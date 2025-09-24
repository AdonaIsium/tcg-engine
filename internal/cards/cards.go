package cards

type Type string

const (
	TypeCreature Type = "creature"
	TypeSpell    Type = "spell"
)

type EffectKind string

const (
	EffectDamage     EffectKind = "damage"      // amount -> target (player/creature)
	EffectHeal       EffectKind = "heal"        // amount -> target (player/creature)
	EffectDrawCards  EffectKind = "draw_cards"  // amount -> self
	EffectBuffAttack EffectKind = "buff_attack" // amount -> target creature
	EffectBuffHealth EffectKind = "buff_health" // amount -> target creature (non-heal perm buff)
)

type TargetKind string

const (
	TargetNone          TargetKind = "none"
	TargetEnemyPlayer   TargetKind = "enemy_player"
	TargetSelfPlayer    TargetKind = "self_player"
	TargetAnyCreature   TargetKind = "any_creature"
	TargetEnemyCreature TargetKind = "enemy_creature"
	TargetAllyCreature  TargetKind = "ally_creature"
)

type Effect struct {
	Kind   EffectKind `json:"kind"`
	Amount int        `json:"amount,omitempty"`
	Target TargetKind `json:"target,omitempty"`
}

type CardDef struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        Type    `json:"type"`
	Cost        int     `json:"cost"`
	Attack      int     `json:"attack,omitempty"`
	Health      int     `json:"health,omitempty"`
	Text        string  `json:"text,omitempty"`
	SpellEffect *Effect `json:"spell_effect,omitempty"`
}
