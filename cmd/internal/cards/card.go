package cards

import "strings"

type Suit uint8

const (
	Spades Suit = iota + 1
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

func (card Card) IsRed() bool {
	return card.Suit == Hearts || card.Suit == Diamonds
}

func (card Card) String() string {
	return rankSymbols[card.Rank] + suitSymbols[card.Suit]
}

// RenderFace returns a 3-line card face:
// suit in the top-left, rank centered, and suit in the bottom-right.
func (card Card) RenderFace() string {
	const faceWidth = 3

	if card.Suit == 0 || card.Rank == 0 {
		return strings.Repeat(" ", faceWidth) + "\n" +
			strings.Repeat(" ", faceWidth) + "\n" +
			strings.Repeat(" ", faceWidth)
	}

	suit := suitSymbols[card.Suit]
	rank := rankSymbols[card.Rank]

	top := suit + strings.Repeat(" ", faceWidth-1)
	middle := centerText(faceWidth, rank)
	bottom := strings.Repeat(" ", faceWidth-1) + suit

	return strings.Join([]string{top, middle, bottom}, "\n")
}

func centerText(width int, text string) string {
	if len(text) >= width {
		return text
	}
	leftPad := (width - len(text)) / 2
	rightPad := width - len(text) - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}
