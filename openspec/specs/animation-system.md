# Animation System — Madokita

## Overview

Frame-based 2D sprite animation system with phase timing support for fighting game attacks (windup/active/recover) and movement animations. Every frame composes **N sprites** (separate images drawn together — base body + weapon + halo) and M hurtboxes.

## Core Types

Two type systems coexist with different responsibilities:

| Package | Purpose |
|---------|---------|
| `internal/project/` | Editor domain model — data definitions, zero Ebitengine deps |
| `internal/animation/` | Runtime playback — Animator, Frame/FrameSprite, constructors |

### Editor Domain Model (`internal/project/types.go`)

```go
type FramePhase int

const (
    PhaseWindup  FramePhase = iota
    PhaseActive
    PhaseRecover
    PhaseArmed
)

type FrameSpriteEntry struct {
    SpriteIdx      int     // which sprite (image) of the project
    SpriteFrameIdx int     // which sub-frame of that sprite's sheet
    OffsetX        float64
    OffsetY        float64
    Rotation       float64
    ScaleX         float64
    ScaleY         float64
    OriginX        float64
    OriginY        float64
}

type AnimationFrame struct {
    Sprites   []FrameSpriteEntry
    Hurtboxes []HurtboxRow
    Phase     FramePhase
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
    X        float64
    Y        float64
    Width    float64
    Height   float64
    Rotation float64
    DmgMult  float64
}

type ProjectData struct {
    AssetName      string
    AssetKey       string
    DefaultOriginX float64
    DefaultOriginY float64
    Animations     []AnimationRow
    Sprites        []SpriteRow
    HitDefs        []HitboxRow
}
```

### Runtime Playback (`internal/animation/`)

- **Animator**: frame-based playback with FPS timing (`NewAnimator(def)`)
- **`FrameSprite`**: one sprite instance inside a frame — sheet index + sub-frame index plus its own transform (offset/rotation/scale/origin)
- **`Frame`**: `Sprites []FrameSprite` + `Hurtboxes []FrameHurtbox` — N sprites and M hurtboxes per frame
- Used by `entity/player/player.go` and `internal/debug/overlay.go`

## Animator

`internal/animation/animator.go`

- `Animator` struct plays frame-based animations
- FPS-based timing: advances frame each `1/FPS` seconds
- Supports phase progression for attacks: windup → active → recover → idle
- Used by `entity/player/player.go` and `internal/debug/overlay.go`

## Multi-Sprite Rendering

`entity/player/player.go`:

- `SetupAnim(def *animation.Movement, am *assets.AssetManager)` — the player holds the AssetManager; there is no pre-sliced frame array anymore
- `Draw()` iterates `frame.Sprites` and draws `am.GetFrame(def.AssetKey, s.SpriteIdx, s.SpriteFrameIdx)` with the entry's own transform
- **Origin per sprite**: when the entry's sheet exists in `def.Sprites`, origin = `s.OriginX * sheet.FrameW` / `s.OriginY * sheet.FrameH`; otherwise falls back to `DefaultOriginX/Y * FrameSize`
- **Footprint**: `footHeight()` uses sheet 0's `FrameH` (fallback `FrameSize` = 256) and `DefaultOriginY`

## Asset Loading

`Movement`/`Attack` declare their images in `Sprites []SpriteSheetDef`:

```go
type SpriteSheetDef struct {
    File       string // relative to assets/
    FrameW     int
    FrameH     int
    FrameCount int
}
```

`cmd/game/bootstrap.go` converts them to `assets.CharacterSheets` and calls `assetMgr.LoadCharacter(...)`; the player fetches frames with `GetFrame(charKey, sheetIdx, frameIdx)`. `assets.SliceFrames` cuts sheets into tiles (the old hand-rolled `loadFramesFromPNG` in bootstrap is gone).

**Convention**: `File` is relative to `assets/` (same as `StageDef.Images`), e.g. `sprites/players/sayaka_miki/sayaka_miki.png`.

## Movement Data Files

Movement data lives in `internal/data/characters/movements/*.go`. Each file defines one `animation.Movement` variable.

**Pattern**: dot-import `madokita/internal/animation` + constructors for zero-noise data definitions.

Available constructors (defined in `internal/animation/types.go`):

| Constructor | Signature | Purpose |
|-------------|-----------|---------|
| `Anim` | `(fps float64, loop bool, frames ...Frame)` | Builds `MovementAnimDef` |
| `F` | `(parts ...any)` | Builds `Frame`; accepts `S()` entries and `HB()`/`HBR()` hurtboxes mixed, in any order |
| `S` | `(spriteIdx, spriteFrameIdx int, rest ...float64)` | Builds `FrameSprite`; rest = offsetX, offsetY, rotation, scaleX, scaleY, originX, originY (defaults: 0, 0, 0, 1, 1, 0.5, 0.5) |
| `HB` | `(w, h, ox, oy float64)` | Builds `FrameHurtbox` with defaults (scale=1, rot=0, mult=1) |
| `HBR` | `(w, h, ox, oy, rot float64)` | Builds `FrameHurtbox` with custom rotation |
| `AttackF` | `(phase AttackPhase, sprites ...FrameSprite)` | Builds `AttackFrame` with one or more sprites (at least 1 required) |
| `AttackAnim` | `(fps float64, loop bool, windup, active, recover, armed, armedFPS float64, wuF, atkF, rcF int, frames ...AttackFrame)` | Builds `AttackAnimDef` with phase timings (incl. armed) and frame counts |

**Rules**:
- MUST use dot-import — movement files are pure data, no logic
- MUST use constructors, never struct literals (the `Sprites []SpriteSheetDef` block is the only allowed literal)
- MUST declare at least one `S(...)` per `F(...)`/`AttackF(...)` — sprite-less frames are rejected
- SHOULD keep frames inline (one `F(...)` per line)
- MAY define shared vars only when the same set repeats across many frames AND hurts readability less than repetition

**Example (`sayaka.go`):**

```go
package movements

import . "madokita/internal/animation"

var SayakaMovement = Movement{
    AssetKey:       "sayaka_movement",
    DefaultOriginX: 0.506,
    DefaultOriginY: 0.586,
    Sprites: []SpriteSheetDef{
        {File: "sprites/players/sayaka_miki/sayaka_miki.png", FrameW: 256, FrameH: 256, FrameCount: 25},
    },
    Animations: map[string]MovementAnimDef{
        "walk": Anim(10, true,
            F(S(0, 4, 0, 0, 0, 1, 1, 0.506, 0.586), HB(100, 57, 1, -32.5), HB(52, 130, 1, 61)),
            F(S(0, 5, 0, 0, 0, 1, 1, 0.506, 0.586), HB(100, 57, 1, -32.5), HB(52, 130, 1, 61)),
        ),
    },
}
```

Multi-sprite frame (base + weapon rendered together):

```go
F(S(0, 4, 0, 0, 0, 1, 1, 0.506, 0.586), S(1, 0, 30, -10, 0, 1, 1, 0.5, 0.5), HB(100, 57, 1, -32.5)),
```

The animprite editor generates exactly this format (`S` with all 9 args) and imports it back — edit data in the editor, never by hand.

## Character Registry

`internal/data/registry.go`

- Global `Registry` map: maps string names to `CharacterData`
- `CharacterData` contains: `Animations []animation.Movement`, `Attack *animation.Attack`, `Hurtboxes []combat.HurtboxConfig`, `AttackConfigs []combat.AttackConfig`, `AdditionalTextures`, `Effects`
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
op.GeoM.Translate(-originX, -originY)    // entry origin → scaled
op.GeoM.Translate(s.OffsetX, s.OffsetY)  // per-entry offset → scaled
// optional: op.GeoM.Rotate(s.Rotation)
op.GeoM.Scale(s.ScaleX*p.Scale*flip, s.ScaleY*p.Scale)   // scales everything above
op.GeoM.Translate(p.X, p.Y)              // world → NOT scaled
op.GeoM.Translate(-cameraX, 0)           // camera → NOT scaled
```

`originX/Y` are per-sprite (`s.OriginX * sheet.FrameW`), not global.

**Do NOT reorder** to the mathematically conventional chain (`S → R → T` from left to right). It will break because the origin subtraction and per-frame offset would end up unscaled while the world position would get incorrectly scaled.
