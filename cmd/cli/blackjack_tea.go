package cli

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/blackjack"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

// PlayBlackjackWithTeaUI starts a blackjack game with Bubble Tea UI
func PlayBlackjackWithTeaUI() {
	p := tea.NewProgram(NewGameModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

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
	hit   key.Binding
	stand key.Binding
	yes   key.Binding
	no    key.Binding
	up    key.Binding
	down  key.Binding
	left  key.Binding
	right key.Binding
	start key.Binding
	quit  key.Binding
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
	help          help.Model
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
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("q", "quit"),
		),
	}

	return GameModel{
		game:          blackjack.NewGame(),
		state:         StateInitial,
		selectedIndex: 0,
		message:       "",
		keymap:        km,
		help:          help.New(),
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
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch {
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
func (m GameModel) View() tea.View {
	var content strings.Builder

	// Title with styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		MarginBottom(1)
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	title := titleStyle.Render("BLACKJACK")
	content.WriteString(title)
	content.WriteString("\n")

	switch m.state {
	case StateInitial:
		content.WriteString("Welcome to Blackjack!\n\n")
		content.WriteString("Press ENTER or SPACE to start a new game\n")
		content.WriteString("Press Q or CTRL+C to quit")

	case StatePlaying:
		dealerCard, _ := m.game.DealerVisibleCard()
		playerCards := m.game.PlayerHand()
		playerScore := m.game.PlayerScore()

		content.WriteString("┌─────────────────────────────────────────────────┐\n")
		content.WriteString("│ DEALER'S HAND                                   │\n")
		content.WriteString("├─────────────────────────────────────────────────┤\n")
		dealerDisplay := formatCards([]cards.Card{dealerCard})
		content.WriteString(fmt.Sprintf("│ %s\n", padString(dealerDisplay, 47)))
		content.WriteString(fmt.Sprintf("│ Score: %s\n", padString(fmt.Sprintf("%d", dealerCard.Rank), 29)))
		content.WriteString("└─────────────────────────────────────────────────┘\n\n")

		content.WriteString("┌─────────────────────────────────────────────────┐\n")
		content.WriteString("│ " + userStyle.Render("YOUR HAND") + "                                       │\n")
		content.WriteString("├─────────────────────────────────────────────────┤\n")
		playerDisplay := formatCards(playerCards)
		content.WriteString(fmt.Sprintf("│ %s\n", padString(playerDisplay, 47)))

		scoreStr := fmt.Sprintf("Score: %d", playerScore)
		if playerScore > 21 {
			scoreStr += " (BUST!)"
		}
		content.WriteString(fmt.Sprintf("│ %s\n", padString(scoreStr, 47)))
		content.WriteString("└─────────────────────────────────────────────────┘\n\n")

		content.WriteString("Choose your action:\n")
		hitStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(map[bool]string{true: "10", false: "7"}[m.selectedIndex == 0]))
		standStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(map[bool]string{true: "10", false: "7"}[m.selectedIndex == 1]))

		content.WriteString(hitStyle.Render(renderOption("Hit", m.selectedIndex == 0)))
		content.WriteString("  ")
		content.WriteString(standStyle.Render(renderOption("Stand", m.selectedIndex == 1)))
		content.WriteString("\n")

	case StateShowingResult:
		dealerCards, _ := m.game.DealerHand()
		playerCards := m.game.PlayerHand()

		content.WriteString("┌─────────────────────────────────────────────────┐\n")
		content.WriteString("│ DEALER'S HAND                                   │\n")
		content.WriteString("├─────────────────────────────────────────────────┤\n")
		dealerDisplay := formatCards(dealerCards)
		content.WriteString(fmt.Sprintf("│ %s\n", padString(dealerDisplay, 38)))
		content.WriteString(fmt.Sprintf("│ Score: %s\n", padString(fmt.Sprintf("%d", m.game.DealerScore()), 40)))
		content.WriteString("└─────────────────────────────────────────────────┘\n\n")

		content.WriteString("┌─────────────────────────────────────────────────┐\n")
		content.WriteString("│ " + userStyle.Render("YOUR HAND") + "                                       │\n")
		content.WriteString("├─────────────────────────────────────────────────┤\n")
		playerDisplay := formatCards(playerCards)
		content.WriteString(fmt.Sprintf("│ %s\n", padString(playerDisplay, 47)))
		scoreStr := fmt.Sprintf("Score: %d", m.game.PlayerScore())
		content.WriteString(fmt.Sprintf("│ %s\n", padString(scoreStr, 47)))
		content.WriteString("└─────────────────────────────────────────────────┘\n\n")

		// Determine winner
		winner := m.game.DetermineWinner()
		resultStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(map[int]string{0: "3", 1: "1", 2: "2"}[winner]))
		switch winner {
		case 0:
			content.WriteString(resultStyle.Render("PUSH! It's a tie!"))
		case 1:
			content.WriteString(resultStyle.Render("Dealer wins!"))
		case 2:
			content.WriteString(resultStyle.Render("YOU WIN! Congratulations!"))
		}

		content.WriteString("\n\n")

	case StatePlayAgain:
		content.WriteString("Play another round?\n\n")
		yesStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(map[bool]string{true: "10", false: "7"}[m.selectedIndex == 0]))
		noStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(map[bool]string{true: "10", false: "7"}[m.selectedIndex == 1]))

		content.WriteString(yesStyle.Render(renderOption("Yes", m.selectedIndex == 0)))
		content.WriteString("  ")
		content.WriteString(noStyle.Render(renderOption("No", m.selectedIndex == 1)))
		content.WriteString("\n")
	}

	if m.message != "" {
		content.WriteString("\n")
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		content.WriteString(errorStyle.Render("Error: " + m.message))
		content.WriteString("\n")
	}

	// Add help view
	content.WriteString("\n")
	content.WriteString(m.helpView())

	view := tea.NewView(content.String())
	view.AltScreen = true

	return view
}

// helpView returns the help text with keybindings
func (m GameModel) helpView() string {
	return "\n" + m.help.ShortHelpView(m.getActiveKeybindings())
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

func padString(text string, width int) string {
	if len(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-len(text))
}

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
