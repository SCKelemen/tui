package style

import (
  "github.com/SCKelemen/text"
)

// terminal is a shared Text instance configured for terminal rendering.
var terminal = text.NewTerminal()

// StringWidth returns the display width of a string in terminal cells,
// correctly handling CJK characters, emoji, and combining marks.
func StringWidth(s string) int {
  return int(terminal.Width(s))
}

// Truncate truncates a string to fit within maxWidth terminal cells,
// appending ellipsis if truncation occurs. Uses proper Unicode width.
func Truncate(s string, maxWidth int, ellipsis string) string {
  if maxWidth <= 0 { return "" }
  w := int(terminal.Width(s))
  if w <= maxWidth { return s }
  return terminal.Truncate(s, text.TruncateOptions{
    MaxWidth: float64(maxWidth),
    Ellipsis: ellipsis,
    Strategy: text.TruncateEnd,
  })
}

// TruncateMiddle truncates from the middle, preserving start and end.
func TruncateMiddle(s string, maxWidth int, ellipsis string) string {
  if maxWidth <= 0 { return "" }
  w := int(terminal.Width(s))
  if w <= maxWidth { return s }
  return terminal.Truncate(s, text.TruncateOptions{
    MaxWidth: float64(maxWidth),
    Ellipsis: ellipsis,
    Strategy: text.TruncateMiddle,
  })
}

// ElidePath smartly shortens a file path to fit within maxWidth.
func ElidePath(path string, maxWidth int) string {
  if maxWidth <= 0 { return "" }
  return terminal.ElidePath(path, float64(maxWidth))
}

// ElideURL smartly shortens a URL to fit within maxWidth.
func ElideURL(rawURL string, maxWidth int) string {
  if maxWidth <= 0 { return "" }
  return terminal.ElideURL(rawURL, float64(maxWidth))
}

// Pad pads a string to exactly the given width with spaces.
func Pad(s string, width int) string {
  w := int(terminal.Width(s))
  if w >= width { return s }
  pad := make([]byte, width-w)
  for i := range pad { pad[i] = ' ' }
  return s + string(pad)
}
