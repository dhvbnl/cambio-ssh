package cambioui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	boxStyle         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 1)
	selectedBoxStyle = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("10")).Padding(0, 1)
	messageStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	cardStyle                = newCardStyle(lipgloss.RoundedBorder(), "8")
	candidateCardStyle       = newCardStyle(lipgloss.DoubleBorder(), "8")
	selectedReplaceCardStyle = newCardStyle(lipgloss.DoubleBorder(), "10")
	selectedLookCardStyle    = newCardStyle(lipgloss.DoubleBorder(), "6")

	redCardStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	blackCardStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	cardBackStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

func newCardStyle(border lipgloss.Border, borderColor string) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(5).
		Align(lipgloss.Center)
}
