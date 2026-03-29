package style

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// TerminalColorScheme represents whether the terminal theme is dark or light.
type TerminalColorScheme string

const (
	// ColorSchemeDark indicates a dark terminal background.
	ColorSchemeDark TerminalColorScheme = "dark"
	// ColorSchemeLight indicates a light terminal background.
	ColorSchemeLight TerminalColorScheme = "light"
)

// TerminalColors contains detected foreground, background, and 16-color palette values.
type TerminalColors struct {
	Background string
	Foreground string
	Palette    [16]string
}

const terminalQueryTimeout = 500 * time.Millisecond

// DetectColorScheme queries the terminal background color (OSC 11) and returns dark or light.
// If detection fails or times out, it falls back to dark.
func DetectColorScheme() TerminalColorScheme {
	r, g, b, err := querySingleColor("11", terminalQueryTimeout)
	if err != nil {
		return ColorSchemeDark
	}

	if luminance(r, g, b) >= 128 {
		return ColorSchemeLight
	}

	return ColorSchemeDark
}

// QueryTerminalColors queries terminal foreground/background/palette via OSC 10, 11, and 4.
// It uses a total timeout of 500ms.
func QueryTerminalColors() (*TerminalColors, error) {
	stdinFD := int(os.Stdin.Fd())
	if !term.IsTerminal(stdinFD) {
		return nil, fmt.Errorf("stdin is not a terminal")
	}

	state, err := term.MakeRaw(stdinFD)
	if err != nil {
		return nil, fmt.Errorf("enable raw mode: %w", err)
	}
	defer func() {
		_ = term.Restore(stdinFD, state)
	}()

	deadline := time.Now().Add(terminalQueryTimeout)
	colors := &TerminalColors{}

	r, g, b, err := querySingleColorInRaw("10", deadline)
	if err != nil {
		return nil, err
	}
	colors.Foreground = rgbHex(r, g, b)

	r, g, b, err = querySingleColorInRaw("11", deadline)
	if err != nil {
		return nil, err
	}
	colors.Background = rgbHex(r, g, b)

	for i := range colors.Palette {
		code := fmt.Sprintf("4;%d", i)
		r, g, b, err = querySingleColorInRaw(code, deadline)
		if err != nil {
			return nil, err
		}
		colors.Palette[i] = rgbHex(r, g, b)
	}

	return colors, nil
}

func querySingleColor(code string, timeout time.Duration) (uint8, uint8, uint8, error) {
	stdinFD := int(os.Stdin.Fd())
	if !term.IsTerminal(stdinFD) {
		return 0, 0, 0, fmt.Errorf("stdin is not a terminal")
	}

	state, err := term.MakeRaw(stdinFD)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("enable raw mode: %w", err)
	}
	defer func() {
		_ = term.Restore(stdinFD, state)
	}()

	return querySingleColorInRaw(code, time.Now().Add(timeout))
}

func querySingleColorInRaw(code string, deadline time.Time) (uint8, uint8, uint8, error) {
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return 0, 0, 0, fmt.Errorf("terminal color query timeout")
	}

	request := fmt.Sprintf("\x1b]%s;?\x07", code)
	if _, err := os.Stdout.WriteString(request); err != nil {
		return 0, 0, 0, fmt.Errorf("write osc query %q: %w", code, err)
	}

	response, err := readOSCResponse(remaining)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read osc response %q: %w", code, err)
	}

	payload, ok := extractOSCPayload(response)
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid osc response %q", response)
	}

	r, g, b, ok := parseColorFromPayload(payload)
	if !ok {
		return 0, 0, 0, fmt.Errorf("unable to parse osc color from %q", payload)
	}

	return r, g, b, nil
}

func parseColorFromPayload(payload string) (uint8, uint8, uint8, bool) {
	parts := strings.Split(payload, ";")
	for i := len(parts) - 1; i >= 0; i-- {
		r, g, b, ok := parseOSCColor(strings.TrimSpace(parts[i]))
		if ok {
			return r, g, b, true
		}
	}
	return 0, 0, 0, false
}

func readOSCResponse(timeout time.Duration) (string, error) {
	type readResult struct {
		response string
		err      error
	}

	ch := make(chan readResult, 1)
	go func() {
		var data []byte
		buf := make([]byte, 1)
		esc := false

		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				ch <- readResult{err: err}
				return
			}
			if n == 0 {
				continue
			}

			b := buf[0]
			data = append(data, b)
			if b == '\a' {
				ch <- readResult{response: string(data)}
				return
			}
			if esc && b == '\\' {
				ch <- readResult{response: string(data)}
				return
			}
			esc = b == 0x1b
		}
	}()

	select {
	case result := <-ch:
		if result.err != nil {
			return "", result.err
		}
		return result.response, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout")
	}
}

func extractOSCPayload(response string) (string, bool) {
	if !strings.HasPrefix(response, "\x1b]") {
		return "", false
	}

	payload := strings.TrimPrefix(response, "\x1b]")
	if strings.HasSuffix(payload, "\a") {
		return strings.TrimSuffix(payload, "\a"), true
	}
	if strings.HasSuffix(payload, "\x1b\\") {
		return strings.TrimSuffix(payload, "\x1b\\"), true
	}

	return "", false
}

// parseOSCColor parses terminal OSC color values in either of these forms:
//   - rgb:RRRR/GGGG/BBBB
//   - #RRGGBB
func parseOSCColor(s string) (r, g, b uint8, ok bool) {
	token := strings.TrimSpace(s)
	if strings.HasPrefix(token, "#") {
		if len(token) != 7 {
			return 0, 0, 0, false
		}
		value, err := strconv.ParseUint(token[1:], 16, 32)
		if err != nil {
			return 0, 0, 0, false
		}
		return uint8(value >> 16), uint8((value >> 8) & 0xFF), uint8(value & 0xFF), true
	}

	if !strings.HasPrefix(token, "rgb:") {
		return 0, 0, 0, false
	}

	channels := strings.Split(strings.TrimPrefix(token, "rgb:"), "/")
	if len(channels) != 3 {
		return 0, 0, 0, false
	}

	parseChannel := func(part string) (uint8, bool) {
		if len(part) == 0 || len(part) > 4 {
			return 0, false
		}
		value, err := strconv.ParseUint(part, 16, 16)
		if err != nil {
			return 0, false
		}
		max := (uint64(1) << (uint(len(part)) * 4)) - 1
		if max == 0 {
			return 0, false
		}
		return uint8((value*255 + max/2) / max), true
	}

	r, ok = parseChannel(channels[0])
	if !ok {
		return 0, 0, 0, false
	}
	g, ok = parseChannel(channels[1])
	if !ok {
		return 0, 0, 0, false
	}
	b, ok = parseChannel(channels[2])
	if !ok {
		return 0, 0, 0, false
	}

	return r, g, b, true
}

// luminance returns relative luminance using an RGB weighted average.
func luminance(r, g, b uint8) float64 {
	return 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
}

func rgbHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
