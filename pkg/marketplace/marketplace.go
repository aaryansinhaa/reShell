package marketplace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reshell/pkg/aliases"
	"reshell/pkg/config"
	"reshell/pkg/env"
	"reshell/pkg/functions"
	"reshell/pkg/snippets"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// MarketplaceManifest represents the schema of a marketplace pack's reshell.toml.
type MarketplaceManifest struct {
	Package struct {
		Name        string `toml:"name"`
		Description string `toml:"description"`
	} `toml:"package"`
	Aliases   []config.Alias   `toml:"aliases"`
	Variables []config.EnvVar  `toml:"variables"`
	Snippets  []config.Snippet `toml:"snippets"`
	Config    struct {
		Packages []string `toml:"packages"`
	} `toml:"config"`
}

// FetchManifest clones the git repository to a temporary directory, parses the reshell.toml manifest, and returns it.
// The caller is responsible for deleting the tempDir (using os.RemoveAll).
func FetchManifest(repoURL string) (*MarketplaceManifest, string, error) {
	// Normalize URL, e.g. github.com/username/repo -> https://github.com/username/repo
	fullURL := repoURL
	if !os.IsPathSeparator(repoURL[0]) && !strings.Contains(repoURL, "://") {
		fullURL = "https://" + repoURL
	}

	tempDir, err := os.MkdirTemp("", "reshell-marketplace-*")
	if err != nil {
		return nil, "", err
	}

	// 1. Clone repository
	cmd := exec.Command("git", "clone", "--depth", "1", fullURL, tempDir)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("failed to clone repository '%s': %w", repoURL, err)
	}

	// 2. Read reshell.toml manifest
	manifestPath := filepath.Join(tempDir, "reshell.toml")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("repository does not contain a reshell.toml: %w", err)
	}

	var manifest MarketplaceManifest
	if err := toml.Unmarshal(manifestData, &manifest); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("invalid reshell.toml format: %w", err)
	}

	return &manifest, tempDir, nil
}

// MergeManifest merges the configuration, functions, and scripts from the cloned repository.
func MergeManifest(manifest *MarketplaceManifest, tempDir string) error {
	// 3. Merge environment variables
	for _, v := range manifest.Variables {
		if err := env.AddOrUpdate(v.Name, v.Value, v.Description, v.Enabled); err != nil {
			return fmt.Errorf("failed to merge env var '%s': %w", v.Name, err)
		}
	}

	// 4. Merge aliases
	for _, al := range manifest.Aliases {
		if err := aliases.AddOrUpdate(al.Name, al.Value, al.Description, al.Shell, al.Enabled); err != nil {
			return fmt.Errorf("failed to merge alias '%s': %w", al.Name, err)
		}
	}

	// 5. Merge snippets
	for _, snip := range manifest.Snippets {
		if err := snippets.AddOrUpdate(snip.Name, snip.Code, snip.Description, snip.Tags, snip.Language, snip.Favorite); err != nil {
			return fmt.Errorf("failed to merge snippet '%s': %w", snip.Name, err)
		}
	}

	// 6. Merge packages
	if len(manifest.Config.Packages) > 0 {
		cfg, err := config.LoadConfig()
		if err == nil {
			// Append package if not already in global config list
			pkgMap := make(map[string]bool)
			for _, p := range cfg.Packages {
				pkgMap[p] = true
			}
			for _, newPkg := range manifest.Config.Packages {
				if !pkgMap[newPkg] {
					cfg.Packages = append(cfg.Packages, newPkg)
				}
			}
			_ = config.SaveConfig(cfg)
		}
	}

	// 7. Copy functions
	funcsSourceDir := filepath.Join(tempDir, "functions")
	if info, err := os.Stat(funcsSourceDir); err == nil && info.IsDir() {
		files, err := os.ReadDir(funcsSourceDir)
		if err == nil {
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				data, err := os.ReadFile(filepath.Join(funcsSourceDir, file.Name()))
				if err == nil {
					nameWithoutExt := filepath.Base(file.Name())
					nameWithoutExt = nameWithoutExt[:len(nameWithoutExt)-len(filepath.Ext(nameWithoutExt))]
					_ = functions.CreateOrUpdate(nameWithoutExt, string(data))
				}
			}
		}
	}

	// 8. Copy scripts
	scriptsSourceDir := filepath.Join(tempDir, "scripts")
	if info, err := os.Stat(scriptsSourceDir); err == nil && info.IsDir() {
		err = filepath.Walk(scriptsSourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(scriptsSourceDir, path)
			category := filepath.Dir(rel)
			if category == "." {
				category = "marketplace"
			}
			name := info.Name()
			if filepath.Ext(name) == ".sh" {
				name = name[:len(name)-len(".sh")]
			}
			data, err := os.ReadFile(path)
			if err == nil {
				scriptsDir, _ := config.GetScriptsDir()
				catDir := filepath.Join(scriptsDir, category)
				_ = os.MkdirAll(catDir, 0755)
				_ = os.WriteFile(filepath.Join(catDir, info.Name()), data, 0755)
			}
			return nil
		})
	}

	return nil
}

// Install clones the git repo, reads its reshell.toml manifest, and merges assets.
func Install(repoURL string) (*MarketplaceManifest, error) {
	manifest, tempDir, err := FetchManifest(repoURL)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	if err := MergeManifest(manifest, tempDir); err != nil {
		return nil, err
	}

	return manifest, nil
}
