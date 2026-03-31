package nav

import "sync"

// ViewPool is a generic recycling pool for reusable values.
type ViewPool[T any] struct {
	available []T
	maxSize   int
	mu        sync.Mutex
}

// NewViewPool creates a new bounded view pool.
func NewViewPool[T any](maxSize int) *ViewPool[T] {
	if maxSize < 0 {
		maxSize = 0
	}
	return &ViewPool[T]{
		available: make([]T, 0, maxSize),
		maxSize:   maxSize,
	}
}

// Get pops a value from the pool.
func (p *ViewPool[T]) Get() (T, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var zero T
	if len(p.available) == 0 {
		return zero, false
	}

	last := len(p.available) - 1
	v := p.available[last]
	p.available = p.available[:last]
	return v, true
}

// Put returns a value to the pool, dropping it when full.
func (p *ViewPool[T]) Put(v T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.maxSize == 0 || len(p.available) >= p.maxSize {
		return
	}
	p.available = append(p.available, v)
}

// Clear removes all pooled values.
func (p *ViewPool[T]) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.available = p.available[:0]
}

// Size reports current pool size.
func (p *ViewPool[T]) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.available)
}
