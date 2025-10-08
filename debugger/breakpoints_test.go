package debugger

import (
	"testing"
)

func TestBreakpointManager_AddBreakpoint(t *testing.T) {
	bm := NewBreakpointManager()

	bp := bm.AddBreakpoint(0x1000, false, "")

	if bp == nil {
		t.Fatal("AddBreakpoint returned nil")
	}

	if bp.ID != 1 {
		t.Errorf("Expected ID 1, got %d", bp.ID)
	}

	if bp.Address != 0x1000 {
		t.Errorf("Expected address 0x1000, got 0x%08X", bp.Address)
	}

	if !bp.Enabled {
		t.Error("Breakpoint should be enabled by default")
	}

	if bp.Temporary {
		t.Error("Breakpoint should not be temporary")
	}

	if bp.HitCount != 0 {
		t.Errorf("Initial hit count should be 0, got %d", bp.HitCount)
	}
}

func TestBreakpointManager_AddMultiple(t *testing.T) {
	bm := NewBreakpointManager()

	bp1 := bm.AddBreakpoint(0x1000, false, "")
	bp2 := bm.AddBreakpoint(0x2000, false, "")

	if bp1.ID == bp2.ID {
		t.Error("Breakpoint IDs should be unique")
	}

	if bm.Count() != 2 {
		t.Errorf("Expected 2 breakpoints, got %d", bm.Count())
	}
}

func TestBreakpointManager_AddDuplicate(t *testing.T) {
	bm := NewBreakpointManager()

	bp1 := bm.AddBreakpoint(0x1000, false, "")
	bp2 := bm.AddBreakpoint(0x1000, false, "r0 == 5")

	// Adding to same address should update existing breakpoint
	if bp1.ID != bp2.ID {
		t.Error("Duplicate address should update existing breakpoint")
	}

	if bp2.Condition != "r0 == 5" {
		t.Error("Condition not updated")
	}
}

func TestBreakpointManager_DeleteBreakpoint(t *testing.T) {
	bm := NewBreakpointManager()

	bp := bm.AddBreakpoint(0x1000, false, "")

	err := bm.DeleteBreakpoint(bp.ID)
	if err != nil {
		t.Fatalf("DeleteBreakpoint failed: %v", err)
	}

	if bm.GetBreakpoint(0x1000) != nil {
		t.Error("Breakpoint not deleted")
	}

	// Try to delete non-existent breakpoint
	err = bm.DeleteBreakpoint(999)
	if err == nil {
		t.Error("Expected error when deleting non-existent breakpoint")
	}
}

func TestBreakpointManager_EnableDisable(t *testing.T) {
	bm := NewBreakpointManager()

	bp := bm.AddBreakpoint(0x1000, false, "")

	// Disable
	err := bm.DisableBreakpoint(bp.ID)
	if err != nil {
		t.Fatalf("DisableBreakpoint failed: %v", err)
	}

	if bp.Enabled {
		t.Error("Breakpoint not disabled")
	}

	// Enable
	err = bm.EnableBreakpoint(bp.ID)
	if err != nil {
		t.Fatalf("EnableBreakpoint failed: %v", err)
	}

	if !bp.Enabled {
		t.Error("Breakpoint not enabled")
	}
}

func TestBreakpointManager_GetBreakpoint(t *testing.T) {
	bm := NewBreakpointManager()

	bm.AddBreakpoint(0x1000, false, "")
	bm.AddBreakpoint(0x2000, false, "")

	bp := bm.GetBreakpoint(0x1000)
	if bp == nil {
		t.Fatal("GetBreakpoint returned nil")
	}

	if bp.Address != 0x1000 {
		t.Errorf("Wrong breakpoint returned: got 0x%08X, want 0x1000", bp.Address)
	}

	bp = bm.GetBreakpoint(0x3000)
	if bp != nil {
		t.Error("GetBreakpoint should return nil for non-existent address")
	}
}

func TestBreakpointManager_GetBreakpointByID(t *testing.T) {
	bm := NewBreakpointManager()

	bp1 := bm.AddBreakpoint(0x1000, false, "")
	bp2 := bm.AddBreakpoint(0x2000, false, "")

	found := bm.GetBreakpointByID(bp1.ID)
	if found != bp1 {
		t.Error("GetBreakpointByID returned wrong breakpoint")
	}

	found = bm.GetBreakpointByID(bp2.ID)
	if found != bp2 {
		t.Error("GetBreakpointByID returned wrong breakpoint")
	}

	found = bm.GetBreakpointByID(999)
	if found != nil {
		t.Error("GetBreakpointByID should return nil for non-existent ID")
	}
}

func TestBreakpointManager_GetAllBreakpoints(t *testing.T) {
	bm := NewBreakpointManager()

	bm.AddBreakpoint(0x1000, false, "")
	bm.AddBreakpoint(0x2000, false, "")
	bm.AddBreakpoint(0x3000, false, "")

	all := bm.GetAllBreakpoints()

	if len(all) != 3 {
		t.Errorf("Expected 3 breakpoints, got %d", len(all))
	}
}

func TestBreakpointManager_Clear(t *testing.T) {
	bm := NewBreakpointManager()

	bm.AddBreakpoint(0x1000, false, "")
	bm.AddBreakpoint(0x2000, false, "")

	bm.Clear()

	if bm.Count() != 0 {
		t.Errorf("Expected 0 breakpoints after clear, got %d", bm.Count())
	}
}

func TestBreakpointManager_HasBreakpoint(t *testing.T) {
	bm := NewBreakpointManager()

	bm.AddBreakpoint(0x1000, false, "")

	if !bm.HasBreakpoint(0x1000) {
		t.Error("HasBreakpoint returned false for existing breakpoint")
	}

	if bm.HasBreakpoint(0x2000) {
		t.Error("HasBreakpoint returned true for non-existent breakpoint")
	}
}

func TestBreakpoint_Temporary(t *testing.T) {
	bm := NewBreakpointManager()

	bp := bm.AddBreakpoint(0x1000, true, "")

	if !bp.Temporary {
		t.Error("Breakpoint should be temporary")
	}
}

func TestBreakpoint_Condition(t *testing.T) {
	bm := NewBreakpointManager()

	condition := "r0 == 42"
	bp := bm.AddBreakpoint(0x1000, false, condition)

	if bp.Condition != condition {
		t.Errorf("Condition = %s, want %s", bp.Condition, condition)
	}
}

func TestBreakpoint_HitCount(t *testing.T) {
	bm := NewBreakpointManager()

	bp := bm.AddBreakpoint(0x1000, false, "")

	if bp.HitCount != 0 {
		t.Errorf("Initial hit count = %d, want 0", bp.HitCount)
	}

	bp.HitCount++
	bp.HitCount++

	if bp.HitCount != 2 {
		t.Errorf("Hit count = %d, want 2", bp.HitCount)
	}
}
