package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type tickMsg time.Time

type model struct {
	// Components gallery
	header             *tui.Header
	activityBar        *tui.ActivityBar
	statusBar          *tui.StatusBar
	structuredData1    *tui.StructuredData
	structuredData2    *tui.StructuredData
	structuredData3    *tui.StructuredData
	toolBlock1         *tui.ToolBlock
	toolBlock2         *tui.ToolBlock
	textInput          *tui.TextInput
	commandPalette     *tui.CommandPalette
	fileExplorer       *tui.FileExplorer
	modal              *tui.Modal

	// State
	width           int
	height          int
	step            int
	currentSection  int // 0=all, 1=status, 2=data, 3=tools, 4=input
	activityRunning bool
}

func initialModel() model {
	cwd, _ := os.Getwd()

	// Header with multiple sections
	header := tui.NewHeader(
		tui.WithColumns(
			tui.HeaderColumn{
				Width:   20,
				Align:   tui.AlignLeft,
				Content: []string{"🎨 Kitchen Sink", "All Components"},
			},
			tui.HeaderColumn{
				Width:   30,
				Align:   tui.AlignCenter,
				Content: []string{"TUI Component Gallery", "Press 1-5 for sections"},
			},
			tui.HeaderColumn{
				Width:   25,
				Align:   tui.AlignRight,
				Content: []string{"228 tests passing ✓", "v1.0.0"},
			},
		),
	)

	// Activity bar
	activityBar := tui.NewActivityBar()

	// Status bar
	statusBar := tui.NewStatusBar()
	statusBar.SetMessage("Press 'q' to quit | 'm' for modal | 'p' for palette | 'r' to run activity")

	// StructuredData with different configurations
	sd1 := tui.NewStructuredData("Cost Summary",
		tui.WithSpinner(tui.SpinnerThinking),
		tui.WithIconSet(tui.IconSetClaude))
	sd1.AddRow("Total cost", "$122.25")
	sd1.AddRow("Duration", "6h 10m 48s")
	sd1.AddSeparator()
	sd1.AddHeader("Usage by model")
	sd1.AddIndentedRow("claude-haiku", "$1.61", 1)
	sd1.AddIndentedRow("claude-sonnet", "$120.63", 1)

	sd2 := tui.NewStructuredData("System Info",
		tui.WithSpinner(tui.SpinnerDots),
		tui.WithIconSet(tui.IconSetSymbols))
	sd2.AddRow("OS", "macOS 14.2.1")
	sd2.AddRow("Arch", "arm64")
	sd2.AddRow("CPU", "Apple M2 Pro")
	sd2.AddRow("Memory", "32 GB")

	sd3 := tui.NewStructuredData("Test Results",
		tui.WithSpinner(tui.SpinnerPulse),
		tui.WithIconSet(tui.IconSetEmoji),
		tui.WithStructuredDataMaxLines(3))
	sd3.AddColoredRow("Total", "228", "\033[32m")
	sd3.AddColoredRow("Passed", "228", "\033[32m")
	sd3.AddColoredRow("Failed", "0", "\033[2m")
	sd3.AddRow("Duration", "11.64s")

	// ToolBlocks
	tb1 := tui.NewToolBlock("Bash", "go test -v", []string{}, tui.WithMaxLines(5))
	tb1.AppendLine("=== RUN   TestStructuredData")
	tb1.AppendLine("--- PASS: TestStructuredData (0.00s)")
	tb1.AppendLine("=== RUN   TestSpinner")
	tb1.AppendLine("--- PASS: TestSpinner (0.00s)")
	tb1.AppendLine("PASS")
	tb1.AppendLine("ok  	github.com/SCKelemen/tui	11.64s")
	tb1.SetStatus(tui.StatusComplete)

	tb2 := tui.NewToolBlock("Bash", "git status", []string{})
	tb2.AppendLine("On branch main")
	tb2.AppendLine("Your branch is up to date with 'origin/main'.")
	tb2.AppendLine("")
	tb2.AppendLine("nothing to commit, working tree clean")
	tb2.SetStatus(tui.StatusComplete)

	// TextInput
	textInput := tui.NewTextInput()
	textInput.OnSubmit(func(text string) tea.Cmd {
		statusBar.SetMessage(fmt.Sprintf("You typed: %s", text))
		return nil
	})

	// Command palette
	commands := []tui.Command{
		{
			Name:        "Show Modal",
			Description: "Display a modal dialog",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return "show-modal"
				}
			},
		},
		{
			Name:        "Run Activity",
			Description: "Start activity bar animation",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return "start-activity"
				}
			},
		},
		{
			Name:        "Toggle Section",
			Description: "Cycle through different sections",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return "toggle-section"
				}
			},
		},
	}
	commandPalette := tui.NewCommandPalette(commands)

	// File explorer
	fileExplorer := tui.NewFileExplorer(cwd)

	// Modal
	modal := tui.NewModal(
		tui.WithModalTitle("Welcome to Kitchen Sink!"),
		tui.WithModalMessage(
			"This demo showcases all TUI components:\n\n"+
				"• Headers with multiple columns\n"+
				"• Activity bar with spinner\n"+
				"• Structured data with animations\n"+
				"• Tool blocks with status\n"+
				"• Text input with callbacks\n"+
				"• Command palette\n"+
				"• File explorer\n"+
				"• Status bar\n"+
				"• And this modal!\n\n"+
				"Press numbers 1-5 to switch sections\n"+
				"Press 'r' to run activity bar\n"+
				"Press 'p' for command palette\n"+
				"Press 'm' to toggle this modal",
		),
		tui.WithModalType(tui.ModalAlert),
	)

	// Show and focus modal initially
	modal.Show()
	modal.Focus()

	return model{
		header:          header,
		activityBar:     activityBar,
		statusBar:       statusBar,
		structuredData1: sd1,
		structuredData2: sd2,
		structuredData3: sd3,
		toolBlock1:      tb1,
		toolBlock2:      tb2,
		textInput:       textInput,
		commandPalette:  commandPalette,
		fileExplorer:    fileExplorer,
		modal:           modal,
		currentSection:  0, // Show all
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.structuredData1.StartRunning(),
		m.structuredData2.StartRunning(),
		m.structuredData3.StartRunning(),
		tickCmd(),
	)
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
		// Always allow quitting with Ctrl+C or q
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		// Modal gets priority when visible
		if m.modal.IsVisible() {
			// Check for 'm' to toggle modal
			if msg.String() == "m" {
				m.modal.Hide()
				m.modal.Blur()
				return m, nil
			}
			// Let modal handle all other keys (including Esc, Enter)
			comp, cmd := m.modal.Update(msg)
			m.modal = comp.(*tui.Modal)
			return m, cmd
		}

		// Command palette gets priority when visible
		if m.commandPalette.IsVisible() {
			// Let command palette handle all keys (including Esc, Enter)
			comp, cmd := m.commandPalette.Update(msg)
			m.commandPalette = comp.(*tui.CommandPalette)

			// If palette was hidden, blur it
			if !m.commandPalette.IsVisible() {
				m.commandPalette.Blur()
			}
			return m, cmd
		}

		switch msg.String() {
		case "m":
			m.modal.Show()
			m.modal.Focus()

		case "p":
			m.commandPalette.Show()
			m.commandPalette.Focus()

		case "r":
			if !m.activityRunning {
				m.activityBar.Start("Processing kitchen sink demo...")
				m.activityRunning = true
				m.statusBar.SetMessage("Activity running... Press 's' to stop")
			}

		case "s":
			if m.activityRunning {
				m.activityBar.Stop()
				m.activityRunning = false
				m.statusBar.SetMessage("Activity stopped. Press 'r' to restart")
			}

		case "1":
			m.currentSection = 1 // Status indicators
			m.statusBar.SetMessage("Section 1: Status Indicators & Activities")

		case "2":
			m.currentSection = 2 // Data display
			m.statusBar.SetMessage("Section 2: Structured Data Components")

		case "3":
			m.currentSection = 3 // Tool blocks
			m.statusBar.SetMessage("Section 3: Tool Blocks & Output")

		case "4":
			m.currentSection = 4 // Input components
			m.statusBar.SetMessage("Section 4: Input Components")

		case "5":
			m.currentSection = 0 // Show all
			m.statusBar.SetMessage("Showing all components")
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.step++

		// Cycle structured data statuses
		switch m.step % 4 {
		case 0:
			m.structuredData1.MarkSuccess()
			m.structuredData2.MarkSuccess()
			m.structuredData3.MarkSuccess()
		case 1:
			m.structuredData1.MarkError()
			m.structuredData2.MarkWarning()
			m.structuredData3.MarkInfo()
		case 2:
			m.structuredData1.StartRunning()
			m.structuredData2.StartRunning()
			m.structuredData3.StartRunning()
		case 3:
			m.structuredData1.MarkWarning()
			m.structuredData2.MarkInfo()
			m.structuredData3.MarkSuccess()
		}

		// Update activity bar progress
		if m.activityRunning {
			m.activityBar.SetProgress(fmt.Sprintf("Step %d...", m.step))
		}

		return m, tickCmd()

	case string:
		switch msg {
		case "show-modal":
			m.modal.Show()
			m.modal.Focus()
		case "start-activity":
			if !m.activityRunning {
				m.activityBar.Start("Running from command palette...")
				m.activityRunning = true
			}
		case "toggle-section":
			m.currentSection = (m.currentSection + 1) % 5
		}
	}

	// Update all components
	comp, cmd := m.header.Update(msg)
	m.header = comp.(*tui.Header)
	cmds = append(cmds, cmd)

	comp, cmd = m.activityBar.Update(msg)
	m.activityBar = comp.(*tui.ActivityBar)
	cmds = append(cmds, cmd)

	comp, cmd = m.statusBar.Update(msg)
	m.statusBar = comp.(*tui.StatusBar)
	cmds = append(cmds, cmd)

	comp, cmd = m.structuredData1.Update(msg)
	m.structuredData1 = comp.(*tui.StructuredData)
	cmds = append(cmds, cmd)

	comp, cmd = m.structuredData2.Update(msg)
	m.structuredData2 = comp.(*tui.StructuredData)
	cmds = append(cmds, cmd)

	comp, cmd = m.structuredData3.Update(msg)
	m.structuredData3 = comp.(*tui.StructuredData)
	cmds = append(cmds, cmd)

	comp, cmd = m.toolBlock1.Update(msg)
	m.toolBlock1 = comp.(*tui.ToolBlock)
	cmds = append(cmds, cmd)

	comp, cmd = m.toolBlock2.Update(msg)
	m.toolBlock2 = comp.(*tui.ToolBlock)
	cmds = append(cmds, cmd)

	comp, cmd = m.textInput.Update(msg)
	m.textInput = comp.(*tui.TextInput)
	cmds = append(cmds, cmd)

	comp, cmd = m.commandPalette.Update(msg)
	m.commandPalette = comp.(*tui.CommandPalette)
	cmds = append(cmds, cmd)

	comp, cmd = m.fileExplorer.Update(msg)
	m.fileExplorer = comp.(*tui.FileExplorer)
	cmds = append(cmds, cmd)

	comp, cmd = m.modal.Update(msg)
	m.modal = comp.(*tui.Modal)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Modal overlay
	if m.modal.IsVisible() {
		return m.modal.View()
	}

	// Command palette overlay
	if m.commandPalette.IsVisible() {
		return m.commandPalette.View()
	}

	s := ""

	// Header (always visible)
	s += m.header.View() + "\n\n"

	// Activity bar (section 0 and 1)
	if m.currentSection == 0 || m.currentSection == 1 {
		s += "=== Status Indicators & Activities ===\n\n"
		s += m.activityBar.View() + "\n"
		s += m.statusBar.View() + "\n\n"
	}

	// Structured data (section 0 and 2)
	if m.currentSection == 0 || m.currentSection == 2 {
		s += "=== Structured Data (Multiple Spinners & Icon Sets) ===\n\n"
		s += m.structuredData1.View() + "\n"
		s += m.structuredData2.View() + "\n"
		s += m.structuredData3.View() + "\n"
	}

	// Tool blocks (section 0 and 3)
	if m.currentSection == 0 || m.currentSection == 3 {
		s += "=== Tool Blocks (Command Output) ===\n\n"
		s += m.toolBlock1.View() + "\n"
		s += m.toolBlock2.View() + "\n"
	}

	// Input components (section 0 and 4)
	if m.currentSection == 0 || m.currentSection == 4 {
		s += "=== Input Components ===\n\n"
		s += "Text Input:\n"
		s += m.textInput.View() + "\n\n"
		s += "File Explorer (↑↓ to navigate, Enter to select):\n"
		s += m.fileExplorer.View() + "\n"
	}

	s += "\n"
	s += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	s += "Keyboard shortcuts: 1-5 (sections) | m (modal) | p (palette) | r (run) | s (stop) | q (quit)\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
