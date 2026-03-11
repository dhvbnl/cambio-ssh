package cambioui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	internalcambio "github.com/dhvbnl/cambio-ssh/cmd/internal/cambio"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

func (m *Model) handleGameplayKey(msg tea.KeyMsg) {
	if m.game == nil {
		return
	}

	if key.Matches(msg, m.keymap.start) {
		if m.game.GetGameState() == internalcambio.GameStart {
			m.game.StartGame()
			m.message = ""
		}
		return
	}

	if m.playerID != m.game.GetPlayerTurn() {
		return
	}

	if key.Matches(msg, m.keymap.draw) {
		if m.game.GetGameState() == internalcambio.WaitingForTurn {
			m.game.DrawCard()
			m.message = ""
		}
		return
	}

	if key.Matches(msg, m.keymap.discard) {
		m.game.SelectTurnType(internalcambio.Discard)
		if err := m.game.DiscardCard(); err != nil {
			m.message = err.Error()
			return
		}
		m.message = ""
		return
	}

	if key.Matches(msg, m.keymap.replace) {
		m.resetSelectionState()
		m.state = StateReplacingCard
		return
	}

	if key.Matches(msg, m.keymap.lookAtSelf) {
		m.resetSelectionState()
		m.state = StateLookingAtOwnCard
		return
	}

	if key.Matches(msg, m.keymap.lookAtOpponent) {
		m.resetSelectionState()
		m.selectedOpponent = m.firstOpponentIndex()
		m.state = StateLookingAtOpponentCard
		if m.selectedOpponent >= 0 {
			m.message = fmt.Sprintf("Player %d selected. Use left/right to switch opponents, choose a card, then press o.", relativePlayerNumber(m.playerID, m.selectedOpponent, len(m.game.GetAllPlayerHands())))
		}
	}
}

func (m *Model) handleReplaceCardKey(msg tea.KeyMsg) {
	if m.game == nil {
		return
	}

	if key.Matches(msg, m.keymap.replace) {
		if m.selectedCard < 0 {
			m.state = StatePlaying
			m.message = ""
			return
		}

		m.game.SelectTurnType(internalcambio.Replace)
		if err := m.game.ReplaceCard(m.selectedCard); err != nil {
			m.message = err.Error()
			return
		}

		m.selectedCard = -1
		m.state = StatePlaying
		m.message = ""
		return
	}

	selected, ok := m.getSelectedCardIndexFromKey(msg)
	if !ok {
		return
	}

	hand := m.game.GetPlayerHand(m.playerID)
	if selected < 0 || selected >= len(hand) {
		return
	}
	if hand[selected] == (cards.Card{}) {
		return
	}

	m.selectedCard = selected
}

func (m *Model) handleLookOwnCardKey(msg tea.KeyMsg) {
	if m.game == nil {
		return
	}

	if key.Matches(msg, m.keymap.lookAtSelf) {
		if m.peekActive {
			m.finishOwnCardPeek()
			return
		}

		if m.selectedCard < 0 {
			m.message = "Select a card, then press s to peek at it."
			return
		}

		m.game.SelectTurnType(internalcambio.LookAtSelf)
		card, err := m.game.LookAtOwnCard(m.selectedCard)
		if err != nil {
			m.message = err.Error()
			return
		}

		m.peekActive = true
		m.message = fmt.Sprintf("Looking at card %d: %s. Press s again to continue.", m.selectedCard+1, card.String())
		return
	}

	if m.peekActive {
		return
	}

	selected, ok := m.getSelectedCardIndexFromKey(msg)
	if !ok {
		return
	}

	hand := m.game.GetPlayerHand(m.playerID)
	if selected < 0 || selected >= len(hand) {
		return
	}
	if hand[selected] == (cards.Card{}) {
		return
	}

	m.selectedCard = selected
	m.message = fmt.Sprintf("Card %d selected. Press s to peek at it.", m.selectedCard+1)
}

func (m *Model) finishOwnCardPeek() {
	m.game.EndTurn()
	m.resetSelectionState()
	m.state = StatePlaying
}

func (m *Model) handleLookOpponentCardKey(msg tea.KeyMsg) {
	if m.game == nil {
		return
	}

	if key.Matches(msg, m.keymap.lookAtOpponent) {
		if m.peekActive {
			m.finishOpponentCardPeek()
			return
		}

		if m.selectedOpponent < 0 {
			m.message = "Select an opponent, then choose a card and press o to peek."
			return
		}

		if m.selectedCard < 0 {
			m.message = "Select an opponent card, then press o to peek at it."
			return
		}

		m.game.SelectTurnType(internalcambio.LookAtOpponent)
		card, err := m.game.LookAtOpponentCard(m.selectedOpponent, m.selectedCard)
		if err != nil {
			m.message = err.Error()
			return
		}

		m.peekActive = true
		m.message = fmt.Sprintf("Looking at Player %d card %d: %s. Press o again to continue.", relativePlayerNumber(m.playerID, m.selectedOpponent, len(m.game.GetAllPlayerHands())), m.selectedCard+1, card.String())
		return
	}

	if m.peekActive {
		return
	}

	if key.Matches(msg, m.keymap.left) {
		m.shiftSelectedOpponent(-1)
		return
	}

	if key.Matches(msg, m.keymap.right) {
		m.shiftSelectedOpponent(1)
		return
	}

	selected, ok := m.getSelectedCardIndexFromKey(msg)
	if !ok || m.selectedOpponent < 0 {
		return
	}

	hand := m.game.GetPlayerHand(m.selectedOpponent)
	if selected < 0 || selected >= len(hand) {
		return
	}
	if hand[selected] == (cards.Card{}) {
		return
	}

	m.selectedCard = selected
	m.message = fmt.Sprintf("Player %d card %d selected. Press o to peek at it.", relativePlayerNumber(m.playerID, m.selectedOpponent, len(m.game.GetAllPlayerHands())), m.selectedCard+1)
}

func (m *Model) finishOpponentCardPeek() {
	m.game.EndTurn()
	m.resetSelectionState()
	m.state = StatePlaying
}

func (m *Model) firstOpponentIndex() int {
	if m.game == nil {
		return -1
	}
	playerCount := len(m.game.GetAllPlayerHands())
	if playerCount <= 1 {
		return -1
	}
	return (m.playerID + 1) % playerCount
}

func (m *Model) shiftSelectedOpponent(delta int) {
	if m.game == nil {
		return
	}

	playerCount := len(m.game.GetAllPlayerHands())
	if playerCount <= 1 {
		return
	}

	if m.selectedOpponent < 0 || m.selectedOpponent >= playerCount {
		m.selectedOpponent = m.firstOpponentIndex()
	} else {
		candidate := m.selectedOpponent
		for step := 0; step < playerCount; step++ {
			candidate = (candidate + delta + playerCount) % playerCount
			if candidate != m.playerID {
				m.selectedOpponent = candidate
				break
			}
		}
	}

	m.selectedCard = -1
	m.message = fmt.Sprintf("Player %d selected. Choose a card, then press o to peek.", relativePlayerNumber(m.playerID, m.selectedOpponent, playerCount))
}

func relativePlayerNumber(currentPlayer int, targetPlayer int, playerCount int) int {
	if playerCount <= 0 {
		return 1
	}
	return ((targetPlayer-currentPlayer+playerCount)%playerCount + 1)
}
