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

type CombatPhase string

const (
	PhaseNone      CombatPhase = "none"
	PhaseAttackers CombatPhase = "declare_attackers"
	PhaseBlockers  CombatPhase = "declare_blockers"
	PhaseDamage    CombatPhase = "resolve_damage"
)

type InstanceID string

type Options struct {
	StartingLife     int
	StartingHand     int
	MaxEnergy        int
	MaxBoardSize     int
	FirstPlayerDraws bool
	Seed             int64
}

type CardInstance struct {
	InstanceID InstanceID
	Def        *cards.CardDef
	Owner      string
	Controller string

	// Permanent buffs applied to attack/health
	PermAttackBuff int
	PermHealthBuff int

	// Temporary effects to attack/health
	TempAttackBuff int
	TempHealthBuff int
	CurrentDamage  int

	// Current calculated values for attack/health
	CurrentAttack int
	CurrentHealth int

	SummoningSick bool
	Exhausted     bool
}

type PlayerState struct {
	PlayerID  string
	Name      string
	Life      int
	Deck      []CardInstance
	Hand      []CardInstance
	Board     []CardInstance
	Graveyard []CardInstance

	CurrentEnergy int
	MaxEnergy     int
}

type Event struct {
	Turn   int
	Player string
	Type   string
	Msg    string
}

type Game struct {
	ID        string
	Players   [2]*PlayerState
	Active    int
	Turn      int
	Options   Options
	Rand      randSource
	Log       []Event
	GameEnded bool
	Winner    string

	// Combat state tracking
	CombatPhase   CombatPhase
	AttackingIDs  []InstanceID
	BlockingPairs map[InstanceID]InstanceID // attacker -> blocker
}

type randSource interface {
	Intn(n int) int
}
