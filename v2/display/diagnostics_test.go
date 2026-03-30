package display

import (
	"strings"
	"testing"
)

func TestDiagnosticListConstructor(t *testing.T) {
	diags := []Diagnostic{{Message: "broken", Severity: SeverityError, File: "main.go", Line: 4, Column: 2}}

	list := NewDiagnosticList(diags)
	if list == nil {
		t.Fatal("NewDiagnosticList returned nil")
	}
	if len(list.diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(list.diagnostics))
	}
	if !list.groupByFile {
		t.Fatal("expected groupByFile=true by default")
	}
	if !list.showCounts {
		t.Fatal("expected showCounts=true by default")
	}
}

func TestDiagnosticListItemsRender(t *testing.T) {
	diags := []Diagnostic{
		{Message: "unused variable", Severity: SeverityWarning, File: "a.go", Line: 10, Column: 3, Source: "golangci", Code: "unused"},
		{Message: "nil pointer", Severity: SeverityError, File: "b.go", Line: 5, Column: 1},
	}

	list := NewDiagnosticList(diags)
	list.Focus()

	plain := stripANSI(list.View())
	if !strings.Contains(plain, "a.go") || !strings.Contains(plain, "b.go") {
		t.Fatalf("expected grouped file headers in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "unused variable") || !strings.Contains(plain, "nil pointer") {
		t.Fatalf("expected diagnostic messages in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "a.go:10:3") {
		t.Fatalf("expected diagnostic location in view, got:\n%s", plain)
	}
}

func TestDiagnosticSeverityIcons(t *testing.T) {
	cases := []struct {
		severity DiagnosticSeverity
		want     string
	}{
		{SeverityError, "✗"},
		{SeverityWarning, "⚠"},
		{SeverityInfo, "ℹ"},
		{SeverityHint, "💡"},
	}

	for _, tc := range cases {
		if got := SeverityIcon(tc.severity); got != tc.want {
			t.Fatalf("SeverityIcon(%v)=%q, want %q", tc.severity, got, tc.want)
		}
	}
}
