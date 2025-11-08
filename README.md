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
- Variable expansion from root `merlin.toml` (`{home_dir}`, `{config_dir}`)

Advanced:
- Profile support for per-machine setups
- Config validation (syntax, duplicates, broken links, missing scripts)
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
```

Flags: `--dry-run`, `--verbose` (global), plus command‑specific ones (`--all`, `--formulae-only`, `--casks-only`, `--strategy`, `--run-scripts`, `--profile`, `--strict`).

### Interactive TUI

Running `merlin` (or `merlin tui`) launches an interactive interface where you can:
- Browse and select packages to install with checkboxes
- Manage dotfiles (link/unlink configs)
- Run tool scripts
- Check system prerequisites

Navigate with arrow keys or vim keys (j/k), select with space, confirm with enter.

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

## Scripts

Defined under `[scripts]` in a tool `merlin.toml`; run with `merlin run <tool>` or after linking using `--run-scripts`.

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

Script flow in TUI is currently a placeholder; use `merlin run <tool>`.

## Build & Test

```
go build -o merlin
go test ./...
```

## License

See `LICENSE`.