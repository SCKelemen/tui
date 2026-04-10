package tui

import tea "github.com/charmbracelet/bubbletea"

// Focusable describes a renderable item that can participate in focus management.
type Focusable interface {
	ID() string
	IsFocusable() bool
	SetFocused(focused bool)
	IsFocused() bool
	TabIndex() int
}

// FocusManager manages focus order and keyboard navigation for a set of focusable items.
type FocusManager struct {
	focusables   []Focusable
	currentIndex int
	trapFocus    bool
}

// NewFocusManager creates a new focus manager.
func NewFocusManager() *FocusManager {
	return &FocusManager{currentIndex: -1}
}

// Register adds a focusable item in tab order.
func (m *FocusManager) Register(f Focusable) {
	if m == nil || f == nil {
		return
	}

	m.normalizeCurrentIndex()
	m.Unregister(f.ID())

	insertAt := len(m.focusables)
	for i, existing := range m.focusables {
		if f.TabIndex() < existing.TabIndex() {
			insertAt = i
			break
		}
	}

	m.focusables = append(m.focusables, nil)
	copy(m.focusables[insertAt+1:], m.focusables[insertAt:])
	m.focusables[insertAt] = f

	if m.currentIndex >= insertAt {
		m.currentIndex++
	}

	if f.IsFocused() {
		m.FocusByIndex(insertAt)
	}
}

// Unregister removes a focusable item by ID.
func (m *FocusManager) Unregister(id string) {
	if m == nil || id == "" || len(m.focusables) == 0 {
		return
	}

	m.normalizeCurrentIndex()

	for i, f := range m.focusables {
		if f == nil || f.ID() != id {
			continue
		}

		wasCurrent := i == m.currentIndex
		f.SetFocused(false)

		m.focusables = append(m.focusables[:i], m.focusables[i+1:]...)

		switch {
		case len(m.focusables) == 0:
			m.currentIndex = -1
		case wasCurrent:
			m.currentIndex = -1
			m.focusFrom(i, 1)
		case i < m.currentIndex:
			m.currentIndex--
		}

		return
	}
}

// FocusNext moves focus forward, wrapping around the registered items.
func (m *FocusManager) FocusNext() {
	if m == nil || len(m.focusables) == 0 {
		return
	}

	m.normalizeCurrentIndex()

	start := m.currentIndex + 1
	if m.currentIndex < 0 {
		start = 0
	}

	m.focusFrom(start, 1)
}

// FocusPrev moves focus backward, wrapping around the registered items.
func (m *FocusManager) FocusPrev() {
	if m == nil || len(m.focusables) == 0 {
		return
	}

	m.normalizeCurrentIndex()

	start := m.currentIndex - 1
	if m.currentIndex < 0 {
		start = len(m.focusables) - 1
	}

	m.focusFrom(start, -1)
}

// FocusByID focuses an item by ID.
func (m *FocusManager) FocusByID(id string) bool {
	if m == nil || id == "" {
		return false
	}

	for i, f := range m.focusables {
		if f != nil && f.ID() == id {
			return m.FocusByIndex(i)
		}
	}

	return false
}

// FocusByIndex focuses an item by index.
func (m *FocusManager) FocusByIndex(idx int) bool {
	if m == nil || idx < 0 || idx >= len(m.focusables) {
		return false
	}

	target := m.focusables[idx]
	if target == nil || !target.IsFocusable() {
		return false
	}

	m.normalizeCurrentIndex()

	if m.currentIndex >= 0 && m.currentIndex < len(m.focusables) {
		current := m.focusables[m.currentIndex]
		if current != nil && current != target {
			current.SetFocused(false)
		}
	}

	for i, f := range m.focusables {
		if f == nil || i == idx {
			continue
		}
		if f.IsFocused() {
			f.SetFocused(false)
		}
	}

	target.SetFocused(true)
	m.currentIndex = idx
	return true
}

// Current returns the currently focused item.
func (m *FocusManager) Current() Focusable {
	if m == nil {
		return nil
	}

	m.normalizeCurrentIndex()
	if m.currentIndex < 0 || m.currentIndex >= len(m.focusables) {
		return nil
	}

	return m.focusables[m.currentIndex]
}

// CurrentID returns the ID of the currently focused item.
func (m *FocusManager) CurrentID() string {
	current := m.Current()
	if current == nil {
		return ""
	}

	return current.ID()
}

// HandleKeyMsg handles Tab and Shift+Tab focus navigation.
func (m *FocusManager) HandleKeyMsg(msg tea.KeyMsg) bool {
	if m == nil {
		return false
	}

	switch msg.Type {
	case tea.KeyTab:
		m.FocusNext()
		return true
	case tea.KeyShiftTab:
		m.FocusPrev()
		return true
	}

	switch msg.String() {
	case "tab":
		m.FocusNext()
		return true
	case "shift+tab":
		m.FocusPrev()
		return true
	}

	return false
}

// SetTrapFocus enables or disables focus trapping for this manager.
func (m *FocusManager) SetTrapFocus(enabled bool) {
	if m == nil {
		return
	}

	m.trapFocus = enabled
}

// Count returns the number of registered items.
func (m *FocusManager) Count() int {
	if m == nil {
		return 0
	}

	return len(m.focusables)
}

func (m *FocusManager) focusFrom(start, step int) bool {
	if len(m.focusables) == 0 {
		m.currentIndex = -1
		return false
	}

	count := len(m.focusables)
	for i := 0; i < count; i++ {
		idx := start + (i * step)
		for idx < 0 {
			idx += count
		}
		idx %= count

		candidate := m.focusables[idx]
		if candidate == nil || !candidate.IsFocusable() {
			continue
		}

		return m.FocusByIndex(idx)
	}

	m.currentIndex = -1
	return false
}

func (m *FocusManager) normalizeCurrentIndex() {
	if m == nil {
		return
	}

	if m.currentIndex >= 0 && m.currentIndex < len(m.focusables) {
		current := m.focusables[m.currentIndex]
		if current != nil && current.IsFocused() {
			return
		}
	}

	for i, f := range m.focusables {
		if f != nil && f.IsFocused() {
			m.currentIndex = i
			return
		}
	}

	m.currentIndex = -1
}

// FocusGroup scopes focus management to a nested container such as a modal or dialog.
type FocusGroup struct {
	manager          *FocusManager
	parent           *FocusManager
	active           bool
	previousParentID string
}

// NewFocusGroup creates a new scoped focus group with its own sub-manager.
func NewFocusGroup(parent *FocusManager) *FocusGroup {
	return &FocusGroup{
		manager: NewFocusManager(),
		parent:  parent,
	}
}

// Manager returns the group's sub-manager.
func (g *FocusGroup) Manager() *FocusManager {
	if g == nil {
		return nil
	}

	if g.manager == nil {
		g.manager = NewFocusManager()
	}

	return g.manager
}

// Enter activates the group and focuses the first available item inside it.
func (g *FocusGroup) Enter() {
	if g == nil {
		return
	}

	g.active = true
	g.Manager().SetTrapFocus(true)

	if g.parent != nil {
		g.previousParentID = g.parent.CurrentID()
		g.parent.SetTrapFocus(true)
	}

	if g.manager.Current() == nil {
		g.manager.FocusNext()
	}
}

// Exit deactivates the group and restores focus handling to the parent manager.
func (g *FocusGroup) Exit() {
	if g == nil {
		return
	}

	g.active = false

	if g.manager != nil {
		g.manager.SetTrapFocus(false)
	}

	if g.parent != nil {
		g.parent.SetTrapFocus(false)
		if g.previousParentID != "" {
			g.parent.FocusByID(g.previousParentID)
		}
	}
}

// IsActive reports whether the group is active.
func (g *FocusGroup) IsActive() bool {
	if g == nil {
		return false
	}

	return g.active
}

// FocusRing provides simple circular focus navigation for a list of IDs.
type FocusRing struct {
	items        []string
	currentIndex int
}

// NewFocusRing creates a new focus ring.
func NewFocusRing(items ...string) *FocusRing {
	return &FocusRing{
		items:        append([]string(nil), items...),
		currentIndex: -1,
	}
}

// Next advances to the next item and returns its ID.
func (r *FocusRing) Next() string {
	if r == nil || len(r.items) == 0 {
		return ""
	}

	r.normalizeCurrentIndex()
	r.currentIndex = (r.currentIndex + 1) % len(r.items)
	return r.items[r.currentIndex]
}

// Prev moves to the previous item and returns its ID.
func (r *FocusRing) Prev() string {
	if r == nil || len(r.items) == 0 {
		return ""
	}

	r.normalizeCurrentIndex()
	r.currentIndex--
	if r.currentIndex < 0 {
		r.currentIndex = len(r.items) - 1
	}

	return r.items[r.currentIndex]
}

// Current returns the current item ID.
func (r *FocusRing) Current() string {
	if r == nil || len(r.items) == 0 {
		return ""
	}

	r.normalizeCurrentIndex()
	if r.currentIndex < 0 {
		return ""
	}

	return r.items[r.currentIndex]
}

// SetCurrent sets the current item by ID if it exists in the ring.
func (r *FocusRing) SetCurrent(id string) {
	if r == nil {
		return
	}

	for i, item := range r.items {
		if item == id {
			r.currentIndex = i
			return
		}
	}
}

func (r *FocusRing) normalizeCurrentIndex() {
	if len(r.items) == 0 {
		r.currentIndex = -1
		return
	}

	if r.currentIndex < -1 || r.currentIndex >= len(r.items) {
		r.currentIndex = -1
	}
}
