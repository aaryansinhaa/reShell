package tui

import (
	"fmt"
	"io"
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
	"reshell/pkg/snippets"
	"reshell/pkg/workflows"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) openEditorForFunction(name, content string) tea.Cmd {
	funcDir, err := config.GetFunctionsDir()
	if err != nil {
		m.showStatus(fmt.Sprintf("Error creating function: %v", err), 3*time.Second)
		return nil
	}

	path := filepath.Join(funcDir, name+".sh")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.WriteFile(path, []byte(content), 0755)
	}

	editor := m.getPreferredEditor()

	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m *model) openEditorForScript(category, name, content string) tea.Cmd {
	scriptsDir, err := config.GetScriptsDir()
	if err != nil {
		m.showStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
		return nil
	}

	catDir := filepath.Join(scriptsDir, category)
	_ = os.MkdirAll(catDir, 0755)

	path := filepath.Join(catDir, name+".sh")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.WriteFile(path, []byte(content), 0755)
	}

	editor := m.getPreferredEditor()

	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m *model) openEditorForWorkflows() tea.Cmd {
	dir, err := config.GetConfigDir()
	if err != nil {
		m.showStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
		return nil
	}
	path := filepath.Join(dir, "workflows.toml")
	
	// Create empty workflows.toml with boilerplate template if it does not exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		boilerplate := `[[workflows]]
name = "deploy-app"
description = "Build and deploy application"

  [[workflows.steps]]
  command = "npm run build"
  dir = "~/projects/myapp"

  [[workflows.steps]]
  command = "scp -r ./dist server:/var/www"
  dir = "~/projects/myapp"
`
		_ = os.WriteFile(path, []byte(boilerplate), 0644)
	}

	editor := m.getPreferredEditor()
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m *model) editSelected() tea.Cmd {
	switch m.activeTab {
	case TabSnippets:
		if len(m.snippetsData) == 0 {
			return nil
		}
		selected := m.snippetsData[m.selectedIdx]
		m.initFormForEditSnippet(selected)
		return nil

	case TabAliases:
		if len(m.aliasesData) == 0 {
			return nil
		}
		selected := m.aliasesData[m.selectedIdx]
		m.initFormForEditAlias(selected)
		return nil

	case TabFunctions:
		if len(m.functionsData) == 0 {
			return nil
		}
		selected := m.functionsData[m.selectedIdx]
		funcDir, _ := config.GetFunctionsDir()
		path := filepath.Join(funcDir, selected+".sh")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			path = filepath.Join(funcDir, selected+".fish")
		}

		editor := m.getPreferredEditor()
		c := exec.Command(editor, path)
		return tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err: err}
		})

	case TabScripts:
		if len(m.scriptsData) == 0 {
			return nil
		}
		selected := m.scriptsData[m.selectedIdx]
		editor := m.getPreferredEditor()
		c := exec.Command(editor, selected.Path)
		return tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err: err}
		})

	case TabWorkflows:
		return m.openEditorForWorkflows()

	case TabEnv:
		if len(m.envData) == 0 {
			return nil
		}
		selected := m.envData[m.selectedIdx]
		m.initFormForEditEnv(selected)
		return nil
	}
	return nil
}

func (m *model) deleteSelected() {
	if m.maxListIndex() < 0 {
		return
	}

	switch m.activeTab {
	case TabSnippets:
		selected := m.snippetsData[m.selectedIdx]
		_ = snippets.Remove(selected.Name)
	case TabAliases:
		selected := m.aliasesData[m.selectedIdx]
		_ = aliases.Remove(selected.Name)
	case TabFunctions:
		selected := m.functionsData[m.selectedIdx]
		_ = functions.Remove(selected)
	case TabScripts:
		selected := m.scriptsData[m.selectedIdx]
		_ = scripts.Remove(selected.Category, selected.Name)
	case TabWorkflows:
		selected := m.workflowsData[m.selectedIdx]
		_ = workflows.Remove(selected.Name)
	case TabEnv:
		selected := m.envData[m.selectedIdx]
		_ = env.Remove(selected.Name)
	case TabPackages:
		selected := m.packagesData[m.selectedIdx]
		_ = packages.Remove(selected)
	case TabProfiles:
		selected := m.profilesData[m.selectedIdx]
		err := config.DeleteProfile(selected)
		if err != nil {
			m.showStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
			return
		}
	}

	if m.selectedIdx > 0 {
		m.selectedIdx--
	}
	m.loadData()
}

func (m *model) copySelected() {
	if m.activeTab == TabSnippets && len(m.snippetsData) > 0 {
		selected := m.snippetsData[m.selectedIdx]
		err := snippets.CopyToClipboard(selected.Name)
		if err != nil {
			m.showStatus(fmt.Sprintf("Failed to copy: %v", err), 2*time.Second)
		} else {
			m.showStatus("Snippet code copied to clipboard!", 2*time.Second)
		}
	}
}

func (m *model) runCommandInViewport(cmd *exec.Cmd, title string) tea.Cmd {
	m.viewingLogs = true
	m.installLogs = fmt.Sprintf("Running: %s...\n\n", title)
	m.viewport.SetContent(m.installLogs)
	m.scriptOutputChan = make(chan string)

	stdoutPipe, errOut := cmd.StdoutPipe()
	stderrPipe, errErr := cmd.StderrPipe()
	if errOut != nil || errErr != nil {
		m.installLogs += fmt.Sprintf("Failed to establish output pipes\n")
		m.viewport.SetContent(m.installLogs)
		return nil
	}

	if err := cmd.Start(); err != nil {
		m.installLogs += fmt.Sprintf("Failed to start process: %v\n", err)
		m.viewport.SetContent(m.installLogs)
		return nil
	}

	go func() {
		defer close(m.scriptOutputChan)

		var wg sync.WaitGroup
		wg.Add(2)

		streamPipe := func(pipe io.ReadCloser) {
			defer wg.Done()
			buf := make([]byte, 1024)
			for {
				n, err := pipe.Read(buf)
				if n > 0 {
					m.scriptOutputChan <- string(buf[:n])
				}
				if err != nil {
					break
				}
			}
		}

		go streamPipe(stdoutPipe)
		go streamPipe(stderrPipe)

		wg.Wait()
		_ = cmd.Wait()
	}()

	return m.listenScriptOutput()
}

func (m *model) executeSelected() tea.Cmd {
	switch m.activeTab {
	case TabScripts:
		if len(m.scriptsData) == 0 {
			return nil
		}
		selected := m.scriptsData[m.selectedIdx]
		c := exec.Command("bash", selected.Path)
		return m.runCommandInViewport(c, "Script '"+selected.Name+"'")

	case TabWorkflows:
		if len(m.workflowsData) == 0 {
			return nil
		}
		selected := m.workflowsData[m.selectedIdx]
		wfCopy := new(config.Workflow)
		wfCopy.Name = selected.Name
		wfCopy.Description = selected.Description
		wfCopy.Steps = make([]config.WorkflowStep, len(selected.Steps))
		copy(wfCopy.Steps, selected.Steps)

		m.runningWorkflow = wfCopy
		m.wfStepsStatus = make([]workflows.StepStatus, len(wfCopy.Steps))
		for i, s := range wfCopy.Steps {
			m.wfStepsStatus[i] = workflows.StepStatus{
				Index:    i,
				Command:  s.Command,
				Finished: false,
			}
		}

		m.wfStepChan = make(chan workflows.StepStatus)
		go workflows.Run(wfCopy, m.wfStepChan)

		return m.listenWorkflowSteps()
	}
	return nil
}

func (m *model) listenWorkflowSteps() tea.Cmd {
	return func() tea.Msg {
		status, ok := <-m.wfStepChan
		if !ok {
			return workflowFinishedMsg{}
		}

		if !status.Finished {
			return stepStartedMsg{index: status.Index}
		} else {
			return stepFinishedMsg{
				index:  status.Index,
				stdout: status.Stdout,
				stderr: status.Stderr,
				err:    status.Error,
			}
		}
	}
}

func (m *model) runSynchronizedInstaller() tea.Cmd {
	_, manager := packages.DetectOS()
	m.pkgInstallChan = make(chan string)

	// Localize state to avoid concurrency data races
	pkgsCopy := make([]string, len(m.packagesData))
	copy(pkgsCopy, m.packagesData)

	var sudoPasswordCopy []byte
	if len(m.sudoPassword) > 0 {
		sudoPasswordCopy = make([]byte, len(m.sudoPassword))
		copy(sudoPasswordCopy, m.sudoPassword)
	}

	go func() {
		defer close(m.pkgInstallChan)
		if len(sudoPasswordCopy) > 0 {
			defer func() {
				for i := range sudoPasswordCopy {
					sudoPasswordCopy[i] = 0
				}
			}()
		}

		for _, pkg := range pkgsCopy {
			if packages.IsInstalled(pkg) {
				m.pkgInstallChan <- fmt.Sprintf("[%s] Already installed.\n", pkg)
				continue
			}

			m.pkgInstallChan <- fmt.Sprintf("[%s] Starting installation...\n", pkg)

			var pkgSudoPassword []byte
			if len(sudoPasswordCopy) > 0 {
				pkgSudoPassword = make([]byte, len(sudoPasswordCopy))
				copy(pkgSudoPassword, sudoPasswordCopy)
			}

			err := packages.Install(pkg, manager, pkgSudoPassword, m.pkgInstallChan)
			if err != nil {
				m.pkgInstallChan <- fmt.Sprintf("[%s] FAILED: %v\n", pkg, err)
			} else {
				m.pkgInstallChan <- fmt.Sprintf("[%s] INSTALLED SUCCESSFULLY.\n", pkg)
			}
		}
	}()

	return m.listenPackageInstall()
}

func (m *model) runSystemUninstaller() tea.Cmd {
	_, manager := packages.DetectOS()
	if len(m.packagesData) == 0 {
		return nil
	}
	pkg := m.packagesData[m.selectedIdx]
	m.pkgInstallChan = make(chan string)

	var sudoPasswordCopy []byte
	if len(m.sudoPassword) > 0 {
		sudoPasswordCopy = make([]byte, len(m.sudoPassword))
		copy(sudoPasswordCopy, m.sudoPassword)
	}

	go func() {
		defer close(m.pkgInstallChan)
		m.pkgInstallChan <- fmt.Sprintf("[%s] Starting uninstallation...\n", pkg)

		err := packages.Uninstall(pkg, manager, sudoPasswordCopy, m.pkgInstallChan)
		if err != nil {
			m.pkgInstallChan <- fmt.Sprintf("[%s] FAILED: %v\n", pkg, err)
		} else {
			m.pkgInstallChan <- fmt.Sprintf("[%s] UNINSTALLED SUCCESSFULLY.\n", pkg)
		}
	}()

	return m.listenPackageInstall()
}

func (m *model) listenPackageInstall() tea.Cmd {
	return func() tea.Msg {
		text, ok := <-m.pkgInstallChan
		if !ok {
			return packageInstallFinishedMsg{}
		}
		return packageInstallOutputMsg{text: text}
	}
}

func (m *model) runMarketplaceFetcher() tea.Cmd {
	url := m.marketplaceURL
	return func() tea.Msg {
		manifest, tempDir, err := marketplace.FetchManifest(url)
		return marketplaceFetchedMsg{manifest: manifest, tempDir: tempDir, err: err}
	}
}

func (m *model) validateFunction() {
	if len(m.functionsData) == 0 {
		return
	}
	selected := m.functionsData[m.selectedIdx]
	output, err := functions.Validate(selected)
	if err != nil {
		m.showStatus(fmt.Sprintf("Syntax error: %s", strings.TrimSpace(output)), 4*time.Second)
	} else {
		m.showStatus("Function syntax is valid!", 2*time.Second)
	}
}

func (m *model) toggleAlias() {
	if len(m.aliasesData) == 0 {
		return
	}
	selected := m.aliasesData[m.selectedIdx]
	_ = aliases.Toggle(selected.Name)
	m.loadData()
}

func (m *model) toggleEnv() {
	if len(m.envData) == 0 {
		return
	}
	selected := m.envData[m.selectedIdx]
	_ = env.Toggle(selected.Name)
	m.loadData()
}

func (m *model) cycleTheme() {
	cfg, err := config.LoadConfig()
	if err != nil {
		return
	}

	themes := []string{"dark", "light", "catppuccin", "gruvbox", "tokyo-night"}
	currentIdx := 0
	for i, t := range themes {
		if t == cfg.Theme {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + 1) % len(themes)
	nextTheme := themes[nextIdx]

	cfg.Theme = nextTheme
	_ = config.SaveConfig(cfg)

	InitTheme(nextTheme)
	m.themeName = nextTheme

	m.viewport.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(GrayColor)
	m.showStatus(fmt.Sprintf("Theme switched to: %s", nextTheme), 2*time.Second)
}

func (m *model) loadData() {
	if a, err := config.LoadAliases(); err == nil {
		m.aliasesData = a.Aliases
	}
	if s, err := config.LoadSnippets(); err == nil {
		m.snippetsData = s.Snippets
	}
	if f, err := functions.List(); err == nil {
		m.functionsData = f
	}
	if sc, err := scripts.List(); err == nil {
		m.scriptsData = sc
	}
	if w, err := config.LoadWorkflows(); err == nil {
		m.workflowsData = w.Workflows
	}
	if cfg, err := config.LoadConfig(); err == nil {
		m.packagesData = cfg.Packages
	}
	if e, err := config.LoadEnv(); err == nil {
		m.envData = e.Variables
	}
	if g, err := git.GetConfig(); err == nil {
		m.gitData = g
	}
	if history, err := git.GetHistory(); err == nil {
		m.gitCommits = history
	}
	if pList, err := config.ListProfiles(); err == nil {
		m.profilesData = pList
	}
	if activeP, err := config.GetActiveProfile(); err == nil {
		m.activeProfile = activeP
	}

	for _, pkg := range m.packagesData {
		m.pkgStatus[pkg] = packages.IsInstalled(pkg)
	}

	m.updateSearchResults()
}

func (m *model) updateSearchResults() {
	query := strings.ToLower(m.searchInput.Value())
	m.searchResults = nil

	// 1. Match Snippets
	for i, s := range m.snippetsData {
		if query == "" || strings.Contains(strings.ToLower(s.Name), query) || strings.Contains(strings.ToLower(s.Description), query) || strings.Contains(strings.ToLower(s.Code), query) {
			m.searchResults = append(m.searchResults, SearchResult{
				Type:        "Snippet",
				Name:        s.Name,
				Value:       s.Code,
				Description: s.Description,
				OriginalIdx: i,
			})
		}
	}

	// 2. Match Aliases
	for i, a := range m.aliasesData {
		if query == "" || strings.Contains(strings.ToLower(a.Name), query) || strings.Contains(strings.ToLower(a.Value), query) || strings.Contains(strings.ToLower(a.Description), query) {
			m.searchResults = append(m.searchResults, SearchResult{
				Type:        "Alias",
				Name:        a.Name,
				Value:       a.Value,
				Description: a.Description,
				OriginalIdx: i,
			})
		}
	}

	// 3. Match Functions
	for i, f := range m.functionsData {
		if query == "" || strings.Contains(strings.ToLower(f), query) {
			m.searchResults = append(m.searchResults, SearchResult{
				Type:        "Function",
				Name:        f,
				Value:       "Custom function script",
				Description: "Shell function",
				OriginalIdx: i,
			})
		}
	}

	// 4. Match Scripts
	for i, s := range m.scriptsData {
		if query == "" || strings.Contains(strings.ToLower(s.Name), query) || strings.Contains(strings.ToLower(s.Description), query) || strings.Contains(strings.ToLower(s.Category), query) {
			m.searchResults = append(m.searchResults, SearchResult{
				Type:        "Script",
				Name:        s.Name,
				Value:       s.Path,
				Description: fmt.Sprintf("[%s] %s", s.Category, s.Description),
				OriginalIdx: i,
			})
		}
	}

	// 5. Match Workflows
	for i, w := range m.workflowsData {
		if query == "" || strings.Contains(strings.ToLower(w.Name), query) || strings.Contains(strings.ToLower(w.Description), query) {
			var steps []string
			for _, step := range w.Steps {
				steps = append(steps, step.Command)
			}
			m.searchResults = append(m.searchResults, SearchResult{
				Type:        "Workflow",
				Name:        w.Name,
				Value:       strings.Join(steps, " -> "),
				Description: w.Description,
				OriginalIdx: i,
			})
		}
	}

	if m.selectedIdx >= len(m.searchResults) {
		m.selectedIdx = 0
	}
}

func (m *model) executeSearchResult() tea.Cmd {
	if len(m.searchResults) == 0 {
		return nil
	}

	selected := m.searchResults[m.selectedIdx]
	switch selected.Type {
	case "Snippet":
		err := snippets.CopyToClipboard(selected.Name)
		if err != nil {
			m.showStatus(fmt.Sprintf("Failed to copy snippet: %v", err), 2*time.Second)
		} else {
			m.showStatus("Snippet code copied to clipboard!", 2*time.Second)
		}
		return nil

	case "Alias":
		_ = aliases.Toggle(selected.Name)
		m.loadData()
		m.showStatus(fmt.Sprintf("Toggled alias: %s", selected.Name), 2*time.Second)
		return nil

	case "Function":
		boilerplate := fmt.Sprintf("#!/bin/bash\n# Custom function: %s\n", selected.Name)
		return m.openEditorForFunction(selected.Name, boilerplate)

	case "Script":
		scriptObj := m.scriptsData[selected.OriginalIdx]
		c := exec.Command("bash", scriptObj.Path)
		return m.runCommandInViewport(c, "Script '"+scriptObj.Name+"'")

	case "Workflow":
		wfObj := m.workflowsData[selected.OriginalIdx]
		wfCopy := new(config.Workflow)
		wfCopy.Name = wfObj.Name
		wfCopy.Description = wfObj.Description
		wfCopy.Steps = make([]config.WorkflowStep, len(wfObj.Steps))
		copy(wfCopy.Steps, wfObj.Steps)

		m.runningWorkflow = wfCopy
		m.wfStepsStatus = make([]workflows.StepStatus, len(wfCopy.Steps))
		for i, s := range wfCopy.Steps {
			m.wfStepsStatus[i] = workflows.StepStatus{
				Index:    i,
				Command:  s.Command,
				Finished: false,
			}
		}

		m.wfStepChan = make(chan workflows.StepStatus)
		go workflows.Run(wfCopy, m.wfStepChan)

		m.activeTab = TabWorkflows
		return m.listenWorkflowSteps()
	}

	return nil
}

func (m *model) editSnippetCode() tea.Cmd {
	if len(m.snippetsData) == 0 {
		return nil
	}
	selected := m.snippetsData[m.selectedIdx]
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("reshell-snippet-%s.txt", selected.Name))
	_ = os.WriteFile(tempFile, []byte(selected.Code), 0644)

	editor := m.getPreferredEditor()
	c := exec.Command(editor, tempFile)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err == nil {
			if updatedCodeBytes, errRead := os.ReadFile(tempFile); errRead == nil {
				_ = snippets.AddOrUpdate(selected.Name, string(updatedCodeBytes), selected.Description, selected.Tags, selected.Language, selected.Favorite)
			}
			os.Remove(tempFile)
		}
		return editorFinishedMsg{err: err}
	})
}

func (m *model) toggleSnippetFavorite() {
	if len(m.snippetsData) == 0 {
		return
	}
	selected := m.snippetsData[m.selectedIdx]
	_ = snippets.ToggleFavorite(selected.Name)
	m.loadData()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func (m *model) getPreferredEditor() string {
	for _, envVar := range m.envData {
		if envVar.Name == "EDITOR" && envVar.Enabled && envVar.Value != "" {
			return envVar.Value
		}
	}

	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}

	if m.preferredEditor != "" {
		return m.preferredEditor
	}

	return "nano"
}
