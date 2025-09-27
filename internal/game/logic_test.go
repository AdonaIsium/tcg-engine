package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdonaIsium/tcg-engine/internal/cards"
)

// Tests
func TestNewGame_Basics(t *testing.T) {
	d1 := makeDeck(10)
	d2 := makeDeck(10)

	opts := Options{
		StartingLife: 20,
		StartingHand: 3,
		MaxEnergy:    10,
		Seed:         420,
	}

	g, err := NewGame("p1", "p2", d1, d2, opts)
	require.NoError(t, err)

	require.NotNil(t, g.Players[0])
	require.NotNil(t, g.Players[1])

	p0, p1 := g.Players[0], g.Players[1]

	assert.Equal(t, 20, p0.Life)
	assert.Equal(t, 20, p1.Life)

	// NOTE: these are looking for the *player* energy/max energy, not the *game* energy/max energy
	assert.Zero(t, p0.Energy)
	assert.Zero(t, p0.MaxEnergy)
	assert.Zero(t, p1.Energy)
	assert.Zero(t, p1.MaxEnergy)

	assert.Equal(t, 0, g.Active)
	assert.Equal(t, 0, g.Turn)

	assert.Len(t, p0.Hand, opts.StartingHand)
	assert.Len(t, p1.Hand, opts.StartingHand)
	assert.Len(t, p0.Deck, 10-opts.StartingHand)
	assert.Len(t, p1.Deck, 10-opts.StartingHand)

	for _, c := range p0.Hand {
		assert.False(t, c.SummoningSick, "hand card should not be summoning sick: %v", c.InstanceID)
	}
	for _, c := range p1.Hand {
		assert.False(t, c.SummoningSick, "hand card should not be summoning sick: %v", c.InstanceID)
	}

	checkRuntime := func(insts []CardInstance) {
		for _, ci := range insts {
			if ci.Def.Type == cards.TypeCreature {
				assert.Equal(t, ci.Def.Attack, ci.CurrentAttack, "atk mismatch for %s", ci.Def.ID)
				assert.Equal(t, ci.Def.Health, ci.CurrentHealth, "hp mismatch for %s", ci.Def.ID)
			} else {
				assert.Zero(t, ci.CurrentAttack, "spell atk should be 0 for %s", ci.Def.ID)
				assert.Zero(t, ci.CurrentHealth, "spell hp should be 0 for %s", ci.Def.ID)
			}
		}
	}

	checkRuntime(p0.Deck)
	checkRuntime(p1.Deck)
	checkRuntime(p0.Hand)
	checkRuntime(p1.Hand)

	assert.GreaterOrEqual(t, len(g.Log), 3)
}

func TestNewGame_SeedDeterminism(t *testing.T) {
	d1 := makeDeck(20)
	d2 := makeDeck(20)

	opts := Options{
		StartingLife: 20,
		StartingHand: 5,
		MaxEnergy:    10,
		Seed:         999,
	}

	g1, err := NewGame("p1", "p2", d1, d2, opts)
	require.NoError(t, err)
	g2, err := NewGame("p1", "p2", d1, d2, opts)
	require.NoError(t, err)

	assert.Equal(t, collectIDs(g1.Players[0].Hand), collectIDs(g2.Players[0].Hand))
	assert.Equal(t, collectIDs(g1.Players[0].Hand), collectIDs(g2.Players[0].Hand))
	assert.Equal(t, collectIDs(g1.Players[0].Deck), collectIDs(g2.Players[0].Deck))
	assert.Equal(t, collectIDs(g1.Players[0].Deck), collectIDs(g2.Players[0].Deck))

	assert.Equal(t, g1.ID, g2.ID)
}

func TestNewGame_DifferentSeedsDifferentShuffle(t *testing.T) {
	d1 := makeDeck(24)
	d2 := makeDeck(24)

	g1, err := NewGame("p1", "p2", d1, d2, Options{StartingLife: 20, StartingHand: 5, MaxEnergy: 10, Seed: 1})
	require.NoError(t, err)
	g2, err := NewGame("p1", "p2", d1, d2, Options{StartingLife: 20, StartingHand: 5, MaxEnergy: 10, Seed: 2})
	require.NoError(t, err)

	assert.NotEqual(t, collectIDs(g1.Players[0].Deck), collectIDs(g2.Players[1].Deck))
}

func TestNewGame_InvalidInputs(t *testing.T) {
	deck := makeDeck(5)

	_, err := NewGame("", "p2", deck, deck, Options{})
	require.Error(t, err)

	_, err = NewGame("p1", "", deck, deck, Options{})
	require.Error(t, err)

	_, err = NewGame("p1", "p2", nil, deck, Options{})
	require.Error(t, err)

	_, err = NewGame("p1", "p2", deck, nil, Options{})
	require.Error(t, err)

	g, err := NewGame("p1", "p2", makeDeck(6), makeDeck(6), Options{StartingHand: -5, Seed: 77})
	require.NoError(t, err)
	assert.Len(t, g.Players[0].Hand, 3)
	assert.Len(t, g.Players[1].Hand, 3)
}

func TestNewGame_InstanceIDUniqueness(t *testing.T) {
	d1 := makeDeck(30)
	d2 := makeDeck(30)

	g, err := NewGame("p1", "p2", d1, d2, Options{StartingHand: 5, Seed: 123})
	require.NoError(t, err)

	p0, p1 := g.Players[0], g.Players[1]

	all := make([]string, 0, len(p0.Deck)+len(p0.Hand)+len(p1.Deck)+len(p1.Hand))
	all = append(all, collectIDs(p0.Deck)...)
	all = append(all, collectIDs(p0.Hand)...)
	all = append(all, collectIDs(p1.Deck)...)
	all = append(all, collectIDs(p1.Hand)...)

	seen := make(map[string]struct{}, len(all))
	for _, id := range all {
		_, exists := seen[id]
		assert.False(t, exists, "duplicate InstanceID: %s", id)
		seen[id] = struct{}{}
		assert.NotEmpty(t, id)
		assert.Contains(t, id, "#", "InstanceID should contain '#': %s", id)
	}
}
