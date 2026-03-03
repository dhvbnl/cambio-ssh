package blackjack

import "github.com/dhvbnl/cambio-ssh/cmd/internal/cards"

type Player struct {
	hand     []cards.Card
	isDealer bool
}

func (player *Player) Score() int {
	score := 0
	aces := 0
	for _, card := range player.hand {
		if card.Rank == cards.Ace {
			aces++
			score += 11
		} else if card.Rank >= cards.Jack {
			score += 10
		} else {
			score += int(card.Rank)
		}
	}
	for i := 0; i < aces; i++ {
		if score > 21 {
			score -= 10
		}
	}
	return score
}
