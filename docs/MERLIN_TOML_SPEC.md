# merlin.toml Specification

This document defines the structure and behavior of Merlin's configuration files.

---

## Overview

Merlin uses a **modular configuration approach**:
- **Root `merlin.toml`** - Global settings, variables, and profiles
- **Per-tool `merlin.toml`** - Tool-specific instructions

This keeps each tool self-contained and makes the configuration more maintainable.

---

## File Locations

```
your-dotfiles/
├── merlin.toml                    # Global settings & profiles
└── config/
    ├── git/
    │   ├── config/                # Dotfiles to symlink
    │   └── merlin.toml            # Git-specific config
    ├── zsh/
    │   ├── config/
    │   └── merlin.toml
    └── cursor/
        ├── config/
        ├── scripts/
        │   └── install_extensions.sh
        └── merlin.toml            # Cursor-specific config
```

---

## Root merlin.toml

Located at the root of your dotfiles repository.

**Purpose:** Global settings, variables, and profiles.

```toml
[metadata]
name = "my-dotfiles"
version = "1.0.0"
description = "Personal dotfiles managed by Merlin"

[settings]
auto_link = false                 # Auto-link configs after package install
confirm_before_install = false    # Ask before installing packages
conflict_strategy = "backup"      # Default: backup, skip, overwrite, interactive

# Variables (can be overridden by Merlin at runtime)
home_dir = "~"
config_dir = "{home_dir}/.config"

# System requirements (installed FIRST, before profiles)
[preinstall]
tools = [
  "xcode",                # Xcode Command Line Tools
  "brew",                 # Homebrew package manager
  "mas",                  # Mac App Store CLI
  "git",                  # Git version control
  "zsh"                   # Z shell
]

# Profiles for different machines
[[profile]]
name = "full"
default = true
description = "Full setup with all tools"
tools = ["cursor", "eza", "ghostty", "karabiner", "lazygit", "misc", "mise", "zellij"]

[[profile]]
name = "work"
hostname = "Work-MacBook"
description = "Work machine setup"
tools = ["git", "zsh", "cursor", "mise"]
```

### Variables

Variables allow flexible directory configuration:

```toml
[settings]
home_dir = "~"
config_dir = "{home_dir}/.config"
```

- Variables can reference other variables
- Merlin expands them at runtime
- User can override via CLI: `merlin link --config-dir ~/.dotfiles`
- Available in all target paths: `target = "{config_dir}/tool"`

---

## Per-Tool merlin.toml

Located at `config/TOOL/merlin.toml` for each tool.

**Purpose:** Tool-specific configuration.

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

---

## Tool Configuration - Simple Case

**File:** `config/eza/merlin.toml`

For tools that just need `config/TOOL/config/*` → `~/.config/TOOL/*`:

```toml
[tool]
name = "eza"
description = "Modern ls replacement"
dependencies = ["brew"]

[[link]]
target = "{config_dir}/eza"
```

**That's it!** Merlin will symlink `config/eza/config/*` → `~/.config/eza/*`

---

## Tool Configuration - Multiple Links

For tools needing specific symlink paths:

### Example 1: git (multiple symlink targets)

**File:** `config/git/merlin.toml`

```toml
[tool]
name = "git"
description = "Git version control configuration"
dependencies = []

[[link]]
target = "{config_dir}/git"

[[link]]
source = "config/.gitconfig"
target = "{home_dir}/.gitconfig"
```

**Result:**
- `config/git/config/*` → `~/.config/git/*`
- `config/git/config/.gitconfig` → `~/.gitconfig`

### Example 2: zsh (dotfiles in home directory)

**File:** `config/zsh/merlin.toml`

```toml
[tool]
name = "zsh"
description = "Z shell configuration"
dependencies = []

[[link]]
source = "config/.zshrc"
target = "{home_dir}/.zshrc"

[[link]]
target = "{config_dir}/zsh"
```

---

## Link Patterns

All links use unified `source` + `target` keys with variable support:

### Pattern 1: Directory (implicit source)

```toml
[[link]]
target = "{config_dir}/tool"
# Implies: source = "config/"
```

### Pattern 2: File (explicit source)

```toml
[[link]]
source = "config/.zshrc"
target = "{home_dir}/.zshrc"
```

### Pattern 3: Multiple files to base

```toml
[[link]]
target = "{home_dir}/Library/Application Support/Cursor/User"
files = [
  { source = "config/settings.json", target = "settings.json" },
  { source = "config/keybindings.json", target = "keybindings.json" }
]
```

### Pattern 4: Directory to directory

```toml
[[link]]
source = "config/themes"
target = "/Applications/Ghostty.app/Contents/Resources/ghostty/themes"
```

---

## Tool Configuration - Scripts & Tags

For tools needing custom setup logic:

**File:** `config/cursor/merlin.toml`

```toml
[tool]
name = "cursor"
description = "AI code editor configuration"
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
scripts = [
  "install_extensions.sh",                      # Simple form
  { file = "post_install_cleanup.sh" },         # Extended form (no tags)
  { file = "large_setup.sh", tags = ["full"] }, # Tagged script
  { file = "dev_only.sh", tags = ["dev", "fast"] }
]
```

**Scripts are executed in order after symlinking.** Each entry can be either:

1. A plain string (backward compatible): `"script.sh"`
2. A table with a `file` (or `name`) key and optional `tags` array: `{ file = "script.sh", tags = ["tag"] }`

### Script Tags

Tags allow selective execution or filtering (e.g., in the TUI script selection flow):

```toml
[scripts]
directory = "scripts"
scripts = [
  { file = "base.sh", tags = ["core"] },
  { file = "optional_fonts.sh", tags = ["fonts", "ui"] },
  { file = "heavy_index.sh", tags = ["full","slow"] }
]
```

Potential usage scenarios:

- Skip long-running scripts by deselecting those tagged `slow`
- Run only `dev` scripts on a development machine
- Filter for `core` scripts in a minimal setup

> NOTE: Tag-based filtering is optional and surfaced primarily through interactive tooling (Phase 10 TUI). Non-interactive CLI flows continue to run all listed scripts.

---

## Tool Configuration - Dependencies

Tools can declare dependencies:

```toml
[tool]
name = "cursor"
dependencies = ["brew"]  # Requires brew to be installed first

[tool]
name = "karabiner"
dependencies = ["brew", "yq", "jq"]  # Multiple dependencies
```

Merlin resolves dependencies and installs in correct order.

---

## Tool Configuration - Tool-Specific Data

Some tools store configuration data in separate TOML files:

### Example: Karabiner profiles

**File:** `config/karabiner/merlin.toml`

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

**File:** `config/karabiner/profiles.toml`

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

The `generate_config.sh` script reads `profiles.toml` to build `karabiner.json`.

**Pattern:** Tool-specific data lives in separate `.toml` files, `merlin.toml` contains only Merlin instructions.

---

## Execution Flow

When `merlin link <tool>` is run:

1. **Resolve variables** - Expand `{home_dir}`, `{config_dir}`, etc.
2. **Check dependencies** - Verify required tools are available
3. **Create symlinks** - Process all `[[link]]` entries
4. **Run scripts** - Execute scripts in order from `[scripts]` table
5. **Handle conflicts** - Use `conflict_strategy` for existing files

---

## Profiles

Define different setups for different machines:

```toml
[[profile]]
name = "full"
default = true
description = "Full setup with all tools"
tools = ["cursor", "eza", "ghostty", "karabiner", "lazygit", "misc", "mise", "zellij"]

[[profile]]
name = "work"
hostname = "Work-MacBook"
description = "Work machine setup"
tools = ["git", "zsh", "cursor", "mise"]

[[profile]]
name = "minimal"
hostname = "Server"
tools = ["git", "zsh"]
```

**Usage:**
```bash
# Auto-selects profile based on hostname
merlin install

# Or explicitly specify
merlin install --profile work
```

---

## Preinstall Tools

System requirements installed BEFORE any profile tools:

```toml
[preinstall]
tools = [
  "xcode",       # Xcode Command Line Tools
  "brew",        # Homebrew
  "mas",         # Mac App Store CLI
  "git",         # Git
  "zsh"          # Z shell
]
```

**Purpose:** Bootstrap tools needed by everything else.

**Execution order:**
1. Preinstall tools (always installed)
2. Profile tools (based on selected profile)

---

## Complete Example

### Root merlin.toml

```toml
[metadata]
name = "covenant"
version = "1.0.0"
description = "Personal dotfiles managed by Merlin"

[settings]
auto_link = false
confirm_before_install = false
conflict_strategy = "backup"

# Variables
home_dir = "~"
config_dir = "{home_dir}/.config"

[preinstall]
tools = ["xcode", "brew", "mas", "git", "zsh"]

[[profile]]
name = "full"
default = true
tools = ["cursor", "eza", "ghostty", "karabiner", "lazygit", "misc", "mise", "zellij"]
```

### config/git/merlin.toml

```toml
[tool]
name = "git"
description = "Git version control configuration"
dependencies = []

[[link]]
target = "{config_dir}/git"

[[link]]
source = "config/.gitconfig"
target = "{home_dir}/.gitconfig"
```

### config/cursor/merlin.toml

```toml
[tool]
name = "cursor"
description = "AI code editor configuration"
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

### config/karabiner/merlin.toml

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

### config/zellij/merlin.toml

```toml
[tool]
name = "zellij"
description = "Terminal multiplexer"
dependencies = ["brew"]

[[link]]
target = "{config_dir}/zellij"

[scripts]
directory = "scripts"
scripts = ["generate_config.sh"]
```

**Note:** Tools like `eza`, `lazygit`, `mas`, `mise` use simple configs and don't need special handling.

---

## Configuration Reference

### Root merlin.toml Fields

**[metadata]**
- `name` (string) - Dotfiles repository name
- `version` (string) - Version
- `description` (string) - Description

**[settings]**
- `auto_link` (boolean, default: false) - Auto-link configs after package install
- `confirm_before_install` (boolean, default: true) - Ask before installing packages
- `conflict_strategy` (string, default: "interactive") - backup|skip|overwrite|interactive
- `home_dir` (string, default: "~") - Home directory variable
- `config_dir` (string, default: "{home_dir}/.config") - Config directory variable

**[preinstall]**
- `tools` (array of strings) - Tools to install before profiles

**[[profile]]**
- `name` (string, required) - Profile name
- `hostname` (string) - Auto-select profile if hostname matches
- `default` (boolean, default: false) - Use if no hostname match
- `description` (string) - Human-readable description
- `tools` (array of strings) - Tools to enable in this profile

### Per-Tool merlin.toml Fields

**[tool]**
- `name` (string, required) - Tool name, must match directory in `config/`
- `description` (string) - Human-readable description
- `dependencies` (array of strings) - Tools that must be installed first

**[[link]]**
- `source` (string, optional) - Path relative to `config/TOOL/` (defaults to "config/")
- `target` (string, required) - Destination path with variable support
- `files` (array, optional) - For multiple files to same base (Pattern 3)
  - `source` (string) - Source file path
  - `target` (string) - Target file name (relative to parent target)

**[scripts]**
- `directory` (string) - Directory containing scripts (relative to tool dir)
- `scripts` (array) - Scripts to execute in order. Each element may be:
  - Plain string: `"script.sh"`
  - Table: `{ file = "script.sh", tags = ["tag1", "tag2"] }`
  - Alternate key `name` accepted instead of `file` for convenience

**Script object fields:**
- `file` (string) - Script file name (required in table form)
- `tags` (array of strings, optional) - Classification labels for selection/filtering

---

## Variables

### Built-in Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `{home_dir}` | `~` | User's home directory |
| `{config_dir}` | `{home_dir}/.config` | Base config directory |

### Variable Expansion

Variables are expanded recursively:

```toml
home_dir = "~"
config_dir = "{home_dir}/.config"  # Expands to ~/.config
```

Used in targets:

```toml
[[link]]
target = "{config_dir}/git"  # Expands to ~/.config/git
```

### Runtime Override

User can override at runtime:

```bash
# Use custom config directory
merlin link git --config-dir ~/.dotfiles

# Override home directory
merlin link git --home-dir /Users/custom
```

---

## Benefits of This Approach

✅ **Declarative** - Clear, readable configuration  
✅ **Flexible** - Handles simple and complex cases  
✅ **Variables** - User-customizable paths  
✅ **Dependencies** - Automatic resolution and ordering  
✅ **Scripts** - Ordered execution for complex setup  
✅ **Profiles** - Different configs for different machines  
✅ **Validation** - Merlin can validate before executing  

---

## Migration Notes

### From v0.x (Old Format)

**Old:** `install_script = "install.sh"`  
**New:** `[scripts]` table with directory and array

**Old:** `target_base` key in links  
**New:** Just `target` (unified keys)

**Old:** Hardcoded paths like `~/.config/tool`  
**New:** Variables like `{config_dir}/tool`

**Old:** No dependencies field  
**New:** `dependencies = ["brew"]`

---

## Implementation Notes

When implementing Merlin:
- Parse TOML files using a robust library
- Expand variables recursively (detect circular references)
- Build dependency graph and topological sort
- Create symlinks with proper conflict handling
- Execute scripts with proper error handling
- Provide clear feedback for each step

---

**Next:** See [DOTFILES_STRUCTURE.md](./DOTFILES_STRUCTURE.md) for how to structure your dotfiles repository.
