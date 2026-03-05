package cli

import (
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/chat"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	focusStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))
)

type ChatModel struct {
	width    int
	height   int
	viewport viewport.Model
	chat     *chat.Chat
	textarea textarea.Model
}

func NewChatModel() ChatModel {
	//
	textarea := textarea.New()
	textarea.Placeholder = "send a message..."
	textarea.Focus()

	textarea.Prompt = "| "
	textarea.CharLimit = 256

	textarea.SetWidth(30)
	textarea.SetHeight(1)

	textarea.FocusedStyle.CursorLine = focusStyle

	textarea.ShowLineNumbers = false

	viewport := viewport.New(30, 10)
	viewport.SetContent("Welcome to the chat!")
	viewport.KeyMap.Left.SetEnabled(false)
	viewport.KeyMap.Right.SetEnabled(false)

	textarea.KeyMap.InsertNewline.SetEnabled(false)

	chat := chat.NewChat()

	return ChatModel{
		textarea: textarea,
		viewport: viewport,
		chat:     chat,
	}
}

func (m ChatModel) Init() tea.Cmd { return textarea.Blink }

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
		m.textarea.SetWidth(m.width)

		headerHeight := 2
		usableHeight := m.height - m.textarea.Height() - headerHeight
		if usableHeight < 0 {
			usableHeight = 0
		}
		m.viewport.Width = m.width
		m.viewport.Height = usableHeight

		messages := m.chat.GetFormattedMessages()
		if len(messages) > 0 {
			formattedMessages := lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(messages, "\n"))
			m.viewport.SetContent(formattedMessages)
		}
		m.viewport.GotoBottom()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "b", "esc":
			return m, navigate(screenHome)
		case "enter":
			content := m.textarea.Value()
			if strings.TrimSpace(content) != "" {
				m.chat.AddMessage("You", content)
				m.textarea.Reset()

				messages := m.chat.GetFormattedMessages()
				formattedMessages := lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(messages, "\n"))
				m.viewport.SetContent(formattedMessages)
				m.viewport.GotoBottom()
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
	case cursor.BlinkMsg:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m ChatModel) View() string {
	viewportView := m.viewport.View()
	textareaView := m.textarea.View()

	header := titleStyle.Render("Chat")
	body := lipgloss.JoinVertical(lipgloss.Left, header, "", viewportView, textareaView)

	keys := []key.Binding{
		key.NewBinding(key.WithKeys("b", "esc"), key.WithHelp("b/esc", "back")),
		key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	}

	content := renderWithFooter(body, renderKeyHelp(m.width, keys), m.width, m.height)
	return renderApp(content)
}
