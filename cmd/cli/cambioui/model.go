package cambioui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	internalcambio "github.com/dhvbnl/cambio-ssh/cmd/internal/cambio"
)

type State int

const (
	StateInitial State = iota
	StatePlaying
	StateReplacingCard
	StateLookingAtOwnCard
	StateLookingAtOpponentCard
	StateShowingResult
	StatePlayAgain
	StateQuit
)

type Model struct {
	game             *internalcambio.Game
	state            State
	selectedCard     int
	selectedOpponent int
	playerID         int
	message          string
	width            int
	height           int
	keymap           keymap
	onBack           tea.Cmd
}

func NewModel(onBack tea.Cmd) Model {
	return Model{
		state:            StateInitial,
		selectedCard:     -1,
		selectedOpponent: -1,
		playerID:         0,
		keymap:           newKeymap(),
		onBack:           onBack,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.state = StateQuit
			return m, tea.Quit
		case key.Matches(msg, m.keymap.escape):
			if m.state == StateReplacingCard {
				m.resetSelectionState()
				m.state = StatePlaying
				return m, nil
			}
			if m.state == StateLookingAtOwnCard {
				if m.isPeeking() {
					m.finishOwnCardPeek()
					return m, nil
				}
				m.resetSelectionState()
				m.state = StatePlaying
				return m, nil
			}
			if m.state == StateLookingAtOpponentCard {
				if m.isPeeking() {
					m.finishOpponentCardPeek()
					return m, nil
				}
				m.resetSelectionState()
				m.state = StatePlaying
				return m, nil
			}
			if m.onBack != nil {
				return m, m.onBack
			}
			return m, nil
		case m.state == StateInitial && key.Matches(msg, m.keymap.join):
			m.game = internalcambio.NewGame(2)
			m.state = StatePlaying
			m.message = ""
			return m, nil
		case m.state == StatePlaying:
			m.handleGameplayKey(msg)
			return m, nil
		case m.state == StateReplacingCard:
			m.handleReplaceCardKey(msg)
			return m, nil
		case m.state == StateLookingAtOwnCard:
			m.handleLookOwnCardKey(msg)
			return m, nil
		case m.state == StateLookingAtOpponentCard:
			m.handleLookOpponentCardKey(msg)
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	var body strings.Builder
	body.WriteString(m.renderTitle())
	body.WriteString("\n")

	switch m.state {
	case StateInitial:
		body.WriteString(m.renderInitial())
	case StatePlaying, StateReplacingCard, StateLookingAtOwnCard, StateLookingAtOpponentCard:
		body.WriteString(m.renderPlaying())
	case StateShowingResult:
		body.WriteString(m.renderResult())
	case StatePlayAgain:
		body.WriteString(m.renderPlayAgain())
	}

	if m.message != "" {
		body.WriteString("\n")
		body.WriteString(messageStyle.Render(m.message))
		body.WriteString("\n")
	}

	helpStr := renderKeyHelp(m.width, m.getActiveKeybindings())
	content := renderWithFooter(body.String(), helpStr, m.width, m.height)
	return renderApp(content)
}

func (m Model) renderTitle() string {
	return titleStyle.Render("CAMBIO - 2 PLAYER")
}

func (m *Model) resetSelectionState() {
	m.selectedCard = -1
	m.selectedOpponent = -1
	m.message = ""
}

func (m Model) isPeeking() bool {
	if m.game == nil {
		return false
	}

	_, _, _, canSeeFace, revealedActive := m.game.GetRevealedCard(m.playerID)
	if !revealedActive || !canSeeFace {
		return false
	}

	return m.state == StateLookingAtOwnCard || m.state == StateLookingAtOpponentCard
}

func (m Model) playerCount() int {
	if m.game == nil {
		return 0
	}

	return len(m.game.GetAllPlayerHands())
}
