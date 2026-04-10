package tui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var defaultByteSliceBucketSizes = [...]int{256, 1024, 4096, 16384, 65536}

type poolCounterSet struct {
	gets        atomic.Uint64
	puts        atomic.Uint64
	misses      atomic.Uint64
	allocations atomic.Uint64
}

func (p *poolCounterSet) snapshot() PoolStats {
	return PoolStats{
		Gets:        p.gets.Load(),
		Puts:        p.puts.Load(),
		Misses:      p.misses.Load(),
		Allocations: p.allocations.Load(),
	}
}

// PoolStats contains usage statistics for a single pool.
type PoolStats struct {
	Gets        uint64
	Puts        uint64
	Misses      uint64
	Allocations uint64
}

// RenderPoolStats contains usage statistics for all render-related pools.
type RenderPoolStats struct {
	StringBuilderPool PoolStats
	ByteSlicePool     PoolStats
	RuneSlicePool     PoolStats
}

type byteSliceBucket struct {
	size int
	pool sync.Pool
}

// RenderPool provides pooled allocations for render-heavy operations.
type RenderPool struct {
	stringBuilders sync.Pool
	runeSlices     sync.Pool
	byteBuckets    []byteSliceBucket

	stringBuilderStats poolCounterSet
	byteSliceStats     poolCounterSet
	runeSliceStats     poolCounterSet
}

// NewRenderPool creates a RenderPool with default buffer buckets.
func NewRenderPool() *RenderPool {
	rp := &RenderPool{
		byteBuckets: make([]byteSliceBucket, 0, len(defaultByteSliceBucketSizes)),
	}

	rp.stringBuilders.New = func() any {
		rp.stringBuilderStats.misses.Add(1)
		rp.stringBuilderStats.allocations.Add(1)
		return &strings.Builder{}
	}

	rp.runeSlices.New = func() any {
		rp.runeSliceStats.misses.Add(1)
		rp.runeSliceStats.allocations.Add(1)
		buf := make([]rune, 0, 256)
		return &buf
	}

	for _, size := range defaultByteSliceBucketSizes {
		bucketSize := size
		bucket := byteSliceBucket{size: bucketSize}
		bucket.pool.New = func() any {
			rp.byteSliceStats.misses.Add(1)
			rp.byteSliceStats.allocations.Add(1)
			buf := make([]byte, 0, bucketSize)
			return &buf
		}
		rp.byteBuckets = append(rp.byteBuckets, bucket)
	}
	return rp
}

// GetStringBuilder returns a reset strings.Builder from the pool.
func (rp *RenderPool) GetStringBuilder() *strings.Builder {
	rp.stringBuilderStats.gets.Add(1)
	builder := rp.stringBuilders.Get().(*strings.Builder)
	builder.Reset()
	return builder
}

// PutStringBuilder returns a strings.Builder to the pool.
func (rp *RenderPool) PutStringBuilder(builder *strings.Builder) {
	if builder == nil {
		return
	}

	rp.stringBuilderStats.puts.Add(1)
	builder.Reset()
	rp.stringBuilders.Put(builder)
}

// GetBytes returns a byte buffer sized to the nearest pool bucket or the exact requested size.
func (rp *RenderPool) GetBytes(size int) []byte {
	rp.byteSliceStats.gets.Add(1)

	if size <= 0 {
		return nil
	}

	bucket := rp.byteBucketForSize(size)
	if bucket == nil {
		rp.byteSliceStats.misses.Add(1)
		rp.byteSliceStats.allocations.Add(1)
		return make([]byte, 0, size)
	}

	buf := *(bucket.pool.Get().(*[]byte))
	return buf[:0]
}

// PutBytes returns a byte slice to the matching pool bucket when possible.
func (rp *RenderPool) PutBytes(buf []byte) {
	if buf == nil {
		return
	}

	rp.byteSliceStats.puts.Add(1)
	bucket := rp.byteBucketForCapacity(cap(buf))
	if bucket == nil {
		return
	}

	buf = buf[:0]
	bucket.pool.Put(&buf)
}

// GetRunes returns a rune slice with at least the requested capacity.
func (rp *RenderPool) GetRunes(size int) []rune {
	rp.runeSliceStats.gets.Add(1)

	if size <= 0 {
		return nil
	}

	buf := *(rp.runeSlices.Get().(*[]rune))
	if cap(buf) < size {
		rp.runeSliceStats.misses.Add(1)
		rp.runeSliceStats.allocations.Add(1)
		return make([]rune, 0, size)
	}

	return buf[:0]
}

// PutRunes returns a rune slice to the pool.
func (rp *RenderPool) PutRunes(buf []rune) {
	if buf == nil {
		return
	}

	rp.runeSliceStats.puts.Add(1)
	buf = buf[:0]
	rp.runeSlices.Put(&buf)
}

// Stats returns a snapshot of the current pool usage statistics.
func (rp *RenderPool) Stats() RenderPoolStats {
	return RenderPoolStats{
		StringBuilderPool: rp.stringBuilderStats.snapshot(),
		ByteSlicePool:     rp.byteSliceStats.snapshot(),
		RuneSlicePool:     rp.runeSliceStats.snapshot(),
	}
}

func (rp *RenderPool) byteBucketForSize(size int) *byteSliceBucket {
	for i := range rp.byteBuckets {
		if size <= rp.byteBuckets[i].size {
			return &rp.byteBuckets[i]
		}
	}

	return nil
}

func (rp *RenderPool) byteBucketForCapacity(capacity int) *byteSliceBucket {
	for i := range rp.byteBuckets {
		if capacity == rp.byteBuckets[i].size {
			return &rp.byteBuckets[i]
		}
	}

	return nil
}

// RenderSchedulerStats contains scheduler state and timing statistics.
type RenderSchedulerStats struct {
	TargetFPS     int
	MaxFPS        int
	FrameInterval time.Duration
	FrameCount    uint64
	DroppedFrames uint64
	AvgFrameTime  time.Duration
	Pending       bool
	Running       bool
}

// RenderScheduler throttles render work to a target frame rate.
type RenderScheduler struct {
	mu            sync.Mutex
	targetFPS     int
	maxFPS        int
	frameInterval time.Duration
	pending       bool
	running       bool
	stopCh        chan struct{}
	doneCh        chan struct{}

	frameCallbacks []func(dt time.Duration)
	rafCallbacks   []func()

	frameCount     atomic.Uint64
	droppedFrames  atomic.Uint64
	totalFrameTime atomic.Int64
}

// NewRenderScheduler creates a scheduler with the provided target FPS.
func NewRenderScheduler(targetFPS int) *RenderScheduler {
	maxFPS := 120
	if targetFPS <= 0 {
		targetFPS = 60
	}
	if targetFPS > maxFPS {
		targetFPS = maxFPS
	}

	return &RenderScheduler{
		targetFPS:      targetFPS,
		maxFPS:         maxFPS,
		frameInterval:  time.Second / time.Duration(targetFPS),
		frameCallbacks: make([]func(dt time.Duration), 0, 1),
		rafCallbacks:   make([]func(), 0, 1),
	}
}

// RequestRender schedules a render on the next frame if one is not already pending.
func (rs *RenderScheduler) RequestRender() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.pending {
		return
	}

	rs.pending = true
}

// RequestAnimationFrame schedules a callback to run on the next frame.
func (rs *RenderScheduler) RequestAnimationFrame(fn func()) {
	if fn == nil {
		return
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.rafCallbacks = append(rs.rafCallbacks, fn)
	rs.pending = true
}

// OnFrame registers a callback that runs whenever a frame is processed.
func (rs *RenderScheduler) OnFrame(fn func(dt time.Duration)) {
	if fn == nil {
		return
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.frameCallbacks = append(rs.frameCallbacks, fn)
}

// Start begins the scheduler loop.
func (rs *RenderScheduler) Start() {
	rs.mu.Lock()
	if rs.running {
		rs.mu.Unlock()
		return
	}

	rs.running = true
	rs.stopCh = make(chan struct{})
	rs.doneCh = make(chan struct{})
	stopCh := rs.stopCh
	doneCh := rs.doneCh
	interval := rs.frameInterval
	rs.mu.Unlock()

	go rs.loop(stopCh, doneCh, interval)
}

// Stop ends the scheduler loop and waits for it to exit.
func (rs *RenderScheduler) Stop() {
	rs.mu.Lock()
	if !rs.running {
		rs.mu.Unlock()
		return
	}

	stopCh := rs.stopCh
	doneCh := rs.doneCh
	rs.running = false
	rs.stopCh = nil
	rs.doneCh = nil
	rs.mu.Unlock()

	close(stopCh)
	<-doneCh
}

// Stats returns a snapshot of scheduler timing and state.
func (rs *RenderScheduler) Stats() RenderSchedulerStats {
	rs.mu.Lock()
	pending := rs.pending
	running := rs.running
	targetFPS := rs.targetFPS
	maxFPS := rs.maxFPS
	frameInterval := rs.frameInterval
	rs.mu.Unlock()

	frameCount := rs.frameCount.Load()
	totalFrameTime := rs.totalFrameTime.Load()
	avgFrameTime := time.Duration(0)
	if frameCount > 0 {
		avgFrameTime = time.Duration(totalFrameTime / int64(frameCount))
	}

	return RenderSchedulerStats{
		TargetFPS:     targetFPS,
		MaxFPS:        maxFPS,
		FrameInterval: frameInterval,
		FrameCount:    frameCount,
		DroppedFrames: rs.droppedFrames.Load(),
		AvgFrameTime:  avgFrameTime,
		Pending:       pending,
		Running:       running,
	}
}

func (rs *RenderScheduler) loop(stopCh, doneCh chan struct{}, interval time.Duration) {
	defer close(doneCh)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	lastFrame := time.Now()
	for {
		select {
		case <-ticker.C:
			rs.processFrame(&lastFrame)
		case <-stopCh:
			return
		}
	}
}

func (rs *RenderScheduler) processFrame(lastFrame *time.Time) {
	rs.mu.Lock()
	shouldRun := rs.pending || len(rs.rafCallbacks) > 0
	if !shouldRun {
		rs.mu.Unlock()
		return
	}

	rafCallbacks := append([]func(){}, rs.rafCallbacks...)
	frameCallbacks := append([]func(time.Duration){}, rs.frameCallbacks...)
	rs.rafCallbacks = rs.rafCallbacks[:0]
	rs.pending = false
	rs.mu.Unlock()

	now := time.Now()
	dt := now.Sub(*lastFrame)
	*lastFrame = now

	started := time.Now()
	for _, fn := range rafCallbacks {
		fn()
	}
	for _, fn := range frameCallbacks {
		fn(dt)
	}
	elapsed := time.Since(started)

	rs.frameCount.Add(1)
	rs.totalFrameTime.Add(int64(elapsed))

	if rs.frameInterval > 0 && elapsed > rs.frameInterval {
		dropped := uint64(elapsed / rs.frameInterval)
		if dropped > 0 {
			rs.droppedFrames.Add(dropped - 1)
		}
	}
}

// BatchWriter batches ANSI output to minimize write syscalls.
type BatchWriter struct {
	mu     sync.Mutex
	writer io.Writer
	buffer strings.Builder
}

// NewBatchWriter creates a batched ANSI writer.
func NewBatchWriter(writer io.Writer) *BatchWriter {
	if writer == nil {
		writer = io.Discard
	}

	return &BatchWriter{writer: writer}
}

// WriteString appends a string to the batch buffer.
func (bw *BatchWriter) WriteString(s string) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	bw.buffer.WriteString(s)
}

// WriteCSI writes a numeric CSI sequence using the SGR final byte.
func (bw *BatchWriter) WriteCSI(params ...int) {
	bw.writeCSICommand('m', params...)
}

// MoveTo moves the cursor to the given x/y terminal position.
func (bw *BatchWriter) MoveTo(x, y int) {
	if x < 1 {
		x = 1
	}
	if y < 1 {
		y = 1
	}

	bw.writeCSICommand('H', y, x)
}

// SetFg sets the foreground color using a named ANSI color.
func (bw *BatchWriter) SetFg(color string) {
	if code, ok := batchWriterANSIColorCode(color, false); ok {
		bw.WriteCSI(code)
	}
}

// SetBg sets the background color using a named ANSI color.
func (bw *BatchWriter) SetBg(color string) {
	if code, ok := batchWriterANSIColorCode(color, true); ok {
		bw.WriteCSI(code)
	}
}

// ResetAttrs resets terminal attributes.
func (bw *BatchWriter) ResetAttrs() {
	bw.WriteCSI(0)
}

// Flush writes the current batch in a single write call.
func (bw *BatchWriter) Flush() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if bw.buffer.Len() == 0 {
		return nil
	}

	_, err := io.WriteString(bw.writer, bw.buffer.String())
	bw.buffer.Reset()
	return err
}

func (bw *BatchWriter) writeCSICommand(final byte, params ...int) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	bw.buffer.WriteString("\x1b[")
	for i, param := range params {
		if i > 0 {
			bw.buffer.WriteByte(';')
		}
		bw.buffer.WriteString(fmt.Sprintf("%d", param))
	}
	bw.buffer.WriteByte(final)
}

var ansiNamedColorCodes = map[string]int{
	"black":          30,
	"red":            31,
	"green":          32,
	"yellow":         33,
	"blue":           34,
	"magenta":        35,
	"cyan":           36,
	"white":          37,
	"default":        39,
	"bright-black":   90,
	"bright-red":     91,
	"bright-green":   92,
	"bright-yellow":  93,
	"bright-blue":    94,
	"bright-magenta": 95,
	"bright-cyan":    96,
	"bright-white":   97,
}

func batchWriterANSIColorCode(color string, background bool) (int, bool) {
	code, ok := ansiNamedColorCodes[strings.ToLower(strings.TrimSpace(color))]
	if !ok {
		return 0, false
	}
	if !background {
		return code, true
	}

	if code >= 90 && code <= 97 {
		return code + 10, true
	}
	if code >= 30 && code <= 37 {
		return code + 10, true
	}
	if code == 39 {
		return 49, true
	}

	return 0, false
}
