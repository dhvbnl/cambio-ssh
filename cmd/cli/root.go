package cli

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screenID string

const (
	screenHome      screenID = "home"
	screenBlackjack screenID = "blackjack"
	screenChat      screenID = "chat"
)

type navMsg struct {
	target screenID
}

func navigate(target screenID) tea.Cmd {
	return func() tea.Msg { return navMsg{target: target} }
}

// RootModel orchestrates navigation between the home menu and feature screens.
type RootModel struct {
	active    screenID
	screens   map[screenID]tea.Model
	factories map[screenID]func() tea.Model
	lastSize  *tea.WindowSizeMsg
}

func NewRootModel() RootModel {
	factories := map[screenID]func() tea.Model{
		screenHome:      func() tea.Model { return NewMenuModel() },
		screenBlackjack: func() tea.Model { return NewGameModel() },
		screenChat:      func() tea.Model { return NewChatModel() },
	}

	screens := make(map[screenID]tea.Model, len(factories))
	screens[screenHome] = factories[screenHome]()

	return RootModel{
		active:    screenHome,
		screens:   screens,
		factories: factories,
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.screens[m.active].Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case navMsg:
		builder, ok := m.factories[msg.target]
		if !ok {
			return m, nil
		}
		m.active = msg.target
		screen := builder()
		var sizeCmd tea.Cmd
		if m.lastSize != nil {
			var cmd tea.Cmd
			screen, cmd = screen.Update(*m.lastSize)
			sizeCmd = cmd
		}
		m.screens[m.active] = screen
		return m, tea.Batch(m.screens[m.active].Init(), sizeCmd)
	case tea.WindowSizeMsg:
		m.lastSize = &msg
		if current, ok := m.screens[m.active]; ok {
			updated, cmd := current.Update(msg)
			m.screens[m.active] = updated
			return m, cmd
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	current := m.screens[m.active]
	updated, cmd := current.Update(msg)
	m.screens[m.active] = updated
	return m, cmd
}

func (m RootModel) View() string {
	if view, ok := m.screens[m.active]; ok {
		return view.View()
	}
	return "Unknown view"
}

// Home menu screen ----------------------------------------------------------------

type menuItem struct {
	title       string
	description string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.description }
func (i menuItem) FilterValue() string { return i.title }

type menuStyles struct {
	title lipgloss.Style
}

func newMenuStyles() menuStyles {
	return menuStyles{
		title: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1),
	}
}

type MenuModel struct {
	list   list.Model
	styles menuStyles
	width  int
	height int
	keys   []key.Binding
}

func NewMenuModel() MenuModel {
	items := []list.Item{
		menuItem{title: "Blackjack", description: "Play the blackjack game"},
		menuItem{title: "Chat (placeholder)", description: "Stub screen for future chat"},
	}

	keys := []key.Binding{
		key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "choose")),
		key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
	}

	styles := newMenuStyles()
	delegate := list.NewDefaultDelegate()

	l := list.New(items, delegate, 0, 0)
	l.Title = "Welcome to Cambio SSH"
	l.Styles.Title = styles.title
	l.DisableQuitKeybindings()
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowHelp(false)

	return MenuModel{list: l, styles: styles, keys: keys}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appFrameSize()
		contentW := msg.Width - h
		contentH := msg.Height - v
		m.list.SetSize(contentW, contentH)
		m.width = contentW
		m.height = contentH
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, m.activateSelection()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MenuModel) View() string {
	content := renderWithFooter(m.list.View(), renderKeyHelp(m.width, m.keys), m.width, m.height)
	return renderApp(content)
}

func (m MenuModel) activateSelection() tea.Cmd {
	if item, ok := m.list.SelectedItem().(menuItem); ok {
		switch item.title {
		case "Blackjack":
			return navigate(screenBlackjack)
		case "Chat (placeholder)":
			return navigate(screenChat)
		}
	}
	return nil
}
