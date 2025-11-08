# Merlin

CLI for managing macOS dotfiles: install packages & apps, link configs, run scripts.

## Note! This repo is in active development, and untested!

<img width="1400" height="700" alt="image" src="https://github.com/user-attachments/assets/fadd014d-03d6-4731-992b-83da7d63ab3e" />

## Features

Core:
- Interactive TUI (Bubble Tea) for installs & dotfiles management
- Homebrew packages (formulae & casks): list, interactive or bulk install
- Mac App Store apps: list, interactive or bulk install (requires signed-in App Store)
- Native symlinking with conflict strategies: skip / backup / overwrite
- Safe unlink (only removes symlinks pointing to the repo)
- Tool scripts: validate & run (or automatically via `link --run-scripts`)
- Drift inspection with `merlin diff` (packages, symlinks, scripts)
- Variable expansion from root `merlin.toml` (`{home_dir}`, `{config_dir}`)

Advanced:
- Profile support for per-machine setups
- Config validation (syntax, duplicates, broken links, missing scripts)
- Backup & restore system with checksums and integrity verification
- Symlink divergence detection (content hashing) for audit
- Optional Git auto-commit for link & backup operations (`auto_commit` setting)
- Logging to `~/.merlin/merlin.log` (enable with `--verbose`)
- Dry-run & verbose flags everywhere
- System doctor for environment checks

See `docs/USAGE.md` for detailed workflows and examples.

## Commands

```
merlin                        # Launch interactive TUI (default)
merlin tui                    # Launch interactive TUI (explicit)
merlin doctor                 # System check
merlin validate               # Validate TOML configs
merlin list                   # Overview (brew, mas, configs)
merlin list brew|mas|configs  # Filtered lists
merlin list profiles          # Show defined profiles
merlin install brew|mas       # Install (interactive unless --all)
merlin link <tool>            # Link one tool
merlin link --all             # Link all
merlin link --profile <name>  # Link tools in profile
merlin link <tool> --strategy backup --run-scripts
merlin unlink <tool>|--all    # Remove symlinks
merlin run <tool>             # Run tool scripts only
merlin backup create <files...> --reason "description"  # Create backup
merlin backup list             # List all backups
merlin backup restore <id>     # Restore backup
merlin backup clean --keep 5   # Clean old backups
merlin diff                    # Show drift (use --json, --packages, --configs, --scripts)
```

Flags: `--dry-run`, `--verbose` (global), plus command‑specific ones (`--all`, `--formulae-only`, `--casks-only`, `--strategy`, `--run-scripts`, `--profile`, `--strict`).

### Interactive TUI

Running `merlin` (or `merlin tui`) launches an interactive interface where you can:
- Browse and select packages to install with checkboxes
- Manage dotfiles (link/unlink configs)
- View and restore configuration backups
- Run tool scripts with multi-select and real-time progress tracking
- Check system prerequisites

Navigate with arrow keys or vim keys (j/k), select with space, confirm with enter.

The scripts flow now includes:
- Tool selection from all tools with scripts
- Multi-select individual scripts (supports tagged scripts for organization)
- Real-time execution progress with status indicators (⏳ Pending, ▶ Running, ✓ Success, ✗ Failed)
- Execution timing and error details
- Summary of successes and failures

## Dotfiles Structure (expected)

```
<dotfiles>/
├── merlin.toml          # Root settings & variables
└── config/
    └── <tool>/
        ├── merlin.toml  # Tool config (links, scripts, deps)
        ├── config/      # Files/directories to link
        └── scripts/     # Optional scripts
```

See the example repo: [covenant](../covenant) or docs:
- `docs/DOTFILES_STRUCTURE.md`
- `docs/MERLIN_TOML_SPEC.md`

## Example Link Definition

```toml
[[link]]
source = "config/omp.toml"
target = "{config_dir}/zsh/omp.toml"
```

Directory link (implicit source of all contents):
```toml
[[link]]
target = "{config_dir}/zellij"
```

Multiple files to one directory:
```toml
[[link]]
target = "{home_dir}/Library/Application Support/Cursor/User"
files = [
  { source = "config/settings.json", target = "settings.json" },
  { source = "config/keybindings.json", target = "keybindings.json" }
]
```

## Safety & Conflict Handling

- Only creates/removes symlinks referring to the tool's source path
- Strategies: skip (default), backup (rename original), overwrite
- Already-linked detection

## Git Auto-Commit (Optional)

Enable lightweight auditing of repository state changes by setting in your root `merlin.toml`. Suppress on a per-run basis with `--no-auto-commit`.

```toml
[settings]
auto_commit = true
```

Behavior:
- After successful `merlin link` operations, Merlin stages only the affected tool directories (`config/<tool>`) and commits with a message like (or creates an empty commit if nothing changed):
  - `chore(link): link zsh`
  - `chore(link): link zsh, git (2 tools)`
  - `chore(link): link 5 tools (zsh, git, eza, …)`
- After `merlin backup create`, Merlin appends a summary entry to `.merlin-meta/backups.json` inside the repo and commits just that file: `chore(backup): record <id> (<n> files)` (empty commit fallback).
- After `merlin unlink`, Merlin records removal actions similarly:
  - `chore(unlink): unlink zsh`
  - `chore(unlink): unlink zsh, git (2 tools)`
  - `chore(unlink): unlink 4 tools (zsh, git, eza, …)`

Design Principles:
- Minimal scope: stages only tool config directories or the backup index file—never auto-add your whole repository.
- Non-intrusive: skipped if Git isn’t available, the directory isn’t a Git repo, or `auto_commit = false`.
- Empty commits allowed: if no paths changed but you requested an operation, Merlin can create an empty commit (audit trail) unless unrelated changes would pollute scope.

Disable if you prefer manual commit control; re-enable anytime. Prefixes (`chore(link):`, `chore(backup):`) let you filter history easily.

Safety Model:
- Whitelisted staging paths only (tool config dirs or backup index file)
- Unrelated changes detection: auto-commit skipped if unstaged/untracked items exist outside whitelisted paths
- Per-run suppression: `--no-auto-commit`
- Graceful skip when git is absent or directory not a repo

Examples:

| Operation | Commit Example | Override Flag |
|-----------|----------------|---------------|
| link single | `chore(link): link zsh` | `--no-auto-commit` |
| link batch (>3) | `chore(link): link 5 tools (zsh, git, eza, …)` | `--no-auto-commit` |
| unlink batch | `chore(unlink): unlink 4 tools (zsh, git, eza, …)` | `--no-auto-commit` |
| backup create | `chore(backup): record 20251108_120301 (3 files)` | `--no-auto-commit` |

Planned Extensions:
- Commit hooks for export/reconcile operations.
- Optional squash mode for initial provisioning.

## Scripts

Defined under `[scripts]` in a tool `merlin.toml`; run with `merlin run <tool>` or after linking using `--run-scripts`, or interactively through the TUI.

Scripts can now include optional tags for organization:

```toml
[scripts]
directory = "scripts"
scripts = [
  "setup.sh",                                      # Plain string
  { file = "install.sh", tags = ["full"] },       # Tagged script
  { file = "dev_tools.sh", tags = ["dev", "optional"] }
]
```

Tags are displayed in the TUI selector and logged, helping you identify script purposes at a glance.

### Diff & Drift Detection

Use `merlin diff` to compare the current system against your declarative repo:

```
merlin diff                 # Full report
merlin diff --packages      # Only Homebrew + MAS differences
merlin diff --configs       # Symlink status (missing, orphaned, broken, divergent)
merlin diff --scripts       # Script presence (namespaced as tool/script)
merlin diff --json          # Machine-readable JSON
```

Symlink categories:
- Missing: declared link not present
- Orphaned: symlink points into repo but not declared
- Broken: symlink target path does not exist
- Divergent: symlink points to file whose content hash differs from declared source

Script categories:
- Added: script file exists but not declared in `[scripts]`
- Missing: declared script not found on disk

JSON schema keys: `brew_formulae`, `brew_casks`, `mas_apps`, `symlinks`, `scripts`.

## Advanced Audit & Automation Roadmap

Recent additions (Phase 12 & 13): drift detection, divergence hashing, script presence diff, and auto-commit hooks. Upcoming plans include:
- `merlin clone <repo>`: bootstrap remote repo & perform initial link/install.
- Uninstall commands for declaratively removing packages.
- Update checks & export tooling (`merlin export` to snapshot current system as TOML).
- Optional reconciliation commands to resolve Missing/Divergent items.

## Quick Start

Install latest via Go:

```
go install github.com/ildx/merlin@latest
```

Or build locally:

```
go build -o merlin
./merlin
```

## Documentation

- Usage guide: `docs/USAGE.md`
- Dotfiles structure: `docs/DOTFILES_STRUCTURE.md`
- TOML specification: `docs/MERLIN_TOML_SPEC.md`
- Implementation roadmap: `docs/IMPLEMENTATION_PLAN.md`

## Safety & Philosophy

- Operates only within your declared dotfiles repository
- Won't delete regular files when unlinking (symlinks only)
- Idempotent operations: safe to repeat installs & links
- Profiles let you tailor minimal setups per host

## Roadmap Notes

Implemented Phase 10 features:
- ✅ Script tags for organization
- ✅ Full interactive TUI scripts flow with selection and execution
- ✅ Real-time progress tracking with status indicators
- ✅ Enhanced logging with per-script timing

Future enhancements may include script retry mechanisms and tag-based CLI filtering.
Diff phase adds future potential for auto-reconciliation and export tooling.

## Build & Test

```
go build -o merlin
go test ./...
```

## License

See `LICENSE`.