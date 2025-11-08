package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ildx/merlin/internal/backup"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage configuration backups",
	Long: `Create, list, restore, and manage backups of your configuration files.
	
Backups provide a safety net for your dotfiles, allowing you to restore
previous versions if something goes wrong.`,
}

var backupCreateCmd = &cobra.Command{
	Use:   "create [files...]",
	Short: "Create a backup of specified files",
	Long: `Create a new backup of one or more configuration files.
	
Examples:
  merlin backup create ~/.zshrc ~/.gitconfig --reason "Before major changes"
  merlin backup create ~/covenant/config/zsh/config/*.zsh`,
	RunE: runBackupCreate,
}

var backupListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available backups",
	Long:    `Show all backups with their IDs, timestamps, reasons, and file counts.`,
	RunE:    runBackupList,
}

var backupShowCmd = &cobra.Command{
	Use:   "show <backup-id>",
	Short: "Show detailed information about a backup",
	Long:  `Display the manifest and file list for a specific backup.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupShow,
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore <backup-id>",
	Short: "Restore files from a backup",
	Long: `Restore configuration files from a previous backup.
	
By default, all files in the backup are restored. Use --files to restore
specific files only.

Examples:
  merlin backup restore 20250108_143022
  merlin backup restore 20250108_143022 --files ~/.zshrc,~/.gitconfig`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupRestore,
}

var backupCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Delete old backups",
	Long: `Remove backups older than a specified number of days.
	
Use --keep to specify how many recent backups to preserve.
Use --older-than to delete backups older than N days.

Examples:
  merlin backup clean --keep 5
  merlin backup clean --older-than 30`,
	RunE: runBackupClean,
}

var backupDeleteCmd = &cobra.Command{
	Use:   "delete <backup-id>",
	Short: "Delete a specific backup",
	Long:  `Permanently remove a backup and all its files.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupDelete,
}

var (
	backupReason    string
	backupFiles     string
	backupKeep      int
	backupOlderThan int
	backupForce     bool
)

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupShowCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	backupCmd.AddCommand(backupCleanCmd)
	backupCmd.AddCommand(backupDeleteCmd)

	// Create flags
	backupCreateCmd.Flags().StringVarP(&backupReason, "reason", "r", "", "Reason for creating this backup")

	// Restore flags
	backupRestoreCmd.Flags().StringVar(&backupFiles, "files", "", "Comma-separated list of files to restore (default: all)")
	backupRestoreCmd.Flags().BoolVar(&backupForce, "force", false, "Skip confirmation prompt")

	// Clean flags
	backupCleanCmd.Flags().IntVar(&backupKeep, "keep", 0, "Number of recent backups to keep (default: keep all)")
	backupCleanCmd.Flags().IntVar(&backupOlderThan, "older-than", 0, "Delete backups older than N days")
	backupCleanCmd.Flags().BoolVar(&backupForce, "force", false, "Skip confirmation prompt")
}

func runBackupCreate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no files specified for backup")
	}

	if backupReason == "" {
		backupReason = "Manual backup"
	}

	// Expand globs in file arguments
	var expandedFiles []string
	for _, pattern := range args {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern %s: %w", pattern, err)
		}
		if len(matches) == 0 {
			// No matches, use pattern as-is (might be exact path)
			expandedFiles = append(expandedFiles, pattern)
		} else {
			expandedFiles = append(expandedFiles, matches...)
		}
	}

	fmt.Printf("Creating backup of %d file(s)...\n", len(expandedFiles))

	manifest, err := backup.CreateBackup(expandedFiles, backupReason)
	if err != nil {
		return fmt.Errorf("create backup: %w", err)
	}

	fmt.Printf("\n✅ Backup created successfully\n")
	fmt.Printf("  ID: %s\n", manifest.ID)
	fmt.Printf("  Files: %d\n", len(manifest.Files))
	fmt.Printf("  Reason: %s\n", manifest.Reason)
	fmt.Printf("\nRestore with: merlin backup restore %s\n", manifest.ID)

	return nil
}

func runBackupList(cmd *cobra.Command, args []string) error {
	backups, err := backup.ListBackups()
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups found.")
		fmt.Println("\nCreate a backup with: merlin backup create <files...>")
		return nil
	}

	fmt.Printf("Found %d backup(s):\n\n", len(backups))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTIMESTAMP\tFILES\tREASON")
	fmt.Fprintln(w, "--\t---------\t-----\t------")

	for _, b := range backups {
		timestamp := b.Timestamp.Format("2006-01-02 15:04:05")
		reason := b.Reason
		if len(reason) > 40 {
			reason = reason[:37] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", b.ID, timestamp, len(b.Files), reason)
	}

	w.Flush()
	fmt.Println("\nUse 'merlin backup show <id>' for detailed information")

	return nil
}

func runBackupShow(cmd *cobra.Command, args []string) error {
	backupID := args[0]

	manifest, err := backup.GetBackupInfo(backupID)
	if err != nil {
		return fmt.Errorf("get backup info: %w", err)
	}

	fmt.Printf("Backup: %s\n", manifest.ID)
	fmt.Printf("Created: %s\n", manifest.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Reason: %s\n", manifest.Reason)
	fmt.Printf("Files: %d\n\n", len(manifest.Files))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ORIGINAL PATH\tSIZE\tCHECKSUM")
	fmt.Fprintln(w, "-------------\t----\t--------")

	for _, entry := range manifest.Files {
		sizeKB := float64(entry.Size) / 1024
		checksum := entry.Checksum[:12] + "..." // Show first 12 chars
		fmt.Fprintf(w, "%s\t%.1f KB\t%s\n", entry.OriginalPath, sizeKB, checksum)
	}

	w.Flush()
	fmt.Printf("\nRestore with: merlin backup restore %s\n", manifest.ID)

	return nil
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
	backupID := args[0]

	// Load backup info
	manifest, err := backup.GetBackupInfo(backupID)
	if err != nil {
		return fmt.Errorf("get backup info: %w", err)
	}

	// Parse selective files if provided
	var selectiveFiles []string
	if backupFiles != "" {
		selectiveFiles = strings.Split(backupFiles, ",")
		for i := range selectiveFiles {
			selectiveFiles[i] = strings.TrimSpace(selectiveFiles[i])
		}
	}

	// Show what will be restored
	fmt.Printf("Backup: %s\n", manifest.ID)
	fmt.Printf("Created: %s\n", manifest.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Reason: %s\n\n", manifest.Reason)

	if len(selectiveFiles) > 0 {
		fmt.Printf("Will restore %d file(s):\n", len(selectiveFiles))
		for _, f := range selectiveFiles {
			fmt.Printf("  • %s\n", f)
		}
	} else {
		fmt.Printf("Will restore all %d file(s) from backup\n", len(manifest.Files))
	}

	// Confirmation prompt (unless --force)
	if !backupForce {
		fmt.Print("\n⚠️  This will overwrite existing files. Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Restore cancelled.")
			return nil
		}
	}

	fmt.Println("\nRestoring files...")
	if err := backup.RestoreBackup(backupID, selectiveFiles); err != nil {
		return fmt.Errorf("restore backup: %w", err)
	}

	fmt.Println("✅ Backup restored successfully")

	return nil
}

func runBackupClean(cmd *cobra.Command, args []string) error {
	backups, err := backup.ListBackups()
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups to clean.")
		return nil
	}

	var toDelete []*backup.BackupManifest

	// Delete based on --keep flag
	if backupKeep > 0 {
		if len(backups) > backupKeep {
			toDelete = backups[backupKeep:]
		}
	}

	// Delete based on --older-than flag
	if backupOlderThan > 0 {
		cutoff := time.Now().AddDate(0, 0, -backupOlderThan)
		for _, b := range backups {
			if b.Timestamp.Before(cutoff) {
				// Check if not already in toDelete list
				found := false
				for _, d := range toDelete {
					if d.ID == b.ID {
						found = true
						break
					}
				}
				if !found {
					toDelete = append(toDelete, b)
				}
			}
		}
	}

	if len(toDelete) == 0 {
		fmt.Println("No backups match deletion criteria.")
		return nil
	}

	fmt.Printf("Will delete %d backup(s):\n\n", len(toDelete))
	for _, b := range toDelete {
		fmt.Printf("  • %s - %s (%d files)\n", b.ID, b.Timestamp.Format("2006-01-02 15:04"), len(b.Files))
	}

	// Confirmation prompt
	if !backupForce {
		fmt.Print("\n⚠️  This cannot be undone. Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Clean cancelled.")
			return nil
		}
	}

	fmt.Println("\nDeleting backups...")
	for _, b := range toDelete {
		if err := backup.DeleteBackup(b.ID); err != nil {
			fmt.Printf("  ⚠️  Failed to delete %s: %v\n", b.ID, err)
		} else {
			fmt.Printf("  ✅ Deleted %s\n", b.ID)
		}
	}

	return nil
}

func runBackupDelete(cmd *cobra.Command, args []string) error {
	backupID := args[0]

	// Get backup info first
	manifest, err := backup.GetBackupInfo(backupID)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	fmt.Printf("Backup: %s\n", manifest.ID)
	fmt.Printf("Created: %s\n", manifest.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Files: %d\n\n", len(manifest.Files))

	// Confirmation
	if !backupForce {
		fmt.Print("⚠️  Delete this backup? This cannot be undone. [y/N]: ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Delete cancelled.")
			return nil
		}
	}

	if err := backup.DeleteBackup(backupID); err != nil {
		return fmt.Errorf("delete backup: %w", err)
	}

	fmt.Println("✅ Backup deleted successfully")
	return nil
}
