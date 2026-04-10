package tui

import (
	"bytes"
	"math/bits"
	"strings"
	"unicode/utf8"
)

// MaxLeafSize controls the maximum number of bytes stored in a rope leaf.
var MaxLeafSize = 512

// Rope is a persistent rope for efficient text editing.
//
// Internal nodes store left and right subtrees plus the weight of the left
// subtree. Leaf nodes store raw bytes in text.
type Rope struct {
	left   *Rope
	right  *Rope
	weight int

	text   []byte
	length int

	depth      int
	lineBreaks int
}

// RopeIterator iterates over a byte range in a rope without materializing the
// entire string.
type RopeIterator struct {
	stack    []ropeIterFrame
	leaf     *Rope
	leafBase int
	index    int
	pos      int
	limit    int
}

type ropeIterFrame struct {
	node *Rope
	base int
}

// NewRope creates a new rope from text.
func NewRope(text string) *Rope {
	return newRopeFromBytes([]byte(text))
}

// Index returns the byte at offset i.
func (r *Rope) Index(i int) byte {
	if i < 0 || i >= r.Length() {
		panic("rope index out of range")
	}

	if r == nil {
		panic("rope index out of range")
	}
	if r.isLeaf() {
		return r.text[i]
	}
	if i < r.weight {
		return r.left.Index(i)
	}
	return r.right.Index(i - r.weight)
}

// String materializes the full rope text.
func (r *Rope) String() string {
	if r == nil || r.length == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(r.length)
	r.writeTo(&b)
	return b.String()
}

// Length returns the total number of bytes in the rope.
func (r *Rope) Length() int {
	if r == nil {
		return 0
	}
	return r.length
}

// Concat concatenates two ropes.
func (r *Rope) Concat(other *Rope) *Rope {
	joined := concatRopes(r, other)
	return maybeRebalance(joined)
}

// Split splits the rope at index and returns the left and right ropes.
func (r *Rope) Split(index int) (*Rope, *Rope) {
	index = clampRopeIndex(index, 0, r.Length())
	left, right := splitRope(r, index)
	return normalizeRope(left), normalizeRope(right)
}

// Insert inserts text at index.
func (r *Rope) Insert(index int, text string) *Rope {
	if text == "" {
		return normalizeRope(r)
	}

	index = clampRopeIndex(index, 0, r.Length())
	left, right := r.Split(index)
	inserted := NewRope(text)
	return maybeRebalance(concatRopes(concatRopes(left, inserted), right))
}

// Delete removes the byte range [start, end).
func (r *Rope) Delete(start, end int) *Rope {
	length := r.Length()
	start = clampRopeIndex(start, 0, length)
	end = clampRopeIndex(end, 0, length)
	if start >= end {
		return normalizeRope(r)
	}

	left, rest := r.Split(start)
	_, right := rest.Split(end - start)
	return maybeRebalance(concatRopes(left, right))
}

// Substring returns the byte range [start, end) as a string.
func (r *Rope) Substring(start, end int) string {
	length := r.Length()
	start = clampRopeIndex(start, 0, length)
	end = clampRopeIndex(end, 0, length)
	if start >= end {
		return ""
	}

	var b strings.Builder
	b.Grow(end - start)
	it := r.IterRange(start, end)
	for {
		ch, ok := it.Next()
		if !ok {
			break
		}
		b.WriteByte(ch)
	}
	return b.String()
}

// LineCount returns the number of logical lines in the rope.
func (r *Rope) LineCount() int {
	if r == nil || r.length == 0 {
		return 1
	}
	return r.lineBreaks + 1
}

// Line returns the nth line without its trailing newline.
func (r *Rope) Line(n int) string {
	if n < 0 || n >= r.LineCount() {
		return ""
	}
	start := r.LineStart(n)
	end := r.LineEnd(n)
	return r.Substring(start, end)
}

// LineStart returns the byte offset of the start of the nth line.
func (r *Rope) LineStart(n int) int {
	if n <= 0 {
		return 0
	}
	if n >= r.LineCount() {
		return r.Length()
	}
	return r.lineStart(n, 0)
}

// LineEnd returns the byte offset of the end of the nth line, excluding the
// trailing newline.
func (r *Rope) LineEnd(n int) int {
	if n < 0 || n >= r.LineCount() {
		return r.Length()
	}

	start := r.LineStart(n)
	nextStart := r.LineStart(n + 1)
	if nextStart > start && nextStart <= r.Length() && r.Index(nextStart-1) == '\n' {
		return nextStart - 1
	}
	return nextStart
}

// LineAt returns the zero-based line number containing offset.
func (r *Rope) LineAt(offset int) int {
	offset = clampRopeIndex(offset, 0, r.Length())
	return r.countLineBreaksBefore(offset)
}

// Iter returns an iterator over the full rope.
func (r *Rope) Iter() *RopeIterator {
	return r.IterRange(0, r.Length())
}

// IterRange returns an iterator over the byte range [start, end).
func (r *Rope) IterRange(start, end int) *RopeIterator {
	length := r.Length()
	start = clampRopeIndex(start, 0, length)
	end = clampRopeIndex(end, 0, length)

	it := &RopeIterator{
		pos:   start,
		limit: end,
	}
	if start >= end || r == nil {
		return it
	}
	it.seek(r, start)
	return it
}

// Rebalance rebuilds the rope into a balanced tree.
func (r *Rope) Rebalance() *Rope {
	if r == nil || r.length <= MaxLeafSize {
		return normalizeRope(r)
	}

	chunks := make([][]byte, 0, max(1, (r.length+effectiveMaxLeafSize()-1)/effectiveMaxLeafSize()))
	collectLeafChunks(r, &chunks)
	chunks = compactChunks(chunks)
	return buildBalancedFromChunks(chunks)
}

// Depth returns the rope depth.
func (r *Rope) Depth() int {
	if r == nil {
		return 0
	}
	return r.depth
}

// IsBalanced reports whether the rope is roughly height-balanced.
func (r *Rope) IsBalanced() bool {
	if r == nil || r.isLeaf() {
		return true
	}

	leftDepth := r.left.Depth()
	rightDepth := r.right.Depth()
	if leftDepth-rightDepth > 1 || rightDepth-leftDepth > 1 {
		return false
	}
	return r.left.IsBalanced() && r.right.IsBalanced()
}

// Next returns the next byte from the iterator.
func (it *RopeIterator) Next() (byte, bool) {
	for {
		if it.leaf == nil || it.pos >= it.limit {
			return 0, false
		}

		if it.index < len(it.leaf.text) {
			b := it.leaf.text[it.index]
			it.index++
			it.pos++
			if it.pos > it.limit {
				return 0, false
			}
			return b, true
		}

		if !it.advanceLeaf() {
			return 0, false
		}
	}
}

// NextRune returns the next rune from the iterator.
func (it *RopeIterator) NextRune() (rune, bool) {
	b, ok := it.Next()
	if !ok {
		return 0, false
	}
	if b < utf8.RuneSelf {
		return rune(b), true
	}

	buf := []byte{b}
	for len(buf) < utf8.UTFMax && !utf8.FullRune(buf) {
		next, ok := it.Next()
		if !ok {
			break
		}
		buf = append(buf, next)
	}

	rn, _ := utf8.DecodeRune(buf)
	return rn, true
}

func newRopeFromBytes(data []byte) *Rope {
	maxLeafSize := effectiveMaxLeafSize()
	if len(data) == 0 {
		return newLeaf(nil)
	}
	if len(data) <= maxLeafSize {
		return newLeaf(data)
	}

	chunks := make([][]byte, 0, (len(data)+maxLeafSize-1)/maxLeafSize)
	for len(data) > 0 {
		size := min(len(data), maxLeafSize)
		chunks = append(chunks, append([]byte(nil), data[:size]...))
		data = data[size:]
	}
	return buildBalancedFromChunks(chunks)
}

func newLeaf(data []byte) *Rope {
	leaf := &Rope{text: append([]byte(nil), data...)}
	leaf.refresh()
	return leaf
}

func buildBalancedFromChunks(chunks [][]byte) *Rope {
	if len(chunks) == 0 {
		return newLeaf(nil)
	}
	if len(chunks) == 1 {
		return newLeaf(chunks[0])
	}

	mid := len(chunks) / 2
	left := buildBalancedFromChunks(chunks[:mid])
	right := buildBalancedFromChunks(chunks[mid:])
	return concatRopes(left, right)
}

func concatRopes(left, right *Rope) *Rope {
	if left == nil || left.length == 0 {
		return normalizeRope(right)
	}
	if right == nil || right.length == 0 {
		return normalizeRope(left)
	}

	node := &Rope{
		left:  left,
		right: right,
	}
	node.refresh()
	return node
}

func splitRope(r *Rope, index int) (*Rope, *Rope) {
	if r == nil {
		empty := newLeaf(nil)
		return empty, empty
	}
	if index <= 0 {
		return newLeaf(nil), r
	}
	if index >= r.Length() {
		return r, newLeaf(nil)
	}
	if r.isLeaf() {
		left := newLeaf(r.text[:index])
		right := newLeaf(r.text[index:])
		return left, right
	}

	if index < r.weight {
		leftLeft, leftRight := splitRope(r.left, index)
		return leftLeft, concatRopes(leftRight, r.right)
	}
	if index == r.weight {
		return normalizeRope(r.left), normalizeRope(r.right)
	}

	rightLeft, rightRight := splitRope(r.right, index-r.weight)
	return concatRopes(r.left, rightLeft), rightRight
}

func (r *Rope) isLeaf() bool {
	if r == nil {
		return true
	}
	return r.left == nil && r.right == nil
}

func (r *Rope) refresh() {
	if r == nil {
		return
	}
	if r.isLeaf() {
		r.length = len(r.text)
		r.weight = r.length
		r.depth = 1
		r.lineBreaks = countLineBreaks(r.text)
		return
	}

	leftLen := 0
	rightLen := 0
	leftDepth := 0
	rightDepth := 0
	leftBreaks := 0
	rightBreaks := 0
	if r.left != nil {
		leftLen = r.left.length
		leftDepth = r.left.depth
		leftBreaks = r.left.lineBreaks
	}
	if r.right != nil {
		rightLen = r.right.length
		rightDepth = r.right.depth
		rightBreaks = r.right.lineBreaks
	}

	r.weight = leftLen
	r.length = leftLen + rightLen
	r.depth = max(leftDepth, rightDepth) + 1
	r.lineBreaks = leftBreaks + rightBreaks
}

func (r *Rope) writeTo(b *strings.Builder) {
	if r == nil || r.length == 0 {
		return
	}
	if r.isLeaf() {
		b.Write(r.text)
		return
	}
	r.left.writeTo(b)
	r.right.writeTo(b)
}

func (r *Rope) lineStart(n, base int) int {
	if r == nil || n <= 0 {
		return base
	}
	if r.isLeaf() {
		seen := 0
		for i, ch := range r.text {
			if ch != '\n' {
				continue
			}
			seen++
			if seen == n {
				return base + i + 1
			}
		}
		return base + len(r.text)
	}

	leftBreaks := 0
	leftLength := 0
	if r.left != nil {
		leftBreaks = r.left.lineBreaks
		leftLength = r.left.length
	}
	if n <= leftBreaks {
		return r.left.lineStart(n, base)
	}
	return r.right.lineStart(n-leftBreaks, base+leftLength)
}

func (r *Rope) countLineBreaksBefore(offset int) int {
	if r == nil || offset <= 0 {
		return 0
	}
	if offset >= r.length {
		return r.lineBreaks
	}
	if r.isLeaf() {
		return countLineBreaks(r.text[:offset])
	}
	if offset <= r.weight {
		return r.left.countLineBreaksBefore(offset)
	}

	leftBreaks := 0
	if r.left != nil {
		leftBreaks = r.left.lineBreaks
	}
	return leftBreaks + r.right.countLineBreaksBefore(offset-r.weight)
}

func (it *RopeIterator) seek(r *Rope, offset int) {
	node := r
	base := 0
	for node != nil && !node.isLeaf() {
		leftLen := 0
		if node.left != nil {
			leftLen = node.left.length
		}
		if offset < base+leftLen {
			if node.right != nil {
				it.stack = append(it.stack, ropeIterFrame{node: node.right, base: base + leftLen})
			}
			node = node.left
			continue
		}
		base += leftLen
		node = node.right
	}

	it.leaf = node
	it.leafBase = base
	if node == nil {
		it.index = 0
		return
	}
	it.index = offset - base
	if it.index >= len(node.text) {
		it.advanceLeaf()
	}
}

func (it *RopeIterator) advanceLeaf() bool {
	for len(it.stack) > 0 {
		last := it.stack[len(it.stack)-1]
		it.stack = it.stack[:len(it.stack)-1]

		node := last.node
		base := last.base
		for node != nil && !node.isLeaf() {
			leftLen := 0
			if node.left != nil {
				leftLen = node.left.length
			}
			if node.right != nil {
				it.stack = append(it.stack, ropeIterFrame{node: node.right, base: base + leftLen})
			}
			node = node.left
		}
		if node == nil {
			continue
		}
		it.leaf = node
		it.leafBase = base
		it.index = 0
		if len(node.text) == 0 {
			continue
		}
		return true
	}

	it.leaf = nil
	it.index = 0
	return false
}

func collectLeafChunks(r *Rope, chunks *[][]byte) {
	if r == nil || r.length == 0 {
		return
	}
	if r.isLeaf() {
		data := r.text
		maxLeafSize := effectiveMaxLeafSize()
		for len(data) > 0 {
			size := min(len(data), maxLeafSize)
			*chunks = append(*chunks, append([]byte(nil), data[:size]...))
			data = data[size:]
		}
		return
	}
	collectLeafChunks(r.left, chunks)
	collectLeafChunks(r.right, chunks)
}

func compactChunks(chunks [][]byte) [][]byte {
	if len(chunks) == 0 {
		return chunks
	}

	maxLeafSize := effectiveMaxLeafSize()
	compacted := make([][]byte, 0, len(chunks))
	var current []byte
	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		if len(current)+len(chunk) > maxLeafSize && len(current) > 0 {
			compacted = append(compacted, current)
			current = nil
		}
		if len(chunk) > maxLeafSize {
			data := chunk
			for len(data) > 0 {
				size := min(len(data), maxLeafSize)
				if len(current) > 0 {
					compacted = append(compacted, current)
					current = nil
				}
				compacted = append(compacted, append([]byte(nil), data[:size]...))
				data = data[size:]
			}
			continue
		}
		current = append(current, chunk...)
	}
	if len(current) > 0 {
		compacted = append(compacted, current)
	}
	if len(compacted) == 0 {
		return [][]byte{nil}
	}
	return compacted
}

func maybeRebalance(r *Rope) *Rope {
	r = normalizeRope(r)
	if r.length <= effectiveMaxLeafSize() || r.IsBalanced() {
		return r
	}

	leafCount := r.leafCount()
	if leafCount <= 1 {
		return r
	}
	idealDepth := bits.Len(uint(leafCount))
	if r.depth > idealDepth*2 {
		return r.Rebalance()
	}
	return r
}

func (r *Rope) leafCount() int {
	if r == nil || r.length == 0 {
		return 0
	}
	if r.isLeaf() {
		return 1
	}
	return r.left.leafCount() + r.right.leafCount()
}

func normalizeRope(r *Rope) *Rope {
	if r == nil {
		return newLeaf(nil)
	}
	return r
}

func countLineBreaks(data []byte) int {
	return bytes.Count(data, []byte{'\n'})
}

func effectiveMaxLeafSize() int {
	if MaxLeafSize <= 0 {
		return 512
	}
	return MaxLeafSize
}

func clampRopeIndex(v, minValue, maxValue int) int {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}
