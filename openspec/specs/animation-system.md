# Animation System — Madokita

## Overview

Frame-based 2D sprite animation system with phase timing support for fighting game attacks (windup/active/recover) and movement animations.

## Core Types

Two type systems coexist with different responsibilities:

| Package | Purpose |
|---------|---------|
| `internal/project/` | Editor domain model — data definitions, zero Ebitengine deps |
| `internal/animation/` | Runtime playback — Animator, MultiSpriteAnimator |

### Editor Domain Model (`internal/project/types.go`)

```go
type FramePhase int

const (
    PhaseWindup  FramePhase = iota
    PhaseActive
    PhaseRecover
    PhaseArmed
)

type AnimationFrame struct {
    SpriteIdx      int
    SpriteFrameIdx int
    OffsetX        float64
    OffsetY        float64
    Rotation       float64
    ScaleX         float64
    ScaleY         float64
    OriginX        float64
    OriginY        float64
    Hurtboxes      []HurtboxRow
    Phase          FramePhase
}

type AnimationRow struct {
    Name       string
    CurrentIdx int
    Frames     []AnimationFrame

    Windup   float64
    Active   float64
    Recover  float64
    Armed    float64
    ArmedFPS float64

    FPS float64
}

type SpriteRow struct {
    Name       string
    File       string
    Width      int
    Height     int
    FrameCount int
    CurrentIdx int

    OffsetX  float64
    OffsetY  float64
    ScaleX   float64
    ScaleY   float64
    Rotation float64
    OriginX  float64
    OriginY  float64
}

type HurtboxRow struct {
    X          float64
    Y          float64
    Width      float64
    Height     float64
    Rotation   float64
    Damage     float64
    Multiplier float64
}

type HitboxRow struct {
    Width  float64
    Height float64
}

type ProjectData struct {
    Animations []AnimationRow
    Sprites    []SpriteRow
    HitDefs    []HitboxRow
}
```

### Runtime Playback (`internal/animation/`)

- **Animator**: frame-based playback with FPS timing
- **Types**: `Frame`, `Movement`, `Attack`, `MovementAnimDef`, `AttackAnimDef`
- Used by `entity/player/player.go` and `entity/enemy/enemy.go`

## Animator

`internal/animation/animator.go`

- `Animator` struct plays frame-based animations
- FPS-based timing: advances frame each `1/FPS` seconds
- Supports phase progression for attacks: windup → active → recover → idle
- Used by `entity/player/player.go` and `entity/enemy/enemy.go`

## Multi-Sprite Animation

- `MultiSpriteAnimator` supports N-sprite playback (for multi-part characters or effects)

## Character Registry

`internal/data/registry.go`

- Global `Registry` map: maps string names to `CharacterData`
- `CharacterData` contains: MovementAnimDefs, AttackAnimDefs, hurtboxes, attack configs, effects
- Also holds `StageData` (background, enemies, base speed)

## Rendering Caveats

### Ebitengine v2 GeoM.Translate Behavior

`GeoM.Translate()` does NOT perform a proper matrix multiplication — it **only adds** to `tx`/`ty`. `GeoM.Scale()` DOES multiply all components including `tx`/`ty`.

This means the **order of method calls** is critical:

| Call before Scale | Gets multiplied by scale |
|-------------------|-------------------------|
| Call after Scale  | Stays unscaled (just added) |

**Correct order for sprite rendering in `entity/player/player.go`:**

```go
op.GeoM.Translate(-originX, -originY)    // origin → scaled
op.GeoM.Translate(ox, oy)                // per-frame offset → scaled
// optional: op.GeoM.Rotate(rot)
op.GeoM.Scale(sx*Scale*flip, sy*Scale)   // scales everything above
op.GeoM.Translate(p.X, p.Y)              // world → NOT scaled
op.GeoM.Translate(-cameraX, 0)           // camera → NOT scaled
```

**Do NOT reorder** to the mathematically conventional chain (`S → R → T` from left to right). It will break because the origin subtraction and per-frame offset would end up unscaled while the world position would get incorrectly scaled.
