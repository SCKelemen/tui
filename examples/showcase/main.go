package main

import (
	"strings"

	tui "github.com/SCKelemen/tui"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	tabBar    *tui.TabBar
	threads   *tui.ThreadProgress
	tasks     *tui.Checklist
	toast     *tui.Toast
	activeTab int
	width     int
	height    int
}

func initialModel() model {
	m := model{
		tabBar: tui.NewTabBar(
			tui.WithTabs(
				tui.Tab{ID: "threads", Label: "Threads"},
				tui.Tab{ID: "tasks", Label: "Checklist"},
			),
		),
		threads: tui.NewThreadProgress(),
		tasks: tui.NewChecklist(
			tui.WithChecklistTitle("Demo Tasks"),
			tui.WithShowProgress(true),
		),
		toast:     tui.NewToast(),
		activeTab: 0,
	}
	m.tabBar.Focus()
	m.threads.Focus()
	m.tasks.Blur()
	return m
}

func (m model) Init() tea.Cmd {
	m.threads.AddThread("t1", "Load configuration")
	m.threads.UpdateThread("t1", tui.ThreadCompleted)
	m.threads.AppendOutput("t1", "Configuration loaded")

	m.threads.AddThread("t2", "Start workers")
	m.threads.UpdateThread("t2", tui.ThreadRunning)
	m.threads.AppendOutput("t2", "2/4 workers ready")

	m.threads.AddThread("t3", "Run health checks")
	m.threads.UpdateThread("t3", tui.ThreadFailed)
	m.threads.AppendOutput("t3", "health endpoint timed out")

	m.tasks.AddItem(tui.ChecklistItem{ID: "c1", Label: "Validate input", Status: tui.ItemDone})
	m.tasks.AddItem(tui.ChecklistItem{ID: "c2", Label: "Fetch secrets", Status: tui.ItemDone})
	m.tasks.AddItem(tui.ChecklistItem{ID: "c3", Label: "Connect to DB", Status: tui.ItemInProgress})
	m.tasks.AddItem(tui.ChecklistItem{ID: "c4", Label: "Run migrations", Status: tui.ItemPending})
	m.tasks.AddItem(tui.ChecklistItem{ID: "c5", Label: "Start API server", Status: tui.ItemPending})

	m.toast.Push("Press TAB to switch panels", tui.ToastInfo)
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % 2
			if m.activeTab == 0 {
				m.tabBar.SetActive("threads")
				m.threads.Focus()
				m.tasks.Blur()
				m.toast.Push("Switched to Threads", tui.ToastInfo)
			} else {
				m.tabBar.SetActive("tasks")
				m.tasks.Focus()
				m.threads.Blur()
				m.toast.Push("Switched to Checklist", tui.ToastInfo)
			}
		}
	}

	if comp, cmd := m.toast.Update(msg); cmd != nil {
		if t, ok := comp.(*tui.Toast); ok {
			m.toast = t
		}
		cmds = append(cmds, cmd)
	}

	if comp, cmd := m.tabBar.Update(msg); cmd != nil {
		if tb, ok := comp.(*tui.TabBar); ok {
			m.tabBar = tb
		}
		cmds = append(cmds, cmd)
	}

	if m.activeTab == 0 {
		if comp, cmd := m.threads.Update(msg); cmd != nil {
			if tp, ok := comp.(*tui.ThreadProgress); ok {
				m.threads = tp
			}
			cmds = append(cmds, cmd)
		}
	} else {
		if comp, cmd := m.tasks.Update(msg); cmd != nil {
			if c, ok := comp.(*tui.Checklist); ok {
				m.tasks = c
			}
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	active := m.threads.View()
	if m.activeTab == 1 {
		active = m.tasks.View()
	}
	return strings.Join([]string{
		m.tabBar.View(),
		active,
		m.toast.View(),
	}, "\n")
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
