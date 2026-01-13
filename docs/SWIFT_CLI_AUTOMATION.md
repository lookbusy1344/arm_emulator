# Swift macOS CLI Automation Guide

A comprehensive guide for automating Swift macOS project development using command-line tools, avoiding Xcode GUI and maximizing automation with Claude CLI and other tools.

## Table of Contents

- [Quick Reference](#quick-reference)
- [Essential Tools](#essential-tools)
- [Project Setup](#project-setup)
- [Dependency Management](#dependency-management)
- [Building](#building)
- [Running Your App](#running-your-app)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Xcode vs CLI vs Alternative IDEs](#xcode-vs-cli-vs-alternative-ides)
- [Alternative Editors and IDE Setup](#alternative-editors-and-ide-setup)
- [Common Workflows](#common-workflows)
- [Pitfalls and Solutions](#pitfalls-and-solutions)
- [Claude CLI Integration](#claude-cli-integration)
- [CI/CD Integration](#cicd-integration)
- [Continuous Delivery and Release Automation](#continuous-delivery-and-release-automation)
- [iOS Development from CLI](#ios-development-from-cli)
- [Distribution and Notarization](#distribution-and-notarization)
- [Performance Profiling from CLI](#performance-profiling-from-cli)
- [Expanded Troubleshooting Guide](#expanded-troubleshooting-guide)

## DEVELOPER_DIR setup

Instead of setting DEVELOPER_DIR before every call eg
    DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer swiftlint lint

We can set it in .zshrc, or generally using:

```
# Check current developer directory
xcode-select -p

# If it's not /Applications/Xcode.app/Contents/Developer, set it:
sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
```

## Quick Reference

Common commands for daily Swift development:

```bash
# === BUILDING ===
# Build the project
xcodebuild -project YourApp.xcodeproj -scheme YourScheme build

# Build with clean output
xcodebuild -project YourApp.xcodeproj -scheme YourScheme build | xcpretty

# Clean build
xcodebuild -project YourApp.xcodeproj -scheme YourScheme clean build

# === TESTING ===
# Run all tests
xcodebuild test -project YourApp.xcodeproj -scheme YourScheme -destination 'platform=macOS'

# Run specific test
xcodebuild test -project YourApp.xcodeproj -scheme YourScheme -only-testing:TestClass/testMethod

# === CODE QUALITY ===
# Format code
swiftformat .

# Lint code
swiftlint

# Fix auto-correctable lint issues
swiftlint --fix

# === DEPENDENCIES ===
# Update all dependencies (Xcode projects)
xcodebuild -resolvePackageDependencies -project YourApp.xcodeproj -scheme YourScheme

# Check for outdated packages
swift-outdated

# === RUNNING ===
# Run the built app
open build/Release/YourApp.app

# Or find and run the built app
find ~/Library/Developer/Xcode/DerivedData -name "YourApp.app" -type d -exec open {} \;

# === CLEANING ===
# Clean project
xcodebuild clean -project YourApp.xcodeproj -scheme YourScheme

# Clean derived data
rm -rf ~/Library/Developer/Xcode/DerivedData

# Clean package caches
rm -rf .swiftpm
rm -rf ~/Library/Caches/org.swift.swiftpm

# === INFORMATION ===
# List schemes
xcodebuild -list -project YourApp.xcodeproj

# Show build settings
xcodebuild -showBuildSettings -project YourApp.xcodeproj -scheme YourScheme

# List available SDKs
xcodebuild -showsdks
```

## Essential Tools

### Complete Installation Guide

Here's everything you need to install for full CLI Swift development. Copy and run the entire block, or install selectively based on your needs.

#### Core Requirements (Install These First)

```bash
# Xcode Command Line Tools - Required for xcodebuild, swift, sourcekit-lsp
xcode-select --install

# Homebrew - macOS package manager (if not already installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### Essential Swift Development Tools

```bash
# SwiftLint - Enforces Swift style and conventions
# Use for: Catching style issues, enforcing team coding standards (popular, install)
brew install swiftlint

# SwiftFormat - Automatically formats Swift code
# Use for: Auto-formatting code to consistent style, pre-commit hooks (popular, install)
brew install swiftformat

# xcpretty - Formats xcodebuild output into readable form
# Use for: Making build/test output human-readable
# brew install xcpretty (old, no main commits for 10 years)

# xcbeautify - Modern alternative to xcpretty (faster, better output)
# Use for: Even better build output formatting than xcpretty (1.3k stars, 65k downloads a month)
brew install xcbeautify
```

#### Code Quality and Analysis Tools

```bash
# Periphery - Detects unused code (dead code elimination)
# Use for: Finding and removing unused classes, functions, variables (6k stars)
brew install periphery

# swift-outdated - Checks for outdated Swift Package dependencies
# Use for: Keeping dependencies up-to-date (low popularity, only 93 installs per month, NOT INSTALLED YET)
brew install swift-outdated
```

#### Project Generation (Optional but Recommended)

```bash
# XcodeGen - Generates Xcode projects from YAML specification
# Use for: Version-control-friendly project management, avoiding .pbxproj conflicts (popular, 10k installs per month)
brew install xcodegen
```

#### Developer Productivity Tools

```bash
# jq - JSON processor for command line
# Use for: Parsing build reports, processing Package.resolved, API responses (install)
brew install jq

# bat - Better 'cat' with syntax highlighting and git integration
# Use for: Viewing source files with syntax highlighting in terminal (popular, install)
brew install bat

# fd - Better 'find' command (faster, simpler syntax)
# Use for: Finding files quickly with intuitive syntax (install)
brew install fd

# ripgrep (rg) - Better 'grep' (faster search)
# Use for: Searching code across entire project (faster than grep) (install)
brew install ripgrep

# fzf - Fuzzy finder for command line
# Use for: Interactive file/command searching and selection (install)
brew install fzf

# fswatch - File system monitor for auto-rebuild on save
# Use for: Automatically rebuilding when files change during development (5k stars, 1.2k installs per month)
brew install fswatch
```

#### Git Enhancement Tools

```bash
# git-delta - Better git diff with syntax highlighting
# Use for: More readable git diffs and code review (28k stars, 3k installs per month)
brew install git-delta

# lazygit - Terminal UI for git
# Use for: Visual git interface without leaving the terminal (install)
brew install lazygit

# gh - GitHub CLI
# Use for: Creating PRs, issues, and GitHub operations from terminal (install)
brew install gh
```

#### Optional: Alternative Editors

```bash
# Visual Studio Code - Lightweight IDE with good Swift support
# Use for: Alternative to Xcode, lighter weight, good for quick edits
brew install --cask visual-studio-code

# Neovim - Modern terminal-based editor
# Use for: Terminal-centric development, SSH work, minimal resource usage
# brew install neovim

# Sublime Text - Fast text editor
# Use for: Quick edits, lightweight IDE alternative
# brew install --cask sublime-text
```

#### Legacy Dependency Managers (Only if Needed)

```bash
# CocoaPods - Older dependency manager (being replaced by SPM)
# Use for: Legacy projects that still use CocoaPods
sudo gem install cocoapods

# Carthage - Decentralized dependency manager
# Use for: Legacy projects that use Carthage
brew install carthage
```

#### Complete One-Command Installation

For a full setup, copy this entire block:

```bash
# Install Xcode Command Line Tools first
xcode-select --install

# Install all essential tools
# brew install \
#   swiftlint \
#   swiftformat \
#   xcpretty \
#   xcbeautify \
#   peripheryapp/periphery/periphery \
#   kiliankoe/formulae/swift-outdated \
#   xcodegen \
#   jq \
#   bat \
#   fd \
#   ripgrep \
#   fzf \
#   fswatch \
#   git-delta \
#   lazygit \
#   gh

# Optional: Install editors
brew install --cask visual-studio-code
# brew install neovim

# Verify installations
echo "Verifying installations..."
swift --version
xcodebuild -version
swiftlint version
swiftformat --version
xcpretty --version
xcbeautify --version
echo "✅ All tools installed successfully!"
```

#### Minimal Setup (Just the Essentials)

If you want just the bare minimum for CLI development:

```bash
# Minimal required tools
brew install swiftlint swiftformat xcbeautify

# Verify
swiftlint version && swiftformat --version && xcbeautify --version
```

### Individual Tool Details

#### xcodebuild
The primary tool for building, testing, and archiving Swift projects from the command line.

```bash
# Install (comes with Xcode Command Line Tools)
xcode-select --install

# Verify installation
xcodebuild -version
```

#### xcpretty
Formats xcodebuild output into readable, colorized output.

```bash
# Install via Homebrew
brew install xcpretty

# Or via RubyGems
gem install xcpretty
```

#### SwiftLint
Enforces Swift style and conventions.

```bash
# Install via Homebrew
brew install swiftlint

# Verify installation
swiftlint version
```

#### SwiftFormat
Automatically formats Swift code according to rules.

```bash
# Install via Homebrew
brew install swiftformat

# Verify installation
swiftformat --version
```

### Optional but Recommended

#### xcbeautify
Modern alternative to xcpretty with better performance.

```bash
brew install xcbeautify
```

#### periphery
Detects unused code in Swift projects.

```bash
brew install peripheryapp/periphery/periphery
```

#### swift-outdated
Checks for outdated Swift Package dependencies.

```bash
brew install kiliankoe/formulae/swift-outdated
```

## Project Setup

### Creating a New Project from Command Line

While `xcodebuild` doesn't directly create projects, you can:

1. **Use Swift Package Manager** for pure Swift projects:
```bash
mkdir MyProject
cd MyProject
swift package init --type executable  # or library
```

2. **Copy and modify an existing `.xcodeproj`** bundle:
```bash
# This is the practical approach for macOS apps
cp -r Template.xcodeproj MyApp.xcodeproj
# Then edit project.pbxproj manually or with tools
```

3. **Use XcodeGen** (recommended for new projects):
```bash
brew install xcodegen

# Create project.yml
cat > project.yml << 'EOF'
name: MyApp
options:
  bundleIdPrefix: com.yourcompany
targets:
  MyApp:
    type: application
    platform: macOS
    deploymentTarget: "13.0"
    sources:
      - MyApp
    settings:
      PRODUCT_BUNDLE_IDENTIFIER: com.yourcompany.MyApp
EOF

# Generate Xcode project
xcodegen generate
```

### Essential Configuration Files

Create these files in your project root:

#### `.swiftlint.yml`
```yaml
disabled_rules:
  - trailing_comma
  - identifier_name
  - line_length

included:
  - Sources
  - Tests

excluded:
  - .build
  - DerivedData

line_length:
  warning: 120
  error: 200

identifier_name:
  min_length: 1
  excluded:
    - id
    - db
```

#### `.swiftformat`
```
--allman false
--commas always
--trimwhitespace always
--indent 4
--linebreaks lf
--header strip
```

#### `.swift-version`
```
6.2
```

## Dependency Management

### Pure SPM Projects (Package.swift)

For projects with a `Package.swift` file, dependency management is straightforward:

#### Adding Dependencies

```bash
# Add a dependency (Swift 5.6+)
swift package add-dependency https://github.com/realm/realm-swift.git --from 10.0.0

# Or manually edit Package.swift
# Then resolve
swift package resolve
```

#### Updating Dependencies

```bash
# Update all dependencies to latest compatible versions
swift package update

# Update a specific dependency
swift package update realm-swift

# Update and resolve
swift package update && swift package resolve
```

#### Removing Dependencies

```bash
# Manually remove from Package.swift dependencies array
# Then clean and resolve
swift package clean
swift package resolve
```

#### Listing and Inspecting

```bash
# Show dependency tree
swift package show-dependencies

# Show dependency tree as JSON
swift package show-dependencies --format json

# Check for outdated packages
swift-outdated

# Dump package details
swift package dump-package
```

### Xcode Projects with SPM Integration

**Important**: Xcode projects store SPM dependencies in the `.xcodeproj` bundle, not in `Package.swift`. This makes CLI dependency management more difficult.

#### Adding Dependencies (Limited CLI Support)

**Reality Check**: There's no direct `xcodebuild` command to add dependencies to Xcode projects. You have three options:

**Option 1: Use Xcode GUI (Easiest)**
```bash
# Open Xcode and use File → Add Package Dependencies
open YourApp.xcodeproj
```

**Option 2: Edit project.pbxproj Manually (Advanced)**
```bash
# The project.pbxproj file is complex and error-prone to edit manually
# Not recommended unless you know what you're doing
vim YourApp.xcodeproj/project.pbxproj
```

**Option 3: Use XcodeGen (Recommended for New Projects)**
```bash
# Define dependencies in project.yml
cat >> project.yml << 'EOF'
packages:
  Realm:
    url: https://github.com/realm/realm-swift.git
    from: 10.0.0

targets:
  MyApp:
    dependencies:
      - package: Realm
EOF

# Regenerate project
xcodegen generate
```

**Option 4: Script with PlistBuddy (Experimental)**
```bash
# This is complex and fragile, but possible for automation
# Add package reference to project.pbxproj
# Then resolve dependencies
xcodebuild -resolvePackageDependencies -project YourApp.xcodeproj -scheme YourScheme
```

#### Updating Dependencies

```bash
# Update all packages (this works well from CLI)
xcodebuild -resolvePackageDependencies -project YourApp.xcodeproj -scheme YourScheme

# Force clean and re-resolve
rm -rf .swiftpm
rm -rf ~/Library/Developer/Xcode/DerivedData/YourApp-*/SourcePackages
xcodebuild -resolvePackageDependencies -project YourApp.xcodeproj -scheme YourScheme

# Check what packages are resolved
cat YourApp.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved
```

#### Removing Dependencies

**Reality Check**: You must use Xcode GUI or manually edit `project.pbxproj`.

```bash
# Option 1: Open in Xcode
open YourApp.xcodeproj
# Then: Select project → Package Dependencies → Select package → Remove

# Option 2: If using XcodeGen, remove from project.yml and regenerate
# Edit project.yml to remove package
xcodegen generate
```

#### Listing Dependencies

```bash
# View resolved packages
cat YourApp.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved

# Pretty print as JSON
cat YourApp.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved | python3 -m json.tool

# Check for outdated packages (if swift-outdated supports Xcode projects)
swift-outdated

# List package dependencies from resolved file
grep -A 5 '"package"' YourApp.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved
```

#### Checking for Updates

```bash
# Check which packages have updates available
# First, note current versions
cat YourApp.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved | \
  grep -E '(package|version)' | paste - -

# Then resolve to get latest
xcodebuild -resolvePackageDependencies -project YourApp.xcodeproj -scheme YourScheme

# Compare Package.resolved before and after
git diff YourApp.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved
```

### CocoaPods (Legacy)

Some older projects still use CocoaPods:

```bash
# Install CocoaPods
sudo gem install cocoapods

# Install dependencies from Podfile
pod install

# Update dependencies
pod update

# Update specific pod
pod update PodName

# Add a dependency (edit Podfile, then)
pod install

# Remove a dependency (remove from Podfile, then)
pod install

# Check for outdated pods
pod outdated

# Show pod details
pod search PodName
```

**Note**: Always open the `.xcworkspace` file when using CocoaPods, not the `.xcodeproj`:
```bash
xcodebuild -workspace YourApp.xcworkspace -scheme YourScheme build
```

### Carthage (Legacy)

```bash
# Install Carthage
brew install carthage

# Build dependencies from Cartfile
carthage update --platform macOS

# Build specific dependency
carthage update DependencyName --platform macOS

# Use pre-built binaries (faster)
carthage update --use-xcframeworks --platform macOS
```

### Dependency Management Best Practices

```bash
# Always commit Package.resolved / Podfile.lock
git add YourApp.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved
git commit -m "Lock dependency versions"

# Clean everything before resolving issues
rm -rf .swiftpm
rm -rf ~/Library/Developer/Xcode/DerivedData
rm -rf ~/Library/Caches/org.swift.swiftpm
xcodebuild -resolvePackageDependencies -project YourApp.xcodeproj -scheme YourScheme

# Create a script for reproducible builds
cat > scripts/resolve-dependencies.sh << 'EOF'
#!/bin/bash
set -e
echo "Cleaning package caches..."
rm -rf .swiftpm
rm -rf ~/Library/Caches/org.swift.swiftpm
echo "Resolving dependencies..."
xcodebuild -resolvePackageDependencies -project *.xcodeproj -scheme MyScheme
echo "Done!"
EOF
chmod +x scripts/resolve-dependencies.sh
```

## Building

### Basic Build Commands

```bash
# List all schemes
xcodebuild -list -project YourApp.xcodeproj

# Build a specific scheme
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -configuration Debug \
  build

# Build with pretty output
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  clean build | xcpretty

# Build for release
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -configuration Release \
  build
```

### Common Build Options

```bash
# Specify SDK
-sdk macosx

# Specify architecture
-arch arm64  # or x86_64

# Specify destination
-destination 'platform=macOS,arch=arm64'

# Parallel builds
-parallelizeTargets

# Quiet output
-quiet

# Show build settings
-showBuildSettings
```

### Build and Archive

```bash
# Create archive
xcodebuild archive \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -archivePath ./build/YourApp.xcarchive

# Export app
xcodebuild -exportArchive \
  -archivePath ./build/YourApp.xcarchive \
  -exportPath ./build \
  -exportOptionsPlist ExportOptions.plist
```

## Running Your App

### Finding the Built App

After building, you need to locate the compiled `.app` bundle:

```bash
# Option 1: Build to a specific location
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -configuration Release \
  -derivedDataPath ./build \
  build

# Then run it
open ./build/Build/Products/Release/YourApp.app

# Option 2: Find in DerivedData
find ~/Library/Developer/Xcode/DerivedData -name "YourApp.app" -type d -path "*/Build/Products/*" | head -1 | xargs open

# Option 3: Use xcodebuild to get build dir
BUILD_DIR=$(xcodebuild -project YourApp.xcodeproj -scheme YourScheme -showBuildSettings | grep -m 1 "BUILD_DIR" | sed 's/[ ]*BUILD_DIR = //')
open "$BUILD_DIR/YourApp.app"
```

### Running with Arguments

```bash
# Run the app with command-line arguments
open ./build/Build/Products/Release/YourApp.app --args --debug --verbose

# Or run the binary directly
./build/Build/Products/Release/YourApp.app/Contents/MacOS/YourApp --help

# With environment variables
ENV_VAR=value ./build/Build/Products/Release/YourApp.app/Contents/MacOS/YourApp
```

### Running Debug Builds

```bash
# Build for debugging
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -configuration Debug \
  -derivedDataPath ./build \
  build

# Run debug build
open ./build/Build/Products/Debug/YourApp.app

# Or run with debugger (lldb)
lldb ./build/Build/Products/Debug/YourApp.app/Contents/MacOS/YourApp
# In lldb:
# (lldb) run
# (lldb) bt    # backtrace on crash
# (lldb) quit
```

### Watching Logs

```bash
# View Console logs for your app
log stream --predicate 'process == "YourApp"' --level debug

# Or use Console.app
open -a Console

# For print statements to stderr
./build/Build/Products/Debug/YourApp.app/Contents/MacOS/YourApp 2>&1 | tee app.log
```

### Build and Run Script

Create a helper script for quick iteration:

```bash
cat > scripts/build-and-run.sh << 'EOF'
#!/bin/bash
set -e

PROJECT=$(ls *.xcodeproj | head -1)
SCHEME=$(xcodebuild -list -project "$PROJECT" | grep -A 1 "Schemes:" | tail -1 | xargs)
CONFIG="${1:-Debug}"

echo "Building $SCHEME ($CONFIG)..."
xcodebuild -project "$PROJECT" \
  -scheme "$SCHEME" \
  -configuration "$CONFIG" \
  -derivedDataPath ./build \
  build | xcpretty

echo "Running..."
open "./build/Build/Products/$CONFIG/$SCHEME.app"
EOF
chmod +x scripts/build-and-run.sh

# Usage:
./scripts/build-and-run.sh          # Debug build
./scripts/build-and-run.sh Release  # Release build
```

## Testing

### Running Tests

```bash
# Run all tests
xcodebuild test \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -destination 'platform=macOS' | xcpretty

# Run specific test
xcodebuild test \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -destination 'platform=macOS' \
  -only-testing:YourAppTests/TestClass/testMethod

# Skip specific test
xcodebuild test \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -skip-testing:YourAppTests/FlakyTests

# Run tests with coverage
xcodebuild test \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -destination 'platform=macOS' \
  -enableCodeCoverage YES

# Generate coverage report
xcrun llvm-cov show \
  -instr-profile=coverage.profdata \
  YourApp.app/Contents/MacOS/YourApp
```

### Test Plans

If using `.xctestplan` files:

```bash
# Run with test plan
xcodebuild test \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -testPlan YourTestPlan \
  -destination 'platform=macOS'
```

## Code Quality

### Linting with SwiftLint

```bash
# Lint all files
swiftlint

# Lint with autocorrect
swiftlint --fix

# Lint specific paths
swiftlint --path Sources/

# Lint in strict mode (warnings as errors)
swiftlint --strict

# Generate report
swiftlint --reporter json > swiftlint-report.json
```

### Formatting with SwiftFormat

```bash
# Format all Swift files
swiftformat .

# Dry run (show what would change)
swiftformat --dryrun .

# Format specific files
swiftformat Sources/

# Lint mode (check without changing)
swiftformat --lint .

# Format with specific rules
swiftformat --disable unusedArguments .
```

### Combining Lint and Format

```bash
# Format first, then lint
swiftformat . && swiftlint
```

### Dead Code Detection

```bash
# Scan for unused code
periphery scan \
  --project YourApp.xcodeproj \
  --schemes YourScheme \
  --targets YourApp

# More aggressive scan
periphery scan \
  --project YourApp.xcodeproj \
  --schemes YourScheme \
  --targets YourApp \
  --no-retain-public
```

### Parsing Build Errors

Build errors can be hard to read in raw `xcodebuild` output. Here's how to handle them:

#### Using xcpretty/xcbeautify

```bash
# xcpretty formats errors nicely
xcodebuild build -project YourApp.xcodeproj -scheme YourScheme 2>&1 | xcpretty

# xcbeautify is faster and has better formatting
xcodebuild build -project YourApp.xcodeproj -scheme YourScheme 2>&1 | xcbeautify

# Save formatted output while displaying it
xcodebuild build -project YourApp.xcodeproj -scheme YourScheme 2>&1 | tee build.log | xcpretty
```

#### Extract Errors Programmatically

```bash
# Extract just errors (with xcpretty JSON)
xcodebuild build -project YourApp.xcodeproj -scheme YourScheme 2>&1 | \
  xcpretty --report json --output build-report.json

# Parse errors with jq
cat build-report.json | jq '.[] | select(.message_type == "error")'

# Count errors and warnings
echo "Errors: $(grep -c "error:" build.log || echo 0)"
echo "Warnings: $(grep -c "warning:" build.log || echo 0)"

# Extract file paths with errors
grep "error:" build.log | sed 's/:.*//g' | sort -u
```

#### Script to Parse and Display Errors

```bash
cat > scripts/build-check.sh << 'EOF'
#!/bin/bash
set -o pipefail

PROJECT=$(ls *.xcodeproj | head -1)
SCHEME=$(xcodebuild -list -project "$PROJECT" | grep -A 1 "Schemes:" | tail -1 | xargs)

echo "Building $SCHEME..."
if xcodebuild build \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  2>&1 | tee /tmp/build.log | xcbeautify; then
  echo "✅ Build succeeded"
  exit 0
else
  echo "❌ Build failed"
  echo ""
  echo "Errors found:"
  grep "error:" /tmp/build.log | head -10
  exit 1
fi
EOF
chmod +x scripts/build-check.sh
```

#### Understanding Common Error Patterns

```bash
# Swift compiler errors
grep "error:.*\.swift:" build.log

# Linker errors
grep "Undefined symbols" build.log -A 5

# Code signing errors
grep "errSecInternalComponent" build.log

# Dependency resolution errors
grep "error: Dependencies could not be resolved" build.log
```

## Xcode vs CLI vs Alternative IDEs

### Can You Still Use Xcode with CLI Automation?

**Absolutely!** CLI automation and Xcode are **not mutually exclusive**. In fact, they complement each other well.

#### The Hybrid Approach (Recommended for Most Teams)

Most professional Swift developers use a hybrid workflow:

```bash
# Use Xcode for:
# - Visual debugging
# - Interface Builder / SwiftUI previews
# - Adding/removing SPM dependencies
# - Complex refactoring with visual feedback
# - Memory graph debugging
# - Instruments profiling

# Use CLI for:
# - CI/CD pipelines
# - Automated testing
# - Code formatting and linting (pre-commit hooks)
# - Build scripts
# - AI-assisted development (Claude CLI)
```

**Example hybrid workflow:**

1. Developer codes in Xcode with visual feedback
2. Pre-commit hook runs `swiftformat` and `swiftlint` automatically
3. Developer commits changes
4. CI/CD runs CLI build and test pipeline
5. Claude CLI can read/modify code and run tests via CLI
6. Developer debugs complex issues in Xcode

#### Making Xcode Projects CLI-Friendly

To ensure your Xcode project works well with CLI automation:

```bash
# 1. Share your schemes (critical!)
# In Xcode: Product → Scheme → Manage Schemes → Check "Shared"
# This makes schemes visible to xcodebuild

# Verify schemes are shared:
ls -la YourApp.xcodeproj/xcshareddata/xcschemes/

# 2. Disable automatic scheme generation
# In Xcode: File → Project Settings → Uncheck "Autocreate schemes"

# 3. Commit scheme files to git
git add YourApp.xcodeproj/xcshareddata/xcschemes/*.xcscheme
git commit -m "Share Xcode schemes for CLI access"

# 4. Use consistent build settings
# Avoid relying on "user-specific" settings in xcuserdata
```

#### Xcode + CLI Integration

You can even trigger CLI commands from within Xcode:

**Build Phase Script:**

In Xcode, add a "Run Script" build phase:

```bash
# SwiftLint as build phase
if which swiftlint >/dev/null; then
  swiftlint
else
  echo "warning: SwiftLint not installed"
fi
```

**Run Script on Build:**

```bash
# SwiftFormat before each build
if which swiftformat >/dev/null; then
  swiftformat --lint "${SRCROOT}"
else
  echo "warning: SwiftFormat not installed"
fi
```

### When to Use Xcode vs CLI vs Alternative IDEs

#### Use Xcode When:

✅ You need visual debugging with breakpoints and step-through
✅ Working with SwiftUI previews
✅ Using Interface Builder for UIKit/AppKit
✅ Profiling with Instruments
✅ Debugging memory graphs and leaks
✅ Exploring unfamiliar codebases (better code navigation)
✅ Adding/removing dependencies (SPM GUI is easier)
✅ Configuring build settings and targets
✅ Working with asset catalogs and resources

**Pros:**
- Best debugging experience for Swift/macOS
- Integrated profiling tools
- Visual interface builders
- SwiftUI live previews
- Excellent code navigation and autocomplete
- Built and maintained by Apple

**Cons:**
- Resource-intensive (RAM/CPU)
- Slower startup time
- Not great for remote development
- Harder to automate
- Can't easily use AI assistants that need CLI access

#### Use CLI + Simple Editor When:

✅ Building in CI/CD pipelines
✅ Using AI assistants (Claude CLI, GitHub Copilot CLI)
✅ Remote development via SSH
✅ Working on low-resource machines
✅ Scripting and automation
✅ Quick edits to single files
✅ When you need reproducible builds
✅ Working in terminal-heavy workflows

**Pros:**
- Fast and lightweight
- Fully automatable
- Works over SSH
- Great for AI-assisted development
- Reproducible builds
- Easy to script and integrate with other tools

**Cons:**
- No visual debugging (though lldb works)
- No SwiftUI previews
- More manual setup required
- Steeper learning curve for beginners
- Dependency management more difficult

#### Use VSCode When:

✅ You want a middle ground between Xcode and pure CLI
✅ Working on multiple languages/platforms
✅ You prefer a keyboard-centric workflow
✅ Need remote development features
✅ Want extensive customization and extensions
✅ Lighter resource usage than Xcode

**Pros:**
- Much lighter than Xcode
- Excellent extensions ecosystem
- Good SourceKit-LSP integration
- Great for polyglot projects
- Remote development support
- Free and cross-platform
- AI Copilot integration

**Cons:**
- SourceKit-LSP not as robust as Xcode
- No SwiftUI previews
- No Interface Builder
- No Instruments integration
- Dependency management still requires Xcode
- Debugging less polished than Xcode

#### Use Neovim/Vim When:

✅ You're already a Vim power user
✅ Terminal-centric development workflow
✅ Working via SSH frequently
✅ Maximum keyboard efficiency
✅ Minimal resource usage

**Pros:**
- Extremely fast and lightweight
- Ultimate keyboard control
- Works everywhere (including SSH)
- Highly customizable
- Free and open source

**Cons:**
- Steep learning curve
- LSP support requires configuration
- No GUI debugging tools
- No previews or visual tools
- Not beginner-friendly

### Recommended Setups by Role

#### Solo Developer / Indie Developer

```bash
# Primary: Xcode for development
# Secondary: CLI for automation and AI assistance
# Setup:
# - Use Xcode for daily coding and debugging
# - Add pre-commit hooks for formatting/linting
# - Use Claude CLI for code reviews and refactoring
# - Run tests via CLI before pushing
```

#### Team Lead / Senior Developer

```bash
# Primary: Hybrid (Xcode + VSCode)
# Secondary: CLI for everything
# Setup:
# - Xcode for complex debugging and profiling
# - VSCode for quick edits and cross-platform work
# - CLI scripts for all team automation
# - Enforce CLI-based CI/CD
# - Use CLI tools for code reviews
```

#### DevOps / Build Engineer

```bash
# Primary: Pure CLI
# Secondary: VSCode for occasional code inspection
# Setup:
# - Everything scripted and automated
# - XcodeGen for project generation
# - CLI-based builds, tests, and deployment
# - No Xcode dependency in CI/CD
```

#### AI-Assisted Development (Claude CLI)

```bash
# Primary: CLI + Simple Editor (VSCode or Vim)
# Secondary: Xcode for verification
# Setup:
# - Let AI handle most code changes via CLI
# - AI runs tests automatically
# - AI formats and lints all changes
# - Human reviews and debugs in Xcode when needed
# - Maximum automation, minimum manual intervention

# Example workflow:
# 1. Ask Claude to implement feature
# 2. Claude edits files, runs swiftformat
# 3. Claude runs swiftlint --fix
# 4. Claude runs build via xcodebuild
# 5. Claude runs tests and reports results
# 6. Human reviews in Xcode if needed
# 7. Claude commits and pushes changes
```

### The Bottom Line

**You don't have to choose!** The best approach is:

1. **Use Xcode** as your primary IDE for development
2. **Enable CLI automation** for everything that can be automated
3. **Make your project CLI-friendly** by sharing schemes and using standard tools
4. **Let AI assistants** use CLI while you use Xcode
5. **Keep options open** - CLI means anyone can build without Xcode

This guide focuses on CLI automation not to replace Xcode, but to **enable automation and AI-assisted development** while keeping Xcode as an option.

## Alternative Editors and IDE Setup

Beyond Xcode, you can develop Swift entirely from alternative editors. This is especially useful for AI-assisted development, remote work, or when you prefer a lighter-weight environment:

### Visual Studio Code

**Setup:**

```bash
# Install VSCode
brew install --cask visual-studio-code

# Launch from command line
code .
```

**Essential Extensions:**

Install these extensions for Swift development:

```bash
# Install via command line
code --install-extension sswg.swift-lang
code --install-extension vknabel.vscode-swiftformat
code --install-extension vknabel.vscode-swiftlint

# Optional but recommended
code --install-extension vadimcn.vscode-lldb         # Debugging
code --install-extension tamasfe.even-better-toml    # For config files
code --install-extension streetsidesoftware.code-spell-checker
```

**VSCode Settings for Swift:**

Create `.vscode/settings.json`:

```json
{
  "swift.path": "/usr/bin/swift",
  "swift.buildPath": "/usr/bin/swift",
  "sourcekit-lsp.serverPath": "/usr/bin/sourcekit-lsp",
  "swift.autoGenerateLaunchConfigurations": true,

  // SwiftFormat integration
  "[swift]": {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "vknabel.vscode-swiftformat"
  },

  // File associations
  "files.associations": {
    "*.swift": "swift"
  },

  // Exclude build artifacts
  "files.exclude": {
    "**/.build": true,
    "**/DerivedData": true,
    "**/.swiftpm": true
  }
}
```

**VSCode Tasks:**

Create `.vscode/tasks.json` for build tasks:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build",
      "type": "shell",
      "command": "xcodebuild",
      "args": [
        "-project", "${workspaceFolder}/*.xcodeproj",
        "-scheme", "YourScheme",
        "build"
      ],
      "group": {
        "kind": "build",
        "isDefault": true
      },
      "problemMatcher": []
    },
    {
      "label": "Test",
      "type": "shell",
      "command": "xcodebuild",
      "args": [
        "test",
        "-project", "${workspaceFolder}/*.xcodeproj",
        "-scheme", "YourScheme",
        "-destination", "platform=macOS"
      ],
      "group": "test"
    },
    {
      "label": "SwiftLint",
      "type": "shell",
      "command": "swiftlint",
      "args": ["--strict"],
      "problemMatcher": []
    },
    {
      "label": "SwiftFormat",
      "type": "shell",
      "command": "swiftformat",
      "args": ["."],
      "problemMatcher": []
    }
  ]
}
```

**VSCode Launch Configuration:**

Create `.vscode/launch.json` for debugging:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "lldb",
      "request": "launch",
      "name": "Debug",
      "program": "${workspaceFolder}/build/Build/Products/Debug/YourApp.app/Contents/MacOS/YourApp",
      "args": [],
      "cwd": "${workspaceFolder}",
      "preLaunchTask": "Build"
    }
  ]
}
```

### Neovim/Vim

**Setup:**

```bash
# Install Neovim
brew install neovim

# Install vim-plug (plugin manager)
sh -c 'curl -fLo "${XDG_DATA_HOME:-$HOME/.local/share}"/nvim/site/autoload/plug.vim --create-dirs \
       https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim'
```

**Recommended Plugins:**

Add to `~/.config/nvim/init.vim`:

```vim
call plug#begin()

" LSP Support
Plug 'neovim/nvim-lspconfig'
Plug 'williamboman/mason.nvim'
Plug 'williamboman/mason-lspconfig.nvim'

" Autocompletion
Plug 'hrsh7th/nvim-cmp'
Plug 'hrsh7th/cmp-nvim-lsp'

" Swift syntax
Plug 'keith/swift.vim'

" Code formatting
Plug 'sbdchd/neoformat'

" File navigation
Plug 'nvim-telescope/telescope.nvim'
Plug 'nvim-lua/plenary.nvim'

call plug#end()

" SourceKit-LSP configuration
lua << EOF
require'lspconfig'.sourcekit.setup{
  cmd = {'/usr/bin/sourcekit-lsp'},
  filetypes = {'swift'},
  root_dir = require'lspconfig'.util.root_pattern('Package.swift', '.git', '*.xcodeproj')
}
EOF

" Format Swift files on save
autocmd BufWritePre *.swift Neoformat swiftformat
```

### Sublime Text

```bash
# Install Sublime Text
brew install --cask sublime-text

# Install Package Control, then install:
# - LSP
# - LSP-SourceKit
# - Swift-Next (syntax highlighting)
```

### SourceKit-LSP Setup

SourceKit-LSP is Apple's Language Server Protocol implementation for Swift, providing IDE features:

```bash
# Verify sourcekit-lsp is installed
which sourcekit-lsp
# Should output: /usr/bin/sourcekit-lsp (comes with Xcode)

# Test sourcekit-lsp
xcrun sourcekit-lsp

# For Xcode projects, generate compile_commands.json (needed for LSP)
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -destination 'platform=macOS' \
  clean build | \
  xcpretty --report json-compilation-database --output compile_commands.json

# Or use xcbeautify
xcodebuild build -project YourApp.xcodeproj -scheme YourScheme 2>&1 | \
  xcbeautify --report json-compilation-database --report-path ./
```

### Editor-Independent Workflow

Create scripts that work regardless of editor:

```bash
cat > scripts/dev.sh << 'EOF'
#!/bin/bash
# Universal development script

case "$1" in
  format)
    echo "Formatting code..."
    swiftformat .
    ;;
  lint)
    echo "Linting code..."
    swiftlint --strict
    ;;
  build)
    echo "Building..."
    xcodebuild build -project *.xcodeproj -scheme MyScheme | xcbeautify
    ;;
  test)
    echo "Testing..."
    xcodebuild test -project *.xcodeproj -scheme MyScheme -destination 'platform=macOS' | xcbeautify
    ;;
  run)
    echo "Running..."
    ./build/Build/Products/Debug/MyApp.app/Contents/MacOS/MyApp
    ;;
  clean)
    echo "Cleaning..."
    rm -rf build .swiftpm ~/Library/Developer/Xcode/DerivedData
    ;;
  *)
    echo "Usage: $0 {format|lint|build|test|run|clean}"
    exit 1
    ;;
esac
EOF
chmod +x scripts/dev.sh

# Usage from any editor:
./scripts/dev.sh format
./scripts/dev.sh build
./scripts/dev.sh test
```

### Recommended Tools for All Editors

**See the [Complete Installation Guide](#complete-installation-guide) in the Essential Tools section for full details on all recommended packages.**

Quick install command for all tools:

```bash
brew install \
  swiftlint swiftformat xcbeautify \
  peripheryapp/periphery/periphery \
  jq bat fd ripgrep fzf fswatch \
  git-delta lazygit gh
```

### File Watching and Auto-Rebuild

For automatic rebuilds on file changes:

```bash
# Install fswatch
brew install fswatch

# Watch for changes and rebuild
cat > scripts/watch.sh << 'EOF'
#!/bin/bash
echo "Watching for changes..."
fswatch -o . \
  --exclude ".*" \
  --include "\\.swift$" \
  --exclude ".build" \
  --exclude "DerivedData" \
  | while read; do
    echo "Change detected, rebuilding..."
    xcodebuild build -project *.xcodeproj -scheme MyScheme 2>&1 | xcbeautify
  done
EOF
chmod +x scripts/watch.sh
```

## Common Workflows

### Complete Build & Test Pipeline

```bash
#!/bin/bash
set -e

PROJECT="YourApp.xcodeproj"
SCHEME="YourScheme"
DESTINATION="platform=macOS"

echo "🧹 Cleaning..."
xcodebuild clean \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  > /dev/null

echo "📦 Resolving dependencies..."
xcodebuild -resolvePackageDependencies \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  > /dev/null

echo "✨ Formatting code..."
swiftformat .

echo "🔍 Linting..."
swiftlint --strict

echo "🔨 Building..."
xcodebuild build \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  -destination "$DESTINATION" | xcpretty

echo "🧪 Testing..."
xcodebuild test \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  -destination "$DESTINATION" \
  -enableCodeCoverage YES | xcpretty

echo "✅ All checks passed!"
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

# Format code
swiftformat --lint . || {
  echo "❌ SwiftFormat failed. Run 'swiftformat .' to fix."
  exit 1
}

# Lint code
swiftlint --strict || {
  echo "❌ SwiftLint failed. Fix issues before committing."
  exit 1
}

echo "✅ Pre-commit checks passed"
```

### Fast Iteration During Development

```bash
# Quick build without cleaning
xcodebuild build \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -quiet | xcpretty --simple

# Quick test of specific file
xcodebuild test \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -only-testing:YourAppTests/SpecificTest \
  -quiet
```

## Pitfalls and Solutions

### 1. Code Signing Issues

**Problem**: Build fails with code signing errors.

**Solutions**:
```bash
# Disable code signing for local builds
xcodebuild build \
  CODE_SIGN_IDENTITY="" \
  CODE_SIGNING_REQUIRED=NO \
  CODE_SIGNING_ALLOWED=NO

# Or use ad-hoc signing
xcodebuild build \
  CODE_SIGN_IDENTITY="-"

# List available identities
security find-identity -v -p codesigning
```

### 2. Scheme Not Found

**Problem**: `xcodebuild: error: The project does not contain a scheme named 'X'`

**Solution**:
- Schemes must be marked as "Shared" to be visible to command-line tools
- In Xcode: Product → Scheme → Manage Schemes → Check "Shared"
- Or manually edit `.xcodeproj/xcshareddata/xcschemes/`

```bash
# Verify scheme exists and is shared
ls -la YourApp.xcodeproj/xcshareddata/xcschemes/
```

### 3. Derived Data Issues

**Problem**: Stale build artifacts causing mysterious errors.

**Solution**:
```bash
# Clean derived data
rm -rf ~/Library/Developer/Xcode/DerivedData

# Or use specific derived data path
xcodebuild build \
  -derivedDataPath ./build/DerivedData

# Clean and build
xcodebuild clean build \
  -project YourApp.xcodeproj \
  -scheme YourScheme
```

### 4. Simulator Not Available

**Problem**: Tests fail because simulator isn't available.

**Solution**:
```bash
# List available simulators
xcrun simctl list devices

# Boot a simulator
xcrun simctl boot "iPhone 15"

# For macOS, use:
-destination 'platform=macOS'
```

### 5. SwiftLint/SwiftFormat Version Mismatches

**Problem**: Different developers or CI have different tool versions.

**Solution**:
```bash
# Use Mint for version management
brew install mint

# Create Mintfile
cat > Mintfile << 'EOF'
realm/SwiftLint@0.54.0
nicklockwood/SwiftFormat@0.52.11
EOF

# Install versions
mint bootstrap

# Run with specific version
mint run swiftlint
mint run swiftformat .
```

### 6. Build Timeout on CI

**Problem**: Builds time out on CI systems.

**Solution**:
```bash
# Increase timeout
-maximum-concurrent-test-device-destinations 1
-maximum-concurrent-test-simulator-destinations 1

# Disable parallelization
-parallel-testing-enabled NO

# Use quiet mode to reduce output
-quiet
```

### 7. Path and Workspace Issues

**Problem**: Build fails with "not a valid path".

**Solution**:
```bash
# Always use absolute paths or be explicit about relative paths
PROJECT_DIR="$(pwd)"
xcodebuild -project "$PROJECT_DIR/YourApp.xcodeproj"

# For workspaces, use -workspace instead of -project
xcodebuild -workspace YourApp.xcworkspace -scheme YourScheme
```

### 8. Swift Package Resolution Failures

**Problem**: Package resolution hangs or fails.

**Solution**:
```bash
# Clear package cache
rm -rf ~/Library/Caches/org.swift.swiftpm
rm -rf .swiftpm
rm Package.resolved

# Reset package caches in Xcode DerivedData
rm -rf ~/Library/Developer/Xcode/DerivedData/*/SourcePackages

# Resolve with verbose output
xcodebuild -resolvePackageDependencies \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -verbose
```

## Claude CLI Integration

### Setup for Claude CLI

Create hooks in Claude Code settings for automated workflows.

#### Pre-commit Hook Example

In your Claude Code settings, add a hook that runs before commits:

```json
{
  "hooks": {
    "pre-commit": {
      "command": "swiftformat . && swiftlint --strict",
      "description": "Format and lint Swift code"
    }
  }
}
```

#### Test Hook

```json
{
  "hooks": {
    "pre-push": {
      "command": "./scripts/run-tests.sh",
      "description": "Run all tests before push"
    }
  }
}
```

### Common Claude CLI Workflows

#### 1. Automated Code Review

When Claude modifies Swift files:

```bash
# Claude should run after making changes
swiftformat [modified-files]
swiftlint --path [modified-files]
xcodebuild build -quiet
```

#### 2. Test-Driven Development

```bash
# 1. Claude writes test
# 2. Run test to verify it fails
xcodebuild test -only-testing:TestClass/testNewFeature

# 3. Claude implements feature
# 4. Run test to verify it passes
xcodebuild test -only-testing:TestClass/testNewFeature

# 5. Run all tests
xcodebuild test
```

#### 3. Refactoring Workflow

```bash
# 1. Run tests before changes
xcodebuild test > /tmp/before.log

# 2. Claude makes refactoring changes

# 3. Format and lint
swiftformat .
swiftlint --fix

# 4. Verify tests still pass
xcodebuild test > /tmp/after.log

# 5. Compare results
diff /tmp/before.log /tmp/after.log
```

### Scripts for Claude CLI

Create helper scripts that Claude can easily invoke:

#### `scripts/validate.sh`
```bash
#!/bin/bash
set -e

echo "Formatting..."
swiftformat .

echo "Linting..."
swiftlint --strict

echo "Building..."
xcodebuild build \
  -project *.xcodeproj \
  -scheme $(xcodebuild -list | grep -A 1 "Schemes:" | tail -1 | xargs) \
  -quiet | xcpretty --simple

echo "Testing..."
xcodebuild test \
  -project *.xcodeproj \
  -scheme $(xcodebuild -list | grep -A 1 "Schemes:" | tail -1 | xargs) \
  -quiet | xcpretty --simple

echo "✅ All validations passed"
```

#### `scripts/quick-test.sh`
```bash
#!/bin/bash
# Run tests for a specific file

if [ -z "$1" ]; then
  echo "Usage: $0 <TestClassName>"
  exit 1
fi

xcodebuild test \
  -project *.xcodeproj \
  -scheme $(xcodebuild -list | grep -A 1 "Schemes:" | tail -1 | xargs) \
  -only-testing:"${1}" \
  | xcpretty --simple
```

## CI/CD Integration

### GitHub Actions

`.github/workflows/swift.yml`:

```yaml
name: Swift CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build-and-test:
    runs-on: macos-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup Xcode
      uses: maxim-lobanov/setup-xcode@v1
      with:
        xcode-version: latest-stable

    - name: Install dependencies
      run: |
        brew install swiftlint swiftformat xcpretty

    - name: SwiftLint
      run: swiftlint --strict

    - name: SwiftFormat Check
      run: swiftformat --lint .

    - name: Build
      run: |
        xcodebuild build \
          -project *.xcodeproj \
          -scheme MyScheme \
          -destination 'platform=macOS' \
          | xcpretty

    - name: Test
      run: |
        xcodebuild test \
          -project *.xcodeproj \
          -scheme MyScheme \
          -destination 'platform=macOS' \
          -enableCodeCoverage YES \
          | xcpretty

    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.lcov
        fail_ci_if_error: true
```

### GitLab CI

`.gitlab-ci.yml`:

```yaml
image: macos-12-xcode-14

stages:
  - lint
  - build
  - test

variables:
  SCHEME: "MyScheme"
  PROJECT: "MyApp.xcodeproj"

before_script:
  - brew install swiftlint swiftformat xcpretty

lint:
  stage: lint
  script:
    - swiftformat --lint .
    - swiftlint --strict

build:
  stage: build
  script:
    - xcodebuild build -project $PROJECT -scheme $SCHEME | xcpretty
  artifacts:
    paths:
      - build/

test:
  stage: test
  script:
    - xcodebuild test -project $PROJECT -scheme $SCHEME -enableCodeCoverage YES | xcpretty
  coverage: '/Test Coverage: \d+\.\d+%/'
```

### Jenkins Pipeline

```groovy
pipeline {
    agent { label 'macos' }

    environment {
        PROJECT = 'MyApp.xcodeproj'
        SCHEME = 'MyScheme'
    }

    stages {
        stage('Setup') {
            steps {
                sh 'brew install swiftlint swiftformat xcpretty'
            }
        }

        stage('Lint') {
            steps {
                sh 'swiftformat --lint .'
                sh 'swiftlint --strict'
            }
        }

        stage('Build') {
            steps {
                sh """
                    xcodebuild build \
                        -project ${PROJECT} \
                        -scheme ${SCHEME} \
                        | xcpretty
                """
            }
        }

        stage('Test') {
            steps {
                sh """
                    xcodebuild test \
                        -project ${PROJECT} \
                        -scheme ${SCHEME} \
                        -enableCodeCoverage YES \
                        | xcpretty
                """
            }
        }
    }

    post {
        always {
            junit '**/test-results/**/*.xml'
        }
    }
}
```

## Advanced Tips

### 1. Parallel Testing

```bash
# Enable parallel testing
xcodebuild test \
  -parallel-testing-enabled YES \
  -maximum-parallel-testing-workers 4
```

### 2. Build Performance

```bash
# Use build cache
-clonedSourcePackagesDirPath ./SourcePackages

# Disable index-while-building
-enableIndexBuildArena NO

# Use new build system
-UseModernBuildSystem=YES
```

### 3. Custom Build Scripts

Integrate custom scripts in build phases:

```bash
# Add run script to .xcodeproj
# This requires editing project.pbxproj or using Xcode GUI once
# The script can run SwiftLint, SwiftFormat, or custom validations
```

### 4. Environment-Specific Builds

```bash
# Use xcconfig files
xcodebuild build \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -xcconfig Configs/Debug.xcconfig
```

### 5. Automatic Screenshot Generation

```bash
# Use fastlane snapshot for automated screenshots
brew install fastlane
fastlane snapshot
```

## Troubleshooting Commands

```bash
# Check Swift version
swift --version

# Check Xcode version
xcodebuild -version

# Check available SDKs
xcodebuild -showsdks

# List all build settings
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -showBuildSettings

# Verbose build output
xcodebuild build -verbose

# Check if scheme is shared
find . -name "*.xcscheme" -path "*/xcshareddata/*"
```

## Continuous Delivery and Release Automation

Automate your release process from version bumping to distribution. This section covers tools and workflows for managing releases, changelogs, and deployment entirely from the command line.

### Automatic Version Bumping

#### Using agvtool (Apple's Version Management Tool)

```bash
# Enable agvtool by ensuring your Info.plist has these keys:
# CFBundleVersion (build number)
# CFBundleShortVersionString (marketing version)

# Set the marketing version (e.g., 1.2.3)
agvtool new-marketing-version 1.2.3

# Bump the build number
agvtool next-version -all

# Set a specific build number
agvtool new-version -all 42

# Check current version
agvtool what-version
agvtool what-marketing-version
```

#### Using PlistBuddy for More Control

```bash
# Get current version
CURRENT_VERSION=$(/usr/libexec/PlistBuddy -c "Print CFBundleShortVersionString" Info.plist)
CURRENT_BUILD=$(/usr/libexec/PlistBuddy -c "Print CFBundleVersion" Info.plist)

# Set new version
/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString 1.2.3" Info.plist
/usr/libexec/PlistBuddy -c "Set :CFBundleVersion 100" Info.plist

# For SPM projects, update Package.swift version comment
sed -i '' 's/\/\/ Version: .*/\/\/ Version: 1.2.3/' Package.swift
```

#### Semantic Version Bumping Script

```bash
cat > scripts/bump-version.sh << 'EOF'
#!/bin/bash
set -e

PLIST="${1:-Info.plist}"
BUMP_TYPE="${2:-patch}"  # major, minor, or patch

if [ ! -f "$PLIST" ]; then
    echo "Error: $PLIST not found"
    exit 1
fi

# Get current version
CURRENT=$(/usr/libexec/PlistBuddy -c "Print CFBundleShortVersionString" "$PLIST")
echo "Current version: $CURRENT"

# Parse version
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

# Bump version based on type
case "$BUMP_TYPE" in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "Invalid bump type: $BUMP_TYPE (use major, minor, or patch)"
        exit 1
        ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo "New version: $NEW_VERSION"

# Update plist
/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $NEW_VERSION" "$PLIST"

# Bump build number
BUILD=$(/usr/libexec/PlistBuddy -c "Print CFBundleVersion" "$PLIST")
NEW_BUILD=$((BUILD + 1))
/usr/libexec/PlistBuddy -c "Set :CFBundleVersion $NEW_BUILD" "$PLIST"

echo "Updated to version $NEW_VERSION (build $NEW_BUILD)"
echo "NEW_VERSION=$NEW_VERSION" >> $GITHUB_OUTPUT  # For GitHub Actions
EOF
chmod +x scripts/bump-version.sh

# Usage:
# ./scripts/bump-version.sh Info.plist patch  # 1.2.3 -> 1.2.4
# ./scripts/bump-version.sh Info.plist minor  # 1.2.3 -> 1.3.0
# ./scripts/bump-version.sh Info.plist major  # 1.2.3 -> 2.0.0
```

### Changelog Generation

#### Using Conventional Commits

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for automatic changelog generation:

```bash
# Commit message format:
# <type>(<scope>): <description>
#
# Types: feat, fix, docs, style, refactor, perf, test, chore
# Examples:
git commit -m "feat(auth): add biometric authentication"
git commit -m "fix(clipboard): resolve memory leak in observer"
git commit -m "docs: update installation instructions"
```

#### Install Changelog Generators

```bash
# Option 1: git-cliff (Rust-based, fast, highly customizable)
brew install git-cliff

# Option 2: github-changelog-generator (Ruby-based)
brew install github-changelog-generator

# Option 3: standard-version (Node.js-based)
npm install -g standard-version
```

#### Using git-cliff (Recommended)

```bash
# Generate changelog
git cliff --output CHANGELOG.md

# Generate changelog for unreleased commits
git cliff --unreleased --tag v1.2.3

# Generate changelog between tags
git cliff v1.0.0..v1.2.0 --output CHANGELOG-1.2.0.md

# Custom configuration (.cliff.toml)
cat > .cliff.toml << 'EOF'
[changelog]
header = """
# Changelog
All notable changes to this project will be documented in this file.\n
"""
body = """
{% for group, commits in commits | group_by(attribute="group") %}
    ### {{ group | upper_first }}
    {% for commit in commits %}
        - {{ commit.message | upper_first }} ({{ commit.id | truncate(length=7, end="") }})
    {% endfor %}
{% endfor %}
"""

[git]
conventional_commits = true
filter_unconventional = true
commit_parsers = [
    { message = "^feat", group = "Features"},
    { message = "^fix", group = "Bug Fixes"},
    { message = "^doc", group = "Documentation"},
    { message = "^perf", group = "Performance"},
    { message = "^refactor", group = "Refactoring"},
    { message = "^style", group = "Styling"},
    { message = "^test", group = "Testing"},
    { message = "^chore", group = "Miscellaneous"},
]
EOF

git cliff --config .cliff.toml --output CHANGELOG.md
```

#### Using github-changelog-generator

```bash
# Generate changelog (requires GitHub API token)
export CHANGELOG_GITHUB_TOKEN="your_github_token"

github_changelog_generator \
  --user your-username \
  --project your-project \
  --token $CHANGELOG_GITHUB_TOKEN \
  --output CHANGELOG.md

# With options
github_changelog_generator \
  --user your-username \
  --project your-project \
  --exclude-labels duplicate,question,invalid,wontfix \
  --enhancement-label "**Enhancements:**" \
  --bugs-label "**Bug Fixes:**" \
  --since-tag v1.0.0
```

### Git Tagging and Release Preparation

#### Automated Tagging Script

```bash
cat > scripts/create-release.sh << 'EOF'
#!/bin/bash
set -e

VERSION="$1"
MESSAGE="${2:-Release version $VERSION}"

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version> [message]"
    echo "Example: $0 1.2.3 'Release with new features'"
    exit 1
fi

# Validate version format (semantic versioning)
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format X.Y.Z (e.g., 1.2.3)"
    exit 1
fi

echo "Creating release for version $VERSION"

# Update version in project files
if [ -f "Info.plist" ]; then
    echo "Updating Info.plist..."
    /usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $VERSION" Info.plist
fi

# Generate changelog for this version
if command -v git-cliff &> /dev/null; then
    echo "Generating changelog..."
    git cliff --unreleased --tag "v$VERSION" --prepend CHANGELOG.md
fi

# Commit version bump and changelog
git add -A
git commit -m "chore: bump version to $VERSION" || echo "No changes to commit"

# Create annotated tag
git tag -a "v$VERSION" -m "$MESSAGE"

echo "✅ Release v$VERSION created"
echo ""
echo "Next steps:"
echo "  1. Review the changes: git show v$VERSION"
echo "  2. Push to remote: git push origin main --tags"
echo "  3. Create GitHub release: gh release create v$VERSION"
EOF
chmod +x scripts/create-release.sh

# Usage:
# ./scripts/create-release.sh 1.2.3
# ./scripts/create-release.sh 1.2.3 "Major release with breaking changes"
```

### GitHub Releases from CLI

#### Using GitHub CLI (gh)

```bash
# Install gh if not already installed
brew install gh

# Authenticate
gh auth login

# Create a release with auto-generated notes
gh release create v1.2.3 \
  --title "Version 1.2.3" \
  --generate-notes

# Create a release with custom notes
gh release create v1.2.3 \
  --title "Version 1.2.3" \
  --notes "## What's New
- Feature 1
- Feature 2
- Bug fix 3"

# Create a release with notes from file
gh release create v1.2.3 \
  --title "Version 1.2.3" \
  --notes-file CHANGELOG.md

# Upload release assets (binaries, DMGs, etc.)
gh release create v1.2.3 \
  --title "Version 1.2.3" \
  --notes-file CHANGELOG.md \
  ./build/YourApp.dmg \
  ./build/YourApp.zip

# Create a pre-release (beta)
gh release create v1.2.3-beta.1 \
  --title "Version 1.2.3 Beta 1" \
  --prerelease \
  --notes "Beta release for testing"

# List releases
gh release list

# View a specific release
gh release view v1.2.3

# Delete a release
gh release delete v1.2.3 --yes
```

#### Complete Release Script with GitHub Integration

```bash
cat > scripts/publish-release.sh << 'EOF'
#!/bin/bash
set -e

VERSION="$1"
PRERELEASE="${2:-false}"

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version> [prerelease]"
    echo "Example: $0 1.2.3"
    echo "Example: $0 1.2.3-beta.1 true"
    exit 1
fi

echo "🚀 Publishing release v$VERSION"

# Ensure we're on main branch and up to date
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "⚠️  Warning: Not on main branch (currently on $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build release version
echo "📦 Building release..."
xcodebuild archive \
  -project *.xcodeproj \
  -scheme "$(xcodebuild -list | grep -A 1 "Schemes:" | tail -1 | xargs)" \
  -configuration Release \
  -archivePath ./build/Release.xcarchive

# Export app
echo "📤 Exporting app..."
xcodebuild -exportArchive \
  -archivePath ./build/Release.xcarchive \
  -exportPath ./build \
  -exportOptionsPlist ExportOptions.plist

# Create DMG (if create-dmg is installed)
if command -v create-dmg &> /dev/null; then
    echo "💿 Creating DMG..."
    APP_NAME=$(find ./build -name "*.app" -maxdepth 1 | head -1 | xargs basename)
    create-dmg \
      --volname "$APP_NAME $VERSION" \
      --window-pos 200 120 \
      --window-size 600 400 \
      --icon-size 100 \
      --app-drop-link 450 120 \
      "./build/${APP_NAME%.app}-${VERSION}.dmg" \
      "./build/$APP_NAME"
fi

# Generate changelog snippet for this release
echo "📝 Generating changelog..."
if command -v git-cliff &> /dev/null; then
    PREV_TAG=$(git describe --abbrev=0 --tags 2>/dev/null || echo "")
    if [ -n "$PREV_TAG" ]; then
        git cliff $PREV_TAG..HEAD --strip header > /tmp/release-notes.md
    else
        git cliff --unreleased --strip header > /tmp/release-notes.md
    fi
else
    # Fallback: use git log
    git log $(git describe --abbrev=0 --tags 2>/dev/null || echo "")..HEAD \
      --pretty=format:"- %s (%h)" > /tmp/release-notes.md
fi

# Create GitHub release
echo "🎉 Creating GitHub release..."
RELEASE_ARGS=(
    "v$VERSION"
    --title "Version $VERSION"
    --notes-file /tmp/release-notes.md
)

if [ "$PRERELEASE" = "true" ]; then
    RELEASE_ARGS+=(--prerelease)
fi

# Add DMG if it exists
if [ -f "./build/"*".dmg" ]; then
    RELEASE_ARGS+=("./build/"*".dmg")
fi

gh release create "${RELEASE_ARGS[@]}"

echo "✅ Release v$VERSION published successfully!"
echo "View at: $(gh release view v$VERSION --json url -q .url)"

# Clean up
rm -f /tmp/release-notes.md
EOF
chmod +x scripts/publish-release.sh

# Usage:
# ./scripts/publish-release.sh 1.2.3           # Regular release
# ./scripts/publish-release.sh 1.2.3-beta.1 true  # Pre-release
```

### Automated TestFlight Distribution

For iOS apps, automate TestFlight uploads using `altool` or `notarytool`:

```bash
cat > scripts/upload-testflight.sh << 'EOF'
#!/bin/bash
set -e

IPA_PATH="$1"
APPLE_ID="${APPLE_ID:-your@email.com}"
TEAM_ID="${TEAM_ID:-YOUR_TEAM_ID}"
APP_PASSWORD="${APP_PASSWORD}"  # App-specific password

if [ -z "$IPA_PATH" ]; then
    echo "Usage: $0 <ipa-path>"
    exit 1
fi

if [ ! -f "$IPA_PATH" ]; then
    echo "Error: IPA file not found: $IPA_PATH"
    exit 1
fi

echo "📱 Uploading to TestFlight..."

# Upload to App Store Connect
xcrun altool --upload-app \
  --type ios \
  --file "$IPA_PATH" \
  --username "$APPLE_ID" \
  --password "$APP_PASSWORD" \
  --verbose

echo "✅ Upload complete!"
echo "The build will be available in TestFlight after processing (usually 5-15 minutes)"
EOF
chmod +x scripts/upload-testflight.sh

# Set environment variables (add to ~/.zshrc or use .env file)
export APPLE_ID="your@email.com"
export TEAM_ID="YOUR_TEAM_ID"
export APP_PASSWORD="xxxx-xxxx-xxxx-xxxx"  # Generate at appleid.apple.com

# Usage:
# ./scripts/upload-testflight.sh ./build/YourApp.ipa
```

### CI/CD Integration for Automated Releases

#### GitHub Actions Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  release:
    runs-on: macos-latest

    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Needed for changelog generation

      - name: Setup Xcode
        uses: maxim-lobanov/setup-xcode@v1
        with:
          xcode-version: latest-stable

      - name: Install dependencies
        run: |
          brew install git-cliff create-dmg

      - name: Get version from tag
        id: version
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          echo "version=$VERSION" >> $GITHUB_OUTPUT

      - name: Build release
        run: |
          xcodebuild archive \
            -project *.xcodeproj \
            -scheme "YourScheme" \
            -configuration Release \
            -archivePath ./build/Release.xcarchive

          xcodebuild -exportArchive \
            -archivePath ./build/Release.xcarchive \
            -exportPath ./build \
            -exportOptionsPlist ExportOptions.plist

      - name: Create DMG
        run: |
          APP_NAME=$(find ./build -name "*.app" -maxdepth 1 | head -1)
          create-dmg \
            --volname "YourApp ${{ steps.version.outputs.version }}" \
            --window-pos 200 120 \
            --window-size 600 400 \
            --icon-size 100 \
            --app-drop-link 450 120 \
            "./build/YourApp-${{ steps.version.outputs.version }}.dmg" \
            "$APP_NAME"

      - name: Generate changelog
        run: |
          git cliff --latest --strip header > /tmp/release-notes.md

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v1
        with:
          body_path: /tmp/release-notes.md
          files: |
            ./build/*.dmg
            ./build/*.zip
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Upload to TestFlight (if iOS)
        if: contains(github.ref, 'ios')
        env:
          APPLE_ID: ${{ secrets.APPLE_ID }}
          APP_PASSWORD: ${{ secrets.APP_PASSWORD }}
        run: |
          xcrun altool --upload-app \
            --type ios \
            --file ./build/*.ipa \
            --username "$APPLE_ID" \
            --password "$APP_PASSWORD"
```

#### Complete Release Workflow

```bash
cat > scripts/complete-release-workflow.sh << 'EOF'
#!/bin/bash
set -e

VERSION="$1"
BUMP_TYPE="${2:-patch}"  # major, minor, patch

if [ -z "$VERSION" ] && [ -z "$BUMP_TYPE" ]; then
    echo "Usage: $0 <version> OR set BUMP_TYPE to auto-calculate"
    echo "Example: $0 1.2.3"
    echo "Example: BUMP_TYPE=minor $0"
    exit 1
fi

echo "🚀 Starting complete release workflow"

# Step 1: Ensure clean working directory
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ Working directory is not clean. Commit or stash changes first."
    exit 1
fi

# Step 2: Ensure on main branch and up to date
git checkout main
git pull origin main

# Step 3: Run tests
echo "🧪 Running tests..."
xcodebuild test \
  -project *.xcodeproj \
  -scheme "$(xcodebuild -list | grep -A 1 "Schemes:" | tail -1 | xargs)" \
  -destination 'platform=macOS' \
  | xcbeautify

# Step 4: Run linting
echo "🔍 Running SwiftLint..."
swiftlint

# Step 5: Bump version if not provided
if [ -z "$VERSION" ]; then
    echo "📈 Bumping version ($BUMP_TYPE)..."
    ./scripts/bump-version.sh Info.plist "$BUMP_TYPE"
    VERSION=$(/usr/libexec/PlistBuddy -c "Print CFBundleShortVersionString" Info.plist)
fi

# Step 6: Update changelog
echo "📝 Updating changelog..."
if command -v git-cliff &> /dev/null; then
    git cliff --unreleased --tag "v$VERSION" --prepend CHANGELOG.md
fi

# Step 7: Commit version bump
git add -A
git commit -m "chore: release version $VERSION"

# Step 8: Create tag
git tag -a "v$VERSION" -m "Release version $VERSION"

# Step 9: Build and archive
echo "🔨 Building release..."
./scripts/publish-release.sh "$VERSION"

# Step 10: Push to remote
echo "⬆️  Pushing to remote..."
git push origin main --tags

echo "✅ Release workflow complete!"
echo "   Version: $VERSION"
echo "   GitHub Release: https://github.com/$(gh repo view --json nameWithOwner -q .nameWithOwner)/releases/tag/v$VERSION"
EOF
chmod +x scripts/complete-release-workflow.sh

# Usage:
# ./scripts/complete-release-workflow.sh 1.2.3
# BUMP_TYPE=minor ./scripts/complete-release-workflow.sh
```

### Best Practices for Release Automation

1. **Semantic Versioning**: Always use semantic versioning (MAJOR.MINOR.PATCH)
   - MAJOR: Breaking changes
   - MINOR: New features (backwards compatible)
   - PATCH: Bug fixes

2. **Conventional Commits**: Use conventional commit format for automatic changelog generation

3. **Pre-release Testing**:
   ```bash
   # Run full test suite before releasing
   xcodebuild test -scheme YourScheme -destination 'platform=macOS'
   swiftlint
   swiftformat --lint .
   ```

4. **Tag Naming**: Use consistent tag format (e.g., `v1.2.3`, not `1.2.3` or `version-1.2.3`)

5. **Changelog Maintenance**: Keep CHANGELOG.md updated and follow [Keep a Changelog](https://keepachangelog.com/) format

6. **Release Notes**: Always include meaningful release notes, not just commit messages

7. **Automation Safety**:
   ```bash
   # Add confirmation prompts to release scripts
   read -p "Ready to release v$VERSION? (y/n) " -n 1 -r
   echo
   if [[ ! $REPLY =~ ^[Yy]$ ]]; then
       exit 1
   fi
   ```

8. **Backup Before Release**:
   ```bash
   # Create a backup branch before major releases
   git branch backup/pre-release-$(date +%Y%m%d)
   ```

### Troubleshooting Release Issues

#### Version Number Conflicts

```bash
# If agvtool fails, check project settings
grep -r "CURRENT_PROJECT_VERSION" *.xcodeproj/project.pbxproj
grep -r "MARKETING_VERSION" *.xcodeproj/project.pbxproj

# Reset version tracking
agvtool new-version -all 1
agvtool new-marketing-version 0.1.0
```

#### Tag Already Exists

```bash
# Delete local tag
git tag -d v1.2.3

# Delete remote tag
git push origin :refs/tags/v1.2.3

# Or force push new tag
git tag -f -a v1.2.3 -m "Fixed release"
git push -f origin v1.2.3
```

#### GitHub Release Upload Failures

```bash
# Check gh auth status
gh auth status

# Re-authenticate
gh auth login

# Check file size limits (2GB for release assets)
ls -lh ./build/*.dmg

# Upload assets separately if needed
gh release upload v1.2.3 ./build/YourApp.dmg
```

## iOS Development from CLI

While this guide focuses on macOS, many developers work across platforms. Here's how iOS development differs from the command line.

### Key Differences from macOS

```bash
# macOS destination
-destination 'platform=macOS'

# iOS simulator destination
-destination 'platform=iOS Simulator,name=iPhone 15,OS=17.0'

# iOS device destination
-destination 'platform=iOS,id=<device-udid>'
```

### Simulator Management

```bash
# List all available simulators
xcrun simctl list devices

# List available device types
xcrun simctl list devicetypes

# List available runtimes (iOS versions)
xcrun simctl list runtimes

# Create a new simulator
xcrun simctl create "My iPhone 15" "iPhone 15" "iOS17.0"

# Boot a simulator
xcrun simctl boot "iPhone 15"

# Or boot by UDID
xcrun simctl boot <simulator-udid>

# Shutdown a simulator
xcrun simctl shutdown "iPhone 15"

# Delete a simulator
xcrun simctl delete "iPhone 15"

# Erase simulator data
xcrun simctl erase "iPhone 15"

# Install app on simulator
xcrun simctl install booted /path/to/YourApp.app

# Launch app on simulator
xcrun simctl launch booted com.yourcompany.YourApp

# Open Simulator app
open -a Simulator
```

### Building for iOS

```bash
# Build for iOS Simulator
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -sdk iphonesimulator \
  -destination 'platform=iOS Simulator,name=iPhone 15' \
  build

# Build for iOS Device
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -sdk iphoneos \
  -destination 'generic/platform=iOS' \
  build

# Build universal app (multiple architectures)
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -sdk iphoneos \
  -configuration Release \
  ONLY_ACTIVE_ARCH=NO \
  build
```

### Testing on iOS

```bash
# Run tests on iOS Simulator
xcodebuild test \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -destination 'platform=iOS Simulator,name=iPhone 15' \
  | xcbeautify

# Run tests on specific iOS version
xcodebuild test \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -destination 'platform=iOS Simulator,name=iPhone 15,OS=17.0' \
  | xcbeautify

# Run tests on multiple simulators in parallel
xcodebuild test \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -destination 'platform=iOS Simulator,name=iPhone 15' \
  -destination 'platform=iOS Simulator,name=iPhone SE (3rd generation)' \
  -destination 'platform=iOS Simulator,name=iPad Pro (12.9-inch)' \
  -parallel-testing-enabled YES \
  | xcbeautify
```

### Device Testing

```bash
# List connected devices
xcrun xctrace list devices

# Or use instruments
instruments -s devices

# Get device UDID
system_profiler SPUSBDataType | grep "Serial Number"

# Build and install on device
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  -destination 'platform=iOS,id=<device-udid>' \
  build

# Install IPA on device (requires ios-deploy or similar)
brew install ios-deploy
ios-deploy --bundle /path/to/YourApp.app --id <device-udid>
```

### Multi-Platform Projects

```bash
# Build for both macOS and iOS
#!/bin/bash
set -e

PROJECT="YourApp.xcodeproj"
SCHEME="YourApp"

echo "Building macOS..."
xcodebuild build \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  -destination 'platform=macOS' \
  | xcbeautify

echo "Building iOS Simulator..."
xcodebuild build \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  -destination 'platform=iOS Simulator,name=iPhone 15' \
  | xcbeautify

echo "Building iOS Device..."
xcodebuild build \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  -sdk iphoneos \
  -destination 'generic/platform=iOS' \
  | xcbeautify

echo "✅ All platforms built successfully!"
```

## Distribution and Notarization

For shipping production macOS apps outside the App Store.

### Code Signing for Distribution

```bash
# List available signing identities
security find-identity -v -p codesigning

# Build with specific identity
xcodebuild archive \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -archivePath ./build/YourApp.xcarchive \
  CODE_SIGN_IDENTITY="Developer ID Application: Your Name (TEAM_ID)"

# Verify code signature
codesign -vvv --deep --strict /path/to/YourApp.app

# Display signing information
codesign -d --verbose=4 /path/to/YourApp.app
```

### Creating Export Options Plist

Create `ExportOptions.plist` for automated exports:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>method</key>
    <string>developer-id</string>
    <key>teamID</key>
    <string>YOUR_TEAM_ID</string>
    <key>signingStyle</key>
    <string>automatic</string>
    <key>provisioningProfiles</key>
    <dict>
        <key>com.yourcompany.yourapp</key>
        <string>Your Provisioning Profile Name</string>
    </dict>
</dict>
</plist>
```

### Notarization Workflow

```bash
# 1. Archive the app
xcodebuild archive \
  -project YourApp.xcodeproj \
  -scheme YourScheme \
  -archivePath ./build/YourApp.xcarchive

# 2. Export the app
xcodebuild -exportArchive \
  -archivePath ./build/YourApp.xcarchive \
  -exportPath ./build \
  -exportOptionsPlist ExportOptions.plist

# 3. Create a ZIP for notarization
ditto -c -k --keepParent ./build/YourApp.app ./build/YourApp.zip

# 4. Submit for notarization (requires App Store Connect API key or credentials)
xcrun notarytool submit ./build/YourApp.zip \
  --apple-id "your@email.com" \
  --team-id "YOUR_TEAM_ID" \
  --password "app-specific-password" \
  --wait

# Or using API key (recommended for CI/CD)
xcrun notarytool submit ./build/YourApp.zip \
  --key ~/private_keys/AuthKey_KEYID.p8 \
  --key-id "KEY_ID" \
  --issuer "ISSUER_ID" \
  --wait

# 5. Check notarization status
xcrun notarytool info <submission-id> \
  --apple-id "your@email.com" \
  --team-id "YOUR_TEAM_ID" \
  --password "app-specific-password"

# 6. Staple the notarization ticket to the app
xcrun stapler staple ./build/YourApp.app

# 7. Verify stapling
xcrun stapler validate ./build/YourApp.app

# 8. Verify Gatekeeper will accept it
spctl -a -vvv -t install ./build/YourApp.app
```

### Complete Distribution Script

```bash
cat > scripts/notarize.sh << 'EOF'
#!/bin/bash
set -e

APP_NAME="YourApp"
PROJECT="${APP_NAME}.xcodeproj"
SCHEME="${APP_NAME}"
ARCHIVE_PATH="./build/${APP_NAME}.xcarchive"
EXPORT_PATH="./build"
APP_PATH="${EXPORT_PATH}/${APP_NAME}.app"
ZIP_PATH="${EXPORT_PATH}/${APP_NAME}.zip"

# Credentials (use environment variables for security)
APPLE_ID="${NOTARIZE_APPLE_ID}"
TEAM_ID="${NOTARIZE_TEAM_ID}"
PASSWORD="${NOTARIZE_PASSWORD}"  # App-specific password

echo "🏗️  Archiving..."
xcodebuild archive \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  -archivePath "$ARCHIVE_PATH" \
  | xcbeautify

echo "📦 Exporting..."
xcodebuild -exportArchive \
  -archivePath "$ARCHIVE_PATH" \
  -exportPath "$EXPORT_PATH" \
  -exportOptionsPlist ExportOptions.plist \
  | xcbeautify

echo "🤐 Creating ZIP..."
ditto -c -k --keepParent "$APP_PATH" "$ZIP_PATH"

echo "🔐 Submitting for notarization..."
SUBMISSION_ID=$(xcrun notarytool submit "$ZIP_PATH" \
  --apple-id "$APPLE_ID" \
  --team-id "$TEAM_ID" \
  --password "$PASSWORD" \
  --wait \
  --output-format json | jq -r '.id')

echo "📋 Submission ID: $SUBMISSION_ID"

echo "✅ Checking notarization status..."
xcrun notarytool info "$SUBMISSION_ID" \
  --apple-id "$APPLE_ID" \
  --team-id "$TEAM_ID" \
  --password "$PASSWORD"

echo "📌 Stapling notarization ticket..."
xcrun stapler staple "$APP_PATH"

echo "🔍 Verifying staple..."
xcrun stapler validate "$APP_PATH"

echo "🚪 Verifying Gatekeeper acceptance..."
spctl -a -vvv -t install "$APP_PATH"

echo "🎉 Notarization complete! App ready for distribution."
echo "App location: $APP_PATH"
EOF
chmod +x scripts/notarize.sh
```

### Creating DMG for Distribution

```bash
# Install create-dmg
brew install create-dmg

# Create DMG
create-dmg \
  --volname "YourApp" \
  --window-pos 200 120 \
  --window-size 800 400 \
  --icon-size 100 \
  --icon "YourApp.app" 200 190 \
  --hide-extension "YourApp.app" \
  --app-drop-link 600 185 \
  "YourApp.dmg" \
  "./build/YourApp.app"

# Notarize the DMG
xcrun notarytool submit YourApp.dmg \
  --apple-id "your@email.com" \
  --team-id "YOUR_TEAM_ID" \
  --password "app-specific-password" \
  --wait

# Staple to DMG
xcrun stapler staple YourApp.dmg
```

## Performance Profiling from CLI

### Using Instruments from Command Line

```bash
# List available instruments
instruments -s templates

# Profile with Time Profiler
instruments -t "Time Profiler" \
  -D ./profile.trace \
  ./build/Build/Products/Release/YourApp.app

# Profile with Allocations
instruments -t "Allocations" \
  -D ./allocations.trace \
  ./build/Build/Products/Release/YourApp.app

# Profile with Leaks
instruments -t "Leaks" \
  -D ./leaks.trace \
  ./build/Build/Products/Release/YourApp.app

# Run for specific duration (30 seconds)
instruments -t "Time Profiler" \
  -l 30000 \
  -D ./profile.trace \
  ./build/Build/Products/Release/YourApp.app
```

### Analyzing Trace Files

```bash
# Export trace data to XML
instruments -s trace.trace

# Use xctrace for more control
xctrace record \
  --template 'Time Profiler' \
  --output ./profile.trace \
  --launch ./build/Build/Products/Release/YourApp.app

# Export xctrace data
xctrace export \
  --input ./profile.trace \
  --output ./profile-data \
  --xpath '/trace-toc'
```

### Build Time Optimization

```bash
# Measure build time
time xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  clean build

# Enable build timing
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  build \
  OTHER_SWIFT_FLAGS="-Xfrontend -debug-time-compilation" \
  | grep "\.swift" | sort -rn | head -20

# Show which files take longest to compile
xcodebuild -project YourApp.xcodeproj \
  -scheme YourScheme \
  clean build \
  OTHER_SWIFT_FLAGS="-Xfrontend -debug-time-function-bodies" 2>&1 \
  | grep "[0-9]ms" | sort -rn | head -20
```

### Memory Leak Detection

```bash
# Run with MallocStackLogging
MallocStackLogging=1 ./build/Build/Products/Debug/YourApp.app/Contents/MacOS/YourApp

# Use leaks tool
leaks --atExit -- ./build/Build/Products/Debug/YourApp.app/Contents/MacOS/YourApp

# Continuous leak checking
leaks --list -- $(pgrep YourApp)
```

## Expanded Troubleshooting Guide

### Build Performance Issues

#### Slow Build Times

```bash
# Problem: Builds taking too long

# Solution 1: Use incremental builds
xcodebuild build -quiet  # Don't clean unless necessary

# Solution 2: Check what's being rebuilt
xcodebuild build -dry-run -verbose

# Solution 3: Enable parallel builds
xcodebuild build \
  -parallelizeTargets \
  -jobs $(sysctl -n hw.ncpu)

# Solution 4: Use build cache
xcodebuild build \
  -clonedSourcePackagesDirPath ./SourcePackages

# Solution 5: Disable index-while-building for CI
xcodebuild build \
  COMPILER_INDEX_STORE_ENABLE=NO
```

#### High Memory Usage During Build

```bash
# Problem: Build uses too much RAM

# Solution: Limit concurrent Swift compiler tasks
xcodebuild build \
  SWIFT_COMPILATION_MODE=incremental \
  -maximum-concurrent-test-device-destinations 1
```

### Code Signing Problems

#### "No signing identity found"

```bash
# Problem: Build fails with signing identity errors

# Solution 1: List available identities
security find-identity -v -p codesigning

# Solution 2: Unlock keychain
security unlock-keychain ~/Library/Keychains/login.keychain-db

# Solution 3: Disable signing for development
xcodebuild build \
  CODE_SIGN_IDENTITY="" \
  CODE_SIGNING_REQUIRED=NO

# Solution 4: Import certificate
security import certificate.p12 \
  -k ~/Library/Keychains/login.keychain-db \
  -P "password" \
  -T /usr/bin/codesign
```

#### Provisioning Profile Issues

```bash
# List installed provisioning profiles
ls ~/Library/MobileDevice/Provisioning\ Profiles/

# View provisioning profile details
security cms -D -i ~/Library/MobileDevice/Provisioning\ Profiles/profile.mobileprovision

# Clean and reinstall
rm -rf ~/Library/MobileDevice/Provisioning\ Profiles/*
# Then re-download from developer portal
```

### Dependency Resolution Problems

#### SPM Package Resolution Hangs

```bash
# Problem: Package resolution stuck or hanging

# Solution 1: Clear all caches
rm -rf ~/Library/Caches/org.swift.swiftpm
rm -rf ~/Library/Developer/Xcode/DerivedData
rm -rf .swiftpm
rm -rf .build

# Solution 2: Use verbose logging
xcodebuild -resolvePackageDependencies -verbose

# Solution 3: Reset package caches
swift package reset
swift package resolve

# Solution 4: Check for network issues
curl -I https://github.com  # Test GitHub connectivity
```

#### Version Conflicts

```bash
# Problem: "Dependencies could not be resolved"

# Solution 1: Check what versions are available
swift package show-dependencies --format json | jq

# Solution 2: Update Package.resolved
rm Package.resolved
swift package resolve

# Solution 3: Force update to latest compatible
swift package update

# Solution 4: Check for conflicting requirements
# Look at Package.swift dependencies across all packages
```

### Simulator Issues

#### Simulator Won't Boot

```bash
# Problem: "Unable to boot device"

# Solution 1: Reset simulator
xcrun simctl erase all

# Solution 2: Shutdown all simulators
xcrun simctl shutdown all

# Solution 3: Delete and recreate
xcrun simctl delete unavailable
xcrun simctl create "Fresh iPhone" "iPhone 15" "iOS17.0"

# Solution 4: Reset CoreSimulator
killall -9 com.apple.CoreSimulator.CoreSimulatorService
rm -rf ~/Library/Developer/CoreSimulator/Caches
```

#### App Won't Install on Simulator

```bash
# Problem: Installation fails

# Solution 1: Clean build folder
xcodebuild clean
rm -rf ~/Library/Developer/Xcode/DerivedData

# Solution 2: Rebuild for simulator
xcodebuild build \
  -sdk iphonesimulator \
  -destination 'platform=iOS Simulator,name=iPhone 15'

# Solution 3: Manually install
xcrun simctl install booted /path/to/YourApp.app
```

### Test Failures

#### Tests Pass in Xcode, Fail in CLI

```bash
# Problem: Tests inconsistent between Xcode and xcodebuild

# Solution 1: Ensure shared scheme
xcodebuild -list  # Verify scheme is listed

# Solution 2: Match test configuration
xcodebuild test \
  -configuration Debug \
  -destination 'platform=macOS' \
  -only-testing:YourTests

# Solution 3: Check environment differences
env | grep -i xcode  # See what Xcode sets
```

#### Flaky Tests

```bash
# Problem: Tests fail intermittently

# Solution 1: Run tests multiple times
for i in {1..10}; do
  echo "Run $i"
  xcodebuild test -project YourApp.xcodeproj -scheme YourScheme || break
done

# Solution 2: Disable parallel testing
xcodebuild test \
  -parallel-testing-enabled NO \
  -maximum-parallel-testing-workers 1

# Solution 3: Run specific test repeatedly
xcodebuild test \
  -only-testing:TestClass/testFlakyTest \
  -run-tests-until-failure
```

### File System Issues

#### "Operation not permitted"

```bash
# Problem: Build fails with permission errors

# Solution 1: Check Full Disk Access
# System Settings → Privacy & Security → Full Disk Access
# Add Terminal.app or your terminal emulator

# Solution 2: Fix ownership
sudo chown -R $(whoami) ~/Library/Developer/Xcode

# Solution 3: Repair permissions
sudo chmod -R 755 ~/Library/Developer/Xcode/DerivedData
```

#### Disk Space Problems

```bash
# Problem: "No space left on device"

# Solution 1: Clean DerivedData
rm -rf ~/Library/Developer/Xcode/DerivedData

# Solution 2: Clean old archives
rm -rf ~/Library/Developer/Xcode/Archives

# Solution 3: Clean simulator data
xcrun simctl delete unavailable
xcrun simctl erase all

# Solution 4: Clean package caches
rm -rf ~/Library/Caches/org.swift.swiftpm
rm -rf ~/.swiftpm
```

### Debugging Helper Commands

```bash
# Show all xcodebuild-related processes
ps aux | grep xcodebuild

# Kill stuck xcodebuild
pkill -9 xcodebuild

# Show derived data size
du -sh ~/Library/Developer/Xcode/DerivedData

# Find large cache directories
du -sh ~/Library/Caches/* | sort -rh | head -10

# Check Xcode installation
xcode-select -p
xcodebuild -version -sdk

# Verify code signing setup
security find-identity -v
security default-keychain
```

## Resources

- [xcodebuild man page](https://developer.apple.com/library/archive/technotes/tn2339/_index.html)
- [SwiftLint GitHub](https://github.com/realm/SwiftLint)
- [SwiftFormat GitHub](https://github.com/nicklockwood/SwiftFormat)
- [XcodeGen](https://github.com/yonaskolb/XcodeGen)
- [xcpretty GitHub](https://github.com/xcpretty/xcpretty)
- [Fastlane Documentation](https://docs.fastlane.tools/)
- [Notarization Guide](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [Instruments User Guide](https://help.apple.com/instruments/mac/current/)

## Summary

The key to successful CLI automation for Swift macOS projects:

1. **Always use shared schemes** - Make schemes visible to command-line tools
2. **Standardize on tools** - Use SwiftLint and SwiftFormat consistently
3. **Script common workflows** - Create scripts for build, test, validate
4. **Integrate with Claude CLI** - Use hooks for automated validation
5. **Clean aggressively** - When in doubt, clean derived data
6. **Version your tools** - Use Mint or similar to pin tool versions
7. **Test your CI locally** - Use the same commands locally and in CI
8. **Keep it simple** - Don't over-engineer; automation should reduce complexity

With these tools and practices, you can develop Swift macOS applications entirely from the command line, enabling full automation with Claude CLI and other AI-assisted development tools.
