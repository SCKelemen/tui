package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type tickMsg time.Time

type model struct {
	// Various spinners showcased
	spinners      []*tui.StructuredData
	iconSets      []*tui.StructuredData
	width         int
	height        int
	step          int
	showSpinners  bool
	showIconSets  bool
}

func initialModel() model {
	// Create examples with different spinners
	spinners := []*tui.StructuredData{
		tui.NewStructuredData("SpinnerDots", tui.WithSpinner(tui.SpinnerDots)).
			AddRow("Style", "Braille dots (smooth)").
			AddRow("Frames", "10"),
		tui.NewStructuredData("SpinnerThinking", tui.WithSpinner(tui.SpinnerThinking)).
			AddRow("Style", "Codex thinking animation").
			AddRow("Frames", ". * ÷ + •"),
		tui.NewStructuredData("SpinnerLine", tui.WithSpinner(tui.SpinnerLine)).
			AddRow("Style", "Classic line spinner").
			AddRow("Frames", "─ \\ | /"),
		tui.NewStructuredData("SpinnerCircle", tui.WithSpinner(tui.SpinnerCircle)).
			AddRow("Style", "Rotating circle").
			AddRow("Frames", "◴ ◷ ◶ ◵"),
		tui.NewStructuredData("SpinnerPulse", tui.WithSpinner(tui.SpinnerPulse)).
			AddRow("Style", "Pulsing circle").
			AddRow("Frames", "8 frames"),
		tui.NewStructuredData("SpinnerArrows", tui.WithSpinner(tui.SpinnerArrows)).
			AddRow("Style", "Rotating arrows").
			AddRow("Frames", "8 directions"),
	}

	// Create examples with different icon sets
	iconSets := []*tui.StructuredData{
		tui.NewStructuredData("IconSetDefault", tui.WithIconSet(tui.IconSetDefault)).
			AddRow("Success", "⏺").
			AddRow("Error", "⏺").
			AddRow("Warning", "⏺"),
		tui.NewStructuredData("IconSetCodex", tui.WithIconSet(tui.IconSetCodex)).
			AddRow("Success", "✓").
			AddRow("Error", "✗").
			AddRow("Warning", "⚠"),
		tui.NewStructuredData("IconSetSymbols", tui.WithIconSet(tui.IconSetSymbols)).
			AddRow("Success", "✓").
			AddRow("Error", "✗").
			AddRow("Warning", "⚠").
			AddRow("Info", "ℹ"),
		tui.NewStructuredData("IconSetEmoji", tui.WithIconSet(tui.IconSetEmoji)).
			AddRow("Success", "✅").
			AddRow("Error", "❌").
			AddRow("Warning", "⚡").
			AddRow("Info", "💡"),
		tui.NewStructuredData("IconSetCircles", tui.WithIconSet(tui.IconSetCircles)).
			AddRow("Success", "●").
			AddRow("Error", "◯").
			AddRow("Warning", "◐").
			AddRow("Info", "○"),
		tui.NewStructuredData("IconSetMinimal", tui.WithIconSet(tui.IconSetMinimal)).
			AddRow("Success", "+").
			AddRow("Error", "x").
			AddRow("Warning", "!").
			AddRow("Info", "i"),
	}

	return model{
		spinners:     spinners,
		iconSets:     iconSets,
		step:         0,
		showSpinners: true,
		showIconSets: false,
	}
}

func (m model) Init() tea.Cmd {
	// Start all spinners
	var cmds []tea.Cmd
	for _, sd := range m.spinners {
		cmds = append(cmds, sd.StartRunning())
	}
	cmds = append(cmds, tickCmd())
	return tea.Batch(cmds...)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			m.showSpinners = true
			m.showIconSets = false
		case "2":
			m.showSpinners = false
			m.showIconSets = true
		case "3":
			m.showSpinners = true
			m.showIconSets = true
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.step++

		// Cycle icon sets through different statuses
		if m.step%3 == 0 {
			for _, sd := range m.iconSets {
				sd.MarkSuccess()
			}
		} else if m.step%3 == 1 {
			for _, sd := range m.iconSets {
				sd.MarkError()
			}
		} else {
			for _, sd := range m.iconSets {
				sd.MarkWarning()
			}
		}

		return m, tickCmd()
	}

	// Update all components
	for i, sd := range m.spinners {
		comp, cmd := sd.Update(msg)
		m.spinners[i] = comp.(*tui.StructuredData)
		cmds = append(cmds, cmd)
	}

	for i, sd := range m.iconSets {
		comp, cmd := sd.Update(msg)
		m.iconSets[i] = comp.(*tui.StructuredData)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	s := "\n=== Spinner & Icon Set Showcase ===\n\n"
	s += "Press '1' for Spinners | '2' for Icon Sets | '3' for Both | 'q' to quit\n\n"

	if m.showSpinners {
		s += "--- Running Spinners (animated) ---\n\n"
		for _, sd := range m.spinners {
			s += sd.View() + "\n"
		}
	}

	if m.showIconSets {
		s += "--- Icon Sets (cycling statuses) ---\n\n"
		for _, sd := range m.iconSets {
			s += sd.View() + "\n"
		}
	}

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
