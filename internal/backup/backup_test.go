package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateBackupID(t *testing.T) {
	id1 := GenerateBackupID()
	time.Sleep(1 * time.Second) // Need full second for timestamp format
	id2 := GenerateBackupID()

	if id1 == id2 {
		t.Error("Expected different backup IDs")
	}

	// Should match format YYYYMMDD_HHMMSS
	if len(id1) != 15 {
		t.Errorf("Expected ID length 15, got %d", len(id1))
	}
}

func TestCreateBackupAndRestore(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()

	// Override backup location to use temp dir
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")

	content1 := []byte("test content 1")
	content2 := []byte("test content 2")

	if err := os.WriteFile(testFile1, content1, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile2, content2, 0644); err != nil {
		t.Fatal(err)
	}

	// Create backup
	manifest, err := CreateBackup([]string{testFile1, testFile2}, "test backup")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	if len(manifest.Files) != 2 {
		t.Errorf("Expected 2 files in backup, got %d", len(manifest.Files))
	}

	if manifest.Reason != "test backup" {
		t.Errorf("Expected reason 'test backup', got '%s'", manifest.Reason)
	}

	// Verify files were backed up
	for _, entry := range manifest.Files {
		if _, err := os.Stat(entry.BackupPath); os.IsNotExist(err) {
			t.Errorf("Backup file not created: %s", entry.BackupPath)
		}
	}

	// Modify original files
	if err := os.WriteFile(testFile1, []byte("modified content 1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile2, []byte("modified content 2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Restore backup
	if err := RestoreBackup(manifest.ID, nil); err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	// Verify files were restored
	restored1, err := os.ReadFile(testFile1)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored1) != string(content1) {
		t.Errorf("File 1 not restored correctly. Expected '%s', got '%s'", content1, restored1)
	}

	restored2, err := os.ReadFile(testFile2)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored2) != string(content2) {
		t.Errorf("File 2 not restored correctly. Expected '%s', got '%s'", content2, restored2)
	}
}

func TestSelectiveRestore(t *testing.T) {
	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")

	content1 := []byte("content 1")
	content2 := []byte("content 2")

	os.WriteFile(testFile1, content1, 0644)
	os.WriteFile(testFile2, content2, 0644)

	// Create backup
	manifest, err := CreateBackup([]string{testFile1, testFile2}, "selective test")
	if err != nil {
		t.Fatal(err)
	}

	// Modify both files
	os.WriteFile(testFile1, []byte("modified 1"), 0644)
	os.WriteFile(testFile2, []byte("modified 2"), 0644)

	// Restore only file 1
	if err := RestoreBackup(manifest.ID, []string{testFile1}); err != nil {
		t.Fatal(err)
	}

	// Verify file 1 was restored
	restored1, _ := os.ReadFile(testFile1)
	if string(restored1) != string(content1) {
		t.Error("File 1 not restored")
	}

	// Verify file 2 was NOT restored
	restored2, _ := os.ReadFile(testFile2)
	if string(restored2) == string(content2) {
		t.Error("File 2 should not have been restored")
	}
}

func TestListBackups(t *testing.T) {
	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Initially no backups
	backups, err := ListBackups()
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups, got %d", len(backups))
	}

	// Create test file and backups
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	_, err = CreateBackup([]string{testFile}, "backup 1")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second) // Need full second for different timestamps

	_, err = CreateBackup([]string{testFile}, "backup 2")
	if err != nil {
		t.Fatal(err)
	}

	// List backups
	backups, err = ListBackups()
	if err != nil {
		t.Fatal(err)
	}

	if len(backups) != 2 {
		t.Fatalf("Expected 2 backups, got %d", len(backups))
	}

	// Should be sorted newest first
	if backups[0].Reason != "backup 2" {
		t.Error("Backups not sorted correctly")
	}
	if backups[1].Reason != "backup 1" {
		t.Error("Backups not sorted correctly")
	}
}

func TestGetBackupInfo(t *testing.T) {
	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	manifest, err := CreateBackup([]string{testFile}, "info test")
	if err != nil {
		t.Fatal(err)
	}

	// Get backup info
	info, err := GetBackupInfo(manifest.ID)
	if err != nil {
		t.Fatalf("GetBackupInfo failed: %v", err)
	}

	if info.ID != manifest.ID {
		t.Errorf("Expected ID %s, got %s", manifest.ID, info.ID)
	}
	if info.Reason != "info test" {
		t.Errorf("Expected reason 'info test', got '%s'", info.Reason)
	}
}

func TestDeleteBackup(t *testing.T) {
	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	manifest, err := CreateBackup([]string{testFile}, "delete test")
	if err != nil {
		t.Fatal(err)
	}

	// Verify backup exists
	_, err = GetBackupInfo(manifest.ID)
	if err != nil {
		t.Fatal("Backup should exist")
	}

	// Delete backup
	if err := DeleteBackup(manifest.ID); err != nil {
		t.Fatalf("DeleteBackup failed: %v", err)
	}

	// Verify backup no longer exists
	_, err = GetBackupInfo(manifest.ID)
	if err == nil {
		t.Error("Backup should not exist after deletion")
	}
}

func TestChecksumVerification(t *testing.T) {
	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	os.WriteFile(testFile, content, 0644)

	manifest, err := CreateBackup([]string{testFile}, "checksum test")
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt backup file
	entry := manifest.Files[0]
	os.WriteFile(entry.BackupPath, []byte("corrupted"), 0644)

	// Restore should fail due to checksum mismatch
	err = RestoreBackup(manifest.ID, nil)
	if err == nil {
		t.Error("Expected restore to fail with corrupted backup")
	}
}
