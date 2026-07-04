package tui

import (
	"fmt"
	"reshell/pkg/scripts"

	"github.com/charmbracelet/lipgloss"
)

type ScriptsComponent struct{}

func (s ScriptsComponent) View(m model) string {
	if len(m.scriptsData) == 0 {
		return "No scripts loaded. Create files in ~/.config/reshell/scripts/<category>/*.sh."
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

		start, end := m.getVisibleSlice(len(m.scriptsData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			scr := m.scriptsData[i]
			line := fmt.Sprintf("[%s] %s", scr.Category, scr.Name)
			displayName := truncateString(line, leftWidth-5)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render("  "+displayName))
			}
		}
		if end < len(m.scriptsData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}

		leftCol := lipgloss.JoinVertical(lipgloss.Left, list...)

		selected := m.scriptsData[m.selectedIdx]
		code, _ := scripts.Get(selected.Category, selected.Name)

		preview := fmt.Sprintf("%s\n%s\n\nParameters: %v\nPath: %s\n\n--- Script Preview ---\n%s",
			TitleStyle.Render(selected.Name),
			TextMuted.Render(selected.Description),
			selected.Parameters,
			selected.Path,
			HighlightCode(code, "bash"),
		)

		previewCard := CardStyle.Width(cardWidth).Render(preview)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, previewCard)
	} else {
		start, end := m.getVisibleSlice(len(m.scriptsData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			scr := m.scriptsData[i]
			line := fmt.Sprintf("[%s] %s", scr.Category, scr.Name)
			displayName := truncateString(line, availWidth-5)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render("  "+displayName))
			}
		}
		if end < len(m.scriptsData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}
		return lipgloss.JoinVertical(lipgloss.Left, list...)
	}
}
