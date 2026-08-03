package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitUtils(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change working directory to temp dir
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Test IsGitRepo when not a repo
	if IsGitRepo() {
		t.Errorf("Expected IsGitRepo to be false, got true")
	}

	// Initialize git repo
	runGit := func(args ...string) {
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, out)
		}
	}

	runGit("init")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test User")
	runGit("-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "Initial empty commit")

	// Test IsGitRepo when repo
	if !IsGitRepo() {
		t.Errorf("Expected IsGitRepo to be true, got false")
	}

	// Test HasStagedChanges and GetFilteredStagedDiff when no changes
	if HasStagedChanges() {
		t.Errorf("Expected HasStagedChanges to be false, got true")
	}

	// Create a file and stage it
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("hello world"), 0644)
	runGit("add", "test.txt")

	// Test HasStagedChanges when changes
	if !HasStagedChanges() {
		t.Errorf("Expected HasStagedChanges to be true, got false")
	}

	// Test GetFilteredStagedDiff
	diff, err := GetFilteredStagedDiff()
	if err != nil {
		t.Fatalf("GetFilteredStagedDiff failed: %v", err)
	}
	if diff == "" {
		t.Errorf("Expected diff to not be empty")
	}

	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	if branch == "" {
		t.Errorf("Expected branch to not be empty")
	}
}
