package blackjack

import (
	"errors"
	"math/rand"

	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

var (
	ErrRoundInProgress = errors.New("round already in progress")
	ErrRoundNotStarted = errors.New("round not started")
	ErrRoundOver       = errors.New("round is over")
)

type Game struct {
	deck   cards.Deck
	player Player
	dealer Player
	isOver bool
}

func NewGame() Game {
	return Game{
		deck:   cards.NewDeck(4, false),
		player: Player{hand: []cards.Card{}, isDealer: false},
		dealer: Player{hand: []cards.Card{}, isDealer: true},
		isOver: true,
	}
}

func (game *Game) InitialDeal() error {
	if !game.isOver {
		return ErrRoundInProgress
	}

	game.player.hand = game.player.hand[:0]
	game.dealer.hand = game.dealer.hand[:0]

	reshuffleThreshold := rand.Intn(10) + 25
	if game.deck.AvailableCards() < reshuffleThreshold {
		game.deck = cards.NewDeck(4, false)
	}

	for range 2 {
		card, err := game.deck.Draw()
		if err != nil {
			return err
		}
		game.player.hand = append(game.player.hand, card)

		card, err = game.deck.Draw()
		if err != nil {
			return err
		}
		game.dealer.hand = append(game.dealer.hand, card)
	}
	game.isOver = false
	return nil
}

func (game *Game) PlayerHit() (cards.Card, error) {
	if len(game.player.hand) == 0 {
		var emptyCard cards.Card
		return emptyCard, ErrRoundNotStarted
	}
	if game.isOver {
		var emptyCard cards.Card
		return emptyCard, ErrRoundOver
	}
	card, err := game.deck.Draw()
	if err != nil {
		var emptyCard cards.Card
		return emptyCard, err
	}
	game.player.hand = append(game.player.hand, card)
	if game.player.Score() > 21 {
		game.isOver = true
		return card, nil
	}
	return card, nil
}

func (game *Game) PlayerStand() error {
	if len(game.player.hand) == 0 {
		return ErrRoundNotStarted
	}
	if game.isOver {
		return ErrRoundOver
	}
	game.isOver = true
	return nil
}

func (game *Game) DealerPlay() error {
	if len(game.dealer.hand) == 0 {
		return ErrRoundNotStarted
	}
	for game.dealer.Score() < 17 {
		card, err := game.deck.Draw()
		if err != nil {
			return err
		}
		game.dealer.hand = append(game.dealer.hand, card)
	}
	return nil
}

func (game *Game) DetermineWinner() int {
	playerScore := game.PlayerScore()
	dealerScore := game.DealerScore()

	if playerScore > 21 {
		return 1
	} else if dealerScore > 21 {
		return 2
	} else if playerScore > dealerScore {
		return 2
	} else if dealerScore > playerScore {
		return 1
	} else {
		return 0
	}
}

func (game *Game) PlayerScore() int {
	return game.player.Score()
}

func (game *Game) DealerScore() int {
	return game.dealer.Score()
}

func (game *Game) IsOver() bool {
	return game.isOver
}

func (game *Game) PlayerHand() []cards.Card {
	hand := make([]cards.Card, len(game.player.hand))
	copy(hand, game.player.hand)
	return hand
}

func (game *Game) DealerHand() ([]cards.Card, error) {
	if !game.isOver {
		var emptyHand []cards.Card
		return emptyHand, ErrRoundInProgress
	}
	hand := make([]cards.Card, len(game.dealer.hand))
	copy(hand, game.dealer.hand)
	return hand, nil
}

func (game *Game) DealerVisibleCard() (cards.Card, error) {
	if len(game.dealer.hand) == 0 {
		var emptyCard cards.Card
		return emptyCard, ErrRoundNotStarted
	}
	return game.dealer.hand[0], nil
}
