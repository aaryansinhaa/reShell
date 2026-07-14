package marketplace

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reshell/pkg/aliases"
	"reshell/pkg/config"
	"reshell/pkg/env"
	"reshell/pkg/functions"
	"reshell/pkg/git"
	"reshell/pkg/snippets"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// MarketplaceManifest represents the schema of a marketplace pack's reshell.toml.
type MarketplaceManifest struct {
	Package struct {
		Name        string `toml:"name"`
		Description string `toml:"description"`
	} `toml:"package"`
	Aliases   []config.Alias    `toml:"aliases"`
	Variables []config.EnvVar   `toml:"variables"`
	Snippets  []config.Snippet  `toml:"snippets"`
	Workflows []config.Workflow `toml:"workflows"`
	Config    struct {
		Packages []string `toml:"packages"`
	} `toml:"config"`
}

// FetchManifest clones the git repository to a temporary directory, parses the reshell.toml manifest, and returns it.
// The caller is responsible for deleting the tempDir (using os.RemoveAll).
func FetchManifest(repoURL string) (*MarketplaceManifest, string, error) {
	// Normalize URL, e.g. github.com/username/repo -> https://github.com/username/repo
	fullURL := repoURL
	if !os.IsPathSeparator(repoURL[0]) && !strings.Contains(repoURL, "://") && !strings.HasPrefix(repoURL, "git@") {
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
	// 3. Merge environment variables with secret checks
	for _, v := range manifest.Variables {
		if config.IsSecret(v.Name, v.Value) {
			if config.IsTUI {
				// Skip silently in TUI to avoid blocking Bubble Tea interactive loop
				continue
			}
			fmt.Printf("\n[WARNING] Potential secret detected in marketplace pack environment variable %q:\n", v.Name)
			fmt.Println("  (Note: Plaintext storage in env.toml is discouraged.)")
			fmt.Println("  1) Skip importing [Recommended]")
			fmt.Println("  2) Import in plaintext anyway")
			var choice int
			for {
				fmt.Print("Select choice (1-2): ")
				_, scanErr := fmt.Scanln(&choice)
				if scanErr == nil && (choice == 1 || choice == 2) {
					break
				}
				if scanErr != nil {
					if scanErr.Error() == "EOF" {
						choice = 1
						break
					}
					var discard string
					_, _ = fmt.Scanln(&discard)
				}
				fmt.Println("Invalid choice. Enter 1 or 2.")
			}
			if choice == 1 {
				continue
			}
		}
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
					ext := filepath.Ext(nameWithoutExt)
					nameWithoutExt = nameWithoutExt[:len(nameWithoutExt)-len(ext)]
					if !config.IsValidName(nameWithoutExt) {
						return fmt.Errorf("security error: invalid custom function name: %q", nameWithoutExt)
					}
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
			rel, errRel := filepath.Rel(scriptsSourceDir, path)
			if errRel != nil || strings.Contains(rel, "..") {
				return fmt.Errorf("security error: path traversal detected: %s", rel)
			}
			category := filepath.Dir(rel)
			if category == "." {
				category = "marketplace"
			}
			name := info.Name()
			if filepath.Ext(name) == ".sh" {
				name = name[:len(name)-len(".sh")]
			}
			if !config.IsValidName(name) {
				return fmt.Errorf("security error: invalid script name: %q", name)
			}
			data, err := os.ReadFile(path)
			if err == nil {
				scriptsDir, errScripts := config.GetScriptsDir()
				if errScripts != nil {
					return errScripts
				}
				catDir, errCat := config.SafeJoin(scriptsDir, category)
				if errCat != nil {
					return errCat
				}
				_ = os.MkdirAll(catDir, 0700)
				scriptPath, errScriptPath := config.SafeJoin(catDir, info.Name())
				if errScriptPath != nil {
					return errScriptPath
				}
				_ = os.WriteFile(scriptPath, data, 0700)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// 9. Copy README.md
	configDir, errDir := config.GetConfigDir()
	if errDir == nil {
		for _, name := range []string{"README.md", "readme.md", "README"} {
			srcReadme := filepath.Join(tempDir, name)
			if _, err := os.Stat(srcReadme); err == nil {
				if body, err := os.ReadFile(srcReadme); err == nil {
					destReadme := filepath.Join(configDir, name)
					_ = os.WriteFile(destReadme, body, 0600)
					break
				}
			}
		}
	}

	// 10. Merge workflows
	if len(manifest.Workflows) > 0 {
		localCfg, err := config.LoadWorkflows()
		if err == nil {
			localMap := make(map[string]bool)
			for _, wf := range localCfg.Workflows {
				localMap[wf.Name] = true
			}
			for _, remoteWf := range manifest.Workflows {
				if !localMap[remoteWf.Name] {
					localCfg.Workflows = append(localCfg.Workflows, remoteWf)
				}
			}
			_ = config.SaveWorkflows(localCfg)
		}
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

	// 1. Initialize local git if not already initialized
	_ = git.InitWorkspace()

	// 2. Configure/update git remote sync-remote
	configDir, errDir := config.GetConfigDir()
	if errDir == nil {
		fullURL := repoURL
		if !strings.Contains(repoURL, "://") && !strings.HasPrefix(repoURL, "git@") {
			fullURL = "https://" + repoURL
		}
		// Run git remote remove/add
		cmdRemove := exec.Command("git", "remote", "remove", "sync-remote")
		cmdRemove.Dir = configDir
		_ = cmdRemove.Run()

		cmdAdd := exec.Command("git", "remote", "add", "sync-remote", fullURL)
		cmdAdd.Dir = configDir
		_ = cmdAdd.Run()

		// Run git fetch sync-remote in background or foreground
		cmdFetch := exec.Command("git", "fetch", "sync-remote")
		cmdFetch.Dir = configDir
		_ = cmdFetch.Run()
	}

	// Automatically track/link this remote repository URL for future sync runs
	cfg, errConfig := config.LoadConfig()
	if errConfig == nil {
		fullURL := repoURL
		if !strings.Contains(repoURL, "://") && !strings.HasPrefix(repoURL, "git@") {
			fullURL = "https://" + repoURL
		}
		cfg.RemoteSyncURL = fullURL
		_ = config.SaveConfig(cfg)
		go FetchAndCacheGitHubMetrics(cfg)
	}

	return manifest, nil
}

func ParseGitHubRepo(url string) (owner, repo string) {
	url = strings.TrimSuffix(url, ".git")
	if strings.Contains(url, "git@github.com:") {
		parts := strings.Split(url, "git@github.com:")
		if len(parts) > 1 {
			sub := strings.Split(parts[1], "/")
			if len(sub) >= 2 {
				return sub[0], sub[1]
			}
		}
	} else if strings.Contains(url, "github.com/") {
		parts := strings.Split(url, "github.com/")
		if len(parts) > 1 {
			sub := strings.Split(parts[1], "/")
			if len(sub) >= 2 {
				return sub[0], sub[1]
			}
		}
	}
	return "", ""
}

type GitHubRepoResponse struct {
	ForksCount      int    `json:"forks_count"`
	StargazersCount int    `json:"stargazers_count"`
	PushedAt        string `json:"pushed_at"`
	OpenIssuesCount int    `json:"open_issues_count"`
}

func FetchAndCacheGitHubMetrics(cfg *config.Config) {
	owner, repo := ParseGitHubRepo(cfg.RemoteSyncURL)
	if owner == "" || repo == "" {
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo), nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "reshell-sync-client")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var data GitHubRepoResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return
	}

	// Update cached configuration details
	cfg, err = config.LoadConfig()
	if err == nil {
		cfg.Forks = data.ForksCount
		cfg.Stars = data.StargazersCount
		cfg.OpenIssues = data.OpenIssuesCount
		if t, err := time.Parse(time.RFC3339, data.PushedAt); err == nil {
			cfg.LastUpdated = t.Local().Format("2006-01-02 15:04")
		} else {
			cfg.LastUpdated = data.PushedAt
		}
		_ = config.SaveConfig(cfg)
	}
}
