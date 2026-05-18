// Package headless provides an in-memory renderer that drives a
// tui.Component lifecycle without a TTY or a live bubbletea program.
//
// It is intended for component-level tests: mount a component, send
// messages, then assert on the resulting View() / Frame() output.
package headless

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	runewidth "github.com/mattn/go-runewidth"

	tui "github.com/SCKelemen/tui/v2"
)

// maxMessages caps the number of messages processed by a single Send
// call to guard against components whose Update returns a cmd that
// indefinitely re-enqueues a message.
const maxMessages = 1000

// ansiSeq matches CSI escape sequences emitted by lipgloss / termenv.
// It is intentionally narrow: \x1b[ ... <letter>. Components that emit
// other escape forms can still rely on View(), which preserves bytes.
var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// Renderer drives a tui.Component in memory. It is not safe for
// concurrent use; tests should hold one renderer per goroutine.
type Renderer struct {
	width  int
	height int
	comp   tui.Component
}

// NewRenderer constructs a Renderer with the given viewport size.
// A width of 0 means Mount will skip the initial tea.WindowSizeMsg,
// which is useful for components that are size-agnostic.
func NewRenderer(width, height int) *Renderer {
	return &Renderer{width: width, height: height}
}

// Mount attaches c to the renderer, dispatches the initial
// tea.WindowSizeMsg (when width > 0) and runs any cmds returned by
// Init() to completion.
func (r *Renderer) Mount(c tui.Component) error {
	if c == nil {
		return errors.New("headless: nil component")
	}
	r.comp = c

	if r.width > 0 {
		if err := r.dispatch(tea.WindowSizeMsg{Width: r.width, Height: r.height}); err != nil {
			return err
		}
	}

	if cmd := c.Init(); cmd != nil {
		if err := r.runCmd(cmd, new(int)); err != nil {
			return err
		}
	}
	return nil
}

// Send delivers msg to the mounted component and drains any cmds the
// resulting Update chain produces. Processing stops on tea.QuitMsg
// and returns an error if the message budget is exhausted.
func (r *Renderer) Send(msg tea.Msg) error {
	if r.comp == nil {
		return errors.New("headless: Send called before Mount")
	}
	return r.dispatch(msg)
}

// View returns the component's current View() output with ANSI
// sequences preserved.
func (r *Renderer) View() string {
	if r.comp == nil {
		return ""
	}
	return r.comp.View()
}

// ViewPlain returns the component's View() with ANSI CSI sequences
// stripped, suitable for text assertions.
func (r *Renderer) ViewPlain() string {
	return stripANSI(r.View())
}

// Frame renders the current view into a width×height grid of runes.
// Cells past the end of a line are padded with spaces; wide runes
// occupy two cells with the right cell set to 0. Lines beyond the
// viewport height are dropped.
func (r *Renderer) Frame() [][]rune {
	w, h := r.width, r.height
	if w <= 0 {
		w = 0
	}
	if h <= 0 {
		h = 0
	}
	grid := make([][]rune, h)
	for i := range grid {
		grid[i] = make([]rune, w)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	plain := stripANSI(r.View())
	lines := strings.Split(plain, "\n")
	for row := 0; row < h && row < len(lines); row++ {
		col := 0
		for _, ch := range lines[row] {
			if col >= w {
				break
			}
			rw := runewidth.RuneWidth(ch)
			if rw <= 0 {
				// Zero-width runes (combining marks etc.) are
				// dropped; tests rarely care, and they would
				// otherwise need their own column.
				continue
			}
			grid[row][col] = ch
			if rw == 2 && col+1 < w {
				grid[row][col+1] = 0
			}
			col += rw
		}
	}
	return grid
}

// Component returns the mounted component, or nil if Mount has not
// been called.
func (r *Renderer) Component() tui.Component {
	return r.comp
}

// Keystroke sends one tea.KeyMsg per rune in s.
func (r *Renderer) Keystroke(s string) error {
	for _, ch := range s {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		if err := r.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

// dispatch is the shared entry point used by Mount and Send. It
// installs a fresh budget counter and routes through runMsg so that
// the recursion cap covers every cmd chain regardless of origin.
func (r *Renderer) dispatch(msg tea.Msg) error {
	budget := 0
	return r.runMsg(msg, &budget)
}

// runMsg handles a single message: it stops on QuitMsg, expands
// BatchMsg, otherwise calls Update and drains the returned cmd.
func (r *Renderer) runMsg(msg tea.Msg, budget *int) error {
	if msg == nil {
		return nil
	}
	*budget++
	if *budget > maxMessages {
		return fmt.Errorf("headless: exceeded %d messages in a single Send", maxMessages)
	}

	switch m := msg.(type) {
	case tea.QuitMsg:
		return nil
	case tea.BatchMsg:
		for _, cmd := range m {
			if err := r.runCmd(cmd, budget); err != nil {
				return err
			}
		}
		return nil
	}

	next, cmd := r.comp.Update(msg)
	if next != nil {
		r.comp = next
	}
	if cmd != nil {
		return r.runCmd(cmd, budget)
	}
	return nil
}

// runCmd invokes cmd and dispatches whatever message it returns. A
// cmd that produces a nil message is a no-op.
func (r *Renderer) runCmd(cmd tea.Cmd, budget *int) error {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if msg == nil {
		return nil
	}
	return r.runMsg(msg, budget)
}

// stripANSI removes CSI escape sequences from s.
func stripANSI(s string) string {
	if !strings.ContainsRune(s, 0x1b) {
		return s
	}
	return ansiSeq.ReplaceAllString(s, "")
}
