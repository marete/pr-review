package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBackupFile_NoFile tests that backing up a non-existent file does nothing
func TestBackupFile_NoFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "nonexistent.txt")

	err := backupFile(testFile)
	if err != nil {
		t.Errorf("backupFile() returned error for non-existent file: %v", err)
	}

	// Verify no backup was created
	backup := testFile + ".~1~"
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Errorf("Backup file was created for non-existent file")
	}
}

// TestBackupFile_FirstBackup tests creating the first backup
func TestBackupFile_FirstBackup(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := "original content"

	// Create the original file
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Backup the file
	err := backupFile(testFile)
	if err != nil {
		t.Errorf("backupFile() returned error: %v", err)
	}

	// Verify the backup was created
	backup := testFile + ".~1~"
	content, err := os.ReadFile(backup)
	if err != nil {
		t.Errorf("Backup file was not created: %v", err)
	}

	if string(content) != originalContent {
		t.Errorf("Backup content = %q, want %q", string(content), originalContent)
	}

	// Verify the original file no longer exists
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("Original file still exists after backup")
	}
}

// TestBackupFile_MultipleBackups tests creating multiple numbered backups
func TestBackupFile_MultipleBackups(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create and backup first version
	content1 := "version 1"
	if err := os.WriteFile(testFile, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := backupFile(testFile); err != nil {
		t.Fatalf("First backup failed: %v", err)
	}

	// Create and backup second version
	content2 := "version 2"
	if err := os.WriteFile(testFile, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := backupFile(testFile); err != nil {
		t.Fatalf("Second backup failed: %v", err)
	}

	// Create and backup third version
	content3 := "version 3"
	if err := os.WriteFile(testFile, []byte(content3), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := backupFile(testFile); err != nil {
		t.Fatalf("Third backup failed: %v", err)
	}

	// Verify all three backups exist with correct content
	backups := []struct {
		name    string
		content string
	}{
		{testFile + ".~1~", content1},
		{testFile + ".~2~", content2},
		{testFile + ".~3~", content3},
	}

	for _, b := range backups {
		content, err := os.ReadFile(b.name)
		if err != nil {
			t.Errorf("Backup %s was not created: %v", b.name, err)
			continue
		}
		if string(content) != b.content {
			t.Errorf("Backup %s content = %q, want %q", b.name, string(content), b.content)
		}
	}
}

// TestWriteReviewToFile_NewFile tests writing a review to a new file
func TestWriteReviewToFile_NewFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "review.md")
	reviewContent := "# Code Review\n\nThis is a test review."

	err := writeReviewToFile(testFile, reviewContent)
	if err != nil {
		t.Errorf("writeReviewToFile() returned error: %v", err)
	}

	// Verify the file was created with correct content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Review file was not created: %v", err)
	}

	if string(content) != reviewContent {
		t.Errorf("Review content = %q, want %q", string(content), reviewContent)
	}

	// Verify no backup was created
	backup := testFile + ".~1~"
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Errorf("Backup file was created for new file")
	}
}

// TestWriteReviewToFile_WithBackup tests writing a review when file exists
func TestWriteReviewToFile_WithBackup(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "review.md")
	oldReview := "# Old Review\n\nThis is the old review."
	newReview := "# New Review\n\nThis is the new review."

	// Create the original file
	if err := os.WriteFile(testFile, []byte(oldReview), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Write new review (should backup the old one)
	err := writeReviewToFile(testFile, newReview)
	if err != nil {
		t.Errorf("writeReviewToFile() returned error: %v", err)
	}

	// Verify the new review was written
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Review file was not created: %v", err)
	}
	if string(content) != newReview {
		t.Errorf("New review content = %q, want %q", string(content), newReview)
	}

	// Verify the backup was created with old content
	backup := testFile + ".~1~"
	backupContent, err := os.ReadFile(backup)
	if err != nil {
		t.Errorf("Backup file was not created: %v", err)
	}
	if string(backupContent) != oldReview {
		t.Errorf("Backup content = %q, want %q", string(backupContent), oldReview)
	}
}

// TestWriteReviewToFile_MultipleBackups tests multiple writes create multiple backups
func TestWriteReviewToFile_MultipleWrites(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "review.md")

	reviews := []string{
		"Review 1",
		"Review 2",
		"Review 3",
	}

	// Write each review
	for _, review := range reviews {
		if err := writeReviewToFile(testFile, review); err != nil {
			t.Fatalf("writeReviewToFile() failed: %v", err)
		}
	}

	// Verify the latest review is in the main file
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Review file was not created: %v", err)
	}
	if string(content) != reviews[2] {
		t.Errorf("Final review content = %q, want %q", string(content), reviews[2])
	}

	// Verify backups were created
	expectedBackups := []struct {
		name    string
		content string
	}{
		{testFile + ".~1~", reviews[0]},
		{testFile + ".~2~", reviews[1]},
	}

	for _, b := range expectedBackups {
		content, err := os.ReadFile(b.name)
		if err != nil {
			t.Errorf("Backup %s was not created: %v", b.name, err)
			continue
		}
		if string(content) != b.content {
			t.Errorf("Backup %s content = %q, want %q", b.name, string(content), b.content)
		}
	}
}
