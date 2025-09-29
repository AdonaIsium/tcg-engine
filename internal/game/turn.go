package game

import (
	"fmt"

	"github.com/AdonaIsium/tcg-engine/internal/cards"
)

func (g *Game) StartTurn() {
	// Validate active index & player
	if g.Active < 0 || g.Active > 1 || g.Players[g.Active] == nil {
		g.log("error", "", "invalid active index or nil player: active=%d", g.Active)
		return
	}

	activePlayer := g.CurrentPlayer()

	// Advance turn counter
	if g.Turn == 0 {
		g.Turn = 1
	} else {
		g.Turn++
	}

	// Energy ramp then refill
	newCap := min(activePlayer.MaxEnergy+1, g.Options.MaxEnergy)
	activePlayer.MaxEnergy = newCap
	activePlayer.CurrentEnergy = activePlayer.MaxEnergy

	// Draw step (skipping first player's draw when appropriate)
	skipFirst := (g.Turn == 1 && g.Active == 0 && !g.Options.FirstPlayerDraws)
	if !skipFirst {
		_ = g.Draw(activePlayer, 1)
	} else {
		g.log("draw", activePlayer.PlayerID, "no card drawn (first turn skip)")
	}

	g.refreshCreatures(activePlayer)

	g.log("start", activePlayer.PlayerID, "start turn: cap=%d energy=%d", activePlayer.MaxEnergy, activePlayer.CurrentEnergy)
}

func (g *Game) EndTurn() {
	if g.Active < 0 || g.Active > 1 || g.Players[g.Active] == nil {
		g.log("error", "", "invalid active index or nil player in EndTurn: active=%d", g.Active)
		return
	}
	g.log("end", g.Players[g.Active].PlayerID, "end turn")
	g.CleanupTurn()
	g.Active = 1 - g.Active
}

func (g *Game) CurrentPlayer() *PlayerState {
	return g.Players[g.Active]
}

func (g *Game) Opponent() *PlayerState {
	return g.Players[1-g.Active]
}

func (g *Game) Draw(player *PlayerState, n int) int {
	drawn := 0
	for range n {
		if len(player.Deck) == 0 {
			break
		}
		top := len(player.Deck) - 1
		card := player.Deck[top]
		player.Deck = player.Deck[:top]
		player.Hand = append(player.Hand, card)
		drawn++
	}

	g.log("draw", player.PlayerID, "drew %d", drawn)

	return drawn
}

func (g *Game) refreshCreatures(ps *PlayerState) {
	for i := range ps.Board {
		ci := &ps.Board[i]
		if ci.Def.Type == cards.TypeCreature {
			ci.Exhausted = false
			if ci.SummoningSick {
				ci.SummoningSick = false
			}
		}
	}
}

func (g *Game) refreshCreatureHealth() {
	for i := range g.Players {
		for j := range g.Players[i].Board {
			card := &g.Players[i].Board[j]

			if card.CurrentDamage > 0 || card.TempAttackBuff != 0 || card.TempHealthBuff != 0 {
				g.log("refresh_creature_before", g.Players[i].PlayerID, "refreshing %s (%s) attack/health from %d/%d\nbase_attack = %d, base_health = %d\nperm_attack_buff = %d, perm_health_buff = %d\ncurrent temp effects: damage = %d, temp_attack_buff = %d, temp_health_buff = %d", card.Def.Name, card.InstanceID, card.CurrentAttack, card.CurrentHealth, card.Def.Attack, card.Def.Health, card.PermAttackBuff, card.PermHealthBuff, card.CurrentDamage, card.TempAttackBuff, card.TempHealthBuff)

				// Reset temporary status effects
				card.CurrentDamage = 0
				card.TempAttackBuff = 0
				card.TempHealthBuff = 0

				// Recalculate attack/health using only permanent statuses
				card.CurrentAttack = card.Def.Attack + card.PermAttackBuff
				card.CurrentHealth = card.Def.Health + card.PermHealthBuff
				g.log("refresh_creature_after", g.Players[i].PlayerID, "%s (%s) attack/health refreshed to %d/%d", card.Def.Name, card.InstanceID, card.CurrentAttack, card.CurrentHealth)
			}
		}
	}
}

func (g *Game) CleanupTurn() {
	g.refreshCreatureHealth()
}

func (g *Game) log(eventType, playerID, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	g.Log = append(g.Log, Event{Turn: g.Turn, Player: playerID, Type: eventType, Msg: msg})
}
