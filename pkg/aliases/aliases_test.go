package aliases

import (
	"os"
	"reshell/pkg/config"
	"testing"
)

func TestAliasAddAndRemove(t *testing.T) {
	tempHome, err := os.MkdirTemp("", "reshell-test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Add new alias
	err = AddOrUpdate("gs", "git status", "git status shortcut", "all", true)
	if err != nil {
		t.Fatalf("AddOrUpdate failed: %v", err)
	}

	cfg, err := config.LoadAliases()
	if err != nil {
		t.Fatalf("LoadAliases failed: %v", err)
	}

	if len(cfg.Aliases) != 1 {
		t.Fatalf("Expected 1 alias, got %d", len(cfg.Aliases))
	}

	// Toggle state
	err = Toggle("gs")
	if err != nil {
		t.Fatalf("Toggle failed: %v", err)
	}

	cfg, _ = config.LoadAliases()
	if cfg.Aliases[0].Enabled {
		t.Error("Alias should be disabled after toggle")
	}

	// Remove alias
	err = Remove("gs")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	cfg, _ = config.LoadAliases()
	if len(cfg.Aliases) != 0 {
		t.Errorf("Expected 0 aliases, got %d", len(cfg.Aliases))
	}
}

func TestAliasConflictDetection(t *testing.T) {
	tempHome, err := os.MkdirTemp("", "reshell-test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Overwrite active alias test
	_ = AddOrUpdate("customcmd", "echo 1", "desc", "all", true)

	msg, conflict := DetectConflict("customcmd")
	if !conflict {
		t.Error("Expected conflict on existing active alias name")
	}
	if msg == "" {
		t.Error("Expected conflict message to be populated")
	}

	// System utility override (e.g. ls)
	_, conflictSystem := DetectConflict("ls")
	if !conflictSystem {
		t.Error("Expected conflict on system command override 'ls'")
	}
}

func TestAliasValidation(t *testing.T) {
	tempHome, err := os.MkdirTemp("", "reshell-test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Valid names
	validNames := []string{"git-status", "gs_123", "my.alias"}
	for _, n := range validNames {
		if !IsValidAliasName(n) {
			t.Errorf("expected alias name %q to be valid", n)
		}
		err = AddOrUpdate(n, "echo 1", "desc", "all", true)
		if err != nil {
			t.Errorf("AddOrUpdate failed for valid name %q: %v", n, err)
		}
	}

	// Invalid names
	invalidNames := []string{"git status", "gs;rm", "my=alias", "gs\n", "gs\\"}
	for _, n := range invalidNames {
		if IsValidAliasName(n) {
			t.Errorf("expected alias name %q to be invalid", n)
		}
		err = AddOrUpdate(n, "echo 1", "desc", "all", true)
		if err == nil {
			t.Errorf("expected AddOrUpdate to fail for invalid name %q", n)
		}
	}
}
