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
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6"))
	s.WriteString(titleStyle.Render("BLACKJACK"))
	s.WriteString("\n\n")

	// Hand box style — fixed width, like the bubbletea example
	handStyle := lipgloss.NewStyle().
		Width(30).
		Height(5).
		Align(lipgloss.Left, lipgloss.Top).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("8"))

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	switch m.state {
	case StateInitial:
		s.WriteString("Welcome to Blackjack!\n\n")
		s.WriteString("Press ENTER or SPACE to start a new game\n")
		s.WriteString("Press Q or CTRL+C to quit")

	case StatePlaying:
		dealerCard, _ := m.game.DealerVisibleCard()
		playerCards := m.game.PlayerHand()
		playerScore := m.game.PlayerScore()

		dealerText := fmt.Sprintf("DEALER'S HAND\n%s\nScore: %d",
			formatCards([]cards.Card{dealerCard}),
			dealerCard.Rank)
		s.WriteString(handStyle.Render(dealerText))
		s.WriteString("\n")

		scoreStr := fmt.Sprintf("Score: %d", playerScore)
		if playerScore > 21 {
			scoreStr += " (BUST!)"
		}
		playerText := fmt.Sprintf("YOUR HAND\n%s\n%s",
			formatCards(playerCards),
			scoreStr)
		s.WriteString(handStyle.Render(playerText))
		s.WriteString("\n\n")

		hitStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(map[bool]string{true: "10", false: "7"}[m.selectedIndex == 0]))
		standStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(map[bool]string{true: "10", false: "7"}[m.selectedIndex == 1]))

		s.WriteString(hitStyle.Render(renderOption("Hit", m.selectedIndex == 0)))
		s.WriteString("  ")
		s.WriteString(standStyle.Render(renderOption("Stand", m.selectedIndex == 1)))
		s.WriteString("\n")

	case StateShowingResult:
		dealerCards, _ := m.game.DealerHand()
		playerCards := m.game.PlayerHand()

		dealerText := fmt.Sprintf("DEALER'S HAND\n%s\nScore: %d",
			formatCards(dealerCards),
			m.game.DealerScore())
		s.WriteString(handStyle.Render(dealerText))
		s.WriteString("\n")

		playerText := fmt.Sprintf("YOUR HAND\n%s\nScore: %d",
			formatCards(playerCards),
			m.game.PlayerScore())
		s.WriteString(handStyle.Render(playerText))
		s.WriteString("\n\n")

		winner := m.game.DetermineWinner()
		resultStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(map[int]string{0: "3", 1: "1", 2: "2"}[winner]))
		switch winner {
		case 0:
			s.WriteString(resultStyle.Render("PUSH! It's a tie!"))
		case 1:
			s.WriteString(resultStyle.Render("Dealer wins!"))
		case 2:
			s.WriteString(resultStyle.Render("YOU WIN! Congratulations!"))
		}
		s.WriteString("\n\n")

	case StatePlayAgain:
		s.WriteString("Play another round?\n\n")
		yesStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(map[bool]string{true: "10", false: "7"}[m.selectedIndex == 0]))
		noStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(map[bool]string{true: "10", false: "7"}[m.selectedIndex == 1]))

		s.WriteString(yesStyle.Render(renderOption("Yes", m.selectedIndex == 0)))
		s.WriteString("  ")
		s.WriteString(noStyle.Render(renderOption("No", m.selectedIndex == 1)))
		s.WriteString("\n")
	}

	if m.message != "" {
		s.WriteString("\n")
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		s.WriteString(errorStyle.Render("Error: " + m.message))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render(m.helpView()))

	view := tea.NewView(s.String())
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

func formatCards(hand []cards.Card) string {
	if len(hand) == 0 {
		return "No cards"
	}
	strs := make([]string, len(hand))
	for i, c := range hand {
		strs[i] = c.String()
	}
	return strings.Join(strs, " ")
}

func renderOption(text string, selected bool) string {
	if selected {
		return fmt.Sprintf("> %s <", text)
	}
	return fmt.Sprintf("  %s  ", text)
}
