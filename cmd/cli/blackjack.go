package cli

import (
	"fmt"

	"github.com/dhvbnl/cambio-ssh/cmd/internal/blackjack"
	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// PlayBlackjack starts a blackjack game session in the terminal
func PlayBlackjack() {
	game := blackjack.NewGame()

	fmt.Printf("%s%sWelcome to Blackjack!%s\n\n", colorBold, colorCyan, colorReset)

	for {
		fmt.Printf("\n%s%s━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorBold, colorYellow, colorReset)
		fmt.Printf("%s%s       NEW ROUND%s\n", colorBold, colorYellow, colorReset)
		fmt.Printf("%s%s━━━━━━━━━━━━━━━━━━━━━━━━%s\n\n", colorBold, colorYellow, colorReset)

		if game.IsOver() {
			if err := game.InitialDeal(); err != nil {
				fmt.Printf("%sError dealing initial cards: %v%s\n", colorRed, err, colorReset)
				continue
			}
		}

		dealerCard, err := game.DealerVisibleCard()
		if err != nil {
			fmt.Printf("%sError reading dealer card: %v%s\n", colorRed, err, colorReset)
			continue
		}

		for !game.IsOver() {
			displayGameState(&game, dealerCard)

			action := getPlayerAction()
			switch action {
			case "hit", "h":
				fmt.Printf("\n%sYou chose to HIT!%s\n", colorBlue, colorReset)
				card, err := game.PlayerHit()
				if err != nil {
					fmt.Printf("%sError drawing card: %v%s\n", colorRed, err, colorReset)
					continue
				}
				fmt.Printf("%sYou drew: %v%s\n", colorCyan, card, colorReset)
				if game.PlayerScore() > 21 {
					fmt.Printf("%sBUST! You went over 21!%s\n", colorRed, colorReset)
				}
			case "stand", "s":
				fmt.Printf("\n%sYou chose to STAND with score: %d%s\n", colorBlue, game.PlayerScore(), colorReset)
				if err := game.PlayerStand(); err != nil {
					fmt.Printf("%sError standing: %v%s\n", colorRed, err, colorReset)
				}
			default:
				fmt.Printf("%sInvalid action. Please choose 'hit' or 'stand'.%s\n", colorYellow, colorReset)
			}
		}

		fmt.Printf("\n%sDealer's turn...%s\n", colorCyan, colorReset)
		if err := game.DealerPlay(); err != nil {
			fmt.Printf("%sError during dealer's play: %v%s\n", colorRed, err, colorReset)
			continue
		}

		displayEndState(&game)

		if !askPlayAgain() {
			fmt.Printf("\n%s%sThanks for playing! See you next time!%s\n\n", colorBold, colorCyan, colorReset)
			break
		}
	}
}

func displayGameState(game *blackjack.Game, dealerCard cards.Card) {
	playerCards := game.PlayerHand()
	playerScore := game.PlayerScore()

	fmt.Printf("%s┌───────────────────────┐%s\n", colorCyan, colorReset)
	fmt.Printf("%s│%s       YOUR HAND       %s│%s\n", colorCyan, colorBold, colorCyan, colorReset)
	fmt.Printf("%s├───────────────────────┤%s\n", colorCyan, colorReset)
	fmt.Printf("%s│%s  Cards: %v\n", colorCyan, colorReset, playerCards)

	scoreColor := colorGreen
	if playerScore > 21 {
		scoreColor = colorRed
	} else if playerScore >= 17 {
		scoreColor = colorYellow
	}
	fmt.Printf("%s│%s  Score: %s%s%d%s\n", colorCyan, colorReset, colorBold, scoreColor, playerScore, colorReset)
	fmt.Printf("%s└───────────────────────┘%s\n\n", colorCyan, colorReset)

	fmt.Printf("%sDealer's visible card: %s%v%s\n\n", colorYellow, colorBold, dealerCard, colorReset)
}

func displayEndState(game *blackjack.Game) {
	dealerCards, err := game.DealerHand()
	if err != nil {
		fmt.Printf("%sError reading dealer hand: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("\n%s%s━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorBold, colorYellow, colorReset)
	fmt.Printf("%s%s     FINAL RESULTS%s\n", colorBold, colorYellow, colorReset)
	fmt.Printf("%s%s━━━━━━━━━━━━━━━━━━━━━━━━%s\n\n", colorBold, colorYellow, colorReset)

	fmt.Printf("%sDealer's hand:%s  %v\n", colorCyan, colorReset, dealerCards)
	fmt.Printf("%sDealer's score:%s %s%d%s\n\n", colorCyan, colorReset, colorBold, game.DealerScore(), colorReset)

	fmt.Printf("%sYour hand:%s      %v\n", colorCyan, colorReset, game.PlayerHand())
	fmt.Printf("%sYour score:%s     %s%d%s\n\n", colorCyan, colorReset, colorBold, game.PlayerScore(), colorReset)

	winner := game.DetermineWinner()
	switch winner {
	case 0:
		fmt.Printf("%s%sPUSH! It's a tie!%s\n", colorBold, colorYellow, colorReset)
	case 1:
		fmt.Printf("%s%sDealer wins!%s\n", colorBold, colorRed, colorReset)
	case 2:
		fmt.Printf("%s%sYOU WIN! Congratulations!%s\n", colorBold, colorGreen, colorReset)
	}
	fmt.Printf("%s%s━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorBold, colorYellow, colorReset)
}

func getPlayerAction() string {
	var action string
	fmt.Printf("%sChoose your action%s (hit/stand): ", colorBold, colorReset)
	fmt.Scanln(&action)
	return action
}

func askPlayAgain() bool {
	var playAgain string
	fmt.Printf("\n%sPlay another round?%s (yes/no): ", colorCyan, colorReset)
	fmt.Scanln(&playAgain)
	return playAgain == "yes"
}
