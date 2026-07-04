package tui

import (
	"fmt"
	"reshell/pkg/functions"

	"github.com/charmbracelet/lipgloss"
)

type FunctionsComponent struct{}

func (f FunctionsComponent) View(m model) string {
	if len(m.functionsData) == 0 {
		return "No functions registered. Press 'n' to add a custom function script."
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

		start, end := m.getVisibleSlice(len(m.functionsData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			fName := m.functionsData[i]
			displayName := truncateString(fName, leftWidth-5)
			line := fmt.Sprintf("  %s", displayName)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render(line))
			}
		}
		if end < len(m.functionsData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}

		leftCol := lipgloss.JoinVertical(lipgloss.Left, list...)

		selected := m.functionsData[m.selectedIdx]
		code, ext, err := functions.Get(selected)
		var preview string
		if err != nil {
			preview = fmt.Sprintf("Error reading function body: %v", err)
		} else {
			preview = fmt.Sprintf("%s (extension %s)\n\n%s",
				TitleStyle.Render("Function: "+selected),
				ext,
				HighlightCode(code, "bash"),
			)
		}

		previewCard := CardStyle.Width(cardWidth).Render(preview)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, previewCard)
	} else {
		start, end := m.getVisibleSlice(len(m.functionsData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			fName := m.functionsData[i]
			displayName := truncateString(fName, availWidth-5)
			line := fmt.Sprintf("  %s", displayName)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render(line))
			}
		}
		if end < len(m.functionsData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}
		return lipgloss.JoinVertical(lipgloss.Left, list...)
	}
}
