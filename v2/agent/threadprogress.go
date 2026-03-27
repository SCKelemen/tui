package agent

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// Spinner defines an animation sequence.
type Spinner struct {
	Frames []string
}

var (
	// SpinnerDots - Braille dots spinner (smooth).
	SpinnerDots = Spinner{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}

	// SpinnerCircleQuarters - Circle with quarters.
	SpinnerCircleQuarters = Spinner{
		Frames: []string{"◐", "◓", "◑", "◒"},
	}
)

// IconSet defines icons for different statuses.
type IconSet struct {
	Running string
	Success string
	Error   string
	Warning string
	Info    string
	None    string
}

var (
	// IconSetCodex - Codex CLI style.
	IconSetCodex = IconSet{
		Running: "⏺",
		Success: "✓",
		Error:   "✗",
		Warning: "⚠",
		Info:    "⏺",
		None:    "⏺",
	}
)

// GetFrame returns the frame at the given index.
func (s Spinner) GetFrame(index int) string {
	if len(s.Frames) == 0 {
		return ""
	}
	return s.Frames[index%len(s.Frames)]
}

// FrameCount returns the number of frames in the spinner.
func (s Spinner) FrameCount() int {
	return len(s.Frames)
}

// ThreadStatus represents a thread execution state.
type ThreadStatus int

const (
	ThreadRunning ThreadStatus = iota
	ThreadCompleted
	ThreadFailed
	ThreadEscalated
)

// ThreadEntry represents a single thread row and its output.
type ThreadEntry struct {
	ID        string
	Label     string
	Status    ThreadStatus
	Output    []string
	StartTime time.Time
	EndTime   time.Time
	Expanded  bool
	Spinner   Spinner
	spinIdx   int
}

// ThreadProgress displays concurrent thread/task progress.
type ThreadProgress struct {
	width int

	threads []*ThreadEntry
	byID    map[string]*ThreadEntry

	selected int
	focused  bool

	threadSpinner Spinner
	iconSet       IconSet

	maxOutputLines int
	showElapsed    bool

	runningColor   string
	successColor   string
	errorColor     string
	escalatedColor string
	labelColor     string
	timeColor      string
	connectorColor string
	mutedColor     string
}

// ThreadProgressOption configures a ThreadProgress component.
type ThreadProgressOption func(*ThreadProgress)

// WithThreadProgressDesignTokens applies design-system colors.
func WithThreadProgressDesignTokens(tokens *design.DesignTokens) ThreadProgressOption {
	return func(tp *ThreadProgress) {
		tp.applyDesignTokens(tokens)
	}
}

// WithThreadProgressTheme applies a named design-system theme.
func WithThreadProgressTheme(theme string) ThreadProgressOption {
	return func(tp *ThreadProgress) {
		tp.applyDesignTokens(style.DesignTokensForTheme(theme))
	}
}

// WithThreadSpinner sets the default spinner used by running threads.
func WithThreadSpinner(s Spinner) ThreadProgressOption {
	return func(tp *ThreadProgress) {
		tp.threadSpinner = s
	}
}

// WithThreadIconSet sets icons used for thread statuses.
func WithThreadIconSet(is IconSet) ThreadProgressOption {
	return func(tp *ThreadProgress) {
		tp.iconSet = is
	}
}

// WithMaxOutputLines sets max output lines shown when expanded (0 = show all).
func WithMaxOutputLines(n int) ThreadProgressOption {
	return func(tp *ThreadProgress) {
		if n < 0 {
			n = 0
		}
		tp.maxOutputLines = n
	}
}

// WithShowElapsed toggles elapsed-time rendering.
func WithShowElapsed(show bool) ThreadProgressOption {
	return func(tp *ThreadProgress) {
		tp.showElapsed = show
	}
}

// threadProgressTickMsg is emitted periodically to animate running threads.
type threadProgressTickMsg struct{}

// NewThreadProgress creates a new ThreadProgress component.
func NewThreadProgress(opts ...ThreadProgressOption) *ThreadProgress {
	tp := &ThreadProgress{
		threads:        []*ThreadEntry{},
		byID:           map[string]*ThreadEntry{},
		selected:       0,
		threadSpinner:  SpinnerDots,
		iconSet:        IconSetCodex,
		maxOutputLines: 0,
		showElapsed:    true,
		runningColor:   style.ANSICyan,
		successColor:   style.ANSIGreen,
		errorColor:     style.ANSIRed,
		escalatedColor: style.ANSIYellow,
		labelColor:     style.ANSIReset,
		timeColor:      style.ANSIDim,
		connectorColor: style.ANSIDim,
		mutedColor:     style.ANSIDim,
	}

	for _, opt := range opts {
		opt(tp)
	}

	return tp
}

// Init initializes the component.
func (tp *ThreadProgress) Init() tea.Cmd {
	if tp.hasRunningThreads() {
		return tp.tick()
	}
	return nil
}

// Update handles Bubble Tea messages.
func (tp *ThreadProgress) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tp.width = msg.Width

	case threadProgressTickMsg:
		if tp.hasRunningThreads() {
			for _, thread := range tp.threads {
				if thread.Status != ThreadRunning {
					continue
				}
				frames := thread.Spinner.FrameCount()
				if frames == 0 {
					frames = tp.threadSpinner.FrameCount()
				}
				if frames > 0 {
					thread.spinIdx = (thread.spinIdx + 1) % frames
				}
			}
			return tp, tp.tick()
		}

	case tea.KeyMsg:
		if !tp.focused || len(tp.threads) == 0 {
			break
		}
		switch msg.String() {
		case "j", "down":
			tp.selected++
			if tp.selected >= len(tp.threads) {
				tp.selected = len(tp.threads) - 1
			}
		case "k", "up":
			tp.selected--
			if tp.selected < 0 {
				tp.selected = 0
			}
		case "enter", " ":
			thread := tp.threads[tp.selected]
			thread.Expanded = !thread.Expanded
		case "a":
			for _, thread := range tp.threads {
				thread.Expanded = true
			}
		case "c":
			for _, thread := range tp.threads {
				thread.Expanded = false
			}
		}
	}

	return tp, nil
}

// View renders the thread progress UI.
func (tp *ThreadProgress) View() string {
	if tp.width == 0 {
		return ""
	}

	if len(tp.threads) == 0 {
		return fmt.Sprintf("%s(no threads)%s\n", tp.mutedColor, style.ANSIReset)
	}

	var lines []string

	for i, thread := range tp.threads {
		header := tp.renderHeader(thread)
		if tp.focused && i == tp.selected {
			header = style.ANSIInverse + header + style.ANSIReset
		}
		lines = append(lines, tp.fitToWidth(header))

		if !thread.Expanded {
			continue
		}

		if len(thread.Output) == 0 {
			empty := fmt.Sprintf("  %s└──%s %s(no output)%s", tp.connectorColor, style.ANSIReset, tp.mutedColor, style.ANSIReset)
			lines = append(lines, tp.fitToWidth(empty))
			continue
		}

		output := thread.Output
		hidden := 0
		if tp.maxOutputLines > 0 && len(output) > tp.maxOutputLines {
			hidden = len(output) - tp.maxOutputLines
			output = output[:tp.maxOutputLines]
		}

		for outIdx, out := range output {
			connector := "├──"
			if outIdx == len(output)-1 && hidden == 0 {
				connector = "└──"
			}
			line := fmt.Sprintf("  %s%s%s %s", tp.connectorColor, connector, style.ANSIReset, out)
			lines = append(lines, tp.fitToWidth(line))
		}

		if hidden > 0 {
			hint := fmt.Sprintf("  %s└──%s %s… +%d lines%s", tp.connectorColor, style.ANSIReset, tp.mutedColor, hidden, style.ANSIReset)
			lines = append(lines, tp.fitToWidth(hint))
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks the component as focused.
func (tp *ThreadProgress) Focus() {
	tp.focused = true
}

// Blur marks the component as unfocused.
func (tp *ThreadProgress) Blur() {
	tp.focused = false
}

// Focused returns whether the component has focus.
func (tp *ThreadProgress) Focused() bool {
	return tp.focused
}

// AddThread adds (or replaces) a thread by ID.
func (tp *ThreadProgress) AddThread(id, label string) {
	if id == "" {
		return
	}

	if existing := tp.byID[id]; existing != nil {
		existing.Label = label
		return
	}

	entry := &ThreadEntry{
		ID:        id,
		Label:     label,
		Status:    ThreadRunning,
		Output:    []string{},
		StartTime: time.Now(),
		Expanded:  false,
		Spinner:   tp.threadSpinner,
		spinIdx:   0,
	}

	tp.threads = append(tp.threads, entry)
	tp.byID[id] = entry
	if len(tp.threads) == 1 {
		tp.selected = 0
	}
}

// UpdateThread updates thread status and timestamps.
func (tp *ThreadProgress) UpdateThread(id string, status ThreadStatus) {
	thread := tp.byID[id]
	if thread == nil {
		return
	}

	thread.Status = status
	now := time.Now()

	switch status {
	case ThreadRunning:
		if thread.StartTime.IsZero() {
			thread.StartTime = now
		}
		thread.EndTime = time.Time{}
	case ThreadCompleted, ThreadFailed, ThreadEscalated:
		if thread.StartTime.IsZero() {
			thread.StartTime = now
		}
		if thread.EndTime.IsZero() {
			thread.EndTime = now
		}
	}
}

// AppendOutput appends one output line to a thread.
func (tp *ThreadProgress) AppendOutput(id string, line string) {
	thread := tp.byID[id]
	if thread == nil {
		return
	}
	thread.Output = append(thread.Output, line)
}

// SetThreadExpanded sets one thread's expanded state.
func (tp *ThreadProgress) SetThreadExpanded(id string, expanded bool) {
	thread := tp.byID[id]
	if thread == nil {
		return
	}
	thread.Expanded = expanded
}

// RemoveThread removes a thread by ID.
func (tp *ThreadProgress) RemoveThread(id string) {
	if id == "" {
		return
	}
	thread := tp.byID[id]
	if thread == nil {
		return
	}

	idx := -1
	for i, t := range tp.threads {
		if t == thread {
			idx = i
			break
		}
	}
	if idx == -1 {
		delete(tp.byID, id)
		return
	}

	tp.threads = append(tp.threads[:idx], tp.threads[idx+1:]...)
	delete(tp.byID, id)

	if len(tp.threads) == 0 {
		tp.selected = 0
		return
	}
	if tp.selected >= len(tp.threads) {
		tp.selected = len(tp.threads) - 1
	}
}

// GetThread returns a thread entry pointer by ID.
func (tp *ThreadProgress) GetThread(id string) *ThreadEntry {
	return tp.byID[id]
}

// StartThread marks a thread as running and returns animation cmd.
func (tp *ThreadProgress) StartThread(id string) tea.Cmd {
	thread := tp.byID[id]
	if thread == nil {
		return nil
	}
	thread.Status = ThreadRunning
	thread.StartTime = time.Now()
	thread.EndTime = time.Time{}
	thread.spinIdx = 0
	if thread.Spinner.FrameCount() == 0 {
		thread.Spinner = tp.threadSpinner
	}
	return tp.tick()
}

func (tp *ThreadProgress) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return threadProgressTickMsg{}
	})
}

func (tp *ThreadProgress) hasRunningThreads() bool {
	for _, thread := range tp.threads {
		if thread.Status == ThreadRunning {
			return true
		}
	}
	return false
}

func (tp *ThreadProgress) renderHeader(thread *ThreadEntry) string {
	statusIcon := tp.renderStatusIcon(thread.Status)
	parts := []string{statusIcon}

	if thread.Status == ThreadRunning {
		spinner := thread.Spinner
		if spinner.FrameCount() == 0 {
			spinner = tp.threadSpinner
		}
		if spinner.FrameCount() > 0 {
			parts = append(parts, tp.runningColor+spinner.GetFrame(thread.spinIdx)+style.ANSIReset)
		}
	}

	parts = append(parts, tp.labelColor+thread.Label+style.ANSIReset)

	if tp.showElapsed {
		duration := tp.threadDuration(thread)
		if duration >= 0 {
			parts = append(parts, fmt.Sprintf("%s(%s)%s", tp.timeColor, tp.formatDuration(duration), style.ANSIReset))
		}
	}

	return strings.Join(parts, " ")
}

func (tp *ThreadProgress) renderStatusIcon(status ThreadStatus) string {
	switch status {
	case ThreadRunning:
		return tp.runningColor + tp.iconSet.Running + style.ANSIReset
	case ThreadCompleted:
		return tp.successColor + tp.iconSet.Success + style.ANSIReset
	case ThreadFailed:
		return tp.errorColor + tp.iconSet.Error + style.ANSIReset
	case ThreadEscalated:
		return tp.escalatedColor + tp.iconSet.Warning + style.ANSIReset
	default:
		return tp.mutedColor + tp.iconSet.None + style.ANSIReset
	}
}

func (tp *ThreadProgress) threadDuration(thread *ThreadEntry) time.Duration {
	if thread.StartTime.IsZero() {
		return -1
	}
	if thread.Status == ThreadRunning || thread.EndTime.IsZero() {
		return time.Since(thread.StartTime)
	}
	return thread.EndTime.Sub(thread.StartTime)
}

func (tp *ThreadProgress) formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

func (tp *ThreadProgress) fitToWidth(line string) string {
	if tp.width <= 0 {
		return line
	}
	stripped := stripANSI(line)
	if style.StringWidth(stripped) <= tp.width {
		return line
	}
	return style.Truncate(stripped, tp.width, "…")
}
func (tp *ThreadProgress) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	accent := style.ANSIColorFromHex(tokens.Accent)
	foreground := style.ANSIColorFromHex(tokens.Color)

	if accent != "" {
		tp.runningColor = accent
		tp.successColor = accent
		tp.escalatedColor = accent
		tp.connectorColor = accent
	}
	if foreground != "" {
		tp.errorColor = foreground
		tp.labelColor = foreground
		tp.timeColor = foreground
		tp.mutedColor = foreground
	}
}

func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape {
			if r == 'm' {
				inEscape = false
			}
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func truncateANSI(s string, maxWidth int) string {
	var result strings.Builder
	visualWidth := 0
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			result.WriteByte(s[i])
		} else if inEscape {
			result.WriteByte(s[i])
			if s[i] == 'm' {
				inEscape = false
			}
		} else {
			if visualWidth >= maxWidth {
				break
			}
			result.WriteByte(s[i])
			visualWidth++
		}
	}

	return result.String()
}

func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
