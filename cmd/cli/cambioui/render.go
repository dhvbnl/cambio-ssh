package cambioui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	internalcambio "github.com/dhvbnl/cambio-ssh/cmd/internal/cambio"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

func (m Model) renderInitial() string {
	var b strings.Builder
	b.WriteString("Welcome to Cambio!\n\n")
	b.WriteString("Create a new game or join an open game.\n\n")
	b.WriteString("- Press c to create your own game\n")
	b.WriteString("- Press j to view joinable games\n")
	b.WriteString("- Press esc to go back\n")
	return b.String()
}

func (m Model) renderWaitingForJoin() string {
	name := m.username
	if name == "" {
		name = "Your"
	}
	return fmt.Sprintf("%s's game\n\nWaiting for someone to join...", name)
}

func (m Model) renderPlaying() string {
	if m.game == nil {
		return "Game not initialized"
	}
	return m.renderCards(m.playerID) + "\n\n"
}

func (m Model) renderCards(currentPlayer int) string {
	playerHands := m.game.GetAllPlayerHands()
	if len(playerHands) == 0 {
		return "No players"
	}

	peeking := m.isPeeking()

	if currentPlayer < 0 || currentPlayer >= len(playerHands) {
		currentPlayer = 0
	}

	revealedPlayer, revealedIndex, _, _, revealedActive := m.game.GetRevealedCard(currentPlayer)

	topRowHands := make([]string, 0, len(playerHands)-1)
	for offset := 1; offset < len(playerHands); offset++ {
		idx := (currentPlayer + offset) % len(playerHands)
		playerName := m.playerName(idx)
		title := fmt.Sprintf("%s's Hand", playerName)
		highlightedIndex := -1
		if revealedActive && revealedPlayer == idx {
			highlightedIndex = revealedIndex
		}
		targetedPlayer := m.state == StateLookingAtOpponentCard && m.selectedOpponent == idx
		targetedSelectedCard := -1
		if targetedPlayer {
			targetedSelectedCard = m.selectedCard
		}
		topRowHands = append(topRowHands, renderPlayerHandBox(title, playerHands[idx], false, false, true, m.state, peeking, targetedPlayer, targetedSelectedCard, highlightedIndex))
	}

	yourHighlightedIndex := -1
	if revealedActive && revealedPlayer == currentPlayer {
		yourHighlightedIndex = revealedIndex
	}

	yourHand := renderPlayerHandBox(
		"Your Hand",
		playerHands[currentPlayer],
		m.game.GetGameStart(),
		false,
		false,
		m.state,
		peeking,
		false,
		m.selectedCard,
		yourHighlightedIndex,
	)
	if len(topRowHands) == 0 {
		return yourHand
	}

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, topRowHands...)
	middle := m.renderSharedCards(currentPlayer)
	return lipgloss.JoinVertical(lipgloss.Center, topRow, middle, yourHand)
}

func (m Model) renderSharedCards(currentPlayer int) string {
	activeCard := m.game.GetActiveCard(currentPlayer)
	active := renderEmptyCard()
	if activeCard != (cards.Card{}) {
		active = renderCard(activeCard, true, -1, cardStyle)
	}

	deck := renderCardBack(cardStyle)
	discard := renderCard(m.game.GetTopDiscardCard(), true, -1, cardStyle)
	revealed := renderEmptyCard()
	_, _, revealedCard, canSeeFace, revealedActive := m.game.GetRevealedCard(currentPlayer)
	_, _, _, isPublicReveal, _ := m.game.GetRevealedCard(-1)
	if revealedActive && !isPublicReveal {
		if canSeeFace {
			revealed = renderCard(revealedCard, true, -1, cardStyle)
		} else {
			revealed = renderCardBack(cardStyle)
		}
	}

	stack := lipgloss.NewStyle().Padding(1, 1, 0, 1)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		stack.Render(lipgloss.JoinVertical(lipgloss.Center, "Deck", deck)),
		stack.Render(lipgloss.JoinVertical(lipgloss.Center, "Discard", discard)),
		stack.Render(lipgloss.JoinVertical(lipgloss.Center, "Active", active)),
		stack.Render(lipgloss.JoinVertical(lipgloss.Center, "Peek", revealed)),
	)
}

func renderPlayerHandBox(title string, hand []cards.Card, gameStart bool, faceUp bool, invertVertical bool, mode State, peekActive bool, targetedPlayer bool, selectedIndex int, revealedIndex int) string {
	handDisplay := "No cards"
	if len(hand) > 0 {
		columns := make([]string, 0, (len(hand)+1)/2)
		for topIdx := 0; topIdx < len(hand); topIdx += 2 {
			showFace := faceUp || gameStart
			topCard := renderEmptyCard()
			bottomCard := renderEmptyCard()

			renderAtTop := func(cardIdx int) {
				style := selectableCardStyle(mode, peekActive, targetedPlayer, cardIdx, selectedIndex, revealedIndex, hand[cardIdx])
				topCard = renderCard(hand[cardIdx], false, cardIdx+1, style)
			}

			renderAtBottom := func(cardIdx int) {
				style := selectableCardStyle(mode, peekActive, targetedPlayer, cardIdx, selectedIndex, revealedIndex, hand[cardIdx])
				bottomCard = renderCard(hand[cardIdx], showFace, cardIdx+1, style)
			}

			bottomIdx := topIdx + 1
			if !invertVertical {
				renderAtTop(topIdx)
				if bottomIdx < len(hand) {
					renderAtBottom(bottomIdx)
				}
			} else {
				renderAtBottom(topIdx)
				if bottomIdx < len(hand) {
					renderAtTop(bottomIdx)
				}
			}

			column := lipgloss.NewStyle().PaddingRight(1).Render(lipgloss.JoinVertical(lipgloss.Left, topCard, bottomCard))
			columns = append(columns, column)
		}
		handDisplay = lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	}

	containerStyle := boxStyle
	if targetedPlayer {
		containerStyle = selectedBoxStyle
	}

	return containerStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, handDisplay))
}

func selectableCardStyle(mode State, peekActive bool, targetedPlayer bool, idx int, selectedIndex int, revealedIndex int, card cards.Card) lipgloss.Style {
	if card == (cards.Card{}) {
		return cardStyle
	}

	if idx == revealedIndex {
		return selectedLookCardStyle
	}

	switch mode {
	case StateReplacingCard:
		if idx == selectedIndex {
			return selectedReplaceCardStyle
		}
		return candidateCardStyle
	case StateLookingAtOwnCard:
		if peekActive {
			return cardStyle
		}
		if idx == selectedIndex {
			return selectedLookCardStyle
		}
		return candidateCardStyle
	case StateLookingAtOpponentCard:
		if !targetedPlayer {
			return cardStyle
		}
		if peekActive {
			return cardStyle
		}
		if idx == selectedIndex {
			return selectedLookCardStyle
		}
		return candidateCardStyle
	default:
		return cardStyle
	}
}

func renderCard(card cards.Card, faceUp bool, index int, style lipgloss.Style) string {
	if !faceUp {
		return style.Render(fmt.Sprintf(" %d \n     \n", index))
	}

	face := card.RenderFace()
	if card.IsRed() {
		face = redCardStyle.Render(face)
	} else {
		face = blackCardStyle.Render(face)
	}
	return style.Render(face)
}

func renderCardBack(style lipgloss.Style) string {
	return style.Render(cardBackStyle.Render("///\nXXX\n///"))
}

func renderEmptyCard() string {
	return cardStyle.Render(" \n     \n")
}

func (m Model) renderResult() string {
	return "Game Over!\n\n(Results display coming soon...)\n"
}

func (m Model) renderPlayAgain() string {
	return "Play another round?\n\n(Play again option coming soon...)\n"
}

func (m Model) getActiveKeybindings() []key.Binding {
	peeking := m.isPeeking()

	switch m.state {
	case StateInitial:
		return []key.Binding{m.keymap.createGame, m.keymap.join, m.keymap.escape, m.keymap.quit}
	case StateJoinGameList:
		return []key.Binding{m.keymap.escape, m.keymap.start, m.keymap.quit}
	case StateWaitingForJoin:
		return []key.Binding{m.keymap.escape, m.keymap.start, m.keymap.quit}
	case StatePlaying:
		return m.getGameKeybindings()
	case StateReplacingCard:
		return append([]key.Binding{m.keymap.escape, m.keymap.replace}, m.getSelectableCardBindings(m.playerID)...)
	case StateLookingAtOwnCard:
		if peeking {
			return []key.Binding{m.keymap.escape, m.keymap.lookAtSelf}
		}
		return append([]key.Binding{m.keymap.escape, m.keymap.lookAtSelf}, m.getSelectableCardBindings(m.playerID)...)
	case StateLookingAtOpponentCard:
		if peeking {
			return []key.Binding{m.keymap.escape, m.keymap.lookAtOpponent}
		}
		bindings := []key.Binding{m.keymap.escape, m.keymap.left, m.keymap.right, m.keymap.lookAtOpponent}
		if m.selectedOpponent >= 0 {
			bindings = append(bindings, m.getSelectableCardBindings(m.selectedOpponent)...)
		}
		return bindings
	case StateShowingResult:
		return []key.Binding{m.keymap.start}
	case StatePlayAgain:
		return []key.Binding{m.keymap.start, m.keymap.quit}
	default:
		return []key.Binding{m.keymap.quit}
	}
}

func (m Model) getGameKeybindings() []key.Binding {
	if m.game == nil {
		return []key.Binding{m.keymap.quit}
	}
	if m.playerID != m.game.GetPlayerTurn() {
		return []key.Binding{m.keymap.quit}
	}

	switch m.game.GetGameState() {
	case internalcambio.GameStart:
		return []key.Binding{m.keymap.start, m.keymap.quit}
	case internalcambio.WaitingForTurn:
		return []key.Binding{m.keymap.draw, m.keymap.quit}
	case internalcambio.TakingTurn:
		bindings := make([]key.Binding, 0, 8)
		for _, turnType := range m.game.GetValidTurnTypes() {
			switch turnType {
			case internalcambio.Discard:
				bindings = append(bindings, m.keymap.discard)
			case internalcambio.Replace:
				bindings = append(bindings, m.keymap.replace)
			case internalcambio.LookAtSelf:
				bindings = append(bindings, m.keymap.lookAtSelf)
			case internalcambio.LookAtOpponent:
				bindings = append(bindings, m.keymap.lookAtOpponent)
			case internalcambio.Swap:
				bindings = append(bindings, m.keymap.swap)
			case internalcambio.CallGame:
				bindings = append(bindings, m.keymap.cambio)
			}
		}
		return append(bindings, m.keymap.quit)
	default:
		return []key.Binding{m.keymap.quit}
	}
}
