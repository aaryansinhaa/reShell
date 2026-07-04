package env

import (
	"errors"
	"fmt"
	"os"
	"reshell/pkg/config"
	"strings"
)

// IsValidEnvName verifies if the environment variable name matches ^[a-zA-Z_][a-zA-Z0-9_]*$.
func IsValidEnvName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return false
			}
		}
	}
	return true
}

// AddOrUpdate adds or updates an environment variable.
func AddOrUpdate(name, value, desc string, enabled bool) error {
	if !IsValidEnvName(name) {
		return fmt.Errorf("invalid environment variable name %q: must contain only alphanumeric characters or underscores, and cannot start with a digit", name)
	}

	cfg, err := config.LoadEnv()
	if err != nil {
		return err
	}

	found := false
	for i, v := range cfg.Variables {
		if v.Name == name {
			cfg.Variables[i].Value = value
			cfg.Variables[i].Description = desc
			cfg.Variables[i].Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		cfg.Variables = append(cfg.Variables, config.EnvVar{
			Name:        name,
			Value:       value,
			Description: desc,
			Enabled:     enabled,
		})
	}

	return config.SaveEnv(cfg)
}

// Remove deletes an environment variable.
func Remove(name string) error {
	cfg, err := config.LoadEnv()
	if err != nil {
		return err
	}

	newVars := make([]config.EnvVar, 0, len(cfg.Variables))
	found := false
	for _, v := range cfg.Variables {
		if v.Name == name {
			found = true
			continue
		}
		newVars = append(newVars, v)
	}

	if !found {
		return errors.New("environment variable not found")
	}

	cfg.Variables = newVars
	return config.SaveEnv(cfg)
}

// Toggle flips the enabled state of an environment variable.
func Toggle(name string) error {
	cfg, err := config.LoadEnv()
	if err != nil {
		return err
	}

	found := false
	for i, v := range cfg.Variables {
		if v.Name == name {
			cfg.Variables[i].Enabled = !v.Enabled
			found = true
			break
		}
	}

	if !found {
		return errors.New("environment variable not found")
	}

	return config.SaveEnv(cfg)
}

// ValidatePath checks if an environment variable's value refers to an existing directory or file.
func ValidatePath(value string) bool {
	expanded := os.ExpandEnv(value)
	_, err := os.Stat(expanded)
	return err == nil
}

// AddDirToPath appends a directory to the PATH variable definition in env.toml.
func AddDirToPath(dir string) error {
	cfg, err := config.LoadEnv()
	if err != nil {
		return err
	}

	found := false
	for i, v := range cfg.Variables {
		if v.Name == "PATH" {
			if !strings.Contains(v.Value, dir) {
				cfg.Variables[i].Value = dir + ":" + v.Value
			}
			cfg.Variables[i].Enabled = true
			found = true
			break
		}
	}

	if !found {
		cfg.Variables = append(cfg.Variables, config.EnvVar{
			Name:        "PATH",
			Value:       dir + ":$PATH",
			Description: "reshell bin path hooks",
			Enabled:     true,
		})
	}

	return config.SaveEnv(cfg)
}
