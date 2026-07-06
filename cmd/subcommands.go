package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reshell/pkg/aliases"
	"reshell/pkg/config"
	"reshell/pkg/env"
	"reshell/pkg/functions"
	"reshell/pkg/git"
	"reshell/pkg/marketplace"
	"reshell/pkg/packages"
	"reshell/pkg/scripts"
	"reshell/pkg/shell"
	"reshell/pkg/snippets"
	"reshell/pkg/templates"
	"reshell/pkg/workflows"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	// --- apply command ---
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Compile settings and apply to active shell profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if executable path is in PATH
			if exePath, err := os.Executable(); err == nil {
				exeDir := filepath.Dir(exePath)
				pathEnv := os.Getenv("PATH")
				paths := filepath.SplitList(pathEnv)
				inPath := false
				for _, p := range paths {
					if filepath.Clean(p) == filepath.Clean(exeDir) {
						inPath = true
						break
					}
				}
				if !inPath {
					fmt.Printf("reshell: The directory containing reshell (%s) is not in your PATH.\n", exeDir)
					fmt.Print("Would you like to add it to your reshell PATH configurations? (y/N): ")
					var response string
					fmt.Scanln(&response)
					response = strings.ToLower(strings.TrimSpace(response))
					if response == "y" || response == "yes" {
						if err := env.AddDirToPath(exeDir); err != nil {
							fmt.Printf("Failed to update PATH: %v\n", err)
						} else {
							fmt.Println("PATH configuration updated successfully.")
						}
					}
				}
			}

			if err := shell.Apply(); err != nil {
				return err
			}
			fmt.Println("reshell: Configuration applied successfully! Restart your terminal or source your profile to update.")
			return nil
		},
	}

	// --- clean command ---
	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove reshell integration block from active shell profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := shell.Clean(); err != nil {
				return err
			}
			fmt.Println("reshell: Integration block removed from shell profile.")
			return nil
		},
	}

	// --- snippet subcommands ---
	snippetCmd := &cobra.Command{Use: "snippet", Short: "Manage code snippets"}

	var snippetTags string
	var snippetLang string
	snippetAddCmd := &cobra.Command{
		Use:   "add <name> <code> [description]",
		Short: "Add or update a snippet",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			code := args[1]
			desc := ""
			if len(args) > 2 {
				desc = args[2]
			}
			lang := "bash"
			if snippetLang != "" {
				l := strings.TrimSpace(strings.ToLower(snippetLang))
				if !snippets.IsValidLanguage(l) {
					fmt.Printf("Warning: Language '%s' is not recognized. Defaulting to 'bash'.\n", l)
				} else {
					lang = l
				}
			}
			var tags []string
			if snippetTags != "" {
				parts := strings.Split(snippetTags, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						tags = append(tags, p)
					}
				}
			}
			return snippets.AddOrUpdate(name, code, desc, tags, lang, false)
		},
	}
	snippetAddCmd.Flags().StringVarP(&snippetTags, "tags", "t", "", "Comma-separated tags for the snippet")
	snippetAddCmd.Flags().StringVarP(&snippetLang, "lang", "l", "", "Language/highlighter syntax for the snippet (default: bash)")

	snippetListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all code snippets",
		RunE: func(cmd *cobra.Command, args []string) error {
			snips, err := snippets.Search("")
			if err != nil {
				return err
			}
			for _, s := range snips {
				fmt.Printf("- %s: %s\n  Code: %s\n", s.Name, s.Description, s.Code)
			}
			return nil
		},
	}

	snippetCopyCmd := &cobra.Command{
		Use:   "copy <name>",
		Short: "Copy snippet code to system clipboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return snippets.CopyToClipboard(args[0])
		},
	}

	snippetCmd.AddCommand(snippetAddCmd, snippetListCmd, snippetCopyCmd)

	// --- alias subcommands ---
	aliasCmd := &cobra.Command{Use: "alias", Short: "Manage command aliases"}

	aliasAddCmd := &cobra.Command{
		Use:   "add <name> <value> [description]",
		Short: "Add or update a command alias",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			val := args[1]
			desc := ""
			if len(args) > 2 {
				desc = args[2]
			}
			// Check conflict/override warnings
			if warn, isConflict := aliases.DetectConflict(name); isConflict {
				fmt.Printf("Warning: %s\n", warn)
			}
			return aliases.AddOrUpdate(name, val, desc, "all", true)
		},
	}

	aliasListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all command aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadAliases()
			if err != nil {
				return err
			}
			for _, al := range cfg.Aliases {
				status := "enabled"
				if !al.Enabled {
					status = "disabled"
				}
				fmt.Printf("- %s='%s' (%s, %s)\n", al.Name, al.Value, al.Description, status)
			}
			return nil
		},
	}

	aliasCmd.AddCommand(aliasAddCmd, aliasListCmd)

	// --- function subcommands ---
	functionCmd := &cobra.Command{Use: "function", Short: "Manage custom shell functions"}

	functionAddCmd := &cobra.Command{
		Use:   "add <name> <code>",
		Short: "Add or update a custom shell function",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return functions.CreateOrUpdate(args[0], args[1])
		},
	}

	functionListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all custom function names",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := functions.List()
			if err != nil {
				return err
			}
			for _, f := range list {
				fmt.Println("-", f)
			}
			return nil
		},
	}

	functionValidateCmd := &cobra.Command{
		Use:   "validate <name>",
		Short: "Dry-run check function shell syntax validity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := functions.Validate(args[0])
			if err != nil {
				fmt.Printf("Validation failed:\n%s\n", output)
				return err
			}
			fmt.Println("Function syntax is valid.")
			return nil
		},
	}

	functionCmd.AddCommand(functionAddCmd, functionListCmd, functionValidateCmd)

	// --- script subcommands ---
	scriptCmd := &cobra.Command{Use: "script", Short: "Manage script library"}

	scriptListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all library scripts",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := scripts.List()
			if err != nil {
				return err
			}
			for _, s := range list {
				fmt.Printf("- %s/%s (%s)\n  Params: %v\n", s.Category, s.Name, s.Description, s.Parameters)
			}
			return nil
		},
	}

	scriptRunCmd := &cobra.Command{
		Use:   "run <category> <name> [args...]",
		Short: "Execute a library script with arguments",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cat := args[0]
			name := args[1]
			runArgs := args[2:]

			stdout, stderr, exitCode, err := scripts.Execute(cat, name, runArgs)
			if stdout != "" {
				fmt.Printf("Stdout:\n%s", stdout)
			}
			if stderr != "" {
				fmt.Printf("Stderr:\n%s", stderr)
			}
			if err != nil {
				return fmt.Errorf("script exited with code %d: %w", exitCode, err)
			}
			return nil
		},
	}

	scriptCmd.AddCommand(scriptListCmd, scriptRunCmd)

	// --- workflow subcommands ---
	workflowCmd := &cobra.Command{Use: "workflow", Short: "Manage execution workflows"}

	workflowRunCmd := &cobra.Command{
		Use:   "run <name>",
		Short: "Run workflow sequence sequentially",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wf, err := workflows.Get(args[0])
			if err != nil {
				return err
			}

			statusChan := make(chan workflows.StepStatus)
			go workflows.Run(wf, statusChan)

			for status := range statusChan {
				if !status.Finished {
					fmt.Printf("Running step %d: %s... ", status.Index+1, status.Command)
				} else {
					if status.Error != nil {
						fmt.Printf("FAILED (%v)\n", status.Error)
						if status.Stderr != "" {
							fmt.Fprintf(os.Stderr, "Error output:\n%s\n", status.Stderr)
						}
						return fmt.Errorf("workflow execution aborted due to step failure")
					} else {
						fmt.Println("SUCCESS")
					}
				}
			}
			return nil
		},
	}

	workflowCmd.AddCommand(workflowRunCmd)

	// --- template command ---
	newCmd := &cobra.Command{
		Use:   "new <template> <name>",
		Short: "Generate new project template directory",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpl := args[0]
			name := args[1]
			err := templates.Generate(tmpl, name, ".")
			if err != nil {
				return err
			}
			fmt.Printf("Created project '%s' from template '%s' successfully.\n", name, tmpl)
			return nil
		},
	}

	// --- install command ---
	installCmd := &cobra.Command{
		Use:   "install [marketplace-url-or-package]",
		Short: "Install system packages or download marketplace modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// Marketplace package install
				target := args[0]
				fmt.Printf("Fetching marketplace package from '%s'...\n", target)
				manifest, tempDir, err := marketplace.FetchManifest(target)
				if err != nil {
					return err
				}
				defer os.RemoveAll(tempDir)

				fmt.Printf("\nMarketplace Package: %s\n", manifest.Package.Name)
				fmt.Printf("Description:         %s\n", manifest.Package.Description)
				fmt.Println("\nThis package will configure:")
				fmt.Printf(" - %d environment variables\n", len(manifest.Variables))
				fmt.Printf(" - %d command aliases\n", len(manifest.Aliases))
				fmt.Printf(" - %d code snippets\n", len(manifest.Snippets))
				fmt.Printf(" - %d system package dependencies\n", len(manifest.Config.Packages))

				funcsCount := 0
				funcsSourceDir := filepath.Join(tempDir, "functions")
				if files, err := os.ReadDir(funcsSourceDir); err == nil {
					for _, f := range files {
						if !f.IsDir() {
							funcsCount++
						}
					}
				}
				fmt.Printf(" - %d custom functions\n", funcsCount)

				scriptsCount := 0
				scriptsSourceDir := filepath.Join(tempDir, "scripts")
				_ = filepath.Walk(scriptsSourceDir, func(path string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() {
						scriptsCount++
					}
					return nil
				})
				fmt.Printf(" - %d library scripts\n", scriptsCount)

				fmt.Println("\nWARNING: Installing third-party configuration packs merges them directly into your local environment and copies scripts/functions that will run on your machine.")
				fmt.Print("Do you want to proceed with the installation? (y/N): ")
				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Installation aborted.")
					return nil
				}

				if err := marketplace.MergeManifest(manifest, tempDir); err != nil {
					return err
				}

				fmt.Printf("\nMarketplace package '%s' installed successfully!\n", manifest.Package.Name)
				fmt.Println("Summary of changes:")
				fmt.Printf(" - Environment Variables: %d configured\n", len(manifest.Variables))
				fmt.Printf(" - Command Aliases:       %d configured\n", len(manifest.Aliases))
				fmt.Printf(" - Code Snippets:          %d configured\n", len(manifest.Snippets))
				fmt.Printf(" - System Packages:       %d configured\n", len(manifest.Config.Packages))
				fmt.Printf(" - Custom Functions:      %d configured\n", funcsCount)
				fmt.Printf(" - Library Scripts:       %d configured\n", scriptsCount)
				fmt.Println("\nRun 'reshell apply' to bind new environments and configurations.")
				return nil
			}

			// Install standard package lists from config.toml
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			if len(cfg.Packages) == 0 {
				fmt.Println("No packages listed in configuration file.")
				return nil
			}

			osName, manager := packages.DetectOS()
			fmt.Printf("Detected OS: %s (Manager: %s)\n", osName, manager)

			// Determine if sudo elevation is needed
			needsSudo := manager == "apt" || manager == "dnf" || manager == "pacman"
			sudoPassword := ""
			if needsSudo {
				// Check if sudo credentials are cached
				sudoCached := false
				chkCmd := exec.Command("sudo", "-n", "true")
				if err := chkCmd.Run(); err == nil {
					sudoCached = true
				}

				if !sudoCached {
					fmt.Print("This operation requires superuser privileges. Enter sudo password: ")
					pwBytes, err := term.ReadPassword(int(syscall.Stdin))
					if err != nil {
						return fmt.Errorf("failed to read password: %w", err)
					}
					fmt.Println()
					sudoPassword = string(pwBytes)
				} else {
					fmt.Println("reshell: Cached sudo credentials detected. Bypassing password prompt.")
				}
			}

			for _, pkg := range cfg.Packages {
				if packages.IsInstalled(pkg) {
					fmt.Printf("[%s] Already installed.\n", pkg)
					continue
				}

				fmt.Printf("[%s] Installing... ", pkg)
				stdoutChan := make(chan string)
				// Run in background and collect prints
				go func() {
					for range stdoutChan {
						// we discard detailed output or can verbose it
					}
				}()

				err := packages.Install(pkg, manager, sudoPassword, stdoutChan)
				if err != nil {
					fmt.Printf("FAILED: %v\n", err)
				} else {
					fmt.Println("SUCCESS")
				}
			}

			return nil
		},
	}

	// --- env subcommands ---
	envCmd := &cobra.Command{Use: "env", Short: "Manage environment variables"}

	envAddCmd := &cobra.Command{
		Use:   "add <name> <value> [description]",
		Short: "Add or update an environment variable",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			val := args[1]
			desc := ""
			if len(args) > 2 {
				desc = args[2]
			}
			return env.AddOrUpdate(name, val, desc, true)
		},
	}

	envListCmd := &cobra.Command{
		Use:   "list",
		Short: "List environment variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadEnv()
			if err != nil {
				return err
			}
			for _, v := range cfg.Variables {
				status := "enabled"
				if !v.Enabled {
					status = "disabled"
				}
				fmt.Printf("- export %s=%q (%s, %s)\n", v.Name, v.Value, v.Description, status)
			}
			return nil
		},
	}

	envCmd.AddCommand(envAddCmd, envListCmd)

	// --- git subcommands ---
	gitCmd := &cobra.Command{
		Use:   "git",
		Short: "Manage Git configuration profile",
	}

	gitApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Synchronize Git configs into local global ~/.gitconfig",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := git.GetConfig()
			if err != nil {
				return err
			}
			// This reads and saves the configs to local git config files.
			// Let's ensure the user is updated.
			err = git.ApplyConfig(cfg)
			if err != nil {
				return err
			}
			fmt.Println("reshell: Git settings applied globally successfully!")
			return nil
		},
	}

	gitClearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear version control history for the active profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := git.ClearHistory()
			if err != nil {
				return err
			}
			fmt.Println("reshell: Version control history cleared successfully for the active profile!")
			return nil
		},
	}

	gitCmd.AddCommand(gitApplyCmd, gitClearCmd)

	// --- export and import commands ---
	exportCmd := &cobra.Command{
		Use:   "export <output-toml-path>",
		Short: "Export all configurations and scripts into a single reshell.toml file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dest := args[0]
			err := config.ExportConfig(dest)
			if err != nil {
				return err
			}
			fmt.Printf("reshell: Configuration exported successfully to '%s'\n", dest)
			return nil
		},
	}

	importCmd := &cobra.Command{
		Use:   "import <toml-path>",
		Short: "Import configurations and scripts from a reshell.toml manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			err := config.ImportConfig(src)
			if err != nil {
				return err
			}
			fmt.Printf("reshell: Configuration imported successfully from '%s'\n", src)
			return nil
		},
	}

	// --- setup/bootstrap command ---
	setupCmd := &cobra.Command{
		Use:     "setup [directory_path]",
		Aliases: []string{"bootstrap", "install-self"},
		Short:   "Install reshell binary globally and bootstrap shell configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			// 1. Prompt for profile name
			var profileName string
			fmt.Print("Enter profile name to import configurations into (default: \"default\"): ")
			_, _ = fmt.Scanln(&profileName)
			profileName = strings.TrimSpace(profileName)
			if profileName == "" {
				profileName = "default"
			}

			// Check if profile exists, if not create it
			profiles, err := config.ListProfiles()
			if err == nil {
				exists := false
				for _, p := range profiles {
					if p == profileName {
						exists = true
						break
					}
				}
				if !exists && profileName != "default" {
					if err := config.CreateProfile(profileName); err != nil {
						return fmt.Errorf("failed to create profile %q: %w", profileName, err)
					}
				}
			}
			if err := config.SetActiveProfile(profileName); err != nil {
				return fmt.Errorf("failed to activate profile %q: %w", profileName, err)
			}

			// 2. Scan targets
			var scanResults config.DiscoveryResults
			var scanErr error
			if len(args) > 0 {
				fmt.Printf("Scanning directory '%s' recursively for configurations...\n", args[0])
				scanResults, scanErr = config.WalkAndParse(args[0])
			} else {
				fmt.Println("No directory specified. Scanning default system profiles and ~/.config...")
				scanResults, scanErr = config.DiscoverDefault(home)
			}
			if scanErr != nil {
				return fmt.Errorf("failed to scan for configurations: %w", scanErr)
			}

			// 3. Load active config files for merging
			activeAliases, err := config.LoadAliases()
			if err != nil {
				return err
			}
			activeEnv, err := config.LoadEnv()
			if err != nil {
				return err
			}
			activeSnippets, err := config.LoadSnippets()
			if err != nil {
				return err
			}

			var importedAliasesCount int
			var importedEnvCount int
			var importedFuncsCount int
			var importedSnippetsCount int
			var skippedSecretsCount int

			// Merge Aliases
			existingAliases := make(map[string]config.Alias)
			for _, al := range activeAliases.Aliases {
				existingAliases[al.Name] = al
			}

			for _, parsed := range scanResults.Aliases {
				existing, found := existingAliases[parsed.Name]
				if !found {
					activeAliases.Aliases = append(activeAliases.Aliases, config.Alias{
						Name:        parsed.Name,
						Value:       parsed.Value,
						Description: fmt.Sprintf("Imported from %s", filepath.Base(parsed.Source)),
						Shell:       "all",
						Enabled:     true,
					})
					existingAliases[parsed.Name] = config.Alias{Name: parsed.Name, Value: parsed.Value}
					importedAliasesCount++
				} else {
					if existing.Value != parsed.Value {
						choice := resolveConflict("alias", parsed.Name, existing.Value, parsed.Value, "active profile config", filepath.Base(parsed.Source))
						switch choice {
						case 2: // override
							for i, al := range activeAliases.Aliases {
								if al.Name == parsed.Name {
									activeAliases.Aliases[i].Value = parsed.Value
									activeAliases.Aliases[i].Description = fmt.Sprintf("Imported from %s (overwrote previous)", filepath.Base(parsed.Source))
									break
								}
							}
							importedAliasesCount++
						case 3: // rename
							newName := promptRename("alias")
							activeAliases.Aliases = append(activeAliases.Aliases, config.Alias{
								Name:        newName,
								Value:       parsed.Value,
								Description: fmt.Sprintf("Imported from %s (renamed from %s)", filepath.Base(parsed.Source), parsed.Name),
								Shell:       "all",
								Enabled:     true,
							})
							importedAliasesCount++
						}
					}
				}
			}
			if err := config.SaveAliases(activeAliases); err != nil {
				return err
			}

			// Merge Environment Variables
			existingEnv := make(map[string]config.EnvVar)
			for _, v := range activeEnv.Variables {
				existingEnv[v.Name] = v
			}

			for _, parsed := range scanResults.EnvVars {
				// Check for secrets
				if config.IsSecret(parsed.Name, parsed.Value) {
					fmt.Printf("\n[WARNING] Potential secret detected in environment variable %q:\n", parsed.Name)
					fmt.Println("  (Note: Although reshell's Git repository is purely local and never pushed to any remote, plaintext storage is discouraged.)")
					fmt.Println("  1) Skip importing (keep in original profile) [Recommended]")
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
								choice = 1 // default to skip
								break
							}
							// Clear buffer
							var discard string
							_, _ = fmt.Scanln(&discard)
						}
						fmt.Println("Invalid choice. Enter 1 or 2.")
					}
					if choice == 1 {
						skippedSecretsCount++
						continue
					}
				}

				existing, found := existingEnv[parsed.Name]
				if !found {
					activeEnv.Variables = append(activeEnv.Variables, config.EnvVar{
						Name:        parsed.Name,
						Value:       parsed.Value,
						Description: fmt.Sprintf("Imported from %s", filepath.Base(parsed.Source)),
						Enabled:     true,
					})
					existingEnv[parsed.Name] = config.EnvVar{Name: parsed.Name, Value: parsed.Value}
					importedEnvCount++
				} else {
					if existing.Value != parsed.Value {
						choice := resolveConflict("environment variable", parsed.Name, existing.Value, parsed.Value, "active profile config", filepath.Base(parsed.Source))
						switch choice {
						case 2: // override
							for i, ev := range activeEnv.Variables {
								if ev.Name == parsed.Name {
									activeEnv.Variables[i].Value = parsed.Value
									activeEnv.Variables[i].Description = fmt.Sprintf("Imported from %s (overwrote previous)", filepath.Base(parsed.Source))
									break
								}
							}
							importedEnvCount++
						case 3: // rename
							newName := promptRename("environment variable")
							activeEnv.Variables = append(activeEnv.Variables, config.EnvVar{
								Name:        newName,
								Value:       parsed.Value,
								Description: fmt.Sprintf("Imported from %s (renamed from %s)", filepath.Base(parsed.Source), parsed.Name),
								Enabled:     true,
							})
							importedEnvCount++
						}
					}
				}
			}
			if err := config.SaveEnv(activeEnv); err != nil {
				return err
			}

			// Merge Snippets
			existingSnippets := make(map[string]config.Snippet)
			for _, s := range activeSnippets.Snippets {
				existingSnippets[s.Name] = s
			}

			for _, parsed := range scanResults.Snippets {
				existing, found := existingSnippets[parsed.Name]
				if !found {
					activeSnippets.Snippets = append(activeSnippets.Snippets, config.Snippet{
						Name:        parsed.Name,
						Code:        parsed.Code,
						Description: parsed.Description,
						Tags:        parsed.Tags,
						Favorite:    false,
					})
					existingSnippets[parsed.Name] = config.Snippet{Name: parsed.Name, Code: parsed.Code}
					importedSnippetsCount++
				} else {
					if existing.Code != parsed.Code {
						choice := resolveConflict("snippet", parsed.Name, existing.Code, parsed.Code, "active profile config", filepath.Base(parsed.Source))
						switch choice {
						case 2: // override
							for i, s := range activeSnippets.Snippets {
								if s.Name == parsed.Name {
									activeSnippets.Snippets[i].Code = parsed.Code
									activeSnippets.Snippets[i].Description = parsed.Description
									break
								}
							}
							importedSnippetsCount++
						case 3: // rename
							newName := promptRename("snippet")
							activeSnippets.Snippets = append(activeSnippets.Snippets, config.Snippet{
								Name:        newName,
								Code:        parsed.Code,
								Description: parsed.Description,
								Tags:        parsed.Tags,
								Favorite:    false,
							})
							importedSnippetsCount++
						}
					}
				}
			}
			if err := config.SaveSnippets(activeSnippets); err != nil {
				return err
			}

			// Merge Functions
			existingFuncsList, _ := functions.List()
			existingFuncs := make(map[string]bool)
			for _, fName := range existingFuncsList {
				existingFuncs[fName] = true
			}

			for _, parsed := range scanResults.Functions {
				_, found := existingFuncs[parsed.Name]
				if !found {
					_ = functions.CreateOrUpdate(parsed.Name, parsed.Code)
					existingFuncs[parsed.Name] = true
					importedFuncsCount++
				} else {
					existingCode, _, err := functions.Get(parsed.Name)
					if err == nil && strings.TrimSpace(existingCode) != strings.TrimSpace(parsed.Code) {
						choice := resolveConflict("function", parsed.Name, existingCode, parsed.Code, "active profile config", filepath.Base(parsed.Source))
						switch choice {
						case 2: // override
							_ = functions.CreateOrUpdate(parsed.Name, parsed.Code)
							importedFuncsCount++
						case 3: // rename
							newName := promptRename("function")
							_ = functions.CreateOrUpdate(newName, parsed.Code)
							importedFuncsCount++
						}
					}
				}
			}

			// Standard Setup steps:
			localBin := filepath.Join(home, ".local", "bin")
			if err := os.MkdirAll(localBin, 0755); err != nil {
				return fmt.Errorf("failed to create local bin directory: %w", err)
			}

			exePath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to locate current executable: %w", err)
			}

			targetPath := filepath.Join(localBin, "reshell")

			cleanExe, err := filepath.EvalSymlinks(exePath)
			if err != nil {
				cleanExe = filepath.Clean(exePath)
			}
			cleanTarget, err := filepath.EvalSymlinks(targetPath)
			if err != nil {
				cleanTarget = filepath.Clean(targetPath)
			}

			if cleanExe != cleanTarget {
				fmt.Printf("reshell: Copying binary to '%s'...\n", targetPath)
				_ = os.Remove(targetPath)
				if err := shell.CopyFile(exePath, targetPath); err != nil {
					return fmt.Errorf("failed to copy executable: %w", err)
				}
				if err := os.Chmod(targetPath, 0755); err != nil {
					return fmt.Errorf("failed to set executable permissions: %w", err)
				}
			} else {
				fmt.Println("reshell: Binary is already installed in the target directory.")
			}

			if err := env.AddDirToPath(localBin); err != nil {
				return fmt.Errorf("failed to add local bin to PATH: %w", err)
			}

			if err := shell.Apply(); err != nil {
				return fmt.Errorf("failed to apply shell configurations: %w", err)
			}

			fmt.Println("\n+--------------------------------------------------------------+")
			fmt.Println("|            reshell SETUP COMPLETED SUCCESSFULLY!              |")
			fmt.Println("+--------------------------------------------------------------+")
			fmt.Printf(" Target Profile:     %s\n", profileName)
			fmt.Printf(" Import Summary:\n")
			fmt.Printf("   - Aliases:        %d imported\n", importedAliasesCount)
			fmt.Printf("   - Env Variables:  %d imported (skipped %d secrets)\n", importedEnvCount, skippedSecretsCount)
			fmt.Printf("   - Custom Funcs:   %d imported\n", importedFuncsCount)
			fmt.Printf("   - Code Snippets:  %d imported\n", importedSnippetsCount)
			fmt.Printf(" Binary installed to: %s\n", targetPath)
			fmt.Println(" Directory added to: reshell PATH variables")
			fmt.Println(" Shell integration hooks applied to active shell profile.")
			fmt.Println(" Please restart your terminal or source your profile to start using 'reshell' globally.")
			return nil
		},
	}

	// --- profile subcommands ---
	profileCmd := &cobra.Command{Use: "profile", Short: "Manage isolated configuration profiles"}

	profileListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := config.ListProfiles()
			if err != nil {
				return err
			}
			active, err := config.GetActiveProfile()
			if err != nil {
				return err
			}
			for _, p := range list {
				if p == active {
					fmt.Printf("- %s (active)\n", p)
				} else {
					fmt.Printf("- %s\n", p)
				}
			}
			return nil
		},
	}

	profileCreateCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			err := config.CreateProfile(name)
			if err != nil {
				return err
			}
			fmt.Printf("Profile '%s' created successfully.\n", name)
			return nil
		},
	}

	profileSwitchCmd := &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch to a different configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			list, err := config.ListProfiles()
			if err != nil {
				return err
			}
			exists := false
			for _, p := range list {
				if p == name {
					exists = true
					break
				}
			}
			if !exists {
				return fmt.Errorf("profile '%s' does not exist", name)
			}

			err = config.SetActiveProfile(name)
			if err != nil {
				return err
			}

			// Compile new profile hooks
			if err := shell.Apply(); err != nil {
				return fmt.Errorf("failed to compile configurations for profile '%s': %w", name, err)
			}

			fmt.Printf("Switched active profile to '%s'. Restart your terminal or source your profile to apply changes.\n", name)
			return nil
		},
	}

	profileDeleteCmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			err := config.DeleteProfile(name)
			if err != nil {
				return err
			}
			fmt.Printf("Profile '%s' deleted successfully.\n", name)
			return nil
		},
	}

	profileCmd.AddCommand(profileListCmd, profileCreateCmd, profileSwitchCmd, profileDeleteCmd)

	// Add commands to rootCmd
	rootCmd.AddCommand(
		applyCmd,
		cleanCmd,
		setupCmd,
		snippetCmd,
		aliasCmd,
		functionCmd,
		scriptCmd,
		workflowCmd,
		newCmd,
		installCmd,
		envCmd,
		gitCmd,
		exportCmd,
		importCmd,
		profileCmd,
	)
}

func resolveConflict(itemType, name, oldVal, newVal, oldSrc, newSrc string) int {
	fmt.Printf("\nConflict detected for %s %q:\n", itemType, name)
	fmt.Printf("  1) Keep existing value: %q (from %s)\n", oldVal, oldSrc)
	fmt.Printf("  2) Override with new value:  %q (from %s)\n", newVal, newSrc)
	fmt.Printf("  3) Keep both (rename the new one)\n")
	fmt.Printf("  4) Skip this import\n")
	for {
		fmt.Print("Select choice (1-4): ")
		var choice int
		_, err := fmt.Scanln(&choice)
		if err == nil && choice >= 1 && choice <= 4 {
			return choice
		}
		if err != nil {
			if err.Error() == "EOF" {
				return 4 // Skip on EOF
			}
			// Clear stdin buffer on error
			var discard string
			_, _ = fmt.Scanln(&discard)
		}
		fmt.Println("Invalid choice. Please enter a number between 1 and 4.")
	}
}

func promptRename(itemType string) string {
	for {
		fmt.Printf("Enter new name for the %s: ", itemType)
		var newName string
		_, err := fmt.Scanln(&newName)
		if err != nil {
			if err.Error() == "EOF" {
				return "imported-" + itemType
			}
		}
		newName = strings.TrimSpace(newName)
		if err == nil && newName != "" {
			return newName
		}
		fmt.Println("Invalid name. Please try again.")
	}
}
