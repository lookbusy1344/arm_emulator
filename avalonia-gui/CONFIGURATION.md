# Configuration Guide

Complete guide to configuring the ARM Emulator Avalonia GUI.

## Settings Overview

Settings are accessible via:
- **Menu**: File → Preferences
- **Keyboard**: `Ctrl+,` (Windows/Linux) or `Cmd+,` (macOS)

All settings are stored in memory during the current session. **Settings do not currently persist across application restarts** (persistence planned for future release).

## General Settings

### Backend URL

**Setting**: Backend URL
**Default**: `http://localhost:8080`
**Description**: The URL where the Go backend is running

**Configuration Options**:
- `http://localhost:8080` (default) - Local development
- `http://localhost:CUSTOM_PORT` - Custom port
- `http://REMOTE_IP:8080` - Remote backend (network access required)

**Notes**:
- Backend must be running before starting the GUI
- Changes take effect after restarting the application or creating a new session
- Ensure firewall allows connections if using remote backend

### Theme

**Setting**: Color Scheme
**Options**:
- **Auto** (default) - Follow system theme (light/dark)
- **Light** - Always use light theme
- **Dark** - Always use dark theme

**Notes**:
- Auto mode automatically switches when system theme changes
- Theme changes require application restart to take full effect
- High contrast system themes are respected

### Auto-Scroll Memory Writes

**Setting**: Auto-scroll to memory writes
**Default**: `true`
**Description**: Automatically navigate memory view to show recent write operations

**Options**:
- **Enabled** - Memory view scrolls to show writes during execution
- **Disabled** - Memory view stays at current address

**Use Cases**:
- **Enable**: When debugging memory corruption or tracking writes
- **Disable**: When focusing on specific memory region

## Editor Settings

### Font Size

**Setting**: Editor Font Size
**Default**: `14` pt
**Range**: 10-24 pt
**Description**: Size of monospace font in code editor

**Recommendations**:
- **10-12 pt**: Small screens, dense code viewing
- **14 pt**: Default, comfortable for most users
- **16-18 pt**: Large screens, presentations
- **20-24 pt**: Accessibility, vision impairment

**Notes**:
- Changes apply immediately with live preview
- Affects line numbers and gutter display
- Does not affect console or other UI elements

### Recent Files

**Setting**: Recent Files Limit
**Default**: `10`
**Range**: 1-50
**Description**: Maximum number of recently opened files to track

**Notes**:
- Recent files list appears in File menu
- Files are added when opened (Open or Examples)
- List is cleared when limit is reduced below current size
- **Not currently persisted across sessions**

## Advanced Configuration

### Configuration File (Future Feature)

Settings will be stored in a JSON configuration file (not yet implemented):

**Location**:
- **Windows**: `%APPDATA%\ARMEmulator\settings.json`
- **macOS**: `~/Library/Application Support/ARMEmulator/settings.json`
- **Linux**: `~/.config/ARMEmulator/settings.json`

**Future Schema**:
```json
{
  "backendUrl": "http://localhost:8080",
  "theme": "Auto",
  "editorFontSize": 14,
  "autoScrollMemoryWrites": true,
  "recentFilesLimit": 10,
  "recentFiles": [
    "/path/to/file1.s",
    "/path/to/file2.s"
  ]
}
```

## Platform-Specific Settings

### Windows

- Settings stored in roaming profile (future)
- Native file dialogs respect Windows theme
- Font rendering uses ClearType

### macOS

- Settings follow macOS conventions (future)
- Native menu bar integration
- Automatic dark mode switching based on system appearance
- Font rendering uses Core Text

### Linux

- Settings location follows XDG Base Directory spec (future)
- GTK file dialogs on GTK-based desktops
- KDE integration on KDE Plasma
- Font rendering depends on desktop environment

## Keyboard Shortcuts

Keyboard shortcuts cannot currently be customized. See [KEYBOARD_SHORTCUTS.md](KEYBOARD_SHORTCUTS.md) for complete list.

## Resetting to Defaults

To reset all settings to defaults:
1. Close the application
2. Delete the configuration file (when persistence is implemented)
3. Restart the application

**Current Version**: Since settings are not persisted, simply restart the application to reset all settings.

## Environment Variables

The following environment variables can override settings:

| Variable | Default | Description |
|----------|---------|-------------|
| `ARM_EMULATOR_BACKEND_URL` | `http://localhost:8080` | Backend URL (not yet implemented) |
| `ARM_EMULATOR_THEME` | `Auto` | Force theme: `Auto`, `Light`, or `Dark` (not yet implemented) |

**Note**: Environment variable support planned for future release.

## Troubleshooting

### Settings Not Saving

**Issue**: Settings reset after restarting application
**Cause**: Settings persistence not yet implemented
**Workaround**: Reconfigure settings each session
**Status**: Planned for Phase 12 (Polish & Release)

### Backend Connection Failed

**Issue**: "Cannot connect to backend" error
**Solutions**:
1. Verify backend is running: `./arm-emulator`
2. Check backend URL in preferences matches running instance
3. Verify firewall allows port 8080 (or configured port)
4. Test connection: `curl http://localhost:8080/api/v1/version`

### Theme Not Updating

**Issue**: Theme changes don't take effect
**Cause**: Theme switching requires application restart
**Solution**: Restart application after changing theme

## See Also

- [README.md](README.md) - Build and run instructions
- [KEYBOARD_SHORTCUTS.md](KEYBOARD_SHORTCUTS.md) - Keyboard shortcut reference
- [INTEGRATION_TESTING.md](INTEGRATION_TESTING.md) - Running integration tests
- [../docs/AVALONIA_IMPLEMENTATION_PLAN.md](../docs/AVALONIA_IMPLEMENTATION_PLAN.md) - Implementation details
