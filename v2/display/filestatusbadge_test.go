package display

import "testing"

func TestFileStatusAllBadgeTypesRender(t *testing.T) {
	tests := []struct {
		name   string
		status FileStatus
		want   string
	}{
		{name: "modified", status: FileStatusModified, want: "[M]"},
		{name: "added", status: FileStatusAdded, want: "[A]"},
		{name: "deleted", status: FileStatusDeleted, want: "[D]"},
		{name: "renamed", status: FileStatusRenamed, want: "[R]"},
		{name: "copied", status: FileStatusCopied, want: "[C]"},
		{name: "untracked", status: FileStatusUntracked, want: "[U]"},
		{name: "conflicted", status: FileStatusConflicted, want: "[!]"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			badge := NewFileStatusBadge(tc.status)
			plain := stripANSI(badge.View())
			if plain != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, plain)
			}
		})
	}
}

func TestFileStatusViewNonEmpty(t *testing.T) {
	badge := NewFileStatusBadge(FileStatusModified)
	if got := badge.View(); got == "" {
		t.Fatal("expected non-empty badge view")
	}
}

func TestFileStatusCompactRender(t *testing.T) {
	badge := NewFileStatusBadge(FileStatusAdded, WithFileStatusBadgeCompact(true))
	plain := stripANSI(badge.View())
	if plain != "A" {
		t.Fatalf("expected compact badge to render as %q, got %q", "A", plain)
	}
}
