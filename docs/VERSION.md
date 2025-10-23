# Version Management

This document describes how version information is managed in the ARM Emulator project.

## Overview

The project uses **git tags** as the single source of truth for version numbers, following semantic versioning (SemVer). Version information is embedded into the binary at build time using Go's `-ldflags` mechanism.

## Version Information

Three pieces of information are embedded in each build:

1. **Version** - Semantic version from git tag (e.g., `v1.0.1`)
2. **Commit** - Short git commit hash (e.g., `7330b81`)
3. **Date** - ISO 8601 build timestamp (e.g., `2025-10-23T18:41:23Z`)

## Viewing Version Information

Check the version of any built binary:

```bash
./arm-emulator --version
```

Example output:
```
ARM2 Emulator v1.0.1
Commit: 7330b81
Built: 2025-10-23T18:41:23Z
```

## Building with Version Information

### Using Make (Recommended)

The Makefile automatically extracts version information from git:

```bash
make build
```

To see what version will be embedded:

```bash
make version
```

### Manual Build

Extract version info and build manually:

```bash
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
go build -ldflags "-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.Date=$DATE" -o arm-emulator
```

### Development Builds

If no git tag exists, the version defaults to `dev`:

```bash
go build -o arm-emulator
./arm-emulator --version
# Output: ARM2 Emulator dev
```

## Creating Releases

### Local Tagging

1. Tag the commit:
   ```bash
   git tag v1.0.1
   ```

2. Build with the new version:
   ```bash
   make build
   ./arm-emulator --version  # Shows v1.0.1
   ```

3. Push the tag to trigger automated release builds:
   ```bash
   git push origin v1.0.1
   ```

### Automated Release Process

When a tag starting with `v` is pushed to GitHub:

1. **GitHub Actions** workflow triggers (`.github/workflows/build-release.yml`)
2. Builds optimized binaries for multiple platforms:
   - linux-amd64
   - macos-arm64
   - windows-amd64
   - windows-arm64
3. Embeds version info into each binary
4. Generates SHA256 checksums
5. Creates a GitHub Release with all artifacts

Each released binary contains the exact version, commit, and build timestamp.

## Semantic Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** version (v1.0.0 → v2.0.0): Breaking changes to ARM assembly compatibility or command-line interface
- **MINOR** version (v1.0.0 → v1.1.0): New features, new syscalls, debugger enhancements (backward compatible)
- **PATCH** version (v1.0.0 → v1.0.1): Bug fixes, documentation updates (backward compatible)

## Implementation Details

### Code Location

Version variables are defined in `main.go`:

```go
var (
    Version = "dev"      // Version number (set by git tag at build time)
    Commit  = "unknown"  // Git commit hash
    Date    = "unknown"  // Build date
)
```

### Build-time Injection

The `-ldflags` mechanism replaces these default values:

```bash
-ldflags "-X main.Version=v1.0.1 -X main.Commit=7330b81 -X main.Date=2025-10-23T18:41:23Z"
```

### Version Detection Logic

The Makefile uses `git describe` to automatically determine the version:

```makefile
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
```

This provides:
- **Tagged commits**: Exact tag (e.g., `v1.0.1`)
- **After tag**: Tag + commits + hash (e.g., `v1.0.1-3-g7330b81`)
- **Dirty working tree**: Appends `-dirty` (e.g., `v1.0.1-dirty`)
- **No tags**: Commit hash or `dev` fallback

## Best Practices

1. **Always tag releases** - Use `git tag vX.Y.Z` for all releases
2. **Build with Make** - Use `make build` to ensure version info is embedded
3. **Check version output** - Verify `--version` shows correct information
4. **Clean working tree** - Avoid building with uncommitted changes (shows `-dirty`)
5. **Push tags** - Don't forget `git push origin vX.Y.Z` to trigger automated releases

## Troubleshooting

### "dev" version shown

**Cause**: No git tags exist in the repository.

**Solution**: Create and push a tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```

### "-dirty" suffix in version

**Cause**: Uncommitted changes in the working tree.

**Solution**: Commit or stash your changes before building.

### Version not updating

**Cause**: Binary wasn't rebuilt with version flags.

**Solution**: Use `make build` or ensure `-ldflags` are passed to `go build`.

## Version History

- **v1.0.1** (2025-10-23): Professional version management implementation
- **v1.0.0** (2025-10): Initial stable release
- **v0.9.0** (2025-10): Release automation and multi-platform builds
