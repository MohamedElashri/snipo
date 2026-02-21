package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("13")). // Bright Magenta (ANSI 13)
			MarginLeft(2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Bright Black (Grey) (ANSI 8) - universally works as dimmed
			MarginLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("5")). // Magenta (ANSI 5)
				Bold(true).
				PaddingLeft(2)

	normalItemStyle = lipgloss.NewStyle().
		// No Foreground set implies terminal default (Best for adaptive text)
		PaddingLeft(2)

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // Bright Black (ANSI 8)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")). // Red (ANSI 1)
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")). // Green (ANSI 2)
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Bright Black (ANSI 8)
			MarginTop(1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("5")). // Magenta (ANSI 5) - simple, standard accent
			Padding(1, 2)

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")). // Black text (ANSI 0)
			Background(lipgloss.Color("4")). // Blue background (ANSI 4)
			Padding(0, 1).
			MarginRight(1)

	favoriteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")). // Yellow (ANSI 3)
			Bold(true)

	languageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")). // Cyan (ANSI 6)
			Italic(true)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("13")). // Bright Magenta (ANSI 13)
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("5")). // Magenta
			MarginBottom(1)

	codeBlockStyle = lipgloss.NewStyle().
		// No Foreground set - adapt to terminal default logic
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")). // Magenta (ANSI 5)
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
		// Default foreground
		Foreground(lipgloss.Color("7")). // White/Light Grey (standard text)
		Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")). // Bright White
				Padding(0, 1)

	focusedPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("5")). // Magenta
				Bold(true).
				Padding(0, 1)
)

func renderHelpText(text string) string {
	parts := strings.Split(text, " • ")
	var renderedParts []string

	shortcutKeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))  // Magenta
	shortcutDescStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Grey

	for _, part := range parts {
		fields := strings.SplitN(part, " ", 2)
		if len(fields) == 2 {
			renderedParts = append(renderedParts, shortcutKeyStyle.Render(fields[0])+" "+shortcutDescStyle.Render(fields[1]))
		} else {
			renderedParts = append(renderedParts, shortcutDescStyle.Render(part))
		}
	}

	return strings.Join(renderedParts, shortcutDescStyle.Render(" • "))
}
