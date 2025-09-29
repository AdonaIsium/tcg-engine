package game

import (
	"fmt"
	"testing"

	"github.com/AdonaIsium/tcg-engine/internal/cards"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanPlayCard_Legality(t *testing.T) {
	// Baseline setup: two players with simple decks
	d1 := smallDeck(5)
	d2 := smallDeck(5)

	opts := Options{
		StartingLife:     20,
		StartingHand:     1,
		MaxEnergy:        10,
		FirstPlayerDraws: false,
		Seed:             123,
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	// Give player 0 plenty of energy
	p0 := g.Players[0]
	p0.MaxEnergy, p0.CurrentEnergy = 10, 10

	// Dummy instance to simulate a creature on the board
	creature := CardInstance{
		InstanceID:    "c_dummy#1",
		Def:           &cards.CardDef{ID: "c_dummy", Type: cards.TypeCreature, Attack: 1, Health: 1},
		CurrentAttack: 1,
		CurrentHealth: 1,
	}
	p0.Board = append(p0.Board, creature)

	opp := g.Players[1]
	opp.Board = append(opp.Board, CardInstance{
		InstanceID:    "c_enemy#1",
		Def:           &cards.CardDef{ID: "c_enemy", Type: cards.TypeCreature, Attack: 2, Health: 2},
		CurrentAttack: 2,
		CurrentHealth: 2,
	})

	// Cases
	cases := []struct {
		name     string
		playerID string
		handIdx  int
		card     cards.CardDef
		targets  []*TargetRef
		wantErr  error
	}{
		{
			name:     "legal creature play",
			playerID: "p0",
			handIdx:  0,
			card:     cards.CardDef{ID: "c1", Type: cards.TypeCreature, Cost: 1},
			targets:  nil, // Creatures don't need targets
			wantErr:  nil,
		},
		{
			name:     "wrong player turn",
			playerID: "p1",
			handIdx:  0,
			card:     cards.CardDef{ID: "c2", Type: cards.TypeCreature, Cost: 1},
			targets:  nil,
			wantErr:  ErrNotYourTurn,
		},
		{
			name:     "hand index out of range",
			playerID: "p0",
			handIdx:  99,
			card:     cards.CardDef{ID: "c3", Type: cards.TypeCreature, Cost: 1},
			targets:  nil,
			wantErr:  ErrInvalidHandIndex,
		},
		{
			name:     "not enough energy",
			playerID: "p0",
			handIdx:  0,
			card:     cards.CardDef{ID: "c4", Type: cards.TypeCreature, Cost: 20},
			targets:  nil,
			wantErr:  ErrNotEnoughEnergy,
		},
		{
			name:     "spell missing required target",
			playerID: "p0",
			handIdx:  0,
			card: cards.CardDef{
				ID: "s1", Type: cards.TypeSpell, Cost: 1,
				Effects: []cards.Effect{
					{Kind: cards.EffectDamage, Amount: 2, Target: cards.TargetAnyCreature},
				},
			},
			targets: []*TargetRef{nil}, // Missing required target
			wantErr: ErrMissingTarget,
		},
		{
			name:     "spell bad creature id",
			playerID: "p0",
			handIdx:  0,
			card: cards.CardDef{
				ID: "s2", Type: cards.TypeSpell, Cost: 1,
				Effects: []cards.Effect{
					{Kind: cards.EffectDamage, Amount: 2, Target: cards.TargetAnyCreature},
				},
			},
			targets: []*TargetRef{{InstanceID: ptrInstance("bogus#99")}},
			wantErr: ErrInvalidTarget,
		},
		{
			name:     "spell valid ally creature target",
			playerID: "p0",
			handIdx:  0,
			card: cards.CardDef{
				ID: "s3", Type: cards.TypeSpell, Cost: 1,
				Effects: []cards.Effect{
					{Kind: cards.EffectDamage, Amount: 2, Target: cards.TargetAllyCreature},
				},
			},
			targets: []*TargetRef{{InstanceID: &p0.Board[0].InstanceID}},
			wantErr: nil,
		},
		{
			name:     "spell valid enemy creature target",
			playerID: "p0",
			handIdx:  0,
			card: cards.CardDef{
				ID: "s4", Type: cards.TypeSpell, Cost: 1,
				Effects: []cards.Effect{
					{Kind: cards.EffectDamage, Amount: 3, Target: cards.TargetEnemyCreature},
				},
			},
			targets: []*TargetRef{{InstanceID: &opp.Board[0].InstanceID}},
			wantErr: nil,
		},
		{
			name:     "spell auto-target self player",
			playerID: "p0",
			handIdx:  0,
			card: cards.CardDef{
				ID: "s5", Type: cards.TypeSpell, Cost: 1,
				Effects: []cards.Effect{
					{Kind: cards.EffectDrawCards, Amount: 1, Target: cards.TargetSelfPlayer},
				},
			},
			targets: []*TargetRef{nil}, // Auto-populated
			wantErr: nil,
		},
		{
			name:     "spell auto-target enemy player",
			playerID: "p0",
			handIdx:  0,
			card: cards.CardDef{
				ID: "s6", Type: cards.TypeSpell, Cost: 1,
				Effects: []cards.Effect{
					{Kind: cards.EffectDamage, Amount: 2, Target: cards.TargetEnemyPlayer},
				},
			},
			targets: []*TargetRef{nil}, // Auto-populated
			wantErr: nil,
		},
		{
			name:     "multi-effect spell: damage + draw",
			playerID: "p0",
			handIdx:  0,
			card: cards.CardDef{
				ID: "s_multi", Type: cards.TypeSpell, Cost: 2,
				Text: "Deal 1 damage to target creature, then draw 1 card",
				Effects: []cards.Effect{
					{Kind: cards.EffectDamage, Amount: 1, Target: cards.TargetAnyCreature},
					{Kind: cards.EffectDrawCards, Amount: 1, Target: cards.TargetSelfPlayer},
				},
			},
			targets: []*TargetRef{
				{InstanceID: &p0.Board[0].InstanceID}, // Target for damage
				nil,                                   // Auto-populated for draw
			},
			wantErr: nil,
		},
		{
			name:     "multi-effect spell: wrong target count",
			playerID: "p0",
			handIdx:  0,
			card: cards.CardDef{
				ID: "s_multi2", Type: cards.TypeSpell, Cost: 2,
				Effects: []cards.Effect{
					{Kind: cards.EffectDamage, Amount: 1, Target: cards.TargetAnyCreature},
					{Kind: cards.EffectHeal, Amount: 2, Target: cards.TargetSelfPlayer},
				},
			},
			targets: []*TargetRef{{InstanceID: &p0.Board[0].InstanceID}}, // Missing second target
			wantErr: fmt.Errorf("expected 2 targets, got 1"),
		},
		{
			name:     "buff spell: +2/+2 to ally",
			playerID: "p0",
			handIdx:  0,
			card: cards.CardDef{
				ID: "s_buff", Type: cards.TypeSpell, Cost: 1,
				Text: "Give target creature +2/+2",
				Effects: []cards.Effect{
					{Kind: cards.EffectBuffStatsPerm, BuffAttack: 2, BuffHealth: 2, Target: cards.TargetAllyCreature},
				},
			},
			targets: []*TargetRef{{InstanceID: &p0.Board[0].InstanceID}},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset p0 hand to contain just this card
			p0.Hand = []CardInstance{
				{InstanceID: "h#1", Def: &tc.card},
			}

			err := g.CanPlayCard(tc.playerID, tc.handIdx, tc.targets)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				if tc.name == "multi-effect spell: wrong target count" {
					// Special case: error message comparison for target count
					assert.Contains(t, err.Error(), "expected 2 targets, got 1")
				} else {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

// TestPlayCard_ExecutionEffects tests that playing cards actually modifies game state correctly
// This is different from TestCanPlayCard_Legality which only tests validation
func TestPlayCard_DamageEffect(t *testing.T) {
	// === SETUP PHASE ===
	// Create a minimal game with controlled conditions
	d1 := smallDeck(10) // Player 1's deck
	d2 := smallDeck(10) // Player 2's deck

	opts := Options{
		StartingLife:     20,
		StartingHand:     0, // Start with empty hands for control
		MaxEnergy:        10,
		MaxBoardSize:     7,
		FirstPlayerDraws: false,
		Seed:             42, // Deterministic for reproducibility
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	// Get references to both players
	activePlayer := g.Players[0] // p0 is active
	opponent := g.Players[1]     // p1 is opponent

	// === GAME STATE SETUP ===
	// Give active player enough energy to cast spells
	activePlayer.MaxEnergy = 10
	activePlayer.CurrentEnergy = 10

	// Place a damage spell in active player's hand
	damageSpell := CardInstance{
		InstanceID: "spell#1",
		Def: &cards.CardDef{
			ID:   "lightning_bolt",
			Name: "Lightning Bolt",
			Type: cards.TypeSpell,
			Cost: 1,
			Text: "Deal 3 damage to any target",
			Effects: []cards.Effect{
				{
					Kind:   cards.EffectDamage,
					Amount: 3,
					Target: cards.TargetAnyCreature,
				},
			},
		},
		Owner:      activePlayer.PlayerID,
		Controller: activePlayer.PlayerID,
	}
	activePlayer.Hand = append(activePlayer.Hand, damageSpell)

	// Place a target creature on opponent's board
	targetCreature := CardInstance{
		InstanceID: "creature#1",
		Def: &cards.CardDef{
			ID:     "goblin",
			Name:   "Goblin",
			Type:   cards.TypeCreature,
			Cost:   1,
			Attack: 2,
			Health: 2, // Will die from 3 damage
		},
		Owner:         opponent.PlayerID,
		Controller:    opponent.PlayerID,
		CurrentAttack: 2,
		CurrentHealth: 2, // Starting health - will track changes
		SummoningSick: false,
		Exhausted:     false,
	}
	opponent.Board = append(opponent.Board, targetCreature)

	// === RECORD INITIAL STATE ===
	initialCreatureHealth := targetCreature.CurrentHealth
	initialOppBoardSize := len(opponent.Board)
	initialOppGraveyardSize := len(opponent.Graveyard)
	initialPlayerEnergy := activePlayer.CurrentEnergy

	// === EXECUTE THE PLAY ===
	// Create target reference for the creature
	targets := []*TargetRef{
		{InstanceID: &targetCreature.InstanceID},
	}

	// TODO: Call PlayCard with the spell
	err = g.PlayCard(activePlayer.PlayerID, 0, targets)
	require.NoError(t, err, "PlayCard should succeed")

	// === ASSERTIONS PHASE ===
	assert.Equal(t, initialPlayerEnergy-damageSpell.Def.Cost, activePlayer.CurrentEnergy, "Energy should be reduced by spell cost")
	assert.Equal(t, initialCreatureHealth-damageSpell.Def.Effects[0].Amount, opponent.Graveyard[0].CurrentHealth, "Creature health should be -1 (overkilled by 1)")

	assert.Len(t, activePlayer.Hand, 0, "Spell should be removed from hand")
	assert.Len(t, activePlayer.Graveyard, 1, "Spell should be in graveyard")

	assert.Len(t, opponent.Board, 0, "Dead creature should be removed from board")
	assert.Less(t, len(opponent.Board), initialOppBoardSize, "Dead creature should no longer be on the board")
	assert.Len(t, opponent.Graveyard, 1, "Dead creature should be in graveyard")
	assert.Greater(t, len(opponent.Graveyard), initialOppGraveyardSize, "Graveyard should have creature in it")

	if len(opponent.Graveyard) > 0 {
		assert.Equal(t, -1, opponent.Graveyard[0].CurrentHealth, "Creature should have -1 health (overkilled by 1)")
	}
}

func TestPlayCard_HealEffect(t *testing.T) {
	// === SETUP PHASE ===
	d1 := smallDeck(10)
	d2 := smallDeck(10)

	opts := Options{
		StartingLife:     20,
		StartingHand:     0,
		MaxEnergy:        10,
		FirstPlayerDraws: false,
		Seed:             42,
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	activePlayer := g.Players[0]

	// === GAME STATE SETUP ===
	activePlayer.MaxEnergy = 10
	activePlayer.CurrentEnergy = 10
	activePlayer.Life = 15 // Reduced from starting 20

	// Create healing spell
	healSpell := CardInstance{
		InstanceID: "heal#1",
		Def: &cards.CardDef{
			ID:   "healing_potion",
			Name: "Healing Potion",
			Type: cards.TypeSpell,
			Cost: 2,
			Text: "Heal 5 life",
			Effects: []cards.Effect{
				{
					Kind:   cards.EffectHeal,
					Amount: 5,
					Target: cards.TargetSelfPlayer,
				},
			},
		},
		Owner:      activePlayer.PlayerID,
		Controller: activePlayer.PlayerID,
	}
	activePlayer.Hand = append(activePlayer.Hand, healSpell)

	// === RECORD INITIAL STATE ===
	initialLife := activePlayer.Life
	initialEnergy := activePlayer.CurrentEnergy

	// === EXECUTE ===
	targets := []*TargetRef{nil} // Auto-populated for self
	err = g.PlayCard(activePlayer.PlayerID, 0, targets)
	require.NoError(t, err, "PlayCard should succeed")

	// === ASSERTIONS ===
	assert.Equal(t, initialLife+5, activePlayer.Life, "Life should increase by heal amount")
	assert.Equal(t, initialEnergy-2, activePlayer.CurrentEnergy, "Energy should be reduced by spell cost")
	assert.Len(t, activePlayer.Hand, 0, "Spell should be removed from hand")
	assert.Len(t, activePlayer.Graveyard, 1, "Spell should be in graveyard")
}

func TestPlayCard_DrawCardsEffect(t *testing.T) {
	// === SETUP PHASE ===
	d1 := smallDeck(10)
	d2 := smallDeck(10)

	opts := Options{
		StartingLife:     20,
		StartingHand:     0,
		MaxEnergy:        10,
		FirstPlayerDraws: false,
		Seed:             42,
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	activePlayer := g.Players[0]

	// === GAME STATE SETUP ===
	activePlayer.MaxEnergy = 10
	activePlayer.CurrentEnergy = 10

	// Create draw spell
	drawSpell := CardInstance{
		InstanceID: "draw#1",
		Def: &cards.CardDef{
			ID:   "divination",
			Name: "Divination",
			Type: cards.TypeSpell,
			Cost: 3,
			Text: "Draw 2 cards",
			Effects: []cards.Effect{
				{
					Kind:   cards.EffectDrawCards,
					Amount: 2,
					Target: cards.TargetSelfPlayer,
				},
			},
		},
		Owner:      activePlayer.PlayerID,
		Controller: activePlayer.PlayerID,
	}
	activePlayer.Hand = append(activePlayer.Hand, drawSpell)

	// === RECORD INITIAL STATE ===
	initialHandSize := len(activePlayer.Hand)
	initialDeckSize := len(activePlayer.Deck)
	initialEnergy := activePlayer.CurrentEnergy

	// === EXECUTE ===
	targets := []*TargetRef{nil} // Auto-populated for self
	err = g.PlayCard(activePlayer.PlayerID, 0, targets)
	require.NoError(t, err, "PlayCard should succeed")

	// === ASSERTIONS ===
	assert.Equal(t, initialHandSize-1+2, len(activePlayer.Hand), "Hand should have 2 more cards (minus the spell played)")
	assert.Equal(t, initialDeckSize-2, len(activePlayer.Deck), "Deck should have 2 fewer cards")
	assert.Equal(t, initialEnergy-3, activePlayer.CurrentEnergy, "Energy should be reduced by spell cost")
	assert.Len(t, activePlayer.Graveyard, 1, "Spell should be in graveyard")
}

func TestPlayCard_BuffEffect(t *testing.T) {
	// === SETUP PHASE ===
	d1 := smallDeck(10)
	d2 := smallDeck(10)

	opts := Options{
		StartingLife:     20,
		StartingHand:     0,
		MaxEnergy:        10,
		FirstPlayerDraws: false,
		Seed:             42,
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	activePlayer := g.Players[0]

	// === GAME STATE SETUP ===
	activePlayer.MaxEnergy = 10
	activePlayer.CurrentEnergy = 10

	// Place target creature on board
	targetCreature := CardInstance{
		InstanceID: "creature#1",
		Def: &cards.CardDef{
			ID:     "bear",
			Name:   "Bear",
			Type:   cards.TypeCreature,
			Attack: 2,
			Health: 2,
		},
		Owner:         activePlayer.PlayerID,
		Controller:    activePlayer.PlayerID,
		CurrentAttack: 2,
		CurrentHealth: 2,
	}
	activePlayer.Board = append(activePlayer.Board, targetCreature)

	// Create buff spell
	buffSpell := CardInstance{
		InstanceID: "buff#1",
		Def: &cards.CardDef{
			ID:   "giant_growth",
			Name: "Giant Growth",
			Type: cards.TypeSpell,
			Cost: 1,
			Text: "Give target creature +3/+3",
			Effects: []cards.Effect{
				{
					Kind:       cards.EffectBuffStatsPerm,
					BuffAttack: 3,
					BuffHealth: 3,
					Target:     cards.TargetAllyCreature,
				},
			},
		},
		Owner:      activePlayer.PlayerID,
		Controller: activePlayer.PlayerID,
	}
	activePlayer.Hand = append(activePlayer.Hand, buffSpell)

	// === RECORD INITIAL STATE ===
	initialAttack := activePlayer.Board[0].CurrentAttack
	initialHealth := activePlayer.Board[0].CurrentHealth
	initialEnergy := activePlayer.CurrentEnergy

	// === EXECUTE ===
	targets := []*TargetRef{
		{InstanceID: &activePlayer.Board[0].InstanceID},
	}
	err = g.PlayCard(activePlayer.PlayerID, 0, targets)
	require.NoError(t, err, "PlayCard should succeed")

	// === ASSERTIONS ===
	assert.Equal(t, initialAttack+3, activePlayer.Board[0].CurrentAttack, "Creature attack should increase by 3")
	assert.Equal(t, initialHealth+3, activePlayer.Board[0].CurrentHealth, "Creature health should increase by 3")
	assert.Equal(t, initialEnergy-1, activePlayer.CurrentEnergy, "Energy should be reduced by spell cost")
	assert.Len(t, activePlayer.Hand, 0, "Spell should be removed from hand")
	assert.Len(t, activePlayer.Graveyard, 1, "Spell should be in graveyard")
}

func TestPlayCard_TempBuffEffect(t *testing.T) {
	// === SETUP PHASE ===
	d1 := smallDeck(10)
	d2 := smallDeck(10)

	opts := Options{
		StartingLife:     20,
		StartingHand:     0,
		MaxEnergy:        10,
		FirstPlayerDraws: false,
		Seed:             42,
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	activePlayer := g.Players[0]

	// === GAME STATE SETUP ===
	activePlayer.MaxEnergy = 10
	activePlayer.CurrentEnergy = 10

	// Place target creature on board
	targetCreature := CardInstance{
		InstanceID: "creature#1",
		Def: &cards.CardDef{
			ID:     "bear",
			Name:   "Bear",
			Type:   cards.TypeCreature,
			Attack: 2,
			Health: 2,
		},
		Owner:         activePlayer.PlayerID,
		Controller:    activePlayer.PlayerID,
		CurrentAttack: 2,
		CurrentHealth: 2,
	}
	activePlayer.Board = append(activePlayer.Board, targetCreature)

	// Create temp buff spell (like Giant Growth in MTG - until end of turn)
	tempBuffSpell := CardInstance{
		InstanceID: "tempbuff#1",
		Def: &cards.CardDef{
			ID:   "battle_fury",
			Name: "Battle Fury",
			Type: cards.TypeSpell,
			Cost: 1,
			Text: "Give target creature +3/+3 until end of turn",
			Effects: []cards.Effect{
				{
					Kind:       cards.EffectBuffStatsTemp,
					BuffAttack: 3,
					BuffHealth: 3,
					Target:     cards.TargetAllyCreature,
				},
			},
		},
		Owner:      activePlayer.PlayerID,
		Controller: activePlayer.PlayerID,
	}
	activePlayer.Hand = append(activePlayer.Hand, tempBuffSpell)

	// === RECORD INITIAL STATE ===
	initialAttack := activePlayer.Board[0].CurrentAttack
	initialHealth := activePlayer.Board[0].CurrentHealth
	initialEnergy := activePlayer.CurrentEnergy

	// === EXECUTE ===
	targets := []*TargetRef{
		{InstanceID: &activePlayer.Board[0].InstanceID},
	}
	err = g.PlayCard(activePlayer.PlayerID, 0, targets)
	require.NoError(t, err, "PlayCard should succeed")

	// === ASSERTIONS ===
	assert.Equal(t, initialAttack+3, activePlayer.Board[0].CurrentAttack, "Creature attack should increase by 3")
	assert.Equal(t, initialHealth+3, activePlayer.Board[0].CurrentHealth, "Creature health should increase by 3")
	assert.Equal(t, 3, activePlayer.Board[0].TempAttackBuff, "Temp attack buff should be 3")
	assert.Equal(t, 3, activePlayer.Board[0].TempHealthBuff, "Temp health buff should be 3")
	assert.Equal(t, 0, activePlayer.Board[0].PermAttackBuff, "Perm attack buff should be 0")
	assert.Equal(t, 0, activePlayer.Board[0].PermHealthBuff, "Perm health buff should be 0")
	assert.Equal(t, initialEnergy-1, activePlayer.CurrentEnergy, "Energy should be reduced by spell cost")
	assert.Len(t, activePlayer.Graveyard, 1, "Spell should be in graveyard")
}

func TestPlayCard_MultiEffect(t *testing.T) {
	// === SETUP PHASE ===
	d1 := smallDeck(10)
	d2 := smallDeck(10)

	opts := Options{
		StartingLife:     20,
		StartingHand:     0,
		MaxEnergy:        10,
		FirstPlayerDraws: false,
		Seed:             42,
	}

	g, err := NewGame("p0", "p1", d1, d2, opts)
	require.NoError(t, err)

	activePlayer := g.Players[0]
	opponent := g.Players[1]

	// === GAME STATE SETUP ===
	activePlayer.MaxEnergy = 10
	activePlayer.CurrentEnergy = 10

	// Place target creature on opponent's board
	targetCreature := CardInstance{
		InstanceID: "creature#1",
		Def: &cards.CardDef{
			ID:     "goblin",
			Name:   "Goblin",
			Type:   cards.TypeCreature,
			Attack: 1,
			Health: 1,
		},
		Owner:         opponent.PlayerID,
		Controller:    opponent.PlayerID,
		CurrentAttack: 1,
		CurrentHealth: 1,
	}
	opponent.Board = append(opponent.Board, targetCreature)

	// Create multi-effect spell
	multiSpell := CardInstance{
		InstanceID: "multi#1",
		Def: &cards.CardDef{
			ID:   "bolt_and_draw",
			Name: "Lightning Strike and Study",
			Type: cards.TypeSpell,
			Cost: 3,
			Text: "Deal 2 damage to target creature, then draw 1 card",
			Effects: []cards.Effect{
				{
					Kind:   cards.EffectDamage,
					Amount: 2,
					Target: cards.TargetAnyCreature,
				},
				{
					Kind:   cards.EffectDrawCards,
					Amount: 1,
					Target: cards.TargetSelfPlayer,
				},
			},
		},
		Owner:      activePlayer.PlayerID,
		Controller: activePlayer.PlayerID,
	}
	activePlayer.Hand = append(activePlayer.Hand, multiSpell)

	// === RECORD INITIAL STATE ===
	initialHandSize := len(activePlayer.Hand)
	initialDeckSize := len(activePlayer.Deck)
	initialOppBoardSize := len(opponent.Board)
	initialEnergy := activePlayer.CurrentEnergy

	// === EXECUTE ===
	targets := []*TargetRef{
		{InstanceID: &targetCreature.InstanceID}, // Target for damage
		nil,                                      // Auto-populated for self draw
	}
	err = g.PlayCard(activePlayer.PlayerID, 0, targets)
	require.NoError(t, err, "PlayCard should succeed")

	// === ASSERTIONS ===
	// Damage effect: creature should be dead (1 health - 2 damage = dead)
	assert.Len(t, opponent.Board, 0, "Creature should be dead from damage")
	assert.Len(t, opponent.Graveyard, 1, "Dead creature should be in graveyard")
	assert.Less(t, len(opponent.Board), initialOppBoardSize, "Opponent board should be smaller")

	// Draw effect: should have drawn 1 card
	assert.Equal(t, initialHandSize-1+1, len(activePlayer.Hand), "Hand should have 1 more card (minus spell, plus draw)")
	assert.Equal(t, initialDeckSize-1, len(activePlayer.Deck), "Deck should have 1 fewer card")

	// General spell effects
	assert.Equal(t, initialEnergy-3, activePlayer.CurrentEnergy, "Energy should be reduced by spell cost")
	assert.Len(t, activePlayer.Graveyard, 1, "Spell should be in graveyard")
}
