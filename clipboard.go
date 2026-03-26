package tui

import (
	"encoding/base64"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// ClipboardTarget represents the OSC 52 clipboard target.
type ClipboardTarget string

const (
	// ClipboardSystem is the system clipboard target.
	ClipboardSystem ClipboardTarget = "c"
	// ClipboardPrimary is the primary selection target (X11).
	ClipboardPrimary ClipboardTarget = "p"
)

// ClipboardWriteMsg is returned after attempting to write clipboard content.
type ClipboardWriteMsg struct {
	Content string
	Target  ClipboardTarget
	Err     error
}

// WriteClipboard writes content to the system clipboard via OSC 52.
func WriteClipboard(content string) tea.Cmd {
	return WriteClipboardTarget(content, ClipboardSystem)
}

// WriteClipboardTarget writes content to a specific clipboard target via OSC 52.
func WriteClipboardTarget(content string, target ClipboardTarget) tea.Cmd {
	return func() tea.Msg {
		sequence := osc52Sequence(content, target)
		_, err := os.Stdout.WriteString(sequence)
		return ClipboardWriteMsg{
			Content: content,
			Target:  target,
			Err:     err,
		}
	}
}

// osc52Sequence builds an OSC 52 sequence for the given content and target.
func osc52Sequence(content string, target ClipboardTarget) string {
	base64Content := base64.StdEncoding.EncodeToString([]byte(content))
	return fmt.Sprintf(string([]byte{0x1b})+"]52;%s;%s"+string([]byte{0x07}), target, base64Content)
}
