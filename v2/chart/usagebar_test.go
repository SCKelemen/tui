package chart

import (
	"strings"
	"testing"
)

func TestUsageBarCreation(t *testing.T) {
	ub := NewUsageBar(25, 100)
	if ub == nil {
		t.Fatal("NewUsageBar returned nil")
	}

	if ub.used != 25 {
		t.Errorf("expected used=25, got %v", ub.used)
	}
	if ub.total != 100 {
		t.Errorf("expected total=100, got %v", ub.total)
	}
	if ub.width != 24 {
		t.Errorf("expected default width=24, got %d", ub.width)
	}
	if ub.baseColor != "#D19A66" {
		t.Errorf("expected default baseColor #D19A66, got %q", ub.baseColor)
	}
	if !ub.showLabels {
		t.Error("showLabels should default to true")
	}
}

func TestUsageBarOptions(t *testing.T) {
	ub := NewUsageBar(10, 100,
		WithUsageBarLabel("Subscription"),
		WithUsageBarShowLabels(false),
		WithUsageBarBaseColor("#123456"),
		WithUsageBarWidth(40),
	)

	if ub.label != "Subscription" {
		t.Errorf("expected label Subscription, got %q", ub.label)
	}
	if ub.showLabels {
		t.Error("expected showLabels=false")
	}
	if ub.baseColor != "#123456" {
		t.Errorf("expected baseColor #123456, got %q", ub.baseColor)
	}
	if ub.width != 40 {
		t.Errorf("expected width=40, got %d", ub.width)
	}
}

func TestUsageBarViewNonEmpty(t *testing.T) {
	ub := NewUsageBar(50, 100)
	view := ub.View()
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty view")
	}
}

func TestUsageBarEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		used          float64
		total         float64
		wantContains  []string
		wantExcludes  []string
	}{
		{
			name:         "0 of 100",
			used:         0,
			total:        100,
			wantContains: []string{"Used: 0", "Remaining: 100", "Total: 100"},
		},
		{
			name:         "50 of 100",
			used:         50,
			total:        100,
			wantContains: []string{"Used: 50", "Remaining: 50", "Total: 100"},
		},
		{
			name:         "100 of 100",
			used:         100,
			total:        100,
			wantContains: []string{"Used: 100", "Remaining: 0", "Total: 100"},
			wantExcludes: []string{"░"},
		},
		{
			name:         "used greater than max clamps",
			used:         150,
			total:        100,
			wantContains: []string{"Used: 100", "Remaining: 0", "Total: 100"},
			wantExcludes: []string{"░"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ub := NewUsageBar(tt.used, tt.total)
			plain := stripANSI(ub.View())

			for _, want := range tt.wantContains {
				if !strings.Contains(plain, want) {
					t.Fatalf("expected %q in %q", want, plain)
				}
			}
			for _, excluded := range tt.wantExcludes {
				if strings.Contains(plain, excluded) {
					t.Fatalf("did not expect %q in %q", excluded, plain)
				}
			}
		})
	}
}
