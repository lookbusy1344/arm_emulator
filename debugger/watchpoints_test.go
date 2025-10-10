package debugger

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestWatchpointManager_AddWatchpoint(t *testing.T) {
	wm := NewWatchpointManager()

	wp := wm.AddWatchpoint(WatchWrite, "r0", 0, true, 0)

	if wp == nil {
		t.Fatal("AddWatchpoint returned nil")
	}

	if wp.ID != 1 {
		t.Errorf("Expected ID 1, got %d", wp.ID)
	}

	if wp.Type != WatchWrite {
		t.Errorf("Wrong watchpoint type: got %d, want %d", wp.Type, WatchWrite)
	}

	if wp.Expression != "r0" {
		t.Errorf("Expression = %s, want r0", wp.Expression)
	}

	if !wp.IsRegister {
		t.Error("Should be register watchpoint")
	}

	if !wp.Enabled {
		t.Error("Watchpoint should be enabled by default")
	}

	if wp.HitCount != 0 {
		t.Errorf("Initial hit count should be 0, got %d", wp.HitCount)
	}
}

func TestWatchpointManager_AddMultiple(t *testing.T) {
	wm := NewWatchpointManager()

	wp1 := wm.AddWatchpoint(WatchWrite, "r0", 0, true, 0)
	wp2 := wm.AddWatchpoint(WatchRead, "[0x1000]", 0x1000, false, 0)

	if wp1.ID == wp2.ID {
		t.Error("Watchpoint IDs should be unique")
	}

	if wm.Count() != 2 {
		t.Errorf("Expected 2 watchpoints, got %d", wm.Count())
	}
}

func TestWatchpointManager_DeleteWatchpoint(t *testing.T) {
	wm := NewWatchpointManager()

	wp := wm.AddWatchpoint(WatchWrite, "r0", 0, true, 0)

	err := wm.DeleteWatchpoint(wp.ID)
	if err != nil {
		t.Fatalf("DeleteWatchpoint failed: %v", err)
	}

	if wm.GetWatchpoint(wp.ID) != nil {
		t.Error("Watchpoint not deleted")
	}

	// Try to delete non-existent watchpoint
	err = wm.DeleteWatchpoint(999)
	if err == nil {
		t.Error("Expected error when deleting non-existent watchpoint")
	}
}

func TestWatchpointManager_EnableDisable(t *testing.T) {
	wm := NewWatchpointManager()

	wp := wm.AddWatchpoint(WatchWrite, "r0", 0, true, 0)

	// Disable
	err := wm.DisableWatchpoint(wp.ID)
	if err != nil {
		t.Fatalf("DisableWatchpoint failed: %v", err)
	}

	if wp.Enabled {
		t.Error("Watchpoint not disabled")
	}

	// Enable
	err = wm.EnableWatchpoint(wp.ID)
	if err != nil {
		t.Fatalf("EnableWatchpoint failed: %v", err)
	}

	if !wp.Enabled {
		t.Error("Watchpoint not enabled")
	}
}

func TestWatchpointManager_CheckWatchpoints_Register(t *testing.T) {
	wm := NewWatchpointManager()
	machine := vm.NewVM()

	// Add register watchpoint
	wp := wm.AddWatchpoint(WatchWrite, "r0", 0, true, 0)

	// Initialize watchpoint
	machine.CPU.R[0] = 100
	err := wm.InitializeWatchpoint(wp.ID, machine)
	if err != nil {
		t.Fatalf("InitializeWatchpoint failed: %v", err)
	}

	if wp.LastValue != 100 {
		t.Errorf("LastValue = %d, want 100", wp.LastValue)
	}

	// No change
	triggered, changed := wm.CheckWatchpoints(machine)
	if triggered != nil || changed {
		t.Error("Should not trigger when value hasn't changed")
	}

	// Change value
	machine.CPU.R[0] = 200
	triggered, changed = wm.CheckWatchpoints(machine)
	if triggered == nil || !changed {
		t.Fatal("Should trigger when value changes")
	}

	if triggered.ID != wp.ID {
		t.Errorf("Wrong watchpoint triggered: got %d, want %d", triggered.ID, wp.ID)
	}

	if wp.HitCount != 1 {
		t.Errorf("Hit count = %d, want 1", wp.HitCount)
	}

	if wp.LastValue != 200 {
		t.Errorf("LastValue not updated: got %d, want 200", wp.LastValue)
	}
}

func TestWatchpointManager_CheckWatchpoints_Memory(t *testing.T) {
	wm := NewWatchpointManager()
	machine := vm.NewVM()

	addr := uint32(0x00020000) // Data segment address

	// Add memory watchpoint
	wp := wm.AddWatchpoint(WatchWrite, "[0x00020000]", addr, false, 0)

	// Initialize watchpoint
	machine.Memory.WriteWord(addr, 0x12345678)
	err := wm.InitializeWatchpoint(wp.ID, machine)
	if err != nil {
		t.Fatalf("InitializeWatchpoint failed: %v", err)
	}

	// No change
	triggered, changed := wm.CheckWatchpoints(machine)
	if triggered != nil || changed {
		t.Error("Should not trigger when value hasn't changed")
	}

	// Change value
	machine.Memory.WriteWord(addr, 0xABCDEF00)
	triggered, changed = wm.CheckWatchpoints(machine)
	if triggered == nil || !changed {
		t.Fatal("Should trigger when value changes")
	}

	if triggered.ID != wp.ID {
		t.Errorf("Wrong watchpoint triggered: got %d, want %d", triggered.ID, wp.ID)
	}
}

func TestWatchpointManager_Disabled(t *testing.T) {
	wm := NewWatchpointManager()
	machine := vm.NewVM()

	// Add and disable watchpoint
	wp := wm.AddWatchpoint(WatchWrite, "r0", 0, true, 0)
	wm.InitializeWatchpoint(wp.ID, machine)
	wm.DisableWatchpoint(wp.ID)

	// Change value
	machine.CPU.R[0] = 100

	// Should not trigger
	triggered, _ := wm.CheckWatchpoints(machine)
	if triggered != nil {
		t.Error("Disabled watchpoint should not trigger")
	}
}

func TestWatchpointManager_GetAllWatchpoints(t *testing.T) {
	wm := NewWatchpointManager()

	wm.AddWatchpoint(WatchWrite, "r0", 0, true, 0)
	wm.AddWatchpoint(WatchRead, "r1", 0, true, 1)
	wm.AddWatchpoint(WatchReadWrite, "[0x1000]", 0x1000, false, 0)

	all := wm.GetAllWatchpoints()

	if len(all) != 3 {
		t.Errorf("Expected 3 watchpoints, got %d", len(all))
	}
}

func TestWatchpointManager_Clear(t *testing.T) {
	wm := NewWatchpointManager()

	wm.AddWatchpoint(WatchWrite, "r0", 0, true, 0)
	wm.AddWatchpoint(WatchRead, "r1", 0, true, 1)

	wm.Clear()

	if wm.Count() != 0 {
		t.Errorf("Expected 0 watchpoints after clear, got %d", wm.Count())
	}
}

func TestWatchpoint_Types(t *testing.T) {
	wm := NewWatchpointManager()

	wpWrite := wm.AddWatchpoint(WatchWrite, "r0", 0, true, 0)
	wpRead := wm.AddWatchpoint(WatchRead, "r1", 0, true, 1)
	wpAccess := wm.AddWatchpoint(WatchReadWrite, "r2", 0, true, 2)

	if wpWrite.Type != WatchWrite {
		t.Error("Wrong type for write watchpoint")
	}

	if wpRead.Type != WatchRead {
		t.Error("Wrong type for read watchpoint")
	}

	if wpAccess.Type != WatchReadWrite {
		t.Error("Wrong type for access watchpoint")
	}
}
