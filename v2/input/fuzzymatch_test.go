package input

import (
	"reflect"
	"strings"
	"testing"

	"github.com/SCKelemen/tui/v2/style"
)

func TestFuzzyMatcherBasic(t *testing.T) {
	t.Run("empty pattern matches everything with score 0", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		match := matcher.Match("", "anything")

		if !match.Matched {
			t.Fatal("expected empty pattern to match")
		}
		if match.Score != 0 {
			t.Fatalf("expected score 0 for empty pattern, got %d", match.Score)
		}
	})

	t.Run("empty candidate with non-empty pattern has no match", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		match := matcher.Match("abc", "")

		if match.Matched {
			t.Fatal("expected no match for non-empty pattern against empty candidate")
		}
	})

	t.Run("exact match scores highest", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		exact := matcher.Match("test", "test")
		inexact := matcher.Match("test", "te-st")

		if !exact.Matched || !inexact.Matched {
			t.Fatal("expected both candidates to match")
		}
		if exact.Score <= inexact.Score {
			t.Fatalf("expected exact match score (%d) > inexact score (%d)", exact.Score, inexact.Score)
		}
	})

	t.Run("prefix match scores higher than middle match", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		prefix := matcher.Match("sp", "Spinners")
		middle := matcher.Match("sp", "Display")

		if !prefix.Matched || !middle.Matched {
			t.Fatal("expected both prefix and middle candidates to match")
		}
		if prefix.Score <= middle.Score {
			t.Fatalf("expected prefix score (%d) > middle score (%d)", prefix.Score, middle.Score)
		}
	})

	t.Run("no match returns Matched=false", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		match := matcher.Match("xyz", "abc")

		if match.Matched {
			t.Fatal("expected no match")
		}
	})
}

func TestFuzzyMatcherSmartCase(t *testing.T) {
	t.Run("all-lowercase pattern uses case-insensitive matching", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		match := matcher.Match("cmd", "CoMmAnD")

		if !match.Matched {
			t.Fatal("expected case-insensitive match for lowercase pattern")
		}
	})

	t.Run("pattern with uppercase uses case-sensitive matching", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		match := matcher.Match("Cmd", "command")

		if match.Matched {
			t.Fatal("expected case-sensitive smart-case mismatch")
		}
	})

	t.Run("WithFuzzyCaseSensitive(true) forces case-sensitive matching", func(t *testing.T) {
		matcher := NewFuzzyMatcher(WithFuzzyCaseSensitive(true))
		match := matcher.Match("cmd", "COMMAND")

		if match.Matched {
			t.Fatal("expected forced case-sensitive mismatch")
		}
	})
}

func TestFuzzyMatcherScoring(t *testing.T) {
	t.Run("consecutive matches score higher than scattered matches", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		consecutive := matcher.Match("abc", "abcxyz")
		scattered := matcher.Match("abc", "axbycz")

		if !consecutive.Matched || !scattered.Matched {
			t.Fatal("expected both candidates to match")
		}
		if consecutive.Score <= scattered.Score {
			t.Fatalf("expected consecutive score (%d) > scattered score (%d)", consecutive.Score, scattered.Score)
		}
	})

	t.Run("word boundary matches score higher than mid-word matches", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		boundary := matcher.Match("go", "go-lang")
		midWord := matcher.Match("go", "algorithm")

		if !boundary.Matched || !midWord.Matched {
			t.Fatal("expected both candidates to match")
		}
		if boundary.Score <= midWord.Score {
			t.Fatalf("expected boundary score (%d) > mid-word score (%d)", boundary.Score, midWord.Score)
		}
	})

	t.Run("camelCase boundary matches get bonuses", func(t *testing.T) {
		matcher := NewFuzzyMatcher(WithFuzzyCaseSensitive(true))
		camel := matcher.Match("CP", "CommandPalette")
		nonCamel := matcher.Match("CP", "CommandXPalette")

		if !camel.Matched || !nonCamel.Matched {
			t.Fatal("expected both candidates to match")
		}
		if camel.Score <= nonCamel.Score {
			t.Fatalf("expected camelCase score (%d) > non-camel score (%d)", camel.Score, nonCamel.Score)
		}
	})

	t.Run("first character match gets bonus", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		startsFirst := matcher.Match("sp", "spinners")
		notFirst := matcher.Match("sp", "xspinners")

		if !startsFirst.Matched || !notFirst.Matched {
			t.Fatal("expected both candidates to match")
		}
		if startsFirst.Score <= notFirst.Score {
			t.Fatalf("expected first-char score (%d) > non-first score (%d)", startsFirst.Score, notFirst.Score)
		}
	})

	t.Run("gap penalties reduce score", func(t *testing.T) {
		matcher := NewFuzzyMatcher()
		smallGap := matcher.Match("abc", "abxc")
		largeGap := matcher.Match("abc", "abxxxxc")

		if !smallGap.Matched || !largeGap.Matched {
			t.Fatal("expected both candidates to match")
		}
		if smallGap.Score <= largeGap.Score {
			t.Fatalf("expected small-gap score (%d) > large-gap score (%d)", smallGap.Score, largeGap.Score)
		}
	})
}

func TestFuzzyMatcherPositions(t *testing.T) {
	matcher := NewFuzzyMatcher()
	match := matcher.Match("cpt", "CommandPaletteTest")
	if !match.Matched {
		t.Fatal("expected match")
	}

	t.Run("positions contain correct indices", func(t *testing.T) {
		expected := []int{0, 7, 11}
		if !reflect.DeepEqual(match.Positions, expected) {
			t.Fatalf("expected positions %v, got %v", expected, match.Positions)
		}
	})

	t.Run("positions length equals pattern length", func(t *testing.T) {
		if len(match.Positions) != len([]rune("cpt")) {
			t.Fatalf("expected %d positions, got %d", len([]rune("cpt")), len(match.Positions))
		}
	})

	t.Run("positions are in ascending order", func(t *testing.T) {
		for i := 1; i < len(match.Positions); i++ {
			if match.Positions[i-1] >= match.Positions[i] {
				t.Fatalf("positions not strictly ascending: %v", match.Positions)
			}
		}
	})
}

func TestFuzzyMatcherRankMatches(t *testing.T) {
	matcher := NewFuzzyMatcher()

	t.Run("returns only matching candidates", func(t *testing.T) {
		candidates := []string{"alpha", "beta", "alp", "zzz"}
		results := matcher.RankMatches("alp", candidates)

		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
		for _, r := range results {
			if !r.Matched {
				t.Fatalf("expected all ranked results to be matched, got %+v", r)
			}
		}
	})

	t.Run("results are sorted by score descending", func(t *testing.T) {
		candidates := []string{"alpha", "alp", "a-l-p"}
		results := matcher.RankMatches("alp", candidates)
		if len(results) < 2 {
			t.Fatalf("expected at least 2 ranked results, got %d", len(results))
		}
		for i := 1; i < len(results); i++ {
			if results[i-1].Score < results[i].Score {
				t.Fatalf("scores not descending at %d: %d < %d", i, results[i-1].Score, results[i].Score)
			}
		}
	})

	t.Run("empty pattern returns all candidates", func(t *testing.T) {
		candidates := []string{"alpha", "beta", "gamma"}
		results := matcher.RankMatches("", candidates)
		if len(results) != len(candidates) {
			t.Fatalf("expected %d results, got %d", len(candidates), len(results))
		}
	})

	t.Run("non-matching candidates are excluded", func(t *testing.T) {
		candidates := []string{"alpha", "beta", "gamma"}
		results := matcher.RankMatches("zzz", candidates)
		if len(results) != 0 {
			t.Fatalf("expected no matches, got %d", len(results))
		}
	})
}

func TestHighlightMatch(t *testing.T) {
	t.Run("matched characters are wrapped with ANSI", func(t *testing.T) {
		m := FuzzyMatch{Matched: true, Candidate: "abc", Positions: []int{0, 2}}
		color := "#ffaa00"
		got := HighlightMatch(m, color)

		highlight := style.Fg(color) + style.ANSIBold
		if !strings.Contains(got, highlight+"a"+style.ANSIReset) {
			t.Fatalf("expected highlighted 'a' in output: %q", got)
		}
		if !strings.Contains(got, highlight+"c"+style.ANSIReset) {
			t.Fatalf("expected highlighted 'c' in output: %q", got)
		}
	})

	t.Run("non-matched characters are unchanged", func(t *testing.T) {
		m := FuzzyMatch{Matched: true, Candidate: "abc", Positions: []int{0, 2}}
		got := HighlightMatch(m, "#00ff00")
		if !strings.Contains(got, "b") {
			t.Fatalf("expected non-matched character to remain unchanged: %q", got)
		}
	})

	t.Run("empty positions returns original string", func(t *testing.T) {
		m := FuzzyMatch{Matched: true, Candidate: "plain", Positions: nil}
		got := HighlightMatch(m, "#ff00ff")
		if got != "plain" {
			t.Fatalf("expected original string, got %q", got)
		}
	})
}

func TestFuzzyMatcherRealWorldScenarios(t *testing.T) {
	matcher := NewFuzzyMatcher()

	t.Run("cp matches CommandPalette with camelCase boundary bonus", func(t *testing.T) {
		csMatcher := NewFuzzyMatcher(WithFuzzyCaseSensitive(true))
		camel := csMatcher.Match("CP", "CommandPalette")
		nonCamel := csMatcher.Match("CP", "CommandXPalette")

		if !camel.Matched {
			t.Fatal("expected 'CP' to match 'CommandPalette'")
		}
		if !nonCamel.Matched {
			t.Fatal("expected 'CP' to match 'CommandXPalette'")
		}
		if camel.Score <= nonCamel.Score {
			t.Fatalf("expected camelCase score (%d) > non-camel score (%d)", camel.Score, nonCamel.Score)
		}
	})

	t.Run("fzf matches FuzzyFinder better than freezing", func(t *testing.T) {
		ff := matcher.Match("fzf", "FuzzyFinder")
		fr := matcher.Match("fzf", "freezing")

		if !ff.Matched {
			t.Fatal("expected 'fzf' to match 'FuzzyFinder'")
		}
		if fr.Matched && ff.Score <= fr.Score {
			t.Fatalf("expected FuzzyFinder score (%d) > freezing score (%d)", ff.Score, fr.Score)
		}
	})

	t.Run("git matches Git & Diffs better than Diagnostics", func(t *testing.T) {
		good := matcher.Match("git", "Git & Diffs")
		weak := matcher.Match("git", "Diagnostics")

		if !good.Matched {
			t.Fatal("expected 'git' to match 'Git & Diffs'")
		}
		if weak.Matched && good.Score <= weak.Score {
			t.Fatalf("expected Git & Diffs score (%d) > Diagnostics score (%d)", good.Score, weak.Score)
		}
	})

	t.Run("sp matches Spinners better than Display", func(t *testing.T) {
		prefix := matcher.Match("sp", "Spinners")
		mid := matcher.Match("sp", "Display")

		if !prefix.Matched || !mid.Matched {
			t.Fatal("expected both candidates to match")
		}
		if prefix.Score <= mid.Score {
			t.Fatalf("expected Spinners score (%d) > Display score (%d)", prefix.Score, mid.Score)
		}
	})

	t.Run("path-like matching hg matches handler.go at boundary", func(t *testing.T) {
		match := matcher.Match("hg", "handler.go")
		if !match.Matched {
			t.Fatal("expected 'hg' to match 'handler.go'")
		}
		if len(match.Positions) != 2 {
			t.Fatalf("expected 2 positions, got %d", len(match.Positions))
		}
		if match.Positions[1] <= match.Positions[0] {
			t.Fatalf("expected ascending positions, got %v", match.Positions)
		}
	})
}
