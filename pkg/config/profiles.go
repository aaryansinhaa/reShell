package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetActiveProfile returns the name of the currently active profile.
func GetActiveProfile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "default", err
	}
	baseDir := filepath.Join(home, ".config", "reshell")
	activeProfileFile := filepath.Join(baseDir, "active_profile.txt")

	data, err := os.ReadFile(activeProfileFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "default", nil
		}
		return "default", err
	}

	profile := strings.TrimSpace(string(data))
	if profile == "" {
		return "default", nil
	}
	return profile, nil
}

// IsValidProfileName verifies if the profile name contains only alphanumeric characters, underscores, and hyphens.
func IsValidProfileName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

// SetActiveProfile updates the active_profile.txt state file.
func SetActiveProfile(name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	baseDir := filepath.Join(home, ".config", "reshell")
	activeProfileFile := filepath.Join(baseDir, "active_profile.txt")

	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}

	if name != "default" && !IsValidProfileName(name) {
		return fmt.Errorf("invalid profile name: %s (must contain only alphanumeric characters, underscores, and hyphens)", name)
	}

	// Make sure the base config dir exists before writing
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(activeProfileFile, []byte(name), 0644)
}

// ListProfiles returns a list of all profiles including "default".
func ListProfiles() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Join(home, ".config", "reshell")
	profilesDir := filepath.Join(baseDir, "profiles")

	profiles := []string{"default"}

	files, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return profiles, nil
		}
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			profiles = append(profiles, file.Name())
		}
	}

	return profiles, nil
}

// CreateProfile creates directories for a new profile under profiles/<name>/.
func CreateProfile(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || name == "default" {
		return fmt.Errorf("invalid profile name: %s", name)
	}

	if !IsValidProfileName(name) {
		return fmt.Errorf("invalid profile name: %s (must contain only alphanumeric characters, underscores, and hyphens)", name)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	profileDir := filepath.Join(home, ".config", "reshell", "profiles", name)
	subdirs := []string{
		"functions",
		"scripts",
		"logs",
		"logs/scripts",
		"logs/workflows",
		"shell",
	}

	for _, sub := range subdirs {
		path := filepath.Join(profileDir, sub)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

// DeleteProfile recursively removes a profile directory (cannot delete active profile).
func DeleteProfile(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || name == "default" {
		return fmt.Errorf("cannot delete default profile")
	}

	if !IsValidProfileName(name) {
		return fmt.Errorf("invalid profile name: %s (must contain only alphanumeric characters, underscores, and hyphens)", name)
	}

	active, err := GetActiveProfile()
	if err != nil {
		return err
	}
	if active == name {
		return fmt.Errorf("cannot delete active profile: %s", name)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	profileDir := filepath.Join(home, ".config", "reshell", "profiles", name)
	return os.RemoveAll(profileDir)
}
