package headless

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// namedKeys maps the short, lowercase names accepted by Key onto the
// bubbletea KeyType constants. The names cover the common navigation
// and editing keys plus ctrl+c, which tests use to confirm quit
// handling.
var namedKeys = map[string]tea.KeyType{
	"enter":     tea.KeyEnter,
	"tab":       tea.KeyTab,
	"esc":       tea.KeyEsc,
	"escape":    tea.KeyEsc,
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"backspace": tea.KeyBackspace,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
	"pgup":      tea.KeyPgUp,
	"pgdown":    tea.KeyPgDown,
	"delete":    tea.KeyDelete,
	"space":     tea.KeySpace,
	"ctrl+c":    tea.KeyCtrlC,
}

// Key sends a single named key as a tea.KeyMsg. The accepted names
// are listed in the package documentation.
func (r *Renderer) Key(name string) error {
	t, ok := namedKeys[name]
	if !ok {
		return fmt.Errorf("headless: unknown key %q", name)
	}
	return r.Send(tea.KeyMsg{Type: t})
}
