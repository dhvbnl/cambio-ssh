package cli

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

var appFrameStyle = lipgloss.NewStyle().Padding(1, 2)

func appFrameSize() (int, int) {
	return appFrameStyle.GetFrameSize()
}

func renderApp(content string) string {
	return appFrameStyle.Render(content)
}

// keyHelpAdapter satisfies help.KeyMap for an arbitrary slice of bindings.
type keyHelpAdapter struct {
	keys []key.Binding
}

func (k keyHelpAdapter) ShortHelp() []key.Binding  { return k.keys }
func (k keyHelpAdapter) FullHelp() [][]key.Binding { return [][]key.Binding{k.keys} }

// renderKeyHelp renders a shared key-hint bar consistent across screens.
func renderKeyHelp(width int, bindings []key.Binding) string {
	h := help.New()
	h.Width = width
	h.ShowAll = false
	return h.View(keyHelpAdapter{keys: bindings})
}

// renderWithFooter pins a footer (like key help) to the bottom of the view area.
func renderWithFooter(content, footer string, width, height int) string {
	const topPad = 1
	pad := strings.Repeat("\n", topPad)
	paddedContent := pad + content
	if width > 0 && height > 0 {
		contentHeight := lipgloss.Height(content)
		footerHeight := lipgloss.Height(footer)
		spacerLines := height - contentHeight - footerHeight - topPad
		if spacerLines < 0 {
			spacerLines = 0
		}
		spacer := strings.Repeat("\n", spacerLines)
		return lipgloss.JoinVertical(lipgloss.Left, paddedContent, spacer, footer)
	}
	return lipgloss.JoinVertical(lipgloss.Left, paddedContent, footer)
}
