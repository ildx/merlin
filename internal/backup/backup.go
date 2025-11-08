package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// BackupManifest contains metadata about a backup operation
type BackupManifest struct {
	ID        string        `json:"id"`         // Timestamp-based unique identifier
	Timestamp time.Time     `json:"timestamp"`  // When backup was created
	Reason    string        `json:"reason"`     // Why this backup was created
	Files     []BackupEntry `json:"files"`      // Files included in this backup
	MerlinDir string        `json:"merlin_dir"` // Base Merlin directory at time of backup
}

// BackupEntry represents a single backed up file
type BackupEntry struct {
	OriginalPath string `json:"original_path"` // Original file location
	BackupPath   string `json:"backup_path"`   // Location in backup directory
	Size         int64  `json:"size"`          // File size in bytes
	Checksum     string `json:"checksum"`      // SHA256 hash for integrity verification
}

// BackupLocation returns the base directory for all backups
func BackupLocation() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, ".merlin", "backups"), nil
}

// GenerateBackupID creates a unique backup identifier from current timestamp
func GenerateBackupID() string {
	return time.Now().Format("20060102_150405")
}

// CreateBackup copies files to a new backup location and generates manifest
func CreateBackup(files []string, reason string) (*BackupManifest, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files specified for backup")
	}

	backupID := GenerateBackupID()
	baseDir, err := BackupLocation()
	if err != nil {
		return nil, err
	}

	backupDir := filepath.Join(baseDir, backupID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}

	manifest := &BackupManifest{
		ID:        backupID,
		Timestamp: time.Now(),
		Reason:    reason,
		Files:     make([]BackupEntry, 0, len(files)),
	}

	// Get Merlin directory for reference
	home, _ := os.UserHomeDir()
	manifest.MerlinDir = filepath.Join(home, ".merlin")

	// Copy each file to backup location
	for _, originalPath := range files {
		// Expand home directory
		if len(originalPath) > 0 && originalPath[0] == '~' {
			originalPath = filepath.Join(home, originalPath[1:])
		}

		// Check if file exists
		info, err := os.Stat(originalPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Skip non-existent files
			}
			return nil, fmt.Errorf("stat file %s: %w", originalPath, err)
		}

		// Skip directories for now
		if info.IsDir() {
			continue
		}

		// Calculate relative backup path to preserve structure
		relPath := filepath.Base(originalPath)
		backupFilePath := filepath.Join(backupDir, relPath)

		// Copy file
		if err := copyFile(originalPath, backupFilePath); err != nil {
			return nil, fmt.Errorf("copy file %s: %w", originalPath, err)
		}

		// Calculate checksum
		checksum, err := calculateChecksum(backupFilePath)
		if err != nil {
			return nil, fmt.Errorf("calculate checksum for %s: %w", originalPath, err)
		}

		entry := BackupEntry{
			OriginalPath: originalPath,
			BackupPath:   backupFilePath,
			Size:         info.Size(),
			Checksum:     checksum,
		}
		manifest.Files = append(manifest.Files, entry)
	}

	// Save manifest
	manifestPath := filepath.Join(backupDir, "manifest.json")
	if err := saveManifest(manifest, manifestPath); err != nil {
		return nil, fmt.Errorf("save manifest: %w", err)
	}

	return manifest, nil
}

// ListBackups returns all available backups sorted by timestamp (newest first)
func ListBackups() ([]*BackupManifest, error) {
	baseDir, err := BackupLocation()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*BackupManifest{}, nil // No backups yet
		}
		return nil, fmt.Errorf("read backup directory: %w", err)
	}

	var manifests []*BackupManifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(baseDir, entry.Name(), "manifest.json")
		manifest, err := loadManifest(manifestPath)
		if err != nil {
			continue // Skip invalid backups
		}
		manifests = append(manifests, manifest)
	}

	// Sort by timestamp, newest first
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].Timestamp.After(manifests[j].Timestamp)
	})

	return manifests, nil
}

// GetBackupInfo loads and returns a specific backup manifest
func GetBackupInfo(backupID string) (*BackupManifest, error) {
	baseDir, err := BackupLocation()
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(baseDir, backupID, "manifest.json")
	return loadManifest(manifestPath)
}

// RestoreBackup restores files from a backup, optionally filtering by specific files
func RestoreBackup(backupID string, selectiveFiles []string) error {
	manifest, err := GetBackupInfo(backupID)
	if err != nil {
		return fmt.Errorf("load backup manifest: %w", err)
	}

	// Create set of selective files for quick lookup
	selective := make(map[string]bool)
	for _, f := range selectiveFiles {
		selective[f] = true
	}

	for _, entry := range manifest.Files {
		// Skip if selective restore and file not in list
		if len(selectiveFiles) > 0 && !selective[entry.OriginalPath] {
			continue
		}

		// Verify backup file still exists and checksum matches
		if err := verifyBackupFile(entry); err != nil {
			return fmt.Errorf("verify backup file %s: %w", entry.BackupPath, err)
		}

		// Ensure target directory exists
		targetDir := filepath.Dir(entry.OriginalPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("create target directory %s: %w", targetDir, err)
		}

		// Copy file back to original location
		if err := copyFile(entry.BackupPath, entry.OriginalPath); err != nil {
			return fmt.Errorf("restore file %s: %w", entry.OriginalPath, err)
		}
	}

	return nil
}

// DeleteBackup removes a backup and its manifest
func DeleteBackup(backupID string) error {
	baseDir, err := BackupLocation()
	if err != nil {
		return err
	}

	backupDir := filepath.Join(baseDir, backupID)
	return os.RemoveAll(backupDir)
}

// Helper functions

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func verifyBackupFile(entry BackupEntry) error {
	// Check file exists
	info, err := os.Stat(entry.BackupPath)
	if err != nil {
		return fmt.Errorf("backup file missing: %w", err)
	}

	// Verify size
	if info.Size() != entry.Size {
		return fmt.Errorf("size mismatch: expected %d, got %d", entry.Size, info.Size())
	}

	// Verify checksum
	checksum, err := calculateChecksum(entry.BackupPath)
	if err != nil {
		return fmt.Errorf("calculate checksum: %w", err)
	}
	if checksum != entry.Checksum {
		return fmt.Errorf("checksum mismatch")
	}

	return nil
}

func saveManifest(manifest *BackupManifest, path string) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func loadManifest(path string) (*BackupManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest BackupManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}
