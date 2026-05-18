package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestDisplayWidthPlainASCII(t *testing.T) {
	if got := displayWidth("hello"); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestDisplayWidthSkipsANSICSI(t *testing.T) {
	// red text "hi" with reset, plus a truecolor sequence.
	s := "\x1b[31mhi\x1b[0m \x1b[38;2;255;128;0mxy\x1b[0m"
	if got := displayWidth(s); got != 5 {
		t.Fatalf("expected width 5 for %q, got %d", s, got)
	}
}

func TestDisplayWidthSkipsOSC(t *testing.T) {
	// Hyperlink OSC 8 with BEL terminator.
	s := "\x1b]8;;https://example.com\x07link\x1b]8;;\x07"
	if got := displayWidth(s); got != 4 {
		t.Fatalf("expected width 4 for hyperlink %q, got %d", s, got)
	}

	// OSC terminated by ST (ESC \\).
	s2 := "\x1b]0;title\x1b\\done"
	if got := displayWidth(s2); got != 4 {
		t.Fatalf("expected width 4 for OSC+ST %q, got %d", s2, got)
	}
}

func TestDisplayWidthCJK(t *testing.T) {
	// Each Han character is 2 cells.
	s := "日本語"
	if got := displayWidth(s); got != 6 {
		t.Fatalf("expected width 6 for CJK %q, got %d", s, got)
	}
}

func TestDisplayWidthEmoji(t *testing.T) {
	// A single base emoji is 2 cells in most terminals.
	if got := displayWidth("🙂"); got != 2 {
		t.Fatalf("expected width 2 for emoji, got %d", got)
	}
}

func TestNormalizeFrameLinePlainASCII(t *testing.T) {
	out := normalizeFrameLine("abc", 5)
	if displayWidth(out) != 5 {
		t.Fatalf("expected display width 5, got %d (%q)", displayWidth(out), out)
	}
	if out != "abc  " {
		t.Fatalf("expected padded 'abc  ', got %q", out)
	}

	out = normalizeFrameLine("abcdefg", 4)
	if displayWidth(out) != 4 {
		t.Fatalf("expected display width 4 after truncation, got %d (%q)", displayWidth(out), out)
	}
}

func TestNormalizeFrameLineTrueColor(t *testing.T) {
	line := "\x1b[38;2;0;255;0mgreen\x1b[0m and plain"
	// Source display width = 5 + 1 + 9 = 15.
	if dw := displayWidth(line); dw != 15 {
		t.Fatalf("baseline displayWidth wrong: %d", dw)
	}

	out := normalizeFrameLine(line, 20)
	if displayWidth(out) != 20 {
		t.Fatalf("expected padded width 20, got %d (%q)", displayWidth(out), out)
	}
	if !strings.Contains(out, "\x1b[38;2;0;255;0m") {
		t.Fatalf("expected truecolor escape preserved: %q", out)
	}

	// Truncate inside the colored region — must emit a reset.
	out = normalizeFrameLine(line, 3)
	if displayWidth(out) != 3 {
		t.Fatalf("expected truncated width 3, got %d (%q)", displayWidth(out), out)
	}
	if !strings.HasSuffix(out, "\x1b[0m") {
		t.Fatalf("expected SGR reset after truncating styled content: %q", out)
	}
}

func TestNormalizeFrameLineCJK(t *testing.T) {
	// "日本語" => 6 cells. Pad to 10.
	out := normalizeFrameLine("日本語", 10)
	if displayWidth(out) != 10 {
		t.Fatalf("expected width 10, got %d", displayWidth(out))
	}

	// Truncate to width 4: must contain "日本" (4 cells) and no half-char.
	out = normalizeFrameLine("日本語", 4)
	if displayWidth(out) != 4 {
		t.Fatalf("expected width 4 after truncation, got %d (%q)", displayWidth(out), out)
	}
	if !strings.HasPrefix(out, "日本") {
		t.Fatalf("expected truncation to keep whole wide chars, got %q", out)
	}

	// Truncate to width 3: must NOT split the wide char — should keep one (2)
	// then pad with a single space to reach 3.
	out = normalizeFrameLine("日本語", 3)
	if displayWidth(out) != 3 {
		t.Fatalf("expected width 3 after constrained truncation, got %d (%q)", displayWidth(out), out)
	}
	if !strings.HasPrefix(out, "日") {
		t.Fatalf("expected first wide char to survive, got %q", out)
	}
}

func TestNormalizeFrameLineEmojiZWJ(t *testing.T) {
	// "👩‍💻" is woman + ZWJ + laptop = visually 1 glyph; many terminals
	// render it as 2 cells (woman) plus the ZWJ joiner (0) plus laptop (2)
	// = up to 4 cells depending on terminfo. We just assert padding makes
	// the output width match the target.
	line := "👩\u200d💻 hi"
	out := normalizeFrameLine(line, 12)
	if got := displayWidth(out); got != 12 {
		t.Fatalf("expected width 12 for ZWJ line, got %d (%q)", got, out)
	}
}

func TestNormalizeFrameLineMixedANSIAndWide(t *testing.T) {
	line := "\x1b[1mBOLD\x1b[0m日本"
	// Width = 4 + 4 = 8.
	if dw := displayWidth(line); dw != 8 {
		t.Fatalf("baseline width wrong: %d", dw)
	}
	out := normalizeFrameLine(line, 8)
	if displayWidth(out) != 8 {
		t.Fatalf("expected width 8, got %d (%q)", displayWidth(out), out)
	}

	// Truncate inside the wide part — width 5 should fit BOLD + half of 日本
	// safely (BOLD = 4, then "日" would push to 6, so we stop at 4 and pad).
	out = normalizeFrameLine(line, 5)
	if displayWidth(out) != 5 {
		t.Fatalf("expected width 5 truncated, got %d (%q)", displayWidth(out), out)
	}
}

func TestNewFrameBufferDefaults(t *testing.T) {
	fb := NewFrameBuffer(nil, 0, 0)
	if fb == nil {
		t.Fatal("NewFrameBuffer returned nil")
	}
	if fb.writer == nil {
		t.Fatal("expected writer to default to io.Discard")
	}
}

func TestFrameBufferRenderProducesOutput(t *testing.T) {
	var buf bytes.Buffer
	fb := NewFrameBuffer(&buf, 10, 2)
	out := fb.Render("hello\nworld")
	if out == "" {
		t.Fatal("expected non-empty render on first frame")
	}
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Fatalf("expected frame content, got %q", out)
	}
}

func TestFrameBufferDiffHidesCursor(t *testing.T) {
	var buf bytes.Buffer
	fb := NewFrameBuffer(&buf, 10, 6)

	frame1 := "row0\nrow1\nrow2\nrow3\nrow4\nrow5"
	_ = fb.Render(frame1)
	buf.Reset()

	// Change only row index 4 (1-based: row 5).
	frame2 := "row0\nrow1\nrow2\nrow3\nCHANGED\nrow5"
	out := fb.Render(frame2)
	if out == "" {
		t.Fatal("expected diff output for changed frame")
	}
	if !strings.HasPrefix(out, "\x1b[?25l") {
		t.Fatalf("expected diff output to start with hide-cursor, got %q", out)
	}
	if !strings.HasSuffix(out, "\x1b[?25h") {
		t.Fatalf("expected diff output to end with show-cursor, got %q", out)
	}
	if !strings.Contains(out, "CHANGED") {
		t.Fatalf("expected diff output to include the changed row, got %q", out)
	}
}

func TestFrameBufferIdleRenderEmitsNoCursorToggles(t *testing.T) {
	var buf bytes.Buffer
	fb := NewFrameBuffer(&buf, 5, 2)
	_ = fb.Render("aa\nbb")
	buf.Reset()

	out := fb.Render("aa\nbb")
	if out != "" {
		t.Fatalf("expected empty diff output for identical frame, got %q", out)
	}
	if strings.Contains(buf.String(), "\x1b[?25l") || strings.Contains(buf.String(), "\x1b[?25h") {
		t.Fatalf("idle render must not emit cursor hide/show: %q", buf.String())
	}
}

func TestFrameBufferResizeShrinkHeight(t *testing.T) {
	var buf bytes.Buffer
	fb := NewFrameBuffer(&buf, 10, 5)
	_ = fb.Render("r0\nr1\nr2\nr3\nr4")

	fb.Resize(10, 2)
	if fb.height != 2 {
		t.Fatalf("expected height=2 after shrink, got %d", fb.height)
	}
	if len(fb.back) != 2 {
		t.Fatalf("expected back buffer length 2, got %d", len(fb.back))
	}
	if !fb.dirty {
		t.Fatal("expected dirty=true after resize")
	}
	if fb.front != nil {
		t.Fatal("expected front buffer to be cleared after resize")
	}

	// After resize, next render should emit a full flush.
	buf.Reset()
	out := fb.Render("aa\nbb")
	if out == "" {
		t.Fatal("expected non-empty render after resize")
	}
	if !strings.Contains(out, "aa") || !strings.Contains(out, "bb") {
		t.Fatalf("expected resized frame content, got %q", out)
	}
}

func TestFrameBufferResizeGrowHeight(t *testing.T) {
	var buf bytes.Buffer
	fb := NewFrameBuffer(&buf, 10, 2)
	_ = fb.Render("r0\nr1")

	fb.Resize(10, 5)
	if fb.height != 5 {
		t.Fatalf("expected height=5 after grow, got %d", fb.height)
	}
	if len(fb.back) != 5 {
		t.Fatalf("expected back buffer length 5, got %d", len(fb.back))
	}
	if fb.front != nil {
		t.Fatal("expected front buffer to be cleared after resize")
	}

	buf.Reset()
	out := fb.Render("a\nb\nc\nd\ne")
	if out == "" {
		t.Fatal("expected non-empty render after grow")
	}
	for _, want := range []string{"a", "b", "c", "d", "e"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected grown frame to contain %q, got %q", want, out)
		}
	}
}

func TestFrameBufferClearEmitsClearEscape(t *testing.T) {
	var buf bytes.Buffer
	fb := NewFrameBuffer(&buf, 10, 3)
	_ = fb.Render("aaa\nbbb\nccc")

	buf.Reset()
	fb.Clear()

	written := buf.String()
	if !strings.Contains(written, "\x1b[2J") {
		t.Fatalf("expected Clear to emit ESC[2J (erase display), got %q", written)
	}
	if !strings.Contains(written, "\x1b[H") {
		t.Fatalf("expected Clear to emit ESC[H (cursor home), got %q", written)
	}
	if fb.front != nil {
		t.Fatal("expected front buffer to be cleared")
	}
	if !fb.dirty {
		t.Fatal("expected dirty=true after Clear")
	}
}

func TestFrameBufferCursorToClampsToOne(t *testing.T) {
	var buf bytes.Buffer
	fb := NewFrameBuffer(&buf, 10, 3)

	buf.Reset()
	fb.CursorTo(-5, -10)
	if got := buf.String(); !strings.Contains(got, "\x1b[1;1H") {
		t.Fatalf("expected CursorTo(-5,-10) to clamp to ESC[1;1H, got %q", got)
	}
	if fb.cursorRow != 1 || fb.cursorCol != 1 {
		t.Fatalf("expected cursor (1,1) after clamping, got (%d,%d)", fb.cursorRow, fb.cursorCol)
	}

	buf.Reset()
	fb.CursorTo(0, 0)
	if got := buf.String(); !strings.Contains(got, "\x1b[1;1H") {
		t.Fatalf("expected CursorTo(0,0) to clamp to ESC[1;1H, got %q", got)
	}

	buf.Reset()
	fb.CursorTo(3, 7)
	if got := buf.String(); !strings.Contains(got, "\x1b[3;7H") {
		t.Fatalf("expected CursorTo(3,7) to emit ESC[3;7H, got %q", got)
	}
	if fb.cursorRow != 3 || fb.cursorCol != 7 {
		t.Fatalf("expected cursor (3,7), got (%d,%d)", fb.cursorRow, fb.cursorCol)
	}
}

func TestFrameBufferHideShowCursorEscapes(t *testing.T) {
	var buf bytes.Buffer
	fb := NewFrameBuffer(&buf, 10, 3)

	buf.Reset()
	fb.HideCursor()
	if got := buf.String(); got != "\x1b[?25l" {
		t.Fatalf("expected HideCursor to emit ESC[?25l, got %q", got)
	}

	buf.Reset()
	fb.ShowCursor()
	if got := buf.String(); got != "\x1b[?25h" {
		t.Fatalf("expected ShowCursor to emit ESC[?25h, got %q", got)
	}
}
