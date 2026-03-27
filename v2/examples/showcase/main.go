package main
import (
	"strings"
	"time"
	colorpkg "github.com/SCKelemen/color"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/agent"
	"github.com/SCKelemen/tui/v2/chart"
	"github.com/SCKelemen/tui/v2/display"
	"github.com/SCKelemen/tui/v2/input"
	"github.com/SCKelemen/tui/v2/nav"
	tea "github.com/charmbracelet/bubbletea"
)
type model struct {
	tabBar      *nav.TabBar
	threads     *agent.ThreadProgress
	checklist   *input.Checklist
	progressBar *chart.ProgressBar
	sparkline   *chart.Sparkline
	subagents   *agent.SubagentGroup
	breadcrumb  *nav.Breadcrumb
	tree        tui.Component
	palette     *input.FloatingPalette
	fileCardRow string
	toast       *display.Toast
	activeTab   string
}
func initialModel() model {
	m := model{
		tabBar: nav.NewTabBar(
			nav.WithTabs(
				nav.Tab{ID: "components", Label: "Components"},
				nav.Tab{ID: "panels", Label: "Panels"},
				nav.Tab{ID: "navigation", Label: "Navigation"},
				nav.Tab{ID: "input", Label: "Input"},
			),
		),
		threads: agent.NewThreadProgress(),
		checklist: input.NewChecklist(
			input.WithChecklistTitle("Release v2.0.0"),
			input.WithShowProgress(true),
		),
		progressBar: chart.NewProgressBar(
			0.72,
			chart.WithProgressBarWidth(34),
			chart.WithProgressBarLabel("build pipeline"),
			chart.WithProgressBarGradient(
				colorpkg.NewRGBA(0.20, 0.60, 1.00, 1.00),
				colorpkg.NewRGBA(0.62, 0.30, 0.96, 1.00),
			),
		),
		sparkline: chart.NewSparkline(
			[]float64{18, 24, 22, 29, 34, 33, 37, 41, 45, 42, 49, 56},
			chart.WithSparklineWidth(36),
			chart.WithSparklineLabel("requests/min"),
		),
		palette: input.NewFloatingPalette([]input.PaletteCommand{
			{Name: "Open settings", Description: "Configure showcase options", Shortcut: "⌘,"},
			{Name: "Run tests", Description: "Execute quick test suite", Shortcut: "⌘T"},
			{Name: "Deploy preview", Description: "Push preview environment", Shortcut: "⌘D"},
			{Name: "View logs", Description: "Show latest application logs", Shortcut: "⌘L"},
		}),
		toast:     display.NewToast(),
		activeTab: "components",
	}
	m.seedData()
	m.tabBar.Focus()
	m.applyFocus()
	m.toast.Push("Tab / Shift+Tab: switch pages", display.ToastInfo)
	return m
}
func (m *model) seedData() {
	m.threads.AddThread("t1", "Assemble component registry")
	m.threads.UpdateThread("t1", agent.ThreadCompleted)
	m.threads.AppendOutput("t1", "Loaded v1.5.0 and v2.0.0 widgets")
	m.threads.SetThreadExpanded("t1", true)
	m.threads.AddThread("t2", "Render screenshot assets")
	m.threads.UpdateThread("t2", agent.ThreadRunning)
	m.threads.AppendOutput("t2", "Generated 9/12 previews")
	m.threads.SetThreadExpanded("t2", true)
	m.threads.AddThread("t3", "Publish changelog summary")
	m.threads.UpdateThread("t3", agent.ThreadFailed)
	m.threads.AppendOutput("t3", "Missing release note for FileCard")
	m.checklist.AddItem(input.ChecklistItem{ID: "c1", Label: "ThreadProgress", Status: input.ItemDone})
	m.checklist.AddItem(input.ChecklistItem{ID: "c2", Label: "Checklist", Status: input.ItemDone})
	m.checklist.AddItem(input.ChecklistItem{ID: "c3", Label: "Toast", Status: input.ItemDone})
	m.checklist.AddItem(input.ChecklistItem{ID: "c4", Label: "ProgressBar gradient", Status: input.ItemInProgress})
	m.checklist.AddItem(input.ChecklistItem{ID: "c5", Label: "Sparkline", Status: input.ItemPending})
	left := agent.NewSubagentPanel(agent.WithSubagentTitle("frontend-agent"), agent.WithSubagentModel("gpt-4.1"))
	left.SetTools([]agent.SubagentTool{{Name: "Read design tokens", Status: agent.ToolCompleted}, {Name: "Build tab interactions", Status: agent.ToolCompleted}})
	left.SetStatus(agent.SubagentCompleted)
	left.SetElapsed(41 * time.Second)
	left.SetTokenCount("5.3k tok")
	left.SetCost("$0.08")
	right := agent.NewSubagentPanel(agent.WithSubagentTitle("qa-agent"), agent.WithSubagentModel("gpt-4o-mini"))
	right.SetTools([]agent.SubagentTool{{Name: "Run terminal build", Status: agent.ToolRunning}, {Name: "Review regressions", Status: agent.ToolCompleted}})
	right.SetStatus(agent.SubagentRunning)
	right.SetElapsed(19 * time.Second)
	right.SetTokenCount("2.1k tok")
	right.SetCost("$0.03")
	m.subagents = agent.NewSubagentGroup(agent.WithSubagentGroupGap(3))
	m.subagents.SetPanels([]*agent.SubagentPanel{left, right})
	m.breadcrumb = nav.NewBreadcrumb(nav.WithSeparator(" / "), nav.WithBreadcrumbItems(
		nav.BreadcrumbItem{ID: "root", Label: "workspace"},
		nav.BreadcrumbItem{ID: "stack", Label: "stack"},
		nav.BreadcrumbItem{ID: "showcase", Label: "examples"},
		nav.BreadcrumbItem{ID: "main", Label: "main.go"},
	))
	m.tree = nav.NewTree([]*nav.TreeNode{
		{Label: "stack", Icon: "📁", Expanded: true, Children: []*nav.TreeNode{
			{Label: "tui", Icon: "📁", Expanded: true, Children: []*nav.TreeNode{
				{Label: "v2", Icon: "📁", Expanded: true, Children: []*nav.TreeNode{
					{Label: "examples", Icon: "📁", Expanded: true, Children: []*nav.TreeNode{
						{Label: "showcase", Icon: "📁", Expanded: true, Children: []*nav.TreeNode{{Label: "main.go", Icon: "📄"}}},
					}},
					{Label: "progressbar.go", Icon: "📄"},
					{Label: "floatingpalette.go", Icon: "📄"},
				}},
			}},
		}},
	}, nav.WithTreeShowIcons(true), nav.WithTreeShowGuides(true), nav.WithTreeIndentSize(2))
	m.fileCardRow = display.NewFileCardRow(
		display.NewFileCard("examples/showcase/main.go", 148, 30),
		display.NewFileCard("chart/progressbar.go", 383, 0),
		display.NewFileCard("input/floatingpalette.go", 469, 0),
	)
}
func (m model) Init() tea.Cmd {
	return tea.Batch(m.toast.Init(), m.subagents.Init())
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			m.toast.Push("Toast demo: components updated", display.ToastSuccess)
		case "ctrl+p":
			if m.activeTab == "input" {
				m.palette.Toggle()
				if m.palette.IsVisible() {
					m.toast.Push("FloatingPalette opened", display.ToastInfo)
				} else {
					m.toast.Push("FloatingPalette closed", display.ToastWarning)
				}
			}
		}
	case input.PaletteSelectMsg:
		m.toast.Push("Selected: "+msg.Command.Name, display.ToastSuccess)
	case input.PaletteDismissMsg:
		m.toast.Push("Palette dismissed", display.ToastInfo)
	}
	if updated, cmd := m.tabBar.Update(msg); updated != nil {
		if tb, ok := updated.(*nav.TabBar); ok {
			m.tabBar = tb
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if id := m.tabBar.GetActiveID(); id != "" && id != m.activeTab {
		m.activeTab = id
		m.applyFocus()
		m.toast.Push("Switched to "+id, display.ToastInfo)
	}
	if updated, cmd := m.palette.Update(msg); updated != nil {
		if fp, ok := updated.(*input.FloatingPalette); ok {
			m.palette = fp
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	switch m.activeTab {
	case "components":
		if updated, cmd := m.threads.Update(msg); updated != nil {
			if tp, ok := updated.(*agent.ThreadProgress); ok {
				m.threads = tp
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if updated, cmd := m.checklist.Update(msg); updated != nil {
			if cl, ok := updated.(*input.Checklist); ok {
				m.checklist = cl
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case "panels":
		if updated, cmd := m.subagents.Update(msg); updated != nil {
			if sg, ok := updated.(*agent.SubagentGroup); ok {
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
		if t, ok := updated.(*display.Toast); ok {
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
	return strings.Join([]string{header, m.tabBar.View(), m.tabContent(), m.toast.View()}, "\n")
}
func main() {
	p := tea.NewProgram(initialModel(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
