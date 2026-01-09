package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHeaderCreation(t *testing.T) {
	header := NewHeader()
	if header == nil {
		t.Fatal("NewHeader returned nil")
	}

	if header.showDivider != true {
		t.Error("Header should have divider enabled by default")
	}
}

func TestHeaderWithColumns(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{
				Width:   50,
				Align:   AlignCenter,
				Content: []string{"Test"},
			},
			HeaderColumn{
				Width:   50,
				Align:   AlignLeft,
				Content: []string{"Content"},
			},
		),
	)

	if len(header.columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(header.columns))
	}
}

func TestHeaderView(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{
				Width:   50,
				Align:   AlignCenter,
				Content: []string{"Centered"},
			},
			HeaderColumn{
				Width:   50,
				Align:   AlignLeft,
				Content: []string{"Left aligned"},
			},
		),
	)

	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Check for rounded corners
	if !strings.Contains(view, "╭") {
		t.Error("View should contain top-left rounded corner (╭)")
	}
	if !strings.Contains(view, "╮") {
		t.Error("View should contain top-right rounded corner (╮)")
	}
	if !strings.Contains(view, "╰") {
		t.Error("View should contain bottom-left rounded corner (╰)")
	}
	if !strings.Contains(view, "╯") {
		t.Error("View should contain bottom-right rounded corner (╯)")
	}

	// Check for content
	if !strings.Contains(view, "Centered") {
		t.Error("View should contain 'Centered' text")
	}
	if !strings.Contains(view, "Left aligned") {
		t.Error("View should contain 'Left aligned' text")
	}
}

func TestHeaderWithSections(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignCenter},
			HeaderColumn{Width: 50, Align: AlignLeft},
		),
		WithColumnSections(1,
			HeaderSection{
				Title:   "Section 1",
				Content: []string{"Line 1", "Line 2"},
			},
			HeaderSection{
				Title:   "Section 2",
				Content: []string{"Line 3"},
				Divider: true,
			},
		),
	)

	if len(header.sections[1]) != 2 {
		t.Errorf("Expected 2 sections in column 1, got %d", len(header.sections[1]))
	}

	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := header.View()

	if !strings.Contains(view, "Section 1") {
		t.Error("View should contain 'Section 1'")
	}
	if !strings.Contains(view, "Section 2") {
		t.Error("View should contain 'Section 2'")
	}
}

func TestHeaderVerticalDivider(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 50, Align: AlignLeft, Content: []string{"A"}},
			HeaderColumn{Width: 50, Align: AlignRight, Content: []string{"B"}},
		),
		WithVerticalDivider(true),
	)

	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := header.View()

	// Count vertical bars (should have some in the middle for divider)
	barCount := strings.Count(view, "│")
	// Should have at least borders (2 per line) + dividers
	if barCount < 6 {
		t.Errorf("Expected multiple vertical bars, got %d", barCount)
	}
}

func TestHeaderAlignment(t *testing.T) {
	tests := []struct {
		name    string
		align   ColumnAlign
		content string
		width   int
		want    string
	}{
		{
			name:    "Left align",
			align:   AlignLeft,
			content: "Test",
			width:   10,
			want:    "Test      ",
		},
		{
			name:    "Right align",
			align:   AlignRight,
			content: "Test",
			width:   10,
			want:    "      Test",
		},
		{
			name:    "Center align",
			align:   AlignCenter,
			content: "Test",
			width:   10,
			want:    "   Test   ",
		},
	}

	header := NewHeader()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := header.alignContent(tt.content, tt.width, tt.align)
			if got != tt.want {
				t.Errorf("alignContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHeaderEmptyColumns(t *testing.T) {
	header := NewHeader()
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := header.View()
	if view != "" {
		t.Error("View should be empty when no columns are set")
	}
}

func TestHeaderFocusManagement(t *testing.T) {
	header := NewHeader()

	if header.Focused() {
		t.Error("Header should not be focused initially")
	}

	header.Focus()
	if !header.Focused() {
		t.Error("Header should be focused after Focus()")
	}

	header.Blur()
	if header.Focused() {
		t.Error("Header should not be focused after Blur()")
	}
}
