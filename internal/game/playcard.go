package game

import (
	"errors"
	"fmt"

	"github.com/AdonaIsium/tcg-engine/internal/cards"
)

var (
	ErrNotYourTurn      = errors.New("not your turn")
	ErrInvalidHandIndex = errors.New("invalid hand index")
	ErrNotEnoughEnergy  = errors.New("not enough energy")
	ErrMissingTarget    = errors.New("missing target")
	ErrInvalidTarget    = errors.New("invalid target")
	ErrBoardFull        = errors.New("board is full")
	ErrPlayerNotFound   = errors.New("player not found")
)

// EffectContext holds all the information any effect function might need
type EffectContext struct {
	Game       *Game
	Caster     *PlayerState
	Target     *TargetRef
	Amount     int
	BuffAttack int
	BuffHealth int
}

// effectResolver is our function map - maps effect kinds to their implementation
var effectResolver = map[cards.EffectKind]func(*EffectContext) error{
	cards.EffectDamage:        applyDamage,
	cards.EffectHeal:          applyHealing,
	cards.EffectDrawCards:     applyDrawCards,
	cards.EffectBuffStatsPerm: applyBuffStatsPerm,
	cards.EffectBuffStatsTemp: applyBuffStatsTemp,
}

func (g *Game) autoPopulateTarget(effect cards.Effect, providedTarget *TargetRef, caster *PlayerState) *TargetRef {
	switch effect.Target {
	case cards.TargetSelfPlayer:
		return &TargetRef{PlayerID: caster.PlayerID}
	case cards.TargetEnemyPlayer:
		return &TargetRef{PlayerID: g.Opponent().PlayerID}
	default:
		return providedTarget
	}
}

func (g *Game) PlayCard(playerID string, handIdx int, targets []*TargetRef) error {
	if g.CurrentPlayer().PlayerID != playerID {
		return ErrNotYourTurn
	}

	activePlayer := g.CurrentPlayer()

	if handIdx < 0 || handIdx >= len(activePlayer.Hand) {
		return ErrInvalidHandIndex
	}

	card := activePlayer.Hand[handIdx]

	if activePlayer.CurrentEnergy < card.Def.Cost {
		return ErrNotEnoughEnergy
	}

	if card.Def.Type == cards.TypeSpell {
		if len(card.Def.Effects) <= 0 {
			return fmt.Errorf("spells must have at least 1 effect, got %d", len(card.Def.Effects))
		}
		if len(targets) != len(card.Def.Effects) {
			return fmt.Errorf("expected %d targets, got %d", len(card.Def.Effects), len(targets))
		}

		for i, effect := range card.Def.Effects {
			if effect.Target == cards.TargetSelfPlayer || effect.Target == cards.TargetEnemyPlayer {
				continue
			}

			if err := g.validateTarget(effect.Target, targets[i], activePlayer); err != nil {
				return fmt.Errorf("effect %d validation failed: %w", i, err)
			}
		}
	}

	activePlayer.CurrentEnergy -= card.Def.Cost

	if card.Def.Type == cards.TypeCreature {
		activePlayer.Board = append(activePlayer.Board, card)
	}

	if card.Def.Type == cards.TypeSpell {
		activePlayer.Graveyard = append(activePlayer.Graveyard, card)
	}

	if len(activePlayer.Hand) > handIdx+1 {
		activePlayer.Hand = append(activePlayer.Hand[:handIdx], activePlayer.Hand[handIdx+1:]...)
	} else {
		activePlayer.Hand = activePlayer.Hand[:handIdx]
	}

	if card.Def.Type == cards.TypeSpell {
		for i, effect := range card.Def.Effects {
			applySpell := effectResolver[effect.Kind]
			actualTarget := g.autoPopulateTarget(effect, targets[i], activePlayer)
			effectContext := EffectContext{Game: g, Caster: activePlayer, Target: actualTarget, Amount: effect.Amount, BuffAttack: effect.BuffAttack, BuffHealth: effect.BuffHealth}
			applySpell(&effectContext)
		}
	}

	// Step 7 - Check for state-based effects (creature death, game end, etc.)
	gameEnded, endMessage := g.checkStateBasedEffects()
	if gameEnded {
		g.GameEnded = true
		g.Winner = endMessage
		g.log("state_based_effects", "", "Game ended: %s", endMessage)
		return nil // PlayCard succeeds even if game ends
	}

	return nil
}

// Apply damage to player or creature
func applyDamage(ctx *EffectContext) error {
	// Try player damage first
	if player := ctx.Game.getTargetPlayer(ctx.Target); player != nil {
		player.Life -= ctx.Amount
		ctx.Game.log("damage", ctx.Caster.PlayerID, "%d damage dealt to %s", ctx.Amount, player.PlayerID)
		return nil
	}

	// Try creature damage
	if creature := ctx.Game.getTargetCreature(ctx.Target); creature != nil {
		creature.CurrentDamage += ctx.Amount
		creature.CurrentHealth = creature.Def.Health + creature.PermHealthBuff + creature.TempHealthBuff - creature.CurrentDamage
		ctx.Game.log("damage", ctx.Caster.PlayerID, "%d damage dealt to %s", ctx.Amount, creature.Def.Name)

		// Check for creature death
		if creature.CurrentHealth <= 0 {
			return ctx.Game.moveToGraveyard(creature, fmt.Sprintf("destroyed by %d damage", ctx.Amount))
		}
		return nil
	}

	// Should never happen if validation worked
	return fmt.Errorf("applyDamage: no valid target found")
}

func applyHealing(ctx *EffectContext) error {
	if player := ctx.Game.getTargetPlayer(ctx.Target); player != nil {
		player.Life += ctx.Amount
		ctx.Game.log("healing", ctx.Caster.PlayerID, "%d healing applied to %s", ctx.Amount, player.PlayerID)
		return nil
	}
	return fmt.Errorf("applyHealing: %w", ErrPlayerNotFound)
}

func applyDrawCards(ctx *EffectContext) error {
	if player := ctx.Game.getTargetPlayer(ctx.Target); player != nil {
		drawn := ctx.Game.Draw(player, ctx.Amount)
		ctx.Game.log("draw_cards", ctx.Caster.PlayerID, "%s drew %d cards", player.PlayerID, drawn)
		return nil
	}
	return fmt.Errorf("applyDrawCards: %w", ErrPlayerNotFound)
}

func applyBuffStatsPerm(ctx *EffectContext) error {
	if creature := ctx.Game.getTargetCreature(ctx.Target); creature != nil {
		creature.PermAttackBuff += ctx.BuffAttack
		creature.PermHealthBuff += ctx.BuffHealth
		creature.CurrentAttack = creature.Def.Attack + creature.PermAttackBuff + creature.TempAttackBuff
		creature.CurrentHealth = creature.Def.Health + creature.PermHealthBuff + creature.TempHealthBuff
		ctx.Game.log("buff_creature_perm", ctx.Caster.PlayerID, "+%d/+%d permanent buff applied to %s", ctx.BuffAttack, ctx.BuffHealth, creature.InstanceID)
		return nil
	}
	return fmt.Errorf("applyBuffStatsPerm: %w", ErrInvalidTarget)
}

func applyBuffStatsTemp(ctx *EffectContext) error {
	if creature := ctx.Game.getTargetCreature(ctx.Target); creature != nil {
		creature.TempAttackBuff += ctx.BuffAttack
		creature.TempHealthBuff += ctx.BuffHealth
		creature.CurrentAttack = creature.Def.Attack + creature.PermAttackBuff + creature.TempAttackBuff
		creature.CurrentHealth = creature.Def.Health + creature.PermHealthBuff + creature.TempHealthBuff
		ctx.Game.log("buff_creature_temp", ctx.Caster.PlayerID, "+%d/+%d temporary buff applied to %s", ctx.BuffAttack, ctx.BuffHealth, creature.InstanceID)
		return nil
	}
	return fmt.Errorf("applyBuffStatsTemp: %w", ErrInvalidTarget)
}

// Helper functions for effect targeting - assume validation already passed
func (g *Game) getTargetPlayer(targetRef *TargetRef) *PlayerState {
	if targetRef.PlayerID == "" {
		return nil
	}
	for _, player := range g.Players {
		if player.PlayerID == targetRef.PlayerID {
			return player
		}
	}
	// Should never happen if validation worked
	g.log("error", "", "getTargetPlayer failed to find player %s", targetRef.PlayerID)
	return nil
}

func (g *Game) getTargetCreature(targetRef *TargetRef) *CardInstance {
	if targetRef.InstanceID == nil {
		return nil
	}
	for _, player := range g.Players {
		// Board should be the only place where creatures can take damage
		for i := range player.Board {
			if player.Board[i].InstanceID == *targetRef.InstanceID {
				return &player.Board[i] // Direct reference to slice element
			}
		}
	}
	// Should never happen if validation worked
	g.log("error", "", "getTargetCreature failed to find creature %s", *targetRef.InstanceID)
	return nil
}

func (g *Game) checkStateBasedEffects() (bool, string) {
	p1Alive := g.Players[0].Life > 0
	p2Alive := g.Players[1].Life > 0

	// Check for game-ending conditions first
	switch {
	case !p1Alive && !p2Alive:
		g.log("game_end", "", "Both players died simultaneously - game is a draw!")
		return true, "Draw!"
	case p1Alive && !p2Alive:
		winner := g.Players[0].Name
		g.log("game_end", g.Players[0].PlayerID, "%s wins! %s died with %d life", winner, g.Players[1].Name, g.Players[1].Life)
		return true, fmt.Sprintf("%s a winner is you!", winner)
	case !p1Alive && p2Alive:
		winner := g.Players[1].Name
		g.log("game_end", g.Players[1].PlayerID, "%s wins! %s died with %d life", winner, g.Players[0].Name, g.Players[0].Life)
		return true, fmt.Sprintf("%s a winner is you!", winner)
	}

	// Check for creature deaths (iterate backwards to handle removal safely)
	for i := range g.Players {
		for j := len(g.Players[i].Board) - 1; j >= 0; j-- {
			creature := &g.Players[i].Board[j]
			if creature.CurrentHealth <= 0 {
				g.log("creature_death", g.Players[i].PlayerID, "%s (%s) died with %d health", creature.Def.Name, creature.InstanceID, creature.CurrentHealth)
				g.moveToGraveyard(creature, "life reached 0")
			}
		}
	}

	return false, ""
}

func (g *Game) CanPlayCard(playerID string, handIdx int, targets []*TargetRef) error {
	if g.CurrentPlayer().PlayerID != playerID {
		return ErrNotYourTurn
	}

	activePlayer := g.CurrentPlayer()

	if handIdx < 0 || handIdx >= len(activePlayer.Hand) {
		return ErrInvalidHandIndex
	}

	card := activePlayer.Hand[handIdx]

	if activePlayer.CurrentEnergy < card.Def.Cost {
		return ErrNotEnoughEnergy
	}

	if card.Def.Type == cards.TypeSpell {
		if len(card.Def.Effects) <= 0 {
			return fmt.Errorf("spells must have at least 1 effect, got %d", len(card.Def.Effects))
		}
		if len(targets) != len(card.Def.Effects) {
			return fmt.Errorf("expected %d targets, got %d", len(card.Def.Effects), len(targets))
		}

		for i, effect := range card.Def.Effects {
			if effect.Target == cards.TargetSelfPlayer || effect.Target == cards.TargetEnemyPlayer {
				continue
			}

			if err := g.validateTarget(effect.Target, targets[i], activePlayer); err != nil {
				return fmt.Errorf("effect %d validation failed: %w", i, err)
			}
		}
	}

	if card.Def.Type == cards.TypeCreature {
		if g.Options.MaxBoardSize > 0 && len(activePlayer.Board) >= g.Options.MaxBoardSize {
			return ErrBoardFull
		}
	}

	return nil
}
