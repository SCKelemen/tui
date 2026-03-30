package display

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestGitDiffViewConstructorWithFiles(t *testing.T) {
	files := []GitDiffFile{
		{Path: "main.go", Status: FileStatusModified, Diff: "@@ -1 +1 @@\n-old\n+new", LinesAdded: 1, LinesRemoved: 1},
		{Path: "go.mod", Status: FileStatusAdded, Diff: "+module x", LinesAdded: 1},
	}

	view := NewGitDiffView(files)
	if view == nil {
		t.Fatal("NewGitDiffView returned nil")
	}
	if len(view.files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(view.files))
	}
	if view.fileList == nil {
		t.Fatal("expected fileList to be initialized")
	}
	if view.selectedFile != 0 {
		t.Fatalf("expected selectedFile=0, got %d", view.selectedFile)
	}
}

func TestGitDiffViewNonEmptyView(t *testing.T) {
	files := []GitDiffFile{{
		Path:         "main.go",
		Status:       FileStatusModified,
		Diff:         "@@ -1,2 +1,2 @@\n-old\n+new",
		LinesAdded:   1,
		LinesRemoved: 1,
	}}

	view := NewGitDiffView(files, WithGitDiffViewWidth(80), WithGitDiffViewHeight(8))
	plain := stripANSI(view.View())

	if strings.TrimSpace(plain) == "" {
		t.Fatal("expected non-empty view")
	}
	if !strings.Contains(plain, "main.go") {
		t.Fatalf("expected file path in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "+1 -1") {
		t.Fatalf("expected diff stats in view, got:\n%s", plain)
	}
}

func TestGitDiffViewTabSwitchesPanes(t *testing.T) {
	files := []GitDiffFile{{Path: "main.go", Status: FileStatusModified, Diff: "+new"}}
	view := NewGitDiffView(files)

	if view.focusedPane != 0 {
		t.Fatalf("expected initial focusedPane=0, got %d", view.focusedPane)
	}
	if !view.fileList.Focused() {
		t.Fatal("expected file list to start focused")
	}

	_, _ = view.Update(tea.KeyMsg{Type: tea.KeyTab})
	if view.focusedPane != 1 {
		t.Fatalf("expected focusedPane=1 after tab, got %d", view.focusedPane)
	}
	if view.fileList.Focused() {
		t.Fatal("expected file list to blur when diff pane is focused")
	}

	_, _ = view.Update(tea.KeyMsg{Type: tea.KeyTab})
	if view.focusedPane != 0 {
		t.Fatalf("expected focusedPane=0 after second tab, got %d", view.focusedPane)
	}
	if !view.fileList.Focused() {
		t.Fatal("expected file list to be focused again")
	}
}
