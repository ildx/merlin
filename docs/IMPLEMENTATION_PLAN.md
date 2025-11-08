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
- Execute `mas install <id>`
- Add `merlin install mas` command
- **Test:** Install an app from App Store via ID

**Dependencies:** None

### Step 4.2: MAS Account Check
- Display helpful message if not signed in
- **Test:** Gracefully handle not-signed-in state

**Dependencies:** None

---
**Goal:** Implement native symlinking without external dependencies

- For each tool, check if `config/TOOL/merlin.toml` exists
- If exists, parse tool-specific config; otherwise use defaults

**Dependencies:** None (uses Phase 2 parser)
- For each file: create symlink from target → source
- Preserve directory structure
- If no per-tool config, use defaults: `config/` → `~/.config/TOOL/`
- **Test:** Unit tests for directory walking and symlink creation with custom paths and defaults
### Step 5.3: Conflict Resolution
- Detect existing files/symlinks at target location
  - **Overwrite** - replace existing file
  - **Interactive** - ask user what to do
- **Test:** Link a single config (e.g., eza), handle conflicts gracefully

- Add `merlin link --all` - link all discovered configs
- Add `merlin link --select` - interactive selection

**Dependencies:** None
- Leave regular files untouched
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

## Phase 7: Enhanced TUI with Bubble Tea ✅

**Goal:** Beautiful interactive interface for all operations

### Step 7.1: Install Bubble Tea & Friends ✅
- ✅ Add `github.com/charmbracelet/bubbletea`
- ✅ Add `github.com/charmbracelet/lipgloss`
- ✅ Add `github.com/charmbracelet/bubbles`
- ✅ Create `internal/tui/` package
- **Test:** Basic hello world Bubble Tea app runs

**Dependencies:** `bubbletea`, `lipgloss`, `bubbles`

### Step 7.2: Main Menu TUI ✅
- ✅ Create main menu model with navigation
- ✅ Options: "Install Packages", "Manage Dotfiles", "Run Scripts", "Doctor", "Quit"
- ✅ Add `merlin tui` command
- ✅ Default to TUI when running `merlin` with no subcommand
- **Test:** Navigation works, selecting options triggers actions

**Dependencies:** Uses 7.1

### Step 7.3: Package Selection TUI ✅
- ✅ Create package picker with checkboxes
- ✅ Show brew formulae and casks
- ✅ Group by categories (from TOML)
- ✅ Package type selection submenu (formulae/casks/both)
- ✅ Confirm selection → trigger installation
- **Test:** Can select/deselect packages, install selected ones

**Dependencies:** Uses 7.1

### Step 7.4: Progress Indicators ✅
- ✅ Add spinner for in-progress operations
- ✅ Progress bar for batch operations (e.g., installing 10 packages)
- ✅ Progress models created for single and batch operations
- **Test:** Visual feedback during installations

**Dependencies:** Uses 7.1, bubbles progress/spinner

### Step 7.5: Dotfiles TUI ✅
- ✅ Create config package picker
- ✅ Show which configs are linked (✓ linked, ⚠ conflict, ○ not linked)
- ✅ Support link/unlink operations
- ✅ Action menu for selecting link or unlink
- **Test:** Interactive link/unlink via TUI

**Dependencies:** Uses 7.1

**Status:** Phase 7 complete! All TUI functionality implemented and tested.

---

## Phase 8: Advanced Features ✅

**Goal:** Polish and power-user features

### Step 8.1: Dry Run Mode ✅
- ✅ `--dry-run` flag already implemented on all commands (Phase 1-6)
- ✅ Print what would be done without doing it
- **Test:** Dry run accurately previews actions ✅

**Dependencies:** None

### Step 8.2: Logging ✅
- ✅ Add `github.com/charmbracelet/log`
- ✅ Configurable log levels (debug, info, warn, error)
- ✅ Log file: `~/.merlin/merlin.log`
- ✅ `--verbose` flag integrated with log level
- ✅ Logs initialized on startup via cobra OnInitialize
- **Test:** Logs are written correctly ✅

**Dependencies:** `log`

### Step 8.3: Config Validation ✅
- ✅ Validate TOML files on load
- ✅ Check for duplicate entries (packages, apps, profiles)
- ✅ Validate category references
- ✅ Warn about missing metadata
- ✅ Validate script existence
- ✅ Validate link sources
- ✅ Add `merlin validate` command with --strict flag
- **Test:** Catches common TOML errors ✅

**Dependencies:** None

### Step 8.4: Profiles Support ✅
- ✅ Parse `[[profile]]` sections from root `merlin.toml`
- ✅ Each profile specifies subset of tools
- ✅ Hostname-based matching available
- ✅ Add `merlin link --profile <name>`
- ✅ Add `merlin list profiles` command with auto-detect indicator
- ✅ Profile filtering logic in link command
- **Test:** Profiles correctly filter tools ✅

**Dependencies:** None (uses Phase 2 parser)

**Status:** Phase 8 complete! All advanced features implemented and tested.

---

## Phase 9: Documentation & Polish

**Goal:** Make Merlin production-ready

### Step 9.1: Help Text & Examples
 ✅ Detailed help text added to all commands (root, install, link, unlink, list, run, validate, doctor, tui)
 ✅ Examples embedded consistently
 ✅ Created `docs/USAGE.md`
 **Test:** Manual inspection for clarity & consistency (PASS)

**Dependencies:** None

### Step 9.2: Error Handling
 ### Phase 9.2: Error Handling ✅
 ✅ Central colorized output helpers (`internal/cli/format.go`)
 ✅ Replaced raw stderr prints in commands with formatted output
 ✅ Suggestions retained (e.g., doctor recommends Homebrew install)
 ✅ Colored errors/warnings via ANSI codes
 **Test:** Triggered sample errors (missing tool, missing brew.toml) – formatted output (PASS)

**Dependencies:** lipgloss (already added)

### Step 9.3: README & Install Instructions
 ### Phase 9.3: README & Install Instructions ✅
 ✅ README reorganized: Features, Quick Start, Documentation refs, Safety, Roadmap note
 ✅ Added `go install` instructions
 ☐ Screenshot reference retained (existing image); may add more later
 **Test:** Walkthrough of install & basic commands from README (PASS)

**Dependencies:** None

### Phase 9.4: USAGE Guide ✅
✅ Added comprehensive `docs/USAGE.md` (workflows, profiles, linking, validation, doctor, TUI, logging, troubleshooting)
**Test:** Cross-checked commands & flags for accuracy (PASS)

### Phase 9.5: Colored Output Integration ✅
✅ All primary commands now use `cli.Error`, `cli.Warning`, `cli.Success`
✅ Validation summary retains structured formatting
**Test:** Build + tests PASS

### Phase 9.6: Script TUI Placeholder ✅
✅ Help text and USAGE.md clarify TUI scripts flow is pending
**Test:** Visibility confirmed in `tui` command help and USAGE guide

---

## Phase 10: TUI Scripts & Orchestration ✅

**Goal:** Complete the final TUI placeholder with full interactive script execution flow including tagging, selection, progress tracking, and error handling.

**Status:** COMPLETE (November 8, 2025)

### Step 10.1: Script Tagging Model ✅
- Extend `ScriptItem` model to support optional tags
- Add custom TOML unmarshaller for backward compatibility:
  - Plain string: `"script.sh"`
  - Tagged: `{ file = "script.sh", tags = ["tag1", "tag2"] }`
  - Alternate key: `{ name = "script.sh", tags = [...] }`
- Add helper methods: `HasScriptTag()`, `FilterScriptsByTag()`
- Update `internal/scripts/runner.go` to use `ScriptItem.File`
- Update `cmd/validate.go` for ScriptItem validation
- **Test:** Parser tests for tagged scripts + backward compatibility (PASS)
- **Files:** `internal/models/merlin_tool.go`, `internal/parser/parser_test.go`, `internal/models/models_test.go`

**Dependencies:** None

### Step 10.2: Update Specification ✅
- Document script tagging syntax in `MERLIN_TOML_SPEC.md`
- Add examples showing plain and tagged script formats
- Explain tag usage (organization, filtering potential)
- Update configuration reference section
- **Test:** Documentation reviewed for accuracy
- **Files:** `docs/MERLIN_TOML_SPEC.md`

**Dependencies:** Step 10.1

### Step 10.3: Tool Script Selector UI ✅
- Create `ToolScriptSelectorModel` in `internal/tui/scripts.go`
- List all tools with scripts (show script count)
- Display tool descriptions
- Navigate and select tool
- **Test:** Model compiles, UI renders correctly
- **Files:** `internal/tui/scripts.go`

**Dependencies:** Step 10.1

### Step 10.4: Script Multi-Selector UI ✅
- Create `ScriptSelectorModel` in `internal/tui/scripts.go`
- Multi-select interface with checkboxes
- All scripts pre-selected by default
- Display tags with color coding `[tag1, tag2]`
- Keyboard shortcuts: space (toggle), `a` (all), `n` (none)
- Show selection count
- **Test:** Model compiles, selection logic works
- **Files:** `internal/tui/scripts.go`

**Dependencies:** Step 10.3

### Step 10.5: Script Execution Progress UI ✅
- Create `ScriptRunnerModel` in `internal/tui/script_runner.go`
- Real-time execution progress display
- Status indicators:
  - ⏳ Pending
  - ▶ Running
  - ✓ Success (with duration)
  - ✗ Failed (with error details)
- Progress counter (N/M scripts)
- Success/failure summary at completion
- Error message display inline
- **Test:** Model compiles, async execution works
- **Files:** `internal/tui/script_runner.go`

**Dependencies:** Step 10.4

### Step 10.6: Flow Integration ✅
- Create `LaunchScriptRunner()` in `internal/tui/flows.go`
- Tool discovery and filtering (scripts only)
- Sequential UI transitions (tool → scripts → execution)
- Script runner initialization with proper environment
- Failure summary after execution
- Wire into `cmd/tui.go` replacing placeholder
- **Test:** End-to-end flow from TUI menu to script completion (PASS)
- **Files:** `internal/tui/flows.go`, `cmd/tui.go`

**Dependencies:** Step 10.5

### Step 10.7: Logging Enrichment ✅
- Add structured logging to `internal/scripts/runner.go`
- Log script start with path
- Log dry-run mode execution
- Log completion with duration
- Log failures with exit code and error
- All logs written to `~/.merlin/merlin.log`
- **Test:** Verify log entries created during execution
- **Files:** `internal/scripts/runner.go`

**Dependencies:** Step 10.6

### Step 10.8: Error Handling & UX ✅
- Failed scripts displayed inline with error messages
- Post-execution summary lists all failures
- Error details preserved in execution model
- Helper methods: `HasFailures()`, `GetFailedScripts()`
- **Test:** Simulate script failure, verify UI shows errors
- **Files:** `internal/tui/script_runner.go`, `internal/tui/flows.go`

**Dependencies:** Step 10.7

### Step 10.9: Documentation Updates ✅
- Update `README.md`: TUI features, script tagging examples, roadmap notes
- Update `USAGE.md`: Add "Scripts Flow" section with tag examples
- Update `cmd/tui.go`: Remove "coming soon" placeholder from help text
- Create `PHASE_10_SUMMARY.md`: Complete feature summary
- **Test:** Documentation reviewed, examples accurate
- **Files:** `README.md`, `docs/USAGE.md`, `cmd/tui.go`, `docs/PHASE_10_SUMMARY.md`

**Dependencies:** Step 10.8

### Step 10.10: Quality Gates ✅
- Build: `go build` (SUCCESS)
- Tests: `go test ./...` (ALL PASS)
- Backward compatibility verified
- No compile errors or linter warnings
- **Test:** Full test suite execution (PASS)

**Dependencies:** Step 10.9

### Success Criteria (All Met ✅)
- ✅ Script tags model implemented with backward compatibility
- ✅ Full TUI script selection flow working
- ✅ Real-time progress UI with status indicators
- ✅ Batch runner executes scripts sequentially
- ✅ Error handling shows failures with details
- ✅ Logging enrichment captures script lifecycle
- ✅ Documentation updated (README, USAGE, spec)
- ✅ Tests pass, build succeeds
- ✅ No breaking changes

### Out of Scope (Deferred)
- Script retry mechanism (requires additional UI state)
- Tag-based filtering in CLI (non-interactive selection)
- Parallel script execution (complexity vs benefit)
- Script dependencies (requires dependency graph)

**Dependencies:** Phase 9

---

## Phase 11: Backup & Restore

**Goal:** Add backup/restore functionality to protect existing configurations and enable safe rollback.

**Status:** ✅ COMPLETED (November 8, 2025)

### Step 11.1: Backup Data Model
- Create `internal/backup/` package
- Define `BackupManifest` structure (metadata, timestamp, files)
- Define `BackupEntry` (original path, backup path, size, checksum)
- Backup location: `~/.merlin/backups/<timestamp>/`
- JSON manifest format for easy inspection
- **Test:** Structures compile, JSON serialization works
- **Files:** `internal/backup/backup.go`

**Dependencies:** None

### Step 11.2: Backup Implementation
- Implement `CreateBackup(files []string, reason string)` function
- Copy files to backup directory preserving structure
- Generate manifest with metadata
- Calculate checksums for integrity
- Return backup ID (timestamp)
- **Test:** Backup creation with various file types
- **Files:** `internal/backup/backup.go`

**Dependencies:** Step 11.1

### Step 11.3: Restore Implementation
- Implement `ListBackups()` function (discover available backups)
- Implement `GetBackupInfo(backupID)` (read manifest)
- Implement `RestoreBackup(backupID, selective []string)` function
- Verify checksums before restore
- Support full and selective restore
- **Test:** Restore backed up files correctly
- **Files:** `internal/backup/backup.go`

**Dependencies:** Step 11.2

### Step 11.4: Backup Commands
- Add `merlin backup create --reason "description"` command
- Add `merlin backup list` command (show all backups)
- Add `merlin backup show <id>` command (manifest details)
- Add `merlin backup restore <id>` command
- Add `merlin backup restore <id> --files file1,file2` (selective)
- Add `merlin backup clean` command (remove old backups)
- **Test:** All commands work, output is clear
- **Files:** `cmd/backup.go`

**Dependencies:** Step 11.3

### Step 11.5: Integration with Link
- Modify backup strategy in `link` command to use backup system
- Auto-backup before overwriting with `--strategy backup`
- Store backup ID in link operation log
- Add `--restore-backup <id>` flag to `merlin unlink`
- **Test:** Linking with backup strategy creates backups
- **Files:** `internal/symlink/linker.go`, `cmd/link.go`

**Dependencies:** Step 11.4

### Step 11.6: TUI Backup Management
- Add backup/restore option to TUI main menu
- Show list of backups with details (date, reason, file count)
- Allow restore selection from TUI
- Show restore confirmation with affected files
- **Test:** TUI navigation and restore flow
- **Files:** `internal/tui/backup.go`, `internal/tui/menu.go`, `internal/tui/flows.go`

**Dependencies:** Step 11.5

### Step 11.7: Documentation
- Update `USAGE.md` with backup/restore workflows
- Update `README.md` feature list
- Add backup examples to documentation
- **Test:** Documentation reviewed
- **Files:** `docs/USAGE.md`, `README.md`

**Dependencies:** Step 11.6

### Step 11.8: Testing & Quality
- Unit tests for backup/restore operations
- Test checksum validation
- Test selective restore
- Integration test: link → backup → unlink → restore
- **Test:** All tests pass
- **Files:** `internal/backup/backup_test.go`

**Dependencies:** Step 11.7

### Success Criteria
- [x] Backups created automatically when using backup strategy
- [x] Manual backup creation via CLI
- [x] List and inspect available backups
- [x] Full and selective restore functionality
- [x] Checksums verify backup integrity
- [x] TUI integration for backup management
- [x] Old backup cleanup capability
- [x] Comprehensive documentation

**Dependencies:** Phase 10

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

