package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test execution defaults
	if cfg.Execution.MaxCycles != 1000000 {
		t.Errorf("Expected MaxCycles=1000000, got %d", cfg.Execution.MaxCycles)
	}
	if cfg.Execution.StackSize != 65536 {
		t.Errorf("Expected StackSize=65536, got %d", cfg.Execution.StackSize)
	}
	if cfg.Execution.DefaultEntry != "0x8000" {
		t.Errorf("Expected DefaultEntry=0x8000, got %s", cfg.Execution.DefaultEntry)
	}

	// Test debugger defaults
	if cfg.Debugger.HistorySize != 1000 {
		t.Errorf("Expected HistorySize=1000, got %d", cfg.Debugger.HistorySize)
	}
	if !cfg.Debugger.ShowSource {
		t.Error("Expected ShowSource=true")
	}

	// Test display defaults
	if cfg.Display.BytesPerLine != 16 {
		t.Errorf("Expected BytesPerLine=16, got %d", cfg.Display.BytesPerLine)
	}
	if cfg.Display.NumberFormat != "hex" {
		t.Errorf("Expected NumberFormat=hex, got %s", cfg.Display.NumberFormat)
	}

	// Test trace defaults
	if cfg.Trace.MaxEntries != 100000 {
		t.Errorf("Expected MaxEntries=100000, got %d", cfg.Trace.MaxEntries)
	}

	// Test statistics defaults
	if cfg.Statistics.Format != "json" {
		t.Errorf("Expected Format=json, got %s", cfg.Statistics.Format)
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()

	// Verify path is not empty
	if path == "" {
		t.Error("GetConfigPath returned empty string")
	}

	// Verify path ends with config.toml
	if filepath.Base(path) != "config.toml" {
		t.Errorf("Expected path to end with config.toml, got %s", path)
	}

	// Platform-specific checks
	switch runtime.GOOS {
	case "windows":
		// Should contain arm-emu
		if !filepath.IsAbs(path) && path != "config.toml" {
			t.Errorf("Expected absolute path on Windows, got %s", path)
		}

	case "darwin", "linux":
		// Should be in .config/arm-emu or be fallback
		dir := filepath.Dir(path)
		if filepath.Base(dir) != "arm-emu" && path != "config.toml" {
			t.Errorf("Expected path in arm-emu directory or fallback, got %s", path)
		}
	}
}

func TestGetLogPath(t *testing.T) {
	path := GetLogPath()

	// Verify path is not empty
	if path == "" {
		t.Error("GetLogPath returned empty string")
	}

	// Platform-specific checks
	switch runtime.GOOS {
	case "windows":
		// Should contain arm-emu\logs or be fallback
		if !filepath.IsAbs(path) && path != "logs" {
			t.Errorf("Expected absolute path on Windows, got %s", path)
		}

	case "darwin", "linux":
		// Should be in .local/share/arm-emu/logs or be fallback
		if filepath.Base(path) != "logs" {
			t.Errorf("Expected path to end with logs, got %s", path)
		}
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.toml")

	// Create a config with custom values
	cfg := DefaultConfig()
	cfg.Execution.MaxCycles = 5000000
	cfg.Execution.EnableTrace = true
	cfg.Debugger.HistorySize = 500
	cfg.Display.ColorOutput = false
	cfg.Trace.FilterRegs = "R0,R1,PC"

	// Save config
	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loaded, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values match
	if loaded.Execution.MaxCycles != 5000000 {
		t.Errorf("Expected MaxCycles=5000000, got %d", loaded.Execution.MaxCycles)
	}
	if !loaded.Execution.EnableTrace {
		t.Error("Expected EnableTrace=true")
	}
	if loaded.Debugger.HistorySize != 500 {
		t.Errorf("Expected HistorySize=500, got %d", loaded.Debugger.HistorySize)
	}
	if loaded.Display.ColorOutput {
		t.Error("Expected ColorOutput=false")
	}
	if loaded.Trace.FilterRegs != "R0,R1,PC" {
		t.Errorf("Expected FilterRegs=R0,R1,PC, got %s", loaded.Trace.FilterRegs)
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Try to load from a non-existent file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.toml")

	// Should return default config without error
	cfg, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom should not error on non-existent file: %v", err)
	}

	// Verify we got default config
	if cfg.Execution.MaxCycles != 1000000 {
		t.Error("Expected default config when file doesn't exist")
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	// Create a temporary file with invalid TOML
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.toml")

	invalidTOML := `
[execution]
max_cycles = "not a number"  # Invalid: should be uint64
`
	if err := os.WriteFile(configPath, []byte(invalidTOML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should return error
	_, err := LoadFrom(configPath)
	if err == nil {
		t.Error("Expected error when loading invalid TOML")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Try to save to a path with non-existent subdirectories
	configPath := filepath.Join(tempDir, "subdir1", "subdir2", "config.toml")

	cfg := DefaultConfig()
	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify directories were created
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Parent directories were not created")
	}
}
