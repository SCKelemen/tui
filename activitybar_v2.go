//go:build stack
// +build stack

package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/cli/renderer"
	"github.com/SCKelemen/color"
	design "github.com/SCKelemen/design-system"
	"github.com/SCKelemen/layout"
	"github.com/SCKelemen/text"
)

// ActivityBarV2 uses the full SCKelemen stack for rendering
type ActivityBarV2 struct {
	width       int
	height      int
	message     string
	active      bool
	startTime   time.Time
	elapsed     time.Duration
	spinner     int
	focused     bool
	progress    string
	cancelable  bool
	tokens      *design.DesignTokens
	accentColor *color.Color
}

// NewActivityBarV2 creates an activity bar using the full stack
func NewActivityBarV2(tokens *design.DesignTokens) *ActivityBarV2 {
	accent, _ := color.ParseColor(tokens.Accent)
	return &ActivityBarV2{
		message:     "Ready",
		cancelable:  true,
		tokens:      tokens,
		accentColor: &accent,
	}
}

// Init initializes the activity bar
func (a *ActivityBarV2) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (a *ActivityBarV2) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tickMsg:
		if a.active {
			a.spinner = (a.spinner + 1) % len(spinnerFrames)
			a.elapsed = time.Since(a.startTime)
			return a, a.tick()
		}

	case tea.KeyMsg:
		if a.focused && a.active && a.cancelable && msg.String() == "esc" {
			a.Stop()
		}
	}
	return a, nil
}

// View renders using the SCKelemen stack
func (a *ActivityBarV2) View() string {
	if a.width == 0 {
		return ""
	}

	// Create layout context
	ctx := layout.NewLayoutContext(float64(a.width), float64(a.height), 16)

	// Create root node
	root := &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
			AlignItems:    layout.AlignItemsCenter,
			Width:         layout.Px(float64(a.width)),
			Height:        layout.Ch(1),
			Gap:           layout.Ch(1),
		},
	}

	// Build content
	var content strings.Builder

	if !a.active {
		// Inactive state
		dimmed, _ := color.ParseColor(a.tokens.Color)
		dimmed = dimmed.Darken(0.3)

		txt := text.NewTerminal()
		msg := a.message
		if txt.Width(msg) > float64(a.width) {
			msg = msg[:a.width-3] + "..."
		}

		content.WriteString(dimmed.ToANSI() + msg + "\033[0m")
	} else {
		// Active state with spinner
		spinner := spinnerFrames[a.spinner]

		// Spinner in accent color
		content.WriteString(a.accentColor.ToANSI() + "✳ " + "\033[0m")
		content.WriteString(a.message)

		// Build status parts
		var status []string

		if a.cancelable {
			dimmed, _ := color.ParseColor(a.tokens.Color)
			dimmed = dimmed.Darken(0.3)
			status = append(status, dimmed.ToANSI()+"esc to interrupt"+"\033[0m")
		}

		if a.elapsed > 0 {
			status = append(status, a.formatDuration(a.elapsed))
		}

		if a.progress != "" {
			status = append(status, a.accentColor.ToANSI()+a.progress+"\033[0m")
		}

		if len(status) > 0 {
			content.WriteString(" (")
			content.WriteString(strings.Join(status, " · "))
			content.WriteString(")")
		}
	}

	// Create styled node
	white, _ := color.ParseColor("#FAFAFA")
	style := &renderer.Style{
		Foreground: &white,
	}

	rootStyled := renderer.NewStyledNode(root, style)
	rootStyled.Content = content.String()

	// Layout and render
	constraints := layout.Tight(float64(a.width), float64(a.height))
	layout.Layout(root, constraints, ctx)

	screen := renderer.NewScreen(a.width, a.height)
	screen.Render(rootStyled)

	return screen.String()
}

// Focus is called when this component receives focus
func (a *ActivityBarV2) Focus() {
	a.focused = true
}

// Blur is called when this component loses focus
func (a *ActivityBarV2) Blur() {
	a.focused = false
}

// Focused returns whether this component is currently focused
func (a *ActivityBarV2) Focused() bool {
	return a.focused
}

// Start begins the activity animation
func (a *ActivityBarV2) Start(message string) tea.Cmd {
	a.message = message
	a.active = true
	a.startTime = time.Now()
	a.elapsed = 0
	a.spinner = 0
	return a.tick()
}

// Stop stops the activity animation
func (a *ActivityBarV2) Stop() {
	a.active = false
	a.message = "Ready"
	a.progress = ""
}

// SetProgress updates the progress indicator
func (a *ActivityBarV2) SetProgress(progress string) {
	a.progress = progress
}

// tick returns a command that sends a tickMsg after a delay
func (a *ActivityBarV2) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// formatDuration formats a duration as "1m 14s" or "14s"
func (a *ActivityBarV2) formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}
