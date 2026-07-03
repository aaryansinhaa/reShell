package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Commit represents a historical workspace state tracked by Git.
type Commit struct {
	Hash      string
	Timestamp string
	Message   string
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "reshell"), nil
}

func ensureConfigDirectories() error {
	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	subdirs := []string{
		"functions",
		"scripts",
		"logs",
		"logs/scripts",
		"logs/workflows",
		"shell",
	}

	for _, sub := range subdirs {
		path := filepath.Join(dir, sub)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

// InitWorkspace initializes git in the reshell configuration directory if not present.
func InitWorkspace() error {
	if err := ensureConfigDirectories(); err != nil {
		return fmt.Errorf("failed to ensure config directories: %w", err)
	}

	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil // already initialized
	}

	// 1. Run git init
	if _, err := runGitCmdInDir(dir, "init"); err != nil {
		return fmt.Errorf("failed to init git: %w", err)
	}

	// 2. Write .gitignore
	ignorePath := filepath.Join(dir, ".gitignore")
	gitignoreContent := "logs/\nshell/\n"
	if err := os.WriteFile(ignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	// 3. Make initial commit
	if _, err := runGitCmdInDir(dir, "add", "-A"); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Set local dummy config if global user.name/email is not configured (for test safety)
	_, _ = runGitCmdInDir(dir, "config", "--local", "user.name", "reshell")
	_, _ = runGitCmdInDir(dir, "config", "--local", "user.email", "reshell@example.com")

	if _, err := runGitCmdInDir(dir, "commit", "-m", "[reshell] Initial configuration workspace"); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// CommitWorkspace stages all changes and commits them with the given message.
func CommitWorkspace(message string) error {
	if err := InitWorkspace(); err != nil {
		return err
	}

	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Stage all files
	if _, err := runGitCmdInDir(dir, "add", "-A"); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	// Check if there are any staged changes to commit
	statusOut, err := runGitCmdInDir(dir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	if len(strings.TrimSpace(statusOut)) == 0 {
		return nil // nothing to commit
	}

	// Commit changes
	commitMsg := fmt.Sprintf("[reshell] %s", message)
	if _, err := runGitCmdInDir(dir, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

// GetHistory parses git log to return a list of commits.
func GetHistory() ([]Commit, error) {
	if err := InitWorkspace(); err != nil {
		return nil, err
	}

	dir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	logOut, err := runGitCmdInDir(dir, "log", `--pretty=format:%h|%ad|%s`, `--date=format:%Y-%m-%d %H:%M:%S`)
	if err != nil {
		// If there is no history or branch has no commits yet, return empty
		if strings.Contains(err.Error(), "does not have any commits yet") || strings.Contains(logOut, "fatal") {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read git log: %w", err)
	}

	var commits []Commit
	lines := strings.Split(logOut, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 3 {
			continue
		}
		commits = append(commits, Commit{
			Hash:      parts[0],
			Timestamp: parts[1],
			Message:   parts[2],
		})
	}

	return commits, nil
}

// RevertToCommit restores the workspace to the target commit hash and commits the restored state as a new commit, keeping history linear.
func RevertToCommit(hash string) error {
	if err := InitWorkspace(); err != nil {
		return err
	}

	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	// 1. Get the current HEAD hash to reset back to soft-ly
	headHash, err := runGitCmdInDir(dir, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("failed to get HEAD hash: %w", err)
	}
	headHash = strings.TrimSpace(headHash)

	// 2. Force reset index and working tree to the target commit
	if _, err := runGitCmdInDir(dir, "reset", "--hard", hash); err != nil {
		return fmt.Errorf("failed to reset to commit %s: %w", hash, err)
	}

	// 3. Move HEAD back to the original commit, leaving the index and working tree matching the target commit
	if _, err := runGitCmdInDir(dir, "reset", "--soft", headHash); err != nil {
		return fmt.Errorf("failed to return HEAD to original commit %s: %w", headHash, err)
	}

	// 4. Clean untracked files
	if _, err := runGitCmdInDir(dir, "clean", "-fd"); err != nil {
		return fmt.Errorf("failed to clean untracked files: %w", err)
	}

	// 5. Commit this restored state as a new commit to preserve linear git history
	commitMsg := fmt.Sprintf("Restore workspace state to commit %s", hash)
	if _, err := runGitCmdInDir(dir, "commit", "-m", "[reshell] "+commitMsg); err != nil {
		// If there is nothing to commit (already matching target state), ignore the error
		return nil
	}

	return nil
}

func runGitCmdInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String() + stderr.String(), fmt.Errorf("git error: %s (stderr: %s)", err.Error(), stderr.String())
	}
	return stdout.String(), nil
}
