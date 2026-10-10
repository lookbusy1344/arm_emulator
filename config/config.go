package config

import (
	"fmt"
	"io"
	"log"
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

// UserDir returns the per-user arm-emu directory: %APPDATA%\arm-emu on Windows and
// ~/.config/arm-emu on macOS and Linux. It does not create the directory.
func UserDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		return filepath.Join(appData, "arm-emu"), nil
	case "darwin", "linux":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".config", "arm-emu"), nil
	default:
		return "", fmt.Errorf("no per-user directory on %s", runtime.GOOS)
	}
}

// GetConfigPath returns the platform-specific config file path, or config.toml in the
// current directory when the per-user directory is unavailable.
func GetConfigPath() string {
	configDir, err := UserDir()
	if err != nil {
		return "config.toml"
	}
	if err := os.MkdirAll(configDir, 0750); err != nil { // #nosec G703 -- path derived from OS-provided home dir, not user input
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
	if err := os.MkdirAll(logDir, 0750); err != nil { // #nosec G703 -- path derived from OS-provided home dir, not user input
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
	meta, err := toml.DecodeFile(path, cfg)
	if err != nil {
		// Log warning and return defaults instead of failing
		log.Printf("Warning: failed to parse config file %s: %v (using defaults)", path, err)
		return DefaultConfig(), nil
	}

	// Warn about unrecognized keys in config file
	undecoded := meta.Undecoded()
	if len(undecoded) > 0 {
		log.Printf("Warning: unrecognized keys in config file %s: %v", path, undecoded)
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

// debugEnv turns on debug logging when set to any non-empty value.
const debugEnv = "ARM_EMULATOR_DEBUG"

// DebugLogger returns a logger that appends to name in the log directory when
// ARM_EMULATOR_DEBUG is set, and discards output otherwise. The log directory is
// per user, so another account cannot plant a file or symlink at the log path.
// The file stays open for the life of the process.
func DebugLogger(name, prefix string) *log.Logger {
	if os.Getenv(debugEnv) == "" {
		return log.New(io.Discard, "", 0)
	}
	const flags = log.Ltime | log.Lmicroseconds | log.Lshortfile
	path := filepath.Join(GetLogPath(), name)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600) // #nosec G304 -- fixed name in the per-user log directory
	if err != nil {
		return log.New(os.Stderr, prefix, flags)
	}
	return log.New(f, prefix, flags)
}
