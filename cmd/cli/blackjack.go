package cli

import (
	"fmt"

	"github.com/dhvbnl/cambio-ssh/cmd/internal/blackjack"
)

// PlayBlackjack starts a blackjack game session in the terminal
func PlayBlackjack() {
	game := blackjack.NewGame()

	for {
		fmt.Printf("\n\n=== New Round ===\n")
		if game.IsOver() {
			if err := game.InitialDeal(); err != nil {
				fmt.Printf("Error dealing initial cards: %v\n", err)
				continue
			}
		}

		dealerCard, err := game.DealerVisibleCard()
		if err != nil {
			fmt.Printf("Error reading dealer card: %v\n", err)
			continue
		}

		for !game.IsOver() {
			displayGameState(&game, dealerCard)

			action := getPlayerAction()
			switch action {
			case "hit":
				fmt.Printf("You chose to hit.\n")
				card, err := game.PlayerHit()
				if err != nil {
					fmt.Printf("Error drawing card: %v\n", err)
					continue
				}
				fmt.Printf("You drew: %v\n", card)
			case "stand":
				fmt.Printf("You stand with score: %d\n", game.PlayerScore())
				if err := game.PlayerStand(); err != nil {
					fmt.Printf("Error standing: %v\n", err)
				}
			default:
				fmt.Println("Invalid action. Please choose 'hit' or 'stand'.")
			}
		}

		if err := game.DealerPlay(); err != nil {
			fmt.Printf("Error during dealer's play: %v\n", err)
			continue
		}

		displayEndState(&game)

		if !askPlayAgain() {
			fmt.Println("Thanks for playing!")
			break
		}
	}
}

func displayGameState(game *blackjack.Game, dealerCard interface{}) {
	playerCards := game.PlayerHand()
	fmt.Printf("Your score and hand: %d - %v\n", game.PlayerScore(), playerCards)
	fmt.Printf("Dealer's visible card: %v\n", dealerCard)
}

func displayEndState(game *blackjack.Game) {
	dealerCards, err := game.DealerHand()
	if err != nil {
		fmt.Printf("Error reading dealer hand: %v\n", err)
		return
	}

	fmt.Printf("\nDealer's hand: %v\n", dealerCards)
	fmt.Printf("Player's score: %d\n", game.PlayerScore())
	fmt.Printf("Dealer's score: %d\n", game.DealerScore())
	fmt.Printf("Result: %s\n", game.DetermineWinner())
}

func getPlayerAction() string {
	var action string
	fmt.Print("Choose action (hit/stand): ")
	fmt.Scanln(&action)
	return action
}

func askPlayAgain() bool {
	var playAgain string
	fmt.Print("\nPlay again? (yes/no): ")
	fmt.Scanln(&playAgain)
	return playAgain == "yes"
}
