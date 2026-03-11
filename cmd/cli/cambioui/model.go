package cambioui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	internalcambio "github.com/dhvbnl/cambio-ssh/cmd/internal/cambio"
)

type State int

const (
	StateInitial State = iota
	StateJoinGameList
	StateWaitingForJoin
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
	username         string
	clientID         string
	lobbyID          string
	joinList         list.Model
	playerID         int
	message          string
	width            int
	height           int
	keymap           keymap
	onBack           tea.Cmd
}

type lobbyTickMsg struct{}

type lobbyListItem struct {
	id       string
	hostName string
}

func (i lobbyListItem) Title() string       { return i.hostName + "'s game" }
func (i lobbyListItem) Description() string { return "Press ENTER to join" }
func (i lobbyListItem) FilterValue() string { return i.hostName }

func NewModel(onBack tea.Cmd, username string, clientID string) Model {
	delegate := list.NewDefaultDelegate()
	joinList := list.New([]list.Item{}, delegate, 0, 0)
	joinList.Title = "Join A Cambio Game"
	joinList.DisableQuitKeybindings()
	joinList.SetFilteringEnabled(false)
	joinList.SetShowStatusBar(false)
	joinList.SetShowPagination(false)
	joinList.SetShowHelp(false)

	return Model{
		state:            StateInitial,
		selectedCard:     -1,
		selectedOpponent: -1,
		username:         username,
		clientID:         clientID,
		joinList:         joinList,
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
		m.joinList.SetSize(m.width, m.height)
		return m, nil
	case lobbyTickMsg:
		return m, m.handleLobbyTick()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.leaveLobby()
			m.state = StateQuit
			return m, tea.Quit
		case key.Matches(msg, m.keymap.escape):
			if m.state == StateJoinGameList || m.state == StateWaitingForJoin {
				m.resetSelectionState()
				m.state = StateInitial
				return m, nil
			}
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
				m.message = "Power card already committed. Select a card to continue."
				return m, nil
			}
			if m.state == StateLookingAtOpponentCard {
				if m.isPeeking() {
					m.finishOpponentCardPeek()
					return m, nil
				}
				m.message = "Power card already committed. Select an opponent card to continue."
				return m, nil
			}
			m.leaveLobby()
			if m.onBack != nil {
				return m, m.onBack
			}
			return m, nil
		case m.state == StateInitial:
			prevState := m.state
			m.handleInitialKey(msg)
			if prevState != m.state && (m.state == StateJoinGameList || m.state == StateWaitingForJoin) {
				return m, tickLobby()
			}
			return m, nil
		case m.state == StateJoinGameList:
			prevState := m.state
			m.handleJoinListKey(msg)
			if prevState != m.state && m.state == StatePlaying && m.lobbyID != "" {
				return m, tickLobby()
			}
			return m, nil
		case m.state == StateWaitingForJoin:
			prevState := m.state
			m.handleWaitingForJoinKey(msg)
			if prevState != m.state && m.state == StatePlaying && m.lobbyID != "" {
				return m, tickLobby()
			}
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
	case StateJoinGameList:
		body.WriteString(m.joinList.View())
	case StateWaitingForJoin:
		body.WriteString(m.renderWaitingForJoin())
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
	return titleStyle.Render("CAMBIO")
}

func (m *Model) resetSelectionState() {
	m.selectedCard = -1
	m.selectedOpponent = -1
	m.message = ""
}

func (m *Model) handleInitialKey(msg tea.KeyMsg) {
	if key.Matches(msg, m.keymap.createGame) {
		session, playerID, err := sharedLobbies.createLobby(m.username, m.clientID)
		if err != nil {
			m.message = err.Error()
			return
		}
		m.lobbyID = session.id
		m.game = session.game
		m.playerID = playerID
		m.state = StateWaitingForJoin
		m.message = "Game created. Waiting for someone to join..."
		return
	}

	if key.Matches(msg, m.keymap.join) {
		m.refreshJoinList()
		m.state = StateJoinGameList
	}
}

func (m *Model) handleJoinListKey(msg tea.KeyMsg) {
	if key.Matches(msg, m.keymap.start) {
		selected, ok := m.joinList.SelectedItem().(lobbyListItem)
		if !ok {
			m.message = "No game selected"
			return
		}

		session, playerID, err := sharedLobbies.joinLobby(selected.id, m.username, m.clientID)
		if err != nil {
			m.message = err.Error()
			m.refreshJoinList()
			return
		}

		m.lobbyID = session.id
		m.game = session.game
		m.playerID = playerID
		m.state = StatePlaying
		m.message = "Joined game. Press ENTER when you're ready to start."
		return
	}

	var cmd tea.Cmd
	m.joinList, cmd = m.joinList.Update(msg)
	_ = cmd
}

func (m *Model) handleWaitingForJoinKey(msg tea.KeyMsg) {
	if key.Matches(msg, m.keymap.start) {
		if sharedLobbies.hasGuest(m.lobbyID) {
			m.state = StatePlaying
			m.message = "Player joined. Press ENTER when you're ready to start."
			return
		}
	}
}

func (m *Model) handleLobbyTick() tea.Cmd {
	if m.lobbyID == "" {
		return nil
	}

	switch m.state {
	case StateJoinGameList:
		m.refreshJoinList()
		return tickLobby()
	case StateWaitingForJoin:
		if sharedLobbies.hasGuest(m.lobbyID) {
			m.state = StatePlaying
			m.message = "Player joined. Press ENTER when you're ready to start."
		}
		return tickLobby()
	case StatePlaying, StateReplacingCard, StateLookingAtOwnCard, StateLookingAtOpponentCard:
		if m.game == nil {
			return tickLobby()
		}

		if idx, ok := sharedLobbies.playerIndexForClient(m.lobbyID, m.clientID); ok {
			m.playerID = idx
		}

		if m.game.GetGameState() != internalcambio.GameStart {
			if strings.Contains(m.message, "Ready confirmed. Waiting for the other player") ||
				strings.Contains(m.message, "Both players ready. Game started.") ||
				strings.Contains(m.message, "Press ENTER when you're ready to start") {
				m.message = ""
			}
		}

		return tickLobby()
	default:
		return tickLobby()
	}
}

func tickLobby() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return lobbyTickMsg{} })
}

func (m *Model) refreshJoinList() {
	available := sharedLobbies.listJoinableLobbies(m.clientID)
	items := make([]list.Item, 0, len(available))
	for _, lobby := range available {
		items = append(items, lobbyListItem{id: lobby.id, hostName: lobby.hostName})
	}
	if len(items) == 0 {
		m.message = "No open games right now. Waiting for hosts to create one..."
	} else if m.state == StateJoinGameList {
		m.message = ""
	}
	m.joinList.SetItems(items)
}

func (m *Model) leaveLobby() {
	if m.lobbyID == "" {
		return
	}
	sharedLobbies.leave(m.lobbyID, m.clientID)
	m.lobbyID = ""
	m.game = nil
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

func (m Model) playerName(index int) string {
	names, ok := sharedLobbies.getPlayerNames(m.lobbyID)
	if !ok || index < 0 || index >= len(names) {
		return "Player"
	}
	return names[index]
}
