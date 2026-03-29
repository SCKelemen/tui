package display

import "github.com/charmbracelet/lipgloss"

// Link renders terminal hyperlinks using the OSC 8 escape sequence.
type Link struct {
	URL   string
	Text  string
	color string
}

// LinkOption configures a Link.
type LinkOption func(*Link)

// NewLink creates a new Link with optional configuration.
func NewLink(url, text string, opts ...LinkOption) *Link {
	l := &Link{
		URL:  url,
		Text: text,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// WithLinkColor sets the rendered text color for the link.
func WithLinkColor(color string) LinkOption {
	return func(l *Link) {
		l.color = color
	}
}

// View renders the link as an OSC 8 hyperlink with optional color styling.
func (l *Link) View() string {
	text := l.Text
	if l.color != "" {
		text = lipgloss.NewStyle().Foreground(lipgloss.Color(l.color)).Render(text)
	}

	return "\033]8;;" + l.URL + "\033\\" + text + "\033]8;;\033\\"
}

// RenderLink renders an OSC 8 hyperlink for quick one-off use.
func RenderLink(url, text string) string {
	return NewLink(url, text).View()
}
