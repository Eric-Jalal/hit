package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#7C3AED")
	ColorSecondary = lipgloss.Color("#06B6D4")
	ColorSuccess   = lipgloss.Color("#22C55E")
	ColorWarning   = lipgloss.Color("#EAB308")
	ColorError     = lipgloss.Color("#EF4444")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorText      = lipgloss.Color("#F9FAFB")
	ColorBg        = lipgloss.Color("#111827")

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 2)

	TabBarStyle = lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted).
			MarginBottom(1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	BadgeSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	BadgeFailure = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	BadgePending = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	BadgeNeutral = lipgloss.NewStyle().
			Foreground(ColorMuted)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ErrorLineStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)
