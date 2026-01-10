package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFileExplorerInvalidPath(t *testing.T) {
	invalidPath := "/this/path/does/not/exist/at/all"
	fe := NewFileExplorer(invalidPath)

	// Should not panic
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := fe.View()

	// View might be empty or show error, but shouldn't panic
	_ = view
}

func TestFileExplorerEmptyDirectory(t *testing.T) {
	// Create a temporary empty directory
	tempDir := os.TempDir()
	emptyDir := filepath.Join(tempDir, "tui_test_empty")
	os.MkdirAll(emptyDir, 0755)
	defer os.RemoveAll(emptyDir)

	fe := NewFileExplorer(emptyDir)
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := fe.View()
	if view == "" {
		t.Error("View should not be empty even with empty directory")
	}
}

func TestFileExplorerVeryDeepNesting(t *testing.T) {
	// Create a deep nested structure
	tempDir := os.TempDir()
	deepDir := filepath.Join(tempDir, "tui_test_deep")
	currentPath := deepDir

	// Create 10 levels deep
	for i := 0; i < 10; i++ {
		currentPath = filepath.Join(currentPath, "level")
	}
	os.MkdirAll(currentPath, 0755)
	defer os.RemoveAll(deepDir)

	fe := NewFileExplorer(deepDir)
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Expand nodes to go deep
	fe.Focus()
	for i := 0; i < 10; i++ {
		fe.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Expand
		fe.Update(tea.KeyMsg{Type: tea.KeyDown})  // Move down
	}

	view := fe.View()
	if view == "" {
		t.Error("View should not be empty with deep nesting")
	}
}

func TestFileExplorerManyFiles(t *testing.T) {
	// Create directory with many files
	tempDir := os.TempDir()
	manyFilesDir := filepath.Join(tempDir, "tui_test_many")
	os.MkdirAll(manyFilesDir, 0755)
	defer os.RemoveAll(manyFilesDir)

	// Create 100 files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(manyFilesDir, "file_"+string(rune(i))+".txt")
		os.WriteFile(filename, []byte("test"), 0644)
	}

	fe := NewFileExplorer(manyFilesDir)
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := fe.View()
	if view == "" {
		t.Error("View should not be empty with many files")
	}
}

func TestFileExplorerScrolling(t *testing.T) {
	// Create directory with enough files to require scrolling
	tempDir := os.TempDir()
	scrollDir := filepath.Join(tempDir, "tui_test_scroll")
	os.MkdirAll(scrollDir, 0755)
	defer os.RemoveAll(scrollDir)

	// Create 50 files
	for i := 0; i < 50; i++ {
		filename := filepath.Join(scrollDir, "file_"+string(rune(i))+".txt")
		os.WriteFile(filename, []byte("test"), 0644)
	}

	fe := NewFileExplorer(scrollDir)
	fe.Focus()
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	initialIndex := fe.selectedIndex

	// Press down many times to scroll
	for i := 0; i < 30; i++ {
		fe.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	if fe.selectedIndex == initialIndex {
		t.Error("Selection should move when pressing Down multiple times")
	}

	// Should not panic or crash
	view := fe.View()
	if view == "" {
		t.Error("View should not be empty after scrolling")
	}
}

func TestFileExplorerEnterKeyExpand(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot get current directory")
	}

	fe := NewFileExplorer(cwd)
	fe.Focus()
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Find a directory node and try to expand it
	// First node should be the root
	fe.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should not panic
	view := fe.View()
	if view == "" {
		t.Error("View should not be empty after Enter")
	}
}

func TestFileExplorerHiddenFiles(t *testing.T) {
	// Create directory with hidden files
	tempDir := os.TempDir()
	hiddenDir := filepath.Join(tempDir, "tui_test_hidden")
	os.MkdirAll(hiddenDir, 0755)
	defer os.RemoveAll(hiddenDir)

	// Create regular and hidden files
	os.WriteFile(filepath.Join(hiddenDir, "visible.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(hiddenDir, ".hidden.txt"), []byte("test"), 0644)

	// Test without showing hidden
	fe := NewFileExplorer(hiddenDir)
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	viewNoHidden := fe.View()

	// Test with showing hidden
	feWithHidden := NewFileExplorer(hiddenDir, WithShowHidden(true))
	feWithHidden.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	viewWithHidden := feWithHidden.View()

	// Views might be different (hard to test exact content due to ANSI codes)
	if viewNoHidden == "" || viewWithHidden == "" {
		t.Error("Views should not be empty")
	}
}

func TestFileExplorerNarrowWidth(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot get current directory")
	}

	fe := NewFileExplorer(cwd)
	fe.Update(tea.WindowSizeMsg{Width: 20, Height: 24})

	// Should not panic with narrow width
	view := fe.View()
	if view == "" {
		t.Error("View should not be empty with narrow width")
	}
}

func TestFileExplorerVeryNarrowWidth(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot get current directory")
	}

	fe := NewFileExplorer(cwd)
	fe.Update(tea.WindowSizeMsg{Width: 5, Height: 24})

	// Should not panic even with very narrow width
	view := fe.View()
	// View might be empty or truncated, but shouldn't panic
	_ = view
}

func TestFileExplorerShortHeight(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot get current directory")
	}

	fe := NewFileExplorer(cwd)
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 5})

	// Should not panic with short height
	view := fe.View()
	if view == "" {
		t.Error("View should not be empty with short height")
	}
}

func TestFileExplorerLongFilenames(t *testing.T) {
	// Create directory with very long filenames
	tempDir := os.TempDir()
	longNameDir := filepath.Join(tempDir, "tui_test_longname")
	os.MkdirAll(longNameDir, 0755)
	defer os.RemoveAll(longNameDir)

	// Create file with very long name
	longName := strings.Repeat("very_long_filename_", 10) + ".txt"
	os.WriteFile(filepath.Join(longNameDir, longName), []byte("test"), 0644)

	fe := NewFileExplorer(longNameDir)
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := fe.View()
	if view == "" {
		t.Error("View should not be empty with long filenames")
	}

	// Should truncate or wrap, not overflow
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		strippedLine := stripANSI(line)
		if len(strippedLine) > 100 {
			t.Errorf("Line should be truncated, got length %d", len(strippedLine))
		}
	}
}

func TestFileExplorerRapidNavigation(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot get current directory")
	}

	fe := NewFileExplorer(cwd)
	fe.Focus()
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Rapidly press keys
	for i := 0; i < 100; i++ {
		fe.Update(tea.KeyMsg{Type: tea.KeyDown})
		fe.Update(tea.KeyMsg{Type: tea.KeyUp})
		fe.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}

	// Should not panic or crash
	view := fe.View()
	if view == "" {
		t.Error("View should not be empty after rapid navigation")
	}
}

func TestFileExplorerSpecialCharactersInFilename(t *testing.T) {
	// Create directory with special characters in filename
	tempDir := os.TempDir()
	specialDir := filepath.Join(tempDir, "tui_test_special")
	os.MkdirAll(specialDir, 0755)
	defer os.RemoveAll(specialDir)

	// Create files with special characters (that are valid on most systems)
	specialNames := []string{
		"file-with-dash.txt",
		"file_with_underscore.txt",
		"file.multiple.dots.txt",
		"file (with parens).txt",
		"file [with brackets].txt",
	}

	for _, name := range specialNames {
		os.WriteFile(filepath.Join(specialDir, name), []byte("test"), 0644)
	}

	fe := NewFileExplorer(specialDir)
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := fe.View()
	if view == "" {
		t.Error("View should not be empty with special characters in filenames")
	}
}

func TestFileExplorerNoPermissions(t *testing.T) {
	// This test might not work on all systems (Windows, root user, etc.)
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := os.TempDir()
	noPermDir := filepath.Join(tempDir, "tui_test_noperm")
	os.MkdirAll(noPermDir, 0755)
	defer func() {
		os.Chmod(noPermDir, 0755) // Restore permissions before removing
		os.RemoveAll(noPermDir)
	}()

	// Create a subdirectory
	subDir := filepath.Join(noPermDir, "subdir")
	os.MkdirAll(subDir, 0755)

	// Remove read permission
	os.Chmod(subDir, 0000)

	fe := NewFileExplorer(noPermDir)
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Should not panic even if we can't read subdirectory
	view := fe.View()
	if view == "" {
		t.Error("View should not be empty even with permission issues")
	}
}
