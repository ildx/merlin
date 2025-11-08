package symlink

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ildx/merlin/internal/backup"
)

// ConflictStrategy defines how to handle conflicts
type ConflictStrategy int

const (
	// StrategySkip skips files that already exist
	StrategySkip ConflictStrategy = iota
	// StrategyBackup backs up existing files before linking
	StrategyBackup
	// StrategyOverwrite overwrites existing files (dangerous!)
	StrategyOverwrite
	// StrategyInteractive prompts the user for each conflict
	StrategyInteractive
)

func (s ConflictStrategy) String() string {
	switch s {
	case StrategySkip:
		return "skip"
	case StrategyBackup:
		return "backup"
	case StrategyOverwrite:
		return "overwrite"
	case StrategyInteractive:
		return "interactive"
	default:
		return "unknown"
	}
}

// ParseStrategy parses a string into a ConflictStrategy
func ParseStrategy(s string) (ConflictStrategy, error) {
	switch s {
	case "skip":
		return StrategySkip, nil
	case "backup":
		return StrategyBackup, nil
	case "overwrite":
		return StrategyOverwrite, nil
	case "interactive":
		return StrategyInteractive, nil
	default:
		return StrategySkip, fmt.Errorf("unknown strategy: %s", s)
	}
}

// ResolveConflict handles a conflict based on the strategy
func ResolveConflict(source, target string, strategy ConflictStrategy, dryRun bool) (*LinkResult, error) {
	result := &LinkResult{
		Source: source,
		Target: target,
	}

	// Check if source exists
	sourceInfo, err := os.Lstat(source)
	if err != nil {
		result.Status = LinkStatusError
		result.Message = fmt.Sprintf("source does not exist: %v", err)
		return result, fmt.Errorf("source %s does not exist: %w", source, err)
	}
	result.IsDir = sourceInfo.IsDir()

	// Check if target exists
	targetInfo, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			// No conflict - create symlink normally
			return CreateSymlink(source, target, dryRun)
		}
		result.Status = LinkStatusError
		result.Message = fmt.Sprintf("failed to check target: %v", err)
		return result, fmt.Errorf("failed to check target: %w", err)
	}

	// Check if target is already correctly linked
	if targetInfo.Mode()&os.ModeSymlink != 0 {
		linkDest, err := os.Readlink(target)
		if err == nil {
			absLinkDest := linkDest
			if !filepath.IsAbs(linkDest) {
				absLinkDest = filepath.Join(filepath.Dir(target), linkDest)
			}
			absLinkDest = filepath.Clean(absLinkDest)
			absSource := filepath.Clean(source)

			if absLinkDest == absSource {
				result.Status = LinkStatusAlreadyLinked
				result.Message = "already correctly linked"
				return result, nil
			}
		}
	}

	// Handle conflict based on strategy
	switch strategy {
	case StrategySkip:
		result.Status = LinkStatusSkipped
		result.Message = "skipped due to conflict"
		return result, nil

	case StrategyBackup:
		if dryRun {
			result.Status = LinkStatusSuccess
			result.Message = "would backup and link (dry-run)"
			return result, nil
		}

		// Create backup using backup system
		manifest, err := backup.CreateBackup([]string{target}, fmt.Sprintf("Before linking %s", source))
		if err != nil {
			result.Status = LinkStatusError
			result.Message = fmt.Sprintf("failed to backup: %v", err)
			return result, fmt.Errorf("failed to backup: %w", err)
		}

		// Remove existing file/directory now that it's backed up
		if err := os.RemoveAll(target); err != nil {
			result.Status = LinkStatusError
			result.Message = fmt.Sprintf("failed to remove after backup: %v", err)
			return result, fmt.Errorf("failed to remove: %w", err)
		}

		// Create symlink
		if err := os.Symlink(source, target); err != nil {
			// Try to restore from backup
			backup.RestoreBackup(manifest.ID, []string{target})
			result.Status = LinkStatusError
			result.Message = fmt.Sprintf("failed to create symlink: %v", err)
			return result, fmt.Errorf("failed to create symlink: %w", err)
		}

		result.Status = LinkStatusSuccess
		result.Message = fmt.Sprintf("backed up (ID: %s) and linked", manifest.ID)
		return result, nil

	case StrategyOverwrite:
		if dryRun {
			result.Status = LinkStatusSuccess
			result.Message = "would overwrite and link (dry-run)"
			return result, nil
		}

		// Remove existing file/directory
		if err := os.RemoveAll(target); err != nil {
			result.Status = LinkStatusError
			result.Message = fmt.Sprintf("failed to remove: %v", err)
			return result, fmt.Errorf("failed to remove: %w", err)
		}

		// Create symlink
		if err := os.Symlink(source, target); err != nil {
			result.Status = LinkStatusError
			result.Message = fmt.Sprintf("failed to create symlink: %v", err)
			return result, fmt.Errorf("failed to create symlink: %w", err)
		}

		result.Status = LinkStatusSuccess
		result.Message = "overwritten and linked"
		return result, nil

	case StrategyInteractive:
		// This will be handled by the caller
		result.Status = LinkStatusConflict
		result.Message = "requires interactive resolution"
		return result, nil

	default:
		result.Status = LinkStatusError
		result.Message = "unknown strategy"
		return result, fmt.Errorf("unknown strategy: %v", strategy)
	}
}

// generateBackupPath generates a backup filename with timestamp
func generateBackupPath(path string) string {
	timestamp := time.Now().Format("20060102_150405")
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	return filepath.Join(dir, fmt.Sprintf("%s.backup_%s", base, timestamp))
}

// LinkToolWithStrategy links all configured links for a tool with conflict resolution
func LinkToolWithStrategy(tool *ToolConfig, strategy ConflictStrategy, dryRun bool) ([]*LinkResult, error) {
	var allResults []*LinkResult

	for _, link := range tool.Links {
		result, err := ResolveConflict(link.Source, link.Target, strategy, dryRun)
		allResults = append(allResults, result)

		// Continue with other links even if one fails
		if err != nil && result.Status == LinkStatusError {
			continue
		}
	}

	return allResults, nil
}
