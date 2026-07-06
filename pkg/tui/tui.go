package tui

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"reshell/pkg/config"
	"reshell/pkg/git"
	"reshell/pkg/marketplace"
	"reshell/pkg/packages"
	"reshell/pkg/scripts"
	"reshell/pkg/workflows"
)

type ActiveTab int

const (
	TabSearch ActiveTab = iota
	TabSnippets
	TabAliases
	TabFunctions
	TabScripts
	TabWorkflows
	TabPackages
	TabMarketplace
	TabEnv
	TabGit
	TabProfiles
)

// Msg structures for async channels
type workflowFinishedMsg struct{}
type stepStartedMsg struct {
	index int
}
type stepFinishedMsg struct {
	index  int
	stdout string
	stderr string
	err    error
}
type packageInstallFinishedMsg struct{}
type packageInstallOutputMsg struct {
	text string
}
type marketplaceFinishedMsg struct {
	manifest *marketplace.MarketplaceManifest
	err      error
}
type marketplaceFetchedMsg struct {
	manifest *marketplace.MarketplaceManifest
	tempDir  string
	err      error
}
type applyFinishedMsg struct {
	err error
}
type editorFinishedMsg struct {
	err error
}

type model struct {
	activeTab           ActiveTab
	width, height       int
	selectedIdx         int
	statusMessage       string
	statusMessageExpiry time.Time
	themeName           string
	userName            string
	preferredEditor     string
	mainHeight          int

	// Search Component State
	searchInput   textinput.Model
	searchResults []SearchResult

	// Local databases
	aliasesData   []config.Alias
	snippetsData  []config.Snippet
	functionsData []string
	scriptsData   []scripts.Script
	workflowsData []config.Workflow
	envData       []config.EnvVar
	gitData        *git.GitConfig
	gitCommits     []git.Commit
	gitHistoryView bool
	gitSelectedIdx int
	profilesData   []string
	activeProfile  string
	packagesData   []string
	pkgStatus      map[string]bool // pkg -> installed

	// Interactive Input Forms
	inputMode  bool
	formType       string // "snippet", "alias", "function", "script", "workflow", "env", "sudo", "marketplace"
	oldEnvName     string
	oldSnippetName string
	formTitle      string
	formInputs []textinput.Model
	inputFocus int

	// Workflow Runner State
	runningWorkflow *config.Workflow
	wfStepChan      chan workflows.StepStatus
	wfStepsStatus   []workflows.StepStatus

	// Package Installer Sudo Authentication State
	sudoPassword     string
	pkgInstallChan   chan string
	scriptOutputChan chan string
	installLogs      string
	viewingLogs      bool
	viewport         viewport.Model

	// Marketplace installer state
	marketplaceURL string
	fetchedManifest *marketplace.MarketplaceManifest
	fetchedTempDir  string

	// Modular sub-components
	search      SearchComponent
	snippets    SnippetsComponent
	aliases     AliasesComponent
	functions   FunctionsComponent
	scripts     ScriptsComponent
	workflows   WorkflowsComponent
	packages    PackagesComponent
	marketplace MarketplaceComponent
	env         EnvComponent
	git         GitComponent
	profiles    ProfilesComponent
	chrome      ChromeComponent
}

func initialModel() model {
	cfg, err := config.LoadConfig()
	var pkgs []string
	theme := "dark"
	var userName string
	var preferredEditor string
	if err == nil {
		pkgs = cfg.Packages
		theme = cfg.Theme
		userName = cfg.UserName
		preferredEditor = cfg.Editor
		InitTheme(cfg.Theme)
	} else {
		InitTheme(theme)
	}

	si := textinput.New()
	si.Placeholder = "Type to search aliases, snippets, scripts..."
	si.Focus()

	m := model{
		activeTab:       TabSearch,
		pkgStatus:       make(map[string]bool),
		packagesData:    pkgs,
		viewport:        viewport.New(80, 20),
		search:          SearchComponent{},
		searchInput:     si,
		themeName:       theme,
		userName:        userName,
		preferredEditor: preferredEditor,
		profiles:        ProfilesComponent{},
	}

	m.viewport.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(GrayColor)

	m.loadData()

	if m.userName == "" {
		m.initFormForSetup()
	}

	return m
}

func (m *model) showStatus(msg string, duration time.Duration) {
	m.statusMessage = msg
	m.statusMessageExpiry = time.Now().Add(duration)
}

func (m model) maxListIndex() int {
	switch m.activeTab {
	case TabSnippets:
		return len(m.snippetsData) - 1
	case TabAliases:
		return len(m.aliasesData) - 1
	case TabFunctions:
		return len(m.functionsData) - 1
	case TabScripts:
		return len(m.scriptsData) - 1
	case TabWorkflows:
		return len(m.workflowsData) - 1
	case TabEnv:
		return len(m.envData) - 1
	case TabPackages:
		return len(m.packagesData) - 1
	case TabProfiles:
		return len(m.profilesData) - 1
	default:
		return -1
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 30
		m.viewport.Height = msg.Height - 12
		return m, nil

	case stepStartedMsg:
		if m.runningWorkflow != nil && msg.index < len(m.wfStepsStatus) {
			m.wfStepsStatus[msg.index].Finished = false
		}
		cmds = append(cmds, m.listenWorkflowSteps())

	case stepFinishedMsg:
		if m.runningWorkflow != nil && msg.index < len(m.wfStepsStatus) {
			m.wfStepsStatus[msg.index].Finished = true
			m.wfStepsStatus[msg.index].Stdout = msg.stdout
			m.wfStepsStatus[msg.index].Stderr = msg.stderr
			m.wfStepsStatus[msg.index].Error = msg.err
		}
		cmds = append(cmds, m.listenWorkflowSteps())

	case workflowFinishedMsg:
		m.runningWorkflow = nil
		m.wfStepChan = nil
		m.showStatus("Workflow execution sequence completed.", 3*time.Second)

	case scriptOutputMsg:
		m.installLogs += msg.text
		m.viewport.SetContent(m.installLogs)
		m.viewport.GotoBottom()
		cmds = append(cmds, m.listenScriptOutput())

	case scriptFinishedMsg:
		m.scriptOutputChan = nil
		m.showStatus("Script execution finished.", 3*time.Second)

	case packageInstallOutputMsg:
		m.installLogs += msg.text
		m.viewport.SetContent(m.installLogs)
		m.viewport.GotoBottom()
		cmds = append(cmds, m.listenPackageInstall())

	case packageInstallFinishedMsg:
		m.pkgInstallChan = nil
		m.sudoPassword = ""
		m.loadData()
		m.showStatus("Synchronized installation checks complete.", 3*time.Second)

	case marketplaceFinishedMsg:
		if msg.err != nil {
			m.showStatus("Failed to import profile: "+msg.err.Error(), 4*time.Second)
		} else {
			m.loadData()
			m.showStatus("Marketplace configuration pack imported successfully!", 3*time.Second)
		}

	case marketplaceFetchedMsg:
		if msg.err != nil {
			m.inputMode = false
			m.showStatus("Failed to fetch pack: "+msg.err.Error(), 4*time.Second)
		} else {
			m.fetchedManifest = msg.manifest
			m.fetchedTempDir = msg.tempDir
			m.inputMode = true
			m.formType = "marketplace_confirm"
			m.formTitle = fmt.Sprintf("Confirm Install: %s", msg.manifest.Package.Name)
			m.formInputs = make([]textinput.Model, 1)
			m.formInputs[0] = textinput.New()
			m.formInputs[0].Placeholder = "Type 'yes' to confirm and install, or 'no'/'esc' to cancel"
			m.formInputs[0].Focus()
		}

	case applyFinishedMsg:
		if msg.err != nil {
			m.showStatus(fmt.Sprintf("Failed to apply settings: %v", msg.err), 3*time.Second)
		} else {
			m.showStatus("Configurations compiled and applied successfully!", 3*time.Second)
		}

	case editorFinishedMsg:
		m.loadData()
		m.showStatus("Index refreshed.", 2*time.Second)

	case tea.KeyMsg:
		if m.activeTab == TabSearch && !m.inputMode && !m.viewingLogs {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "tab":
				m.activeTab = (m.activeTab + 1) % 11
				m.selectedIdx = 0
				return m, nil
			case "shift+tab":
				m.activeTab = (m.activeTab - 1 + 11) % 11
				m.selectedIdx = 0
				return m, nil
			case "ctrl+/", "ctrl+_":
				m.searchInput.Focus()
				return m, nil
			case "up", "ctrl+p":
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
				return m, nil
			case "down", "ctrl+n":
				if m.selectedIdx < len(m.searchResults)-1 {
					m.selectedIdx++
				}
				return m, nil
			case "enter":
				cmd := m.executeSearchResult()
				return m, cmd
			case "esc":
				m.searchInput.SetValue("")
				m.updateSearchResults()
				return m, nil
			case "ctrl+t":
				m.cycleTheme()
				return m, nil
			case "ctrl+a":
				return m, m.applySettings()
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.updateSearchResults()
				return m, cmd
			}
		}

		if m.inputMode {
			cmd := m.handleFormKey(msg)
			if cmd != nil {
				return m, cmd
			}
			var textCmd tea.Cmd
			m.formInputs[m.inputFocus], textCmd = m.formInputs[m.inputFocus].Update(msg)
			return m, textCmd
		}

		if m.viewingLogs {
			switch msg.String() {
			case "q", "esc":
				m.viewingLogs = false
				return m, nil
			}
			var vpCmd tea.Cmd
			m.viewport, vpCmd = m.viewport.Update(msg)
			return m, vpCmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.activeTab = (m.activeTab + 1) % 11
			m.selectedIdx = 0
			m.gitHistoryView = false

		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + 11) % 11
			m.selectedIdx = 0
			m.gitHistoryView = false

		case "ctrl+/", "ctrl+_":
			m.activeTab = TabSearch
			m.selectedIdx = 0
			m.searchInput.Focus()

		case "ctrl+t":
			m.cycleTheme()

		case "ctrl+a":
			return m, m.applySettings()

		case "up", "k":
			if m.activeTab == TabGit && m.gitHistoryView {
				if m.gitSelectedIdx > 0 {
					m.gitSelectedIdx--
				} else {
					m.gitSelectedIdx = len(m.gitCommits) - 1
					if m.gitSelectedIdx < 0 {
						m.gitSelectedIdx = 0
					}
				}
			} else {
				if m.selectedIdx > 0 {
					m.selectedIdx--
				} else {
					m.selectedIdx = m.maxListIndex()
					if m.selectedIdx < 0 {
						m.selectedIdx = 0
					}
				}
			}

		case "down", "j":
			if m.activeTab == TabGit && m.gitHistoryView {
				maxIdx := len(m.gitCommits) - 1
				if maxIdx >= 0 {
					if m.gitSelectedIdx < maxIdx {
						m.gitSelectedIdx++
					} else {
						m.gitSelectedIdx = 0
					}
				}
			} else {
				maxIdx := m.maxListIndex()
				if maxIdx >= 0 {
					if m.selectedIdx < maxIdx {
						m.selectedIdx++
					} else {
						m.selectedIdx = 0
					}
				}
			}

		case "n":
			m.initForm()

		case "e":
			cmd := m.editSelected()
			if cmd != nil {
				return m, cmd
			}

		case "E":
			if m.activeTab == TabSnippets {
				cmd := m.editSnippetCode()
				if cmd != nil {
					return m, cmd
				}
			}

		case "d":
			m.deleteSelected()

		case "c":
			if m.activeTab == TabGit && m.gitHistoryView {
				err := git.ClearHistory()
				if err != nil {
					m.showStatus(fmt.Sprintf("Failed to clear history: %v", err), 3*time.Second)
				} else {
					m.loadData()
					m.gitSelectedIdx = 0
					m.showStatus("Version history cleared successfully!", 3*time.Second)
				}
			} else {
				m.copySelected()
			}

		case "x":
			cmd := m.executeSelected()
			if cmd != nil {
				return m, cmd
			}

		case "v":
			if m.activeTab == TabFunctions {
				m.validateFunction()
			}

		case " ":
			if m.activeTab == TabAliases {
				m.toggleAlias()
			} else if m.activeTab == TabEnv {
				m.toggleEnv()
			}

		case "f":
			if m.activeTab == TabSnippets {
				m.toggleSnippetFavorite()
			}

		case "i":
			if m.activeTab == TabPackages {
				// Check if sudo credentials are cached
				if exec.Command("sudo", "-n", "true").Run() == nil {
					m.sudoPassword = ""
					m.viewingLogs = true
					m.installLogs = "Starting Synchronized package installer...\n"
					m.viewport.SetContent(m.installLogs)
					return m, m.runSynchronizedInstaller()
				} else {
					m.initFormForSudo()
				}
			} else if m.activeTab == TabMarketplace {
				m.initFormForMarketplace()
			}

		case "u":
			if m.activeTab == TabPackages {
				_, manager := packages.DetectOS()
				needsSudo := manager == "apt" || manager == "dnf" || manager == "pacman"
				if needsSudo {
					// Check if sudo credentials are cached
					if exec.Command("sudo", "-n", "true").Run() == nil {
						m.sudoPassword = ""
						m.viewingLogs = true
						m.installLogs = "Starting system package uninstaller...\n"
						m.viewport.SetContent(m.installLogs)
						return m, m.runSystemUninstaller()
					} else {
						m.initFormForSudoUninstall()
					}
				} else {
					m.viewingLogs = true
					m.installLogs = "Starting system package uninstaller...\n"
					m.viewport.SetContent(m.installLogs)
					return m, m.runSystemUninstaller()
				}
			}

		case "h":
			if m.activeTab == TabGit {
				m.gitHistoryView = !m.gitHistoryView
				m.gitSelectedIdx = 0
				return m, nil
			}

		case "s":
			if m.activeTab == TabProfiles && len(m.profilesData) > 0 {
				selected := m.profilesData[m.selectedIdx]
				err := config.SetActiveProfile(selected)
				if err != nil {
					m.showStatus(fmt.Sprintf("Failed to switch profile: %v", err), 3*time.Second)
				} else {
					m.loadData()
					m.selectedIdx = 0
					m.showStatus(fmt.Sprintf("Switched active profile to '%s'!", selected), 3*time.Second)
					return m, m.applySettings()
				}
			}

		case "r":
			if m.activeTab == TabGit && m.gitHistoryView && len(m.gitCommits) > 0 {
				selectedCommit := m.gitCommits[m.gitSelectedIdx]
				err := git.RevertToCommit(selectedCommit.Hash)
				if err != nil {
					m.showStatus(fmt.Sprintf("Revert failed: %v", err), 3*time.Second)
				} else {
					m.loadData()
					m.showStatus(fmt.Sprintf("Reverted workspace to commit %s!", selectedCommit.Hash), 3*time.Second)
					return m, m.applySettings()
				}
			}

		case "enter":
			if m.activeTab == TabGit && m.gitHistoryView && len(m.gitCommits) > 0 {
				selectedCommit := m.gitCommits[m.gitSelectedIdx]
				err := git.RevertToCommit(selectedCommit.Hash)
				if err != nil {
					m.showStatus(fmt.Sprintf("Revert failed: %v", err), 3*time.Second)
				} else {
					m.loadData()
					m.showStatus(fmt.Sprintf("Reverted workspace to commit %s!", selectedCommit.Hash), 3*time.Second)
					return m, m.applySettings()
				}
			} else if m.activeTab == TabProfiles && len(m.profilesData) > 0 {
				selected := m.profilesData[m.selectedIdx]
				err := config.SetActiveProfile(selected)
				if err != nil {
					m.showStatus(fmt.Sprintf("Failed to switch profile: %v", err), 3*time.Second)
				} else {
					m.loadData()
					m.selectedIdx = 0
					m.showStatus(fmt.Sprintf("Switched active profile to '%s'!", selected), 3*time.Second)
					return m, m.applySettings()
				}
			} else if m.activeTab == TabSnippets {
				m.copySelected()
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// TUI layout generation
func (m model) View() string {
	h := m.height - 10
	hasStatus := false
	if time.Now().Before(m.statusMessageExpiry) {
		h = m.height - 11
		hasStatus = true
	}
	m.mainHeight = h

	header := m.chrome.HeaderView(m)
	left := m.chrome.SidebarView(m, h)
	var right string

	switch m.activeTab {
	case TabSearch:
		right = m.search.View(m)
	case TabSnippets:
		right = m.snippets.View(m)
	case TabAliases:
		right = m.aliases.View(m)
	case TabFunctions:
		right = m.functions.View(m)
	case TabScripts:
		right = m.scripts.View(m)
	case TabWorkflows:
		right = m.workflows.View(m)
	case TabPackages:
		right = m.packages.View(m)
	case TabMarketplace:
		right = m.marketplace.View(m)
	case TabEnv:
		right = m.env.View(m)
	case TabGit:
		right = m.git.View(m)
	case TabProfiles:
		right = m.profiles.View(m)
	}

	if m.inputMode {
		right = m.formView()
	} else if m.viewingLogs {
		right = m.logsView()
	}

	mainPanel := lipgloss.JoinHorizontal(lipgloss.Top, left, MainStyle.Height(h).Render(right))
	help := m.chrome.HelpView(m)

	var joined string
	if hasStatus {
		statusLine := lipgloss.NewStyle().Foreground(YellowColor).Bold(true).Padding(0, 2).Render(m.statusMessage)
		joined = lipgloss.JoinVertical(lipgloss.Left, header, mainPanel, statusLine, help)
	} else {
		joined = lipgloss.JoinVertical(lipgloss.Left, header, mainPanel, help)
	}

	return makeOpaque(joined, m.width, m.height, BgColor)
}

func (m model) getVisibleSlice(totalItems int) (int, int) {
	maxVisible := m.mainHeight - 4
	if maxVisible < 5 {
		maxVisible = 5
	}
	if totalItems <= maxVisible {
		return 0, totalItems
	}
	start := m.selectedIdx - maxVisible/2
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > totalItems {
		end = totalItems
		start = end - maxVisible
	}
	return start, end
}

func makeOpaque(s string, width, height int, bg lipgloss.Color) string {
	return s
}

func (m model) applySettings() tea.Cmd {
	return func() tea.Msg {
		c := exec.Command("reshell", "apply")
		err := c.Run()
		return applyFinishedMsg{err: err}
	}
}

type scriptOutputMsg struct {
	text string
}

type scriptFinishedMsg struct{}

func (m *model) listenScriptOutput() tea.Cmd {
	return func() tea.Msg {
		text, ok := <-m.scriptOutputChan
		if !ok {
			return scriptFinishedMsg{}
		}
		return scriptOutputMsg{text: text}
	}
}

func Start() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
