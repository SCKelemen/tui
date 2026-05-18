package selection

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
)

// DefaultMaxOSC52PayloadSize is the maximum number of bytes accepted by the
// OSC 52 clipboard write helpers. Terminals impose varying limits on OSC 52
// sequence length (xterm ~4 KB, iTerm2 ~100 KB by default); 100 KB is a
// pragmatic default that works on most modern terminals. Use
// SetMaxOSC52PayloadSize to override.
const DefaultMaxOSC52PayloadSize = 100 * 1024

// ErrOSC52PayloadTooLarge is returned by clipboard write helpers when the
// content exceeds the configured size limit.
var ErrOSC52PayloadTooLarge = errors.New("osc52: clipboard payload exceeds configured size limit")

// osc52PayloadLimit stores the configured maximum payload size. Defaults to
// DefaultMaxOSC52PayloadSize.
var osc52PayloadLimit atomic.Int64

func init() {
	osc52PayloadLimit.Store(int64(DefaultMaxOSC52PayloadSize))
}

// SetMaxOSC52PayloadSize sets the maximum payload size for OSC 52 clipboard
// writes and returns the previous limit. Pass 0 or a negative value to
// disable the limit entirely.
func SetMaxOSC52PayloadSize(n int) (previous int) {
	return int(osc52PayloadLimit.Swap(int64(n)))
}

// MaxOSC52PayloadSize returns the current maximum payload size for OSC 52
// clipboard writes.
func MaxOSC52PayloadSize() int {
	return int(osc52PayloadLimit.Load())
}

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
	if limit := MaxOSC52PayloadSize(); limit > 0 && len(content) > limit {
		return ErrOSC52PayloadTooLarge
	}
	sequence := OSC52Sequence(target, content)
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

// OSC 52 clipboard contract.
//
// The OSC 52 escape sequence asks the host terminal to replace the
// contents of one of its clipboards with the base64-encoded payload
// supplied in-band. The wire format is fixed by xterm:
//
//	ESC ] 5 2 ; <target> ; <base64 payload> BEL
//
// where <target> is "c" (system clipboard) or "p" (X11 primary
// selection) and BEL is the ASCII bell character (0x07). The full
// reference is documented in the xterm control sequences manual:
// https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Operating-System-Commands
//
// OSC52Sequence is a pure builder: it performs no I/O, no environment
// inspection, and no terminal detection. Callers that want to know
// whether the host terminal will honour the sequence should consult
// SupportsOSC52 separately.
//
// SupportsOSC52 reports whether a modern terminal is likely to accept
// the sequence. The check is heuristic and based on environment
// variables only.

// OSC52Sequence builds an OSC 52 escape sequence for the supplied
// clipboard target and payload. The function is total: any string
// (including empty input or arbitrary binary in UTF-8 form) is valid.
//
// Example: OSC52Sequence(ClipboardSystem, "hello") returns
// "\x1b]52;c;aGVsbG8=\x07".
func OSC52Sequence(target ClipboardTarget, text string) string {
	base64Content := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("\x1b]52;%s;%s\x07", target, base64Content)
}

// osc52KnownTerminals lists TERM_PROGRAM values that are expected to
// honour OSC 52 clipboard writes. The list mirrors the modern-terminal
// allow-list used by tui.DetectCapabilities so the two helpers stay in
// agreement.
var osc52KnownTerminals = map[string]struct{}{
	"iterm.app":      {},
	"vscode":         {},
	"apple_terminal": {},
	"wezterm":        {},
	"ghostty":        {},
	"kitty":          {},
	"alacritty":      {},
	"tmux":           {},
}

// SupportsOSC52 reports whether the host terminal is likely to honour
// OSC 52 clipboard writes. The check is heuristic and based on
// TERM_PROGRAM, TERM, and KITTY_WINDOW_ID — no terminal queries are
// issued. A "true" result is best-effort, not a guarantee.
//
// Most modern terminal emulators (iTerm2, WezTerm, Ghostty, Kitty,
// Alacritty, VS Code's integrated terminal, Apple Terminal, tmux) are
// recognised. Unknown or legacy terminals return false.
func SupportsOSC52() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("TERM")), "dumb") {
		return false
	}
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}
	if term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM"))); term != "" {
		if term == "xterm-kitty" || strings.Contains(term, "kitty") {
			return true
		}
	}
	if program := strings.ToLower(strings.TrimSpace(os.Getenv("TERM_PROGRAM"))); program != "" {
		if _, ok := osc52KnownTerminals[program]; ok {
			return true
		}
	}
	return false
}
