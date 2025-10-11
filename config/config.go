package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Config represents the emulator configuration
type Config struct {
	// Execution settings
	Execution struct {
		MaxCycles      uint64 `toml:"max_cycles"`
		StackSize      uint   `toml:"stack_size"`
		DefaultEntry   string `toml:"default_entry"`
		EnableTrace    bool   `toml:"enable_trace"`
		EnableMemTrace bool   `toml:"enable_mem_trace"`
		EnableStats    bool   `toml:"enable_stats"`
	} `toml:"execution"`

	// Debugger settings
	Debugger struct {
		HistorySize    int  `toml:"history_size"`
		AutoSaveBreaks bool `toml:"auto_save_breakpoints"`
		ShowSource     bool `toml:"show_source"`
		ShowRegisters  bool `toml:"show_registers"`
	} `toml:"debugger"`

	// Display settings
	Display struct {
		ColorOutput   bool   `toml:"color_output"`
		BytesPerLine  int    `toml:"bytes_per_line"`
		DisasmContext int    `toml:"disasm_context"`
		SourceContext int    `toml:"source_context"`
		NumberFormat  string `toml:"number_format"` // hex, dec, both
	} `toml:"display"`

	// Trace settings
	Trace struct {
		OutputFile    string `toml:"output_file"`
		FilterRegs    string `toml:"filter_registers"` // comma-separated: "R0,R1,PC"
		IncludeFlags  bool   `toml:"include_flags"`
		IncludeTiming bool   `toml:"include_timing"`
		MaxEntries    int    `toml:"max_entries"`
	} `toml:"trace"`

	// Statistics settings
	Statistics struct {
		OutputFile     string `toml:"output_file"`
		Format         string `toml:"format"` // json, csv, html
		CollectHotPath bool   `toml:"collect_hotpath"`
		TrackCalls     bool   `toml:"track_calls"`
	} `toml:"statistics"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	cfg := &Config{}

	// Execution defaults
	cfg.Execution.MaxCycles = 1000000
	cfg.Execution.StackSize = 65536 // 64KB
	cfg.Execution.DefaultEntry = "0x8000"
	cfg.Execution.EnableTrace = false
	cfg.Execution.EnableMemTrace = false
	cfg.Execution.EnableStats = false

	// Debugger defaults
	cfg.Debugger.HistorySize = 1000
	cfg.Debugger.AutoSaveBreaks = true
	cfg.Debugger.ShowSource = true
	cfg.Debugger.ShowRegisters = true

	// Display defaults
	cfg.Display.ColorOutput = true
	cfg.Display.BytesPerLine = 16
	cfg.Display.DisasmContext = 5
	cfg.Display.SourceContext = 5
	cfg.Display.NumberFormat = "hex"

	// Trace defaults
	cfg.Trace.OutputFile = "trace.log"
	cfg.Trace.FilterRegs = ""
	cfg.Trace.IncludeFlags = true
	cfg.Trace.IncludeTiming = true
	cfg.Trace.MaxEntries = 100000

	// Statistics defaults
	cfg.Statistics.OutputFile = "stats.json"
	cfg.Statistics.Format = "json"
	cfg.Statistics.CollectHotPath = true
	cfg.Statistics.TrackCalls = true

	return cfg
}

// GetConfigPath returns the platform-specific config file path
func GetConfigPath() string {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\arm-emu\config.toml
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		configDir = filepath.Join(configDir, "arm-emu")

	case "darwin", "linux":
		// macOS/Linux: ~/.config/arm-emu/config.toml
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory
			return "config.toml"
		}
		configDir = filepath.Join(homeDir, ".config", "arm-emu")

	default:
		// Unknown platform: use current directory
		return "config.toml"
	}

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0750); err != nil {
		// If we can't create the directory, fall back to current directory
		return "config.toml"
	}

	return filepath.Join(configDir, "config.toml")
}

// GetLogPath returns the platform-specific log directory path
func GetLogPath() string {
	var logDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\arm-emu\logs
		logDir = os.Getenv("APPDATA")
		if logDir == "" {
			logDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		logDir = filepath.Join(logDir, "arm-emu", "logs")

	case "darwin", "linux":
		// macOS/Linux: ~/.local/share/arm-emu/logs
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "logs"
		}
		logDir = filepath.Join(homeDir, ".local", "share", "arm-emu", "logs")

	default:
		return "logs"
	}

	// Ensure directory exists
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return "logs"
	}

	return logDir
}

// Load loads configuration from the default config file
func Load() (*Config, error) {
	return LoadFrom(GetConfigPath())
}

// LoadFrom loads configuration from the specified file
func LoadFrom(path string) (*Config, error) {
	cfg := DefaultConfig()

	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	// Read and parse config file
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save saves configuration to the default config file
func (c *Config) Save() error {
	return c.SaveTo(GetConfigPath())
}

// SaveTo saves configuration to the specified file
func (c *Config) SaveTo(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create file
	f, err := os.Create(path) // #nosec G304 -- user config file path
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close config file: %w", closeErr)
		}
	}()

	// Encode to TOML
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
