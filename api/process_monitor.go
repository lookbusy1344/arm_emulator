package api

import (
	"log"
	"os"
	"sync"
	"time"
)

// ProcessMonitor watches the parent process and triggers shutdown when it dies.
// This prevents orphaned backend processes when the Swift GUI crashes or is force-quit.
type ProcessMonitor struct {
	parentPID     int
	checkInterval time.Duration
	shutdownFunc  func()
	stopChan      chan struct{}
	stopOnce      sync.Once
}

// NewProcessMonitor creates a monitor that calls shutdownFunc when the parent process dies.
// The parent PID is captured at creation time via os.Getppid().
func NewProcessMonitor(shutdownFunc func()) *ProcessMonitor {
	return &ProcessMonitor{
		parentPID:     os.Getppid(),
		checkInterval: 2 * time.Second,
		shutdownFunc:  shutdownFunc,
		stopChan:      make(chan struct{}),
	}
}

// Start begins monitoring the parent process in a background goroutine.
// The monitor checks every 2 seconds if the parent PID has changed.
// When the parent dies, the OS re-parents the process (typically to PID 1),
// triggering the shutdown callback.
func (pm *ProcessMonitor) Start() {
	go pm.monitorLoop()
}

// Stop gracefully stops the monitor goroutine.
// Safe to call multiple times - only the first call has an effect.
func (pm *ProcessMonitor) Stop() {
	pm.stopOnce.Do(func() {
		close(pm.stopChan)
	})
}

// monitorLoop runs in a goroutine and periodically checks if the parent process is still alive.
func (pm *ProcessMonitor) monitorLoop() {
	ticker := time.NewTicker(pm.checkInterval)
	defer ticker.Stop()

	log.Printf("Process monitor started (parent PID: %d, check interval: %v)", pm.parentPID, pm.checkInterval)

	for {
		select {
		case <-ticker.C:
			currentPPID := os.Getppid()
			if currentPPID != pm.parentPID {
				log.Printf("Parent process died (PPID changed: %d -> %d), initiating graceful shutdown",
					pm.parentPID, currentPPID)
				pm.shutdownFunc()
				return
			}
		case <-pm.stopChan:
			log.Println("Process monitor stopped")
			return
		}
	}
}
