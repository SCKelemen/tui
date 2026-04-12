package tui

import (
	"strings"
	"testing"
)

func TestNewRope(t *testing.T) {
	r := NewRope("hello")
	if r == nil {
		t.Fatal("NewRope returned nil")
	}
	if r.String() != "hello" {
		t.Fatalf("String() = %q, want hello", r.String())
	}
}

func TestLength(t *testing.T) {
	if got := NewRope("hello").Length(); got != 5 {
		t.Fatalf("Length() = %d, want 5", got)
	}
}

func TestString(t *testing.T) {
	input := "hello\nworld"
	if got := NewRope(input).String(); got != input {
		t.Fatalf("String() = %q, want %q", got, input)
	}
}

func TestIndex(t *testing.T) {
	r := NewRope("hello")
	if got := r.Index(1); got != 'e' {
		t.Fatalf("Index(1) = %q, want e", got)
	}
}

func TestConcat(t *testing.T) {
	joined := NewRope("hello").Concat(NewRope(" world"))
	if got := joined.String(); got != "hello world" {
		t.Fatalf("Concat() = %q, want hello world", got)
	}
}

func TestSplit(t *testing.T) {
	left, right := NewRope("hello world").Split(5)
	if left.String() != "hello" || right.String() != " world" {
		t.Fatalf("Split(5) = (%q, %q), want (%q, %q)", left.String(), right.String(), "hello", " world")
	}
}

func TestInsert(t *testing.T) {
	got := NewRope("helo").Insert(3, "l")
	if got.String() != "hello" {
		t.Fatalf("Insert() = %q, want hello", got.String())
	}
}

func TestDelete(t *testing.T) {
	got := NewRope("hello cruel world").Delete(5, 11)
	if got.String() != "hello world" {
		t.Fatalf("Delete() = %q, want hello world", got.String())
	}
}

func TestSubstring(t *testing.T) {
	got := NewRope("hello world").Substring(6, 11)
	if got != "world" {
		t.Fatalf("Substring() = %q, want world", got)
	}
}

func TestLineCount(t *testing.T) {
	r := NewRope("alpha\nbeta\ngamma")
	if got := r.LineCount(); got != 3 {
		t.Fatalf("LineCount() = %d, want 3", got)
	}
}

func TestLine(t *testing.T) {
	r := NewRope("alpha\nbeta\ngamma")
	if got := r.Line(1); got != "beta" {
		t.Fatalf("Line(1) = %q, want beta", got)
	}
}

func TestLineStartEnd(t *testing.T) {
	r := NewRope("alpha\nbeta\ngamma")
	if got := r.LineStart(1); got != 6 {
		t.Fatalf("LineStart(1) = %d, want 6", got)
	}
	if got := r.LineEnd(1); got != 10 {
		t.Fatalf("LineEnd(1) = %d, want 10", got)
	}
}

func TestLineAt(t *testing.T) {
	r := NewRope("alpha\nbeta\ngamma")
	if got := r.LineAt(8); got != 1 {
		t.Fatalf("LineAt(8) = %d, want 1", got)
	}
}

func TestIter(t *testing.T) {
	r := NewRope("hello")
	it := r.Iter()
	var b strings.Builder
	for {
		ch, ok := it.Next()
		if !ok {
			break
		}
		b.WriteByte(ch)
	}
	if got := b.String(); got != "hello" {
		t.Fatalf("Iter() = %q, want hello", got)
	}
}

func TestIterRange(t *testing.T) {
	r := NewRope("hello world")
	it := r.IterRange(6, 11)
	var b strings.Builder
	for {
		ch, ok := it.Next()
		if !ok {
			break
		}
		b.WriteByte(ch)
	}
	if got := b.String(); got != "world" {
		t.Fatalf("IterRange() = %q, want world", got)
	}
}

func TestRebalance(t *testing.T) {
	old := MaxLeafSize
	MaxLeafSize = 4
	defer func() { MaxLeafSize = old }()

	leafA := newLeaf([]byte("aaaa"))
	leafB := newLeaf([]byte("bbbb"))
	leafC := newLeaf([]byte("cccc"))
	leafD := newLeaf([]byte("dddd"))

	left := &Rope{left: leafA, right: leafB}
	left.refresh()
	leftHeavy := &Rope{left: left, right: leafC}
	leftHeavy.refresh()
	unbalanced := &Rope{left: leftHeavy, right: leafD}
	unbalanced.refresh()

	if unbalanced.IsBalanced() {
		t.Fatal("expected constructed rope to be unbalanced")
	}

	rebalanced := unbalanced.Rebalance()
	if got := rebalanced.String(); got != "aaaabbbbccccdddd" {
		t.Fatalf("Rebalance().String() = %q, want aaaabbbbccccdddd", got)
	}
	if !rebalanced.IsBalanced() {
		t.Fatal("rebuilt rope should be balanced")
	}
}

func TestDepth(t *testing.T) {
	old := MaxLeafSize
	MaxLeafSize = 4
	defer func() { MaxLeafSize = old }()

	r := NewRope("abcdefghijklmnop")
	if got := r.Depth(); got < 2 {
		t.Fatalf("Depth() = %d, want >= 2", got)
	}
}
