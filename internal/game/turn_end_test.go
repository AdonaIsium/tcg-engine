package game

import (
	"testing"

	"github.com/AdonaIsium/tcg-engine/internal/cards"
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

func TestEndTurn_DamageClearsAtEndOfTurn(t *testing.T) {
	// This test verifies that creature damage is cleared at the end of turn
	// following Magic: The Gathering rules where damage wears off during cleanup

	opts := Options{
		Seed:         42,
		StartingLife: 20,
		StartingHand: 0,
		MaxEnergy:    10,
	}

	g, err := NewGame("p1", "p2", smallDeck(10), smallDeck(10), opts)
	require.NoError(t, err)

	// Setup: Place a creature on each player's board
	p1Creature := CardInstance{
		InstanceID: "p1_creature#1",
		Def: &cards.CardDef{
			ID:     "dragon",
			Name:   "Dragon",
			Type:   cards.TypeCreature,
			Attack: 5,
			Health: 5,
		},
		Owner:         "p1",
		Controller:    "p1",
		CurrentAttack: 5,
		CurrentHealth: 5,
	}
	g.Players[0].Board = append(g.Players[0].Board, p1Creature)

	p2Creature := CardInstance{
		InstanceID: "p2_creature#1",
		Def: &cards.CardDef{
			ID:     "angel",
			Name:   "Angel",
			Type:   cards.TypeCreature,
			Attack: 4,
			Health: 4,
		},
		Owner:         "p2",
		Controller:    "p2",
		CurrentAttack: 4,
		CurrentHealth: 4,
	}
	g.Players[1].Board = append(g.Players[1].Board, p2Creature)

	// Apply damage to both creatures
	g.Players[0].Board[0].CurrentDamage = 3
	g.Players[0].Board[0].CurrentHealth = 2 // 5 health - 3 damage
	g.Players[1].Board[0].CurrentDamage = 2
	g.Players[1].Board[0].CurrentHealth = 2 // 4 health - 2 damage

	// Verify damage is applied
	assert.Equal(t, 3, g.Players[0].Board[0].CurrentDamage, "P1 creature should have 3 damage")
	assert.Equal(t, 2, g.Players[0].Board[0].CurrentHealth, "P1 creature should be at 2 health")
	assert.Equal(t, 2, g.Players[1].Board[0].CurrentDamage, "P2 creature should have 2 damage")
	assert.Equal(t, 2, g.Players[1].Board[0].CurrentHealth, "P2 creature should be at 2 health")

	// End turn - this should trigger cleanup
	g.EndTurn()

	// Both creatures should be fully healed
	assert.Equal(t, 0, g.Players[0].Board[0].CurrentDamage, "P1 creature damage should be cleared")
	assert.Equal(t, 5, g.Players[0].Board[0].CurrentHealth, "P1 creature should be back to full health")
	assert.Equal(t, 0, g.Players[1].Board[0].CurrentDamage, "P2 creature damage should be cleared")
	assert.Equal(t, 4, g.Players[1].Board[0].CurrentHealth, "P2 creature should be back to full health")
}

func TestEndTurn_TemporaryBuffsExpire(t *testing.T) {
	// This test verifies that temporary buffs (like Giant Growth) expire at end of turn
	// while the creature's base stats remain unchanged

	opts := Options{
		Seed:         42,
		StartingLife: 20,
		StartingHand: 0,
		MaxEnergy:    10,
	}

	g, err := NewGame("p1", "p2", smallDeck(10), smallDeck(10), opts)
	require.NoError(t, err)

	// Setup: Place creatures with temporary buffs
	buffedCreature := CardInstance{
		InstanceID:     "buffed#1",
		Def: &cards.CardDef{
			ID:     "soldier",
			Name:   "Soldier",
			Type:   cards.TypeCreature,
			Attack: 2,
			Health: 2,
		},
		Owner:          "p1",
		Controller:     "p1",
		TempAttackBuff: 3, // Giant Growth effect
		TempHealthBuff: 3,
		CurrentAttack:  5, // 2 base + 3 temp
		CurrentHealth:  5, // 2 base + 3 temp
	}
	g.Players[0].Board = append(g.Players[0].Board, buffedCreature)

	// Another creature with different buffs
	anotherBuffed := CardInstance{
		InstanceID:     "buffed#2",
		Def: &cards.CardDef{
			ID:     "scout",
			Name:   "Scout",
			Type:   cards.TypeCreature,
			Attack: 1,
			Health: 1,
		},
		Owner:          "p2",
		Controller:     "p2",
		TempAttackBuff: 5, // Massive temporary boost
		TempHealthBuff: 2,
		CurrentAttack:  6, // 1 base + 5 temp
		CurrentHealth:  3, // 1 base + 2 temp
	}
	g.Players[1].Board = append(g.Players[1].Board, anotherBuffed)

	// Verify buffs are active
	assert.Equal(t, 3, g.Players[0].Board[0].TempAttackBuff)
	assert.Equal(t, 3, g.Players[0].Board[0].TempHealthBuff)
	assert.Equal(t, 5, g.Players[0].Board[0].CurrentAttack)
	assert.Equal(t, 5, g.Players[0].Board[0].CurrentHealth)

	// End turn - temporary buffs should expire
	g.EndTurn()

	// Verify temp buffs are gone and stats are back to base
	assert.Equal(t, 0, g.Players[0].Board[0].TempAttackBuff, "Temp attack buff should be cleared")
	assert.Equal(t, 0, g.Players[0].Board[0].TempHealthBuff, "Temp health buff should be cleared")
	assert.Equal(t, 2, g.Players[0].Board[0].CurrentAttack, "Attack should return to base value")
	assert.Equal(t, 2, g.Players[0].Board[0].CurrentHealth, "Health should return to base value")

	assert.Equal(t, 0, g.Players[1].Board[0].TempAttackBuff, "P2 temp attack buff should be cleared")
	assert.Equal(t, 0, g.Players[1].Board[0].TempHealthBuff, "P2 temp health buff should be cleared")
	assert.Equal(t, 1, g.Players[1].Board[0].CurrentAttack, "P2 attack should return to base")
	assert.Equal(t, 1, g.Players[1].Board[0].CurrentHealth, "P2 health should return to base")
}

func TestEndTurn_PermanentBuffsPersist(t *testing.T) {
	// This test verifies that permanent buffs (like +1/+1 counters) persist through cleanup
	// even when damage is cleared

	opts := Options{
		Seed:         42,
		StartingLife: 20,
		StartingHand: 0,
		MaxEnergy:    10,
	}

	g, err := NewGame("p1", "p2", smallDeck(10), smallDeck(10), opts)
	require.NoError(t, err)

	// Setup: Creature with permanent buffs and damage
	permBuffedCreature := CardInstance{
		InstanceID:     "permbuffed#1",
		Def: &cards.CardDef{
			ID:     "veteran",
			Name:   "Veteran",
			Type:   cards.TypeCreature,
			Attack: 2,
			Health: 3,
		},
		Owner:          "p1",
		Controller:     "p1",
		PermAttackBuff: 2, // Permanent +2/+2
		PermHealthBuff: 2,
		CurrentDamage:  3, // Took 3 damage
		CurrentAttack:  4, // 2 base + 2 perm
		CurrentHealth:  2, // 3 base + 2 perm - 3 damage = 2
	}
	g.Players[0].Board = append(g.Players[0].Board, permBuffedCreature)

	// Verify initial state
	assert.Equal(t, 2, g.Players[0].Board[0].PermAttackBuff, "Should have +2 permanent attack")
	assert.Equal(t, 2, g.Players[0].Board[0].PermHealthBuff, "Should have +2 permanent health")
	assert.Equal(t, 3, g.Players[0].Board[0].CurrentDamage, "Should have 3 damage")
	assert.Equal(t, 2, g.Players[0].Board[0].CurrentHealth, "Should be at 2 health")

	// End turn - damage clears but permanent buffs remain
	g.EndTurn()

	// Permanent buffs should persist
	assert.Equal(t, 2, g.Players[0].Board[0].PermAttackBuff, "Permanent attack buff should persist")
	assert.Equal(t, 2, g.Players[0].Board[0].PermHealthBuff, "Permanent health buff should persist")
	assert.Equal(t, 0, g.Players[0].Board[0].CurrentDamage, "Damage should be cleared")

	// Stats should be base + permanent (no damage)
	assert.Equal(t, 4, g.Players[0].Board[0].CurrentAttack, "Attack should be base + perm")
	assert.Equal(t, 5, g.Players[0].Board[0].CurrentHealth, "Health should be base + perm (damage cleared)")
}

func TestEndTurn_ComplexCreatureState(t *testing.T) {
	// This test verifies cleanup with complex state: permanent buffs, temporary buffs, and damage
	// This simulates a creature with +1/+1 counters, Giant Growth, and combat damage

	opts := Options{
		Seed:         42,
		StartingLife: 20,
		StartingHand: 0,
		MaxEnergy:    10,
	}

	g, err := NewGame("p1", "p2", smallDeck(10), smallDeck(10), opts)
	require.NoError(t, err)

	// Setup: Complex creature state
	complexCreature := CardInstance{
		InstanceID:     "complex#1",
		Def: &cards.CardDef{
			ID:     "champion",
			Name:   "Champion",
			Type:   cards.TypeCreature,
			Attack: 3,
			Health: 3,
		},
		Owner:          "p1",
		Controller:     "p1",
		PermAttackBuff: 1,  // +1/+1 counter
		PermHealthBuff: 1,
		TempAttackBuff: 3,  // Giant Growth
		TempHealthBuff: 3,
		CurrentDamage:  5,  // Took 5 damage in combat
		CurrentAttack:  7,  // 3 base + 1 perm + 3 temp = 7
		CurrentHealth:  2,  // 3 base + 1 perm + 3 temp - 5 damage = 2
	}
	g.Players[0].Board = append(g.Players[0].Board, complexCreature)

	// Another complex creature on opponent's side
	opponentComplex := CardInstance{
		InstanceID:     "complex#2",
		Def: &cards.CardDef{
			ID:     "knight",
			Name:   "Knight",
			Type:   cards.TypeCreature,
			Attack: 4,
			Health: 5,
		},
		Owner:          "p2",
		Controller:     "p2",
		PermAttackBuff: 2,  // Multiple +1/+1 counters
		PermHealthBuff: 2,
		TempAttackBuff: 1,  // Small temp boost
		TempHealthBuff: 1,
		CurrentDamage:  3,  // Some damage
		CurrentAttack:  7,  // 4 base + 2 perm + 1 temp = 7
		CurrentHealth:  5,  // 5 base + 2 perm + 1 temp - 3 damage = 5
	}
	g.Players[1].Board = append(g.Players[1].Board, opponentComplex)

	// Verify complex initial state
	assert.Equal(t, 7, g.Players[0].Board[0].CurrentAttack)
	assert.Equal(t, 2, g.Players[0].Board[0].CurrentHealth)
	assert.Equal(t, 7, g.Players[1].Board[0].CurrentAttack)
	assert.Equal(t, 5, g.Players[1].Board[0].CurrentHealth)

	// End turn - cleanup should handle everything correctly
	g.EndTurn()

	// P1 creature: Should keep permanent buffs, lose temp buffs and damage
	p1Creature := g.Players[0].Board[0]
	assert.Equal(t, 1, p1Creature.PermAttackBuff, "Permanent attack buff should persist")
	assert.Equal(t, 1, p1Creature.PermHealthBuff, "Permanent health buff should persist")
	assert.Equal(t, 0, p1Creature.TempAttackBuff, "Temp attack buff should be cleared")
	assert.Equal(t, 0, p1Creature.TempHealthBuff, "Temp health buff should be cleared")
	assert.Equal(t, 0, p1Creature.CurrentDamage, "Damage should be cleared")
	assert.Equal(t, 4, p1Creature.CurrentAttack, "Attack should be 3 base + 1 perm")
	assert.Equal(t, 4, p1Creature.CurrentHealth, "Health should be 3 base + 1 perm")

	// P2 creature: Same logic
	p2Creature := g.Players[1].Board[0]
	assert.Equal(t, 2, p2Creature.PermAttackBuff, "P2 permanent attack should persist")
	assert.Equal(t, 2, p2Creature.PermHealthBuff, "P2 permanent health should persist")
	assert.Equal(t, 0, p2Creature.TempAttackBuff, "P2 temp attack should be cleared")
	assert.Equal(t, 0, p2Creature.TempHealthBuff, "P2 temp health should be cleared")
	assert.Equal(t, 0, p2Creature.CurrentDamage, "P2 damage should be cleared")
	assert.Equal(t, 6, p2Creature.CurrentAttack, "P2 attack should be 4 base + 2 perm")
	assert.Equal(t, 7, p2Creature.CurrentHealth, "P2 health should be 5 base + 2 perm")
}

func TestEndTurn_NoCreaturesNoErrors(t *testing.T) {
	// This test verifies that cleanup handles empty boards gracefully

	opts := Options{
		Seed:         42,
		StartingLife: 20,
		StartingHand: 0,
		MaxEnergy:    10,
	}

	g, err := NewGame("p1", "p2", smallDeck(10), smallDeck(10), opts)
	require.NoError(t, err)

	// Both boards are empty
	assert.Empty(t, g.Players[0].Board)
	assert.Empty(t, g.Players[1].Board)

	// Should not panic or error
	g.EndTurn()

	// Game should continue normally
	assert.Equal(t, 1, g.Active, "Should switch to player 2")
}

func TestEndTurn_OnlyUndamagedCreatures(t *testing.T) {
	// This test verifies that creatures without damage or buffs are not unnecessarily processed
	// and that the optimization in refreshCreatureHealth works correctly

	opts := Options{
		Seed:         42,
		StartingLife: 20,
		StartingHand: 0,
		MaxEnergy:    10,
	}

	g, err := NewGame("p1", "p2", smallDeck(10), smallDeck(10), opts)
	require.NoError(t, err)

	// Setup: Healthy creatures with no modifications
	healthyCreature := CardInstance{
		InstanceID:    "healthy#1",
		Def: &cards.CardDef{
			ID:     "golem",
			Name:   "Golem",
			Type:   cards.TypeCreature,
			Attack: 4,
			Health: 6,
		},
		Owner:         "p1",
		Controller:    "p1",
		CurrentAttack: 4,
		CurrentHealth: 6,
	}
	g.Players[0].Board = append(g.Players[0].Board, healthyCreature)

	// Record log length before end turn
	logLenBefore := len(g.Log)

	// End turn
	g.EndTurn()

	// Creature should remain unchanged
	assert.Equal(t, 4, g.Players[0].Board[0].CurrentAttack)
	assert.Equal(t, 6, g.Players[0].Board[0].CurrentHealth)
	assert.Equal(t, 0, g.Players[0].Board[0].CurrentDamage)
	assert.Equal(t, 0, g.Players[0].Board[0].TempAttackBuff)
	assert.Equal(t, 0, g.Players[0].Board[0].TempHealthBuff)

	// Should only have the "end turn" log, no refresh logs (due to optimization)
	logsAdded := len(g.Log) - logLenBefore
	assert.Equal(t, 1, logsAdded, "Should only log end turn, not refresh for unchanged creatures")
}
