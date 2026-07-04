package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupProfilesTestHome(t *testing.T) string {
	tempHome, err := os.MkdirTemp("", "reshell-profiles-test-*")
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

func TestProfileManagement(t *testing.T) {
	tempHome := setupProfilesTestHome(t)
	baseDir := filepath.Join(tempHome, ".config", "reshell")

	// 1. GetActiveProfile defaults to "default"
	active, err := GetActiveProfile()
	if err != nil {
		t.Fatalf("GetActiveProfile failed: %v", err)
	}
	if active != "default" {
		t.Errorf("expected active profile to be 'default', got %q", active)
	}

	// GetConfigDir defaults to baseDir
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}
	if dir != baseDir {
		t.Errorf("expected config dir to be %q, got %q", baseDir, dir)
	}

	// 2. CreateProfile
	err = CreateProfile("work")
	if err != nil {
		t.Fatalf("CreateProfile failed: %v", err)
	}

	// Verify profile folders were created
	profileDir := filepath.Join(baseDir, "profiles", "work")
	if _, err := os.Stat(filepath.Join(profileDir, "functions")); os.IsNotExist(err) {
		t.Errorf("Expected functions subdir to exist under work profile")
	}

	// 3. ListProfiles contains default and work
	list, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 profiles in list, got %d", len(list))
	}
	if list[0] != "default" || list[1] != "work" {
		t.Errorf("unexpected profile list contents: %v", list)
	}

	// 4. SetActiveProfile to "work"
	err = SetActiveProfile("work")
	if err != nil {
		t.Fatalf("SetActiveProfile failed: %v", err)
	}

	active, _ = GetActiveProfile()
	if active != "work" {
		t.Errorf("expected active profile to be 'work', got %q", active)
	}

	// GetConfigDir points to profiles/work
	dir, _ = GetConfigDir()
	if dir != profileDir {
		t.Errorf("expected config dir to be %q, got %q", profileDir, dir)
	}

	// 5. DeleteProfile boundaries
	// Cannot delete active profile
	err = DeleteProfile("work")
	if err == nil || !strings.Contains(err.Error(), "cannot delete active profile") {
		t.Errorf("expected error deleting active profile, got: %v", err)
	}

	// Switch to default
	_ = SetActiveProfile("default")

	// Delete work
	err = DeleteProfile("work")
	if err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}

	// Verify folder is gone
	if _, err := os.Stat(profileDir); !os.IsNotExist(err) {
		t.Errorf("Expected work profile directory to be removed")
	}
}

func TestProfileValidation(t *testing.T) {
	_ = setupProfilesTestHome(t)

	// 1. Test IsValidProfileName
	validNames := []string{"work", "school-projects", "my_workspace_123"}
	for _, n := range validNames {
		if !IsValidProfileName(n) {
			t.Errorf("expected profile name %q to be valid", n)
		}
	}

	invalidNames := []string{"work/space", "school..projects", "../danger", "my profile", "work?", "default"}
	for _, n := range invalidNames {
		if n != "default" && IsValidProfileName(n) {
			t.Errorf("expected profile name %q to be invalid", n)
		}
	}

	// 2. Test CreateProfile with traversal
	err := CreateProfile("../traversal")
	if err == nil {
		t.Error("expected CreateProfile to fail with traversal path")
	}

	// 3. Test SetActiveProfile with traversal
	err = SetActiveProfile("../traversal")
	if err == nil {
		t.Error("expected SetActiveProfile to fail with traversal path")
	}

	// 4. Test DeleteProfile with traversal
	err = DeleteProfile("../traversal")
	if err == nil {
		t.Error("expected DeleteProfile to fail with traversal path")
	}
}
