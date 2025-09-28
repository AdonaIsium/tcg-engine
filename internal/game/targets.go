package game

import (
	"github.com/AdonaIsium/tcg-engine/internal/cards"
)

// TargetRef is a reference to something in the game state
// that a spell/effect can point at.
type TargetRef struct {
	PlayerID   string      // if targeting a player
	InstanceID *InstanceID // if targeting a creature
}

// validateTarget checks that the given TargetRef satisfies
// the requirements of the card's Effect.Target.
func (g *Game) validateTarget(req cards.TargetKind, target *TargetRef, caster *PlayerState) error {
	switch req {

	// No target expected — target must be nil
	case cards.TargetNone:
		if target != nil {
			return ErrInvalidTarget
		}

	// Any creature on either board
	case cards.TargetAnyCreature:
		if target == nil || target.InstanceID == nil {
			return ErrMissingTarget
		}
		if _, ok := g.findCardInstance(*target.InstanceID); !ok {
			return ErrInvalidTarget
		}

	// Creature you control
	case cards.TargetAllyCreature:
		if target == nil || target.InstanceID == nil {
			return ErrMissingTarget
		}
		if ci, ok := g.findCardInstance(*target.InstanceID); !ok || !g.controls(caster, ci) {
			return ErrInvalidTarget
		}

	// Creature opponent controls
	case cards.TargetEnemyCreature:
		if target == nil || target.InstanceID == nil {
			return ErrMissingTarget
		}
		if ci, ok := g.findCardInstance(*target.InstanceID); !ok || g.controls(caster, ci) {
			return ErrInvalidTarget
		}

	// Self player (the caster)
	case cards.TargetSelfPlayer:
		if target == nil || target.PlayerID != caster.PlayerID {
			return ErrInvalidTarget
		}

	// Opponent player
	case cards.TargetEnemyPlayer:
		opp := g.Opponent()
		if target == nil || target.PlayerID != opp.PlayerID {
			return ErrInvalidTarget
		}

	// Unknown/unsupported target kind
	default:
		return ErrInvalidTarget
	}

	return nil
}

// findCardInstance searches all boards for a card with the given instance ID.
func (g *Game) findCardInstance(id InstanceID) (*CardInstance, bool) {
	for _, p := range g.Players {
		for i := range p.Board {
			if p.Board[i].InstanceID == id {
				return &p.Board[i], true
			}
		}
	}
	return nil, false
}

// controls reports whether the given player controls the card instance.
func (g *Game) controls(player *PlayerState, ci *CardInstance) bool {
	for i := range player.Board {
		if &player.Board[i] == ci {
			return true
		}
	}
	return false
}
