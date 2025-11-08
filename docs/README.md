# Merlin Documentation

Welcome to the Merlin documentation! This directory contains everything you need to understand, use, and develop Merlin.

---

## 📚 Documentation Overview

### For Users

1. **[DOTFILES_STRUCTURE.md](./DOTFILES_STRUCTURE.md)** ⭐ **Start here if you're setting up dotfiles**
   - Required directory structure
   - How Merlin discovers and links configs
   - Example dotfiles repositories
   - Migration from other systems
   - Best practices

2. **[MERLIN_TOML_SPEC.md](./MERLIN_TOML_SPEC.md)** 📖 **Configuration reference**
   - Root `merlin.toml` specification
   - Per-tool `merlin.toml` specification
   - Complete examples for all use cases
   - Execution flow and behavior

### For Developers

3. **[IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md)** 🔧 **Development roadmap**
   - 9 phases, broken into small testable steps
   - Architecture principles
   - Tech stack and dependencies
   - Directory structure
   - Testing strategy

---

## Quick Navigation

### "I want to use Merlin with my dotfiles"
→ Read [DOTFILES_STRUCTURE.md](./DOTFILES_STRUCTURE.md) first, then [MERLIN_TOML_SPEC.md](./MERLIN_TOML_SPEC.md) for advanced config.

### "I want to understand how Merlin works"
→ Start with the main [README.md](../README.md), then [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md).

### "I want to contribute to Merlin"
→ Read [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) to understand the architecture and current progress.

### "I'm confused about configuration"
→ [MERLIN_TOML_SPEC.md](./MERLIN_TOML_SPEC.md) has complete examples and reference.

---

## Document Status

| Document | Status | Last Updated | Purpose |
|----------|--------|--------------|---------|
| DOTFILES_STRUCTURE.md | ✅ Complete | Nov 8, 2025 | User guide for dotfiles structure |
| MERLIN_TOML_SPEC.md | ✅ Complete | Nov 8, 2025 | Configuration specification |
| IMPLEMENTATION_PLAN.md | ✅ Complete | Nov 8, 2025 | Development roadmap |
| README.md (this) | ✅ Complete | Nov 8, 2025 | Documentation index |

---

## Key Concepts

### Modular Configuration
- **Root `merlin.toml`**: Global settings, variables, profiles, preinstall tools
- **Per-tool `merlin.toml`**: Tool-specific symlinks, scripts, dependencies
- **Variables**: Flexible paths (`{config_dir}`, `{home_dir}`)
- **Dependencies**: Automatic resolution and ordering

### Directory Structure
```
dotfiles/
├── merlin.toml                  # Root config
└── config/
    ├── TOOL/
    │   ├── config/              # Files to symlink
    │   ├── merlin.toml          # Optional tool config
    │   ├── scripts/             # Optional setup scripts
    │   └── *.toml               # Optional tool-specific data
    ├── brew/config/brew.toml    # Homebrew packages
    └── mas/config/mas.toml      # Mac App Store apps
```

### Tool Lifecycle
1. Resolve dependencies
2. Expand variables
3. Create symlinks
4. Run scripts (in order)
5. Handle conflicts

---

## Common Questions

**Q: Do I need a `merlin.toml` for every tool?**  
A: No! Only if the tool needs custom symlink paths or install scripts. Tools without one use smart defaults.

**Q: What's the difference between root and per-tool `merlin.toml`?**  
A: Root = global settings, variables, profiles, preinstall. Per-tool = dependencies, symlinks, scripts for that specific tool.

**Q: Can I use Merlin with my existing dotfiles?**  
A: Yes! Just reorganize to match the expected structure. See DOTFILES_STRUCTURE.md migration guide.

**Q: Why not use Stow?**  
A: Merlin has native symlinking with better UX, conflict handling, and no external dependencies.

---

## Documentation Philosophy

Our documentation follows these principles:

1. **Examples first** - Show, don't just tell
2. **Progressive disclosure** - Start simple, add complexity
3. **Real-world scenarios** - Based on actual use cases
4. **Keep in sync** - Update docs when code changes
5. **One source of truth** - Each concept documented in one place

---

## Contributing to Docs

Found something unclear or missing? Here's how to improve the docs:

1. **Typos/clarity**: Feel free to fix directly
2. **Missing examples**: Add to the relevant spec file
3. **New features**: Update all relevant docs
4. **Breaking changes**: Update examples and add migration notes

---

## Changelog

### November 8, 2025 (Updated)
- Added variables system (`{config_dir}`, `{home_dir}`)
- Unified link key naming (`source` + `target` only)
- Changed scripts from `install_script` to `[scripts]` table
- Added `dependencies` field for automatic ordering
- Added `[preinstall]` section for system requirements
- Updated all examples to match Covenant structure
- Separated tool-specific data from Merlin instructions

---

Need help? Open an issue or check the main [README.md](../README.md) for contact info.

