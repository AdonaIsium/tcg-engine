package cards

type Type string

const (
	TypeCreature Type = "creature"
	TypeSpell    Type = "spell"
)

type EffectKind string

const (
	EffectDamage        EffectKind = "damage"          // amount -> target (player/creature)
	EffectHeal          EffectKind = "heal"            // amount -> target (player/creature)
	EffectDrawCards     EffectKind = "draw_cards"      // amount -> self
	EffectBuffStatsPerm EffectKind = "buff_stats_perm" // attack_buff/health_buff -> target creature
	EffectBuffStatsTemp EffectKind = "buff_stats_temp" // attack_buff/health_buff -> target creature
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
	Kind       EffectKind `json:"kind"`
	Amount     int        `json:"amount,omitempty"`
	BuffAttack int        `json:"attack_buff,omitempty"`
	BuffHealth int        `json:"health_buff,omitempty"`
	Target     TargetKind `json:"target,omitempty"`
}

type CardDef struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    Type     `json:"type"`
	Cost    int      `json:"cost"`
	Attack  int      `json:"attack,omitempty"`
	Health  int      `json:"health,omitempty"`
	Text    string   `json:"text,omitempty"`
	Effects []Effect `json:"effects,omitempty"`
}
