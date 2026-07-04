package tui

import (
	"fmt"
	"reshell/pkg/aliases"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type AliasesComponent struct{}

func (a AliasesComponent) View(m model) string {
	if len(m.aliasesData) == 0 {
		return "No aliases defined. Press 'n' to add one."
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

		start, end := m.getVisibleSlice(len(m.aliasesData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			al := m.aliasesData[i]
			name := al.Name
			value := al.Value

			maxNameLen := 12
			maxValueLen := leftWidth - maxNameLen - 16
			if maxValueLen < 8 {
				maxValueLen = 8
			}

			displayName := truncateString(name, maxNameLen)
			displayValue := truncateString(value, maxValueLen)

			status := "ok"
			isOverride := false
			if warn, ok := aliases.DetectConflict(al.Name); ok && al.Enabled {
				if !strings.HasPrefix(warn, "collides with another active alias") {
					status = "⚠️ override"
					isOverride = true
				}
			}

			if !al.Enabled {
				displayName = "[Disabled] " + displayName
			}

			line := fmt.Sprintf("%-12s = %-*s (%s)", displayName, maxValueLen, displayValue, status)
			line = truncateString(line, leftWidth-2)

			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+line))
			} else {
				styled := "  " + line
				if isOverride && al.Enabled {
					list = append(list, lipgloss.NewStyle().Background(BgColor).Foreground(YellowColor).Render(styled))
				} else if !al.Enabled {
					list = append(list, TextMuted.Render(styled))
				} else {
					list = append(list, UnselectedStyle.Render(styled))
				}
			}
		}
		if end < len(m.aliasesData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}

		leftCol := lipgloss.JoinVertical(lipgloss.Left, list...)

		selected := m.aliasesData[m.selectedIdx]
		warnMsg := ""
		if warn, ok := aliases.DetectConflict(selected.Name); ok && selected.Enabled {
			if !strings.HasPrefix(warn, "collides with another active alias") {
				warnMsg = "\n\n" + WarningLabel.Render("Conflict Warning: ") + warn
			}
		}

		preview := fmt.Sprintf("%s\n%s\n\nCommand: %s\nShell: %s\nEnabled: %t%s",
			TitleStyle.Render("Alias: "+selected.Name),
			TextMuted.Render(selected.Description),
			lipgloss.NewStyle().Foreground(GreenColor).Render(selected.Value),
			selected.Shell,
			selected.Enabled,
			warnMsg,
		)

		previewCard := CardStyle.Width(cardWidth).Render(preview)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, previewCard)
	} else {
		start, end := m.getVisibleSlice(len(m.aliasesData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			al := m.aliasesData[i]
			name := al.Name
			value := al.Value

			displayName := truncateString(name, 12)
			displayValue := truncateString(value, availWidth-20)

			status := ""
			if warn, ok := aliases.DetectConflict(al.Name); ok && al.Enabled {
				if !strings.HasPrefix(warn, "collides with another active alias") {
					status = " ⚠️"
				}
			}

			if !al.Enabled {
				displayName = "[D] " + displayName
			}

			line := fmt.Sprintf("%-12s = %s%s", displayName, displayValue, status)
			line = truncateString(line, availWidth-4)

			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+line))
			} else {
				styled := "  " + line
				if !al.Enabled {
					list = append(list, TextMuted.Render(styled))
				} else {
					list = append(list, UnselectedStyle.Render(styled))
				}
			}
		}
		if end < len(m.aliasesData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}
		return lipgloss.JoinVertical(lipgloss.Left, list...)
	}
}
