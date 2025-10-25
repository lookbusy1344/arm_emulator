# Documentation

This directory contains comprehensive documentation for the ARM2 Emulator project.

## Getting Started

Start here if you're new to the emulator:

- **[Installation Guide](installation.md)** - How to install and set up the ARM2 Emulator on macOS, Windows, and Linux
- **[Tutorial](TUTORIAL.md)** - Learn ARM2 assembly programming from scratch with hands-on examples
- **[FAQ](FAQ.md)** - Common questions, troubleshooting tips, and platform-specific issues

## Reference Documentation

Complete reference materials for programming:

- **[Instruction Set Reference](INSTRUCTIONS.md)** - Complete ARM2 CPU instruction set with detailed descriptions of all data processing, memory, branch, multiply, and system instructions (including SWI syscalls)
- **[Assembler Directives](ASSEMBLER.md)** - Assembler directives and syntax (.text, .data, .word, .ltorg, character literals, etc.)
- **[Programming Reference](REFERENCE.md)** - Condition codes, addressing modes, shift operations, pseudo-instructions, and register usage conventions
- **[Assembly Language Reference](assembly_reference.md)** - Complete language reference covering program structure, registers, data types, and addressing modes

## Debugging

Tools and guides for debugging ARM2 programs:

- **[Debugger Reference](debugger_reference.md)** - Complete debugger documentation covering both command-line mode and TUI (Text User Interface) with breakpoints, watchpoints, and expression evaluation
- **[Debugging Tutorial](debugging_tutorial.md)** - Practical debugging sessions with real example programs

## Developer Documentation

For developers extending or integrating the emulator:

- **[API Reference](API.md)** - Comprehensive API documentation for all packages (VM, Parser, Debugger, Encoder, Tools, Config)
- **[Architecture Overview](architecture.md)** - Internal architecture, project structure, package organization, and execution pipeline
- **[Literal Pool Implementation](ltorg_implementation.md)** - Technical details of the .ltorg directive and dynamic literal pool management

## Project Information

Release notes, version history, and security:

- **[Changelog](CHANGELOG.md)** - Version history following [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format
- **[Version](VERSION.md)** - Current version information
- **[Release v1.0 Summary](RELEASE_v1.0_SUMMARY.md)** - First production release overview and highlights
- **[Release Checklist](RELEASE_CHECKLIST.md)** - Release preparation and verification checklist
- **[Security](SECURITY.md)** - Security policy, reporting vulnerabilities, and hardening measures
- **[Security Audit Summary](SECURITY_AUDIT_SUMMARY.md)** - Comprehensive security audit report addressing anti-virus false positives

## Internal Documentation

Development process and code review documentation:

- **[Implementation Plan](IMPLEMENTATION_PLAN.md)** - Original project development plan and phases
- **[Code Review](CODE_REVIEW.md)** - Detailed code review findings and recommendations
- **[Review Summary](REVIEW_SUMMARY.md)** - Summary of code review results

---

## Quick Links

### I want to...

- **Learn ARM2 assembly** → Start with [Tutorial](TUTORIAL.md)
- **Look up an instruction** → See [Instruction Set Reference](INSTRUCTIONS.md)
- **Use system calls** → Check [System Instructions](INSTRUCTIONS.md#system-instructions) section
- **Debug a program** → Read [Debugging Tutorial](debugging_tutorial.md)
- **Understand assembler syntax** → Review [Assembler Directives](ASSEMBLER.md)
- **Install the emulator** → Follow [Installation Guide](installation.md)
- **Extend the emulator** → Study [API Reference](API.md) and [Architecture Overview](architecture.md)
- **Report a security issue** → See [Security Policy](SECURITY.md)

## Documentation Standards

All documentation follows these conventions:

- GitHub-flavored Markdown format
- Cross-referenced with relative links
- Code examples use syntax highlighting
- Tables for reference material
- Clear section hierarchy with anchors

## Contributing

When adding or updating documentation:

1. Use clear, concise language
2. Include practical examples
3. Cross-reference related documents
4. Update this README if adding new files
5. Follow the existing formatting style
