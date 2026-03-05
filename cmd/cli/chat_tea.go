package cli

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ChatPlaceholderModel struct {
	width  int
	height int
}

func NewChatPlaceholder() ChatPlaceholderModel {
	return ChatPlaceholderModel{}
}

func (m ChatPlaceholderModel) Init() tea.Cmd { return nil }

func (m ChatPlaceholderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "b", "esc":
			return m, navigate(screenHome)
		}
	}
	return m, nil
}

func (m ChatPlaceholderModel) View() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	b.WriteString(titleStyle.Render("Chat (coming soon)"))
	b.WriteString("\n")
	b.WriteString(bodyStyle.Render("This is a placeholder for a future chat experience."))
	b.WriteString("\n\n")
	b.WriteString(bodyStyle.Render("Press b or Esc to go back, q to quit."))

	keys := []key.Binding{
		key.NewBinding(key.WithKeys("b", "esc"), key.WithHelp("b/esc", "back")),
		key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}

	content := renderWithFooter(b.String(), renderKeyHelp(m.width, keys), m.width, m.height)
	return renderApp(content)
}
