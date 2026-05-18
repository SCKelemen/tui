package display

import (
	"strings"
	"testing"

	"github.com/SCKelemen/tui/v2/style"
)

// stripAllANSI removes ANSI escape sequences so width and content
// assertions can be made against the visible portion of the gutter only.
// A test-local helper avoids colliding with the package-level stripANSI
// used elsewhere in the display package.
func stripAllANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) {
				c := s[j]
				if c >= '@' && c <= '~' {
					j++
					break
				}
				j++
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func TestLineNumbersDefaults(t *testing.T) {
	l := NewLineNumbers()
	if got, want := l.Width(), 2; got != want {
		t.Fatalf("Width() = %d, want %d (1 digit + 1 separator)", got, want)
	}
	out := stripAllANSI(l.Render())
	if out != "1 " {
		t.Fatalf("Render() = %q, want %q", out, "1 ")
	}
}

func TestLineNumbersWidthScalesWithDigitCount(t *testing.T) {
	cases := []struct {
		name      string
		start     int
		count     int
		wantWidth int
	}{
		{"single digit", 1, 9, 2},     // 1 digit + sep
		{"two digits", 1, 99, 3},      // 2 digits + sep
		{"three digits", 1, 999, 4},   // 3 digits + sep
		{"four digits", 1, 1000, 5},   // 4 digits + sep
		{"offset into 10s", 95, 10, 4}, // last = 104, 3 digits + sep
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l := NewLineNumbers(WithLineNumbersStart(c.start), WithLineNumbersCount(c.count))
			if got := l.Width(); got != c.wantWidth {
				t.Fatalf("Width() = %d, want %d", got, c.wantWidth)
			}
		})
	}
}

func TestLineNumbersAbsoluteRenderShape(t *testing.T) {
	l := NewLineNumbers(WithLineNumbersStart(8), WithLineNumbersCount(4))
	// last line = 11, so number column is 2 wide; total width = 3.
	out := stripAllANSI(l.Render())
	want := strings.Join([]string{
		" 8 ",
		" 9 ",
		"10 ",
		"11 ",
	}, "\n")
	if out != want {
		t.Fatalf("Render() = %q, want %q", out, want)
	}
}

func TestLineNumbersRelativeMode(t *testing.T) {
	l := NewLineNumbers(
		WithLineNumbersStart(8),
		WithLineNumbersCount(5),
		WithLineNumbersRelative(true),
		WithLineNumbersCurrentLine(10),
	)
	// width = 2 (10 is the highest absolute number). Current line keeps
	// its absolute number; other lines show distance from 10.
	out := stripAllANSI(l.Render())
	want := strings.Join([]string{
		" 2 ",
		" 1 ",
		"10 ",
		" 1 ",
		" 2 ",
	}, "\n")
	if out != want {
		t.Fatalf("Render() = %q, want %q", out, want)
	}
}

func TestLineNumbersCurrentLineBoldOthersDim(t *testing.T) {
	l := NewLineNumbers(
		WithLineNumbersStart(1),
		WithLineNumbersCount(3),
		WithLineNumbersCurrentLine(2),
	)
	rows := strings.Split(l.Render(), "\n")
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	if !strings.Contains(rows[1], style.ANSIBold) {
		t.Errorf("current row missing ANSIBold: %q", rows[1])
	}
	if strings.Contains(rows[0], style.ANSIBold) || strings.Contains(rows[2], style.ANSIBold) {
		t.Errorf("non-current rows should not contain ANSIBold")
	}
	for _, r := range []string{rows[0], rows[2]} {
		if !strings.Contains(r, style.ANSIDim) {
			t.Errorf("non-current row missing ANSIDim: %q", r)
		}
	}
	for _, r := range rows {
		if !strings.HasSuffix(r, style.ANSIReset) {
			t.Errorf("row missing trailing ANSIReset: %q", r)
		}
	}
}

func TestLineNumbersSignsReplaceExpectedPosition(t *testing.T) {
	signs := map[int]string{2: "●"}
	l := NewLineNumbers(
		WithLineNumbersStart(1),
		WithLineNumbersCount(3),
		WithLineNumbersSigns(signs),
	)
	// width = 1 digit + 1 sign + 1 sep = 3
	if got, want := l.Width(), 3; got != want {
		t.Fatalf("Width() = %d, want %d", got, want)
	}
	out := stripAllANSI(l.Render())
	want := strings.Join([]string{
		"1  ",
		"2● ",
		"3  ",
	}, "\n")
	if out != want {
		t.Fatalf("Render() = %q, want %q", out, want)
	}
}

func TestLineNumbersRenderLineMatchesRender(t *testing.T) {
	l := NewLineNumbers(WithLineNumbersStart(5), WithLineNumbersCount(4), WithLineNumbersCurrentLine(7))
	full := l.Render()
	rows := strings.Split(full, "\n")
	for i := 0; i < 4; i++ {
		got := l.RenderLine(5 + i)
		if got != rows[i] {
			t.Errorf("RenderLine(%d) = %q, want %q", 5+i, got, rows[i])
		}
	}
}

func TestLineNumbersMinWidthOptionExpandsButDoesNotShrink(t *testing.T) {
	// Smaller minWidth than required digits → digits win.
	a := NewLineNumbers(WithLineNumbersStart(1), WithLineNumbersCount(100), WithLineNumbersWidth(1))
	if got, want := a.Width(), 4; got != want { // 3 digits + sep
		t.Fatalf("digits-win Width() = %d, want %d", got, want)
	}
	// Larger minWidth than required digits → minWidth wins.
	b := NewLineNumbers(WithLineNumbersStart(1), WithLineNumbersCount(5), WithLineNumbersWidth(4))
	if got, want := b.Width(), 5; got != want { // 4-wide column + sep
		t.Fatalf("minWidth-wins Width() = %d, want %d", got, want)
	}
}

func TestLineNumbersSetCurrentLineUpdatesBoldRow(t *testing.T) {
	l := NewLineNumbers(WithLineNumbersStart(1), WithLineNumbersCount(3))
	if strings.Contains(l.Render(), style.ANSIBold) {
		t.Fatalf("no current line should mean no bold rows")
	}
	l.SetCurrentLine(3)
	rows := strings.Split(l.Render(), "\n")
	if !strings.Contains(rows[2], style.ANSIBold) {
		t.Errorf("expected last row to be bold after SetCurrentLine(3): %q", rows[2])
	}
}

func TestLineNumbersStartAndCountClampToOne(t *testing.T) {
	l := NewLineNumbers(WithLineNumbersStart(0), WithLineNumbersCount(-5))
	if got, want := stripAllANSI(l.Render()), "1 "; got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}
