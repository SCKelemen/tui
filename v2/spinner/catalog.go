package spinner

import "sort"

// --- Braille spinners ---

// Braille is the standard braille dot spinner (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏).
// This is the recommended default for agent/CLI UIs.
var Braille = Spinner{Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}}

// BrailleTwoPoint is a compact two-point braille spinner.
var BrailleTwoPoint = Spinner{Frames: []string{"⠂", "⠒", "⠐", "⠰", "⠠", "⠤", "⠄", "⠆"}}

// --- Dot spinners ---

// Dots is a simple growing dot pattern.
var Dots = Spinner{Frames: []string{"⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"}}

// Dots2 is an alternative dot spinner.
var Dots2 = Spinner{Frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}}

// Dots3 is a compact vertical dot spinner.
var Dots3 = Spinner{Frames: []string{"⠋", "⠙", "⠚", "⠒", "⠂", "⠂", "⠒", "⠲", "⠴", "⠦", "⠖", "⠒", "⠐", "⠐", "⠒", "⠓", "⠋"}}

// --- Circle spinners ---

// CircleQuarters is the original tui spinner (◐◓◑◒).
var CircleQuarters = Spinner{Frames: []string{"◐", "◓", "◑", "◒"}}

// CircleHalf rotates a half-filled circle.
var CircleHalf = Spinner{Frames: []string{"◐", "◓", "◑", "◒"}}

// Circle uses open/filled circle states.
var Circle = Spinner{Frames: []string{"◡", "⊙", "◠"}}

// --- Line spinners ---

// Line is a rotating line (|/-\).
var Line = Spinner{Frames: []string{"|", "/", "-", "\\"}}

// SimpleDots is a growing ellipsis.
var SimpleDots = Spinner{Frames: []string{".", "..", "...", ".."}}

// --- Arrow spinners ---

// Arrow rotates arrow directions.
var Arrow = Spinner{Frames: []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}}

// Arrow2 uses bold arrow variants.
var Arrow2 = Spinner{Frames: []string{"⬆", "↗", "➡", "↘", "⬇", "↙", "⬅", "↖"}}

// --- Bounce spinners ---

// Bounce bounces a ball between brackets.
var Bounce = Spinner{Frames: []string{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"}}

// BouncingBar bounces a bar across a track.
var BouncingBar = Spinner{Frames: []string{
	"[    ]",
	"[=   ]",
	"[==  ]",
	"[=== ]",
	"[ ===]",
	"[  ==]",
	"[   =]",
	"[    ]",
	"[   =]",
	"[  ==]",
	"[ ===]",
	"[====]",
	"[=== ]",
	"[==  ]",
	"[=   ]",
}}

// --- Bar/block spinners ---

// GrowVertical fills a vertical bar.
var GrowVertical = Spinner{Frames: []string{"▁", "▃", "▄", "▅", "▆", "▇", "▆", "▅", "▄", "▃"}}

// GrowHorizontal fills a horizontal bar.
var GrowHorizontal = Spinner{Frames: []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "▊", "▋", "▌", "▍", "▎"}}

// Pulse pulses a block.
var Pulse = Spinner{Frames: []string{"█", "▓", "▒", "░", "▒", "▓"}}

// --- Aesthetic / decorative ---

// Moon cycles through moon phases.
var Moon = Spinner{Frames: []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}}

// Earth spins the globe.
var Earth = Spinner{Frames: []string{"🌍", "🌎", "🌏"}}

// Clock cycles through clock faces.
var Clock = Spinner{Frames: []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}}

// Star twinkles.
var Star = Spinner{Frames: []string{"✶", "✸", "✹", "✺", "✹", "✷"}}

// Hamburger is a cute toggle.
var Hamburger = Spinner{Frames: []string{"☱", "☲", "☴"}}

// --- Minimal / low-profile ---

// Pipe toggles pipe characters.
var Pipe = Spinner{Frames: []string{"┤", "┘", "┴", "└", "├", "┌", "┬", "┐"}}

// Toggle simple on/off.
var Toggle = Spinner{Frames: []string{"⊶", "⊷"}}

// Toggle2 uses boxes.
var Toggle2 = Spinner{Frames: []string{"▫", "▪"}}

// Ellipsis is a simple three-dot cycle.
var Ellipsis = Spinner{Frames: []string{"   ", ".  ", ".. ", "..."}}

// --- Arc spinners ---

// Arc rotates an arc character.
var Arc = Spinner{Frames: []string{"◜", "◠", "◝", "◞", "◡", "◟"}}

// --- Vertical bars (like Slate's status indicator) ---

// VerticalBars pulses three vertical bars left-to-right.
var VerticalBars = Spinner{Frames: []string{
	"▎  ",
	"▌  ",
	"█  ",
	"▊▎ ",
	"▌▌ ",
	" █ ",
	" ▊▎",
	" ▌▌",
	"  █",
	"  ▊",
	"  ▌",
	"  ▎",
}}

// --- Mini segmented progress (like Slate's mini bars) ---

// MiniBar sweeps small segments across a track.
var MiniBar = Spinner{Frames: []string{
	"▰▱▱▱▱",
	"▰▰▱▱▱",
	"▰▰▰▱▱",
	"▱▰▰▰▱",
	"▱▱▰▰▰",
	"▱▱▱▰▰",
	"▱▱▱▱▰",
}}

// --- Named lookup ---

// All returns a map of all named spinners.
func All() map[string]Spinner {
	return map[string]Spinner{
		"braille":        Braille,
		"braille2":       BrailleTwoPoint,
		"dots":           Dots,
		"dots2":          Dots2,
		"dots3":          Dots3,
		"circle":         Circle,
		"circleQuarters": CircleQuarters,
		"line":           Line,
		"simpleDots":     SimpleDots,
		"arrow":          Arrow,
		"arrow2":         Arrow2,
		"bounce":         Bounce,
		"bouncingBar":    BouncingBar,
		"growVertical":   GrowVertical,
		"growHorizontal": GrowHorizontal,
		"pulse":          Pulse,
		"moon":           Moon,
		"earth":          Earth,
		"clock":          Clock,
		"star":           Star,
		"hamburger":      Hamburger,
		"pipe":           Pipe,
		"toggle":         Toggle,
		"toggle2":        Toggle2,
		"ellipsis":       Ellipsis,
		"arc":            Arc,
		"verticalBars":   VerticalBars,
		"miniBar":        MiniBar,
	}
}

// ByName returns a spinner by name, or the Braille spinner if not found.
func ByName(name string) Spinner {
	all := All()
	if s, ok := all[name]; ok {
		return s
	}
	return Braille
}

// Names returns all available spinner names sorted.
func Names() []string {
	all := All()
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	// Sort for deterministic output
	sort.Strings(names)
	return names
}
