package symlink

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("create new symlink", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source.txt")
		target := filepath.Join(tmpDir, "target.txt")

		// Create source file
		if err := os.WriteFile(source, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create symlink
		result, err := CreateSymlink(source, target, false)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}

		if result.Status != LinkStatusSuccess {
			t.Errorf("Status = %v, want %v", result.Status, LinkStatusSuccess)
		}

		// Verify symlink was created
		linkDest, err := os.Readlink(target)
		if err != nil {
			t.Fatalf("failed to read symlink: %v", err)
		}

		if linkDest != source {
			t.Errorf("symlink points to %s, want %s", linkDest, source)
		}
	})

	t.Run("dry run mode", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source2.txt")
		target := filepath.Join(tmpDir, "target2.txt")

		// Create source file
		if err := os.WriteFile(source, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create symlink in dry-run mode
		result, err := CreateSymlink(source, target, true)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}

		if result.Status != LinkStatusSuccess {
			t.Errorf("Status = %v, want %v", result.Status, LinkStatusSuccess)
		}

		// Verify symlink was NOT created
		if _, err := os.Lstat(target); !os.IsNotExist(err) {
			t.Error("symlink should not exist in dry-run mode")
		}
	})

	t.Run("already linked correctly", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source3.txt")
		target := filepath.Join(tmpDir, "target3.txt")

		// Create source file
		if err := os.WriteFile(source, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create symlink manually
		if err := os.Symlink(source, target); err != nil {
			t.Fatal(err)
		}

		// Try to create again
		result, err := CreateSymlink(source, target, false)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}

		if result.Status != LinkStatusAlreadyLinked {
			t.Errorf("Status = %v, want %v", result.Status, LinkStatusAlreadyLinked)
		}
	})

	t.Run("conflict with existing file", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source4.txt")
		target := filepath.Join(tmpDir, "target4.txt")

		// Create source file
		if err := os.WriteFile(source, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create conflicting file at target
		if err := os.WriteFile(target, []byte("conflict"), 0644); err != nil {
			t.Fatal(err)
		}

		// Try to create symlink
		result, err := CreateSymlink(source, target, false)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}

		if result.Status != LinkStatusConflict {
			t.Errorf("Status = %v, want %v", result.Status, LinkStatusConflict)
		}
	})

	t.Run("conflict with existing symlink to different location", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source5.txt")
		other := filepath.Join(tmpDir, "other.txt")
		target := filepath.Join(tmpDir, "target5.txt")

		// Create source file
		if err := os.WriteFile(source, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create other file
		if err := os.WriteFile(other, []byte("other"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create symlink to other
		if err := os.Symlink(other, target); err != nil {
			t.Fatal(err)
		}

		// Try to create symlink to source
		result, err := CreateSymlink(source, target, false)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}

		if result.Status != LinkStatusConflict {
			t.Errorf("Status = %v, want %v", result.Status, LinkStatusConflict)
		}
	})

	t.Run("create parent directories", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source6.txt")
		target := filepath.Join(tmpDir, "deeply", "nested", "target.txt")

		// Create source file
		if err := os.WriteFile(source, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create symlink (should create parent dirs)
		result, err := CreateSymlink(source, target, false)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}

		if result.Status != LinkStatusSuccess {
			t.Errorf("Status = %v, want %v", result.Status, LinkStatusSuccess)
		}

		// Verify symlink exists
		if _, err := os.Lstat(target); err != nil {
			t.Errorf("symlink should exist: %v", err)
		}
	})

	t.Run("link directory", func(t *testing.T) {
		source := filepath.Join(tmpDir, "sourcedir")
		target := filepath.Join(tmpDir, "targetdir")

		// Create source directory
		if err := os.MkdirAll(source, 0755); err != nil {
			t.Fatal(err)
		}

		// Create symlink to directory
		result, err := CreateSymlink(source, target, false)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}

		if result.Status != LinkStatusSuccess {
			t.Errorf("Status = %v, want %v", result.Status, LinkStatusSuccess)
		}

		if !result.IsDir {
			t.Error("IsDir should be true")
		}

		// Verify symlink was created and points to source
		linkDest, err := os.Readlink(target)
		if err != nil {
			t.Fatalf("failed to read symlink: %v", err)
		}

		if linkDest != source {
			t.Errorf("symlink points to %s, want %s", linkDest, source)
		}
	})
}

func TestIsLinked(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("correctly linked", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source_link.txt")
		target := filepath.Join(tmpDir, "target_link.txt")

		// Create source and symlink
		os.WriteFile(source, []byte("test"), 0644)
		os.Symlink(source, target)

		isLinked, err := IsLinked(source, target)
		if err != nil {
			t.Fatalf("IsLinked() error = %v", err)
		}

		if !isLinked {
			t.Error("should be linked")
		}
	})

	t.Run("not linked - target doesn't exist", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source_nolink.txt")
		target := filepath.Join(tmpDir, "target_nolink.txt")

		// Only create source
		os.WriteFile(source, []byte("test"), 0644)

		isLinked, err := IsLinked(source, target)
		if err != nil {
			t.Fatalf("IsLinked() error = %v", err)
		}

		if isLinked {
			t.Error("should not be linked")
		}
	})

	t.Run("not linked - target is regular file", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source_file.txt")
		target := filepath.Join(tmpDir, "target_file.txt")

		// Create both as regular files
		os.WriteFile(source, []byte("test"), 0644)
		os.WriteFile(target, []byte("test"), 0644)

		isLinked, err := IsLinked(source, target)
		if err != nil {
			t.Fatalf("IsLinked() error = %v", err)
		}

		if isLinked {
			t.Error("should not be linked")
		}
	})

	t.Run("not linked - symlink points elsewhere", func(t *testing.T) {
		source := filepath.Join(tmpDir, "source_wrong.txt")
		other := filepath.Join(tmpDir, "other_wrong.txt")
		target := filepath.Join(tmpDir, "target_wrong.txt")

		// Create files and symlink to other
		os.WriteFile(source, []byte("test"), 0644)
		os.WriteFile(other, []byte("other"), 0644)
		os.Symlink(other, target)

		isLinked, err := IsLinked(source, target)
		if err != nil {
			t.Fatalf("IsLinked() error = %v", err)
		}

		if isLinked {
			t.Error("should not be linked")
		}
	})
}

func TestWalkAndLink(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("link single file", func(t *testing.T) {
		source := filepath.Join(tmpDir, "single.txt")
		target := filepath.Join(tmpDir, "link_single.txt")

		// Create source file
		os.WriteFile(source, []byte("test"), 0644)

		results, err := WalkAndLink(source, target, false)
		if err != nil {
			t.Fatalf("WalkAndLink() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		if results[0].Status != LinkStatusSuccess {
			t.Errorf("Status = %v, want %v", results[0].Status, LinkStatusSuccess)
		}

		// Verify symlink
		isLinked, _ := IsLinked(source, target)
		if !isLinked {
			t.Error("file should be linked")
		}
	})

	t.Run("link directory contents", func(t *testing.T) {
		sourceDir := filepath.Join(tmpDir, "sourcedir_walk")
		targetDir := filepath.Join(tmpDir, "targetdir_walk")

		// Create source structure
		os.MkdirAll(sourceDir, 0755)
		os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("1"), 0644)
		os.WriteFile(filepath.Join(sourceDir, "file2.txt"), []byte("2"), 0644)
		os.MkdirAll(filepath.Join(sourceDir, "subdir"), 0755)
		os.WriteFile(filepath.Join(sourceDir, "subdir", "file3.txt"), []byte("3"), 0644)

		results, err := WalkAndLink(sourceDir, targetDir, false)
		if err != nil {
			t.Fatalf("WalkAndLink() error = %v", err)
		}

		// Should have linked 3 files (file1, file2, subdir/file3)
		successCount := 0
		for _, r := range results {
			if r.Status == LinkStatusSuccess {
				successCount++
			}
		}

		if successCount != 3 {
			t.Errorf("expected 3 successful links, got %d", successCount)
		}

		// Verify files are linked
		file1Target := filepath.Join(targetDir, "file1.txt")
		isLinked, _ := IsLinked(filepath.Join(sourceDir, "file1.txt"), file1Target)
		if !isLinked {
			t.Error("file1.txt should be linked")
		}

		file3Target := filepath.Join(targetDir, "subdir", "file3.txt")
		isLinked, _ = IsLinked(filepath.Join(sourceDir, "subdir", "file3.txt"), file3Target)
		if !isLinked {
			t.Error("subdir/file3.txt should be linked")
		}
	})

	t.Run("skip hidden files", func(t *testing.T) {
		sourceDir := filepath.Join(tmpDir, "sourcedir_hidden")
		targetDir := filepath.Join(tmpDir, "targetdir_hidden")

		// Create source with hidden files
		os.MkdirAll(sourceDir, 0755)
		os.WriteFile(filepath.Join(sourceDir, "visible.txt"), []byte("v"), 0644)
		os.WriteFile(filepath.Join(sourceDir, ".hidden.txt"), []byte("h"), 0644)

		results, err := WalkAndLink(sourceDir, targetDir, false)
		if err != nil {
			t.Fatalf("WalkAndLink() error = %v", err)
		}

		// Should only link visible.txt
		successCount := 0
		for _, r := range results {
			if r.Status == LinkStatusSuccess {
				successCount++
			}
		}

		if successCount != 1 {
			t.Errorf("expected 1 successful link, got %d", successCount)
		}

		// Verify hidden file was not linked
		hiddenTarget := filepath.Join(targetDir, ".hidden.txt")
		if _, err := os.Lstat(hiddenTarget); !os.IsNotExist(err) {
			t.Error("hidden file should not be linked")
		}
	})
}

func TestLinkTool(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("link tool with multiple links", func(t *testing.T) {
		// Create test structure
		sourceDir1 := filepath.Join(tmpDir, "tool", "config1")
		sourceDir2 := filepath.Join(tmpDir, "tool", "config2")
		targetDir1 := filepath.Join(tmpDir, "target1")
		targetDir2 := filepath.Join(tmpDir, "target2")

		os.MkdirAll(sourceDir1, 0755)
		os.MkdirAll(sourceDir2, 0755)

		// Create tool config
		tool := &ToolConfig{
			Name: "testtool",
			Links: []ResolvedLink{
				{
					Source: sourceDir1,
					Target: targetDir1,
					IsDir:  true,
				},
				{
					Source: sourceDir2,
					Target: targetDir2,
					IsDir:  true,
				},
			},
		}

		results, err := LinkTool(tool, false)
		if err != nil {
			t.Fatalf("LinkTool() error = %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}

		// Verify both links were created
		for _, result := range results {
			if result.Status != LinkStatusSuccess {
				t.Errorf("Link status = %v, want %v", result.Status, LinkStatusSuccess)
			}
		}
	})

	t.Run("link tool with file and directory", func(t *testing.T) {
		sourceFile := filepath.Join(tmpDir, "tool2", "file.txt")
		sourceDir := filepath.Join(tmpDir, "tool2", "dir")
		targetFile := filepath.Join(tmpDir, "target_file.txt")
		targetDir := filepath.Join(tmpDir, "target_dir")

		os.MkdirAll(filepath.Dir(sourceFile), 0755)
		os.WriteFile(sourceFile, []byte("test"), 0644)
		os.MkdirAll(sourceDir, 0755)

		tool := &ToolConfig{
			Name: "testtool2",
			Links: []ResolvedLink{
				{
					Source: sourceFile,
					Target: targetFile,
					IsDir:  false,
				},
				{
					Source: sourceDir,
					Target: targetDir,
					IsDir:  true,
				},
			},
		}

		results, err := LinkTool(tool, false)
		if err != nil {
			t.Fatalf("LinkTool() error = %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}

		// Verify links
		isLinked, _ := IsLinked(sourceFile, targetFile)
		if !isLinked {
			t.Error("file should be linked")
		}

		isLinked, _ = IsLinked(sourceDir, targetDir)
		if !isLinked {
			t.Error("directory should be linked")
		}
	})
}

func TestGetLinkStatus(t *testing.T) {
	tmpDir := t.TempDir()

	source1 := filepath.Join(tmpDir, "src1")
	target1 := filepath.Join(tmpDir, "tgt1")
	source2 := filepath.Join(tmpDir, "src2")
	target2 := filepath.Join(tmpDir, "tgt2")
	source3 := filepath.Join(tmpDir, "src3")
	target3 := filepath.Join(tmpDir, "tgt3")

	// Create sources
	os.MkdirAll(source1, 0755)
	os.MkdirAll(source2, 0755)
	os.MkdirAll(source3, 0755)

	// Create different scenarios
	os.Symlink(source1, target1) // Correctly linked
	os.MkdirAll(target2, 0755)   // Conflict (directory exists)
	// target3 doesn't exist (not linked)

	tool := &ToolConfig{
		Name: "statustest",
		Links: []ResolvedLink{
			{Source: source1, Target: target1, IsDir: true},
			{Source: source2, Target: target2, IsDir: true},
			{Source: source3, Target: target3, IsDir: true},
		},
	}

	status := GetLinkStatus(tool)

	if status[target1] != LinkStatusAlreadyLinked {
		t.Errorf("target1 status = %v, want %v", status[target1], LinkStatusAlreadyLinked)
	}

	if status[target2] != LinkStatusConflict {
		t.Errorf("target2 status = %v, want %v", status[target2], LinkStatusConflict)
	}

	if status[target3] != LinkStatusSkipped {
		t.Errorf("target3 status = %v, want %v", status[target3], LinkStatusSkipped)
	}
}

