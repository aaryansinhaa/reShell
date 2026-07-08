package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type WorkflowsComponent struct{}

func (w WorkflowsComponent) View(m model) string {
	if len(m.workflowsData) == 0 {
		descText := "Workflows execute a series of bash commands sequentially in specified directories, halting on failures.\n\nHow to write a workflow:\n1. Define them in ~/.config/reshell/workflows.toml\n2. Template Structure:"

		tomlExample := `[[workflows]]
name = "deploy-web"
description = "Build and upload frontend dist"

  [[workflows.steps]]
  command = "npm run build"
  dir = "~/projects/frontend"
  comment = "Build production bundle"

  [[workflows.steps]]
  command = "scp -r ./dist user@host:/var/www"
  dir = "~/projects/frontend"
  comment = "Transfer assets"`

		highlighted := HighlightCode(tomlExample, "toml")
		borderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GrayColor).
			Padding(0, 1)

		codeBox := borderStyle.Render(highlighted)

		footerHelp := TextMuted.Render("Press 'n' to initialize a new workflow template and open workflows.toml in your editor.")

		content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
			TitleStyle.Render("No Workflows Configured Yet"),
			TextMuted.Render(descText),
			codeBox,
			footerHelp,
		)

		return CardStyle.Width(m.width - 30).Render(content)
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

		start, end := m.getVisibleSlice(len(m.workflowsData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			wf := m.workflowsData[i]
			displayName := truncateString(wf.Name, leftWidth-5)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render("  "+displayName))
			}
		}
		if end < len(m.workflowsData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}

		leftCol := lipgloss.JoinVertical(lipgloss.Left, list...)

		selected := m.workflowsData[m.selectedIdx]
		stepsView := strings.Builder{}
		for idx, step := range selected.Steps {
			marker := "  "
			if m.runningWorkflow != nil && m.runningWorkflow.Name == selected.Name {
				if idx < len(m.wfStepsStatus) {
					status := m.wfStepsStatus[idx]
					if !status.Finished {
						marker = "⏳ "
					} else if status.Error != nil {
						marker = "✘ "
					} else {
						marker = "✔ "
					}
				}
			}
			stepsView.WriteString(fmt.Sprintf("%s%d. %s (dir: %s)\n", marker, idx+1, step.Command, step.Dir))
		}

		preview := fmt.Sprintf("%s\n%s\n\nSteps Sequence:\n%s",
			TitleStyle.Render("Workflow: "+selected.Name),
			TextMuted.Render(selected.Description),
			stepsView.String(),
		)

		previewCard := CardStyle.Width(cardWidth).Render(preview)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, previewCard)
	} else {
		start, end := m.getVisibleSlice(len(m.workflowsData))
		var list []string
		if start > 0 {
			list = append(list, TextMuted.Render("  ▲ ..."))
		}
		for i := start; i < end; i++ {
			wf := m.workflowsData[i]
			displayName := truncateString(wf.Name, availWidth-5)
			if i == m.selectedIdx {
				list = append(list, SelectedStyle.Render("> "+displayName))
			} else {
				list = append(list, UnselectedStyle.Render("  "+displayName))
			}
		}
		if end < len(m.workflowsData) {
			list = append(list, TextMuted.Render("  ▼ ..."))
		}
		return lipgloss.JoinVertical(lipgloss.Left, list...)
	}
}

