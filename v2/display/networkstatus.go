package display

import (
	"fmt"
	"sort"
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/SCKelemen/tui/v2/style/design"
	tea "github.com/charmbracelet/bubbletea"
)

// NetworkRequestState is the approval state of one host.
type NetworkRequestState int

const (
	// NetworkRequestPending means request is awaiting approval.
	NetworkRequestPending NetworkRequestState = iota
	// NetworkRequestApproved means host request was approved.
	NetworkRequestApproved
	// NetworkRequestDenied means host request was denied.
	NetworkRequestDenied
)

// NetworkStatus tracks and renders per-host network approval state.
type NetworkStatus struct {
	hosts        map[string]NetworkRequestState
	focused      bool
	designTokens *design.DesignTokens
}

// NetworkStatusOption configures NetworkStatus.
type NetworkStatusOption func(*NetworkStatus)

// WithNetworkStatusDesignTokens applies design tokens.
func WithNetworkStatusDesignTokens(tokens *design.DesignTokens) NetworkStatusOption {
	return func(n *NetworkStatus) {
		if tokens != nil {
			n.designTokens = tokens
		}
	}
}

// NewNetworkStatus creates a compact inline network status model.
func NewNetworkStatus(opts ...NetworkStatusOption) *NetworkStatus {
	n := &NetworkStatus{
		hosts:        make(map[string]NetworkRequestState),
		designTokens: design.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// Init satisfies Bubble Tea model contract.
func (n *NetworkStatus) Init() tea.Cmd { return nil }

// Update currently handles no external messages but satisfies model contract.
func (n *NetworkStatus) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	return n, nil
}

// SetHostStatus updates one host state.
func (n *NetworkStatus) SetHostStatus(host string, state NetworkRequestState) {
	h := strings.TrimSpace(host)
	if h == "" {
		return
	}
	n.hosts[h] = state
}

// View renders compact inline host status, suitable for status bars.
func (n *NetworkStatus) View() string {
	if len(n.hosts) == 0 {
		return "net: idle"
	}

	hosts := make([]string, 0, len(n.hosts))
	for host := range n.hosts {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	parts := make([]string, 0, len(hosts))
	for _, host := range hosts {
		parts = append(parts, n.renderHost(host, n.hosts[host]))
	}
	return "net: " + strings.Join(parts, " ")
}

// Focus marks focus state.
func (n *NetworkStatus) Focus() { n.focused = true }

// Blur marks blur state.
func (n *NetworkStatus) Blur() { n.focused = false }

// Focused reports focus state.
func (n *NetworkStatus) Focused() bool { return n.focused }

func (n *NetworkStatus) renderHost(host string, state NetworkRequestState) string {
	symbol := "…"
	color := style.ANSIYellow
	if n.designTokens != nil {
		if v := style.Fg(n.designTokens.PendingColor); v != "" {
			color = v
		}
	}

	switch state {
	case NetworkRequestApproved:
		symbol = "✓"
		color = style.ANSIGreen
		if n.designTokens != nil {
			if v := style.Fg(n.designTokens.SuccessBright); v != "" {
				color = v
			}
		}
	case NetworkRequestDenied:
		symbol = "✗"
		color = style.ANSIRed
		if n.designTokens != nil {
			if v := style.Fg(n.designTokens.ErrorBright); v != "" {
				color = v
			}
		}
	}

	return fmt.Sprintf("%s%s%s%s", color, host, symbol, style.ANSIReset)
}

var _ tui.Component = (*NetworkStatus)(nil)
