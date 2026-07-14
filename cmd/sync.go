package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reshell/pkg/config"
	"reshell/pkg/git"
	"reshell/pkg/marketplace"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

func init() {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize current developer environment configurations with a remote deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Display the warning prompt
			fmt.Println("Warning: Syncing configurations interacts with a remote Git repository, which might require Git authentication (SSH keys or HTTPS credentials).")
			fmt.Println("If authentication fails, the sync command will output the Git error to the terminal.")
			fmt.Print("Do you want to continue? (y/N): ")
			var proceed string
			fmt.Scanln(&proceed)
			proceed = strings.ToLower(strings.TrimSpace(proceed))
			if proceed != "y" && proceed != "yes" {
				fmt.Println("Sync operation aborted.")
				return nil
			}

			// Initialize local git workspace if not already done
			if err := git.InitWorkspace(); err != nil {
				return fmt.Errorf("failed to initialize local git workspace: %w", err)
			}

			// Load global configuration
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Get the active configuration directory
			configDir, err := config.GetConfigDir()
			if err != nil {
				return fmt.Errorf("failed to obtain active configuration directory: %w", err)
			}

			// Request remote deployment link if not set
			if cfg.RemoteSyncURL == "" {
				// Try to auto-detect from git remote sync-remote
				if out, err := runGit(configDir, "remote", "get-url", "sync-remote"); err == nil {
					detectedURL := strings.TrimSpace(out)
					if detectedURL != "" {
						cfg.RemoteSyncURL = detectedURL
						_ = config.SaveConfig(cfg)
					}
				}
			}

			if cfg.RemoteSyncURL == "" {
				fmt.Print("Enter remote deployment link (Git URL): ")
				var remoteURL string
				fmt.Scanln(&remoteURL)
				remoteURL = strings.TrimSpace(remoteURL)
				if remoteURL == "" {
					return fmt.Errorf("remote deployment link cannot be empty")
				}
				cfg.RemoteSyncURL = remoteURL
				if err := config.SaveConfig(cfg); err != nil {
					return fmt.Errorf("failed to save remote URL to configuration: %w", err)
				}
			}

			fmt.Printf("Syncing configuration workspace with remote: %s\n", cfg.RemoteSyncURL)

			// Configure/update git remote sync-remote
			_, _ = runGit(configDir, "remote", "remove", "sync-remote")
			if _, err := runGit(configDir, "remote", "add", "sync-remote", cfg.RemoteSyncURL); err != nil {
				return fmt.Errorf("failed to add git remote: %w", err)
			}

			// Check if remote is empty
			remoteHeads, err := runGit(configDir, "ls-remote", "--heads", "sync-remote")
			if err != nil {
				return fmt.Errorf("failed to connect to remote: %v. Please verify your authentication credentials", err)
			}

			// Fetch GitHub repository metrics (Stars, Forks, Last Updated, Open Issues) if applicable
			go marketplace.FetchAndCacheGitHubMetrics(cfg)

			// Clean local workspace first by committing any uncommitted changes
			_ = git.CommitWorkspace("Pre-sync local auto-commit")

			if strings.TrimSpace(remoteHeads) == "" {
				fmt.Println("Remote deployment is empty. Nothing to sync.")
				cfg.LastSync = time.Now().Format("2006-01-02 15:04:05")
				_ = config.SaveConfig(cfg)
				return nil
			}

			// Remote is not empty, fetch remote changes
			fmt.Println("Fetching remote changes...")
			if _, err := runGit(configDir, "fetch", "sync-remote"); err != nil {
				return fmt.Errorf("failed to fetch remote branch: %w", err)
			}

			// Get remote branch name
			remoteBranch, err := getRemoteBranch(configDir)
			if err != nil {
				return fmt.Errorf("failed to resolve remote branch name: %w", err)
			}
			fullRemoteBranch := "sync-remote/" + remoteBranch

			// Read remote files and perform programmatic merge
			fmt.Println("Programmatically merging configurations...")

			// Merge config.toml
			if err := mergeConfigTOML(configDir, fullRemoteBranch); err != nil {
				return fmt.Errorf("failed to merge config.toml: %w", err)
			}

			// Merge aliases.toml
			if err := mergeAliasesTOML(configDir, fullRemoteBranch); err != nil {
				return fmt.Errorf("failed to merge aliases.toml: %w", err)
			}

			// Merge env.toml
			if err := mergeEnvTOML(configDir, fullRemoteBranch); err != nil {
				return fmt.Errorf("failed to merge env.toml: %w", err)
			}

			// Merge snippets.toml
			if err := mergeSnippetsTOML(configDir, fullRemoteBranch); err != nil {
				return fmt.Errorf("failed to merge snippets.toml: %w", err)
			}

			// Merge workflows.toml
			if err := mergeWorkflowsTOML(configDir, fullRemoteBranch); err != nil {
				return fmt.Errorf("failed to merge workflows.toml: %w", err)
			}

			// Merge custom functions
			if err := mergeFunctionsDir(configDir, fullRemoteBranch); err != nil {
				return fmt.Errorf("failed to merge functions directory: %w", err)
			}

			// Merge library scripts
			if err := mergeScriptsDir(configDir, fullRemoteBranch); err != nil {
				return fmt.Errorf("failed to merge scripts directory: %w", err)
			}

			// Merge README.md
			_ = mergeReadme(configDir, fullRemoteBranch)

			// Stage and commit the programmatically merged changes
			if _, err := runGit(configDir, "add", "-A"); err != nil {
				return fmt.Errorf("failed to stage merged files: %w", err)
			}
			_, _ = runGit(configDir, "commit", "-m", "[reshell] programmatically merged remote changes")

			// Cache sync metadata
			cfg, _ = config.LoadConfig()
			cfg.LastSync = time.Now().Format("2006-01-02 15:04:05")
			_ = config.SaveConfig(cfg)

			fmt.Println("\nSync completed successfully! Local workspace updated with remote configurations.")
			fmt.Println("Please run 'reshell apply' to apply any updated aliases, scripts, or environment variables to your current session.")
			return nil
		},
	}

	rootCmd.AddCommand(syncCmd)
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}
	return stdout.String(), nil
}

func getLocalBranch(dir string) (string, error) {
	out, err := runGit(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func getRemoteBranch(dir string) (string, error) {
	_, _ = runGit(dir, "remote", "set-head", "sync-remote", "-a")
	out, err := runGit(dir, "symbolic-ref", "refs/remotes/sync-remote/HEAD")
	if err == nil {
		ref := strings.TrimSpace(out)
		branch := strings.TrimPrefix(ref, "refs/remotes/sync-remote/")
		if branch != ref && branch != "" {
			return branch, nil
		}
	}
	// Fallback verification
	if _, err := runGit(dir, "show-ref", "--verify", "refs/remotes/sync-remote/main"); err == nil {
		return "main", nil
	}
	if _, err := runGit(dir, "show-ref", "--verify", "refs/remotes/sync-remote/master"); err == nil {
		return "master", nil
	}
	local, err := getLocalBranch(dir)
	if err == nil {
		return local, nil
	}
	return "main", nil
}

// PROGRAMMATIC TOML MERGERS

func mergeConfigTOML(dir, remoteBranch string) error {
	remoteData, err := runGit(dir, "show", remoteBranch+":config.toml")
	if err != nil {
		return nil // Remote doesn't have config.toml yet
	}

	var remoteCfg config.Config
	if err := toml.Unmarshal([]byte(remoteData), &remoteCfg); err != nil {
		return err
	}

	localCfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Merge packages
	packageMap := make(map[string]bool)
	for _, p := range localCfg.Packages {
		packageMap[p] = true
	}
	for _, p := range remoteCfg.Packages {
		if !packageMap[p] {
			localCfg.Packages = append(localCfg.Packages, p)
			packageMap[p] = true
		}
	}

	// Merge marketplace
	marketplaceMap := make(map[string]bool)
	for _, m := range localCfg.Marketplace {
		marketplaceMap[m] = true
	}
	for _, m := range remoteCfg.Marketplace {
		if !marketplaceMap[m] {
			localCfg.Marketplace = append(localCfg.Marketplace, m)
			marketplaceMap[m] = true
		}
	}

	// Keep local user preferences (editor, theme, username, remote sync url)
	return config.SaveConfig(localCfg)
}

func mergeAliasesTOML(dir, remoteBranch string) error {
	remoteData, err := runGit(dir, "show", remoteBranch+":aliases.toml")
	if err != nil {
		return nil
	}

	var remoteCfg config.AliasConfig
	if err := toml.Unmarshal([]byte(remoteData), &remoteCfg); err != nil {
		return err
	}

	localCfg, err := config.LoadAliases()
	if err != nil {
		return err
	}

	localMap := make(map[string]config.Alias)
	for _, al := range localCfg.Aliases {
		localMap[al.Name] = al
	}

	for _, remoteAl := range remoteCfg.Aliases {
		localAl, exists := localMap[remoteAl.Name]
		if !exists {
			localCfg.Aliases = append(localCfg.Aliases, remoteAl)
		} else if localAl.Value != remoteAl.Value {
			choice := resolveConflict("alias", remoteAl.Name, localAl.Value, remoteAl.Value, "local workspace", "remote branch")
			switch choice {
			case 2: // override
				for i, al := range localCfg.Aliases {
					if al.Name == remoteAl.Name {
						localCfg.Aliases[i] = remoteAl
						break
					}
				}
			case 3: // rename
				newName := promptRename("alias")
				remoteAl.Name = newName
				localCfg.Aliases = append(localCfg.Aliases, remoteAl)
			}
		}
	}

	return config.SaveAliases(localCfg)
}

func mergeEnvTOML(dir, remoteBranch string) error {
	remoteData, err := runGit(dir, "show", remoteBranch+":env.toml")
	if err != nil {
		return nil
	}

	var remoteCfg config.EnvConfig
	if err := toml.Unmarshal([]byte(remoteData), &remoteCfg); err != nil {
		return err
	}

	localCfg, err := config.LoadEnv()
	if err != nil {
		return err
	}

	localMap := make(map[string]config.EnvVar)
	for _, ev := range localCfg.Variables {
		localMap[ev.Name] = ev
	}

	for _, remoteEv := range remoteCfg.Variables {
		localEv, exists := localMap[remoteEv.Name]
		if !exists {
			// Double check secrets before importing
			if config.IsSecret(remoteEv.Name, remoteEv.Value) {
				fmt.Printf("\n[WARNING] Potential secret detected in remote env variable %q:\n", remoteEv.Name)
				fmt.Println("  1) Skip importing [Recommended]")
				fmt.Println("  2) Import in plaintext anyway")
				var choice int
				for {
					fmt.Print("Select choice (1-2): ")
					_, err := fmt.Scanln(&choice)
					if err == nil && (choice == 1 || choice == 2) {
						break
					}
					var discard string
					_, _ = fmt.Scanln(&discard)
				}
				if choice == 1 {
					continue
				}
			}
			localCfg.Variables = append(localCfg.Variables, remoteEv)
		} else if localEv.Value != remoteEv.Value {
			choice := resolveConflict("environment variable", remoteEv.Name, localEv.Value, remoteEv.Value, "local workspace", "remote branch")
			switch choice {
			case 2: // override
				for i, ev := range localCfg.Variables {
					if ev.Name == remoteEv.Name {
						localCfg.Variables[i] = remoteEv
						break
					}
				}
			case 3: // rename
				newName := promptRename("environment variable")
				remoteEv.Name = newName
				localCfg.Variables = append(localCfg.Variables, remoteEv)
			}
		}
	}

	return config.SaveEnv(localCfg)
}

func mergeSnippetsTOML(dir, remoteBranch string) error {
	remoteData, err := runGit(dir, "show", remoteBranch+":snippets.toml")
	if err != nil {
		return nil
	}

	var remoteCfg config.SnippetConfig
	if err := toml.Unmarshal([]byte(remoteData), &remoteCfg); err != nil {
		return err
	}

	localCfg, err := config.LoadSnippets()
	if err != nil {
		return err
	}

	localMap := make(map[string]config.Snippet)
	for _, snip := range localCfg.Snippets {
		localMap[snip.Name] = snip
	}

	for _, remoteSnip := range remoteCfg.Snippets {
		localSnip, exists := localMap[remoteSnip.Name]
		if !exists {
			localCfg.Snippets = append(localCfg.Snippets, remoteSnip)
		} else if localSnip.Code != remoteSnip.Code {
			choice := resolveConflict("snippet", remoteSnip.Name, localSnip.Code, remoteSnip.Code, "local workspace", "remote branch")
			switch choice {
			case 2: // override
				for i, snip := range localCfg.Snippets {
					if snip.Name == remoteSnip.Name {
						localCfg.Snippets[i] = remoteSnip
						break
					}
				}
			case 3: // rename
				newName := promptRename("snippet")
				remoteSnip.Name = newName
				localCfg.Snippets = append(localCfg.Snippets, remoteSnip)
			}
		}
	}

	return config.SaveSnippets(localCfg)
}

func mergeWorkflowsTOML(dir, remoteBranch string) error {
	remoteData, err := runGit(dir, "show", remoteBranch+":workflows.toml")
	if err != nil {
		return nil
	}

	var remoteCfg config.WorkflowConfig
	if err := toml.Unmarshal([]byte(remoteData), &remoteCfg); err != nil {
		return err
	}

	localCfg, err := config.LoadWorkflows()
	if err != nil {
		return err
	}

	localMap := make(map[string]config.Workflow)
	for _, wf := range localCfg.Workflows {
		localMap[wf.Name] = wf
	}

	for _, remoteWf := range remoteCfg.Workflows {
		localWf, exists := localMap[remoteWf.Name]
		if !exists {
			localCfg.Workflows = append(localCfg.Workflows, remoteWf)
		} else {
			// Compare steps to detect changes
			stepsDiffer := len(localWf.Steps) != len(remoteWf.Steps)
			if !stepsDiffer {
				for i := range localWf.Steps {
					if localWf.Steps[i].Command != remoteWf.Steps[i].Command || localWf.Steps[i].Dir != remoteWf.Steps[i].Dir {
						stepsDiffer = true
						break
					}
				}
			}

			if stepsDiffer {
				localDesc := fmt.Sprintf("%d steps, e.g. %q", len(localWf.Steps), getFirstStepCmd(localWf.Steps))
				remoteDesc := fmt.Sprintf("%d steps, e.g. %q", len(remoteWf.Steps), getFirstStepCmd(remoteWf.Steps))
				choice := resolveConflict("workflow", remoteWf.Name, localDesc, remoteDesc, "local workspace", "remote branch")
				switch choice {
				case 2: // override
					for i, wf := range localCfg.Workflows {
						if wf.Name == remoteWf.Name {
							localCfg.Workflows[i] = remoteWf
							break
						}
					}
				case 3: // rename
					newName := promptRename("workflow")
					remoteWf.Name = newName
					localCfg.Workflows = append(localCfg.Workflows, remoteWf)
				}
			}
		}
	}

	return config.SaveWorkflows(localCfg)
}

func getFirstStepCmd(steps []config.WorkflowStep) string {
	if len(steps) > 0 {
		return steps[0].Command
	}
	return ""
}

// DIRECTORY FILE MERGERS

func mergeFunctionsDir(configDir, remoteBranch string) error {
	localFuncsDir := filepath.Join(configDir, "functions")
	_ = os.MkdirAll(localFuncsDir, 0700)

	// List remote function files
	remoteFilesList, err := runGit(configDir, "ls-tree", "-r", "--name-only", remoteBranch)
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(remoteFilesList, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "functions/") {
			continue
		}

		filename := filepath.Base(line)
		ext := filepath.Ext(filename)
		if ext != ".sh" && ext != ".fish" {
			continue
		}
		funcName := strings.TrimSuffix(filename, ext)

		remoteCode, err := runGit(configDir, "show", remoteBranch+":"+line)
		if err != nil {
			continue
		}

		localFilePath := filepath.Join(localFuncsDir, filename)
		localBytes, localErr := os.ReadFile(localFilePath)

		if localErr != nil {
			// Local doesn't have it, write it directly
			_ = os.WriteFile(localFilePath, []byte(remoteCode), 0700)
		} else if strings.TrimSpace(string(localBytes)) != strings.TrimSpace(remoteCode) {
			// Content conflict
			choice := resolveConflict("function script", funcName, string(localBytes), remoteCode, "local workspace", "remote branch")
			switch choice {
			case 2: // override
				_ = os.WriteFile(localFilePath, []byte(remoteCode), 0700)
			case 3: // rename
				newName := promptRename("function script")
				newFilePath := filepath.Join(localFuncsDir, newName+ext)
				_ = os.WriteFile(newFilePath, []byte(remoteCode), 0700)
			}
		}
	}

	return nil
}

func mergeScriptsDir(configDir, remoteBranch string) error {
	localScriptsDir := filepath.Join(configDir, "scripts")
	_ = os.MkdirAll(localScriptsDir, 0700)

	// List remote script files
	remoteFilesList, err := runGit(configDir, "ls-tree", "-r", "--name-only", remoteBranch)
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(remoteFilesList, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "scripts/") {
			continue
		}

		relPath := strings.TrimPrefix(line, "scripts/")
		filename := filepath.Base(relPath)
		ext := filepath.Ext(filename)
		if ext != ".sh" {
			continue
		}
		scriptName := strings.TrimSuffix(filename, ext)
		category := filepath.Dir(relPath)
		if category == "." {
			category = "general"
		}

		remoteCode, err := runGit(configDir, "show", remoteBranch+":"+line)
		if err != nil {
			continue
		}

		localFilePath := filepath.Join(localScriptsDir, category, filename)
		_ = os.MkdirAll(filepath.Dir(localFilePath), 0700)

		localBytes, localErr := os.ReadFile(localFilePath)

		if localErr != nil {
			// Local doesn't have it, write it
			_ = os.WriteFile(localFilePath, []byte(remoteCode), 0700)
		} else if strings.TrimSpace(string(localBytes)) != strings.TrimSpace(remoteCode) {
			// Content conflict
			choice := resolveConflict("library script", category+"/"+scriptName, string(localBytes), remoteCode, "local workspace", "remote branch")
			switch choice {
			case 2: // override
				_ = os.WriteFile(localFilePath, []byte(remoteCode), 0700)
			case 3: // rename
				newName := promptRename("library script")
				newFilePath := filepath.Join(localScriptsDir, category, newName+ext)
				_ = os.WriteFile(newFilePath, []byte(remoteCode), 0700)
			}
		}
	}

	return nil
}

func mergeReadme(dir, remoteBranch string) error {
	for _, name := range []string{"README.md", "readme.md", "README"} {
		remoteData, err := runGit(dir, "show", remoteBranch+":"+name)
		if err == nil {
			localPath := filepath.Join(dir, name)
			return os.WriteFile(localPath, []byte(remoteData), 0600)
		}
	}
	return nil
}
