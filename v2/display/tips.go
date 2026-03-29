package display

import (
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	highlightOpenTag  = "{highlight}"
	highlightCloseTag = "{/highlight}"
)

// Tip represents a single tip entry.
type Tip struct {
	Text string
}

// Tips rotates through a list of CLI tips and renders styled output.
type Tips struct {
	tips           []Tip
	currentIndex   int
	highlightColor string
	dimColor       string
	width          int
}

// TipsOption configures the Tips component.
type TipsOption func(*Tips)

// NewTips creates a tips display with optional configuration.
func NewTips(tips []Tip, opts ...TipsOption) *Tips {
	t := &Tips{
		tips:           append([]Tip(nil), tips...),
		currentIndex:   0,
		highlightColor: "212",
		dimColor:       "245",
		width:          0,
	}

	for _, opt := range opts {
		opt(t)
	}

	if len(t.tips) == 0 {
		t.currentIndex = 0
	} else if t.currentIndex < 0 || t.currentIndex >= len(t.tips) {
		t.currentIndex = 0
	}

	return t
}

// WithTipsHighlightColor sets the color used for highlighted tip content.
func WithTipsHighlightColor(c string) TipsOption {
	return func(t *Tips) {
		t.highlightColor = c
	}
}

// WithTipsDimColor sets the color used for non-highlighted tip content.
func WithTipsDimColor(c string) TipsOption {
	return func(t *Tips) {
		t.dimColor = c
	}
}

// WithTipsWidth sets the maximum rendered width for the tips view.
func WithTipsWidth(w int) TipsOption {
	return func(t *Tips) {
		if w > 0 {
			t.width = w
		}
	}
}

// Next advances to the next tip and wraps around at the end.
func (t *Tips) Next() {
	if len(t.tips) == 0 {
		return
	}
	t.currentIndex = (t.currentIndex + 1) % len(t.tips)
}

// RandomTip selects a random tip index.
func (t *Tips) RandomTip() {
	if len(t.tips) == 0 {
		return
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	t.currentIndex = r.Intn(len(t.tips))
}

// View renders the current tip as: Tip: <text>.
//
// The markup tags {highlight} and {/highlight} are parsed and the enclosed
// content is rendered with the highlight color.
func (t *Tips) View() string {
	if len(t.tips) == 0 {
		return ""
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.dimColor))
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.highlightColor))

	segments := parseTipMarkup(t.tips[t.currentIndex].Text)

	var b strings.Builder
	b.WriteString(dimStyle.Render("Tip: "))
	for _, segment := range segments {
		if segment.highlight {
			b.WriteString(highlightStyle.Render(segment.text))
			continue
		}
		b.WriteString(dimStyle.Render(segment.text))
	}

	out := b.String()
	if t.width > 0 {
		out = lipgloss.NewStyle().MaxWidth(t.width).Render(out)
	}

	return out
}

type tipSegment struct {
	text      string
	highlight bool
}

func parseTipMarkup(text string) []tipSegment {
	if text == "" {
		return nil
	}

	segments := make([]tipSegment, 0, 4)
	remaining := text

	for len(remaining) > 0 {
		openIndex := strings.Index(remaining, highlightOpenTag)
		if openIndex < 0 {
			segments = append(segments, tipSegment{text: remaining})
			break
		}

		if openIndex > 0 {
			segments = append(segments, tipSegment{text: remaining[:openIndex]})
		}

		highlightStart := openIndex + len(highlightOpenTag)
		closeIndex := strings.Index(remaining[highlightStart:], highlightCloseTag)
		if closeIndex < 0 {
			segments = append(segments, tipSegment{text: remaining[openIndex:]})
			break
		}

		closeIndex += highlightStart
		segments = append(segments, tipSegment{
			text:      remaining[highlightStart:closeIndex],
			highlight: true,
		})

		remaining = remaining[closeIndex+len(highlightCloseTag):]
	}

	return segments
}

// DefaultTips returns a reusable list of practical coding CLI tips.
func DefaultTips() []Tip {
	return []Tip{
		{Text: "Use {highlight}rg --hidden --glob '!vendor' 'pattern'{/highlight} for fast codebase search."},
		{Text: "Preview command effects with {highlight}git diff{/highlight} before committing."},
		{Text: "Use {highlight}git add -p{/highlight} to stage only the exact hunks you want."},
		{Text: "Find large files quickly with {highlight}du -h | sort -h{/highlight}."},
		{Text: "Use {highlight}go test ./...{/highlight} often to catch regressions early."},
		{Text: "Run {highlight}go test -run TestName -v{/highlight} to iterate on a single failing test."},
		{Text: "Pipe long output into {highlight}less -R{/highlight} to keep colorized logs readable."},
		{Text: "Use {highlight}jq{/highlight} to inspect and transform JSON output from APIs."},
		{Text: "Check open ports with {highlight}lsof -i -P -n{/highlight} while debugging local services."},
		{Text: "Use {highlight}git restore --staged <file>{/highlight} to unstage without losing edits."},
		{Text: "Find slow tests by running {highlight}go test -json ./... | jq{/highlight}."},
		{Text: "Use {highlight}xargs -P{/highlight} for safe parallel command execution."},
		{Text: "Search command history quickly with {highlight}Ctrl+R{/highlight}."},
		{Text: "Use {highlight}git bisect{/highlight} to identify the commit that introduced a bug."},
		{Text: "Validate formatting with {highlight}gofmt -w{/highlight} before opening a pull request."},
		{Text: "Use {highlight}go vet ./...{/highlight} to catch suspicious code patterns."},
		{Text: "Trace HTTP calls with {highlight}curl -v{/highlight} when integrations fail."},
		{Text: "Use {highlight}watch -n 1 <cmd>{/highlight} to monitor changing output in real time."},
		{Text: "Generate coverage with {highlight}go test ./... -coverprofile=cover.out{/highlight}."},
		{Text: "Inspect binary/module deps with {highlight}go list -deps ./...{/highlight}."},
	}
}
