package tui

import (
	"fmt"
	"strings"
)

type GitComponent struct{}

func (g GitComponent) View(m model) string {
	if m.gitHistoryView {
		title := TitleStyle.Render("Workspace History (Version Control)")
		if len(m.gitCommits) == 0 {
			preview := fmt.Sprintf("%s\n\nNo workspace history found.", title)
			return CardStyle.Width(m.width - 30).Render(preview)
		}

		var historyLines []string
		start, end := m.getVisibleSlice(len(m.gitCommits))
		for i := start; i < end; i++ {
			c := m.gitCommits[i]
			// format commit line: hash | timestamp | message
			hashStr := TextMuted.Render(c.Hash)
			timeStr := SuccessLabel.Render(c.Timestamp)
			msgStr := c.Message

			// Strip the prefix [reshell] for cleaner visual presentation
			msgStr = strings.TrimPrefix(msgStr, "[reshell] ")

			line := fmt.Sprintf("%s  %s  %s", hashStr, timeStr, msgStr)
			if i == m.gitSelectedIdx {
				historyLines = append(historyLines, SelectedStyle.Render("> "+line))
			} else {
				historyLines = append(historyLines, "  "+line)
			}
		}

		preview := fmt.Sprintf("%s\n\n%s", title, strings.Join(historyLines, "\n"))
		return CardStyle.Width(m.width - 30).Render(preview)
	}

	gitContent := "No global Git configurations read."
	if m.gitData != nil {
		gitContent = fmt.Sprintf("Name: %s\nEmail: %s\nGPG Signing: %t\nSigning Key: %s\n\nGlobal Aliases:\n",
			m.gitData.UserName, m.gitData.UserEmail, m.gitData.GpgSign, m.gitData.SigningKey,
		)
		for alias, value := range m.gitData.Aliases {
			gitContent += fmt.Sprintf("  %s = %s\n", alias, value)
		}
	}

	preview := fmt.Sprintf("%s\n%s", TitleStyle.Render("Git global config"), gitContent)
	return CardStyle.Width(m.width - 30).Render(preview)
}
