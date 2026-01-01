package debugger

import (
	"fmt"
	"sync"
)

// Breakpoint represents a breakpoint at a specific address
type Breakpoint struct {
	ID        int
	Address   uint32
	Enabled   bool
	Temporary bool   // Auto-delete after first hit
	Condition string // Optional condition expression
	HitCount  int    // Number of times this breakpoint was hit
}

// BreakpointManager manages all breakpoints
type BreakpointManager struct {
	mu          sync.RWMutex
	breakpoints map[uint32]*Breakpoint // address -> breakpoint
	nextID      int
}

// NewBreakpointManager creates a new breakpoint manager
func NewBreakpointManager() *BreakpointManager {
	return &BreakpointManager{
		breakpoints: make(map[uint32]*Breakpoint),
		nextID:      1,
	}
}

// AddBreakpoint adds a new breakpoint at the specified address
func (bm *BreakpointManager) AddBreakpoint(address uint32, temporary bool, condition string) *Breakpoint {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Check if breakpoint already exists at this address
	if bp, exists := bm.breakpoints[address]; exists {
		// Update existing breakpoint
		bp.Enabled = true
		bp.Temporary = temporary
		bp.Condition = condition
		return bp
	}

	// Create new breakpoint
	bp := &Breakpoint{
		ID:        bm.nextID,
		Address:   address,
		Enabled:   true,
		Temporary: temporary,
		Condition: condition,
		HitCount:  0,
	}

	bm.breakpoints[address] = bp
	bm.nextID++

	return bp
}

// DeleteBreakpoint removes a breakpoint by ID
func (bm *BreakpointManager) DeleteBreakpoint(id int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Find breakpoint by ID
	for addr, bp := range bm.breakpoints {
		if bp.ID == id {
			delete(bm.breakpoints, addr)
			return nil
		}
	}

	return fmt.Errorf("breakpoint %d not found", id)
}

// DeleteBreakpointAt removes a breakpoint at a specific address
func (bm *BreakpointManager) DeleteBreakpointAt(address uint32) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.breakpoints[address]; !exists {
		return fmt.Errorf("no breakpoint at address 0x%08X", address)
	}

	delete(bm.breakpoints, address)
	return nil
}

// EnableBreakpoint enables a breakpoint by ID
func (bm *BreakpointManager) EnableBreakpoint(id int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, bp := range bm.breakpoints {
		if bp.ID == id {
			bp.Enabled = true
			return nil
		}
	}

	return fmt.Errorf("breakpoint %d not found", id)
}

// DisableBreakpoint disables a breakpoint by ID
func (bm *BreakpointManager) DisableBreakpoint(id int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, bp := range bm.breakpoints {
		if bp.ID == id {
			bp.Enabled = false
			return nil
		}
	}

	return fmt.Errorf("breakpoint %d not found", id)
}

// GetBreakpoint gets a breakpoint at a specific address
func (bm *BreakpointManager) GetBreakpoint(address uint32) *Breakpoint {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return bm.breakpoints[address]
}

// GetBreakpointByID gets a breakpoint by ID
func (bm *BreakpointManager) GetBreakpointByID(id int) *Breakpoint {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for _, bp := range bm.breakpoints {
		if bp.ID == id {
			return bp
		}
	}

	return nil
}

// GetAllBreakpoints returns all breakpoints
func (bm *BreakpointManager) GetAllBreakpoints() []*Breakpoint {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make([]*Breakpoint, 0, len(bm.breakpoints))
	for _, bp := range bm.breakpoints {
		result = append(result, bp)
	}

	return result
}

// Clear removes all breakpoints
func (bm *BreakpointManager) Clear() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.breakpoints = make(map[uint32]*Breakpoint)
}

// HasBreakpoint checks if a breakpoint exists at the given address
func (bm *BreakpointManager) HasBreakpoint(address uint32) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	_, exists := bm.breakpoints[address]
	return exists
}

// Count returns the number of breakpoints
func (bm *BreakpointManager) Count() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return len(bm.breakpoints)
}

// ProcessHit atomically increments hit count and handles temporary breakpoint deletion
// Returns a copy of the breakpoint for safe access after the lock is released
func (bm *BreakpointManager) ProcessHit(address uint32) *Breakpoint {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bp, exists := bm.breakpoints[address]
	if !exists {
		return nil
	}

	// Increment hit count
	bp.HitCount++

	// Make a copy for return
	result := *bp

	// Delete if temporary
	if bp.Temporary {
		delete(bm.breakpoints, address)
	}

	return &result
}
