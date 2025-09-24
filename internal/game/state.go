package game

import (
	"github.com/AdonaIsium/tcg-engine/internal/cards"
)

type Zone string

const (
	ZoneDeck      Zone = "deck"
	ZoneHand      Zone = "hand"
	ZoneBoard     Zone = "board"
	ZoneGraveyard Zone = "graveyard"
)

type InstanceID string

type Options struct {
	StartingLife     int
	StartingHand     int
	MaxEnergy        int
	FirstPlayerDraws bool
	Seed             int64
}

type CardInstance struct {
	InstanceID InstanceID
	Def        *cards.CardDef

	CurrentAttack int
	CurrentHealth int
	SummoningSick bool
	Exhausted     bool
}

type PlayerState struct {
	PlayerID  string
	Life      int
	Deck      []CardInstance
	Hand      []CardInstance
	Board     []CardInstance
	Graveyard []CardInstance

	Energy    int
	MaxEnergy int
}

type Event struct {
	Turn   int
	Player string
	Type   string
	Msg    string
}

type Game struct {
	ID      string
	Players [2]*PlayerState
	Active  int
	Turn    int
	Options Options
	Rand    randSource
	Log     []Event
}

type randSource interface {
	Intn(n int) int
}
