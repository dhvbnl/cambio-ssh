package cards

type Suit uint8

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

var suitStrings = map[Suit]string{
	Spades:   "Spades",
	Hearts:   "Hearts",
	Diamonds: "Diamonds",
	Clubs:    "Clubs",
}

var suitSymbols = map[Suit]string{
	Spades:   "♠",
	Hearts:   "♥",
	Diamonds: "♦",
	Clubs:    "♣",
}

func (suit Suit) String() string {
	return suitStrings[suit]
}

type Rank uint8

const (
	Ace Rank = iota + 1
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

var rankStrings = map[Rank]string{
	Ace:   "Ace",
	Two:   "Two",
	Three: "Three",
	Four:  "Four",
	Five:  "Five",
	Six:   "Six",
	Seven: "Seven",
	Eight: "Eight",
	Nine:  "Nine",
	Ten:   "Ten",
	Jack:  "Jack",
	Queen: "Queen",
	King:  "King",
}

var rankSymbols = map[Rank]string{
	Ace:   "A",
	Two:   "2",
	Three: "3",
	Four:  "4",
	Five:  "5",
	Six:   "6",
	Seven: "7",
	Eight: "8",
	Nine:  "9",
	Ten:   "10",
	Jack:  "J",
	Queen: "Q",
	King:  "K",
}

func (rank Rank) String() string {
	return rankStrings[rank]
}

type Card struct {
	Suit Suit
	Rank Rank
}

func (card Card) String() string {
	return rankSymbols[card.Rank] + suitSymbols[card.Suit]
}
