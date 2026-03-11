package cambioui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

type keymap struct {
	createGame     key.Binding
	join           key.Binding
	start          key.Binding
	quit           key.Binding
	escape         key.Binding
	left           key.Binding
	right          key.Binding
	draw           key.Binding
	replace        key.Binding
	discard        key.Binding
	lookAtSelf     key.Binding
	lookAtOpponent key.Binding
	swap           key.Binding
	cambio         key.Binding
	cardKeys       [9]key.Binding
}

func newKeymap() keymap {
	km := keymap{
		createGame:     key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "create game")),
		join:           key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "join game")),
		start:          key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "start game")),
		quit:           key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		escape:         key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		left:           key.NewBinding(key.WithKeys("left"), key.WithHelp("left", "prev opponent")),
		right:          key.NewBinding(key.WithKeys("right"), key.WithHelp("right", "next opponent")),
		draw:           key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "draw card")),
		replace:        key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "replace / confirm")),
		discard:        key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "discard card")),
		lookAtSelf:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "look at own / confirm")),
		lookAtOpponent: key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "look at opponent cards")),
		swap:           key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "swap with opponent")),
		cambio:         key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "call cambio")),
	}

	for i := range len(km.cardKeys) {
		n := i + 1
		k := fmt.Sprintf("%d", n)
		km.cardKeys[i] = key.NewBinding(key.WithKeys(k), key.WithHelp(k, fmt.Sprintf("select card %d", n)))
	}

	return km
}

func (m Model) getSelectedCardIndexFromKey(msg tea.KeyMsg) (int, bool) {
	for idx, binding := range m.keymap.cardKeys {
		if key.Matches(msg, binding) {
			return idx, true
		}
	}
	return -1, false
}

func (m Model) getSelectableCardBindings(playerID int) []key.Binding {
	hand := m.game.GetPlayerHand(playerID)
	bindings := make([]key.Binding, 0, len(hand))
	for idx, card := range hand {
		if idx >= len(m.keymap.cardKeys) {
			break
		}
		if card == (cards.Card{}) {
			continue
		}
		bindings = append(bindings, m.keymap.cardKeys[idx])
	}
	return bindings
}
