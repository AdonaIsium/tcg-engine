package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndTurn_FlipsActiveOnly(t *testing.T) {
	opts := Options{
		Seed:         123,
		StartingLife: 20,
		StartingHand: 3,
		MaxEnergy:    10,
	}
	p1ID, p2ID := "p1", "p2"
	d1 := smallDeck(8)
	d2 := smallDeck(8)

	g, err := NewGame(p1ID, p2ID, d1, d2, opts)
	require.NoError(t, err)

	startActive := g.Active
	startTurn := g.Turn
	startLogLen := len(g.Log)
	p0Snap := clonePlayerState(g.Players[0])
	p1Snap := clonePlayerState(g.Players[1])

	// Act
	g.EndTurn()

	// Active should flip
	assert.NotEqual(t, startActive, g.Active)
	assert.Equal(t, 1-startActive, g.Active)

	// Turn should NOT change (StartTurn handles incrementing)
	assert.Equal(t, startTurn, g.Turn)

	// Players unchanged
	assert.Equal(t, p0Snap, clonePlayerState(g.Players[0]))
	assert.Equal(t, p1Snap, clonePlayerState(g.Players[1]))

	// Log grew by 1
	assert.Equal(t, startLogLen+1, len(g.Log))
}

func TestEndTurn_MultipleCallsAlternateActive(t *testing.T) {
	opts := Options{Seed: 999, StartingLife: 20, StartingHand: 3, MaxEnergy: 10}
	p1ID, p2ID := "p1", "p2"
	d1 := smallDeck(6)
	d2 := smallDeck(6)

	g, err := NewGame(p1ID, p2ID, d1, d2, opts)
	require.NoError(t, err)

	baseTurn := g.Turn
	baseP0 := clonePlayerState(g.Players[0])
	baseP1 := clonePlayerState(g.Players[1])
	baseLog := len(g.Log)

	want := 1

	for range 4 {
		g.EndTurn()
		assert.Equal(t, want, g.Active, "active should alternate")
		assert.Equal(t, baseTurn, g.Turn, "turn counter should not change")
		assert.Equal(t, baseP0, clonePlayerState(g.Players[0]))
		assert.Equal(t, baseP1, clonePlayerState(g.Players[1]))
		want = 1 - want
	}

	assert.Equal(t, baseLog+4, len(g.Log))
}
