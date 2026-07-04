package aliases

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reshell/pkg/config"
	"strings"
)

// IsValidAliasName verifies if the alias name contains only alphanumeric characters, underscores, hyphens, and periods.
func IsValidAliasName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

// AddOrUpdate creates or updates an alias.
func AddOrUpdate(name, value, desc, shell string, enabled bool) error {
	if !IsValidAliasName(name) {
		return fmt.Errorf("invalid alias name %q: must contain only alphanumeric characters, underscores, hyphens, or periods", name)
	}

	cfg, err := config.LoadAliases()
	if err != nil {
		return err
	}

	found := false
	for i, al := range cfg.Aliases {
		if al.Name == name {
			cfg.Aliases[i].Value = value
			cfg.Aliases[i].Description = desc
			cfg.Aliases[i].Shell = shell
			cfg.Aliases[i].Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		cfg.Aliases = append(cfg.Aliases, config.Alias{
			Name:        name,
			Value:       value,
			Description: desc,
			Shell:       shell,
			Enabled:     enabled,
		})
	}

	return config.SaveAliases(cfg)
}

// Remove deletes an alias.
func Remove(name string) error {
	cfg, err := config.LoadAliases()
	if err != nil {
		return err
	}

	newAliases := make([]config.Alias, 0, len(cfg.Aliases))
	found := false
	for _, al := range cfg.Aliases {
		if al.Name == name {
			found = true
			continue
		}
		newAliases = append(newAliases, al)
	}

	if !found {
		return errors.New("alias not found")
	}

	cfg.Aliases = newAliases
	return config.SaveAliases(cfg)
}

// Toggle toggles the enabled state of an alias.
func Toggle(name string) error {
	cfg, err := config.LoadAliases()
	if err != nil {
		return err
	}

	found := false
	for i, al := range cfg.Aliases {
		if al.Name == name {
			cfg.Aliases[i].Enabled = !al.Enabled
			found = true
			break
		}
	}

	if !found {
		return errors.New("alias not found")
	}

	return config.SaveAliases(cfg)
}

// DetectConflict checks if an alias name collides with:
// 1. A system command (using exec.LookPath)
// 2. An existing custom function name (in ~/.config/reshell/functions/)
// 3. Another registered alias name
func DetectConflict(name string) (string, bool) {
	// 1. Check custom function files
	funcDir, err := config.GetFunctionsDir()
	if err == nil {
		files, err := os.ReadDir(funcDir)
		if err == nil {
			for _, file := range files {
				ext := filepath.Ext(file.Name())
				fName := strings.TrimSuffix(file.Name(), ext)
				if fName == name {
					return fmt.Sprintf("collides with custom function defined in '%s'", file.Name()), true
				}
			}
		}
	}

	// 2. Check system utilities
	if path, err := exec.LookPath(name); err == nil {
		return fmt.Sprintf("overrides system utility located at: %s", path), true
	}

	// 3. Check other active aliases
	cfg, err := config.LoadAliases()
	if err == nil {
		for _, al := range cfg.Aliases {
			if al.Name == name && al.Enabled {
				return fmt.Sprintf("collides with another active alias: '%s=%s'", al.Name, al.Value), true
			}
		}
	}

	return "", false
}
