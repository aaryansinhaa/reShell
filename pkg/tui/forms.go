package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reshell/pkg/aliases"
	"reshell/pkg/config"
	"reshell/pkg/env"
	"reshell/pkg/packages"
	"reshell/pkg/shell"
	"reshell/pkg/marketplace"
	"reshell/pkg/snippets"
	"reshell/pkg/workflows"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Action Trigger Methods
func (m *model) initForm() {
	m.inputMode = true
	m.inputFocus = 0

	switch m.activeTab {
	case TabSnippets:
		m.formType = "snippet"
		m.formTitle = "Create New Code Snippet"
		m.formInputs = make([]textinput.Model, 5)
		m.formInputs[0] = textinput.New()
		m.formInputs[0].Placeholder = "Snippet Name"
		m.formInputs[0].Focus()
		m.formInputs[1] = textinput.New()
		m.formInputs[1].Placeholder = "Code Block Content"
		m.formInputs[2] = textinput.New()
		m.formInputs[2].Placeholder = "Description"
		m.formInputs[3] = textinput.New()
		m.formInputs[3].Placeholder = "Tags (comma-separated, e.g. docker, clean)"
		m.formInputs[4] = textinput.New()
		m.formInputs[4].Placeholder = "Language (e.g. bash, go, python; default: bash)"

	case TabAliases:
		m.formType = "alias"
		m.formTitle = "Create New Command Alias"
		m.formInputs = make([]textinput.Model, 3)
		m.formInputs[0] = textinput.New()
		m.formInputs[0].Placeholder = "Alias Name"
		m.formInputs[0].Focus()
		m.formInputs[1] = textinput.New()
		m.formInputs[1].Placeholder = "Command Value"
		m.formInputs[2] = textinput.New()
		m.formInputs[2].Placeholder = "Description"

	case TabFunctions:
		m.formType = "function"
		m.formTitle = "Create Custom Function Script"
		m.formInputs = make([]textinput.Model, 1)
		m.formInputs[0] = textinput.New()
		m.formInputs[0].Placeholder = "Function Name"
		m.formInputs[0].Focus()

	case TabScripts:
		m.formType = "script"
		m.formTitle = "Create Custom Script File"
		m.formInputs = make([]textinput.Model, 2)
		m.formInputs[0] = textinput.New()
		m.formInputs[0].Placeholder = "Script Name (e.g. build-dist)"
		m.formInputs[0].Focus()
		m.formInputs[1] = textinput.New()
		m.formInputs[1].Placeholder = "Category (default: general)"

	case TabWorkflows:
		m.formType = "workflow_type"
		m.formTitle = "Choose New Workflow Template"
		m.formInputs = make([]textinput.Model, 1)
		m.formInputs[0] = textinput.New()
		m.formInputs[0].Placeholder = "Type 'empty' or 'workspace' [default: empty]"
		m.formInputs[0].Focus()

	case TabEnv:
		m.formType = "env"
		m.formTitle = "Create Environment Variable"
		m.formInputs = make([]textinput.Model, 3)
		m.formInputs[0] = textinput.New()
		m.formInputs[0].Placeholder = "Variable Name"
		m.formInputs[0].Focus()
		m.formInputs[1] = textinput.New()
		m.formInputs[1].Placeholder = "Variable Value"
		m.formInputs[2] = textinput.New()
		m.formInputs[2].Placeholder = "Description"

	case TabPackages:
		m.formType = "package"
		m.formTitle = "Add System Package Requirement"
		m.formInputs = make([]textinput.Model, 1)
		m.formInputs[0] = textinput.New()
		m.formInputs[0].Placeholder = "Package Name (e.g. jq, ripgrep, tmux)"
		m.formInputs[0].Focus()

	case TabProfiles:
		m.formType = "create_profile"
		m.formTitle = "Create New Configuration Profile"
		m.formInputs = make([]textinput.Model, 1)
		m.formInputs[0] = textinput.New()
		m.formInputs[0].Placeholder = "Profile Name (e.g. work, school, chill)"
		m.formInputs[0].Focus()

	default:
		m.inputMode = false
	}
}

func (m *model) initFormForSudo() {
	m.inputMode = true
	m.inputFocus = 0
	m.formType = "sudo"
	m.formTitle = "Synchronized Package Installer Sudo Authentication"
	m.formInputs = make([]textinput.Model, 1)
	m.formInputs[0] = textinput.New()
	m.formInputs[0].Placeholder = "Sudo Password (Hidden)"
	m.formInputs[0].EchoMode = textinput.EchoPassword
	m.formInputs[0].Focus()
}

func (m *model) initFormForSudoUninstall() {
	m.inputMode = true
	m.inputFocus = 0
	m.formType = "sudo_uninstall"
	m.formTitle = "Package Uninstaller Sudo Authentication"
	m.formInputs = make([]textinput.Model, 1)
	m.formInputs[0] = textinput.New()
	m.formInputs[0].Placeholder = "Sudo Password (Hidden)"
	m.formInputs[0].EchoMode = textinput.EchoPassword
	m.formInputs[0].Focus()
}

func detectEditors() []string {
	var editors []string
	candidates := []string{"nvim", "vim", "code", "micro", "nano", "subl", "emacs", "gedit"}
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			editors = append(editors, c)
		}
	}
	if len(editors) == 0 {
		editors = append(editors, "nano")
	}
	return editors
}

func (m *model) initFormForSetup() {
	m.inputMode = true
	m.inputFocus = 0
	m.formType = "first_time_setup"
	m.formTitle = "Welcome to reshell! First-time Setup"
	m.formInputs = make([]textinput.Model, 2)

	m.formInputs[0] = textinput.New()
	m.formInputs[0].Placeholder = "Your Name"
	m.formInputs[0].Focus()

	m.formInputs[1] = textinput.New()
	detected := detectEditors()
	defaultEd := "nano"
	if len(detected) > 0 {
		defaultEd = detected[0]
	}
	m.formInputs[1].Placeholder = fmt.Sprintf("Default Editor (default: %s)", defaultEd)
}

func (m *model) initFormForWorkflowWorkspace() {
	m.inputMode = true
	m.inputFocus = 0
	m.formType = "workflow_workspace"
	m.formTitle = "Generate Workspace Setup Workflow"
	m.formInputs = make([]textinput.Model, 6)

	m.formInputs[0] = textinput.New()
	m.formInputs[0].Placeholder = "Workflow Name (default: work)"
	m.formInputs[0].Focus()

	m.formInputs[1] = textinput.New()
	m.formInputs[1].Placeholder = "Work Directory (default: ~/projects/work)"

	m.formInputs[2] = textinput.New()
	m.formInputs[2].Placeholder = "Browser Command (e.g. xdg-open, firefox, chrome, brave; default: xdg-open)"

	m.formInputs[3] = textinput.New()
	m.formInputs[3].Placeholder = "Ticket Tracker URL (default: https://linear.app)"

	m.formInputs[4] = textinput.New()
	m.formInputs[4].Placeholder = "Music (Playlist URL / Query, default: piano man)"

	m.formInputs[5] = textinput.New()
	m.formInputs[5].Placeholder = "Work Apps / Commands to start (default: slack)"
}

func (m *model) initFormForWorkflowEmpty() {
	m.inputMode = true
	m.inputFocus = 0
	m.formType = "workflow"
	m.formTitle = "Initialize Generic Workflow"
	m.formInputs = make([]textinput.Model, 2)

	m.formInputs[0] = textinput.New()
	m.formInputs[0].Placeholder = "Workflow Name (e.g. deploy-web)"
	m.formInputs[0].Focus()

	m.formInputs[1] = textinput.New()
	m.formInputs[1].Placeholder = "Workflow Description"
}

func (m *model) initFormForMarketplace() {
	m.inputMode = true
	m.inputFocus = 0
	m.formType = "marketplace"
	m.formTitle = "Install Profile Pack from Marketplace"
	m.formInputs = make([]textinput.Model, 1)
	m.formInputs[0] = textinput.New()
	m.formInputs[0].Placeholder = "Git Repository URL (e.g. github.com/user/pack)"
	m.formInputs[0].Focus()
}


func (m *model) initFormForEditEnv(selected config.EnvVar) {
	m.inputMode = true
	m.inputFocus = 0
	m.formType = "edit_env"
	m.oldEnvName = selected.Name
	m.formTitle = fmt.Sprintf("Edit Environment Variable: %s", selected.Name)

	m.formInputs = make([]textinput.Model, 3)
	m.formInputs[0] = textinput.New()
	m.formInputs[0].Placeholder = "Variable Name"
	m.formInputs[0].SetValue(selected.Name)
	m.formInputs[0].Focus()

	m.formInputs[1] = textinput.New()
	m.formInputs[1].Placeholder = "Variable Value"
	m.formInputs[1].SetValue(selected.Value)

	m.formInputs[2] = textinput.New()
	m.formInputs[2].Placeholder = "Description"
	m.formInputs[2].SetValue(selected.Description)
}

func (m *model) initFormForEditSnippet(selected config.Snippet) {
	m.inputMode = true
	m.inputFocus = 0
	m.formType = "edit_snippet"
	m.oldSnippetName = selected.Name
	m.formTitle = fmt.Sprintf("Edit Snippet: %s", selected.Name)

	m.formInputs = make([]textinput.Model, 4)
	m.formInputs[0] = textinput.New()
	m.formInputs[0].Placeholder = "Snippet Name"
	m.formInputs[0].SetValue(selected.Name)
	m.formInputs[0].Focus()

	m.formInputs[1] = textinput.New()
	m.formInputs[1].Placeholder = "Description"
	m.formInputs[1].SetValue(selected.Description)

	m.formInputs[2] = textinput.New()
	m.formInputs[2].Placeholder = "Tags (comma-separated)"
	m.formInputs[2].SetValue(strings.Join(selected.Tags, ", "))

	m.formInputs[3] = textinput.New()
	m.formInputs[3].Placeholder = "Language"
	m.formInputs[3].SetValue(selected.Language)
}

func (m *model) handleFormKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		if m.formType == "marketplace_confirm" && m.fetchedTempDir != "" {
			os.RemoveAll(m.fetchedTempDir)
			m.fetchedTempDir = ""
			m.fetchedManifest = nil
		}
		m.inputMode = false
		return nil
	case "tab", "down":
		m.formInputs[m.inputFocus].Blur()
		m.inputFocus = (m.inputFocus + 1) % len(m.formInputs)
		m.formInputs[m.inputFocus].Focus()
		return nil
	case "up":
		m.formInputs[m.inputFocus].Blur()
		if m.inputFocus == 0 {
			m.inputFocus = len(m.formInputs) - 1
		} else {
			m.inputFocus--
		}
		m.formInputs[m.inputFocus].Focus()
		return nil
	case "enter":
		m.inputMode = false
		return m.submitForm()
	}
	return nil
}

func (m *model) submitForm() tea.Cmd {
	switch m.formType {
	case "first_time_setup":
		name := m.formInputs[0].Value()
		if name == "" {
			name = "Developer"
		}
		edChoice := m.formInputs[1].Value()
		if edChoice == "" {
			detected := detectEditors()
			if len(detected) > 0 {
				edChoice = detected[0]
			} else {
				edChoice = "nano"
			}
		}

		cfg, err := config.LoadConfig()
		if err == nil {
			cfg.UserName = name
			cfg.Editor = edChoice
			_ = config.SaveConfig(cfg)
		}

		_ = env.AddOrUpdate("EDITOR", edChoice, "reshell default text editor", true)

		m.userName = name
		m.preferredEditor = edChoice
		m.loadData()

	case "snippet":
		name := m.formInputs[0].Value()
		code := m.formInputs[1].Value()
		desc := m.formInputs[2].Value()
		tagsStr := m.formInputs[3].Value()
		lang := strings.TrimSpace(strings.ToLower(m.formInputs[4].Value()))
		if lang == "" {
			lang = "bash"
		} else if !snippets.IsValidLanguage(lang) {
			m.showStatus(fmt.Sprintf("Language '%s' is not recognized. Defaulting to 'bash'.", lang), 4*time.Second)
			lang = "bash"
		}
		var tags []string
		if tagsStr != "" {
			parts := strings.Split(tagsStr, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					tags = append(tags, p)
				}
			}
		}
		if name != "" && code != "" {
			_ = snippets.AddOrUpdate(name, code, desc, tags, lang, false)
		}
	case "alias":
		name := m.formInputs[0].Value()
		val := m.formInputs[1].Value()
		desc := m.formInputs[2].Value()
		if name != "" && val != "" {
			_ = aliases.AddOrUpdate(name, val, desc, "all", true)
		}
	case "function":
		name := m.formInputs[0].Value()
		if name != "" {
			boilerplate := fmt.Sprintf("#!/bin/bash\n# Custom function: %s\n# Add your shell script lines below\n\nfunction %s() {\n    echo \"Hello from %s function!\"\n}\n", name, name, name)
			return m.openEditorForFunction(name, boilerplate)
		}
	case "script":
		name := m.formInputs[0].Value()
		category := m.formInputs[1].Value()
		if name != "" {
			if category == "" {
				category = "general"
			}
			boilerplate := fmt.Sprintf("#!/bin/bash\n# Script: %s/%s\n# Description: Add script logic here\n\necho \"Running %s/%s...\"\n", category, name, category, name)
			return m.openEditorForScript(category, name, boilerplate)
		}
	case "workflow_type":
		choice := strings.ToLower(strings.TrimSpace(m.formInputs[0].Value()))
		if choice == "workspace" || choice == "work" {
			m.initFormForWorkflowWorkspace()
			return nil
		} else {
			m.initFormForWorkflowEmpty()
			return nil
		}
	case "workflow_workspace":
		name := m.formInputs[0].Value()
		if name == "" {
			name = "work"
		}
		dir := m.formInputs[1].Value()
		if dir == "" {
			dir = "~/projects/work"
		}
		browser := m.formInputs[2].Value()
		if browser == "" {
			browser = "xdg-open"
		}
		tracker := m.formInputs[3].Value()
		if tracker == "" {
			tracker = "https://linear.app"
		}
		music := m.formInputs[4].Value()
		if music == "" {
			music = "piano man"
		}
		apps := m.formInputs[5].Value()
		if apps == "" {
			apps = "slack"
		}

		steps := []config.WorkflowStep{}

		// 1. Close/minimize other windows
		steps = append(steps, config.WorkflowStep{
			Command: `if command -v xdotool >/dev/null; then
  ACTIVE_WIN=$(xdotool getactivewindow)
  for win in $(xdotool search --onlyvisible --class ""); do
    if [ "$win" != "$ACTIVE_WIN" ]; then
      xdotool windowminimize "$win" 2>/dev/null || xdotool windowclose "$win" 2>/dev/null
    fi
  done
fi`,
			Dir:     dir,
			Comment: "Minimize/Close all other open windows except current terminal",
		})

		// 2. Open Ticket Tracker
		steps = append(steps, config.WorkflowStep{
			Command: fmt.Sprintf(`%s "%s" &`, browser, tracker),
			Dir:     dir,
			Comment: "Launch Ticket Tracker",
		})

		// 3. Open Music
		musicURL := music
		if !strings.HasPrefix(music, "http://") && !strings.HasPrefix(music, "https://") {
			musicURL = fmt.Sprintf("https://www.youtube.com/results?search_query=%s", strings.ReplaceAll(music, " ", "+"))
		}
		steps = append(steps, config.WorkflowStep{
			Command: fmt.Sprintf(`%s "%s" &`, browser, musicURL),
			Dir:     dir,
			Comment: "Open Music playlist / search",
		})

		// 4. Open Work Apps
		steps = append(steps, config.WorkflowStep{
			Command: fmt.Sprintf(`%s &`, apps),
			Dir:     dir,
			Comment: "Start Work Application(s)",
		})

		_ = workflows.AddOrUpdate(name, "Auto-generated Workspace Setup Workflow", steps)
		m.inputMode = false
		m.loadData()

	case "workflow":
		name := m.formInputs[0].Value()
		desc := m.formInputs[1].Value()
		if name != "" {
			_ = workflows.AddOrUpdate(name, desc, []config.WorkflowStep{
				{Command: "echo 'Step 1'", Dir: "~", Comment: "Initialize"},
			})
			return m.openEditorForWorkflows()
		}
	case "package":
		name := m.formInputs[0].Value()
		if name != "" {
			_ = packages.Add(name)
		}
	case "create_profile":
		name := m.formInputs[0].Value()
		if name != "" {
			err := config.CreateProfile(name)
			if err != nil {
				m.showStatus(fmt.Sprintf("Error creating profile: %v", err), 3*time.Second)
			} else {
				_ = config.SetActiveProfile(name)
				_ = shell.Apply()
				m.showStatus(fmt.Sprintf("Profile '%s' created and activated!", name), 3*time.Second)
			}
		}
	case "env":
		name := m.formInputs[0].Value()
		val := m.formInputs[1].Value()
		desc := m.formInputs[2].Value()
		if name != "" {
			_ = env.AddOrUpdate(name, val, desc, true)
		}
	case "edit_env":
		name := m.formInputs[0].Value()
		val := m.formInputs[1].Value()
		desc := m.formInputs[2].Value()
		if name != "" {
			if m.oldEnvName != "" && name != m.oldEnvName {
				_ = env.Remove(m.oldEnvName)
			}
			_ = env.AddOrUpdate(name, val, desc, true)
		}
	case "edit_snippet":
		name := m.formInputs[0].Value()
		desc := m.formInputs[1].Value()
		tagsStr := m.formInputs[2].Value()
		lang := strings.TrimSpace(strings.ToLower(m.formInputs[3].Value()))
		if lang == "" {
			lang = "bash"
		} else if !snippets.IsValidLanguage(lang) {
			m.showStatus(fmt.Sprintf("Language '%s' is not recognized. Defaulting to 'bash'.", lang), 4*time.Second)
			lang = "bash"
		}
		var tags []string
		if tagsStr != "" {
			parts := strings.Split(tagsStr, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					tags = append(tags, p)
				}
			}
		}

		if name != "" {
			var code string
			var favorite bool
			for _, snip := range m.snippetsData {
				if snip.Name == m.oldSnippetName {
					code = snip.Code
					favorite = snip.Favorite
					break
				}
			}

			if m.oldSnippetName != "" && name != m.oldSnippetName {
				_ = snippets.Remove(m.oldSnippetName)
			}
			_ = snippets.AddOrUpdate(name, code, desc, tags, lang, favorite)
		}
	case "sudo":
		m.sudoPassword = m.formInputs[0].Value()
		m.viewingLogs = true
		m.installLogs = "Starting Synchronized package installer...\n"
		m.viewport.SetContent(m.installLogs)
		return m.runSynchronizedInstaller()

	case "sudo_uninstall":
		m.sudoPassword = m.formInputs[0].Value()
		m.viewingLogs = true
		m.installLogs = "Starting system package uninstaller...\n"
		m.viewport.SetContent(m.installLogs)
		return m.runSystemUninstaller()

	case "marketplace":
		m.marketplaceURL = m.formInputs[0].Value()
		if m.marketplaceURL != "" {
			m.inputMode = true
			m.formType = "marketplace_fetching"
			m.formTitle = "Fetching marketplace pack..."
			m.formInputs = nil
			return m.runMarketplaceFetcher()
		}

	case "marketplace_confirm":
		confirmVal := strings.ToLower(strings.TrimSpace(m.formInputs[0].Value()))
		if confirmVal == "yes" || confirmVal == "y" {
			m.inputMode = false

			// Compute how many items are being merged for the success summary
			varsCount := len(m.fetchedManifest.Variables)
			aliasesCount := len(m.fetchedManifest.Aliases)
			snippetsCount := len(m.fetchedManifest.Snippets)
			pkgsCount := len(m.fetchedManifest.Config.Packages)

			funcsCount := 0
			funcsSourceDir := filepath.Join(m.fetchedTempDir, "functions")
			if files, err := os.ReadDir(funcsSourceDir); err == nil {
				for _, f := range files {
					if !f.IsDir() {
						funcsCount++
					}
				}
			}

			scriptsCount := 0
			scriptsSourceDir := filepath.Join(m.fetchedTempDir, "scripts")
			_ = filepath.Walk(scriptsSourceDir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					scriptsCount++
				}
				return nil
			})

			err := marketplace.MergeManifest(m.fetchedManifest, m.fetchedTempDir)
			os.RemoveAll(m.fetchedTempDir)
			m.fetchedTempDir = ""
			m.fetchedManifest = nil

			if err != nil {
				m.showStatus("Failed to install pack: "+err.Error(), 4*time.Second)
			} else {
				m.loadData()
				summaryMsg := fmt.Sprintf("Imported pack successfully! Summary: +%d env vars, +%d aliases, +%d snippets, +%d packages, +%d functions, +%d scripts.",
					varsCount, aliasesCount, snippetsCount, pkgsCount, funcsCount, scriptsCount)
				m.showStatus(summaryMsg, 5*time.Second)
			}
		} else {
			m.inputMode = false
			if m.fetchedTempDir != "" {
				os.RemoveAll(m.fetchedTempDir)
				m.fetchedTempDir = ""
			}
			m.fetchedManifest = nil
			m.showStatus("Installation cancelled.", 2*time.Second)
		}
	}

	m.loadData()
	return nil
}

func (m model) formView() string {
	s := strings.Builder{}
	s.WriteString(TitleStyle.Render(m.formTitle) + "\n\n")

	for i, input := range m.formInputs {
		prompt := input.Placeholder
		if m.inputFocus == i {
			prompt = SelectedStyle.Render(prompt)
		}
		s.WriteString(fmt.Sprintf("%s:\n%s\n\n", prompt, input.View()))
	}

	s.WriteString(TextMuted.Render("Press [tab]/[down] to switch fields, [enter] to submit, [esc] to cancel."))
	return CardStyle.Width(m.width - 32).Render(s.String())
}

func (m model) logsView() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("Viewport Logger Output (Press 'q' or 'esc' to exit)"),
		m.viewport.View(),
	)
}
