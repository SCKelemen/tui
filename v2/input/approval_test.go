package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestApprovalView(t *testing.T) {
	a := NewApproval(ApprovalAction{
		Tool:        "bash",
		Description: "Run shell command",
		Risk:        "high",
		Details:     "rm -rf /tmp/example",
	})

	view := a.View()

	checks := []string{
		"Permission required",
		"Tool:",
		"bash",
		"Description:",
		"Run shell command",
		"Risk:",
		"HIGH",
		"[✓ Approve]",
		"[✗ Deny]",
		"[⟳ Always Approve]",
	}

	for _, want := range checks {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestApprovalApproveKey(t *testing.T) {
	a := NewApproval(ApprovalAction{Tool: "bash"})

	_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected command for y key")
	}

	msg := cmd()
	result, ok := msg.(ApprovalResultMsg)
	if !ok {
		t.Fatalf("expected ApprovalResultMsg, got %T", msg)
	}

	if !result.Approved {
		t.Error("expected Approved=true for y key")
	}
	if result.Always {
		t.Error("expected Always=false for y key")
	}
}

func TestApprovalDenyKey(t *testing.T) {
	a := NewApproval(ApprovalAction{Tool: "bash"})

	_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatal("expected command for n key")
	}

	msg := cmd()
	result, ok := msg.(ApprovalResultMsg)
	if !ok {
		t.Fatalf("expected ApprovalResultMsg, got %T", msg)
	}

	if result.Approved {
		t.Error("expected Approved=false for n key")
	}
	if result.Always {
		t.Error("expected Always=false for n key")
	}
}

func TestApprovalAlwaysKey(t *testing.T) {
	a := NewApproval(ApprovalAction{Tool: "bash"})

	_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd == nil {
		t.Fatal("expected command for a key")
	}

	msg := cmd()
	result, ok := msg.(ApprovalResultMsg)
	if !ok {
		t.Fatalf("expected ApprovalResultMsg, got %T", msg)
	}

	if !result.Approved {
		t.Error("expected Approved=true for a key")
	}
	if !result.Always {
		t.Error("expected Always=true for a key")
	}
}

func TestApprovalNavigation(t *testing.T) {
	a := NewApproval(ApprovalAction{Tool: "bash"})

	if a.selected != 0 {
		t.Fatalf("expected initial selected=0, got %d", a.selected)
	}

	a.Update(tea.KeyMsg{Type: tea.KeyRight})
	if a.selected != 1 {
		t.Fatalf("expected selected=1 after right, got %d", a.selected)
	}

	a.Update(tea.KeyMsg{Type: tea.KeyRight})
	if a.selected != 2 {
		t.Fatalf("expected selected=2 after right, got %d", a.selected)
	}

	a.Update(tea.KeyMsg{Type: tea.KeyRight})
	if a.selected != 0 {
		t.Fatalf("expected selected=0 after wrap-around right, got %d", a.selected)
	}

	a.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if a.selected != 2 {
		t.Fatalf("expected selected=2 after wrap-around left, got %d", a.selected)
	}
}

func TestApprovalRiskColors(t *testing.T) {
	low := NewApproval(ApprovalAction{Risk: "low"}).riskBadge()
	medium := NewApproval(ApprovalAction{Risk: "medium"}).riskBadge()
	high := NewApproval(ApprovalAction{Risk: "high"}).riskBadge()

	if !strings.Contains(low, "LOW") {
		t.Errorf("expected LOW label, got %q", low)
	}
	if !strings.Contains(medium, "MEDIUM") {
		t.Errorf("expected MEDIUM label, got %q", medium)
	}
	if !strings.Contains(high, "HIGH") {
		t.Errorf("expected HIGH label, got %q", high)
	}

	// In color-capable terminals, lipgloss includes ANSI color escapes. Verify
	// risk levels map to different colors when ANSI output is present.
	if strings.Contains(low, "\x1b[") || strings.Contains(medium, "\x1b[") || strings.Contains(high, "\x1b[") {
		if !strings.Contains(low, "38;5;34") {
			t.Errorf("low risk badge missing expected color code: %q", low)
		}
		if !strings.Contains(medium, "38;5;220") {
			t.Errorf("medium risk badge missing expected color code: %q", medium)
		}
		if !strings.Contains(high, "38;5;160") {
			t.Errorf("high risk badge missing expected color code: %q", high)
		}
	}
}
func TestApprovalEnterConfirm(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Approval)
		approved bool
		always   bool
	}{
		{
			name: "approve",
			setup: func(a *Approval) {
				a.selected = 0
			},
			approved: true,
			always:   false,
		},
		{
			name: "deny",
			setup: func(a *Approval) {
				a.selected = 1
			},
			approved: false,
			always:   false,
		},
		{
			name: "always",
			setup: func(a *Approval) {
				a.selected = 2
			},
			approved: true,
			always:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewApproval(ApprovalAction{Tool: "bash"})
			tt.setup(a)

			_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if cmd == nil {
				t.Fatal("expected command for enter key")
			}

			msg := cmd()
			result, ok := msg.(ApprovalResultMsg)
			if !ok {
				t.Fatalf("expected ApprovalResultMsg, got %T", msg)
			}

			if result.Approved != tt.approved {
				t.Errorf("expected Approved=%v, got %v", tt.approved, result.Approved)
			}
			if result.Always != tt.always {
				t.Errorf("expected Always=%v, got %v", tt.always, result.Always)
			}
		})
	}
}
