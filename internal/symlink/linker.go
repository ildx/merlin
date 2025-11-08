package symlink

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LinkResult represents the outcome of a symlink operation
type LinkResult struct {
	Source  string
	Target  string
	Status  LinkStatus
	Message string
	IsDir   bool
}

// LinkStatus represents the status of a link operation
type LinkStatus int

const (
	LinkStatusSuccess LinkStatus = iota
	LinkStatusSkipped
	LinkStatusError
	LinkStatusAlreadyLinked
	LinkStatusConflict
)

func (s LinkStatus) String() string {
	switch s {
	case LinkStatusSuccess:
		return "success"
	case LinkStatusSkipped:
		return "skipped"
	case LinkStatusError:
		return "error"
	case LinkStatusAlreadyLinked:
		return "already_linked"
	case LinkStatusConflict:
		return "conflict"
	default:
		return "unknown"
	}
}

// CreateSymlink creates a symbolic link from target to source
func CreateSymlink(source, target string, dryRun bool) (*LinkResult, error) {
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

	// Check if target already exists
	targetInfo, err := os.Lstat(target)
	if err == nil {
		// Target exists - check if it's already our symlink
		if targetInfo.Mode()&os.ModeSymlink != 0 {
			// It's a symlink - check where it points
			linkDest, err := os.Readlink(target)
			if err != nil {
				result.Status = LinkStatusError
				result.Message = fmt.Sprintf("failed to read existing symlink: %v", err)
				return result, fmt.Errorf("failed to read symlink %s: %w", target, err)
			}

			// Resolve to absolute path for comparison
			absLinkDest := linkDest
			if !filepath.IsAbs(linkDest) {
				absLinkDest = filepath.Join(filepath.Dir(target), linkDest)
			}

			// Clean both paths for comparison
			absLinkDest = filepath.Clean(absLinkDest)
			absSource := filepath.Clean(source)

			if absLinkDest == absSource {
				result.Status = LinkStatusAlreadyLinked
				result.Message = "already correctly linked"
				return result, nil
			}

			// Symlink exists but points elsewhere
			result.Status = LinkStatusConflict
			result.Message = fmt.Sprintf("symlink exists but points to: %s", linkDest)
			return result, nil
		}

		// Target exists but is not a symlink (file or directory conflict)
		result.Status = LinkStatusConflict
		if targetInfo.IsDir() {
			result.Message = "directory already exists at target"
		} else {
			result.Message = "file already exists at target"
		}
		return result, nil
	}

	// Target doesn't exist - we can create the symlink
	if dryRun {
		result.Status = LinkStatusSuccess
		result.Message = "would create symlink (dry-run)"
		return result, nil
	}

	// Ensure parent directory exists
	targetDir := filepath.Dir(target)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		result.Status = LinkStatusError
		result.Message = fmt.Sprintf("failed to create parent directory: %v", err)
		return result, fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Create the symlink
	if err := os.Symlink(source, target); err != nil {
		result.Status = LinkStatusError
		result.Message = fmt.Sprintf("failed to create symlink: %v", err)
		return result, fmt.Errorf("failed to create symlink: %w", err)
	}

	result.Status = LinkStatusSuccess
	result.Message = "symlink created successfully"
	return result, nil
}

// WalkAndLink recursively walks a source directory and creates symlinks
// for all files and subdirectories in the target directory
func WalkAndLink(source, target string, dryRun bool) ([]*LinkResult, error) {
	var results []*LinkResult

	// Check if source is a directory
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("failed to stat source %s: %w", source, err)
	}

	// If source is not a directory, just link the single file/directory
	if !sourceInfo.IsDir() {
		result, err := CreateSymlink(source, target, dryRun)
		if err != nil && result.Status == LinkStatusError {
			return []*LinkResult{result}, err
		}
		return []*LinkResult{result}, nil
	}

	// Source is a directory - we need to decide whether to link the whole
	// directory or its contents
	// For now, we'll link individual files to preserve the ability to have
	// some files from dotfiles and some from elsewhere

	err = filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root source directory itself
		if path == source {
			return nil
		}

		// Calculate relative path from source
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Calculate target path
		targetPath := filepath.Join(target, relPath)

		// Skip hidden files and directories (starting with .)
		if strings.HasPrefix(d.Name(), ".") && path != source {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// If it's a directory, just ensure it exists at target
		if d.IsDir() {
			if !dryRun {
				if err := os.MkdirAll(targetPath, 0755); err != nil {
					result := &LinkResult{
						Source:  path,
						Target:  targetPath,
						Status:  LinkStatusError,
						Message: fmt.Sprintf("failed to create directory: %v", err),
						IsDir:   true,
					}
					results = append(results, result)
					return nil // Continue walking
				}
			}
			// Don't add directory creation to results (too verbose)
			return nil
		}

		// It's a file - create symlink
		result, err := CreateSymlink(path, targetPath, dryRun)
		results = append(results, result)
		
		// Continue even if there was an error linking this file
		return nil
	})

	if err != nil {
		return results, fmt.Errorf("failed to walk directory: %w", err)
	}

	return results, nil
}

// LinkTool links all configured links for a tool
func LinkTool(tool *ToolConfig, dryRun bool) ([]*LinkResult, error) {
	var allResults []*LinkResult

	for _, link := range tool.Links {
		var results []*LinkResult

		if link.IsDir {
			// If we want to link the whole directory as one symlink
			// (not its contents), use CreateSymlink
			// Otherwise use WalkAndLink to link contents
			
			// Check if we should link the directory itself or its contents
			// For now, we'll link the directory itself
			result, err := CreateSymlink(link.Source, link.Target, dryRun)
			results = []*LinkResult{result}
			if err != nil && result.Status == LinkStatusError {
				// Continue with other links even if one fails
				allResults = append(allResults, results...)
				continue
			}
		} else {
			// Single file
			result, err := CreateSymlink(link.Source, link.Target, dryRun)
			results = []*LinkResult{result}
			if err != nil && result.Status == LinkStatusError {
				// Continue with other links even if one fails
				allResults = append(allResults, results...)
				continue
			}
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// IsLinked checks if a target path is already correctly symlinked to source
func IsLinked(source, target string) (bool, error) {
	// Check if target exists
	targetInfo, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat target: %w", err)
	}

	// Check if it's a symlink
	if targetInfo.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	// Read where the symlink points
	linkDest, err := os.Readlink(target)
	if err != nil {
		return false, fmt.Errorf("failed to read symlink: %w", err)
	}

	// Resolve to absolute path
	absLinkDest := linkDest
	if !filepath.IsAbs(linkDest) {
		absLinkDest = filepath.Join(filepath.Dir(target), linkDest)
	}

	// Clean and compare
	absLinkDest = filepath.Clean(absLinkDest)
	absSource := filepath.Clean(source)

	return absLinkDest == absSource, nil
}

// GetLinkStatus checks the status of all links for a tool
func GetLinkStatus(tool *ToolConfig) map[string]LinkStatus {
	status := make(map[string]LinkStatus)

	for _, link := range tool.Links {
		isLinked, err := IsLinked(link.Source, link.Target)
		if err != nil {
			status[link.Target] = LinkStatusError
			continue
		}

		if isLinked {
			status[link.Target] = LinkStatusAlreadyLinked
		} else {
			// Check if target exists (conflict)
			if _, err := os.Stat(link.Target); err == nil {
				status[link.Target] = LinkStatusConflict
			} else {
				status[link.Target] = LinkStatusSkipped
			}
		}
	}

	return status
}

// UnlinkResult represents the outcome of an unlink operation
type UnlinkResult struct {
	Target  string
	Status  LinkStatus
	Message string
}

// RemoveSymlink removes a symlink if it points to the expected source
func RemoveSymlink(source, target string, dryRun bool) (*UnlinkResult, error) {
	result := &UnlinkResult{
		Target: target,
	}

	// Check if target exists
	targetInfo, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = LinkStatusSkipped
			result.Message = "target does not exist"
			return result, nil
		}
		result.Status = LinkStatusError
		result.Message = fmt.Sprintf("failed to check target: %v", err)
		return result, fmt.Errorf("failed to check target: %w", err)
	}

	// Check if target is a symlink
	if targetInfo.Mode()&os.ModeSymlink == 0 {
		result.Status = LinkStatusSkipped
		result.Message = "target is not a symlink (safety check)"
		return result, nil
	}

	// Read where the symlink points
	linkDest, err := os.Readlink(target)
	if err != nil {
		result.Status = LinkStatusError
		result.Message = fmt.Sprintf("failed to read symlink: %v", err)
		return result, fmt.Errorf("failed to read symlink: %w", err)
	}

	// Resolve to absolute path
	absLinkDest := linkDest
	if !filepath.IsAbs(linkDest) {
		absLinkDest = filepath.Join(filepath.Dir(target), linkDest)
	}
	absLinkDest = filepath.Clean(absLinkDest)
	absSource := filepath.Clean(source)

	// Safety check: only remove if it points to our source
	if absLinkDest != absSource {
		result.Status = LinkStatusSkipped
		result.Message = fmt.Sprintf("symlink points to %s, not our source (safety check)", linkDest)
		return result, nil
	}

	// Remove the symlink
	if dryRun {
		result.Status = LinkStatusSuccess
		result.Message = "would remove symlink (dry-run)"
		return result, nil
	}

	if err := os.Remove(target); err != nil {
		result.Status = LinkStatusError
		result.Message = fmt.Sprintf("failed to remove: %v", err)
		return result, fmt.Errorf("failed to remove: %w", err)
	}

	result.Status = LinkStatusSuccess
	result.Message = "symlink removed"
	return result, nil
}

// UnlinkTool removes all symlinks for a tool
func UnlinkTool(tool *ToolConfig, dryRun bool) ([]*UnlinkResult, error) {
	var results []*UnlinkResult

	for _, link := range tool.Links {
		result, err := RemoveSymlink(link.Source, link.Target, dryRun)
		results = append(results, result)
		
		// Continue with other links even if one fails
		if err != nil && result.Status == LinkStatusError {
			continue
		}
	}

	return results, nil
}

