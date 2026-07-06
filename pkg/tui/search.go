package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type SearchResult struct {
	Type        string // "Snippet", "Alias", "Function", "Script", "Workflow"
	Name        string
	Value       string
	Description string
	OriginalIdx int
}

type SearchComponent struct{}

func (s SearchComponent) View(m model) string {
	availWidth := m.width - 30
	if availWidth < 45 {
		availWidth = 45
	}

	searchBar := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PurpleColor).
		Padding(0, 1).
		Width(availWidth - 2).
		Render("🔍 Search: " + m.searchInput.View())

	if len(m.searchResults) == 0 {
		emptyMsg := "No matches found. Start typing to search snippets, aliases, functions, scripts, and workflows."
		return lipgloss.JoinVertical(lipgloss.Left,
			searchBar,
			"\n",
			TextMuted.Render(emptyMsg),
		)
	}

	if availWidth >= 70 {
		leftWidth := availWidth * 3 / 8
		if leftWidth < 30 {
			leftWidth = 30
		}
		if leftWidth > 45 {
			leftWidth = 45
		}
		cardWidth := availWidth - leftWidth - 2

		start, end := m.getVisibleSlice(len(m.searchResults))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			res := m.searchResults[i]
			typeLabel := lipgloss.NewStyle().Foreground(CyanColor).Render(fmt.Sprintf("[%s]", res.Type))
			switch res.Type {
			case "Snippet":
				typeLabel = lipgloss.NewStyle().Foreground(PinkColor).Render("[Snippet]")
			case "Alias":
				typeLabel = lipgloss.NewStyle().Foreground(GreenColor).Render("[Alias]")
			case "Function":
				typeLabel = lipgloss.NewStyle().Foreground(YellowColor).Render("[Func]")
			case "Script":
				typeLabel = lipgloss.NewStyle().Foreground(CyanColor).Render("[Script]")
			case "Workflow":
				typeLabel = lipgloss.NewStyle().Foreground(PurpleColor).Render("[Workfl]")
			}

			displayName := truncateString(res.Name, leftWidth-16)
			line := fmt.Sprintf("  %s %s", typeLabel, displayName)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+typeLabel+" "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render(line))
			}
		}
		if end < len(m.searchResults) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}

		leftCol := lipgloss.JoinVertical(lipgloss.Left, list...)

		selected := m.searchResults[m.selectedIdx]
		actionHint := "Press [Enter] to execute / copy / toggle"
		if selected.Type == "Function" {
			actionHint = "Press [Enter] to edit"
		} else if selected.Type == "Snippet" {
			actionHint = "Press [Enter] to copy"
		}

		lexer := "bash"
		if selected.Type == "Snippet" && selected.OriginalIdx < len(m.snippetsData) {
			lexer = m.snippetsData[selected.OriginalIdx].Language
		}

		preview := fmt.Sprintf("%s %s\n%s\n\n%s\n\nContent / Command / Path:\n%s",
			lipgloss.NewStyle().Foreground(PurpleColor).Bold(true).Render(selected.Type+":"),
			TitleStyle.Render(selected.Name),
			TextMuted.Render(selected.Description),
			SuccessLabel.Render(actionHint),
			GetTruncatedCodeBlock(selected.Value, lexer, 10),
		)

		previewCard := CardStyle.Width(cardWidth).Render(preview)

		mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, previewCard)
		return lipgloss.JoinVertical(lipgloss.Left, searchBar, "\n", mainContent)
	} else {
		start, end := m.getVisibleSlice(len(m.searchResults))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			res := m.searchResults[i]
			displayName := truncateString(res.Name, availWidth-15)
			line := fmt.Sprintf("  [%s] %s", res.Type[:3], displayName)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> ["+res.Type[:3]+"] "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render(line))
			}
		}
		if end < len(m.searchResults) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}

		mainContent := lipgloss.JoinVertical(lipgloss.Left, list...)
		return lipgloss.JoinVertical(lipgloss.Left, searchBar, "\n", mainContent)
	}
}
