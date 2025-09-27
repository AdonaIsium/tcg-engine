package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdonaIsium/tcg-engine/internal/cards"
)

func TestStartTurn_RampsEnergy_And_DrawsWhenAllowed(t *testing.T) {
	// p0/p1 each have 6 cards; opening hand 3.
	d1 := smallDeck(6)
	d2 := smallDeck(6)

	opts := Options{
		StartingLife:     20,
		StartingHand:     3,
		MaxEnergy:        10,   // cap
		FirstPlayerDraws: true, // p0 draws at start of first turn
		Seed:             123,
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	p0 := g.Players[0]
	require.Equal(t, 3, len(p0.Hand))
	require.Equal(t, 3, len(g.Players[1].Hand))

	// Exercise
	g.StartTurn()

	// Turn bookkeeping
	assert.Equal(t, 1, g.Turn, "first StartTurn should set Turn=1")
	assert.Equal(t, 0, g.Active, "Active player should still be p0 until EndTurn")

	// Energy ramp + refill
	assert.Equal(t, 1, p0.MaxEnergy)
	assert.Equal(t, 1, p0.Energy)

	// Draw exactly 1 (since allowed)
	assert.Equal(t, 4, len(p0.Hand))
	assert.Equal(t, 6-3-1, len(p0.Deck)) // 6 total - 3 opening - 1 draw
}

func TestStartTurn_SkipDraw_WhenFirstPlayerDrawsFalse(t *testing.T) {
	d1 := smallDeck(6)
	d2 := smallDeck(6)

	opts := Options{
		StartingLife:     20,
		StartingHand:     3,
		MaxEnergy:        10,
		FirstPlayerDraws: false, // p0 skips first draw
		Seed:             999,
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	p0 := g.Players[0]
	require.Equal(t, 3, len(p0.Hand))

	g.StartTurn()

	// Energy ramp still happens
	assert.Equal(t, 1, p0.MaxEnergy)
	assert.Equal(t, 1, p0.Energy)

	// No draw on the very first turn for p0
	assert.Equal(t, 3, len(p0.Hand))
	assert.Equal(t, 6-3, len(p0.Deck))
}

func TestStartTurn_RefreshesCreatures(t *testing.T) {
	d1 := smallDeck(6)
	d2 := smallDeck(6)

	opts := Options{
		StartingLife:     20,
		StartingHand:     3,
		MaxEnergy:        10,
		FirstPlayerDraws: true,
		Seed:             42,
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	// Simulate a creature that was played on the previous turn (already on board)
	ci := CardInstance{
		InstanceID:    InstanceID("x#1"),
		Def:           &cards.CardDef{ID: "c_soldier", Name: "Soldier", Type: cards.TypeCreature, Attack: 2, Health: 2},
		CurrentAttack: 2,
		CurrentHealth: 2,
		SummoningSick: true, // should clear on StartTurn
		Exhausted:     true, // should reset on StartTurn
	}
	g.Players[0].Board = append(g.Players[0].Board, ci)

	g.StartTurn()

	got := g.Players[0].Board[0]
	assert.False(t, got.SummoningSick, "summoning sickness should clear at start of controller's turn")
	assert.False(t, got.Exhausted, "creatures should be readied at start of turn")
}
