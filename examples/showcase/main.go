package main

import (
	"strings"
	"time"
	colorpkg "github.com/SCKelemen/color"
	tui "github.com/SCKelemen/tui"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	tabBar       *tui.TabBar
	threads      *tui.ThreadProgress
	checklist    *tui.Checklist
	progressBar  *tui.ProgressBar
	sparkline    *tui.Sparkline
	subagents    *tui.SubagentGroup
	breadcrumb   *tui.Breadcrumb
	tree         tui.Component
	palette      *tui.FloatingPalette
	fileCardRow  string
	toast        *tui.Toast
	activeTab    string
}
func initialModel() model {
	m := model{
		tabBar: tui.NewTabBar(
			tui.WithTabs(
				tui.Tab{ID: "components", Label: "Components"},
				tui.Tab{ID: "panels", Label: "Panels"},
				tui.Tab{ID: "navigation", Label: "Navigation"},
				tui.Tab{ID: "input", Label: "Input"},
			),
		),
		threads: tui.NewThreadProgress(),
		checklist: tui.NewChecklist(
			tui.WithChecklistTitle("Release v1.6.0"),
			tui.WithShowProgress(true),
		),
		progressBar: tui.NewProgressBar(
			0.72,
			tui.WithProgressBarWidth(34),
			tui.WithProgressBarLabel("build pipeline"),
			tui.WithProgressBarGradient(
				colorpkg.NewRGBA(0.20, 0.60, 1.00, 1.00),
				colorpkg.NewRGBA(0.62, 0.30, 0.96, 1.00),
			),
		),
		sparkline: tui.NewSparkline(
			[]float64{18, 24, 22, 29, 34, 33, 37, 41, 45, 42, 49, 56},
			tui.WithSparklineWidth(36),
			tui.WithSparklineLabel("requests/min"),
		),
		palette: tui.NewFloatingPalette([]tui.PaletteCommand{
			{Name: "Open settings", Description: "Configure showcase options", Shortcut: "⌘,"},
			{Name: "Run tests", Description: "Execute quick test suite", Shortcut: "⌘T"},
			{Name: "Deploy preview", Description: "Push preview environment", Shortcut: "⌘D"},
			{Name: "View logs", Description: "Show latest application logs", Shortcut: "⌘L"},
		}),
		toast:     tui.NewToast(),
		activeTab: "components",
	}

	m.seedData()
	m.tabBar.Focus()
	m.applyFocus()
	m.toast.Push("Tab / Shift+Tab: switch pages", tui.ToastInfo)
	return m
}

func (m *model) seedData() {
	m.threads.AddThread("t1", "Assemble component registry")
	m.threads.UpdateThread("t1", tui.ThreadCompleted)
	m.threads.AppendOutput("t1", "Loaded v1.5.0 and v1.6.0 widgets")
	m.threads.SetThreadExpanded("t1", true)

	m.threads.AddThread("t2", "Render screenshot assets")
	m.threads.UpdateThread("t2", tui.ThreadRunning)
	m.threads.AppendOutput("t2", "Generated 9/12 previews")
	m.threads.SetThreadExpanded("t2", true)

	m.threads.AddThread("t3", "Publish changelog summary")
	m.threads.UpdateThread("t3", tui.ThreadFailed)
	m.threads.AppendOutput("t3", "Missing release note for FileCard")

	m.checklist.AddItem(tui.ChecklistItem{ID: "c1", Label: "ThreadProgress", Status: tui.ItemDone})
	m.checklist.AddItem(tui.ChecklistItem{ID: "c2", Label: "Checklist", Status: tui.ItemDone})
	m.checklist.AddItem(tui.ChecklistItem{ID: "c3", Label: "Toast", Status: tui.ItemDone})
	m.checklist.AddItem(tui.ChecklistItem{ID: "c4", Label: "ProgressBar gradient", Status: tui.ItemInProgress})
	m.checklist.AddItem(tui.ChecklistItem{ID: "c5", Label: "Sparkline", Status: tui.ItemPending})

	left := tui.NewSubagentPanel(tui.WithSubagentTitle("frontend-agent"), tui.WithSubagentModel("gpt-4.1"))
	left.SetTools([]tui.SubagentTool{{Name: "Read design tokens", Status: tui.ToolCompleted}, {Name: "Build tab interactions", Status: tui.ToolCompleted}})
	left.SetStatus(tui.SubagentCompleted)
	left.SetElapsed(41 * time.Second)
	left.SetTokenCount("5.3k tok")
	left.SetCost("$0.08")

	right := tui.NewSubagentPanel(tui.WithSubagentTitle("qa-agent"), tui.WithSubagentModel("gpt-4o-mini"))
	right.SetTools([]tui.SubagentTool{{Name: "Run terminal build", Status: tui.ToolRunning}, {Name: "Review regressions", Status: tui.ToolCompleted}})
	right.SetStatus(tui.SubagentRunning)
	right.SetElapsed(19 * time.Second)
	right.SetTokenCount("2.1k tok")
	right.SetCost("$0.03")

	m.subagents = tui.NewSubagentGroup(tui.WithSubagentGroupGap(3))
	m.subagents.SetPanels([]*tui.SubagentPanel{left, right})

	m.breadcrumb = tui.NewBreadcrumb(tui.WithSeparator(" / "), tui.WithBreadcrumbItems(
		tui.BreadcrumbItem{ID: "root", Label: "workspace"},
		tui.BreadcrumbItem{ID: "stack", Label: "stack"},
		tui.BreadcrumbItem{ID: "showcase", Label: "examples"},
		tui.BreadcrumbItem{ID: "main", Label: "main.go"},
	))

	m.tree = tui.NewTree([]*tui.TreeNode{
		{Label: "stack", Icon: "📁", Expanded: true, Children: []*tui.TreeNode{
			{Label: "tui", Icon: "📁", Expanded: true, Children: []*tui.TreeNode{
				{Label: "examples", Icon: "📁", Expanded: true, Children: []*tui.TreeNode{
					{Label: "showcase", Icon: "📁", Expanded: true, Children: []*tui.TreeNode{{Label: "main.go", Icon: "📄"}}},
				}},
				{Label: "progressbar.go", Icon: "📄"},
				{Label: "floatingpalette.go", Icon: "📄"},
			}},
		}},
	}, tui.WithTreeShowIcons(true), tui.WithTreeShowGuides(true), tui.WithTreeIndentSize(2))

	m.fileCardRow = tui.NewFileCardRow(
		tui.NewFileCard("showcase/main.go", 143, 28),
		tui.NewFileCard("progressbar.go", 38, 6),
		tui.NewFileCard("floatingpalette.go", 24, 4),
	)
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.toast.Init(), m.subagents.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			m.toast.Push("Toast demo: components updated", tui.ToastSuccess)
		case "ctrl+p":
			if m.activeTab == "input" {
				m.palette.Toggle()
				if m.palette.IsVisible() {
					m.toast.Push("FloatingPalette opened", tui.ToastInfo)
				} else {
					m.toast.Push("FloatingPalette closed", tui.ToastWarning)
				}
			}
		}
	case tui.PaletteSelectMsg:
		m.toast.Push("Selected: "+msg.Command.Name, tui.ToastSuccess)
	case tui.PaletteDismissMsg:
		m.toast.Push("Palette dismissed", tui.ToastInfo)
	}

	if updated, cmd := m.tabBar.Update(msg); updated != nil {
		if tb, ok := updated.(*tui.TabBar); ok {
			m.tabBar = tb
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if id := m.tabBar.GetActiveID(); id != "" && id != m.activeTab {
		m.activeTab = id
		m.applyFocus()
		m.toast.Push("Switched to "+id, tui.ToastInfo)	}

	if updated, cmd := m.palette.Update(msg); updated != nil {
		if fp, ok := updated.(*tui.FloatingPalette); ok {
			m.palette = fp
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	switch m.activeTab {
	case "components":
		if updated, cmd := m.threads.Update(msg); updated != nil {
			if tp, ok := updated.(*tui.ThreadProgress); ok {
				m.threads = tp
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if updated, cmd := m.checklist.Update(msg); updated != nil {
			if cl, ok := updated.(*tui.Checklist); ok {
				m.checklist = cl
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case "panels":
		if updated, cmd := m.subagents.Update(msg); updated != nil {
			if sg, ok := updated.(*tui.SubagentGroup); ok {
				m.subagents = sg
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case "navigation":
		m.breadcrumb.Update(msg)
		m.tree, _ = m.tree.Update(msg)
	}

	if updated, cmd := m.toast.Update(msg); updated != nil {
		if t, ok := updated.(*tui.Toast); ok {
			m.toast = t
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) applyFocus() {
	m.threads.Blur()
	m.checklist.Blur()
	m.subagents.Blur()
	m.breadcrumb.Blur()
	m.tree.Blur()
	m.palette.Blur()

	switch m.activeTab {
	case "components":
		m.threads.Focus()
		m.checklist.Focus()
	case "panels":
		m.subagents.Focus()
	case "navigation":
		m.tree.Focus()
		m.breadcrumb.Focus()
	case "input":
		if m.palette.IsVisible() {
			m.palette.Focus()
		}
	}
}

func (m model) tabContent() string {
	switch m.activeTab {
	case "components":
		return strings.Join([]string{
			"ThreadProgress",
			m.threads.View(),
			"Checklist",
			m.checklist.View(),
			"ProgressBar (gradient)",
			m.progressBar.View(),
			"Sparkline",
			m.sparkline.View(),
		}, "\n")
	case "panels":
		return "SubagentPanel + SubagentGroup\n" + m.subagents.View()
	case "navigation":
		return strings.Join([]string{"Breadcrumb", m.breadcrumb.View(), "", "Tree", m.tree.View()}, "\n")
	case "input":
		base := strings.Join([]string{
			"Ctrl+P toggles FloatingPalette",
			"",
			"FileCard row",
			m.fileCardRow,
		}, "\n")
		if m.palette.IsVisible() {
			return base + "\n\n" + m.palette.View()
		}
		return base
	default:
		return ""
	}
}

func (m model) View() string {
	header := "Showcase demo • active tab: " + m.activeTab
	return strings.Join([]string{header, m.tabBar.View(), m.tabContent(), m.toast.View()}, "\n")}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
