package tui

import (
	"fmt"
	"reshell/pkg/shell"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var tabNames = map[ActiveTab]string{
	TabSearch:      "Finder 🔍",
	TabSnippets:    "Snippets",
	TabAliases:     "Aliases",
	TabFunctions:   "Functions",
	TabScripts:     "Scripts",
	TabWorkflows:   "Workflows",
	TabPackages:    "Packages",
	TabMarketplace: "Marketplace",
	TabEnv:         "Environment",
	TabGit:         "Git Config",
	TabProfiles:    "Profiles",
}

type ChromeComponent struct{}

func (c ChromeComponent) HeaderView(m model) string {
	logo := " ⚒️  reshell "
	shellName := shell.DetectShell()
	profile, _ := shell.GetShellProfile(shellName)

	status := fmt.Sprintf("Theme: %s | Shell: %s | Profile: %s",
		SuccessLabel.Render(m.themeName),
		SuccessLabel.Render(shellName),
		TextMuted.Render(profile),
	)
	headerText := SelectedStyle.Render(logo)
	if m.userName != "" {
		greeting := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff79c6")).
			Italic(true).
			Bold(true).
			Render(fmt.Sprintf(" ✨ Hello, %s ✨", m.userName))
		headerText += greeting
	}

	contentWidth := m.width - 8
	wLeft := lipgloss.Width(headerText)
	wRight := lipgloss.Width(status)

	spaces := contentWidth - wLeft - wRight
	if spaces < 0 {
		spaces = 0
	}

	content := headerText + strings.Repeat(" ", spaces) + status
	return HeaderStyle.Width(m.width - 4).Render(content)
}

func (c ChromeComponent) SidebarView(m model, h int) string {
	var tabs []string
	for i := ActiveTab(0); i < 11; i++ {
		name := tabNames[i]
		if m.activeTab == i {
			tabs = append(tabs, TabActiveStyle.Width(20).Render(" "+name))
		} else {
			tabs = append(tabs, TabInactiveStyle.Width(20).Render(" "+name))
		}
	}
	return SidebarStyle.Height(h).Render(lipgloss.JoinVertical(lipgloss.Left, tabs...))
}

func (c ChromeComponent) HelpView(m model) string {
	keys := []string{"Tab/S-Tab: Cycle tabs", "Ctrl+/: Finder", "Ctrl+t: Theme", "Ctrl+a: Apply", "q/Ctrl+c: Quit"}
	switch m.activeTab {
	case TabSearch:
		keys = append(keys, "Type: Filter results", "Up/Down: Nav matches", "Enter: Exec/Copy/Toggle", "Esc: Clear")
	case TabSnippets:
		keys = append(keys, "n: Add snippet", "e: Edit details", "E: Edit code", "d: Delete snippet", "c: Copy snippet", "f: Favorite snippet")
	case TabAliases:
		keys = append(keys, "n: Add alias", "e: Edit alias", "d: Delete alias", "Space: Toggle enable/disable")
	case TabFunctions:
		keys = append(keys, "n: Create function", "e: Edit body", "d: Remove", "v: Dry-run check syntax")
	case TabScripts:
		keys = append(keys, "n: Create script", "e: Edit body", "d: Remove", "x: Execute script")
	case TabWorkflows:
		keys = append(keys, "n: Initialize workflow", "e: Edit workflows.toml", "x: Run workflow", "d: Delete")
	case TabPackages:
		keys = append(keys, "n: Add package", "d: Delete", "i: Install packages", "u: Uninstall package")
	case TabMarketplace:
		keys = append(keys, "i: Install profile package")
	case TabEnv:
		keys = append(keys, "n: Add variable", "e: Edit variable", "d: Delete", "Space: Toggle enable/disable")
	case TabGit:
		if m.gitHistoryView {
			keys = append(keys, "h: Config view", "Up/Down: Nav commits", "r/Enter: Revert to version", "c: Clear history")
		} else {
			keys = append(keys, "h: History view")
		}
	case TabProfiles:
		keys = append(keys, "s/Enter: Activate profile", "n: Create profile", "d: Delete profile")
	}

	return HelpStyle.Width(m.width - 4).Render(strings.Join(keys, "  |  "))
}
