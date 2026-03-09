package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/cambio"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

var cambioCardStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("8")).
	Width(5).
	Align(lipgloss.Center)

var cambioRedCardStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
var cambioBlackCardStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

// CambioGameState represents the current state of the game UI
type CambioGameState int

const (
	CambioStateInitial CambioGameState = iota
	CambioStatePlaying
	CambioStateShowingResult
	CambioStatePlayAgain
	CambioStateQuit
)

type cambioKeymap struct {
	join           key.Binding
	start          key.Binding
	quit           key.Binding
	escape         key.Binding
	draw           key.Binding
	replace        key.Binding
	discard        key.Binding
	lookAtSelf     key.Binding
	lookAtOpponent key.Binding
	swap           key.Binding
	cambio         key.Binding
}

// CambioGameModel represents the Bubble Tea model for the cambio game
type CambioGameModel struct {
	game     *cambio.Game
	state    CambioGameState
	playerId int
	message  string
	width    int
	height   int
	keymap   cambioKeymap
}

// NewCambioGameModel creates a new cambio game model
func NewCambioGameModel() CambioGameModel {
	km := cambioKeymap{
		join: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("enter", "join game"),
		),
		start: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "start game"),
		),
		quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		draw: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "draw card"),
		),
		replace: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "replace card 0"),
		),
		discard: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "discard card"),
		),
		lookAtSelf: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "look at your cards"),
		),
		lookAtOpponent: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "look at opponent's cards"),
		),
		swap: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "swap a card with opponent"),
		),
		cambio: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "call cambio"),
		),
	}

	return CambioGameModel{
		game:     nil,
		state:    CambioStateInitial,
		playerId: 0,
		message:  "",
		keymap:   km,
	}
}

// Init initializes the model
func (m CambioGameModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m CambioGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.escape):
			return m, navigate(screenHome)
		case key.Matches(msg, m.keymap.quit):
			m.state = CambioStateQuit
			return m, tea.Quit

		case m.state == CambioStateInitial && key.Matches(msg, m.keymap.join):
			m.game = cambio.NewGame(2)
			m.state = CambioStatePlaying
		case m.state == CambioStatePlaying:
			m.handleGameplayKey(msg)
		}
	}

	return m, nil
}

func (m *CambioGameModel) handleGameplayKey(msg tea.KeyMsg) {
	if m.game == nil {
		return
	}

	if key.Matches(msg, m.keymap.start) {
		if m.game.GetGameState() == cambio.GameStart {
			m.game.StartGame()
			m.message = ""
		}
		return
	}

	if m.playerId != m.game.GetPlayerTurn() {
		return
	}

	if key.Matches(msg, m.keymap.draw) {
		if m.game.GetGameState() == cambio.WaitingForTurn {
			m.game.DrawCard()
			m.message = ""
		}
		return
	}

	if key.Matches(msg, m.keymap.discard) {
		m.game.SelectTurnType(cambio.Discard)
		if err := m.game.DiscardCard(); err != nil {
			m.message = err.Error()
			return
		}
		m.message = ""
		return
	}

	if key.Matches(msg, m.keymap.replace) {
		m.game.SelectTurnType(cambio.Replace)
		if err := m.game.ReplaceCard(0); err != nil {
			m.message = err.Error()
			return
		}
		m.message = ""
	}
}

// View renders the model
func (m CambioGameModel) View() string {
	var body strings.Builder

	body.WriteString(m.renderTitle())
	body.WriteString("\n")

	switch m.state {
	case CambioStateInitial:
		body.WriteString(m.renderInitial())
	case CambioStatePlaying:
		body.WriteString(m.renderPlaying())
	case CambioStateShowingResult:
		body.WriteString(m.renderResult())
	case CambioStatePlayAgain:
		body.WriteString(m.renderPlayAgain())
	}

	if m.message != "" {
		body.WriteString("\n")
		body.WriteString(errorStyle.Render("Error: " + m.message))
		body.WriteString("\n")
	}

	helpStr := renderKeyHelp(m.width, m.getActiveKeybindings())
	content := renderWithFooter(body.String(), helpStr, m.width, m.height)
	return renderApp(content)
}

func (m CambioGameModel) renderTitle() string { return titleStyle.Render("CAMBIO - 2 PLAYER") }

func (m CambioGameModel) renderInitial() string {
	var b strings.Builder
	b.WriteString("Welcome to Cambio!\n\n")
	b.WriteString("Press ENTER or SPACE to start a new game\n")
	b.WriteString("Press Q or CTRL+C to quit\n\n")
	b.WriteString("How to play:\n")
	b.WriteString("- Goal: Have the lowest total card value at the end of the game.\n")
	b.WriteString("- Each player starts with 4 face-down cards.\n")
	b.WriteString("- On your turn, draw a card and choose to replace or discard it.\n")
	b.WriteString("- Special cards let you peek at cards, swap with opponent, etc.\n")
	b.WriteString("- Call 'Cambio' when you think you have the lowest score!\n")
	return b.String()
}

func (m CambioGameModel) renderPlaying() string {
	if m.game == nil {
		return "Game not initialized"
	}

	var b strings.Builder
	b.WriteString(m.renderCards(m.playerId))
	b.WriteString("\n\n")
	b.WriteString("Start with ENTER, draw with D, discard with X, or replace card index 0 with R.\n")
	return b.String()
}

func (m CambioGameModel) renderCards(currentPlayer int) string {
	playerCards := m.game.GetAllPlayerHands()
	gameStart := m.game.GetGameStart()
	if len(playerCards) == 0 {
		return "No players"
	}

	if currentPlayer < 0 || currentPlayer >= len(playerCards) {
		currentPlayer = 0
	}
	currentPlayerIdx := currentPlayer

	topRowHands := make([]string, 0, len(playerCards)-1)
	relativePlayerNum := 2
	for offset := 1; offset < len(playerCards); offset++ {
		idx := (currentPlayerIdx + offset) % len(playerCards)
		title := fmt.Sprintf("PLAYER %d'S HAND", relativePlayerNum)
		topRowHands = append(topRowHands, renderPlayerHandBox(title, playerCards[idx], false, false))
		relativePlayerNum++
	}

	currentPlayerBox := renderPlayerHandBox("YOUR HAND", playerCards[currentPlayerIdx], gameStart, true)

	if len(topRowHands) == 0 {
		return currentPlayerBox
	}

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, topRowHands...)
	middleRow := m.renderSharedCards(currentPlayer)
	return lipgloss.JoinVertical(lipgloss.Center, topRow, middleRow, currentPlayerBox)
}

func (m CambioGameModel) renderSharedCards(currentPlayer int) string {
	activeCard := renderCambioCard(m.game.GetActiveCard(currentPlayer), true)
	deckCard := renderEmptyCambioCard()
	topDiscardCard := renderCambioCard(m.game.GetTopDiscardCard(), true)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().PaddingRight(1).Render(lipgloss.JoinVertical(lipgloss.Center, "Deck", deckCard)),
		lipgloss.NewStyle().PaddingRight(1).Render(lipgloss.JoinVertical(lipgloss.Center, "Discard", topDiscardCard)),
		lipgloss.NewStyle().PaddingRight(1).Render(lipgloss.JoinVertical(lipgloss.Center, "Active", activeCard)),
	)

}

func renderPlayerHandBox(title string, playerCards []cards.Card, gameStart bool, faceUp bool) string {
	handDisplay := "No cards"
	if len(playerCards) > 0 {
		columns := make([]string, 0, (len(playerCards)+1)/2)
		for topIdx := 0; topIdx < len(playerCards); topIdx += 2 {
			displayFaceUp := faceUp || gameStart
			topCard := renderCambioCard(playerCards[topIdx], false)

			bottomCard := renderEmptyCambioCard()
			bottomIdx := topIdx + 1
			if bottomIdx < len(playerCards) {
				bottomCard = renderCambioCard(playerCards[bottomIdx], displayFaceUp)
			}

			column := lipgloss.NewStyle().PaddingRight(1).Render(
				lipgloss.JoinVertical(lipgloss.Left, topCard, bottomCard),
			)
			columns = append(columns, column)
		}
		handDisplay = lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	}

	return boxStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		handDisplay,
	))
}

func renderCambioCard(card cards.Card, faceUp bool) string {
	face := card.RenderFace()
	if card.IsRed() {
		face = cambioRedCardStyle.Render(face)
	} else {
		face = cambioBlackCardStyle.Render(face)
	}
	if !faceUp {
		return renderEmptyCambioCard()
	}
	return cambioCardStyle.Render(face)
}

func renderEmptyCambioCard() string {
	return cambioCardStyle.Render(" \n     \n")
}

func (m CambioGameModel) renderResult() string {
	var b strings.Builder
	b.WriteString("Game Over!\n\n")
	b.WriteString("(Results display coming soon...)\n")
	return b.String()
}

func (m CambioGameModel) renderPlayAgain() string {
	var b strings.Builder
	b.WriteString("Play another round?\n\n")
	b.WriteString("(Play again option coming soon...)\n")
	return b.String()
}

// getActiveKeybindings returns the appropriate keybindings for the current state
func (m CambioGameModel) getActiveKeybindings() []key.Binding {
	switch m.state {
	case CambioStateInitial:
		return []key.Binding{
			m.keymap.join,
			m.keymap.quit,
		}
	case CambioStatePlaying:
		return append([]key.Binding{
			m.keymap.quit}, m.getGameKeybindings()...)
	case CambioStateShowingResult:
		return []key.Binding{
			m.keymap.start,
		}
	case CambioStatePlayAgain:
		return []key.Binding{
			m.keymap.start,
			m.keymap.quit,
		}
	default:
		return []key.Binding{m.keymap.quit}
	}
}

func (m CambioGameModel) getGameKeybindings() []key.Binding {
	gameState := m.game.GetGameState()
	if m.playerId != m.game.GetPlayerTurn() {
		// if it's not the player's turn, only allow quit
		return []key.Binding{m.keymap.quit}
	}
	switch gameState {
	case cambio.GameStart:
		return []key.Binding{
			m.keymap.start,
		}
	case cambio.WaitingForTurn:
		return []key.Binding{
			m.keymap.draw,
		}
	case cambio.TakingTurn:
		var bindings []key.Binding
		validTurnType := m.game.GetValidTurnTypes()
		for _, turnType := range validTurnType {
			switch turnType {
			case cambio.Discard:
				bindings = append(bindings, m.keymap.discard)
			case cambio.Replace:
				bindings = append(bindings, m.keymap.replace)
			case cambio.LookAtSelf:
				bindings = append(bindings, m.keymap.lookAtSelf)
			case cambio.LookAtOpponent:
				bindings = append(bindings, m.keymap.lookAtOpponent)
			case cambio.Swap:
				bindings = append(bindings, m.keymap.swap)
			case cambio.CallGame:
				bindings = append(bindings, m.keymap.cambio)
			}
		}
		return append(bindings, m.keymap.quit)
	default:
		return []key.Binding{}
	}
}
