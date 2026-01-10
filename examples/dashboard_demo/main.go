package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

// tickMsg is sent on every tick to update metrics
type tickMsg time.Time

// model holds the application state
type model struct {
	dashboard *tui.Dashboard
	width     int
	height    int
	startTime time.Time
}

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// generateTrendData generates realistic trend data
func generateTrendData(points int, baseValue float64, volatility float64) []float64 {
	trend := make([]float64, points)
	value := baseValue

	for i := 0; i < points; i++ {
		// Add some randomness
		change := (rand.Float64() - 0.5) * volatility
		value += change

		// Keep value positive
		if value < 0 {
			value = 0
		}

		trend[i] = value
	}

	return trend
}

// initialModel creates the initial model with sample stat cards
func initialModel() model {
	rand.Seed(time.Now().UnixNano())

	// Create stat cards with different metrics
	cpuCard := tui.NewStatCard(
		tui.WithTitle("CPU Usage"),
		tui.WithValue("42%"),
		tui.WithSubtitle("8 cores active"),
		tui.WithChange(5, 13.5),
		tui.WithColor("#2196F3"),
		tui.WithTrendColor("#4CAF50"),
		tui.WithTrend(generateTrendData(30, 40, 10)),
	)

	memoryCard := tui.NewStatCard(
		tui.WithTitle("Memory"),
		tui.WithValue("8.2 GB"),
		tui.WithSubtitle("of 16 GB total"),
		tui.WithChange(-200, -2.4),
		tui.WithColor("#9C27B0"),
		tui.WithTrendColor("#E91E63"),
		tui.WithTrend(generateTrendData(30, 8000, 500)),
	)

	networkCard := tui.NewStatCard(
		tui.WithTitle("Network"),
		tui.WithValue("125 Mbps"),
		tui.WithSubtitle("Download speed"),
		tui.WithChange(25, 25.0),
		tui.WithColor("#FF9800"),
		tui.WithTrendColor("#FFC107"),
		tui.WithTrend(generateTrendData(30, 100, 30)),
	)

	diskCard := tui.NewStatCard(
		tui.WithTitle("Disk I/O"),
		tui.WithValue("450 MB/s"),
		tui.WithSubtitle("Read/Write"),
		tui.WithChange(0, 0.0),
		tui.WithColor("#00BCD4"),
		tui.WithTrendColor("#03A9F4"),
		tui.WithTrend(generateTrendData(30, 450, 100)),
	)

	activeUsersCard := tui.NewStatCard(
		tui.WithTitle("Active Users"),
		tui.WithValue("1,247"),
		tui.WithSubtitle("Online now"),
		tui.WithChange(127, 11.3),
		tui.WithColor("#4CAF50"),
		tui.WithTrendColor("#8BC34A"),
		tui.WithTrend(generateTrendData(30, 1100, 200)),
	)

	requestsCard := tui.NewStatCard(
		tui.WithTitle("Requests"),
		tui.WithValue("45.2k"),
		tui.WithSubtitle("per minute"),
		tui.WithChange(3200, 7.6),
		tui.WithColor("#F44336"),
		tui.WithTrendColor("#FF5722"),
		tui.WithTrend(generateTrendData(30, 42000, 5000)),
	)

	errorRateCard := tui.NewStatCard(
		tui.WithTitle("Error Rate"),
		tui.WithValue("0.23%"),
		tui.WithSubtitle("Last hour"),
		tui.WithChange(-10, -4.2),
		tui.WithColor("#FF5722"),
		tui.WithTrendColor("#F44336"),
		tui.WithTrend(generateTrendData(30, 0.25, 0.05)),
	)

	latencyCard := tui.NewStatCard(
		tui.WithTitle("Avg Latency"),
		tui.WithValue("42ms"),
		tui.WithSubtitle("p95: 125ms"),
		tui.WithChange(-8, -16.0),
		tui.WithColor("#3F51B5"),
		tui.WithTrendColor("#2196F3"),
		tui.WithTrend(generateTrendData(30, 45, 10)),
	)

	uptimeCard := tui.NewStatCard(
		tui.WithTitle("Uptime"),
		tui.WithValue("99.97%"),
		tui.WithSubtitle("30 days"),
		tui.WithChange(0, 0.02),
		tui.WithColor("#009688"),
		tui.WithTrendColor("#00BCD4"),
		tui.WithTrend(generateTrendData(30, 99.95, 0.1)),
	)

	// Create dashboard with responsive layout
	dashboard := tui.NewDashboard(
		tui.WithDashboardTitle("System Metrics Dashboard"),
		tui.WithResponsiveLayout(30), // Min card width of 30 characters
		tui.WithGap(2),
		tui.WithCards(
			cpuCard,
			memoryCard,
			networkCard,
			diskCard,
			activeUsersCard,
			requestsCard,
			errorRateCard,
			latencyCard,
			uptimeCard,
		),
	)

	// Enable keyboard navigation
	dashboard.Focus()

	return model{
		dashboard: dashboard,
		startTime: time.Now(),
	}
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
	)
}

// Update handles messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle quit/refresh first
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			// Refresh - regenerate data
			return initialModel(), tickCmd()
		default:
			// Forward other keys to dashboard for navigation
			m.dashboard.Update(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Forward to dashboard
		m.dashboard.Update(msg)

	case tickMsg:
		// Update metrics (in a real app, fetch from data source)
		m.updateMetrics()

		return m, tickCmd()
	}

	return m, nil
}

// updateMetrics simulates updating metric values
func (m *model) updateMetrics() {
	cards := m.dashboard.GetCards()

	// Simulate metric changes
	if len(cards) > 0 {
		// CPU Usage
		cpuValue := 35 + rand.Float64()*30
		cards[0] = tui.NewStatCard(
			tui.WithTitle("CPU Usage"),
			tui.WithValue(fmt.Sprintf("%.0f%%", cpuValue)),
			tui.WithSubtitle("8 cores active"),
			tui.WithChange(int(rand.Float64()*10-5), rand.Float64()*20-10),
			tui.WithColor("#2196F3"),
			tui.WithTrend(generateTrendData(30, cpuValue, 10)),
		)
	}

	if len(cards) > 4 {
		// Active Users
		users := 1000 + rand.Intn(500)
		cards[4] = tui.NewStatCard(
			tui.WithTitle("Active Users"),
			tui.WithValue(fmt.Sprintf("%d", users)),
			tui.WithSubtitle("Online now"),
			tui.WithChange(int(rand.Float64()*200-100), rand.Float64()*20-10),
			tui.WithColor("#4CAF50"),
			tui.WithTrend(generateTrendData(30, float64(users), 200)),
		)
	}

	m.dashboard.SetCards(cards)
}

// View renders the view
func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	s := m.dashboard.View()

	// Add footer with help
	s += "\n"
	elapsed := time.Since(m.startTime).Round(time.Second)
	footer := fmt.Sprintf("Uptime: %s | ←→↑↓ navigate | enter select | esc deselect | r refresh | q quit", elapsed)
	s += footer

	return s
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// abs returns absolute value
func abs(x float64) float64 {
	return math.Abs(x)
}
