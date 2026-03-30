package input

import (
	"sort"
	"strings"
	"unicode"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"golang.org/x/text/unicode/norm"
)

const (
	scoreMatch       = 16
	bonusConsecutive = 8
	bonusBoundary    = 8
	bonusCamelCase   = 7
	bonusFirstChar   = 10
	penaltyGap       = -3
	penaltyGapStart  = -5
	penaltyMaxGap    = -9
)

// FuzzyMatch is a scored fuzzy match result for a single candidate string.
type FuzzyMatch struct {
	Score     int
	Matched   bool
	Positions []int
	Candidate string
}

// FuzzyMatcherOption configures a FuzzyMatcher.
type FuzzyMatcherOption func(*FuzzyMatcher)

// FuzzyMatcher provides fzf-style fuzzy matching and ranking.
type FuzzyMatcher struct {
	caseSensitive bool
	normalize     bool
}

// Reference tui package explicitly per package conventions in this module.
var _ tui.Component

// NewFuzzyMatcher creates a matcher with optional behavior overrides.
func NewFuzzyMatcher(opts ...FuzzyMatcherOption) *FuzzyMatcher {
	m := &FuzzyMatcher{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// WithFuzzyCaseSensitive forces case-sensitive matching when true.
func WithFuzzyCaseSensitive(v bool) FuzzyMatcherOption {
	return func(m *FuzzyMatcher) {
		m.caseSensitive = v
	}
}

// WithFuzzyNormalize strips accents/diacritics when true.
func WithFuzzyNormalize(v bool) FuzzyMatcherOption {
	return func(m *FuzzyMatcher) {
		m.normalize = v
	}
}

// Match scores a single candidate against a pattern.
func (m *FuzzyMatcher) Match(pattern, candidate string) FuzzyMatch {
	result := FuzzyMatch{Candidate: candidate}

	if pattern == "" {
		result.Matched = true
		return result
	}
	if candidate == "" {
		return result
	}

	patternRunes := []rune(pattern)
	candidateRunes, candidateMap := []rune(candidate), make([]int, 0)
	if m.normalize {
		patternRunes, _ = normalizeRunesWithMap(pattern)
		candidateRunes, candidateMap = normalizeRunesWithMap(candidate)
	} else {
		candidateMap = make([]int, len(candidateRunes))
		for i := range candidateRunes {
			candidateMap[i] = i
		}
	}

	if len(patternRunes) == 0 {
		result.Matched = true
		return result
	}
	if len(candidateRunes) == 0 {
		return result
	}

	caseSensitive := m.smartCaseSensitive(patternRunes)
	if !caseSensitive {
		patternRunes = toLowerRunes(patternRunes)
		candidateRunes = toLowerRunes(candidateRunes)
	}

	positions, score, ok := scoreOptimalMatch(patternRunes, candidateRunes)
	if !ok {
		return result
	}

	result.Matched = true
	result.Score = score
	result.Positions = make([]int, len(positions))
	for i, p := range positions {
		if p >= 0 && p < len(candidateMap) {
			result.Positions[i] = candidateMap[p]
		}
	}
	return result
}

// RankMatches scores and sorts multiple candidates, returning only matches.
func (m *FuzzyMatcher) RankMatches(pattern string, candidates []string) []FuzzyMatch {
	matches := make([]FuzzyMatch, 0, len(candidates))
	for _, candidate := range candidates {
		match := m.Match(pattern, candidate)
		if match.Matched {
			matches = append(matches, match)
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		if len(matches[i].Positions) != len(matches[j].Positions) {
			return len(matches[i].Positions) < len(matches[j].Positions)
		}
		return strings.ToLower(matches[i].Candidate) < strings.ToLower(matches[j].Candidate)
	})

	return matches
}

// HighlightMatch returns candidate text with matched runes ANSI-highlighted.
func HighlightMatch(match FuzzyMatch, highlightColor string) string {
	if !match.Matched || len(match.Positions) == 0 {
		return match.Candidate
	}

	highlight := style.Fg(highlightColor) + style.ANSIBold
	posSet := make(map[int]struct{}, len(match.Positions))
	for _, p := range match.Positions {
		posSet[p] = struct{}{}
	}

	var b strings.Builder
	for i, r := range []rune(match.Candidate) {
		if _, ok := posSet[i]; ok {
			b.WriteString(highlight)
			b.WriteRune(r)
			b.WriteString(style.ANSIReset)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (m *FuzzyMatcher) smartCaseSensitive(pattern []rune) bool {
	if m.caseSensitive {
		return true
	}
	for _, r := range pattern {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func scoreOptimalMatch(pattern, candidate []rune) ([]int, int, bool) {
	m, n := len(pattern), len(candidate)
	if m == 0 {
		return []int{}, 0, true
	}
	if m > n {
		return nil, 0, false
	}

	const negInf = -1 << 30

	dp := make([][]int, m)
	prev := make([][]int, m)
	for i := 0; i < m; i++ {
		dp[i] = make([]int, n)
		prev[i] = make([]int, n)
		for j := 0; j < n; j++ {
			dp[i][j] = negInf
			prev[i][j] = -1
		}
	}

	bonuses := make([]int, n)
	for j := 0; j < n; j++ {
		if isBoundary(string(candidate), j) {
			bonuses[j] += bonusBoundary
		}
		if isCamelCaseTransition(string(candidate), j) {
			bonuses[j] += bonusCamelCase
		}
		if j == 0 {
			bonuses[j] += bonusFirstChar
		}
	}

	for j := 0; j < n; j++ {
		if pattern[0] == candidate[j] {
			dp[0][j] = scoreMatch + bonuses[j]
		}
	}

	for i := 1; i < m; i++ {
		for j := 0; j < n; j++ {
			if pattern[i] != candidate[j] {
				continue
			}

			bestScore := negInf
			bestPrev := -1
			for k := 0; k < j; k++ {
				if dp[i-1][k] == negInf {
					continue
				}

				s := dp[i-1][k] + scoreMatch + bonuses[j]
				gap := j - k - 1
				if gap == 0 {
					s += bonusConsecutive
				} else {
					s += gapPenalty(gap)
				}

				if s > bestScore || (s == bestScore && isBetterTransition(k, bestPrev, j, candidate)) {
					bestScore = s
					bestPrev = k
				}
			}

			dp[i][j] = bestScore
			prev[i][j] = bestPrev
		}
	}

	bestEnd := -1
	bestScore := negInf
	for j := 0; j < n; j++ {
		if dp[m-1][j] > bestScore {
			bestScore = dp[m-1][j]
			bestEnd = j
		}
	}
	if bestEnd == -1 || bestScore == negInf {
		return nil, 0, false
	}

	positions := make([]int, m)
	j := bestEnd
	for i := m - 1; i >= 0; i-- {
		positions[i] = j
		j = prev[i][j]
	}

	return positions, bestScore, true
}

func isBetterTransition(newPrev, oldPrev, current int, _ []rune) bool {
	if oldPrev == -1 {
		return true
	}

	newGap := current - newPrev - 1
	oldGap := current - oldPrev - 1
	if newGap != oldGap {
		return newGap < oldGap
	}

	return newPrev < oldPrev
}
func gapPenalty(gap int) int {
	if gap <= 0 {
		return 0
	}
	p := penaltyGapStart + (gap-1)*penaltyGap
	if p < penaltyMaxGap {
		return penaltyMaxGap
	}
	return p
}

func normalizeRunesWithMap(s string) ([]rune, []int) {
	in := []rune(s)
	out := make([]rune, 0, len(in))
	mapping := make([]int, 0, len(in))

	for i, r := range in {
		decomposed := norm.NFD.String(string(r))
		for _, dr := range decomposed {
			if unicode.Is(unicode.Mn, dr) {
				continue
			}
			out = append(out, dr)
			mapping = append(mapping, i)
		}
	}

	return out, mapping
}

func toLowerRunes(in []rune) []rune {
	out := make([]rune, len(in))
	for i, r := range in {
		out[i] = unicode.ToLower(r)
	}
	return out
}

func isBoundary(candidate string, pos int) bool {
	runes := []rune(candidate)
	return boundaryAtRunes(runes, pos)
}

func boundaryAtRunes(runes []rune, pos int) bool {
	if pos < 0 || pos >= len(runes) {
		return false
	}
	if pos == 0 {
		return true
	}
	prev := runes[pos-1]
	curr := runes[pos]

	if prev == ' ' || prev == '_' || prev == '-' || prev == '.' || prev == '/' {
		return true
	}
	return !isWordChar(prev) && isWordChar(curr)
}

func isCamelCaseTransition(candidate string, pos int) bool {
	runes := []rune(candidate)
	if pos <= 0 || pos >= len(runes) {
		return false
	}
	prev := runes[pos-1]
	curr := runes[pos]
	return unicode.IsLower(prev) && unicode.IsUpper(curr)
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
