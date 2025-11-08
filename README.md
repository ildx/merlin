# Merlin

> A macOS-focused CLI tool for managing dotfiles with style ✨

<img width="1400" height="700" alt="image" src="https://github.com/user-attachments/assets/fadd014d-03d6-4731-992b-83da7d63ab3e" />

Merlin is a powerful yet simple dotfiles manager that handles package installation, configuration symlinking, and custom setup scripts—all from declarative TOML files. Built with Go and [Charm](https://charm.sh/) for a beautiful terminal experience.

---

## What is Merlin?

Merlin manages your entire macOS setup:
- 📦 **Homebrew packages** (formulae & casks)
- 🍎 **Mac App Store apps**
- 🔗 **Dotfiles symlinking** (native, no external tools)
- 🔧 **Custom install scripts** for complex setups
- 📋 **Profiles** for different machines (work, personal, minimal)
- 🎨 **Beautiful TUI** powered by Bubble Tea

**Framework-agnostic:** Works with any dotfiles repository following the expected structure (see [DOTFILES_STRUCTURE.md](./docs/DOTFILES_STRUCTURE.md)).

---

## Status

🚧 **In Development** - Merlin is currently being built. See [IMPLEMENTATION_PLAN.md](./docs/IMPLEMENTATION_PLAN.md) for progress.

---

## Quick Start (Future)

```bash
# Install merlin
go install github.com/yourusername/merlin@latest

# In your dotfiles repository
cd ~/dotfiles

# List available tools
merlin list configs

# Link a specific tool's configs
merlin link zsh

# Link all configs
merlin link --all

# Install Homebrew packages
merlin install brew

# Interactive TUI
merlin tui
```

---

## Example Dotfiles Structure

```
your-dotfiles/
├── merlin.toml              # Global settings & profiles
└── config/
    ├── git/
    │   ├── config/          # Files to symlink
    │   └── merlin.toml      # Optional: custom config
    ├── zsh/
    │   ├── config/
    │   │   ├── .zshrc
    │   │   └── defaults/
    │   └── merlin.toml
    ├── cursor/
    │   ├── config/
    │   ├── install.sh       # Custom install script
    │   └── merlin.toml
    ├── brew/
    │   └── config/
    │       └── brew.toml    # Homebrew packages
    └── mas/
        └── config/
            └── mas.toml     # Mac App Store apps
```

---

## Key Features

### 📋 Declarative Configuration

Define your entire setup in TOML files:

```toml
# Root merlin.toml - global settings
[metadata]
name = "my-dotfiles"

[settings]
conflict_strategy = "backup"

# Per-tool config/cursor/merlin.toml
[tool]
name = "cursor"

[[link]]
source = "config/settings.json"
target = "~/Library/Application Support/Cursor/User/settings.json"

install_script = "install.sh"
```

### 🔗 Smart Symlinking

- **Native implementation** (no Stow dependency)
- **Conflict resolution** (backup, skip, overwrite, interactive)
- **Custom paths** per tool
- **Smart defaults** for tools without config

### 🎯 Profiles

Different setups for different machines:

```toml
[[profile]]
name = "personal"
hostname = "MacBook-Pro"
tools = ["git", "zsh", "cursor", "eza"]

[[profile]]
name = "work"
hostname = "Work-MacBook"
tools = ["git", "zsh"]
```

### 🎨 Beautiful TUI

Interactive interface powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea):
- Browse and select packages to install
- Choose configs to link
- Live progress indicators
- Grouped by categories

---

## Documentation

- **[DOTFILES_STRUCTURE.md](./docs/DOTFILES_STRUCTURE.md)** - How to structure your dotfiles for Merlin
- **[MERLIN_TOML_SPEC.md](./docs/MERLIN_TOML_SPEC.md)** - Complete merlin.toml specification
- **[IMPLEMENTATION_PLAN.md](./docs/IMPLEMENTATION_PLAN.md)** - Development roadmap

---

## Architecture

```
merlin/
├── cmd/                # CLI commands (cobra)
│   ├── root.go
│   ├── list.go
│   ├── install.go
│   ├── link.go
│   └── tui.go
├── internal/
│   ├── config/         # Dotfiles path resolution
│   ├── models/         # Data structures
│   ├── parser/         # TOML parsing
│   ├── installer/      # brew, mas installers
│   ├── symlink/        # Native symlinking logic
│   ├── scripts/        # Install script execution
│   └── tui/            # Bubble Tea UI
└── docs/
```

### Tech Stack

- **Go** - Fast, simple, single binary
- **Cobra** - CLI framework
- **Bubble Tea** - Terminal UI framework
- **Lip Gloss** - Styling & layout
- **TOML** - Configuration format

### Design Principles

1. **Small, testable steps** - Each phase independently testable
2. **Minimal dependencies** - No external tools (no Stow)
3. **Modular config** - Root + per-tool merlin.toml files
4. **Idempotent operations** - Safe to run multiple times
5. **User feedback** - Always show what's happening

---

## Development

### Prerequisites

- Go 1.21+
- macOS (for full functionality)

### Building

```bash
# Clone the repo
git clone https://github.com/yourusername/merlin.git
cd merlin

# Build
go build -o merlin

# Run
./merlin --help
```

### Testing

```bash
# Run tests
go test ./...

# Test with a dotfiles repo
cd /path/to/your/dotfiles
/path/to/merlin list configs
```

---

## Example Dotfiles Repository

See [covenant](https://github.com/yourusername/covenant) for a complete example dotfiles repository configured for Merlin.

---

## Philosophy

Merlin is built on the belief that dotfiles management should be:
- **Declarative** - Define what you want, not how to get it
- **Transparent** - Always show what's happening
- **Flexible** - Support both simple and complex setups
- **Beautiful** - Make the terminal a joy to use

---

## Roadmap

- [x] Design architecture & configuration format
- [x] Document expected dotfiles structure
- [ ] Phase 1: Foundation & CLI framework
- [ ] Phase 2: TOML parsing
- [ ] Phase 3: Homebrew installation
- [ ] Phase 4: Mac App Store installation
- [ ] Phase 5: Native symlinking
- [ ] Phase 6: Custom install scripts
- [ ] Phase 7: Bubble Tea TUI
- [ ] Phase 8: Advanced features (dry-run, logging, profiles)
- [ ] Phase 9: Documentation & polish
- [ ] 1.0 Release

See [IMPLEMENTATION_PLAN.md](./docs/IMPLEMENTATION_PLAN.md) for detailed progress.

---

## Why "Merlin"?

Because managing dotfiles should feel like magic ✨

---

## License

MIT License - see [LICENSE](./LICENSE) for details

---

## Contributing

Contributions welcome! Please read the [IMPLEMENTATION_PLAN.md](./docs/IMPLEMENTATION_PLAN.md) to understand the architecture first.

---

Built with ❤️ and ☕

