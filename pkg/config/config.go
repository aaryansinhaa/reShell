package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"reshell/pkg/git"
)

// Config represents the global reshell configuration.
type Config struct {
	Packages    []string `toml:"packages"`
	Marketplace []string `toml:"marketplace"`
	Theme       string   `toml:"theme"`
	UserName    string   `toml:"username"`
	Editor      string   `toml:"editor"`
}

// Alias represents a single shell alias.
type Alias struct {
	Name        string `toml:"name"`
	Value       string `toml:"value"`
	Description string `toml:"description"`
	Shell       string `toml:"shell"` // "all", "bash", "zsh", "fish"
	Enabled     bool   `toml:"enabled"`
}

type AliasConfig struct {
	Aliases []Alias `toml:"aliases"`
}

// Snippet represents a code snippet.
type Snippet struct {
	Name        string   `toml:"name"`
	Code        string   `toml:"code"`
	Description string   `toml:"description"`
	Tags        []string `toml:"tags"`
	Language    string   `toml:"language"`
	Favorite    bool     `toml:"favorite"`
}

type SnippetConfig struct {
	Snippets []Snippet `toml:"snippets"`
}

// EnvVar represents an environment variable.
type EnvVar struct {
	Name        string `toml:"name"`
	Value       string `toml:"value"`
	Description string `toml:"description"`
	Enabled     bool   `toml:"enabled"`
}

type EnvConfig struct {
	Variables []EnvVar `toml:"variables"`
}

// WorkflowStep represents a single task within a workflow.
type WorkflowStep struct {
	Command string `toml:"command"`
	Dir     string `toml:"dir"`
	Comment string `toml:"comment"`
}

// Workflow represents a script workflow sequence.
type Workflow struct {
	Name        string         `toml:"name"`
	Description string         `toml:"description"`
	Steps       []WorkflowStep `toml:"steps"`
}

type WorkflowConfig struct {
	Workflows []Workflow `toml:"workflows"`
}

// GetConfigDir returns the active config directory path for reshell.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	baseDir := filepath.Join(home, ".config", "reshell")

	// Read active profile name
	activeProfileFile := filepath.Join(baseDir, "active_profile.txt")
	profileBytes, err := os.ReadFile(activeProfileFile)
	if err != nil || len(strings.TrimSpace(string(profileBytes))) == 0 {
		return baseDir, nil
	}

	profileName := strings.TrimSpace(string(profileBytes))
	if profileName == "default" || profileName == "" {
		return baseDir, nil
	}

	return filepath.Join(baseDir, "profiles", profileName), nil
}

// EnsureDirectories configures the directories required for reshell.
func EnsureDirectories() error {
	dir, err := GetConfigDir()
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

// LoadTOMLFile loads configuration data from a given filename inside the config dir.
func LoadTOMLFile(filename string, v interface{}) error {
	if err := EnsureDirectories(); err != nil {
		return err
	}

	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file does not exist, leave object empty (default initial state)
			return nil
		}
		return err
	}

	return toml.Unmarshal(data, v)
}

// SaveTOMLFile serializes data and saves it to a file inside the config dir.
func SaveTOMLFile(filename string, v interface{}) error {
	if err := EnsureDirectories(); err != nil {
		return err
	}

	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, filename)
	data, err := toml.Marshal(v)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	// Auto-commit configurations to version control
	_ = git.CommitWorkspace("Update config: " + filename)
	return nil
}

// LoadConfig loads the global config.toml.
func LoadConfig() (*Config, error) {
	var cfg Config
	err := LoadTOMLFile("config.toml", &cfg)
	return &cfg, err
}

// SaveConfig saves the global config.toml.
func SaveConfig(cfg *Config) error {
	return SaveTOMLFile("config.toml", cfg)
}

// LoadAliases loads aliases.toml.
func LoadAliases() (*AliasConfig, error) {
	var cfg AliasConfig
	err := LoadTOMLFile("aliases.toml", &cfg)
	return &cfg, err
}

// SaveAliases saves aliases.toml.
func SaveAliases(cfg *AliasConfig) error {
	return SaveTOMLFile("aliases.toml", cfg)
}

// LoadSnippets loads snippets.toml.
func LoadSnippets() (*SnippetConfig, error) {
	var cfg SnippetConfig
	err := LoadTOMLFile("snippets.toml", &cfg)
	return &cfg, err
}

// SaveSnippets saves snippets.toml.
func SaveSnippets(cfg *SnippetConfig) error {
	return SaveTOMLFile("snippets.toml", cfg)
}

// LoadEnv loads env.toml.
func LoadEnv() (*EnvConfig, error) {
	var cfg EnvConfig
	err := LoadTOMLFile("env.toml", &cfg)
	return &cfg, err
}

// SaveEnv saves env.toml.
func SaveEnv(cfg *EnvConfig) error {
	return SaveTOMLFile("env.toml", cfg)
}

// LoadWorkflows loads workflows.toml.
func LoadWorkflows() (*WorkflowConfig, error) {
	var cfg WorkflowConfig
	err := LoadTOMLFile("workflows.toml", &cfg)
	return &cfg, err
}

// SaveWorkflows saves workflows.toml.
func SaveWorkflows(cfg *WorkflowConfig) error {
	return SaveTOMLFile("workflows.toml", cfg)
}

// GetFunctionsDir returns the path to the custom functions script directory.
func GetFunctionsDir() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "functions"), nil
}

// GetScriptsDir returns the path to the script library directory.
func GetScriptsDir() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "scripts"), nil
}
