package git

import (
	"bytes"
	"os/exec"
	"strings"
)

// GitConfig represents the global git configuration states.
type GitConfig struct {
	UserName   string            `toml:"username"`
	UserEmail  string            `toml:"email"`
	GpgSign    bool              `toml:"signing"`
	SigningKey string            `toml:"signing_key"`
	Aliases    map[string]string `toml:"aliases"`
}

// GetConfig reads the active global git settings.
func GetConfig() (*GitConfig, error) {
	username, _ := runGitCmd("config", "--global", "user.name")
	email, _ := runGitCmd("config", "--global", "user.email")
	signing, _ := runGitCmd("config", "--global", "commit.gpgsign")
	signingKey, _ := runGitCmd("config", "--global", "user.signingkey")

	aliasesMap := make(map[string]string)
	aliasLines, err := runGitCmd("config", "--global", "--get-regexp", "^alias\\.")
	if err == nil {
		lines := strings.Split(aliasLines, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				aliasName := strings.TrimPrefix(parts[0], "alias.")
				aliasesMap[aliasName] = parts[1]
			}
		}
	}

	return &GitConfig{
		UserName:   strings.TrimSpace(username),
		UserEmail:  strings.TrimSpace(email),
		GpgSign:    strings.TrimSpace(signing) == "true",
		SigningKey: strings.TrimSpace(signingKey),
		Aliases:    aliasesMap,
	}, nil
}

// ApplyConfig writes the GitConfig state to global ~/.gitconfig.
func ApplyConfig(cfg *GitConfig) error {
	if cfg.UserName != "" {
		if _, err := runGitCmd("config", "--global", "user.name", cfg.UserName); err != nil {
			return err
		}
	}
	if cfg.UserEmail != "" {
		if _, err := runGitCmd("config", "--global", "user.email", cfg.UserEmail); err != nil {
			return err
		}
	}

	signVal := "false"
	if cfg.GpgSign {
		signVal = "true"
	}
	if _, err := runGitCmd("config", "--global", "commit.gpgsign", signVal); err != nil {
		return err
	}

	if cfg.SigningKey != "" {
		if _, err := runGitCmd("config", "--global", "user.signingkey", cfg.SigningKey); err != nil {
			return err
		}
	} else {
		_ = exec.Command("git", "config", "--global", "--unset", "user.signingkey").Run()
	}

	// Apply aliases
	for name, value := range cfg.Aliases {
		if _, err := runGitCmd("config", "--global", "alias."+name, value); err != nil {
			return err
		}
	}

	return nil
}

func runGitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}
