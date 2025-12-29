package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "205", Dark: "205"}). // Keep Pink/Magenta for now, maybe darker for light mode? "205" is quite bright. Let's try "161" for light.
			// Actually "205" (HotPink) might be hard to read on white. "161" (DeepPink3) is safe.
			Foreground(lipgloss.AdaptiveColor{Light: "161", Dark: "205"}).
			MarginLeft(2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "237", Dark: "241"}). // Darker grey for light mode (was 241)
			MarginLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "133", Dark: "170"}).
				Bold(true).
				PaddingLeft(2)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}). // Dark text for light mode, Light text for dark mode.
			PaddingLeft(2)

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "241"}) // Darker grey for light (was 245)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "196"}). // Red
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "42"}). // Green
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "241"}). // Darker grey for light (was 243)
			MarginTop(1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "57", Dark: "62"}). // Blue/Purple
			Padding(1, 2)

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "229", Dark: "229"}). // Light Yellow text
			Background(lipgloss.AdaptiveColor{Light: "57", Dark: "57"}). // Blue background
			Padding(0, 1).
			MarginRight(1)

	favoriteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "178", Dark: "226"}). // Yellow/Gold
			Bold(true)

	languageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "39", Dark: "117"}). // Blue/Cyan
			Italic(true)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "161", Dark: "205"}).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.AdaptiveColor{Light: "57", Dark: "62"}).
			MarginBottom(1)

	codeBlockStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "57", Dark: "62"}).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "236"}). // Light grey for light, Dark grey for dark
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "161", Dark: "205"}).
				Background(lipgloss.AdaptiveColor{Light: "254", Dark: "236"}).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "161", Dark: "205"})
)
