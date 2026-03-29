package selection

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

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

type clipboardCommand struct {
	name string
	args []string
}

// WriteClipboard writes content to the system clipboard via OSC 52.
func WriteClipboard(content string) tea.Cmd {
	return WriteClipboardTarget(content, ClipboardSystem)
}

// WriteClipboardTarget writes content to a specific clipboard target via OSC 52.
func WriteClipboardTarget(content string, target ClipboardTarget) tea.Cmd {
	return func() tea.Msg {
		err := writeOSC52(content, target)
		return ClipboardWriteMsg{
			Content: content,
			Target:  target,
			Err:     err,
		}
	}
}

// CopyWithFallback copies text using OSC 52 first, then platform-specific commands.
func CopyWithFallback(text string) error {
	if canUseOSC52() {
		if err := writeOSC52(text, ClipboardSystem); err == nil {
			return nil
		}
	}

	cmd, ok := detectCopyCommand()
	if !ok {
		return fmt.Errorf("no supported clipboard copy command found")
	}

	return runCopyCommand(cmd, text)
}

// Read reads clipboard text using platform-specific commands.
func Read() (string, error) {
	cmd, ok := detectReadCommand()
	if !ok {
		return "", fmt.Errorf("no supported clipboard read command found")
	}

	return runReadCommand(cmd)
}

// DetectMethod returns the clipboard method that would be used for copying.
func DetectMethod() string {
	if canUseOSC52() {
		return "osc52"
	}

	cmd, ok := detectCopyCommand()
	if !ok {
		return "none"
	}

	return cmd.name
}

func writeOSC52(content string, target ClipboardTarget) error {
	sequence := osc52Sequence(content, target)
	_, err := os.Stdout.WriteString(sequence)
	return err
}

func canUseOSC52() bool {
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func detectCopyCommand() (clipboardCommand, bool) {
	switch runtime.GOOS {
	case "darwin":
		return commandIfExists("pbcopy")
	case "linux":
		if cmd, ok := commandIfExists("wl-copy"); ok {
			return cmd, true
		}
		if cmd, ok := commandIfExists("xclip", "-selection", "clipboard"); ok {
			return cmd, true
		}
		return commandIfExists("xsel", "--clipboard", "--input")
	default:
		return clipboardCommand{}, false
	}
}

func detectReadCommand() (clipboardCommand, bool) {
	switch runtime.GOOS {
	case "darwin":
		return commandIfExists("pbpaste")
	case "linux":
		if cmd, ok := commandIfExists("wl-paste"); ok {
			return cmd, true
		}
		if cmd, ok := commandIfExists("xclip", "-selection", "clipboard", "-o"); ok {
			return cmd, true
		}
		return commandIfExists("xsel", "--clipboard", "--output")
	default:
		return clipboardCommand{}, false
	}
}

func commandIfExists(name string, args ...string) (clipboardCommand, bool) {
	if _, err := exec.LookPath(name); err != nil {
		return clipboardCommand{}, false
	}

	return clipboardCommand{name: name, args: args}, true
}

func runCopyCommand(cmd clipboardCommand, text string) error {
	execCmd := exec.Command(cmd.name, cmd.args...)
	execCmd.Stdin = strings.NewReader(text)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clipboard copy with %s failed: %w: %s", cmd.name, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runReadCommand(cmd clipboardCommand) (string, error) {
	execCmd := exec.Command(cmd.name, cmd.args...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("clipboard read with %s failed: %w: %s", cmd.name, err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

// osc52Sequence builds an OSC 52 sequence for the given content and target.
func osc52Sequence(content string, target ClipboardTarget) string {
	base64Content := base64.StdEncoding.EncodeToString([]byte(content))
	return fmt.Sprintf("\033]52;%s;%s\a", target, base64Content)
}
