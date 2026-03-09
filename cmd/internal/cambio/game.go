package cambio

import (
	"errors"
	"slices"
	"sync"

	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

var ErrInvalidTurn = errors.New("invalid turn")

const startingHandSize = 4

type turnType string

const (
	discard        turnType = "d"
	replace        turnType = "r"
	lookAtSelf     turnType = "e"
	lookAtOpponent turnType = "l"
	swap           turnType = "s"
	callGame       turnType = "c"
	unselected     turnType = "u"
)

type gameState string

const (
	gameStart      gameState = "game_start"
	waitingForTurn gameState = "waiting_for_turn"
	takingTurn     gameState = "taking_turn"
	removingCard   gameState = "removing_card"
	gameOver       gameState = "game_over"
)

type replaceInfo struct {
	playerReplacing      int
	playersCardReplacing int
	cardIndex            int
}

type Game struct {
	mu               sync.RWMutex
	playerCount      int // always 2 for now, but could be extended in the future
	deck             cards.Deck
	playerCards      [][]cards.Card // playerCards[0] is player 1's hand, playerCards[1] is player 2's hand
	activeCard       cards.Card     // the card that is currently being played, i.e. the card that was just drawn
	playerCalledGame int            // -1 means no one has called game yet; otherwise stores player index
	playerTurn       int            // corresponds to player index, so 0 for player 1, 1 for player 2
	turnType         turnType       // the type of action the current player is taking
	validTurnTypes   []turnType     // the valid turn types for the current active card
	gameState        gameState      // the current state of the game
	savedGameState   gameState      // used for when a player removes card so we can return to the previous game state after removal is done
	replacing        *replaceInfo   // information about the card being replaced
}

func NewGame(playerCount int) *Game {
	game := Game{
		playerCount:      playerCount,
		deck:             cards.NewDeck(1, true),
		playerCards:      make([][]cards.Card, playerCount),
		playerCalledGame: -1,
		playerTurn:       0,
		gameState:        gameStart,
		turnType:         unselected,
	}

	game.dealStartingHands(startingHandSize)
	return &game
}

func (game *Game) dealStartingHands(cardsPerPlayer int) {
	for i := 0; i < cardsPerPlayer; i++ {
		for player := 0; player < game.playerCount; player++ {
			game.playerCards[player] = append(game.playerCards[player], game.deck.DrawWithReshuffle())
		}
	}
}

func (game *Game) clearActiveCard(discard bool) {
	if game.activeCard == (cards.Card{}) {
		return
	}
	if discard {
		game.deck.Discard(game.activeCard)
	}
	game.activeCard = cards.Card{}
}

func (game *Game) advanceTurn() {
	if game.playerCalledGame >= 0 && game.playerCalledGame == (game.playerTurn+1)%game.playerCount {
		game.gameState = gameOver
		game.turnType = unselected
		game.validTurnTypes = nil
		return
	}

	game.playerTurn = (game.playerTurn + 1) % game.playerCount
	game.gameState = waitingForTurn
	game.turnType = unselected
	game.validTurnTypes = nil
}

// State machine validators

// validateTurnAction checks if a consuming action (discard/replace) can execute.
func (game *Game) validateTurnAction(expectedTurnType turnType) error {
	if game.gameState != takingTurn || game.turnType != expectedTurnType {
		return ErrInvalidTurn
	}
	return nil
}

// validateViewAction checks if a non-consuming action (look/swap) can execute in takingTurn.
func (game *Game) validateViewAction(expectedTurnType turnType) error {
	if game.gameState != takingTurn || game.turnType != expectedTurnType {
		return ErrInvalidTurn
	}
	return nil
}

// validateRemoveAction checks if a card can be removed during the removingCard state.
func (game *Game) validateRemoveAction() error {
	if game.gameState != removingCard && game.gameState != takingTurn {
		return ErrInvalidTurn
	}
	return nil
}

func (game *Game) StartGame() {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != gameStart {
		return
	}
	game.gameState = waitingForTurn
}

func DrawCard(game *Game) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != waitingForTurn {
		return
	}
	card := game.deck.DrawWithReshuffle()
	game.activeCard = card
	game.gameState = takingTurn
	game.turnType = unselected
	game.validTurnTypes = getValidTurnType(game)
}

func SelectTurnType(game *Game, turn turnType) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != takingTurn || !slices.Contains(game.validTurnTypes, turn) {
		return
	}
	game.turnType = turn
}

func DiscardCard(game *Game) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateTurnAction(discard); err != nil {
		return err
	}

	game.clearActiveCard(true)
	game.advanceTurn()
	return nil
}

func ReplaceCard(game *Game, cardIndex int) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateTurnAction(replace); err != nil {
		return err
	}

	if cardIndex < 0 || cardIndex >= len(game.playerCards[game.playerTurn]) {
		return errors.New("invalid card index")
	}

	replacedCard := game.playerCards[game.playerTurn][cardIndex]
	game.playerCards[game.playerTurn][cardIndex] = game.activeCard
	game.deck.Discard(replacedCard)
	game.clearActiveCard(false)
	game.advanceTurn()
	return nil
}

func LookAtOwnCard(game *Game, cardIndex int) (cards.Card, error) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateViewAction(lookAtSelf); err != nil {
		return cards.Card{}, err
	}

	if cardIndex < 0 || cardIndex >= len(game.playerCards[game.playerTurn]) {
		return cards.Card{}, errors.New("invalid card index")
	}

	return game.playerCards[game.playerTurn][cardIndex], nil
}

func LookAtOpponentCard(game *Game, opponentIndex int, cardIndex int) (cards.Card, error) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateViewAction(lookAtOpponent); err != nil {
		return cards.Card{}, err
	}

	if opponentIndex < 0 || opponentIndex >= game.playerCount || opponentIndex == game.playerTurn {
		return cards.Card{}, errors.New("invalid opponent index")
	}

	if cardIndex < 0 || cardIndex >= len(game.playerCards[opponentIndex]) {
		return cards.Card{}, errors.New("invalid card index")
	}

	return game.playerCards[opponentIndex][cardIndex], nil
}

func SwapCards(game *Game, opponentIndex int, ownCardIndex int, opponentCardIndex int) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateViewAction(swap); err != nil {
		return err
	}

	if opponentIndex < 0 || opponentIndex >= game.playerCount || opponentIndex == game.playerTurn {
		return errors.New("invalid opponent index")
	}

	if ownCardIndex < 0 || ownCardIndex >= len(game.playerCards[game.playerTurn]) {
		return errors.New("invalid own card index")
	}

	if opponentCardIndex < 0 || opponentCardIndex >= len(game.playerCards[opponentIndex]) {
		return errors.New("invalid opponent card index")
	}

	game.playerCards[game.playerTurn][ownCardIndex], game.playerCards[opponentIndex][opponentCardIndex] =
		game.playerCards[opponentIndex][opponentCardIndex], game.playerCards[game.playerTurn][ownCardIndex]

	return nil
}

func RemoveCard(game *Game, playerReplacing int, playersCardReplacing int, cardIndex int) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateRemoveAction(); err != nil {
		return err
	}

	if playerReplacing < 0 || playerReplacing >= game.playerCount {
		return errors.New("invalid player index")
	}

	if playersCardReplacing < 0 || playersCardReplacing >= game.playerCount {
		return errors.New("invalid player index")
	}

	if cardIndex < 0 || cardIndex >= len(game.playerCards[playersCardReplacing]) {
		return errors.New("invalid card index")
	}

	if game.playerCards[playersCardReplacing][cardIndex].Rank != game.deck.GetTopDiscardCard().Rank {
		// add a new card to players hand as penalty for trying to replace with a card that doesn't match the rank of the top card of the discard pile
		newCard := game.deck.DrawWithReshuffle()
		game.playerCards[playersCardReplacing] = append(game.playerCards[playersCardReplacing], newCard)
		return errors.New("can only replace with a card that matches the rank of the top card of the discard pile, penalty card added to hand")
	}

	// remove the card from the player's hand and leave blank to preserve card indices if player replacing own card
	if playerReplacing == playersCardReplacing {
		game.playerCards[playersCardReplacing][cardIndex] = cards.Card{}
		return nil
	}

	// remove card from opponent's hand if player replacing opponent's card and set state to player can give opponent a card of their choice to replace the removed card
	game.playerCards[playersCardReplacing][cardIndex] = cards.Card{}
	game.savedGameState = game.gameState
	game.gameState = removingCard
	game.replacing = &replaceInfo{
		playerReplacing:      playerReplacing,
		playersCardReplacing: playersCardReplacing,
		cardIndex:            cardIndex,
	}
	return nil
}

// EndTurn commits a look/swap action and advances to the next player.
func EndTurn(game *Game) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != takingTurn {
		return
	}

	game.clearActiveCard(true)
	game.advanceTurn()
}

func EndGame(game *Game) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != waitingForTurn {
		return
	}

	game.playerCalledGame = game.playerTurn
	game.playerTurn = (game.playerTurn + 1) % game.playerCount
	game.gameState = waitingForTurn
	game.turnType = unselected
}

func GetGameStart(game *Game) bool {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.gameState == gameStart
}

func GetActivePlayer(game *Game) int {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.playerTurn + 0
}

func GetTopDiscardCard(game *Game) cards.Card {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.deck.GetTopDiscardCard()
}

func GetPlayerHand(game *Game, playerIndex int) []cards.Card {
	game.mu.RLock()
	defer game.mu.RUnlock()
	if playerIndex < 0 || playerIndex >= game.playerCount {
		return nil
	}
	hand := make([]cards.Card, len(game.playerCards[playerIndex]))
	copy(hand, game.playerCards[playerIndex])
	return hand
}

func GetAllPlayerHands(game *Game) [][]cards.Card {
	game.mu.RLock()
	defer game.mu.RUnlock()

	hands := make([][]cards.Card, game.playerCount)
	for i := 0; i < game.playerCount; i++ {
		hands[i] = make([]cards.Card, len(game.playerCards[i]))
		copy(hands[i], game.playerCards[i])
	}
	return hands
}

func GetActiveCard(game *Game, playerIndex int) cards.Card {
	game.mu.RLock()
	defer game.mu.RUnlock()
	if playerIndex != game.playerTurn {
		return cards.Card{}
	}
	return game.activeCard
}

func getValidTurnType(game *Game) []turnType {
	switch game.activeCard.Rank {
	case cards.Ace, cards.Two, cards.Three, cards.Four, cards.Five, cards.Six:
		return []turnType{discard, replace}
	case cards.Seven, cards.Eight:
		return []turnType{lookAtSelf, discard, replace}
	case cards.Nine, cards.Ten:
		return []turnType{lookAtOpponent, discard, replace}
	case cards.Jack, cards.Queen, cards.King:
		if game.activeCard.Suit == cards.Hearts || game.activeCard.Suit == cards.Diamonds {
			return []turnType{discard, replace}
		} else {
			return []turnType{swap, discard, replace}
		}
	default:
		return []turnType{callGame}
	}
}
