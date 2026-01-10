package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHeaderSingleColumn(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 100, Align: AlignCenter, Content: []string{"Only column"}},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty with single column")
	}

	if !strings.Contains(view, "Only column") {
		t.Error("View should contain column content")
	}
}

func TestHeaderColumnWidthsNotHundred(t *testing.T) {
	// Columns that don't add up to 100%
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 30, Align: AlignLeft, Content: []string{"Left"}},
			HeaderColumn{Width: 30, Align: AlignRight, Content: []string{"Right"}},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Should not panic
	view := header.View()
	if view == "" {
		t.Error("View should not be empty even with non-100% widths")
	}
}

func TestHeaderColumnWidthsOverHundred(t *testing.T) {
	// Columns that add up to more than 100%
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 80, Align: AlignLeft, Content: []string{"Left"}},
			HeaderColumn{Width: 80, Align: AlignRight, Content: []string{"Right"}},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Should not panic (might overflow or truncate)
	view := header.View()
	_ = view
}

func TestHeaderVeryLongContent(t *testing.T) {
	longContent := make([]string, 100)
	for i := range longContent {
		longContent[i] = strings.Repeat("Long line ", 20)
	}

	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft, Content: longContent},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Should not panic with very long content
	view := header.View()
	if view == "" {
		t.Error("View should not be empty with long content")
	}
}

func TestHeaderEmptyContent(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft, Content: []string{}},
			HeaderColumn{Width: 50, Align: AlignRight, Content: []string{}},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty even with empty content")
	}

	// Should still have borders
	if !strings.Contains(view, "╭") || !strings.Contains(view, "╯") {
		t.Error("Header should have borders even with empty content")
	}
}

func TestHeaderManySections(t *testing.T) {
	// Create many sections
	sections := make([]HeaderSection, 20)
	for i := range sections {
		sections[i] = HeaderSection{
			Title:   "Section",
			Content: []string{"Line 1", "Line 2"},
			Divider: i%2 == 0,
		}
	}

	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft},
			HeaderColumn{Width: 50, Align: AlignLeft},
		),
		WithColumnSections(1, sections...),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Should not panic with many sections
	view := header.View()
	if view == "" {
		t.Error("View should not be empty with many sections")
	}
}

func TestHeaderVeryNarrowWidth(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft, Content: []string{"Test"}},
			HeaderColumn{Width: 50, Align: AlignRight, Content: []string{"Test"}},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 20, Height: 24})

	// Should not panic with narrow width
	view := header.View()
	_ = view
}

func TestHeaderVeryWideWidth(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft, Content: []string{"Test"}},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 300, Height: 24})

	// Should not panic with wide width
	view := header.View()
	if view == "" {
		t.Error("View should not be empty with wide width")
	}
}

func TestHeaderUnicodeContent(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{
				Width: 50,
				Align: AlignCenter,
				Content: []string{
					"日本語",
					"한국어",
					"中文",
					"العربية",
				},
			},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty with unicode content")
	}

	// Should contain unicode characters
	if !strings.Contains(view, "日本語") {
		t.Error("View should contain unicode content")
	}
}

func TestHeaderEmojiContent(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{
				Width: 50,
				Align: AlignCenter,
				Content: []string{
					"🎉 Welcome! 🎊",
					"🚀 Launch 🌟",
					"❤️ Love ✨",
				},
			},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty with emoji content")
	}
}

func TestHeaderSectionWithoutTitle(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft},
			HeaderColumn{Width: 50, Align: AlignLeft},
		),
		WithColumnSections(1,
			HeaderSection{
				Title:   "", // No title
				Content: []string{"Content line 1", "Content line 2"},
			},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty with section without title")
	}
}

func TestHeaderSectionWithoutContent(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft},
			HeaderColumn{Width: 50, Align: AlignLeft},
		),
		WithColumnSections(1,
			HeaderSection{
				Title:   "Empty Section",
				Content: []string{}, // No content
			},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty with empty section content")
	}

	if !strings.Contains(view, "Empty Section") {
		t.Error("View should contain section title")
	}
}

func TestHeaderAllAlignments(t *testing.T) {
	alignments := []ColumnAlign{AlignLeft, AlignCenter, AlignRight}

	for _, align := range alignments {
		header := NewHeader(
			WithColumns(
				HeaderColumn{
					Width:   100,
					Align:   align,
					Content: []string{"Test content"},
				},
			),
		)
		header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		view := header.View()
		if view == "" {
			t.Errorf("View should not be empty with alignment %v", align)
		}

		if !strings.Contains(view, "Test content") {
			t.Errorf("View should contain content with alignment %v", align)
		}
	}
}

func TestHeaderSectionDividers(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft},
			HeaderColumn{Width: 50, Align: AlignLeft},
		),
		WithColumnSections(1,
			HeaderSection{
				Title:   "Section 1",
				Content: []string{"Content 1"},
				Divider: true,
			},
			HeaderSection{
				Title:   "Section 2",
				Content: []string{"Content 2"},
				Divider: true,
			},
			HeaderSection{
				Title:   "Section 3",
				Content: []string{"Content 3"},
				Divider: false,
			},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty with section dividers")
	}

	// Should have horizontal dividers (─)
	if !strings.Contains(view, "─") {
		t.Error("View should contain divider characters")
	}
}

func TestHeaderWithVerticalDivider(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft, Content: []string{"Left"}},
			HeaderColumn{Width: 50, Align: AlignRight, Content: []string{"Right"}},
		),
		WithVerticalDivider(true),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty with vertical divider")
	}

	// Should have vertical dividers (│)
	verticalBarCount := strings.Count(view, "│")
	if verticalBarCount < 6 {
		t.Errorf("View should have multiple vertical bars for divider, got %d", verticalBarCount)
	}
}

func TestHeaderWithoutVerticalDivider(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft, Content: []string{"Left"}},
			HeaderColumn{Width: 50, Align: AlignRight, Content: []string{"Right"}},
		),
		WithVerticalDivider(false),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty without vertical divider")
	}

	// Should have fewer vertical bars (only borders)
	verticalBarCount := strings.Count(view, "│")
	if verticalBarCount > 10 {
		t.Errorf("View should have fewer vertical bars without divider, got %d", verticalBarCount)
	}
}

func TestHeaderThreeColumns(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 33, Align: AlignLeft, Content: []string{"Left"}},
			HeaderColumn{Width: 34, Align: AlignCenter, Content: []string{"Center"}},
			HeaderColumn{Width: 33, Align: AlignRight, Content: []string{"Right"}},
		),
		WithVerticalDivider(true),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty with three columns")
	}

	if !strings.Contains(view, "Left") || !strings.Contains(view, "Center") || !strings.Contains(view, "Right") {
		t.Error("View should contain all three column contents")
	}
}

func TestHeaderZeroWidth(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft, Content: []string{"Test"}},
		),
	)

	// Don't set width
	view := header.View()
	if view != "" {
		t.Error("View should be empty when width is not set")
	}
}

func TestHeaderAfterWindowResize(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft, Content: []string{"Test content"}},
		),
	)

	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view1 := header.View()

	header.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	view2 := header.View()

	if view1 == "" || view2 == "" {
		t.Error("Views should not be empty")
	}

	// Views should be different after resize (wider)
	if len(view1) >= len(view2) {
		t.Error("View should be wider after window resize")
	}
}

func TestHeaderSingleLineContent(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 100, Align: AlignCenter, Content: []string{"Single line"}},
		),
	)
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty with single line")
	}

	// Should have minimal height
	lineCount := strings.Count(view, "\n")
	if lineCount > 5 {
		t.Errorf("Single line header should have minimal lines, got %d", lineCount)
	}
}
