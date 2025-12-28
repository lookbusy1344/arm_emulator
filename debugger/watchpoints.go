package debugger

import (
	"fmt"
	"sync"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// WatchType represents the type of watchpoint
// NOTE: The current implementation can only detect value changes, not specific
// read/write operations. All watchpoint types behave the same way - they trigger
// when the monitored value differs from its previous value. True read-only or
// write-only tracking would require integration with the VM's memory access layer.
type WatchType int

const (
	WatchWrite     WatchType = iota // Trigger on write (currently same as WatchReadWrite)
	WatchRead                       // Trigger on read (currently same as WatchReadWrite)
	WatchReadWrite                  // Trigger on read or write (value change detection)
)

// Watchpoint represents a watchpoint for monitoring memory or register changes
type Watchpoint struct {
	ID         int
	Type       WatchType
	Expression string // Expression to watch (e.g., "r0", "[0x1000]", "myvar")
	Address    uint32 // Resolved address for memory watchpoints
	IsRegister bool   // True if watching a register
	Register   int    // Register number if IsRegister is true
	Enabled    bool
	LastValue  uint32 // Last known value
	HitCount   int
}

// WatchpointManager manages all watchpoints
type WatchpointManager struct {
	mu          sync.RWMutex
	watchpoints map[int]*Watchpoint
	nextID      int
}

// NewWatchpointManager creates a new watchpoint manager
func NewWatchpointManager() *WatchpointManager {
	return &WatchpointManager{
		watchpoints: make(map[int]*Watchpoint),
		nextID:      1,
	}
}

// AddWatchpoint adds a new watchpoint
func (wm *WatchpointManager) AddWatchpoint(wpType WatchType, expression string, address uint32, isRegister bool, register int) *Watchpoint {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wp := &Watchpoint{
		ID:         wm.nextID,
		Type:       wpType,
		Expression: expression,
		Address:    address,
		IsRegister: isRegister,
		Register:   register,
		Enabled:    true,
		LastValue:  0,
		HitCount:   0,
	}

	wm.watchpoints[wp.ID] = wp
	wm.nextID++

	return wp
}

// DeleteWatchpoint removes a watchpoint by ID
func (wm *WatchpointManager) DeleteWatchpoint(id int) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, exists := wm.watchpoints[id]; !exists {
		return fmt.Errorf("watchpoint %d not found", id)
	}

	delete(wm.watchpoints, id)
	return nil
}

// EnableWatchpoint enables a watchpoint by ID
func (wm *WatchpointManager) EnableWatchpoint(id int) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wp, exists := wm.watchpoints[id]
	if !exists {
		return fmt.Errorf("watchpoint %d not found", id)
	}

	wp.Enabled = true
	return nil
}

// DisableWatchpoint disables a watchpoint by ID
func (wm *WatchpointManager) DisableWatchpoint(id int) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wp, exists := wm.watchpoints[id]
	if !exists {
		return fmt.Errorf("watchpoint %d not found", id)
	}

	wp.Enabled = false
	return nil
}

// GetWatchpoint gets a watchpoint by ID
func (wm *WatchpointManager) GetWatchpoint(id int) *Watchpoint {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	return wm.watchpoints[id]
}

// GetAllWatchpoints returns all watchpoints
func (wm *WatchpointManager) GetAllWatchpoints() []*Watchpoint {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	result := make([]*Watchpoint, 0, len(wm.watchpoints))
	for _, wp := range wm.watchpoints {
		result = append(result, wp)
	}

	return result
}

// CheckWatchpoints checks all watchpoints and returns the first that has changed
// NOTE: This implementation uses value change detection, not true read/write tracking.
// The watchpoint Type field is currently not enforced - all types behave the same way,
// triggering when the monitored value differs from its previous value.
func (wm *WatchpointManager) CheckWatchpoints(machine *vm.VM) (*Watchpoint, bool) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for _, wp := range wm.watchpoints {
		if !wp.Enabled {
			continue
		}

		var currentValue uint32
		var err error

		if wp.IsRegister {
			// Check register value
			currentValue = machine.CPU.GetRegister(wp.Register)
		} else {
			// Check memory value
			currentValue, err = machine.Memory.ReadWord(wp.Address)
			if err != nil {
				// Skip if memory read fails
				continue
			}
		}

		// Check if value has changed
		if currentValue != wp.LastValue {
			wp.HitCount++
			wp.LastValue = currentValue
			return wp, true
		}
	}

	return nil, false
}

// InitializeWatchpoint initializes the last value for a watchpoint
func (wm *WatchpointManager) InitializeWatchpoint(id int, machine *vm.VM) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wp, exists := wm.watchpoints[id]
	if !exists {
		return fmt.Errorf("watchpoint %d not found", id)
	}

	if wp.IsRegister {
		wp.LastValue = machine.CPU.GetRegister(wp.Register)
	} else {
		value, err := machine.Memory.ReadWord(wp.Address)
		if err != nil {
			return fmt.Errorf("failed to initialize watchpoint: %w", err)
		}
		wp.LastValue = value
	}

	return nil
}

// Clear removes all watchpoints
func (wm *WatchpointManager) Clear() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.watchpoints = make(map[int]*Watchpoint)
}

// Count returns the number of watchpoints
func (wm *WatchpointManager) Count() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	return len(wm.watchpoints)
}
