package tui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderPool(t *testing.T) {
	rp := NewRenderPool()

	builder := rp.GetStringBuilder()
	builder.WriteString("hello")
	if builder.String() != "hello" {
		t.Fatalf("builder contents = %q, want hello", builder.String())
	}
	rp.PutStringBuilder(builder)

	reused := rp.GetStringBuilder()
	if reused.Len() != 0 {
		t.Fatalf("reused builder len = %d, want 0", reused.Len())
	}
	rp.PutStringBuilder(reused)

	for _, size := range defaultByteSliceBucketSizes {
		buf := rp.GetBytes(size)
		if len(buf) != 0 {
			t.Fatalf("GetBytes(%d) len = %d, want 0", size, len(buf))
		}
		if cap(buf) != size {
			t.Fatalf("GetBytes(%d) cap = %d, want %d", size, cap(buf), size)
		}
		buf = append(buf, 'x')
		rp.PutBytes(buf)

		again := rp.GetBytes(size)
		if cap(again) != size {
			t.Fatalf("GetBytes(%d) after PutBytes cap = %d, want %d", size, cap(again), size)
		}
		rp.PutBytes(again)
	}

	oversized := rp.GetBytes(defaultByteSliceBucketSizes[len(defaultByteSliceBucketSizes)-1] + 1)
	if cap(oversized) != defaultByteSliceBucketSizes[len(defaultByteSliceBucketSizes)-1]+1 {
		t.Fatalf("oversized GetBytes cap = %d, want exact size", cap(oversized))
	}

	runes := rp.GetRunes(32)
	if len(runes) != 0 {
		t.Fatalf("GetRunes len = %d, want 0", len(runes))
	}
	if cap(runes) < 32 {
		t.Fatalf("GetRunes cap = %d, want >= 32", cap(runes))
	}
	runes = append(runes, 'α', 'β')
	rp.PutRunes(runes)

	runes = rp.GetRunes(16)
	if len(runes) != 0 {
		t.Fatalf("reused rune slice len = %d, want 0", len(runes))
	}
	rp.PutRunes(runes)

	stats := rp.Stats()
	if stats.StringBuilderPool.Gets < 2 || stats.StringBuilderPool.Puts < 2 {
		t.Fatalf("unexpected string builder stats: %+v", stats.StringBuilderPool)
	}
	if stats.ByteSlicePool.Gets < uint64(len(defaultByteSliceBucketSizes)+1) {
		t.Fatalf("unexpected byte slice stats: %+v", stats.ByteSlicePool)
	}
	if stats.RuneSlicePool.Gets < 2 || stats.RuneSlicePool.Puts < 2 {
		t.Fatalf("unexpected rune slice stats: %+v", stats.RuneSlicePool)
	}
}

func TestRenderScheduler(t *testing.T) {
	rs := NewRenderScheduler(120)
	if rs.Stats().Running {
		t.Fatal("scheduler should not be running before Start")
	}

	frameCh := make(chan time.Duration, 2)
	rafCh := make(chan struct{}, 1)
	rs.OnFrame(func(dt time.Duration) {
		frameCh <- dt
	})
	rs.Start()
	defer rs.Stop()

	rs.RequestAnimationFrame(func() {
		rafCh <- struct{}{}
	})
	rs.RequestRender()

	select {
	case <-rafCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for animation frame callback")
	}

	select {
	case dt := <-frameCh:
		if dt <= 0 {
			t.Fatalf("frame dt = %v, want > 0", dt)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for frame callback")
	}

	stats := rs.Stats()
	if !stats.Running {
		t.Fatal("scheduler should report running after Start")
	}
	if stats.Pending {
		t.Fatal("scheduler should not remain pending after frame processed")
	}
	if stats.FrameCount < 1 {
		t.Fatalf("FrameCount = %d, want >= 1", stats.FrameCount)
	}
	if stats.TargetFPS != 120 || stats.MaxFPS != 120 {
		t.Fatalf("unexpected scheduler stats: %+v", stats)
	}

	rs.Stop()
	if rs.Stats().Running {
		t.Fatal("scheduler should not be running after Stop")
	}
}

func TestBatchWriter(t *testing.T) {
	var buf bytes.Buffer
	bw := NewBatchWriter(&buf)

	bw.WriteString("hello")
	bw.MoveTo(2, 3)
	bw.SetFg("red")
	bw.SetBg("blue")
	bw.WriteCSI(1, 4)
	bw.writeCSICommand('K', 2)

	if err := bw.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	got := buf.String()
	for _, want := range []string{"hello", "\x1b[3;2H", "\x1b[31m", "\x1b[44m", "\x1b[1;4m", "\x1b[2K"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Flush output %q missing %q", got, want)
		}
	}

	before := buf.Len()
	if err := bw.Flush(); err != nil {
		t.Fatalf("second Flush() error = %v", err)
	}
	if buf.Len() != before {
		t.Fatalf("second Flush changed buffer len from %d to %d", before, buf.Len())
	}
}
