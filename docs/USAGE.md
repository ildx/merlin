# Merlin Usage Guide

A practical guide to installing packages, managing dotfiles, running scripts, and using profiles with Merlin.

---
## Quick Start

```bash
# Install (from repository root or anywhere with Go installed)
go install github.com/ildx/merlin@latest

# Or build locally
cd merlin
go build -o merlin
./merlin
```

Merlin will launch the interactive TUI by default if no subcommand is provided.

---
## Global Flags

These flags work on most commands:

- `--dry-run`  Preview actions without making changes
- `--verbose` / `-v`  More detailed output and debug logging (written to `~/.merlin/merlin.log`)

You can combine them with subcommands:

```bash
merlin link zsh --dry-run
merlin install brew --all --verbose
```

---
## Dotfiles Structure Overview

Merlin expects a dotfiles repository with this structure:

```
<dotfiles>/
├── merlin.toml          # Root settings, profiles, variables
└── config/
    └── <tool>/
        ├── merlin.toml  # Tool config (links, scripts, dependencies)
        ├── config/      # Files/directories to link
        └── scripts/     # Optional scripts
```

Variable placeholders like `{home_dir}` and `{config_dir}` are expanded in link targets.

---
## Installing Packages

### Homebrew
Install formulae and casks defined in `config/brew/config/brew.toml`.

```bash
# Interactive picker
merlin install brew

# Install all packages
merlin install brew --all

# Only formulae or only casks
merlin install brew --formulae-only
merlin install brew --casks-only

# Preview without installing
merlin install brew --all --dry-run
```

Already-installed items are skipped. Use `merlin list brew` to inspect package definitions.

### Mac App Store (MAS)
Install apps from `config/mas/config/mas.toml`.

```bash
merlin install mas          # Interactive picker
merlin install mas --all    # Install all apps
merlin install mas --dry-run --all
```

You must be signed into the App Store and have `mas` CLI installed.

---
## Listing Resources

```bash
merlin list                 # Overview (brew, mas, configs)
merlin list brew            # Homebrew packages
merlin list mas             # App Store apps
merlin list configs         # Config tools
merlin list profiles        # Profiles from root merlin.toml
```

Filter Homebrew or MAS by category:

```bash
merlin list brew -c dev
merlin list mas -c productivity
```

---
## Profiles

Profiles let you define subsets of tools and optional hostname targeting in `merlin.toml`:

```toml
[[profile]]
name = "personal"
hostname = "MacBook-Pro"
tools = ["zsh", "git", "eza", "cursor"]
```

To use a profile when linking:

```bash
merlin link --all --profile personal
```

If a profile lists no tools, all tools are used. If `hostname` matches the current machine, you can choose to adopt that profile explicitly.

List profiles:

```bash
merlin list profiles
```

---
## Linking Configurations

Create symlinks for a single tool or all tools.

```bash
# Link one tool
merlin link zsh

# Link all tools
merlin link --all

# Profile-based linking
merlin link --all --profile personal

# Preview changes
merlin link git --dry-run
```

Conflict strategies:

- `skip` (default): keep existing files
- `backup`: rename existing file to `.backup.<timestamp>`
- `overwrite`: replace existing file/symlink

Specify strategy:

```bash
merlin link eza --strategy backup
```

Run tool scripts immediately after linking if defined:

```bash
merlin link zellij --run-scripts
```

---
## Unlinking

Remove symlinks created by Merlin (safe – only those pointing into your repo):

```bash
merlin unlink git
merlin unlink --all
merlin unlink zsh --dry-run
```

---
## Scripts

Tool scripts live under `config/<tool>/scripts/` and are configured in the tool's `merlin.toml` under `[scripts]`.

Run scripts directly:

```bash
merlin run cursor
merlin run zellij --dry-run
merlin run git --verbose
```

Or run them after linking with `--run-scripts`.

Note: The dedicated scripts flow in the TUI is a placeholder for now. Use the CLI commands above.

---
## Validation

Validate configuration integrity:

```bash
merlin validate
merlin validate --strict   # Treat warnings as errors
```

Checks include: syntax errors, duplicates, invalid strategies, missing scripts, broken link definitions.

Use before linking or installing to catch issues early.

---
## System Doctor

Check prerequisites, macOS suitability, and optional tools:

```bash
merlin doctor
```

Reports Homebrew, mas-cli, and common utilities (git, curl, jq, yq). Suggests installation if missing.

---
## Interactive TUI

Launch with:

```bash
merlin
# or
merlin tui
```

Provides menus for:

- Installing packages
- Managing dotfiles (link/unlink)
- (Scripts: coming soon)
- System doctor shortcut

Navigation:

- Arrow keys / j k: move
- Space: toggle selection
- Enter: confirm
- Esc / q: back/quit

### Scripts Flow

The TUI now includes full script execution support with interactive selection:

1. From the main menu, select **Run Scripts**
2. Choose a tool that has scripts defined
3. Multi-select which scripts to run (all selected by default)
   - Space: toggle individual script
   - `a`: select all
   - `n`: select none
   - Scripts with tags are displayed with `[tag1, tag2]`
4. Watch real-time execution progress with status indicators:
   - ⏳ Pending
   - ▶ Running
   - ✓ Success (with timing)
   - ✗ Failed (with error details)
5. Review summary showing successes and failures

Scripts can now include optional tags in `merlin.toml` for better organization:

```toml
[scripts]
directory = "scripts"
scripts = [
  "setup.sh",                                 # Plain string (backward compatible)
  { file = "install.sh", tags = ["full"] },  # Tagged script
  { file = "dev_setup.sh", tags = ["dev", "optional"] }
]
```

Tags help categorize scripts for selection but don't affect execution order—all selected scripts run sequentially.

---
## Dry-Run Strategy

Use `--dry-run` anywhere to preview actions:

```bash
merlin link --all --dry-run
merlin install mas --all --dry-run
merlin run cursor --dry-run
```

Dry-run ensures no changes are made; summaries still display.

---
## Logging

Verbose mode enables debug-level logging. Logs are written to:

```
~/.merlin/merlin.log
```

Enable with `--verbose` (or `-v`).

---
## Troubleshooting

| Issue | Solution |
|-------|----------|
| Homebrew not installed | Install from https://brew.sh |
| mas-cli missing | `brew install mas` |
| Not signed into App Store | Open App Store, sign in, retry `merlin install mas` |
| Link sources missing | Run `merlin validate` to identify missing files |
| Scripts failing | Use `--verbose` to stream output and inspect errors |

---
## Exit Codes

- 0: Success
- Non-zero: Validation errors, script failures, install errors, prerequisite failures

Strict validation (`--strict`) treats warnings as errors (non-zero exit).

---
## Recommended Workflow

1. `merlin doctor` – confirm environment
2. `merlin validate` – ensure configs are clean
3. `merlin install brew|mas` – provision packages/apps
4. `merlin link --all` – apply configuration
5. `merlin run <tool>` – run any remaining scripts
6. `merlin tui` – day-to-day management

Use profiles if managing multiple machines.

---
## Future Enhancements (Planned)

- Script retry mechanism after failures
- Backup/restore and diff capabilities
- Remote repository cloning
- Tag-based script filtering in CLI

---
## Reference

For data model details:
- `docs/MERLIN_TOML_SPEC.md`
- `docs/DOTFILES_STRUCTURE.md`

---
Happy dotfiles managing! ✨
