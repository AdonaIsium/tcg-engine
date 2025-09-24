package game

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdonaIsium/tcg-engine/internal/cards"
)

// Helper functions
func makeDeck(n int) []cards.CardDef {
	deck := make([]cards.CardDef, 0, n)
	for i := 0; i < n; i++ {
		if i%3 == 2 {
			deck = append(deck, cards.CardDef{
				ID:   "s_firebolt_" + strconv.Itoa(i),
				Name: "Fire Bolt " + strconv.Itoa(i),
				Type: cards.TypeSpell,
				Cost: 1,
				Text: "Deal 2 damage to any target.",
				SpellEffect: &cards.Effect{
					Kind:   cards.EffectDamage,
					Amount: 2,
					Target: cards.TargetAnyCreature,
				},
			})
		} else {
			deck = append(deck, cards.CardDef{
				ID:     "c_unit_" + strconv.Itoa(i),
				Name:   "Unit " + strconv.Itoa(i),
				Type:   cards.TypeCreature,
				Cost:   2,
				Attack: 2 + (i % 3),
				Health: 2 + ((i + 1) % 3),
				Text:   "A basic creature.",
			})
		}
	}
	return deck
}

func collectIDs(insts []CardInstance) []string {
	out := make([]string, len(insts))
	for i := range insts {
		out[i] = string(insts[i].InstanceID)
	}
	return out
}

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
