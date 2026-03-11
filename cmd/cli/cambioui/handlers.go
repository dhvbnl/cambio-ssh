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
		m.selectedCard = -1
		m.peekActive = false
		m.state = StateReplacingCard
		m.message = ""
		return
	}

	if key.Matches(msg, m.keymap.lookAtSelf) {
		m.selectedCard = -1
		m.peekActive = false
		m.state = StateLookingAtOwnCard
		m.message = ""
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
	m.selectedCard = -1
	m.peekActive = false
	m.state = StatePlaying
	m.message = ""
}
