package display

import (
	"strings"
	"testing"
)

func TestShellOutputConstructor(t *testing.T) {
	v := NewShellOutputView("echo hi", "hi")
	if v == nil {
		t.Fatal("expected non-nil ShellOutputView")
	}
	if v.command != "echo hi" {
		t.Fatalf("expected command %q, got %q", "echo hi", v.command)
	}
	if v.output != "hi" {
		t.Fatalf("expected output %q, got %q", "hi", v.output)
	}
}

func TestShellOutputCommandAndOutputRender(t *testing.T) {
	v := NewShellOutputView("ls -la", "line1\nline2", WithShellOutputViewWidth(30))
	plain := stripANSI(v.View())

	if !strings.Contains(plain, "$ ls -la") {
		t.Fatalf("expected command in view, got %q", plain)
	}
	if !strings.Contains(plain, "line1") || !strings.Contains(plain, "line2") {
		t.Fatalf("expected output lines in view, got %q", plain)
	}
}

func TestShellOutputExitCodeBadge(t *testing.T) {
	ok := NewShellOutputView("true", "", WithShellOutputViewExitCode(0))
	if !strings.Contains(stripANSI(ok.View()), "✓ Exit 0") {
		t.Fatalf("expected success exit badge, got %q", stripANSI(ok.View()))
	}

	fail := NewShellOutputView("false", "", WithShellOutputViewExitCode(2))
	if !strings.Contains(stripANSI(fail.View()), "✗ Exit 2") {
		t.Fatalf("expected failure exit badge, got %q", stripANSI(fail.View()))
	}
}
