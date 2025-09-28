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
					{Kind: cards.EffectBuffStats, BuffAttack: 2, BuffHealth: 2, Target: cards.TargetAllyCreature},
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
