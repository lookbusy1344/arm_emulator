package debugger

import (
	"testing"
)

func TestCommandHistory_Add(t *testing.T) {
	h := NewCommandHistory()

	h.Add("step")
	h.Add("continue")
	h.Add("break 0x1000")

	if h.Size() != 3 {
		t.Errorf("Size = %d, want 3", h.Size())
	}

	all := h.GetAll()
	if len(all) != 3 {
		t.Errorf("GetAll() length = %d, want 3", len(all))
	}

	if all[0] != "step" {
		t.Errorf("First command = %s, want step", all[0])
	}
}

func TestCommandHistory_IgnoreEmpty(t *testing.T) {
	h := NewCommandHistory()

	h.Add("step")
	h.Add("")
	h.Add("continue")

	if h.Size() != 2 {
		t.Errorf("Size = %d, want 2 (empty commands should be ignored)", h.Size())
	}
}

func TestCommandHistory_IgnoreDuplicates(t *testing.T) {
	h := NewCommandHistory()

	h.Add("step")
	h.Add("step")
	h.Add("continue")

	if h.Size() != 2 {
		t.Errorf("Size = %d, want 2 (duplicate should be ignored)", h.Size())
	}

	all := h.GetAll()
	if all[0] != "step" || all[1] != "continue" {
		t.Error("Duplicate command was not ignored correctly")
	}
}

func TestCommandHistory_Previous(t *testing.T) {
	h := NewCommandHistory()

	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	// Navigate backwards
	prev := h.Previous()
	if prev != "cmd3" {
		t.Errorf("Previous() = %s, want cmd3", prev)
	}

	prev = h.Previous()
	if prev != "cmd2" {
		t.Errorf("Previous() = %s, want cmd2", prev)
	}

	prev = h.Previous()
	if prev != "cmd1" {
		t.Errorf("Previous() = %s, want cmd1", prev)
	}

	// At start, should return empty
	prev = h.Previous()
	if prev != "" {
		t.Errorf("Previous() at start = %s, want empty", prev)
	}
}

func TestCommandHistory_Next(t *testing.T) {
	h := NewCommandHistory()

	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	// Navigate backwards first
	h.Previous()
	h.Previous()
	h.Previous()

	// Now navigate forwards
	next := h.Next()
	if next != "cmd2" {
		t.Errorf("Next() = %s, want cmd2", next)
	}

	next = h.Next()
	if next != "cmd3" {
		t.Errorf("Next() = %s, want cmd3", next)
	}

	// At end, should return empty
	next = h.Next()
	if next != "" {
		t.Errorf("Next() at end = %s, want empty", next)
	}
}

func TestCommandHistory_GetLast(t *testing.T) {
	h := NewCommandHistory()

	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	last := h.GetLast()
	if last != "cmd3" {
		t.Errorf("GetLast() = %s, want cmd3", last)
	}

	// GetLast should not change position
	last = h.GetLast()
	if last != "cmd3" {
		t.Errorf("GetLast() = %s, want cmd3", last)
	}
}

func TestCommandHistory_Clear(t *testing.T) {
	h := NewCommandHistory()

	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	h.Clear()

	if h.Size() != 0 {
		t.Errorf("Size after clear = %d, want 0", h.Size())
	}

	last := h.GetLast()
	if last != "" {
		t.Errorf("GetLast after clear = %s, want empty", last)
	}
}

func TestCommandHistory_Search(t *testing.T) {
	h := NewCommandHistory()

	h.Add("break 0x1000")
	h.Add("break 0x2000")
	h.Add("step")
	h.Add("continue")

	results := h.Search("break")

	if len(results) != 2 {
		t.Errorf("Search results length = %d, want 2", len(results))
	}

	if results[0] != "break 0x1000" {
		t.Errorf("Search result[0] = %s, want 'break 0x1000'", results[0])
	}

	if results[1] != "break 0x2000" {
		t.Errorf("Search result[1] = %s, want 'break 0x2000'", results[1])
	}
}

func TestCommandHistory_SearchNoMatches(t *testing.T) {
	h := NewCommandHistory()

	h.Add("step")
	h.Add("continue")

	results := h.Search("break")

	if len(results) != 0 {
		t.Errorf("Search with no matches should return empty slice, got %d results", len(results))
	}
}

func TestCommandHistory_MaxSize(t *testing.T) {
	h := NewCommandHistory()

	// Add more than max size
	for i := 0; i < 1100; i++ {
		h.Add("cmd")
	}

	// Should be trimmed to max size
	if h.Size() > 1000 {
		t.Errorf("Size = %d, should not exceed max size of 1000", h.Size())
	}
}

func TestCommandHistory_EmptyHistory(t *testing.T) {
	h := NewCommandHistory()

	if h.Size() != 0 {
		t.Errorf("New history size = %d, want 0", h.Size())
	}

	last := h.GetLast()
	if last != "" {
		t.Errorf("GetLast on empty history = %s, want empty", last)
	}

	prev := h.Previous()
	if prev != "" {
		t.Errorf("Previous on empty history = %s, want empty", prev)
	}

	next := h.Next()
	if next != "" {
		t.Errorf("Next on empty history = %s, want empty", next)
	}
}
