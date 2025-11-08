# Merlin Implementation Plan

**Version:** 1.0  
**Date:** November 8, 2025  
**Tech Stack:** Go, [Charm](https://charm.sh/) libs

---

## Overview

Merlin is a **macOS-focused** CLI tool for managing dotfiles repositories. It will:
- Parse TOML configuration files
- Install Homebrew packages and Mac App Store apps
- Manage dotfiles via native symlinking
- Execute custom install scripts for complex configurations
- Provide both CLI commands and interactive TUI

**Framework-agnostic:** Works with any dotfiles repo following the expected structure.

---

## Architecture Principles

1. **Small, testable steps** - Each phase can be tested independently
2. **Minimal dependencies** - Only add libraries when needed
3. **Hybrid approach** - Use native symlinking for simple configs, keep custom install scripts for complex setups
4. **Expected structure** - All configs follow `dotfiles/config/TOOL/config/FILES` pattern
5. **No external tools** - Self-contained, no dependency on Stow or other symlink managers

---

## Phase 1: Foundation & Project Setup

**Goal:** Create a working Go project with basic structure and CLI framework

### Step 1.1: Initialize Go Project
- Create `merlin/` directory structure
- Initialize `go.mod`
- Set up basic `main.go` with entry point
- Create `cmd/` directory for commands
- **Test:** Run `go run main.go` - should print version/help

**Dependencies:** None (stdlib only)

### Step 1.2: Add Cobra for CLI Framework
- Install `github.com/spf13/cobra`
- Set up root command with version flag
- Add `merlin --version` and `merlin --help`
- **Test:** Commands should execute and print expected output

**Dependencies:** `cobra`

### Step 1.3: Create Config Package
- Create `internal/config/` package
- Add `config.go` with types for:
  - Dotfiles repository path resolution
  - Config directory discovery
- Add function to find dotfiles root (check current dir, parent dirs, env var)
- **Test:** Unit tests for dotfiles path resolution

**Dependencies:** None (stdlib only)

---

## Phase 2: TOML Parsing

**Goal:** Parse covenant TOML files into Go structs

### Step 2.1: Define Data Models
- Create `internal/models/` package
- Define structs for:
  - `BrewConfig` (formulae, casks, categories)
  - `MASConfig` (apps, categories)
  - `MerlinConfig` (tools, profiles, settings)
  - `ToolConfig` (links, scripts, hooks)
  - `Metadata` (version, description)
- **Test:** Structs compile and have correct tags

**Dependencies:** None

### Step 2.2: Add TOML Parser
- Install `github.com/BurntSushi/toml`
- Create `internal/parser/` package
- Implement `ParseBrewTOML()` function
- Implement `ParseMASTOML()` function
- Implement `ParseRootMerlinTOML()` function (root config)
- Implement `ParseToolMerlinTOML()` function (per-tool config)
- **Test:** Parse real `brew.toml`, `mas.toml`, root and tool merlin.toml files, verify structs

**Dependencies:** `toml`

### Step 2.3: Create "list" Command
- Add `cmd/list.go` for `merlin list` command
- Support `merlin list brew` - shows all brew packages
- Support `merlin list mas` - shows all MAS apps
- Display in simple text format (name, description, category)
- **Test:** Run commands, verify output matches TOML contents

**Dependencies:** None (uses Phase 2.1-2.2)

---

## Phase 3: Homebrew Installation

**Goal:** Install Homebrew packages from brew.toml

### Step 3.1: Check Prerequisites
- Create `internal/system/` package
- Add `CheckHomebrew()` - verify brew is installed
- Add `CheckCommand()` - generic command existence check
- **Test:** Functions correctly detect installed/missing commands

**Dependencies:** None (stdlib only)

### Step 3.2: Brew Install - Formulae Only
- Create `internal/installer/brew.go`
- Implement `InstallFormulae()` - installs CLI packages
- Check if already installed (skip if yes)
- Execute `brew install <package>` with output streaming
- Add `merlin install brew` command (formulae only for now)
- **Test:** Install a single formulae, verify it works

**Dependencies:** None

### Step 3.3: Brew Install - Casks
- Extend `InstallCasks()` in `brew.go`
- Execute `brew install --cask <app>`
- Add flag: `merlin install brew --casks`
- **Test:** Install a single cask, verify it appears in Applications

**Dependencies:** None

### Step 3.4: Interactive Selection
- Add basic prompt: "Install all or select?" (use stdlib for now)
- If select: show numbered list, user enters numbers
- **Test:** User can select subset of packages to install

**Dependencies:** None (basic stdin/stdout)

---

## Phase 4: Mac App Store Installation

**Goal:** Install MAS apps from mas.toml

### Step 4.1: MAS Integration
- Check if `mas` is installed
- Implement `InstallMASApp()` in `internal/installer/mas.go`
- Execute `mas install <id>`
- Check if already installed: `mas list | grep <id>`
- Add `merlin install mas` command
- **Test:** Install an app from App Store via ID

**Dependencies:** None

### Step 4.2: MAS Account Check
- Add `CheckMASAccount()` - verify signed into App Store
- Display helpful message if not signed in
- **Test:** Gracefully handle not-signed-in state

**Dependencies:** None

---

## Phase 5: Native Dotfiles Symlinking

**Goal:** Implement native symlinking without external dependencies

### Step 5.1: Config Discovery
- Create `internal/symlink/` package
- Add `DiscoverConfigPackages()` - scan `<dotfiles>/config/*/config/` dirs
- For each tool, check if `config/TOOL/merlin.toml` exists
- If exists, parse tool-specific config; otherwise use defaults
- Read root `merlin.toml` for global settings
- Add `merlin list configs` command
- **Test:** Lists all available config packages, respects per-tool merlin.toml settings

**Dependencies:** None (uses Phase 2 parser)

### Step 5.2: Symlink Core Logic
- Implement `WalkAndLink(sourceDir, targetDir)` - recursively walk directories
- For each file: create symlink from target → source
- Preserve directory structure
- Handle nested directories (create parent dirs as needed)
- Add `internal/symlink/conflict.go` for conflict detection
- Support custom source/target from per-tool `merlin.toml` `[[link]]` entries
- If no per-tool config, use defaults: `config/` → `~/.config/TOOL/`
- **Test:** Unit tests for directory walking and symlink creation with custom paths and defaults

**Dependencies:** None (stdlib `os`, `filepath`)

### Step 5.3: Conflict Resolution
- Detect existing files/symlinks at target location
- Implement strategies:
  - **Skip** - leave existing file alone
  - **Backup** - move existing file to `.backup.timestamp`
  - **Overwrite** - replace existing file
  - **Interactive** - ask user what to do
- Respect `conflict_strategy` from `merlin.toml` settings
- Add `merlin link <tool>` command with `--strategy` flag (overrides config)
- **Test:** Link a single config (e.g., eza), handle conflicts gracefully

**Dependencies:** None

### Step 5.4: Batch Linking & Status
- Add `merlin link --all` - link all discovered configs
- Add `merlin link --select` - interactive selection
- Implement `GetLinkStatus(tool)` - check if tool's configs are already linked
- Show status: ✓ linked, ✗ not linked, ⚠ conflict
- **Test:** Link multiple configs, verify status reporting

**Dependencies:** None

### Step 5.5: Unlinking Support
- Implement `UnlinkConfig(tool)` - remove symlinks for a tool
- Only remove if symlink points to our dotfiles (safety check)
- Leave regular files untouched
- Add `merlin unlink <tool>` command
- Add `merlin unlink --all`
- **Test:** Unlink removes only our symlinks, preserves other files

**Dependencies:** None

---

## Phase 6: Custom Install Scripts

**Goal:** Execute tool-specific install scripts for complex setups

### Step 6.1: Script Discovery
- Add `DiscoverInstallScripts()` in `internal/scripts/` package
- Scan for `<dotfiles>/config/*/install.sh` files
- Add to `merlin list scripts` output
- **Test:** Correctly finds all install.sh scripts

**Dependencies:** None

### Step 6.2: Script Execution & Hooks
- Implement `ExecuteInstallScript(tool string)`
- Run with proper working directory
- Stream output to terminal
- Capture exit code
- Support `pre_install`, `install_script`, and `post_install` from per-tool `merlin.toml`
- Execute in order: pre → symlinks → install script → post
- Add `merlin install script <tool>` command
- **Test:** Execute cursor install.sh, verify extensions install; test zellij post_install

**Dependencies:** None

### Step 6.3: Requirements Check
- Implement `CheckRequirements(tool)` - verify `requires` packages from per-tool `merlin.toml`
- Check if required brew formulae/casks are installed
- Suggest installation command if missing
- Add `--skip-requirements` flag to bypass checks
- **Test:** Helpful warnings for missing dependencies, suggest installation

**Dependencies:** None

---

## Phase 7: Enhanced TUI with Bubble Tea

**Goal:** Beautiful interactive interface for all operations

### Step 7.1: Install Bubble Tea & Friends
- Add `github.com/charmbracelet/bubbletea`
- Add `github.com/charmbracelet/lipgloss`
- Add `github.com/charmbracelet/bubbles`
- Create `internal/tui/` package
- **Test:** Basic hello world Bubble Tea app runs

**Dependencies:** `bubbletea`, `lipgloss`, `bubbles`

### Step 7.2: Main Menu TUI
- Create main menu model with bubbles list component
- Options: "Install Packages", "Manage Dotfiles", "Run Scripts", "Quit"
- Add `merlin tui` or just `merlin` (interactive mode)
- **Test:** Navigation works, selecting options triggers placeholder actions

**Dependencies:** Uses 7.1

### Step 7.3: Package Selection TUI
- Create package picker with checkboxes
- Show brew formulae, casks, and MAS apps
- Group by categories (from TOML)
- Filter/search functionality
- Confirm selection → trigger installation
- **Test:** Can select/deselect packages, install selected ones

**Dependencies:** Uses 7.1

### Step 7.4: Progress Indicators
- Add spinner for in-progress operations
- Progress bar for batch operations (e.g., installing 10 packages)
- Real-time output streaming in TUI
- **Test:** Visual feedback during installations

**Dependencies:** Uses 7.1, bubbles progress/spinner

### Step 7.5: Dotfiles TUI
- Create config package picker
- Show which configs are already linked (✓ linked, ✗ not linked, ⚠ conflict)
- Support link/unlink operations with conflict resolution options
- **Test:** Interactive link/unlink via TUI

**Dependencies:** Uses 7.1

---

## Phase 8: Advanced Features

**Goal:** Polish and power-user features

### Step 8.1: Dry Run Mode
- Add `--dry-run` flag to all commands
- Print what would be done without doing it
- **Test:** Dry run accurately previews actions

**Dependencies:** None

### Step 8.2: Logging
- Add `github.com/charmbracelet/log`
- Configurable log levels (debug, info, warn, error)
- Log file: `~/.merlin/merlin.log`
- Add `--verbose` flag
- **Test:** Logs are written correctly

**Dependencies:** `log`

### Step 8.3: Config Validation
- Validate TOML files on load
- Check for duplicate entries
- Validate category references
- Warn about missing icons/metadata
- Add `merlin validate` command
- **Test:** Catches common TOML errors

**Dependencies:** None

### Step 8.4: Profiles Support
- Parse `[[profile]]` sections from root `merlin.toml`
- Each profile specifies subset of tools/packages
- Support hostname-based auto-selection
- Add `merlin link --profile <name>` 
- Add `merlin install --profile <name>`
- Add `merlin list profiles` command
- **Test:** Profiles correctly filter tools and packages

**Dependencies:** None (uses Phase 2 parser)

---

## Phase 9: Documentation & Polish

**Goal:** Make Merlin production-ready

### Step 9.1: Help Text & Examples
- Add detailed help text to all commands
- Add examples in help output
- Create `docs/USAGE.md`
- **Test:** Help text is clear and accurate

**Dependencies:** None

### Step 9.2: Error Handling
- Graceful error messages (no stack traces for users)
- Suggestions for common errors (e.g., "brew not installed? Run: ...")
- Add colors to error output (via lipgloss)
- **Test:** Intentionally trigger errors, verify messages are helpful

**Dependencies:** lipgloss (already added)

### Step 9.3: README & Install Instructions
- Create comprehensive README for merlin
- Installation via `go install` or build script
- Screenshots of TUI
- **Test:** Fresh user can follow README and install merlin

**Dependencies:** None

---

## Testing Strategy

After each step:
1. **Manual testing** - Run the command/feature, verify it works
2. **Unit tests** (where applicable) - Test pure functions
3. **Integration check** - Ensure new code doesn't break previous steps

### Test Cases to Track:
- Parse valid/invalid TOML files
- Install packages that exist/don't exist
- Stow files when symlinks already exist
- Execute scripts with/without permissions
- TUI navigation and selection
- Dry-run mode accuracy

---

## Directory Structure (Final)

```
merlin/
├── docs/
│   ├── IMPLEMENTATION_PLAN.md    (this file)
│   ├── DOTFILES_STRUCTURE.md     (dotfiles repo structure guide)
│   ├── MERLIN_TOML_SPEC.md       (merlin.toml specification)
│   └── USAGE.md                  (Phase 9)
├── cmd/
│   ├── root.go                   (cobra root command)
│   ├── list.go                   (list packages/configs)
│   ├── install.go                (install packages)
│   ├── link.go                   (link/unlink dotfiles)
│   ├── validate.go               (validate TOML)
│   └── tui.go                    (interactive TUI)
├── internal/
│   ├── config/                   (dotfiles repo path resolution)
│   ├── models/                   (data structures)
│   ├── parser/                   (TOML parsing)
│   ├── installer/                (brew, mas installers)
│   ├── symlink/                  (native symlinking logic)
│   ├── scripts/                  (install script execution)
│   ├── system/                   (system checks)
│   └── tui/                      (Bubble Tea UI)
├── go.mod
├── go.sum
├── main.go
└── README.md
```

---

## Success Criteria

- ✅ User can install all brew packages with one command
- ✅ User can install all MAS apps with one command
- ✅ User can link all configs with one command
- ✅ User can run custom install scripts via merlin
- ✅ Interactive TUI makes everything easy and beautiful
- ✅ Dry-run mode allows safe previews
- ✅ Works on fresh macOS installation
- ✅ No external dependencies (Stow, etc.) required

---

## Notes

- **Modular config:** Root `merlin.toml` for global settings/profiles, per-tool `merlin.toml` for tool-specific behavior
- **Symlinks vs Scripts:** Use native symlinking for simple configs (eza, mise). Keep custom scripts for complex setups (JSON generation, extension installation, theme symlinking).
- **Smart defaults:** Tools without per-tool `merlin.toml` get automatic config discovery and default symlink paths
- **Order matters:** Some tools depend on others (e.g., an editor needs to be installed before running its extension install script).
- **Idempotency:** All operations should be safe to run multiple times.
- **User feedback:** Always show what's happening, never silent operations.
- **Dotfiles discovery:** Merlin finds dotfiles via current dir → parent dirs → `MERLIN_DOTFILES` env var
- **Symlink safety:** Always verify symlink target before removing; never delete regular files

---

## Future Enhancements (Post-MVP)

- Backup/restore functionality
- Diff between current state and dotfiles configs
- Git integration (auto-commit changes to dotfiles)
- Remote dotfiles repos (clone from GitHub/custom URL)
- Uninstall mode (remove packages)
- Update check (compare installed vs TOML versions)
- Export current system state to TOML

