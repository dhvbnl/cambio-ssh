package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/blackjack"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

var (
	boxStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 1)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	actionOn    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	actionOff   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	winnerColor = map[int]string{0: "3", 1: "1", 2: "2"}
)

// GameState represents the current state of the game UI
type GameState int

const (
	StateInitial GameState = iota
	StatePlaying
	StateShowingResult
	StatePlayAgain
	StateQuit
)

type keymap struct {
	hit    key.Binding
	stand  key.Binding
	yes    key.Binding
	no     key.Binding
	start  key.Binding
	quit   key.Binding
	escape key.Binding
}

// GameModel represents the Bubble Tea model for the blackjack game
type GameModel struct {
	game          blackjack.Game
	state         GameState
	selectedIndex int // 0 = Hit, 1 = Stand (when playing)
	message       string
	width         int
	height        int
	keymap        keymap
}

// NewGameModel creates a new game model
func NewGameModel() GameModel {
	km := keymap{
		hit: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "hit"),
		),
		stand: key.NewBinding(
			key.WithKeys("s", "right"),
			key.WithHelp("s/→", "stand"),
		),
		yes: key.NewBinding(
			key.WithKeys("y", "left"),
			key.WithHelp("y/←", "yes"),
		),
		no: key.NewBinding(
			key.WithKeys("n", "right"),
			key.WithHelp("n/→", "no"),
		),
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

	return GameModel{
		game:          blackjack.NewGame(),
		state:         StateInitial,
		selectedIndex: 0,
		message:       "",
		keymap:        km,
	}
}

// Init initializes the model
func (m GameModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.state = StateQuit
			return m, tea.Quit

		case m.state == StateInitial && key.Matches(msg, m.keymap.start):
			if err := m.game.InitialDeal(); err != nil {
				m.message = fmt.Sprintf("Error dealing: %v", err)
				return m, nil
			}
			m.state = StatePlaying
			m.selectedIndex = 0

		case m.state == StatePlaying && key.Matches(msg, m.keymap.hit):
			m.selectedIndex = 0
		case m.state == StatePlaying && key.Matches(msg, m.keymap.stand):
			m.selectedIndex = 1
		case m.state == StatePlaying && key.Matches(msg, m.keymap.start):
			if m.selectedIndex == 0 {
				// Hit
				_, err := m.game.PlayerHit()
				if err != nil {
					m.message = fmt.Sprintf("Error on hit: %v", err)
					return m, nil
				}
				if m.game.PlayerScore() > 21 {
					m.state = StateShowingResult
				}
			} else {
				// Stand
				if err := m.game.PlayerStand(); err != nil {
					m.message = fmt.Sprintf("Error on stand: %v", err)
					return m, nil
				}
				if err := m.game.DealerPlay(); err != nil {
					m.message = fmt.Sprintf("Error during dealer play: %v", err)
					return m, nil
				}
				m.state = StateShowingResult
			}

		case m.state == StateShowingResult && key.Matches(msg, m.keymap.start):
			m.state = StatePlayAgain
			m.selectedIndex = 0

		case m.state == StatePlayAgain && key.Matches(msg, m.keymap.yes):
			m.selectedIndex = 0
		case m.state == StatePlayAgain && key.Matches(msg, m.keymap.no):
			m.selectedIndex = 1
		case m.state == StatePlayAgain && key.Matches(msg, m.keymap.start):
			if m.selectedIndex == 0 {
				m.game = blackjack.NewGame()
				if err := m.game.InitialDeal(); err != nil {
					m.message = fmt.Sprintf("Error dealing: %v", err)
					return m, nil
				}
				m.state = StatePlaying
				m.selectedIndex = 0
			} else {
				m.state = StateQuit
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View renders the model
func (m GameModel) View() string {
	var body strings.Builder

	body.WriteString(m.renderTitle())
	body.WriteString("\n")

	switch m.state {
	case StateInitial:
		body.WriteString(m.renderInitial())
	case StatePlaying:
		body.WriteString(m.renderPlaying())
	case StateShowingResult:
		body.WriteString(m.renderResult())
	case StatePlayAgain:
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

func (m GameModel) renderTitle() string { return titleStyle.Render("BLACKJACK") }

func (m GameModel) renderInitial() string {
	var b strings.Builder
	b.WriteString("Welcome to Blackjack!\n\n")
	b.WriteString("Press ENTER or SPACE to start a new game\n")
	b.WriteString("Press Q or CTRL+C to quit\n\n")
	b.WriteString("How to play:\n")
	b.WriteString("- Goal: finish with a hand value closer to 21 than the dealer without busting.\n")
	b.WriteString("- Card values: 2-10 face value, J/Q/K = 10, Ace = 1 or 11 (best for you).\n")
	b.WriteString("- Your turn: hit to take a card or stand to hold. Busting (over 21) loses immediately.\n")
	b.WriteString("- Dealer: reveals hand after you stand, then hits until at least 17.\n")
	b.WriteString("- Outcomes: higher total wins; tie is a push; busting always loses.\n")
	return b.String()
}

func (m GameModel) renderPlaying() string {
	dealerCard, _ := m.game.DealerVisibleCard()
	playerCards := m.game.PlayerHand()
	playerScore := m.game.PlayerScore()

	dealerDisplay := formatCards([]cards.Card{dealerCard})
	dealerScoreStr := fmt.Sprintf("Score: %d", dealerCard.Rank)
	dealerBox := boxStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"DEALER'S HAND",
		dealerDisplay,
		dealerScoreStr,
	))

	playerDisplay := formatCards(playerCards)
	scoreStr := fmt.Sprintf("Score: %d", playerScore)
	if playerScore > 21 {
		scoreStr += " (BUST!)"
	}
	playerBox := boxStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"YOUR HAND",
		playerDisplay,
		scoreStr,
	))

	var b strings.Builder
	b.WriteString(dealerBox)
	b.WriteString("\n\n")
	b.WriteString(playerBox)
	b.WriteString("\n\n")
	b.WriteString("Choose your action:\n")
	b.WriteString(actionStyle(m.selectedIndex == 0).Render(renderOption("Hit", m.selectedIndex == 0)))
	b.WriteString("  ")
	b.WriteString(actionStyle(m.selectedIndex == 1).Render(renderOption("Stand", m.selectedIndex == 1)))
	b.WriteString("\n")
	return b.String()
}

func (m GameModel) renderResult() string {
	dealerCards, _ := m.game.DealerHand()
	playerCards := m.game.PlayerHand()

	dealerDisplay := formatCards(dealerCards)
	dealerScoreStr := fmt.Sprintf("Score: %d", m.game.DealerScore())
	dealerBox := boxStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"DEALER'S HAND",
		dealerDisplay,
		dealerScoreStr,
	))

	playerDisplay := formatCards(playerCards)
	playerScoreStr := fmt.Sprintf("Score: %d", m.game.PlayerScore())
	playerBox := boxStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"YOUR HAND",
		playerDisplay,
		playerScoreStr,
	))

	winner := m.game.DetermineWinner()
	resultStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(winnerColor[winner]))

	var b strings.Builder
	b.WriteString(dealerBox)
	b.WriteString("\n\n")
	b.WriteString(playerBox)
	b.WriteString("\n\n")
	switch winner {
	case 0:
		b.WriteString(resultStyle.Render("PUSH! It's a tie!"))
	case 1:
		b.WriteString(resultStyle.Render("Dealer wins!"))
	case 2:
		b.WriteString(resultStyle.Render("YOU WIN! Congratulations!"))
	}
	b.WriteString("\n\n")
	return b.String()
}

func (m GameModel) renderPlayAgain() string {
	var b strings.Builder
	b.WriteString("Play another round?\n\n")
	b.WriteString(actionStyle(m.selectedIndex == 0).Render(renderOption("Yes", m.selectedIndex == 0)))
	b.WriteString("  ")
	b.WriteString(actionStyle(m.selectedIndex == 1).Render(renderOption("No", m.selectedIndex == 1)))
	b.WriteString("\n")
	return b.String()
}

func actionStyle(active bool) lipgloss.Style {
	if active {
		return actionOn
	}
	return actionOff
}

// getActiveKeybindings returns the appropriate keybindings for the current state
func (m GameModel) getActiveKeybindings() []key.Binding {
	switch m.state {
	case StateInitial:
		return []key.Binding{
			m.keymap.start,
			m.keymap.quit,
		}
	case StatePlaying:
		return []key.Binding{
			m.keymap.hit,
			m.keymap.stand,
			m.keymap.start,
			m.keymap.quit,
		}
	case StateShowingResult:
		return []key.Binding{
			m.keymap.start,
		}
	case StatePlayAgain:
		return []key.Binding{
			m.keymap.yes,
			m.keymap.no,
			m.keymap.start,
			m.keymap.quit,
		}
	default:
		return []key.Binding{m.keymap.quit}
	}
}

// Helper functions

func formatCards(cards []cards.Card) string {
	if len(cards) == 0 {
		return "No cards"
	}
	strs := make([]string, len(cards))
	for i, c := range cards {
		strs[i] = c.String()
	}
	return strings.Join(strs, " ")
}

func renderOption(text string, selected bool) string {
	if selected {
		return fmt.Sprintf("►%s◄", text)
	}
	return fmt.Sprintf(" %s ", text)
}
