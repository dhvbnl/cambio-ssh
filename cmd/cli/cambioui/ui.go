package cambioui

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

type keyHelpAdapter struct {
	keys []key.Binding
}

func (k keyHelpAdapter) ShortHelp() []key.Binding {
	return k.keys
}

func (k keyHelpAdapter) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.keys}
}

func renderKeyHelp(width int, bindings []key.Binding) string {
	h := help.New()
	h.Width = width
	h.ShowAll = false
	return h.View(keyHelpAdapter{keys: bindings})
}

func renderWithFooter(content, footer string, width, height int) string {
	const topPad = 1
	paddedContent := strings.Repeat("\n", topPad) + content

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
