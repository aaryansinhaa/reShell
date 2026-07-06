package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Palette Colors (dynamic)
	BgColor        lipgloss.Color
	HeaderBg       lipgloss.Color
	IndigoColor    lipgloss.Color
	PurpleColor    lipgloss.Color
	PinkColor      lipgloss.Color
	CyanColor      lipgloss.Color
	GreenColor     lipgloss.Color
	YellowColor    lipgloss.Color
	RedColor       lipgloss.Color
	GrayColor      lipgloss.Color
	TextMutedColor lipgloss.Color
	TextBright     lipgloss.Color
	SelectionBg    lipgloss.Color

	// Structural Styles
	AppStyle     lipgloss.Style
	HeaderStyle  lipgloss.Style
	SidebarStyle lipgloss.Style
	MainStyle    lipgloss.Style
	HelpStyle    lipgloss.Style

	// Interactive Elements
	TabActiveStyle   lipgloss.Style
	TabInactiveStyle lipgloss.Style
	TitleStyle       lipgloss.Style
	SelectedStyle    lipgloss.Style
	UnselectedStyle  lipgloss.Style

	// Status Labels
	SuccessLabel lipgloss.Style
	WarningLabel lipgloss.Style
	ErrorLabel   lipgloss.Style

	// Card/Border Containers
	CardStyle lipgloss.Style
	TextMuted lipgloss.Style
)

// InitTheme initializes the active theme colors and generates Lipgloss styling elements.
func InitTheme(themeName string) {
	switch themeName {
	case "light":
		BgColor = lipgloss.Color("#eff1f5")
		HeaderBg = lipgloss.Color("#dce0e8")
		IndigoColor = lipgloss.Color("#1e66f5")
		PurpleColor = lipgloss.Color("#8839ef")
		PinkColor = lipgloss.Color("#ea76cb")
		CyanColor = lipgloss.Color("#04a5e5")
		GreenColor = lipgloss.Color("#40a02b")
		YellowColor = lipgloss.Color("#df8e1d")
		RedColor = lipgloss.Color("#d20f39")
		GrayColor = lipgloss.Color("#bcc0cc")
		TextMutedColor = lipgloss.Color("#6c6f85")
		TextBright = lipgloss.Color("#202020")
		SelectionBg = lipgloss.Color("#ccd0da")

	case "catppuccin":
		BgColor = lipgloss.Color("#24273a")
		HeaderBg = lipgloss.Color("#363a4f")
		IndigoColor = lipgloss.Color("#8aadf4")
		PurpleColor = lipgloss.Color("#c6a0f6")
		PinkColor = lipgloss.Color("#f5bde6")
		CyanColor = lipgloss.Color("#91d7e3")
		GreenColor = lipgloss.Color("#a6da95")
		YellowColor = lipgloss.Color("#eed49f")
		RedColor = lipgloss.Color("#ed8796")
		GrayColor = lipgloss.Color("#5b6078")
		TextMutedColor = lipgloss.Color("#a5adcb")
		TextBright = lipgloss.Color("#cad3f5")
		SelectionBg = lipgloss.Color("#494d64")

	case "gruvbox":
		BgColor = lipgloss.Color("#282828")
		HeaderBg = lipgloss.Color("#3c3836")
		IndigoColor = lipgloss.Color("#83a598")
		PurpleColor = lipgloss.Color("#d3869b")
		PinkColor = lipgloss.Color("#d3869b")
		CyanColor = lipgloss.Color("#8ec07c")
		GreenColor = lipgloss.Color("#b8bb26")
		YellowColor = lipgloss.Color("#fabd2f")
		RedColor = lipgloss.Color("#fb4934")
		GrayColor = lipgloss.Color("#504945")
		TextMutedColor = lipgloss.Color("#a89984")
		TextBright = lipgloss.Color("#ebdbb2")
		SelectionBg = lipgloss.Color("#665c54")

	case "tokyo-night":
		BgColor = lipgloss.Color("#1a1b26")
		HeaderBg = lipgloss.Color("#24283b")
		IndigoColor = lipgloss.Color("#7aa2f7")
		PurpleColor = lipgloss.Color("#bb9af7")
		PinkColor = lipgloss.Color("#f7768e")
		CyanColor = lipgloss.Color("#7db9f5")
		GreenColor = lipgloss.Color("#9ece6a")
		YellowColor = lipgloss.Color("#e0af68")
		RedColor = lipgloss.Color("#f7768e")
		GrayColor = lipgloss.Color("#414868")
		TextMutedColor = lipgloss.Color("#a9b1d6")
		TextBright = lipgloss.Color("#c0caf5")
		SelectionBg = lipgloss.Color("#2f3549")

	default: // "dark"
		BgColor = lipgloss.Color("#1e1e2e")
		HeaderBg = lipgloss.Color("#313244")
		IndigoColor = lipgloss.Color("#89b4fa")
		PurpleColor = lipgloss.Color("#cba6f7")
		PinkColor = lipgloss.Color("#f5c2e7")
		CyanColor = lipgloss.Color("#89dceb")
		GreenColor = lipgloss.Color("#a6e3a1")
		YellowColor = lipgloss.Color("#f9e2af")
		RedColor = lipgloss.Color("#f38ba8")
		GrayColor = lipgloss.Color("#585b70")
		TextMutedColor = lipgloss.Color("#a6adc8")
		TextBright = lipgloss.Color("#cdd6f4")
		SelectionBg = lipgloss.Color("#45475a")
	}

	// Recompile Style structures using updated theme palettes (Translucent layout - background colors removed except for active selections/tabs)
	AppStyle = lipgloss.NewStyle().
		Foreground(TextBright)

	HeaderStyle = lipgloss.NewStyle().
		Foreground(IndigoColor).
		Padding(1, 2).
		Bold(true).
		Border(lipgloss.DoubleBorder(), false, false, true, false).
		BorderForeground(GrayColor)

	SidebarStyle = lipgloss.NewStyle().
		Foreground(TextBright).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(GrayColor).
		Padding(1, 1).
		Width(22)

	MainStyle = lipgloss.NewStyle().
		Foreground(TextBright).
		Padding(1, 2)

	HelpStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(GrayColor).
		Padding(0, 2).
		Foreground(TextMutedColor)

	TabActiveStyle = lipgloss.NewStyle().
		Background(PurpleColor).
		Foreground(lipgloss.Color("#11111b")).
		Padding(0, 1).
		Bold(true)

	TabInactiveStyle = lipgloss.NewStyle().
		Foreground(TextMutedColor).
		Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
		Foreground(IndigoColor).
		Bold(true).
		MarginBottom(1)

	SelectedStyle = lipgloss.NewStyle().
		Background(SelectionBg).
		Foreground(PurpleColor).
		Padding(0, 1).
		Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
		Foreground(TextBright).
		Padding(0, 1)

	SuccessLabel = lipgloss.NewStyle().Foreground(GreenColor).Bold(true)
	WarningLabel = lipgloss.NewStyle().Foreground(YellowColor).Bold(true)
	ErrorLabel = lipgloss.NewStyle().Foreground(RedColor).Bold(true)

	CardStyle = lipgloss.NewStyle().
		Foreground(TextBright).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(GrayColor).
		Padding(1, 2).
		MarginBottom(1)

	TextMuted = lipgloss.NewStyle().Foreground(TextMutedColor)
}

func HighlightCode(code, lexer string) string {
	var buf bytes.Buffer
	err := quick.Highlight(&buf, code, lexer, "terminal256", "monokai")
	if err != nil {
		return code
	}
	return buf.String()
}

func GetTruncatedCodeBlock(code, lexer string, maxLines int) string {
	highlighted := HighlightCode(code, lexer)
	lines := strings.Split(highlighted, "\n")
	totalLines := len(lines)
	if totalLines <= maxLines {
		return highlighted
	}

	truncatedLines := lines[:maxLines]
	joined := strings.Join(truncatedLines, "\n")
	notice := "\n" + TextMuted.Render(fmt.Sprintf("... (truncated, %d lines remaining. Press Enter to copy)", totalLines-maxLines))
	return joined + notice
}
