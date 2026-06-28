# Combat System — Madokita

## Overview

Fighting-game combat system with frame-based collision detection, team-aware targeting, multi-phase attacks, and AI controllers.

## Collision System

`internal/combat/collision.go`

- `CollisionSystem` registers hitboxes and actors
- Each frame checks overlap between active hitboxes and eligible hurtboxes
- Team filtering: hitboxes only affect hostile teams
- Produces `HitResult` on successful collision

## Combat Actor

`internal/combat/actor.go`

Represents a combat participant with:

- `Team` affiliation
- `Stats` (health, stamina, poise, attack, defense)
- State flags: invincibility, stagger, flinch, hyper-armor
- `ReceiveHit(hit HitResult)` — applies damage and stagger effects

**States**: stagger immunity, hyper armor absorption, death handling.

## Stats

`internal/combat/stats.go`

```go
type Stats struct {
    Health, MaxHealth       float64
    Stamina, MaxStamina     float64
    Poise, MaxPoise         float64
    Attack, Defense         float64
    StaminaRegen            float64
    StaggerImmunityDuration float64
    HyperArmorAbsorption    float64
}
```

## Hit Result

`internal/combat/hit_result.go`

`CalculateHit()` formula:

```
damage = attackStats.Attack - target.Defense/2 + hitbox.Damage
```

**Stagger levels**: flinch, knockdown, launch.

## Attack System

`internal/combat/attack.go`

- `AttackConfig` — defines an attack's properties
- `AttackNode` / `AttackGraph` — combo tree structure
- `Controller` — manages attack lifecycle:
  1. **Windup** — startup frames
  2. **Active** — hitbox active window
  3. **Recover** — recovery frames
  4. **Cooldown** — cooldown before next attack

## Hitbox / Hurtbox

- **Hitbox** (`internal/combat/hitbox.go`): rect + offset, damage, poise damage, stagger level. `Activate()`/`Deactivate()` lifecycle per attack phase.
- **Hurtbox** (`internal/combat/hurtbox.go`): rect + damage multiplier. `HurtboxSystem` manages collections per actor.

## Targeting

`internal/combat/target.go`

- `TargetSystem` registers actors
- `GetClosestEnemy(source)` — finds nearest hostile actor by distance

## Teams

`internal/combat/team.go`

```go
type Team int
const (
    TeamPlayer Team = iota
    TeamAlly
    TeamEnemy
    TeamNeutral
)
func (t Team) IsHostile(other Team) bool
```

## AI Controller

`internal/combat/ai.go`

- `AIController` with states: idle, approach, attack, retreat, staggered, dead
- State transitions based on distance to target, cooldowns, and random decisions
