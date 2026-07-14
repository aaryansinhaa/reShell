package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type MarketplaceComponent struct{}

func (mp MarketplaceComponent) View(m model) string {
	if m.remoteSyncURL == "" {
		preview := `Browse & install reshell terminal profiles.
Specify repositories to automatically import setup configurations (aliases, variables, functions, and package lists).

Example repositories:
- github.com/aaryansinhaa/reshell-java (Setup JDK, Maven, workspace env-vars, aliases)
- github.com/aaryansinhaa/reshell-react (Setup TypeScript environment, formatters, and templates)

Press 'i' key to install a package.`

		return CardStyle.Width(m.width - 30).Render(preview)
	}

	availWidth := m.width - 30
	if availWidth < 45 {
		availWidth = 45
	}

	// Gather metrics
	lastSyncStr := m.lastSync
	if lastSyncStr == "" {
		lastSyncStr = "Never"
	}
	forksStr := fmt.Sprintf("%d", m.forksCount)
	starsStr := fmt.Sprintf("%d", m.starsCount)
	issuesStr := fmt.Sprintf("%d", m.openIssuesCount)
	lastUpdatedStr := m.lastUpdated
	if lastUpdatedStr == "" {
		lastUpdatedStr = "N/A"
	}

	// Extract owner and repository name dynamically from remote sync URL
	ownerName := "remote"
	repoName := "repository"
	if parts := strings.Split(m.remoteSyncURL, "/"); len(parts) >= 2 {
		repoName = strings.TrimSuffix(parts[len(parts)-1], ".git")
		ownerName = parts[len(parts)-2]
		if strings.Contains(ownerName, ":") {
			subParts := strings.Split(ownerName, ":")
			ownerName = subParts[len(subParts)-1]
		}
	}

	repoHeader := fmt.Sprintf("📦 %s / %s",
		lipgloss.NewStyle().Foreground(TextMutedColor).Render(ownerName),
		lipgloss.NewStyle().Foreground(IndigoColor).Bold(true).Render(repoName),
	)

	statusColor := GreenColor
	statusText := "✔ Up to date"
	if m.updatesCount > 0 {
		statusColor = YellowColor
		statusText = fmt.Sprintf("⚠️  %d commits behind", m.updatesCount)
	}

	statusBadge := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true).
		Render(statusText)

	metricsRow := fmt.Sprintf("⭐ %-4s  🍴 %-4s  ☉ %-4s open issues  |  📅 Push: %s  |  Status: %s",
		starsStr,
		forksStr,
		issuesStr,
		lastUpdatedStr,
		statusBadge,
	)

	keysRow := fmt.Sprintf("Keys: %s  |  %s  |  URL: %s",
		lipgloss.NewStyle().Foreground(CyanColor).Render("[i] Install Pack"),
		lipgloss.NewStyle().Foreground(PurpleColor).Render("[s] Sync Remote"),
		TextMuted.Render(m.remoteSyncURL),
	)

	topContent := fmt.Sprintf("%s\n\n%s\n%s",
		repoHeader,
		metricsRow,
		keysRow,
	)

	wrappedTop := lipgloss.NewStyle().Width(availWidth - 8).Render(topContent)
	topCard := CardStyle.Width(availWidth).MarginBottom(0).Render(wrappedTop)

	// Bottom Card: README (Taking full width, auto-truncated to remaining height)
	var bottomCard string
	if m.readmeContent == "" {
		bottomCard = CardStyle.Width(availWidth).Render(TitleStyle.Render("README.md") + "\n\n(No README found in workspace configuration)")
	} else {
		// Calculate available height dynamically
		topHeight := lipgloss.Height(topCard)
		maxHeight := m.mainHeight - topHeight - 3
		if maxHeight < 8 {
			maxHeight = 8
		}
		// Subtract borders/padding from maxHeight to get max lines for content
		maxLines := maxHeight - 4
		if maxLines < 4 {
			maxLines = 4
		}

		renderedMD := renderMarkdown(m.readmeContent, availWidth-8)
		lines := strings.Split(renderedMD, "\n")

		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, TextMuted.Render(fmt.Sprintf("\n... (%d lines remaining)", len(strings.Split(renderedMD, "\n"))-maxLines)))
		}
		truncatedMD := strings.Join(lines, "\n")

		wrappedBottom := TitleStyle.Render("README.md") + "\n" + truncatedMD
		bottomCard = CardStyle.Width(availWidth).Height(maxHeight).Render(wrappedBottom)
	}

	return lipgloss.JoinVertical(lipgloss.Left, topCard, bottomCard)
}

func wrapLine(text string, limit int) string {
	if len(text) <= limit || limit <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	var currentLine []string
	currentLen := 0
	for _, w := range words {
		if currentLen+len(w)+1 > limit && currentLen > 0 {
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{w}
			currentLen = len(w)
		} else {
			currentLine = append(currentLine, w)
			if currentLen == 0 {
				currentLen = len(w)
			} else {
				currentLen += len(w) + 1
			}
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}
	return strings.Join(lines, "\n")
}

func renderMarkdown(md string, limit int) string {
	lines := strings.Split(md, "\n")
	var rendered []string
	inCodeBlock := false

	// Define Regexes for inline styles
	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	reCode := regexp.MustCompile("`(.*?)`")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for code blocks
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			// Truncate code lines to fit limit, avoiding code wraps
			if len(line) > limit-4 {
				line = line[:limit-7] + "..."
			}
			rendered = append(rendered, lipgloss.NewStyle().Foreground(PinkColor).Background(SelectionBg).Render("  "+line))
			continue
		}

		// Headers
		if strings.HasPrefix(trimmed, "# ") {
			title := strings.TrimPrefix(trimmed, "# ")
			wrapped := wrapLine(title, limit)
			prefix := ""
			if len(rendered) > 0 {
				prefix = "\n"
			}
			rendered = append(rendered, prefix+lipgloss.NewStyle().Foreground(IndigoColor).Bold(true).Render(wrapped))
		} else if strings.HasPrefix(trimmed, "## ") {
			title := strings.TrimPrefix(trimmed, "## ")
			wrapped := wrapLine(title, limit)
			prefix := ""
			if len(rendered) > 0 {
				prefix = "\n"
			}
			rendered = append(rendered, prefix+lipgloss.NewStyle().Foreground(PurpleColor).Bold(true).Render(wrapped))
		} else if strings.HasPrefix(trimmed, "### ") {
			title := strings.TrimPrefix(trimmed, "### ")
			wrapped := wrapLine(title, limit)
			prefix := ""
			if len(rendered) > 0 {
				prefix = "\n"
			}
			rendered = append(rendered, prefix+lipgloss.NewStyle().Foreground(CyanColor).Bold(true).Render(wrapped))
		} else if strings.HasPrefix(trimmed, "- ") {
			item := strings.TrimPrefix(trimmed, "- ")
			wrapped := wrapLine(item, limit-4)
			wrappedLines := strings.Split(wrapped, "\n")
			for i, wl := range wrappedLines {
				wl = reBold.ReplaceAllStringFunc(wl, func(match string) string {
					content := match[2 : len(match)-2]
					return lipgloss.NewStyle().Bold(true).Foreground(TextBright).Render(content)
				})
				wl = reCode.ReplaceAllStringFunc(wl, func(match string) string {
					content := match[1 : len(match)-1]
					return lipgloss.NewStyle().Foreground(PinkColor).Background(SelectionBg).Padding(0, 1).Render(content)
				})
				if i == 0 {
					rendered = append(rendered, "  • "+wl)
				} else {
					rendered = append(rendered, "    "+wl)
				}
			}
		} else if strings.HasPrefix(trimmed, "* ") {
			item := strings.TrimPrefix(trimmed, "* ")
			wrapped := wrapLine(item, limit-4)
			wrappedLines := strings.Split(wrapped, "\n")
			for i, wl := range wrappedLines {
				wl = reBold.ReplaceAllStringFunc(wl, func(match string) string {
					content := match[2 : len(match)-2]
					return lipgloss.NewStyle().Bold(true).Foreground(TextBright).Render(content)
				})
				wl = reCode.ReplaceAllStringFunc(wl, func(match string) string {
					content := match[1 : len(match)-1]
					return lipgloss.NewStyle().Foreground(PinkColor).Background(SelectionBg).Padding(0, 1).Render(content)
				})
				if i == 0 {
					rendered = append(rendered, "  • "+wl)
				} else {
					rendered = append(rendered, "    "+wl)
				}
			}
		} else if strings.HasPrefix(trimmed, "- ") {
			item := strings.TrimPrefix(trimmed, "- ")
			wrapped := wrapLine(item, limit-4)
			wrappedLines := strings.Split(wrapped, "\n")
			for i, wl := range wrappedLines {
				wl = reBold.ReplaceAllStringFunc(wl, func(match string) string {
					content := match[2 : len(match)-2]
					return lipgloss.NewStyle().Bold(true).Foreground(TextBright).Render(content)
				})
				wl = reCode.ReplaceAllStringFunc(wl, func(match string) string {
					content := match[1 : len(match)-1]
					return lipgloss.NewStyle().Foreground(PinkColor).Background(SelectionBg).Padding(0, 1).Render(content)
				})
				if i == 0 {
					rendered = append(rendered, "  • "+wl)
				} else {
					rendered = append(rendered, "    "+wl)
				}
			}
		} else if strings.HasPrefix(trimmed, "> ") {
			quote := strings.TrimPrefix(trimmed, "> ")
			wrapped := wrapLine(quote, limit-4)
			wrappedLines := strings.Split(wrapped, "\n")
			for _, wl := range wrappedLines {
				rendered = append(rendered, lipgloss.NewStyle().Foreground(YellowColor).Italic(true).Render("  │ "+wl))
			}
		} else {
			wrapped := wrapLine(line, limit)
			wrappedLines := strings.Split(wrapped, "\n")
			for _, wl := range wrappedLines {
				wl = reBold.ReplaceAllStringFunc(wl, func(match string) string {
					content := match[2 : len(match)-2]
					return lipgloss.NewStyle().Bold(true).Foreground(TextBright).Render(content)
				})
				wl = reCode.ReplaceAllStringFunc(wl, func(match string) string {
					content := match[1 : len(match)-1]
					return lipgloss.NewStyle().Foreground(PinkColor).Background(SelectionBg).Padding(0, 1).Render(content)
				})
				rendered = append(rendered, wl)
			}
		}
	}
	return strings.Join(rendered, "\n")
}
