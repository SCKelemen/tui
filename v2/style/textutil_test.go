package style

import (
	"strings"
	"testing"
)

func TestStringWidth(t *testing.T) {
	if got := StringWidth("hello"); got != 5 {
		t.Fatalf("StringWidth(ASCII) = %d, want 5", got)
	}

	if got := StringWidth("你好"); got != 4 {
		t.Fatalf("StringWidth(CJK) = %d, want 4", got)
	}

	if got := StringWidth("🙂"); got != 2 {
		t.Fatalf("StringWidth(emoji) = %d, want 2", got)
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("hello world", 5, "…"); got != "hell…" {
		t.Fatalf("Truncate(normal) = %q, want %q", got, "hell…")
	}

	if got := Truncate("你好世界", 5, "…"); got != "你好…" {
		t.Fatalf("Truncate(CJK) = %q, want %q", got, "你好…")
	}

	if got := Truncate("hello", 5, "…"); got != "hello" {
		t.Fatalf("Truncate(no-op when fits) = %q, want %q", got, "hello")
	}

	if got := Truncate("", 5, "…"); got != "" {
		t.Fatalf("Truncate(empty) = %q, want empty string", got)
	}

	if got := Truncate("hello", 0, "…"); got != "" {
		t.Fatalf("Truncate(zero width) = %q, want empty string", got)
	}
}

func TestTruncateMiddle(t *testing.T) {
	got := TruncateMiddle("abcdefghij", 7, "…")
	if !strings.HasPrefix(got, "abc") {
		t.Fatalf("TruncateMiddle should preserve start, got %q", got)
	}
	if !strings.HasSuffix(got, "hij") {
		t.Fatalf("TruncateMiddle should preserve end, got %q", got)
	}
	if !strings.Contains(got, "…") {
		t.Fatalf("TruncateMiddle should include ellipsis, got %q", got)
	}
	if w := StringWidth(got); w > 7 {
		t.Fatalf("TruncateMiddle result width = %d, want <= 7", w)
	}
}

func TestElidePath(t *testing.T) {
	path := "/Users/sam/code/SCKelemen/SCK-Stack/stack/tui/v2/style/textutil.go"
	got := ElidePath(path, 24)

	if got == "" {
		t.Fatal("ElidePath returned empty string")
	}
	if got == path {
		t.Fatalf("ElidePath should shorten long path, got unchanged %q", got)
	}
	if w := StringWidth(got); w > 24 {
		t.Fatalf("ElidePath result width = %d, want <= 24 (value %q)", w, got)
	}
}

func TestElideURL(t *testing.T) {
	url := "https://docs.lovable.dev/reference/api/v0.5/embeds?source=security"
	got := ElideURL(url, 30)

	if got == "" {
		t.Fatal("ElideURL returned empty string")
	}
	if got == url {
		t.Fatalf("ElideURL should shorten long URL, got unchanged %q", got)
	}
	if w := StringWidth(got); w > 30 {
		t.Fatalf("ElideURL result width = %d, want <= 30 (value %q)", w, got)
	}
}

func TestPad(t *testing.T) {
	got := Pad("abc", 5)
	if got != "abc  " {
		t.Fatalf("Pad("+"%q"+") = %q, want %q", "abc", got, "abc  ")
	}
	if w := StringWidth(got); w != 5 {
		t.Fatalf("Pad width = %d, want 5", w)
	}

	if unchanged := Pad("abcdef", 3); unchanged != "abcdef" {
		t.Fatalf("Pad should return original string when already wide enough, got %q", unchanged)
	}
}
