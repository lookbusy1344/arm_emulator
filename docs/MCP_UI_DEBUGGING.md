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
