package display

import (
	"strings"
	"testing"
)

func TestGitFileListConstructorWithEntries(t *testing.T) {
	entries := []GitFileEntry{{Path: "main.go", Status: FileStatusModified}, {Path: "go.mod", Status: FileStatusAdded}}

	list := NewGitFileList(entries, WithGitFileListCursor(1))
	if list == nil {
		t.Fatal("NewGitFileList returned nil")
	}
	if len(list.files) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list.files))
	}
	if list.cursor != 1 {
		t.Fatalf("expected cursor=1, got %d", list.cursor)
	}
}

func TestGitFileListViewRendersFileNamesAndStats(t *testing.T) {
	entries := []GitFileEntry{
		{Path: "cmd/main.go", Status: FileStatusModified, LinesAdded: 12, LinesRemoved: 4},
		{Path: "README.md", Status: FileStatusAdded, LinesAdded: 3, LinesRemoved: 0},
	}

	list := NewGitFileList(entries, WithGitFileListWidth(100))
	plain := stripANSI(list.View())

	if !strings.Contains(plain, "cmd/main.go") || !strings.Contains(plain, "README.md") {
		t.Fatalf("expected file names in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "+12") || !strings.Contains(plain, "-4") {
		t.Fatalf("expected per-file stats in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "2 files changed") {
		t.Fatalf("expected summary line, got:\n%s", plain)
	}
}
