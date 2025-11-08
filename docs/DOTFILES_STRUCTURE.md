# Dotfiles Repository Structure

This document describes the expected structure for dotfiles repositories that work with Merlin.

---

## Overview

Merlin is a **macOS-focused** tool designed to work with any dotfiles repository following a consistent, organized structure. Your dotfiles repo doesn't need to be named "covenant" or follow any specific naming convention—just the directory structure and file formats described below.

---

## Required Structure

```
your-dotfiles/
├── merlin.toml              # Root config: settings, variables, profiles
└── config/
    └── <tool-name>/
        ├── config/          # Files to symlink
        ├── merlin.toml      # Tool instructions (optional)
        ├── scripts/         # Setup scripts (optional)
        └── *.toml           # Tool-specific data (optional)
```

---

## Directory Structure

### Root Level

**`merlin.toml`** (required)
- Global settings
- Directory variables (`home_dir`, `config_dir`)
- Preinstall tools
- Profiles

### Tool Level

**`config/<tool>/config/`** (required)
- Actual config files to symlink
- Symlinked to `~/.config/<tool>/` by default
- Or custom paths via `merlin.toml`

**`config/<tool>/merlin.toml`** (optional)
- Tool-specific instructions
- Dependencies
- Custom symlink paths
- Scripts to run
- Only needed if tool requires special handling

**`config/<tool>/scripts/`** (optional)
- Custom setup scripts
- Executed after symlinking
- For extension installs, config generation, etc.

**`config/<tool>/*.toml`** (optional)
- Tool-specific data files
- Read by scripts, not by Merlin
- Examples: `brew.toml`, `mas.toml`, `profiles.toml`

---

## Examples

### Simple Tool (eza)

```
config/eza/
├── config/
│   ├── eza.zsh
│   └── theme.yml
└── merlin.toml
```

**`config/eza/merlin.toml`:**
```toml
[tool]
name = "eza"
description = "Modern ls replacement"
dependencies = ["brew"]

[[link]]
target = "{config_dir}/eza"
```

**Result:** `config/eza/config/*` → `~/.config/eza/*`

---

### Multiple Symlink Targets (git)

```
config/git/
├── config/
│   ├── .gitconfig
│   └── profiles/
│       ├── personal
│       └── work
└── merlin.toml
```

**`config/git/merlin.toml`:**
```toml
[tool]
name = "git"
description = "Git version control"
dependencies = []

[[link]]
target = "{config_dir}/git"

[[link]]
source = "config/.gitconfig"
target = "{home_dir}/.gitconfig"
```

**Result:**
- `config/git/config/profiles/*` → `~/.config/git/profiles/*`
- `config/git/config/.gitconfig` → `~/.gitconfig`

---

### With Scripts (cursor)

```
config/cursor/
├── config/
│   ├── settings.json
│   ├── keybindings.json
│   └── extensions.txt
├── scripts/
│   └── install_extensions.sh
└── merlin.toml
```

**`config/cursor/merlin.toml`:**
```toml
[tool]
name = "cursor"
description = "AI code editor"
dependencies = ["brew"]

[[link]]
target = "{config_dir}/cursor"

[[link]]
target = "{home_dir}/Library/Application Support/Cursor/User"
files = [
  { source = "config/settings.json", target = "settings.json" },
  { source = "config/keybindings.json", target = "keybindings.json" }
]

[scripts]
directory = "scripts"
scripts = ["install_extensions.sh"]
```

**Result:**
1. Config files symlinked
2. Script installs extensions from `extensions.txt`

---

### With Tool-Specific Data (karabiner)

```
config/karabiner/
├── config/
│   ├── assets/
│   │   └── complex_modifications/
│   │       └── shared.json
│   └── karabiner.json (generated)
├── scripts/
│   └── generate_config.sh
├── merlin.toml
└── profiles.toml
```

**`config/karabiner/merlin.toml`:**
```toml
[tool]
name = "karabiner"
description = "Keyboard customizer"
dependencies = ["brew", "yq", "jq"]

[[link]]
target = "{config_dir}/karabiner"

[scripts]
directory = "scripts"
scripts = ["generate_config.sh"]
```

**`config/karabiner/profiles.toml`:**
```toml
[[profiles]]
name = "Personal"
hostname = "iivo"
selected = true
rules = ["shared"]

[[profiles]]
name = "Work"
hostname = "uppis"
selected = false
rules = ["shared"]
```

**Result:**
1. Config directory symlinked
2. Script reads `profiles.toml` and generates `karabiner.json`

---

## Package Definitions

### brew.toml

Defines Homebrew packages (formulae and casks).

**Location:** `config/brew/config/brew.toml`

**Format:**
```toml
[metadata]
version = "1.0.0"
description = "Homebrew configuration"

[[brew]]
name = "bat"
description = "Better cat with syntax highlighting"
category = "cli"
dependencies = []

[[brew]]
name = "eza"
description = "Modern ls replacement"
category = "cli"
dependencies = []

[[cask]]
name = "cursor"
description = "AI code editor"
category = "development"
dependencies = []

[categories]
cli = { display_name = "CLI Tools", icon = "🔧", order = 1 }
development = { display_name = "Development", icon = "💻", order = 2 }
```

### mas.toml

Defines Mac App Store applications.

**Location:** `config/mas/config/mas.toml`

**Format:**
```toml
[metadata]
version = "1.0.0"
description = "Mac App Store applications"

[[app]]
name = "Amphetamine"
id = 937984704
description = "Keep Mac awake"
category = "productivity"
dependencies = []

[categories]
productivity = { display_name = "Productivity", icon = "⚡", order = 1 }
```

**Finding App IDs:**
```bash
mas search "App Name"
mas list
```

---

## Variables

Paths use variables for flexibility:

**Root `merlin.toml`:**
```toml
[settings]
home_dir = "~"
config_dir = "{home_dir}/.config"
```

**Tool configs use variables:**
```toml
[[link]]
target = "{config_dir}/tool"

[[link]]
source = "config/.zshrc"
target = "{home_dir}/.zshrc"
```

**User can override:**
```bash
merlin link tool --config-dir ~/.dotfiles
```

---

## Link Patterns

### Pattern 1: Directory (implicit source)
```toml
[[link]]
target = "{config_dir}/tool"
# Symlinks: config/tool/config/* → ~/.config/tool/*
```

### Pattern 2: File (explicit source)
```toml
[[link]]
source = "config/<file>"
target = "{home_dir}/<file>"
# Symlinks: config/tool/config/<file> → ~/<file>
```

### Pattern 3: Multiple files to base
```toml
[[link]]
target = "{home_dir}/Library/Application Support/<App>/User"
files = [
  { source = "config/<file>", target = "<file>" }
]
# Each file in array symlinked to base target
```

### Pattern 4: Directory to directory
```toml
[[link]]
source = "config/<directory>"
target = "/Applications/<App>.app/Resources/<directory>"
# Directory contents symlinked
```

---

## Script Guidelines

### Location
```
config/tool/
└── scripts/
    ├── setup.sh
    └── install_extensions.sh
```

### Requirements
- Must be executable (`chmod +x`)
- Should be idempotent (safe to run multiple times)
- Should provide clear output
- Exit with non-zero on failure

### Example Script
```bash
#!/bin/bash
set -e

echo "🔧 Installing extensions..."

# Check if tool is installed
if ! command -v tool &> /dev/null; then
    echo "❌ Error: tool not installed"
    exit 1
fi

# Do the work
tool --install-extension ext1
tool --install-extension ext2

echo "✅ Done!"
```

### When to Use Scripts

**Use symlinking for:**
- Config files that just need to exist
- Files that don't need processing

**Use scripts for:**
- Installing extensions
- Generating configs
- Complex setup logic
- Checking prerequisites

---

## Complete Example Repository

```
my-dotfiles/
├── merlin.toml                       # Root config
└── config/
    ├── brew/
    │   ├── config/
    │   │   ├── brew.toml
    │   │   └── brew.zsh
    │   └── merlin.toml
    ├── cursor/
    │   ├── config/
    │   │   ├── settings.json
    │   │   ├── keybindings.json
    │   │   ├── cursor.zsh
    │   │   └── extensions.txt
    │   ├── scripts/
    │   │   └── install_extensions.sh
    │   └── merlin.toml
    ├── eza/
    │   ├── config/
    │   │   ├── eza.zsh
    │   │   └── theme.yml
    │   └── merlin.toml
    ├── git/
    │   ├── config/
    │   │   ├── .gitconfig
    │   │   └── profiles/
    │   │       ├── personal
    │   │       └── work
    │   └── merlin.toml
    ├── karabiner/
    │   ├── config/
    │   │   ├── assets/
    │   │   │   └── complex_modifications/
    │   │   │       └── shared.json
    │   │   └── karabiner.json
    │   ├── scripts/
    │   │   └── generate_config.sh
    │   ├── merlin.toml
    │   └── profiles.toml
    ├── mas/
    │   ├── config/
    │   │   ├── mas.toml
    │   │   └── mas.zsh
    │   └── merlin.toml
    ├── misc/
    │   ├── config/
    │   │   └── .hushlogin
    │   └── merlin.toml
    └── zsh/
        ├── config/
        │   ├── .zshrc
        │   ├── defaults/
        │   │   ├── alias.zsh
        │   │   ├── color.zsh
        │   │   └── plugins.zsh
        │   └── omp.toml
        └── merlin.toml
```

---

## Best Practices

### 1. Keep configs modular
Each tool should be self-contained. Don't mix unrelated configs.

### 2. Use variables
Always use `{config_dir}` and `{home_dir}` instead of hardcoded paths.

### 3. Prefer symlinking
Only create scripts when necessary. Native symlinking is simpler.

### 4. Document scripts
Add comments explaining what complex scripts do.

### 5. Test on fresh system
Regularly test your dotfiles on a clean macOS installation.

### 6. Use dependencies
Declare dependencies in `merlin.toml` so Merlin installs in correct order.

```toml
[tool]
name = "cursor"
dependencies = ["brew"]  # Cursor needs Homebrew first
```

### 7. Separate concerns
- `merlin.toml` = Instructions TO Merlin
- `*.toml` = Data FOR scripts
- `scripts/` = Custom setup logic

---

## Migration from Other Systems

### From GNU Stow

1. Keep existing structure (mostly compatible)
2. Ensure each tool follows `config/<tool>/config/` pattern
3. Add `merlin.toml` at root with settings
4. Add per-tool `merlin.toml` files for custom paths
5. Use variables in targets
6. Run `merlin link <tool>` instead of `stow <tool>`

### From Shell Scripts

1. Extract package lists into `brew.toml` / `mas.toml`
2. Move configs into `config/<tool>/config/` directories
3. Create `merlin.toml` files to define symlinks
4. Keep only complex logic in scripts (move to `scripts/` subdirs)
5. Let Merlin handle symlinking
6. Add dependencies to tool configs

---

## Validation

Verify your structure:

```bash
# Check configuration
merlin validate

# List available tools
merlin list

# Show what would be linked
merlin link <tool> --dry-run

# Check link status
merlin status
```

---

## Common Questions

**Q: Can I use a different name than `config/`?**  
A: Not currently. Merlin expects `config/` as the root directory.

**Q: Do I need both `brew.toml` and `mas.toml`?**  
A: No, include only what you use. Merlin gracefully handles missing files.

**Q: Can I have multiple scripts per tool?**  
A: Yes! Define them in the `scripts` array:
```toml
[scripts]
directory = "scripts"
scripts = ["setup.sh", "install_extensions.sh", "post_setup.sh"]
```

**Q: What if my tool doesn't fit the structure?**  
A: Create a `misc/` directory for one-off configs.

**Q: Where do tool-specific data files go?**  
A: At tool root: `config/tool/data.toml` (NOT in `config/` subdir)

**Q: How do I share configs between profiles?**  
A: Tools listed in a profile get linked. Use different profiles for different machines.

---

**Next:** See [MERLIN_TOML_SPEC.md](./MERLIN_TOML_SPEC.md) for complete configuration reference.
