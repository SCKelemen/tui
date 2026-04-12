package tui

import (
	"reflect"
	"testing"
)

func TestGraphemeIterator(t *testing.T) {
	it := NewGraphemeIterator("A界e\u0301🙂")

	cluster, ok := it.Next()
	if !ok || cluster.String() != "A" || cluster.Width != 1 {
		t.Fatalf("first cluster = (%q,%d,%v), want (A,1,true)", cluster.String(), cluster.Width, ok)
	}
	if got := it.Remaining(); got != "界e\u0301🙂" {
		t.Fatalf("Remaining() after first Next = %q, want %q", got, "界e\u0301🙂")
	}

	var got []string
	var widths []int
	got = append(got, cluster.String())
	widths = append(widths, cluster.Width)
	for {
		cluster, ok = it.Next()
		if !ok {
			break
		}
		got = append(got, cluster.String())
		widths = append(widths, cluster.Width)
	}

	wantClusters := []string{"A", "界", "e\u0301", "🙂"}
	wantWidths := []int{1, 2, 1, 2}
	if !reflect.DeepEqual(got, wantClusters) {
		t.Fatalf("clusters = %#v, want %#v", got, wantClusters)
	}
	if !reflect.DeepEqual(widths, wantWidths) {
		t.Fatalf("widths = %#v, want %#v", widths, wantWidths)
	}
	if got := it.Remaining(); got != "" {
		t.Fatalf("Remaining() at end = %q, want empty", got)
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "ASCII", input: "abc", want: 3},
		{name: "CJK", input: "界", want: 2},
		{name: "emoji", input: "🙂", want: 2},
		{name: "combining", input: "e\u0301", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.input); got != tt.want {
				t.Fatalf("StringWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("A界B", 3); got != "A界" {
		t.Fatalf("Truncate(A界B,3) = %q, want %q", got, "A界")
	}
	if got := Truncate("e\u0301x", 1); got != "e\u0301" {
		t.Fatalf("Truncate(e◌́x,1) = %q, want %q", got, "e\u0301")
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	if got := TruncateWithEllipsis("hello", 4); got != "hel…" {
		t.Fatalf("TruncateWithEllipsis(hello,4) = %q, want %q", got, "hel…")
	}
	if got := TruncateWithEllipsis("hello", 1); got != "…" {
		t.Fatalf("TruncateWithEllipsis(hello,1) = %q, want %q", got, "…")
	}
}

func TestPadRight(t *testing.T) {
	if got := PadRight("界", 4); got != "界  " {
		t.Fatalf("PadRight(界,4) = %q, want %q", got, "界  ")
	}
}

func TestPadLeft(t *testing.T) {
	if got := PadLeft("界", 4); got != "  界" {
		t.Fatalf("PadLeft(界,4) = %q, want %q", got, "  界")
	}
}

func TestPadCenter(t *testing.T) {
	if got := PadCenter("界", 5); got != " 界  " {
		t.Fatalf("PadCenter(界,5) = %q, want %q", got, " 界  ")
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  []string
	}{
		{
			name:  "wrap words",
			input: "hello world",
			width: 5,
			want:  []string{"hello", "world"},
		},
		{
			name:  "split long token",
			input: "abcdef",
			width: 4,
			want:  []string{"abcd", "ef"},
		},
		{
			name:  "respect newlines",
			input: "hi\nthere",
			width: 10,
			want:  []string{"hi", "there"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WrapText(tt.input, tt.width); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("WrapText(%q,%d) = %#v, want %#v", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

func TestIsWide(t *testing.T) {
	if !IsWide('界') {
		t.Fatal("IsWide('界') = false, want true")
	}
	if IsWide('A') {
		t.Fatal("IsWide('A') = true, want false")
	}
}

func TestIsCombining(t *testing.T) {
	if !IsCombining('\u0301') {
		t.Fatal("IsCombining(U+0301) = false, want true")
	}
	if IsCombining('A') {
		t.Fatal("IsCombining('A') = true, want false")
	}
}
