# Merlin

CLI for managing macOS dotfiles: install packages & apps, link configs, run scripts.

<img width="1400" height="700" alt="image" src="https://github.com/user-attachments/assets/fadd014d-03d6-4731-992b-83da7d63ab3e" />

## Features

- Homebrew packages (formulae & casks): list, interactive or bulk install
- Mac App Store apps: list, interactive or bulk install (requires signed-in App Store)
- Native symlinking with conflict strategies (skip, backup, overwrite)
- Link all tools or a single tool
- Safe unlink (only removes symlinks pointing to the dotfiles repo)
- Tool scripts: validate and run in ordered sequence (or via link --run-scripts)
- Variable expansion from root `merlin.toml` (e.g. `{home_dir}`, `{config_dir}`)
- Dry-run & verbose flags
- System doctor (checks prerequisites)

## Commands

```
merlin doctor                 # System check
merlin list                   # Overview (brew, mas, configs)
merlin list brew|mas|configs  # Filtered lists
merlin install brew|mas       # Install (interactive unless --all)
merlin link <tool>            # Link one tool
merlin link --all             # Link all
merlin link <tool> --strategy backup --run-scripts
merlin unlink <tool>|--all    # Remove symlinks
merlin run <tool>             # Run tool scripts only
```

Flags: `--dry-run`, `--verbose` (global), plus command‑specific ones (`--all`, `--formulae-only`, `--casks-only`, `--strategy`, `--run-scripts`).

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

## Build & Test

```
go build -o merlin
go test ./...
```