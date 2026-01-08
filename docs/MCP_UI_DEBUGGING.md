# UI Debugging with MCP Servers

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

Integrates Xcode build tools with AI assistants for iOS/macOS development. Provides programmatic access to builds, testing, and device automation.

**Key Strengths:**
- Native integration with Xcode toolchain
- Device management and code signing
- UI automation for iOS/macOS apps
- Build diagnostics and error reporting

## Setup

### Playwright MCP

**Installation:**
```bash
npx @playwright/mcp@latest
```

**Configuration in Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["-y", "@playwright/mcp@latest"]
    }
  }
}
```

**Command-line Options:**
- `--browser` - Choose browser: chromium (default), firefox, webkit
- `--timeout` - Set default timeout in milliseconds
- `--headless` - Run browser in headless mode

### XcodeBuild MCP

This fixed installation:

claude mcp add --transport stdio XcodeBuildMCP -- npx -y xcodebuildmcp@latest

https://github.com/keskinonur/claude-code-ios-dev-guide?tab=readme-ov-file#7-xcodebuildmcp-integration

**Installation via Smithery:**
```bash
npx -y @smithery/cli@latest install cameroncooke/xcodebuildmcp --client claude
```

**Requirements:**
- macOS 14.5+
- Xcode 16.x+
- Node.js 18.x+
- AXe (for UI automation)

## Discovering Available Tools

Once configured, use the `mcp-cli` command to explore available tools:

```bash
# List all connected MCP servers
mcp-cli servers

# List all available tools
mcp-cli tools

# Search for specific functionality
mcp-cli grep "browser"
mcp-cli grep "build"

# Get detailed schema for a tool (REQUIRED before using any tool)
mcp-cli info <server>/<tool>
```

## Common Workflows

### Web UI Debugging with Playwright

#### 1. Navigate to a Page and Capture Snapshot

**Typical Flow:**
```bash
# First, check the schema
mcp-cli info playwright/browser_navigate

# Navigate to the page
mcp-cli call playwright/browser_navigate '{"url": "http://localhost:5173"}'

# Capture accessibility snapshot
mcp-cli call playwright/browser_snapshot '{}'
```

The snapshot provides a structured view of all interactive elements with unique identifiers, making it easy to target specific UI components.

#### 2. Interact with Elements

**Common Interactions:**
- **Click:** `mcp-cli call playwright/browser_click '{"selector": "button[type=submit]"}'`
- **Type:** `mcp-cli call playwright/browser_type '{"selector": "input#velocity", "text": "0.5c"}'`
- **Select:** `mcp-cli call playwright/browser_select '{"selector": "select#units", "values": ["kilometers"]}'`

**Best Practice:** Always capture a snapshot first to identify correct selectors.

#### 3. Inspect Network Activity

```bash
# Check schema first
mcp-cli info playwright/browser_network_requests

# Inspect all network requests
mcp-cli call playwright/browser_network_requests '{}'
```

Useful for debugging API calls, resource loading, and timing issues.

#### 4. Execute JavaScript for Debugging

```bash
# Evaluate JavaScript in page context
mcp-cli call playwright/browser_evaluate '{
  "script": "document.querySelector(\"#result\").textContent"
}'
```

#### 5. Capture Screenshots and PDFs

```bash
# Take screenshot of current state
mcp-cli call playwright/browser_screenshot '{}'

# Generate PDF (if supported)
mcp-cli call playwright/browser_pdf '{}'
```

### iOS/macOS UI Debugging with XcodeBuild

#### 1. Build the Project

**Typical Flow:**
```bash
# Check available build tools
mcp-cli tools xcodebuild

# Get schema for build command
mcp-cli info xcodebuild/build

# Execute build
mcp-cli call xcodebuild/build '{
  "scheme": "MyApp",
  "configuration": "Debug"
}'
```

#### 2. Run Tests

```bash
# Check test tool schema
mcp-cli info xcodebuild/test

# Run UI tests
mcp-cli call xcodebuild/test '{
  "scheme": "MyAppUITests",
  "destination": "platform=iOS Simulator,name=iPhone 15"
}'
```

#### 3. Device Management

```bash
# List available simulators/devices
mcp-cli call xcodebuild/list_devices '{}'

# Configure code signing
mcp-cli call xcodebuild/configure_signing '{
  "target": "MyApp",
  "team": "TEAM_ID"
}'
```

#### 4. UI Automation

**Note:** Requires AXe installation for accessibility-based automation.

```bash
# Interact with UI elements through automation
mcp-cli call xcodebuild/ui_automation '{
  "action": "tap",
  "element": "button:Login"
}'
```

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

```bash
# Check schema first
mcp-cli info XcodeBuildMCP/session-set-defaults

# Set project and scheme
mcp-cli call XcodeBuildMCP/session-set-defaults '{
  "projectPath": "/path/to/ARMEmulator.xcodeproj",
  "scheme": "ARMEmulator"
}'
```

#### 2. Build and Launch Automatically

```bash
# Build the app
mcp-cli call XcodeBuildMCP/build_macos '{}'

# Get app path
APP_PATH=$(mcp-cli call XcodeBuildMCP/get_mac_app_path '{}' | grep -oE '/.*\.app')

# Launch with test file pre-loaded
mcp-cli call XcodeBuildMCP/launch_mac_app "{
  \"appPath\": \"$APP_PATH\",
  \"args\": [\"/path/to/test/fibonacci.s\"]
}"
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

### Advanced: Parsing Console Output

Filter debug logs programmatically:

```bash
# Capture only Stack View logs
mcp-cli call XcodeBuildMCP/launch_mac_app "{...}" 2>&1 | grep '\[StackView\]'

# Count load attempts
mcp-cli call XcodeBuildMCP/launch_mac_app "{...}" 2>&1 | grep 'loadStack() called' | wc -l

# Detect failures
mcp-cli call XcodeBuildMCP/launch_mac_app "{...}" 2>&1 | grep '❌' | tail -1
```

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

1. **Start with clean builds** - Clear derived data when debugging build issues
2. **Check device availability** - Verify simulators/devices before running tests
3. **Use specific schemes** - Target exactly what you need to test
4. **Review full logs** - Build output often contains hints in warnings
5. **Validate code signing** - Many issues stem from provisioning profile problems

### General MCP Usage

1. **Check schemas first** - Always run `mcp-cli info <server>/<tool>` before using a tool
2. **Use structured JSON** - Properly format all parameters as valid JSON
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

**Problem:** "Command not found"
- **Solution:** Verify Xcode Command Line Tools: `xcode-select --install`
- Check macOS and Xcode version requirements

**Problem:** "Build failed with signing errors"
- **Solution:** Use `configure_signing` tool to set up certificates
- Verify team ID and provisioning profiles

**Problem:** "Simulator not available"
- **Solution:** List devices first to see what's available
- Open Xcode to ensure simulators are properly installed

## Advanced Techniques

### Playwright: Testing Responsive Design

```bash
# Resize window to mobile viewport
mcp-cli call playwright/browser_resize '{
  "width": 375,
  "height": 667
}'

# Take snapshot at different sizes
mcp-cli call playwright/browser_snapshot '{}'
```

### Playwright: Debugging Form Submissions

```bash
# Fill entire form
mcp-cli call playwright/browser_fill_form '{
  "form": "form#calculator",
  "fields": {
    "velocity": "0.8c",
    "distance": "10",
    "units": "lightyears"
  }
}'

# Monitor network for form submission
mcp-cli call playwright/browser_network_requests '{
  "filter": "POST"
}'
```

### XcodeBuild: Continuous Integration Debugging

```bash
# Build with verbose output
mcp-cli call xcodebuild/build '{
  "scheme": "MyApp",
  "configuration": "Release",
  "verbose": true
}'

# Run tests with result bundles for analysis
mcp-cli call xcodebuild/test '{
  "scheme": "MyAppTests",
  "resultBundlePath": "./TestResults"
}'
```

## Resources

### Documentation
- [Playwright MCP GitHub](https://github.com/microsoft/playwright-mcp)
- [XcodeBuild MCP GitHub](https://github.com/cameroncooke/XcodeBuildMCP)
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
| **Automation Type** | Browser automation | Build & test automation |
| **UI Inspection** | Accessibility tree | Xcode UI testing framework |
| **Best For** | Web app debugging | Native app build/test issues |
| **Network Analysis** | ✅ Full support | ❌ Limited |
| **Visual Testing** | ✅ Screenshots/PDFs | ⚠️ Via simulator screenshots |
| **CI/CD Ready** | ✅ Headless mode | ✅ Command-line builds |

Both servers complement each other well: use Playwright for web interfaces and XcodeBuild for native iOS/macOS applications. Together they provide comprehensive UI debugging capabilities across all major platforms.
