package cards

import (
	"errors"
	"math/rand"
)

var ErrDeckEmpty = errors.New("deck is empty")

type Deck struct {
	DrawPile    []Card
	DiscardPile []Card
}

// Creates a new standard deck of 52 playing cards
func NewDeck(deckMultiple int) Deck {
	var cards []Card
	for suit := Spades; suit <= Clubs; suit++ {
		for rank := Ace; rank <= King; rank++ {
			cards = append(cards, Card{Suit: suit, Rank: rank})
		}
	}
	for i := 1; i < deckMultiple; i++ {
		cards = append(cards, cards...)
	}

	var newDeck Deck = Deck{DrawPile: cards, DiscardPile: []Card{}}
	newDeck.Shuffle()
	return newDeck
}

func (deck *Deck) Shuffle() {
	rand.Shuffle(len(deck.DrawPile), func(i, j int) {
		deck.DrawPile[i], deck.DrawPile[j] = deck.DrawPile[j], deck.DrawPile[i]
	})
}

func (deck *Deck) Draw() (Card, error) {
	if len(deck.DrawPile) == 0 {
		return Card{}, ErrDeckEmpty
	}
	card := deck.DrawPile[0]
	deck.DrawPile = deck.DrawPile[1:]
	return card, nil
}

func (deck *Deck) AvailableCards() int {
	return len(deck.DrawPile)
}
