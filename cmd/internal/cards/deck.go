package cards

import (
	"errors"
	"math/rand"
)

var ErrDeckEmpty = errors.New("deck is empty")

type Deck struct {
	drawPile    []Card
	inPlayPile  []Card
	discardPile []Card
}

// Creates a new standard deck of 52 playing cards
func NewDeck(deckMultiple int, inPlayPile bool) Deck {
	var cards []Card
	for suit := Spades; suit <= Clubs; suit++ {
		for rank := Ace; rank <= King; rank++ {
			cards = append(cards, Card{Suit: suit, Rank: rank})
		}
	}
	for i := 1; i < deckMultiple; i++ {
		cards = append(cards, cards...)
	}

	var newDeck Deck = Deck{drawPile: cards, discardPile: []Card{}}
	if inPlayPile {
		newDeck.inPlayPile = []Card{}
	}
	newDeck.Shuffle()
	return newDeck
}

func (deck *Deck) Shuffle() {
	rand.Shuffle(len(deck.drawPile), func(i, j int) {
		deck.drawPile[i], deck.drawPile[j] = deck.drawPile[j], deck.drawPile[i]
	})
}

func (deck *Deck) Draw() (Card, error) {
	if len(deck.drawPile) == 0 {
		return Card{}, ErrDeckEmpty
	}
	card := deck.drawPile[0]
	deck.drawPile = deck.drawPile[1:]
	if deck.inPlayPile != nil {
		deck.inPlayPile = append(deck.inPlayPile, card)
	} else {
		deck.discardPile = append(deck.discardPile, card)
	}
	return card, nil
}

func (deck *Deck) DrawWithReshuffle() Card {
	card, err := deck.Draw()
	if err == ErrDeckEmpty {
		deck.drawPile = append(deck.drawPile, deck.discardPile...)
		deck.discardPile = []Card{}
		deck.Shuffle()
		card, _ = deck.Draw()
	}
	return card
}

func (deck *Deck) Discard(card Card) {
	if deck.inPlayPile != nil {
		for i, c := range deck.inPlayPile {
			if c == card {
				deck.inPlayPile = append(deck.inPlayPile[:i], deck.inPlayPile[i+1:]...)
				break
			}
		}
	}
	deck.discardPile = append(deck.discardPile, card)
}

func (deck *Deck) AvailableCards() int {
	return len(deck.drawPile)
}

func (deck *Deck) GetTopDiscardCard() Card {
	if len(deck.discardPile) == 0 {
		return Card{}
	}
	return deck.discardPile[len(deck.discardPile)-1]
}
