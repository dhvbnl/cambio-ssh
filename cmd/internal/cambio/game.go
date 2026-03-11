package cambio

import (
	"errors"
	"slices"
	"sync"

	"github.com/dhvbnl/cambio-ssh/cmd/internal/cards"
)

var ErrInvalidTurn = errors.New("invalid turn")

const startingHandSize = 4

type TurnType string

const (
	Discard        TurnType = "i"
	Replace        TurnType = "r"
	LookAtSelf     TurnType = "e"
	LookAtOpponent TurnType = "l"
	Swap           TurnType = "s"
	CallGame       TurnType = "c"
	Unselected     TurnType = "u"
)

type GameState string

const (
	GameStart      GameState = "game_start"
	WaitingForTurn GameState = "waiting_for_turn"
	TakingTurn     GameState = "taking_turn"
	RemovingCard   GameState = "removing_card"
	GameOver       GameState = "game_over"
)

type replaceInfo struct {
	playerReplacing      int
	playersCardReplacing int
	cardIndex            int
}

type revealedCard struct {
	playerIndex int
	cardIndex   int
	viewerIndex int
	card        cards.Card
}

type Game struct {
	mu               sync.RWMutex
	playerCount      int // always 2 for now, but could be extended in the future
	deck             cards.Deck
	playerCards      [][]cards.Card // playerCards[0] is player 1's hand, playerCards[1] is player 2's hand
	activeCard       cards.Card     // the card that is currently being played, i.e. the card that was just drawn
	playerCalledGame int            // -1 means no one has called game yet; otherwise stores player index
	playerTurn       int            // corresponds to player index, so 0 for player 1, 1 for player 2
	turnType         TurnType       // the type of action the current player is taking
	validTurnTypes   []TurnType     // the valid turn types for the current active card
	gameState        GameState      // the current state of the game
	savedGameState   GameState      // used for when a player removes card so we can return to the previous game state after removal is done
	replacing        *replaceInfo   // information about the card being replaced
	revealed         *revealedCard  // the most recently revealed card, kept until the next draw
}

func NewGame(playerCount int) *Game {
	game := Game{
		playerCount:      playerCount,
		deck:             cards.NewDeck(1, true),
		playerCards:      make([][]cards.Card, playerCount),
		playerCalledGame: -1,
		playerTurn:       0,
		gameState:        GameStart,
		turnType:         Unselected,
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

func (game *Game) clearRevealedCard() {
	game.revealed = nil
}

func (game *Game) setRevealedCard(playerIndex int, cardIndex int, viewerIndex int, card cards.Card) {
	game.revealed = &revealedCard{
		playerIndex: playerIndex,
		cardIndex:   cardIndex,
		viewerIndex: viewerIndex,
		card:        card,
	}
}

func (game *Game) advanceTurn() {
	if game.playerCalledGame >= 0 && game.playerCalledGame == (game.playerTurn+1)%game.playerCount {
		game.gameState = GameOver
		game.turnType = Unselected
		game.validTurnTypes = nil
		return
	}

	game.playerTurn = (game.playerTurn + 1) % game.playerCount
	game.gameState = WaitingForTurn
	game.turnType = Unselected
	game.validTurnTypes = nil
}

// State machine validators

// validateTurnAction checks if a consuming action (discard/replace) can execute.
func (game *Game) validateTurnAction(expectedTurnType TurnType) error {
	if game.gameState != TakingTurn || game.turnType != expectedTurnType {
		return ErrInvalidTurn
	}
	return nil
}

// validateViewAction checks if a non-consuming action (look/swap) can execute in takingTurn.
func (game *Game) validateViewAction(expectedTurnType TurnType) error {
	if game.gameState != TakingTurn || game.turnType != expectedTurnType {
		return ErrInvalidTurn
	}
	return nil
}

// validateRemoveAction checks if a card can be removed during the removingCard state.
func (game *Game) validateRemoveAction() error {
	if game.gameState != RemovingCard && game.gameState != TakingTurn {
		return ErrInvalidTurn
	}
	return nil
}

func (game *Game) StartGame() {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != GameStart {
		return
	}
	game.gameState = WaitingForTurn
}

func (game *Game) DrawCard() {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != WaitingForTurn {
		return
	}
	game.clearRevealedCard()
	card := game.deck.DrawWithReshuffle()
	game.activeCard = card
	game.gameState = TakingTurn
	game.turnType = Unselected
	game.validTurnTypes = game.getValidTurnType()
}

func (game *Game) SelectTurnType(turn TurnType) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != TakingTurn || !slices.Contains(game.validTurnTypes, turn) {
		return
	}
	game.turnType = turn
}

func (game *Game) DiscardCard() error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateTurnAction(Discard); err != nil {
		return err
	}

	game.clearActiveCard(true)
	game.advanceTurn()
	return nil
}

func (game *Game) ReplaceCard(cardIndex int) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateTurnAction(Replace); err != nil {
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

func (game *Game) LookAtOwnCard(cardIndex int) (cards.Card, error) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateViewAction(LookAtSelf); err != nil {
		return cards.Card{}, err
	}

	if cardIndex < 0 || cardIndex >= len(game.playerCards[game.playerTurn]) {
		return cards.Card{}, errors.New("invalid card index")
	}

	card := game.playerCards[game.playerTurn][cardIndex]
	game.setRevealedCard(game.playerTurn, cardIndex, game.playerTurn, card)
	return card, nil
}

func (game *Game) LookAtOpponentCard(opponentIndex int, cardIndex int) (cards.Card, error) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateViewAction(LookAtOpponent); err != nil {
		return cards.Card{}, err
	}

	if opponentIndex < 0 || opponentIndex >= game.playerCount || opponentIndex == game.playerTurn {
		return cards.Card{}, errors.New("invalid opponent index")
	}

	if cardIndex < 0 || cardIndex >= len(game.playerCards[opponentIndex]) {
		return cards.Card{}, errors.New("invalid card index")
	}

	card := game.playerCards[opponentIndex][cardIndex]
	game.setRevealedCard(opponentIndex, cardIndex, game.playerTurn, card)
	return card, nil
}

func (game *Game) SwapCards(opponentIndex int, ownCardIndex int, opponentCardIndex int) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.validateViewAction(Swap); err != nil {
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

func (game *Game) RemoveCard(playerReplacing int, playersCardReplacing int, cardIndex int) error {
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
	game.gameState = RemovingCard
	game.replacing = &replaceInfo{
		playerReplacing:      playerReplacing,
		playersCardReplacing: playersCardReplacing,
		cardIndex:            cardIndex,
	}
	return nil
}

// EndTurn commits a look/swap action and advances to the next player.
func (game *Game) EndTurn() {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != TakingTurn {
		return
	}

	game.clearRevealedCard()
	game.clearActiveCard(true)
	game.advanceTurn()
}

func (game *Game) EndGame() {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.gameState != WaitingForTurn {
		return
	}

	game.playerCalledGame = game.playerTurn
	game.playerTurn = (game.playerTurn + 1) % game.playerCount
	game.gameState = WaitingForTurn
	game.turnType = Unselected
}

func (game *Game) GetGameStart() bool {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.gameState == GameStart
}

func (game *Game) GetActivePlayer() int {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.playerTurn
}

func (game *Game) GetTopDiscardCard() cards.Card {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.deck.GetTopDiscardCard()
}

func (game *Game) GetPlayerHand(playerIndex int) []cards.Card {
	game.mu.RLock()
	defer game.mu.RUnlock()
	if playerIndex < 0 || playerIndex >= game.playerCount {
		return nil
	}
	hand := make([]cards.Card, len(game.playerCards[playerIndex]))
	copy(hand, game.playerCards[playerIndex])
	return hand
}

func (game *Game) GetAllPlayerHands() [][]cards.Card {
	game.mu.RLock()
	defer game.mu.RUnlock()

	hands := make([][]cards.Card, game.playerCount)
	for i := 0; i < game.playerCount; i++ {
		hands[i] = make([]cards.Card, len(game.playerCards[i]))
		copy(hands[i], game.playerCards[i])
	}
	return hands
}

func (game *Game) GetActiveCard(playerIndex int) cards.Card {
	game.mu.RLock()
	defer game.mu.RUnlock()
	if playerIndex != game.playerTurn {
		return cards.Card{}
	}
	return game.activeCard
}

func (game *Game) GetRevealedCard(viewerIndex int) (int, int, cards.Card, bool, bool) {
	game.mu.RLock()
	defer game.mu.RUnlock()

	if game.revealed == nil {
		return -1, -1, cards.Card{}, false, false
	}

	canSeeFace := viewerIndex == game.revealed.viewerIndex
	return game.revealed.playerIndex, game.revealed.cardIndex, game.revealed.card, canSeeFace, true
}

func (game *Game) GetPlayerTurn() int {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.playerTurn
}

func (game *Game) GetGameState() GameState {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.gameState
}

func (game *Game) GetValidTurnTypes() []TurnType {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.validTurnTypes
}

func (game *Game) getValidTurnType() []TurnType {
	switch game.activeCard.Rank {
	case cards.Ace, cards.Two, cards.Three, cards.Four, cards.Five, cards.Six:
		return []TurnType{Discard, Replace}
	case cards.Seven, cards.Eight:
		return []TurnType{LookAtSelf, Discard, Replace}
	case cards.Nine, cards.Ten:
		return []TurnType{LookAtOpponent, Discard, Replace}
	case cards.Jack, cards.Queen, cards.King:
		if game.activeCard.Suit == cards.Hearts || game.activeCard.Suit == cards.Diamonds {
			return []TurnType{Discard, Replace}
		} else {
			return []TurnType{Swap, Discard, Replace}
		}
	default:
		return []TurnType{CallGame}
	}
}
