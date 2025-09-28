package game

import (
	"strconv"

	"github.com/AdonaIsium/tcg-engine/internal/cards"
)

// Helper functions
func makeDeck(n int) []cards.CardDef {
	deck := make([]cards.CardDef, 0, n)
	for i := range n {
		if i%3 == 2 {
			deck = append(deck, cards.CardDef{
				ID:   "s_firebolt_" + strconv.Itoa(i),
				Name: "Fire Bolt " + strconv.Itoa(i),
				Type: cards.TypeSpell,
				Cost: 1,
				Text: "Deal 2 damage to any target.",
				Effects: []cards.Effect{
					{
						Kind:   cards.EffectDamage,
						Amount: 2,
						Target: cards.TargetAnyCreature,
					},
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

func smallDeck(n int) []cards.CardDef {
	out := make([]cards.CardDef, 0, n)
	for i := range n {
		out = append(out, cards.CardDef{
			ID:     "c_" + strconvItoa(i),
			Name:   "Unit " + strconvItoa(i),
			Type:   cards.TypeCreature,
			Cost:   1,
			Attack: 1,
			Health: 1,
		})
	}
	return out
}

func clonePlayerState(p *PlayerState) PlayerState {
	cp := *p // shallow copy value fields

	cloneInstances := func(src []CardInstance) []CardInstance {
		if src == nil {
			return nil
		}
		dst := make([]CardInstance, len(src))
		copy(dst, src)
		return dst
	}

	cp.Deck = cloneInstances(p.Deck)
	cp.Hand = cloneInstances(p.Hand)
	cp.Board = cloneInstances(p.Board)
	cp.Graveyard = cloneInstances(p.Graveyard)

	return cp
}

func collectIDs(insts []CardInstance) []string {
	out := make([]string, len(insts))
	for i := range insts {
		out[i] = string(insts[i].InstanceID)
	}
	return out
}

// tiny int->string to avoid extra imports in examples
func strconvItoa(i int) string {
	const d = "0123456789"
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	p := len(buf)
	for n := i; n > 0; n /= 10 {
		p--
		buf[p] = d[n%10]
	}
	return string(buf[p:])
}

// helper to quickly make *InstanceID
func ptrInstance(id string) *InstanceID {
	inst := InstanceID(id)
	return &inst
}
