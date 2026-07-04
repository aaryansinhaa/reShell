package tui

import (
	"fmt"
	"reshell/pkg/packages"

	"github.com/charmbracelet/lipgloss"
)

type PackagesComponent struct{}

func (p PackagesComponent) View(m model) string {
	if len(m.packagesData) == 0 {
		return "No packages listed in config.toml. Press 'n' to add a package requirement."
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

		osName, manager := packages.DetectOS()
		start, end := m.getVisibleSlice(len(m.packagesData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			pkg := m.packagesData[i]
			status := ErrorLabel.Render("Missing ⚠️")
			if m.pkgStatus[pkg] {
				status = SuccessLabel.Render("Installed ✔")
			}

			maxPkgLen := leftWidth - 18
			if maxPkgLen < 8 {
				maxPkgLen = 8
			}
			displayName := truncateString(pkg, maxPkgLen)
			line := fmt.Sprintf("%-*s [%s]", maxPkgLen, displayName, status)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+line))
			} else {
				list = append(list, UnselectedStyle.Render("  "+line))
			}
		}
		if end < len(m.packagesData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}

		leftCol := lipgloss.JoinVertical(lipgloss.Left, list...)

		preview := fmt.Sprintf("%s\nDetected OS: %s\nDefault package manager: %s\n\nPress 'i' to trigger automated synchronized install for all packages.",
			TitleStyle.Render("Package Requirements List"),
			osName, manager,
		)

		previewCard := CardStyle.Width(cardWidth).Render(preview)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, previewCard)
	} else {
		start, end := m.getVisibleSlice(len(m.packagesData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			pkg := m.packagesData[i]
			status := "⚠️"
			if m.pkgStatus[pkg] {
				status = "✔"
			}
			displayName := truncateString(pkg, availWidth-10)
			line := fmt.Sprintf("%s [%s]", displayName, status)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+line))
			} else {
				list = append(list, UnselectedStyle.Render("  "+line))
			}
		}
		if end < len(m.packagesData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}
		return lipgloss.JoinVertical(lipgloss.Left, list...)
	}
}
