package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupGitTestHome(t *testing.T) string {
	tempHome, err := os.MkdirTemp("", "reshell-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)

	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(tempHome)
	})

	return tempHome
}

func TestGitVersionControl(t *testing.T) {
	tempHome := setupGitTestHome(t)
	configDir := filepath.Join(tempHome, ".config", "reshell")

	// 1. Test Initialization
	err := InitWorkspace()
	if err != nil {
		t.Fatalf("InitWorkspace failed: %v", err)
	}

	// Verify .git and .gitignore exist
	if _, err := os.Stat(filepath.Join(configDir, ".git")); os.IsNotExist(err) {
		t.Errorf("Expected .git folder to exist in %s", configDir)
	}
	if _, err := os.Stat(filepath.Join(configDir, ".gitignore")); os.IsNotExist(err) {
		t.Errorf("Expected .gitignore file to exist in %s", configDir)
	}

	// 2. Test Commit
	// Create a dummy config file
	testFilePath := filepath.Join(configDir, "aliases.toml")
	err = os.WriteFile(testFilePath, []byte("aliases = []"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err = CommitWorkspace("Create aliases.toml")
	if err != nil {
		t.Fatalf("CommitWorkspace failed: %v", err)
	}

	// 3. Test History
	history, err := GetHistory()
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	// Expecting 2 commits: Initial commit and Create aliases.toml
	if len(history) != 2 {
		t.Fatalf("expected 2 commits in history, got %d", len(history))
	}

	if !strings.Contains(history[0].Message, "Create aliases.toml") {
		t.Errorf("expected first commit message to contain 'Create aliases.toml', got %q", history[0].Message)
	}

	// 4. Test Reversion
	firstCommitHash := history[0].Hash
	initialCommitHash := history[1].Hash

	// Modify aliases.toml and commit
	err = os.WriteFile(testFilePath, []byte("aliases = [{name = 'test'}]"), 0644)
	if err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}
	_ = CommitWorkspace("Modify aliases.toml")

	// Verify modified content
	content, _ := os.ReadFile(testFilePath)
	if !strings.Contains(string(content), "test") {
		t.Fatalf("Expected file to contain modified alias")
	}

	// Revert to firstCommitHash
	err = RevertToCommit(firstCommitHash)
	if err != nil {
		t.Fatalf("RevertToCommit failed: %v", err)
	}

	// Verify content was reverted
	content, _ = os.ReadFile(testFilePath)
	if strings.Contains(string(content), "test") {
		t.Errorf("Expected file content to be reverted, but got: %s", string(content))
	}

	// Revert to initialCommitHash (before aliases.toml was created)
	err = RevertToCommit(initialCommitHash)
	if err != nil {
		t.Fatalf("RevertToCommit failed: %v", err)
	}

	// Verify aliases.toml was deleted during hard reset and clean
	if _, err := os.Stat(testFilePath); !os.IsNotExist(err) {
		t.Errorf("Expected aliases.toml to be deleted, but it still exists")
	}
}
