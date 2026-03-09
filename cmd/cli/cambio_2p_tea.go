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
	start  key.Binding
	quit   key.Binding
	escape key.Binding
}

// CambioGameModel represents the Bubble Tea model for the cambio game
type CambioGameModel struct {
	game    *cambio.Game
	state   CambioGameState
	message string
	width   int
	height  int
	keymap  cambioKeymap
}

// NewCambioGameModel creates a new cambio game model
func NewCambioGameModel() CambioGameModel {
	km := cambioKeymap{
		start: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("enter", "continue"),
		),
		quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}

	return CambioGameModel{
		game:    nil,
		state:   CambioStateInitial,
		message: "",
		keymap:  km,
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

		case m.state == CambioStateInitial && key.Matches(msg, m.keymap.start):
			m.game = cambio.NewGame(2)
			m.state = CambioStatePlaying
		}
	}

	return m, nil
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
	b.WriteString(m.renderCards(1))
	b.WriteString("\n\n")
	b.WriteString("Game initialized! (Full gameplay coming soon...)\n")
	return b.String()
}

func (m CambioGameModel) renderCards(currentPlayer int) string {
	playerCards := cambio.GetAllPlayerHands(m.game)
	gameStart := cambio.GetGameStart(m.game)
	if len(playerCards) == 0 {
		return "No players"
	}

	if currentPlayer < 1 || currentPlayer > len(playerCards) {
		currentPlayer = 1
	}
	currentPlayerIdx := currentPlayer - 1

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
	return lipgloss.JoinVertical(lipgloss.Left, topRow, middleRow, currentPlayerBox)
}

func (m CambioGameModel) renderSharedCards(currentPlayer int) string {
	activeCard := renderCambioCard(cambio.GetActiveCard(m.game, currentPlayer), true)
	deckCard := renderEmptyCambioCard()
	topDiscardCard := renderCambioCard(cambio.GetTopDiscardCard(m.game), true)

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
			m.keymap.start,
			m.keymap.quit,
		}
	case CambioStatePlaying:
		return []key.Binding{
			m.keymap.quit,
		}
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
