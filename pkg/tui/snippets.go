package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type SnippetsComponent struct{}

func (s SnippetsComponent) View(m model) string {
	if len(m.snippetsData) == 0 {
		return "No snippets stored. Press 'n' to add a new snippet."
	}

	availWidth := m.width - 30
	if availWidth < 45 {
		availWidth = 45
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

		start, end := m.getVisibleSlice(len(m.snippetsData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			snip := m.snippetsData[i]
			name := snip.Name
			if snip.Favorite {
				name += " ★"
			}
			displayName := truncateString(name, leftWidth-5)
			line := fmt.Sprintf("  %s", displayName)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render(line))
			}
		}
		if end < len(m.snippetsData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}

		leftCol := lipgloss.JoinVertical(lipgloss.Left, list...)

		selected := m.snippetsData[m.selectedIdx]
		tagsStr := "None"
		if len(selected.Tags) > 0 {
			tagsStr = strings.Join(selected.Tags, ", ")
		}

		preview := fmt.Sprintf("%s\n%s\n\nLanguage: %s\nTags: %s\n\nCode:\n%s",
			TitleStyle.Render("Snippet: "+selected.Name),
			TextMuted.Render(selected.Description),
			selected.Language,
			tagsStr,
			GetTruncatedCodeBlock(selected.Code, selected.Language, 10),
		)

		previewCard := CardStyle.Width(cardWidth).Render(preview)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, previewCard)
	} else {
		start, end := m.getVisibleSlice(len(m.snippetsData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			snip := m.snippetsData[i]
			name := snip.Name
			if snip.Favorite {
				name += " ★"
			}
			displayName := truncateString(name, availWidth-5)
			line := fmt.Sprintf("  %s", displayName)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render(line))
			}
		}
		if end < len(m.snippetsData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}
		return lipgloss.JoinVertical(lipgloss.Left, list...)
	}
}
