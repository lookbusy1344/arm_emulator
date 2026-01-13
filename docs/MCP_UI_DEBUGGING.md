# UI Debugging with MCP Servers

Updated by Opus 4.5 - 13:23, 8 Jan 2026

This guide covers how to use the Playwright and XcodeBuild MCP (Model Context Protocol) servers for debugging user interfaces across web and native iOS/macOS applications.

## Overview

### Playwright MCP Server
[microsoft/playwright-mcp](https://github.com/microsoft/playwright-mcp)

An accessibility-first browser automation server that enables AI assistants to interact with web applications through structured data rather than screenshots. Uses Playwright's browser automation capabilities to provide deterministic UI testing and debugging.

**Key Strengths:**
- Accessibility tree snapshots for precise element targeting
- Cross-browser support (Chromium, Firefox, WebKit)
- Network inspection and JavaScript evaluation
- No vision model required - operates on structured data

### XcodeBuild MCP Server
[cameroncooke/XcodeBuildMCP](https://github.com/cameroncooke/XcodeBuildMCP)

Integrates Xcode build tools with AI assistants for iOS/macOS development. Provides 63+ tools across 12 workflow groups for builds, testing, device automation, and UI testing.

**Key Strengths:**
- Complete iOS/macOS build and test workflows
- Simulator and physical device management
- UI automation via AXe accessibility framework
- Log capture and analysis
- Swift Package Manager support
- Project scaffolding from templates

## Setup

### Playwright MCP

**Installation via client CLIs** (recommended):
```bash
# Claude Code
claude mcp add playwright npx @playwright/mcp@latest

# Cursor - use the MCP settings UI, or:
# Go to Cursor Settings -> MCP -> Add new MCP Server

# VS Code
code --add-mcp '{"name":"playwright","command":"npx","args":["@playwright/mcp@latest"]}'

# Copilot CLI
/mcp add
```

**Manual Configuration** (`claude_desktop_config.json` or similar):
```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["@playwright/mcp@latest"]
    }
  }
}
```

**Key Command-line Options:**
| Option | Description |
|--------|-------------|
| `--browser <name>` | Browser: chromium (default), firefox, webkit, msedge |
| `--headless` | Run browser in headless mode (no visible window) |
| `--device <name>` | Emulate device, e.g., "iPhone 15" |
| `--viewport-size <WxH>` | Set viewport, e.g., "1280x720" |
| `--caps <list>` | Enable capabilities: vision, pdf, testing |
| `--save-trace` | Save Playwright trace for debugging |
| `--save-video <WxH>` | Record video of session |
| `--timeout-action <ms>` | Action timeout (default 5000ms) |
| `--isolated` | Use isolated profile (no persistent state) |

**Requirements:**
- Node.js 18+
- No browser installation needed (Playwright downloads automatically)

### XcodeBuild MCP

## NOTE - this fixed installation

claude mcp add --transport stdio XcodeBuildMCP -- npx -y xcodebuildmcp@latest

For more details see:

https://github.com/keskinonur/claude-code-ios-dev-guide?tab=readme-ov-file#7-xcodebuildmcp-integration


**Installation via Smithery** (recommended):
```bash
# For Claude Code
npx -y @smithery/cli@latest install cameroncooke/xcodebuildmcp --client claude-code

# For Claude Desktop
npx -y @smithery/cli@latest install cameroncooke/xcodebuildmcp --client claude

# For Cursor
npx -y @smithery/cli@latest install cameroncooke/xcodebuildmcp --client cursor

# For VS Code
npx -y @smithery/cli@latest install cameroncooke/xcodebuildmcp --client vscode
```

**Manual installation** (Claude Code):
```bash
claude mcp add --transport stdio XcodeBuildMCP -- npx -y xcodebuildmcp@latest
```

**AXe Installation** (required for UI automation):
```bash
brew install cameroncooke/axe/axe
```

> **Note**: You may see `Failed to fix install linkage` errors during installation - these can be safely ignored as the binaries are already codesigned.

**Requirements:**
- macOS 14.5+
- Xcode 16.x+
- Node.js 18.x+
- AXe (for UI automation, install separately)

## AXe: iOS Simulator UI Automation

[AXe](https://github.com/cameroncooke/axe) is a CLI tool for iOS Simulator automation using Apple's Accessibility APIs and HID (Human Interface Device) functionality. XcodeBuild MCP uses AXe under the hood for its UI testing tools.

### Why AXe?

- **Single binary** - No server/daemon required, just a standalone CLI
- **Direct accessibility access** - Uses Apple's accessibility APIs for reliable element targeting
- **Complete HID coverage** - Full touch, gesture, keyboard, and hardware button simulation
- **Scriptable** - Easy to integrate into automation workflows

### Installation

```bash
# Install via Homebrew (recommended)
brew install cameroncooke/axe/axe

# Verify installation
axe --help

# List available simulators
axe list-simulators
```

### Direct CLI Usage

While XcodeBuild MCP wraps AXe tools, you can also use AXe directly:

#### Touch & Tap
```bash
# Get simulator UDID first
UDID=$(xcrun simctl list devices booted -j | jq -r '.devices[][] | select(.state=="Booted") | .udid' | head -1)

# Tap at coordinates
axe tap -x 100 -y 200 --udid $UDID

# Tap by accessibility identifier
axe tap --id "LoginButton" --udid $UDID

# Tap by accessibility label
axe tap --label "Sign In" --udid $UDID

# Tap with timing controls
axe tap -x 100 -y 200 --pre-delay 1.0 --post-delay 0.5 --udid $UDID
```

#### Swipe & Gestures
```bash
# Custom swipe
axe swipe --start-x 100 --start-y 500 --end-x 100 --end-y 100 --udid $UDID

# Gesture presets (much easier!)
axe gesture scroll-up --udid $UDID
axe gesture scroll-down --udid $UDID
axe gesture swipe-from-left-edge --udid $UDID   # Back navigation
axe gesture swipe-from-bottom-edge --udid $UDID # Open/reveal
```

#### Text Input
```bash
# Type text (use single quotes for special chars)
axe type 'Hello World!' --udid $UDID

# From stdin (best for automation)
echo "user@example.com" | axe type --stdin --udid $UDID

# From file
axe type --file credentials.txt --udid $UDID
```

#### Hardware Buttons
```bash
axe button home --udid $UDID
axe button lock --duration 2.0 --udid $UDID
axe button siri --udid $UDID
axe button apple-pay --udid $UDID
```

#### Screenshots & Video
```bash
# Screenshot (auto-generates filename)
axe screenshot --udid $UDID

# Screenshot to specific path
axe screenshot --output ~/Desktop/test.png --udid $UDID

# Record video to MP4
axe record-video --udid $UDID --fps 15 --output recording.mp4
# Press Ctrl+C to stop recording
```

#### Accessibility Inspection
```bash
# Get full UI hierarchy (critical for automation!)
axe describe-ui --udid $UDID

# Get element at specific point
axe describe-ui --point 100,200 --udid $UDID
```

### Gesture Presets Reference

| Preset | Description | Use Case |
|--------|-------------|----------|
| `scroll-up` | Scroll up in center | Content navigation |
| `scroll-down` | Scroll down in center | Content navigation |
| `scroll-left` | Scroll left in center | Horizontal scrolling |
| `scroll-right` | Scroll right in center | Horizontal scrolling |
| `swipe-from-left-edge` | Left edge to right | Back navigation |
| `swipe-from-right-edge` | Right edge to left | Forward navigation |
| `swipe-from-top-edge` | Top to bottom | Dismiss/close |
| `swipe-from-bottom-edge` | Bottom to top | Open/reveal |

### AXe vs XcodeBuild MCP UI Tools

| Approach | When to Use |
|----------|-------------|
| **XcodeBuild MCP** | Within AI assistant sessions, integrated workflows |
| **AXe CLI directly** | Shell scripts, CI pipelines, standalone automation |

Both use the same underlying automation - XcodeBuild MCP's `tap`, `swipe`, `describe_ui` etc. are wrappers around AXe commands.

## Playwright MCP Tools Reference

Playwright MCP provides ~20 tools for browser automation. The key principle is **accessibility-first**: use `browser_snapshot` to get element references, then interact using those refs.

### Core Tools

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `browser_navigate` | Go to URL | `url` |
| `browser_snapshot` | Get accessibility tree (use this!) | `filename` (optional) |
| `browser_click` | Click element | `element`, `ref` |
| `browser_type` | Type into input | `element`, `ref`, `text`, `submit` |
| `browser_hover` | Hover over element | `element`, `ref` |
| `browser_select_option` | Select dropdown option | `element`, `ref`, `values` |
| `browser_fill_form` | Fill multiple form fields | `fields` (array) |
| `browser_press_key` | Press keyboard key | `key` (e.g., "Enter", "Tab") |
| `browser_drag` | Drag and drop | `startRef`, `endRef` |

### Inspection & Debugging

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `browser_console_messages` | Get console output | `level` (error/warning/info/debug) |
| `browser_network_requests` | List network requests | `includeStatic` |
| `browser_evaluate` | Run JavaScript | `function` |
| `browser_take_screenshot` | Capture image | `filename`, `fullPage` |

### Navigation & Tabs

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `browser_navigate_back` | Go back | - |
| `browser_tabs` | List/create/close tabs | `action` (list/new/close/select) |
| `browser_close` | Close browser | - |
| `browser_resize` | Resize viewport | `width`, `height` |
| `browser_wait_for` | Wait for condition | `text`, `textGone`, `time` |

### Optional Capabilities

Enable with `--caps=vision,pdf,testing`:

| Tool | Capability | Description |
|------|------------|-------------|
| `browser_mouse_click_xy` | vision | Click at coordinates |
| `browser_mouse_move_xy` | vision | Move mouse to coordinates |
| `browser_mouse_drag_xy` | vision | Drag between coordinates |
| `browser_pdf_save` | pdf | Save page as PDF |
| `browser_verify_element_visible` | testing | Assert element visible |
| `browser_verify_text_visible` | testing | Assert text visible |
| `browser_generate_locator` | testing | Generate test locator |

### Understanding Snapshots

The `browser_snapshot` tool returns an **accessibility tree**, not a screenshot. Example output:

```
- document "My App"
  - navigation "Main"
    - link "Home" [ref=e1]
    - link "About" [ref=e2]
  - main
    - heading "Welcome" [ref=e3]
    - textbox "Email" [ref=e4]
    - button "Submit" [ref=e5]
```

Use the `ref` values (e.g., `e4`, `e5`) in subsequent tool calls:
```json
{"element": "Email textbox", "ref": "e4", "text": "user@example.com"}
```

### Profile Management

**Persistent profile** (default) - Keeps login sessions, cookies:
```bash
# Profile location (macOS)
~/Library/Caches/ms-playwright/mcp-chromium-profile
```

**Isolated profile** - Fresh state each time:
```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["@playwright/mcp@latest", "--isolated"]
    }
  }
}
```

**With saved auth state**:
```json
{
  "args": ["@playwright/mcp@latest", "--isolated", "--storage-state=auth.json"]
}
```

### Session Recording

**Save trace for debugging**:
```json
{
  "args": ["@playwright/mcp@latest", "--save-trace", "--output-dir=./traces"]
}
```

View traces at https://trace.playwright.dev

**Record video**:
```json
{
  "args": ["@playwright/mcp@latest", "--save-video=1280x720", "--output-dir=./videos"]
}
```

### Configuration File

For complex setups, use a config file (`playwright-mcp.json`):
```json
{
  "browser": {
    "browserName": "chromium",
    "launchOptions": { "headless": false },
    "contextOptions": { "viewport": { "width": 1280, "height": 720 } }
  },
  "capabilities": ["core", "pdf"],
  "saveTrace": true,
  "outputDir": "./playwright-output",
  "timeouts": {
    "action": 10000,
    "navigation": 60000
  }
}
```

Then: `npx @playwright/mcp@latest --config playwright-mcp.json`

## Common Workflows

### Web UI Debugging with Playwright

#### 1. Navigate and Capture Snapshot

The key workflow: **navigate → snapshot → interact using refs**

```
AI: browser_navigate url="http://localhost:5173"
AI: browser_snapshot
    → Returns accessibility tree with refs like [ref=e1], [ref=e2]
AI: browser_click element="Submit button" ref="e5"
```

The snapshot provides a structured view of all interactive elements with unique identifiers (`ref`), making it easy to target specific UI components without screenshots.

#### 2. Interact with Elements

**Common Interactions** (using refs from snapshot):
```
browser_click element="Submit button" ref="e5"
browser_type element="Email input" ref="e4" text="user@example.com"
browser_select_option element="Country dropdown" ref="e7" values=["US"]
browser_fill_form fields=[{name:"Email", ref:"e4", type:"textbox", value:"test@example.com"}]
```

**Best Practice:** Always capture a snapshot first to get valid element refs.

#### 3. Inspect Network & Console

```
browser_network_requests                    # See all API calls
browser_network_requests includeStatic=true # Include images, scripts
browser_console_messages level="error"      # Get console errors
```

Useful for debugging API calls, resource loading, and JavaScript errors.

#### 4. Execute JavaScript

```
browser_evaluate function="() => document.querySelector('#result').textContent"
browser_evaluate function="() => localStorage.getItem('token')"
browser_evaluate function="(el) => el.innerHTML" element="Results div" ref="e10"
```

#### 5. Screenshots and PDFs

```
browser_take_screenshot                          # Viewport screenshot
browser_take_screenshot fullPage=true            # Full page
browser_take_screenshot element="Chart" ref="e8" # Element only
browser_pdf_save filename="report.pdf"           # Requires --caps=pdf
```

### iOS/macOS Development with XcodeBuild MCP

XcodeBuild MCP organizes its 63+ tools into workflow groups. Key groups:
- **project-discovery**: Find projects, list schemes, examine build settings
- **session-management**: Set defaults for project, scheme, simulator
- **simulator**: Build, run, test on iOS simulators
- **macos**: Build, run, test macOS apps
- **device**: Build, deploy, test on physical devices
- **ui-testing**: Screenshots, gestures, accessibility inspection
- **logging**: Capture app and system logs

#### 1. Set Session Defaults (Critical First Step)

```bash
# Set project and scheme once - used by all subsequent calls
session_set_defaults projectPath="/path/to/MyApp.xcodeproj" scheme="MyApp"

# Or for workspace-based projects
session_set_defaults workspacePath="/path/to/MyApp.xcworkspace" scheme="MyApp"

# Include simulator if targeting iOS
session_set_defaults projectPath="/path/to/MyApp.xcodeproj" scheme="MyApp" simulatorName="iPhone 16"
```

#### 2. Build and Run

**macOS:**
```bash
build_macos                    # Build the app
launch_mac_app appPath="..."   # Launch with optional args
stop_mac_app name="MyApp"      # Stop by name or PID
```

**iOS Simulator:**
```bash
list_sims                      # Find available simulators
build_sim                      # Build for simulator
launch_app_sim                 # Launch in simulator
build_run_sim                  # Build and run in one step
```

#### 3. Run Tests

```bash
# macOS tests
test_macos

# iOS Simulator tests  
test_sim

# Physical device tests
test_device
```

#### 4. UI Automation (Requires AXe)

```bash
# Get precise element coordinates (don't guess from screenshots!)
describe_ui

# Tap at coordinates
tap x=100 y=200

# Tap by accessibility label
tap accessibilityId="LoginButton"

# Type text
type_text text="hello@example.com"

# Gestures
gesture gesture="scroll-down"
swipe startX=100 startY=500 endX=100 endY=100

# Screenshot for visual verification
screenshot
```

#### 5. Log Capture (Programmatic Debug Log Access)

XcodeBuild MCP provides programmatic access to app logs - extremely useful for automated debugging workflows.

**Available Tools:**

| Tool | Platform | Description |
|------|----------|-------------|
| `start_sim_log_cap` | Simulator | Start capturing logs, returns session ID |
| `stop_sim_log_cap` | Simulator | Stop capture and **return logs** |
| `launch_app_logs_sim` | Simulator | Launch app with log capture in one step |
| `start_device_log_cap` | Device | Start capturing device logs |
| `stop_device_log_cap` | Device | Stop and return device logs |

**Key Parameters:**

```bash
# Basic log capture (structured os_log only)
start_sim_log_cap bundleId="com.example.MyApp"

# Capture console output (print/NSLog/DebugLog) - RECOMMENDED for debugging
start_sim_log_cap bundleId="com.example.MyApp" captureConsole=true
```

> **Important:** `captureConsole: true` captures `print()`, `NSLog()`, and custom debug utilities like `DebugLog`. Without it, you only get structured `os_log` entries. Most debug logging ends up in console output.

**Basic Workflow:**

```bash
# 1. Start log capture (returns session ID)
start_sim_log_cap bundleId="com.example.ARMEmulator" captureConsole=true
# Response: { "sessionId": "abc123" }

# 2. Interact with app (run, step, trigger the bug, etc.)
# ...

# 3. Stop and retrieve logs
stop_sim_log_cap logSessionId="abc123"
# Response: { "logs": "🔵 [ViewModel] Loading program...\n❌ [StackView] Error: ..." }
```

**One-Step Launch with Logs:**

```bash
# Launch app and capture logs simultaneously
launch_app_logs_sim bundleId="com.example.ARMEmulator" args=["test.s"]
```

**Example: Automated Log Analysis**

```bash
#!/bin/bash
# debug_with_logs.sh - Capture and analyze app logs programmatically

# Build and get bundle ID
mcp-cli call XcodeBuildMCP/build_sim '{}'
BUNDLE_ID=$(mcp-cli call XcodeBuildMCP/get_sim_app_path '{}' | jq -r '.bundleId')

# Start log capture with console output
SESSION=$(mcp-cli call XcodeBuildMCP/start_sim_log_cap "{
  \"bundleId\": \"$BUNDLE_ID\",
  \"captureConsole\": true
}" | jq -r '.sessionId')

echo "Log session: $SESSION"

# Launch app
mcp-cli call XcodeBuildMCP/launch_app_sim "{\"bundleId\": \"$BUNDLE_ID\"}"

# Wait for app to do something interesting
sleep 5

# Stop capture and get logs
LOGS=$(mcp-cli call XcodeBuildMCP/stop_sim_log_cap "{\"logSessionId\": \"$SESSION\"}")

# Analyze logs
echo "$LOGS" | jq -r '.logs' > app_debug.log

# Search for errors
echo "=== Errors Found ==="
grep -E "❌|Error|error|failed|Failed" app_debug.log

# Search for specific component
echo "=== StackView Logs ==="
grep "StackView" app_debug.log
```

**Structured Logs vs Console Output:**

| Type | Captured By | Contains |
|------|-------------|----------|
| Structured (`os_log`) | Default | System events, lifecycle, crashes |
| Console | `captureConsole: true` | `print()`, `NSLog()`, `DebugLog`, stdout/stderr |

For debugging custom app logic, you almost always want `captureConsole: true`.

## Integration Patterns for AI Assistants

### Workflow: Debug a Web UI Issue

**Step 1:** Describe the problem to your AI assistant:
```
"The velocity calculator isn't updating when I change the input.
Can you help debug this?"
```

**Step 2:** Assistant uses Playwright MCP:
1. Navigates to the page: `browser_navigate`
2. Takes snapshot: `browser_snapshot`
3. Types into input: `browser_type`
4. Inspects DOM state: `browser_evaluate`
5. Checks network requests: `browser_network_requests`
6. Identifies the issue (e.g., event listener not attached)

### Workflow: Debug an iOS Build Failure

**Step 1:** Describe the problem:
```
"My iOS app won't build - getting Swift compiler errors.
Can you investigate?"
```

**Step 2:** Assistant uses XcodeBuild MCP:
1. Attempts build: `build`
2. Parses error output
3. Examines specific files mentioned in errors
4. Suggests fixes based on diagnostics
5. Runs clean build to verify

## Automated Debugging Workflows (Swift GUI)

**Problem**: Manual debugging of Swift macOS apps is tedious:
- Build in Xcode → Launch → Step through UI → Check console logs → Repeat
- Hard to correlate UI state with backend behavior
- Time-consuming to reproduce specific scenarios

**Solution**: Automate the build-launch-debug cycle using XcodeBuild MCP + direct API testing.

### Complete Automation Example: Stack View Bug

This workflow debugged the Stack View display issue in the ARM Emulator Swift GUI.

#### 1. Set Up Session Defaults

**CRITICAL**: Set session defaults once to avoid repeating parameters:

```
session_set_defaults projectPath="/path/to/ARMEmulator.xcodeproj" scheme="ARMEmulator"
```

#### 2. Build and Launch

```
build_macos                                    # Build the app
get_mac_app_path                               # Returns app bundle path
launch_mac_app appPath="..." args=["test.s"]   # Launch with test file
```

**Benefit**: App launches with test program already loaded - no manual "Open File" needed.

#### 3. Debug with Console Logs

The Swift app uses `DebugLog` utility that outputs to stdout. When launched via MCP, these logs appear in the terminal:

```
🔵 [StackView] loadStack() called, SP = 0x00050000
🔵 [StackView] Fetching memory from 0x0004FFC0, length: 128 bytes
❌ [StackView] Failed to fetch stack memory: Server error (500):
    "memory access violation: address 0x00050000 is not mapped"
```

**Key Insight**: This immediately revealed the Stack View was trying to read unmapped memory above SP.

#### 4. Combine with Backend API Testing

While the GUI runs, directly test the backend API to isolate issues:

```bash
# Create session
SESSION=$(curl -s -X POST http://localhost:8080/api/v1/session \
  -H "Content-Type: application/json" -d '{}' | jq -r '.sessionId')

# Load program
curl -s -X POST "http://localhost:8080/api/v1/session/$SESSION/load" \
  -H "Content-Type: application/json" \
  -d "{\"source\": $(cat fibonacci.s | jq -Rs .)}" | jq .

# Step through program
curl -s -X POST "http://localhost:8080/api/v1/session/$SESSION/step"

# Check registers
curl -s "http://localhost:8080/api/v1/session/$SESSION/registers" | jq .

# Test the problematic memory fetch
SP=327680
START=$((SP - 64))
curl -s "http://localhost:8080/api/v1/session/$SESSION/memory?address=$START&length=128"
```

**Benefit**: Confirms whether the issue is in Swift GUI code or backend API.

#### 5. Scripted Debugging Loop

Create a script to automate the entire workflow:

```bash
#!/bin/bash
# debug_stack_view.sh

set -e

echo "=== Building Swift GUI ==="
mcp-cli call XcodeBuildMCP/build_macos '{}'

echo "=== Stopping old instances ==="
killall ARMEmulator 2>/dev/null || true
sleep 1

echo "=== Launching app ==="
APP_PATH=$(mcp-cli call XcodeBuildMCP/get_mac_app_path '{}' | grep -oE '/.*\.app')
mcp-cli call XcodeBuildMCP/launch_mac_app "{
  \"appPath\": \"$APP_PATH\",
  \"args\": [\"$(pwd)/examples/fibonacci.s\"]
}" &

# Give app time to start backend
sleep 3

echo "=== Creating API session ==="
SESSION=$(curl -s -X POST http://localhost:8080/api/v1/session \
  -H "Content-Type: application/json" -d '{}' | jq -r '.sessionId')

echo "=== Loading fibonacci.s ==="
curl -s -X POST "http://localhost:8080/api/v1/session/$SESSION/load" \
  -H "Content-Type: application/json" \
  -d "{\"source\": $(cat examples/fibonacci.s | jq -Rs .)}" | jq .

echo "=== Stepping 3 times to trigger stack operations ==="
for i in 1 2 3; do
  curl -s -X POST "http://localhost:8080/api/v1/session/$SESSION/step" > /dev/null
  echo "Step $i done"
done

echo "=== Checking register state ==="
curl -s "http://localhost:8080/api/v1/session/$SESSION/registers" | jq '{pc, sp, lr}'

echo "=== Testing stack memory fetch ==="
SP=$(curl -s "http://localhost:8080/api/v1/session/$SESSION/registers" | jq -r '.sp')
START=$((SP - 64))
END=$((SP + 64))

echo "Attempting to read from $START to $END (crosses SP=$SP)"
curl -s "http://localhost:8080/api/v1/session/$SESSION/memory?address=$START&length=128" || echo "FAILED as expected"

echo ""
echo "=== Check Xcode console for [StackView] debug logs ==="
```

**Usage**:
```bash
chmod +x debug_stack_view.sh
./debug_stack_view.sh
```

### Benefits of Automated Workflows

1. **Reproducibility**: Exact same steps every time
2. **Speed**: Full cycle takes ~10 seconds vs 2+ minutes manually
3. **Isolation**: Test GUI and backend independently
4. **Documentation**: Script serves as executable documentation
5. **CI/CD Ready**: Same scripts work in automation pipelines

### Troubleshooting Automated Workflows

**Problem**: App doesn't launch
- **Check**: Backend binary built? `ls -la arm-emulator`
- **Fix**: `cd .. && make build`

**Problem**: API connection refused
- **Check**: Backend started? Look for "API server starting on http://127.0.0.1:8080"
- **Fix**: Add `sleep 3` after launch to let backend initialize

**Problem**: No debug logs appear
- **Check**: Building in Debug configuration? Release strips DebugLog calls
- **Fix**: `mcp-cli call XcodeBuildMCP/build_macos '{"configuration": "Debug"}'`

**Problem**: Session not found
- **Check**: Backend process alive? `ps aux | grep arm-emulator`
- **Fix**: Kill and relaunch app

### When to Use Automation vs Manual

**Use Automation When:**
- Reproducing a specific bug scenario repeatedly
- Testing multiple edge cases quickly
- Verifying a fix works across different inputs
- Running regression tests

**Use Manual Debugging When:**
- Exploring unknown behavior
- UI layout/visual issues
- Gesture/interaction testing
- Initial bug triage

## Best Practices

### For Playwright MCP

1. **Always capture snapshots first** - This provides context and correct selectors
2. **Use accessibility selectors** - Prefer ARIA roles and labels over CSS classes
3. **Check network timing** - Use `browser_network_requests` to identify slow resources
4. **Leverage headless mode** - Faster for automated debugging without visual inspection
5. **Record traces** - Use trace recording for complex interaction sequences

### For XcodeBuild MCP

1. **Set session defaults first** - Always call `session_set_defaults` before other tools
2. **Use `describe_ui` before interactions** - Never guess coordinates from screenshots
3. **Clean builds for fresh state** - Use `clean` when debugging build issues
4. **Capture logs for debugging** - Use `start_sim_log_cap`/`stop_sim_log_cap` for diagnostics
5. **Check `doctor` output** - Run `doctor` to verify environment and dependencies

### General MCP Usage

1. **Start with snapshots/describe_ui** - Always get current state before interacting
2. **Use structured parameters** - Properly format all parameters 
3. **Start simple** - Test basic operations before complex workflows
4. **Document your MCP config** - Keep notes on which servers are configured and why

## Troubleshooting

### Playwright MCP Issues

**Problem:** "Server not responding"
- **Solution:** Verify installation: `npx @playwright/mcp@latest --version`
- Check Claude Desktop config JSON syntax

**Problem:** "Element not found"
- **Solution:** Capture fresh snapshot to get updated selectors
- Ensure page has loaded completely (check network idle)

**Problem:** "Timeout waiting for element"
- **Solution:** Increase timeout in configuration
- Check if element is in iframe or shadow DOM

### XcodeBuild MCP Issues

**Problem:** "Command not found" or tools not working
- **Solution:** Run `doctor` to check environment and dependencies
- Verify Xcode Command Line Tools: `xcode-select --install`
- Check macOS 14.5+ and Xcode 16.x+ requirements

**Problem:** "Build failed with signing errors"
- **Solution:** Configure code signing in Xcode before using device tools
- See [docs/DEVICE_CODE_SIGNING.md](https://github.com/cameroncooke/XcodeBuildMCP/blob/main/docs/DEVICE_CODE_SIGNING.md)

**Problem:** "Simulator not available"
- **Solution:** Use `list_sims` to see available simulators
- Boot simulator with `boot_sim` or open Xcode to install more

**Problem:** UI automation not working
- **Solution:** Install AXe: `brew install cameroncooke/axe/axe`
- Ensure simulator is booted and app is running

## Advanced Techniques

### Playwright: Device Emulation

```
# Configure in MCP args for persistent emulation
["@playwright/mcp@latest", "--device=iPhone 15"]

# Or resize dynamically
browser_resize width=375 height=667
browser_snapshot  # Check mobile layout
```

### Playwright: Debugging Form Submissions

```
# Fill form using refs from snapshot
browser_fill_form fields=[
  {name:"Velocity", ref:"e4", type:"textbox", value:"0.8c"},
  {name:"Distance", ref:"e5", type:"textbox", value:"10"},
  {name:"Units", ref:"e6", type:"combobox", value:"lightyears"}
]

# Watch for form submission in network
browser_network_requests  # Look for POST requests
```

### Playwright: Handling Dialogs & Waits

```
# Handle alert/confirm/prompt dialogs
browser_handle_dialog accept=true
browser_handle_dialog accept=true promptText="user input"

# Wait for conditions
browser_wait_for text="Loading complete"
browser_wait_for textGone="Please wait..."
browser_wait_for time=2  # Wait 2 seconds
```

### Playwright: Tab Management

```
browser_tabs action="list"           # See all tabs
browser_tabs action="new"            # Open new tab
browser_tabs action="select" index=0 # Switch to first tab
browser_tabs action="close"          # Close current tab
```

### XcodeBuild: Swift Package Manager

```bash
# Build a Swift package
swift_package_build

# Run package tests
swift_package_test

# Run an executable target
swift_package_run executableName="MyTool"

# Clean build artifacts
swift_package_clean
```

### XcodeBuild: Project Scaffolding

```bash
# Create new iOS project from template
scaffold_ios_project projectName="MyApp" organizationIdentifier="com.example"

# Create new macOS project from template
scaffold_macos_project projectName="MyMacApp" organizationIdentifier="com.example"
```

## Resources

### Documentation
- [Playwright MCP GitHub](https://github.com/microsoft/playwright-mcp)
- [XcodeBuild MCP GitHub](https://github.com/cameroncooke/XcodeBuildMCP)
- [XcodeBuild MCP Tools Reference](https://github.com/cameroncooke/XcodeBuildMCP/blob/main/docs/TOOLS.md)
- [AXe GitHub](https://github.com/cameroncooke/axe) - iOS Simulator UI automation CLI
- [MCP Specification](https://modelcontextprotocol.io/)
- [Playwright Documentation](https://playwright.dev/)
- [Xcode Build Settings Reference](https://developer.apple.com/documentation/xcode)

### Related Tools
- **Claude Code CLI** - AI assistant with MCP integration
- **Cursor** - IDE with MCP support
- **VS Code** - MCP extensions available
- **Smithery** - MCP server registry and installer

## Summary

| Feature | Playwright MCP | XcodeBuild MCP |
|---------|----------------|----------------|
| **Target Platform** | Web (all browsers) | iOS/macOS native |
| **Automation Type** | Browser automation | Build, test, UI automation |
| **UI Inspection** | Accessibility tree | AXe accessibility framework |
| **Best For** | Web app debugging | Native app development |
| **UI Testing** | ✅ Full support | ✅ Via AXe (gestures, taps, screenshots) |
| **Network Analysis** | ✅ Full support | ❌ Not available |
| **Log Capture** | ✅ Console messages | ✅ App and system logs |
| **CI/CD Ready** | ✅ Headless mode | ✅ Command-line builds |
| **Tool Count** | ~20 tools | 63+ tools |

Both servers complement each other: use Playwright for web interfaces and XcodeBuild for native iOS/macOS applications.
