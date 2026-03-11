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
	b.WriteString("Press ENTER or SPACE to start a new game\n")
	b.WriteString("Press CTRL+C to quit\n\n")
	b.WriteString("How to play:\n")
	b.WriteString("- Goal: Have the lowest total card value at the end.\n")
	b.WriteString("- Each player starts with 4 face-down cards.\n")
	b.WriteString("- On your turn, draw a card and replace or discard it.\n")
	b.WriteString("- Special cards allow peeking and swapping effects.\n")
	b.WriteString("- Call Cambio when you think you have the lowest score.\n")
	return b.String()
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

	if currentPlayer < 0 || currentPlayer >= len(playerHands) {
		currentPlayer = 0
	}

	revealedPlayer, revealedIndex, _, _, revealedActive := m.game.GetRevealedCard(currentPlayer)

	topRowHands := make([]string, 0, len(playerHands)-1)
	relativeNum := 2
	for offset := 1; offset < len(playerHands); offset++ {
		idx := (currentPlayer + offset) % len(playerHands)
		title := fmt.Sprintf("Player %d Hand", relativeNum)
		highlightedIndex := -1
		if revealedActive && revealedPlayer == idx {
			highlightedIndex = revealedIndex
		}
		targetedPlayer := m.state == StateLookingAtOpponentCard && m.selectedOpponent == idx
		targetedSelectedCard := -1
		if targetedPlayer {
			targetedSelectedCard = m.selectedCard
		}
		topRowHands = append(topRowHands, renderPlayerHandBox(title, playerHands[idx], false, false, m.state, m.peekActive, targetedPlayer, targetedSelectedCard, highlightedIndex))
		relativeNum++
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
		m.state,
		m.peekActive,
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
	if revealedActive {
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

func renderPlayerHandBox(title string, hand []cards.Card, gameStart bool, faceUp bool, mode State, peekActive bool, targetedPlayer bool, selectedIndex int, revealedIndex int) string {
	handDisplay := "No cards"
	if len(hand) > 0 {
		columns := make([]string, 0, (len(hand)+1)/2)
		for topIdx := 0; topIdx < len(hand); topIdx += 2 {
			showFace := faceUp || gameStart
			topStyle := selectableCardStyle(mode, peekActive, targetedPlayer, topIdx, selectedIndex, revealedIndex, hand[topIdx])
			topCard := renderCard(hand[topIdx], false, topIdx+1, topStyle)

			bottomCard := renderEmptyCard()
			if bottomIdx := topIdx + 1; bottomIdx < len(hand) {
				bottomStyle := selectableCardStyle(mode, peekActive, targetedPlayer, bottomIdx, selectedIndex, revealedIndex, hand[bottomIdx])
				bottomCard = renderCard(hand[bottomIdx], showFace, bottomIdx+1, bottomStyle)
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
	switch m.state {
	case StateInitial:
		return []key.Binding{m.keymap.join, m.keymap.quit}
	case StatePlaying:
		return m.getGameKeybindings()
	case StateReplacingCard:
		return append([]key.Binding{m.keymap.escape, m.keymap.replace}, m.getSelectableCardBindings(m.playerID)...)
	case StateLookingAtOwnCard:
		if m.peekActive {
			return []key.Binding{m.keymap.escape, m.keymap.lookAtSelf}
		}
		return append([]key.Binding{m.keymap.escape, m.keymap.lookAtSelf}, m.getSelectableCardBindings(m.playerID)...)
	case StateLookingAtOpponentCard:
		if m.peekActive {
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
