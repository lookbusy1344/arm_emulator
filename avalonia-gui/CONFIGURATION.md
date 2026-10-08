# Configuration Guide

Complete guide to configuring the ARM Emulator Avalonia GUI.

## Settings Overview

Settings are accessible via:
- **Menu**: File → Preferences
- **Keyboard**: `Ctrl+,` (Windows/Linux) or `Cmd+,` (macOS)

Settings persist across restarts in a JSON file (see [Settings File](#settings-file)). Invalid values are clamped to their range.

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
- The GUI uses a backend already listening on this URL, or starts the bundled one on its port
- A change takes effect when the backend restarts (File → Restart Backend)
- Ensure firewall allows connections if using remote backend

### Theme

**Setting**: Color Scheme
**Options**:
- **Auto** (default) - Follow system theme (light/dark)
- **Light** - Always use light theme
- **Dark** - Always use dark theme

**Notes**:
- Auto mode automatically switches when system theme changes
- Theme changes apply immediately
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
- Persisted across sessions; entries whose file no longer exists are dropped when the menu opens

## Settings File

Preferences, recent files, the selected inspector panel and the window geometry are stored as JSON:

| Platform | Location |
|----------|----------|
| Windows | `%APPDATA%\ARMEmulator\settings.json` |
| macOS | `~/Library/Application Support/ARMEmulator/settings.json` |
| Linux | `~/.config/ARMEmulator/settings.json` |

An unreadable or invalid file is replaced by defaults, a message is shown once, and the damaged file is kept as `settings.json.bak`.

## Platform-Specific Settings

### Windows

- Settings stored in the roaming profile
- Native file dialogs respect Windows theme
- Font rendering uses ClearType

### macOS

- Settings stored in `~/Library/Application Support`
- Native menu bar integration
- Automatic dark mode switching based on system appearance
- Font rendering uses Core Text

### Linux

- Settings stored under `~/.config`
- GTK file dialogs on GTK-based desktops
- KDE integration on KDE Plasma
- Font rendering depends on desktop environment

## Keyboard Shortcuts

Keyboard shortcuts cannot currently be customized. See [KEYBOARD_SHORTCUTS.md](KEYBOARD_SHORTCUTS.md) for complete list.

## Resetting to Defaults

Close the application, delete `settings.json`, and start it again.

## Troubleshooting

### Backend Connection Failed

**Issue**: "Cannot connect to backend" error
**Solutions**:
1. Use Retry on the connection screen, or File → Restart Backend
2. Check the backend binary sits next to the application (see the README) and the backend URL in preferences matches running instance
3. Verify firewall allows port 8080 (or configured port)
4. Test connection: `curl http://localhost:8080/api/v1/version`

## See Also

- [README.md](README.md) - Build and run instructions
- [KEYBOARD_SHORTCUTS.md](KEYBOARD_SHORTCUTS.md) - Keyboard shortcut reference
- [INTEGRATION_TESTING.md](INTEGRATION_TESTING.md) - Running integration tests
- [../docs/plans/2026-10-06-avalonia-gui-plan.md](../docs/plans/2026-10-06-avalonia-gui-plan.md) - Plan and status
