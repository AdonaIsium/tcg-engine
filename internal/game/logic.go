package game

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/AdonaIsium/tcg-engine/internal/cards"
)

func NewGame(p1ID, p2ID string, d1, d2 []cards.CardDef, opts Options) (*Game, error) {
	if p1ID == "" || p2ID == "" {
		return nil, errors.New("player IDs must not be empty")
	}

	if len(d1) == 0 || len(d2) == 0 {
		return nil, errors.New("both players must provide a non-empty deck")
	}

	if opts.StartingLife <= 0 {
		opts.StartingLife = 20
	}
	if opts.StartingHand < 0 {
		opts.StartingHand = 3
	}
	if opts.MaxEnergy <= 0 {
		opts.MaxEnergy = 10
	}

	seed := opts.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	r := rand.New(rand.NewSource(seed))

	var nextInstance int64
	newInstanceID := func(base string) InstanceID {
		nextInstance++
		return InstanceID(fmt.Sprintf("%s#%d", base, nextInstance))
	}

	toInstances := func(defs []cards.CardDef) []CardInstance {
		out := make([]CardInstance, 0, len(defs))
		for i := range defs {
			def := &defs[i]
			inst := CardInstance{
				InstanceID:    newInstanceID(def.ID),
				Def:           def,
				CurrentAttack: def.Attack,
				CurrentHealth: def.Health,
				SummoningSick: false,
				Exhausted:     false,
			}
			out = append(out, inst)
		}
		return out
	}

	shuffle := func(insts []CardInstance) {
		r.Shuffle(len(insts), func(i, j int) { insts[i], insts[j] = insts[j], insts[i] })
	}

	drawN := func(ps *PlayerState, n int) {
		for i := 0; i < n; i++ {
			if len(ps.Deck) == 0 {
				break
			}
			top := len(ps.Deck) - 1
			card := ps.Deck[top]
			ps.Deck = ps.Deck[:top]
			ps.Hand = append(ps.Hand, card)
		}
	}

	p0 := &PlayerState{
		PlayerID:  p1ID,
		Life:      opts.StartingLife,
		Deck:      toInstances(d1),
		Hand:      nil,
		Board:     nil,
		Graveyard: nil,
		Energy:    0,
		MaxEnergy: 0,
	}
	p1 := &PlayerState{
		PlayerID:  p2ID,
		Life:      opts.StartingLife,
		Deck:      toInstances(d2),
		Hand:      nil,
		Board:     nil,
		Graveyard: nil,
		Energy:    0,
		MaxEnergy: 0,
	}

	shuffle(p0.Deck)
	shuffle(p1.Deck)

	g := &Game{
		ID:      fmt.Sprintf("g_%08x", r.Uint64()),
		Players: [2]*PlayerState{p0, p1},
		Active:  0,
		Turn:    0,
		Options: opts,
		Rand:    &randAdapter{r: r},
		Log:     nil,
	}

	if opts.StartingHand > 0 {
		drawN(p0, opts.StartingHand)
		drawN(p1, opts.StartingHand)
	}

	g.Log = append(g.Log,
		Event{Turn: g.Turn, Player: "", Type: "init", Msg: "game created"},
		Event{Turn: g.Turn, Player: p0.PlayerID, Type: "draw", Msg: fmt.Sprintf("opening hand: %d", len(p0.Hand))},
		Event{Turn: g.Turn, Player: p1.PlayerID, Type: "draw", Msg: fmt.Sprintf("opening hand: %d", len(p1.Hand))},
	)

	return g, nil
}

type randAdapter struct {
	r *rand.Rand
}

func (ra *randAdapter) Intn(n int) int {
	return ra.r.Intn(n)
}
