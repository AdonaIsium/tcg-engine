# Card Resolution Architecture Design Decisions

## Overview
This document captures the key design decisions made during the implementation of the card resolution system for the TCG engine. Each decision includes the options considered and rationale for the chosen approach.

---

## 1. Effect Resolution Model

**Problem:** How should card effects be resolved when played?

**Options Considered:**
- **Atomic**: Everything happens instantly, no interruptions possible
- **Stackable**: Like Magic, where responses are possible between effects
- **Sequential**: Effects resolve one by one, but no responses allowed

**Decision:** **Atomic Resolution**
**Rationale:**
- Simpler to implement and test for a learning project
- Reduces complexity exponentially (no need for priority, stack management, response validation)
- Can be refactored to stackable later once fundamentals are mastered
- No other players available for testing interactive elements

---

## 2. Card Play Sequence

**Problem:** What order should events happen when a card is played?

**Decision:** 5-Step Atomic Sequence:
1. Check validation (turn, hand index, energy, targets)
2. Pay the cost (subtract energy)
3. Move card to appropriate zone (board/graveyard)
4. Apply effects (if any)
5. Check state-based effects (creature death, etc.)

**Rationale:** Prevents energy loss on invalid plays, handles zone movement before effects resolve, ensures state consistency.

---

## 3. Zone Movement for Cards

**Problem:** Where should cards go when played?

**Decision:**
- **Creatures**: Hand → Board (card becomes the permanent)
- **Spells**: Hand → Graveyard (after effects resolve)
- **Hand Management**: Shift remaining cards down (no gaps in array)

**Rationale:** Clean, intuitive model where creatures ARE the card, spells are consumed after use.

---

## 4. Function Organization for Effects

**Problem:** How should different effect types be implemented?

**Options Considered:**
- **Big Switch Statement**: One large switch with all effect logic
- **Dedicated Functions**: Separate function for each effect type
- **Strategy Pattern**: Function map with dedicated implementations

**Decision:** **Strategy Pattern with Function Map**
```go
var effectResolver = map[cards.EffectKind]func(*EffectContext) error{
    cards.EffectDamage: applyDamage,
    cards.EffectHeal: applyHeal,
    // etc...
}
```

**Rationale:**
- Avoids maintainability nightmare of large switch statements
- Clean separation of concerns
- Easy to extend with new effect types
- Natural place for complex effect logic to grow

---

## 5. Effect Context Architecture

**Problem:** How should effect functions receive parameters they need?

**Decision:** **EffectContext Struct**
```go
type EffectContext struct {
    Game   *Game
    Caster *PlayerState
    Target *TargetRef
    Amount int
}
```

**Rationale:**
- Uniform function signatures across all effect types
- Easy to extend with new parameters
- Each effect function validates its own requirements
- Clean dispatch mechanism

---

## 6. Multi-Effect Cards Support

**Problem:** How to handle cards with multiple effects (e.g., "Deal 2 damage and draw 1 card")?

**Options Considered:**
- **Expanded Effect Struct**: Multiple amount fields (Damage, Healing, CardDraw, etc.)
- **Multiple Effects Array**: `Effects []Effect` instead of single effect
- **Flexible Parameters**: `Params map[string]int`

**Decision:** **Multiple Effects Array**
```go
type CardDef struct {
    // ...
    Effects []Effect `json:"effects,omitempty"`
}
```

**Rationale:**
- Maximum flexibility for complex cards
- Supports duplicate effects ("Give two creatures +2/+2")
- Scales naturally as new effect types are added
- Compositional design - complex behaviors from simple pieces

---

## 7. Multi-Effect Targeting

**Problem:** How should targeting work when cards have multiple effects?

**Decision:** **Effect-Level Targeting with Auto-Population**
- Each `Effect` has its own `Target` field
- PlayCard accepts `[]*TargetRef` in effect order
- Auto-targeting for self/enemy effects:
  - `TargetSelfPlayer` → Auto-fills with caster
  - `TargetEnemyPlayer` → Auto-fills with opponent
  - `TargetAnyCreature` → Player must choose

**Rationale:**
- Three targeting modes cover 95% of card design space
- Intuitive "left-to-right" targeting matches card text
- Reduces UI complexity (no nil targets for self-effects)
- Explicit validation prevents targeting errors

---

## 8. Buff Effect Consolidation

**Problem:** Should attack and health buffs be separate effects?

**Decision:** **Unified EffectBuffStats**
- Replace `EffectBuffAttack` and `EffectBuffHealth` with single `EffectBuffStats`
- Single effect can buff attack, health, or both simultaneously

**Rationale:**
- Prevents timing issues with "+2/+2" style effects
- Matches player intuition (one effect, one resolution)
- Reduces effect count for complex buff spells
- Atomic application ensures consistency

---

## 9. Effect Resolution Order

**Problem:** Should multiple effects resolve simultaneously or sequentially?

**Decision:** **Sequential with State Checks**
- Effects resolve in card text order
- State-based effects (creature death) checked between each effect
- Example: "Deal 3 damage, then draw a card" → damage first, death check, then draw

**Rationale:**
- Predictable timing for complex interactions
- Prevents "dead creature triggers" paradoxes
- Matches intuitive reading of card text
- Enables conditional effects based on previous results

---

---

## 10. Owner vs Controller Architecture

**Problem:** How should card ownership be tracked, especially for mind-control effects?

**Decision:** **Dual Tracking System**
```go
type CardInstance struct {
    Owner      string // PlayerID who owns this card (original deck)
    Controller string // PlayerID who currently controls this card
    // ...
}
```

**Rationale:**
- **Owner**: Never changes, determines graveyard destination
- **Controller**: Can change via effects, determines who can activate abilities
- **Graveyard Rule**: Cards always return to owner's graveyard, not controller's
- **Mind Control**: Stolen creatures return to original owner when destroyed
- **Future-Proof**: Enables complex control-changing effects

---

## 11. Zone Management System

**Problem:** How to handle card movement between zones (deck, hand, board, graveyard)?

**Decision:** **Centralized Helper with Auto-Detection**
```go
func (g *Game) moveToGraveyard(cardInstance *CardInstance, reason string) error
func (g *Game) findCardInZones(cardInstance *CardInstance) (*PlayerState, Zone, int, error)
```

**Rationale:**
- **Auto-Zone Detection**: Function finds card automatically across all zones
- **Owner Logic**: Always moves to owner's graveyard, regardless of controller
- **Reason Tracking**: Enables detailed logging and future triggered abilities
- **Error Handling**: Comprehensive validation prevents impossible states
- **Reusability**: Same helper for damage, destruction, sacrifice, etc.

**Search Order Optimization**: Board → Hand → Graveyard → Deck (most to least likely)

---

## 12. Turn Cleanup Timing

**Problem:** When should creature healing and other cleanup effects occur?

**Decision:** **CleanupTurn Phase Before Player Switch**
```go
EndTurn() → CleanupTurn() → switch active player
```

**Rationale:**
- **Precise Timing**: Cleanup happens before opponent's turn begins
- **Future Stack Support**: Maintains timing distinction for potential stack system
- **Separation of Concerns**: `EndTurn()` delegates cleanup, stays focused
- **Creature Healing**: All surviving creatures heal to full at turn end
- **State Consistency**: Ensures clean state before next player's actions

---

## 13. Effect Function Organization

**Problem:** How should effect validation and state changes be coordinated?

**Decision:** **Smart Helper Functions with Immediate State Changes**
```go
applyDamage() → check creature death → moveToGraveyard() if needed
```

**Rationale:**
- **Immediate Resolution**: State changes happen during effect application
- **Helper Integration**: Effect functions coordinate with zone management
- **Atomic Operations**: Each effect fully resolves before the next begins
- **Future Triggers**: Foundation for "when creature dies" triggered abilities
- **Graveyard Ordering**: Spells and creatures handled in proper sequence

---

## 14. Validation Philosophy

**Problem:** Should validation be strict or permissive for edge cases?

**Decision:** **Pragmatic Validation - Strict Where It Matters**
- **Target Count**: Strict validation (must match effect count)
- **Auto-Targeting**: Permissive (ignore provided values for self/enemy effects)
- **Zone Location**: Auto-detection rather than requiring caller knowledge
- **Error Messages**: Detailed context for debugging

**Rationale:**
- **Developer Experience**: Helpful errors without over-engineering
- **Future Flexibility**: Auto-detection enables complex card interactions
- **Fail Fast**: Critical errors caught early with clear messaging
- **Simplicity**: Avoid validation that doesn't improve game correctness

---

## Implementation Notes

All decisions prioritize:
1. **Learning-friendly complexity** - Start simple, add sophistication later
2. **Extensibility** - Easy to add new card types and effects
3. **Predictability** - Clear, intuitive behavior for players
4. **Maintainability** - Clean code organization for long-term development
5. **Professional Patterns** - Follows Magic: The Gathering design principles
6. **Owner/Controller Distinction** - Handles complex card interactions correctly
7. **Zone Management** - Robust foundation for all card movement needs

These decisions form the foundation for a robust, scalable TCG engine that can grow in complexity while maintaining clean architecture. The system now supports complex multi-effect cards, proper ownership tracking, and sophisticated targeting while remaining approachable for continued development.