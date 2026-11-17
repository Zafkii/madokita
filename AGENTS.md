# Madokita Gems — AI Project Guide

## Overview

Two-in-one Go project: **Madokita Gems** (fighting game) + **animprite** (animation editor).

- **Stack**: Go 1.26.4, Ebitengine/v2 v2.9.9
- **Architecture**: ECS-influenced, scene-based, phase-driven runtime
- **Editor**: Standalone Ebitengine app with immediate-mode UI (independent module)

---

## Quick Reference

| Action           | Command                                                            |
| ---------------- | ------------------------------------------------------------------ |
| Build game       | `go build -o ./tmp/game.exe ./cmd/game`                            |
| Run game         | `go run ./cmd/game`                                                |
| Hot-reload game  | `air` (from root)                                                  |
| Build editor     | `go build -o ./tmp/animprite.exe ./cmd/editor` (from `animprite/`) |
| Run editor       | `go run ./cmd/editor` (from `animprite/`)                          |
| Run tests        | `go test ./...`                                                    |
| Run editor tests | `go test ./...` (from `animprite/`)                                |

---

## Project Structure

```
madokita/
├── cmd/game/          # Game entry point: main.go, bootstrap.go
├── internal/          # Game engine packages
│   ├── animation/     # Frame-based animation playback (phases: windup/active/recover/idle)
│   ├── assets/        # ICO loader
│   ├── audio/         # OGG Vorbis playback via oto/v3
│   ├── combat/        # Fighting system: collision, actors, AI, attacks, hitboxes, hurtboxes
│   ├── data/          # Registry: characters, stages, enemies (data definitions)
│   ├── database/      # SQLite persistence (modernc.org/sqlite)
│   ├── ecs/           # Minimal ECS: Entity, Component, World, System
│   ├── engine/        # Phase-driven Runtime, SystemManager, DI Container
│   ├── entity/        # Concrete entities: Player, Enemy, Effects
│   ├── event/         # Typed pub/sub EventBus
│   ├── input/         # Input Manager (action → key bindings)
│   ├── localization/  # Translation key→string map
│   ├── math/          # Vec2, Rect
│   ├── menu/          # Intro, MainMenu, Settings scenes
│   ├── platform/      # Platform API abstraction (desktop stub)
│   ├── save/          # Save/load system (SQLite-backed, multi-slot)
│   ├── scene/         # Scene Manager, Scene Interface, Fade transitions
│   ├── settings/      # Persistent settings (SQLite-backed)
│   ├── ui/            # Shared UI widgets (Button, Label, ImageCache)
│   └── windrag/       # Window dragging (win32/linux stubs)
│
├── animprite/          # Animation editor (standalone module)
│   ├── cmd/editor/    # Editor entry point, EditorApp (implements ebiten.Game)
│   ├── internal/
│   │   ├── camera/    # Pan/zoom camera (world↔canvas transforms)
│   │   ├── canvas/    # Viewport with grid, origin crosshair, boundary
│   │   ├── filedialog/ # Native file open/save dialogs (win32/linux)
│   │   ├── fonts/     # Embedded PxPlus IBM VGA 8x14
│   │   ├── project/   # Domain model: AnimationRow, SpriteRow, ProjectData, DeepCopy (zero Ebitengine dep)
│   │   ├── theme/     # Dark/Light palette system (29 color fields)
│   │   ├── ui/        # Editor UI widgets (Button, Dropdown, TextInput, Slider, Table)
│   └── ... (module-level go.mod, air.toml, Makefile)
│
├── assets/            # Game assets (images, audio, etc.)
├── openspec/          # Specification-Driven Development artifacts
│   ├── config.yaml    # SDD project config
│   ├── changes/       # Change proposals, specs, designs
│   └── specs/         # Current-state specs
└── reference/         # Design references from TS/Phaser migration
```

---

## Engine Architecture

### Phase-Driven Runtime (`internal/engine/`)

```
PhasePreInit → PhaseInit → PhasePostInit → PhaseReady → PhaseRunning → PhaseShutdown
```

- **Runtime** owns the phase state machine and the DI Container
- **SystemManager** runs registered `System` implementations through lifecycle phases
- Systems initialized/started/readied in **registration order**, shutdown in **reverse order**
- Container is a simple string-keyed service locator (`Register(key, svc)` / `Get(key)`)

### DI Container (`internal/engine/container.go`)

```go
container.Register("events", eventBus)
container.Register("input", inputMgr)
container.Register("platform", platformAPI)
```

### Event Bus (`internal/event/bus.go`)

Typed pub/sub with auto-unsubscribe closures:

```go
bus.On("player.hit", func(payload any) { ... })  // returns unsubscribe func
bus.Emit("player.hit", hitResult)
```

### Scene Manager (`internal/scene/manager.go`)

- Scenes implement `Interface` (Update/Draw/Enter/Exit/Pause/Resume/IsActive)
- Manager owns a map of named scenes, current scene, fade transition
- **Fixed-canvas rendering**: always renders to 1280x720, then scales to window size
- `SwitchWithFade(name, duration)` handles fade-out → scene swap → fade-in

### System Interface (`internal/engine/system.go`)

```go
type System interface {
    ID() string
    Initialize() error
    Start() error
    Ready() error
    Shutdown() error
}
```

---

## Combat System (`internal/combat/`)

| Package                    | Purpose                                                                      |
| -------------------------- | ---------------------------------------------------------------------------- |
| `collision.go`             | Frame overlap detection (hitbox↔hurtbox), team filtering                     |
| `actor.go`                 | CombatActor: invincibility, stagger, hyper-armor, `ReceiveHit()`             |
| `attack.go`                | Attack lifecycle (windup→active→recover→cooldown), combo trees (AttackGraph) |
| `hit_result.go`            | Damage formula: attack - defense/2 + damage, stagger levels                  |
| `hitbox.go` / `hurtbox.go` | Hitbox/Hurtbox with rect, damage, lifecycle                                  |
| `stats.go`                 | Stats: health, stamina, poise, defense, regen                                |
| `target.go`                | TargetSystem: closest enemy by distance                                      |
| `team.go`                  | Team enum (Player/Ally/Enemy/Neutral), IsHostile()                           |
| `ai.go`                    | AIController: idle/approach/attack/retreat/staggered/dead states             |

### Attack Commitment (player)

The player is locked by phases, derived from `AttackAnimator.Phase()` in `Player.updateAttacking()`:

| Phase    | Move | Re-attack |
| -------- | ---- | --------- |
| Windup   | ✗    | ✗         |
| Active   | ✗    | ✗         |
| Recover  | ✓    | ✗         |
| Armed    | ✓    | ✓ (cancels the pose) |

- `IsAttacking` = attack animation playing (drives `Draw`); `IsMovementLocked`/`IsAnimationLocked` drive input, set on `TryAttack` and per-phase in `updateAttacking`
- Moving cancels the attack from recover on (like the reference TS `cancelGuard`); the armed pose is NOT called "guard" — that concept is reserved for future parries
- `Controller.Start()` has no cooldown gate: the commitment phases pace attacks
- Gravity runs unconditionally in `Player.Update` (not gated by `IsMovementLocked`): attacking mid-air must not freeze the fall

---

## Animation System (`internal/animation/`)

- **Animator**: frame-based playback with FPS timing (`NewAnimator(def)`)
- **Types**: `Frame`, `FrameSprite`, `FrameHurtbox`, `FrameHitbox`, `Movement`, `Attack`, `MovementAnimDef`, `AttackAnimDef`, `SpriteSheetDef`
- Each `Frame` composes **N sprites** (`Sprites []FrameSprite`) + M hurtboxes (`Hurtboxes []FrameHurtbox`); each `AttackFrame` composes N sprites + hurtboxes + hitboxes (`Hitboxes []FrameHitbox`)
- Frame phases: windup/active/recover/armed (for attacks)
- `Movement`/`Attack` declare their images in `Sprites []SpriteSheetDef` (File relative to `assets/`, FrameW/H/Count); bootstrap converts them via `assets.LoadCharacter` and the player fetches frames with `GetFrame(charKey, sheetIdx, frameIdx)`

---

## Editor (`animprite/`)

### Architecture

- `EditorApp` implements `ebiten.Game` (Update/Draw/Layout)
- **Immediate-mode UI**: widgets painted every frame with vector primitives
- No widget hierarchy — flat dispatch in `Update()` (split across 2 files: `update.go` orchestrator, `update_canvas.go` interaction), draw order in `Draw()`
- **`internal/project/`** — domain model (`ProjectData`, `AnimationRow`, `SpriteRow`, etc.) in a zero-dependency package, testable without Ebitengine

### Layout

```
+----------------------------------------------------------+
| Mode Indicator (22px)                                    |
+----------------------------------------------------------+
| Top Panel: ModeDropdown [Open] [Save] [Dark]             |
| 4 Tables (Animation | Sprite | Hurtbox | Hitbox)         |
+----------------------------------------+-----------------+
|                                        | Right Panel     |
|          CANVAS                        | ├─ Selected Elem|
|          (grid, origin,                | ├─ Base Sprite  |
|           boundary)                    | ├─ Animation    |
|          Pan: left drag                | └─ Preview      |
|          Zoom: scroll wheel            |                 |
+----------------------------------------+-----------------+
| Status Bar: Zoom: X%  [Reset View]                      |
+----------------------------------------------------------+
```

### Current State (visual layer complete, functionality pending)

| Feature                                             | Status            |
| --------------------------------------------------- | ----------------- |
| Native window chrome (OS title bar)                 | ✅ Done           |
| Window pref persistence (JSON)                      | ✅ Done           |
| Camera (pan, zoom, reset)                           | ✅ Done           |
| Canvas (grid, crosshair, boundary)                  | ✅ Done           |
| Dark/Light theme toggle                             | ✅ Done           |
| 4 data tables (Animation, Sprite, Hurtbox, Hitbox)  | ✅ Done           |
| Right panel inputs (properties, timing, preview)    | ✅ Done           |
| Preview Play/Loop/Speed slider                      | ✅ Done (UI only) |
| Tab navigation between inputs                       | ✅ Done           |
| Mode toggle (Movement/Attack)                       | ✅ Done           |
| **Sprite loading from disk**                        | ❌ Pending        |
| **Sprite rendering on canvas**                      | ✅ Done           |
| **Hitbox/Hurtbox canvas interaction**               | ❌ Pending        |
| **Animation preview playback**                      | ❌ Pending        |
| **Per-frame sprite frame selection**                | ✅ Done           |
| **Table exclusive selection**                       | ✅ Done           |
| **New frame inherits sprite + frame**               | ✅ Done           |
| **Strict per-layer stage editing**                  | ✅ Done           |
| **Per-object Z depth inside a layer**               | ✅ Done           |
| **Click cycles through overlapped objects**         | ✅ Done           |
| **Cut/Paste stage objects**                         | ✅ Done           |
| **Open/Save project files**                         | ❌ Pending        |
| **Undo/Redo**                                       | ❌ Pending        |

## Stage Layer Editing

An active layer restricts hit-testing, selection, marquee, select-all, drag, delete, duplicate, and collision edits to that layer. The General view edits every layer. Cut/Paste moves selected objects into the active layer, or into the top layer while in General view; the clipboard is cleared when a stage is loaded.

## Stage Info Tab

The stage toolbar has Build/Edit/Info tabs. A canvas edit (drag, scale, rotate) and a focused property field open Info automatically and hand the toolbar back to Edit once the edit has been idle for a moment. Picking the **Info** tab by hand pins it: selection changes, property edits and the idle timer leave the tab alone, because a manual visit means the user wants to keep tuning values. The pin clears as soon as another tab is picked, and automatic opens never set it.

## Stage Edit Panels (Opacity / Scale)

The Edit tab's action row carries **Opacity** and **Scale** buttons beside the create-object buttons. Each one toggles its own panel under the button row; pressing the same button again closes it, and opening one closes the other.

- **Opacity** panel shows a slider plus a value field. The field also stays in the Info tab, but the Info tab no longer paints the slider: dragging opacity belongs to Edit, while reading or typing an exact value stays in Info.
- **Scale** panel resizes the selection as a percentage of the size it had when the drag started: `applyStageScale` captures a `scaleBase` snapshot of the selected objects on the first step and every later slider move multiplies that snapshot, so dragging back and forth never compounds. `resetStageScaleBurst` drops the snapshot (and rewinds the slider/field to 100) whenever the selection changes, the panel toggles, or the drag goes idle. `releaseStageSliders` hands the button release to every stage slider, because a slider that never receives it keeps following the pointer and its idle guards (which check `Dragging()`) never close the undo or scale burst. Width and height never drop below 1px, and one undo entry covers the whole drag.

Editing from a panel never switches tabs: dragging its slider or typing in its field keeps the toolbar in Edit. Canvas drags and the Info property fields still open Info automatically, so `stageEditPanelOwnsInput` suppresses that only for a field painted by the open panel.

Because the panels own their widgets, the shared hide pass (`hideStageObjectInputs`) hides the opacity slider, the scale slider and the scale field on every other path, so no stale control can be clicked outside its panel.

## Stage Depth (Z)

Each stage object carries a `Z` depth edited in the Info tab next to Rotation and Opacity. Layers still draw in slice order, so Z only orders objects **inside their own layer**: a larger Z draws closer to the camera, and at equal Z the category order is kept (backgrounds, static and animated props, static and animated blocks). Default Z is 0, so existing stages render exactly as before.

One ordering rule feeds both renderers so the editor previews what the game paints: `StageData.RenderOrder()` (`animprite/internal/project/stage.go`) and `stage.DrawOrder()` (`internal/stage/stage.go`) return object indices sorted by layer, then Z, then category rank, then slice position. Rendering walks the list forwards and picking walks it backwards, so the topmost object receives the click. Start markers sort last inside their layer (they never paint) which keeps them pickable ahead of the artwork. Z is display only: collision and hurtboxes ignore it, and the export writes `Z` only when it is not zero.

## Cycling Overlapped Objects

`stagePickCandidates()` (`animprite/internal/editor/stage.go`) returns every object under the pointer front to back, and the press resolution is two steps. The **first** click on a spot uses `stagePressTarget()`: a press inside the current selection keeps it (so a freshly placed object still drags instead of stealing whatever sits above it), and any other press picks the front-most object under the pointer. Clicking the **same spot again** (within `stagePickRepeatSlop` screen pixels) steps through `stageCycleTarget()` to the object behind the selected one and wraps around at the back of the stack, so several PNGs dropped on one point are reachable one click at a time like the Geometry Dash editor. The repeat spot is armed on the press that hits objects and, while the button is up, dropped as soon as the pointer moves off it (`stageClearPickAnchorIfPointerMoved`), so leaving the spot and coming back starts over at the front object.

Picking is a plain rectangle hit test (`stageObjectContainsWorld`): the object's full box picks, transparent padding included, so a big tree sprite can still be grabbed by its empty area the way it always could.

## Per-Frame Sprite Composition

Each `AnimationFrame` stores `Sprites []FrameSpriteEntry` — one entry per sprite resource rendered in that frame. Each `FrameSpriteEntry` has `SpriteIdx` (which sprite resource) and `SpriteFrameIdx` (which sub-frame of that spritesheet) plus its own Offset/Rotation/Scale/Origin. The ◀▶ buttons in the sprite table, when the right panel is in animation frame mode, write to the matching `FrameSpriteEntry` of the active frame instead of the global `SpriteRow.CurrentIdx`. New animation frames inherit all sprite entries (and hurtboxes) from the previous frame.

When switching frames via `loadAnimFrameProps()`, `SpriteRow.CurrentIdx` is synced from the frame's entry so the sprite table display matches.

**Structs**: `AnimationFrame.Sprites []FrameSpriteEntry` (`internal/project/types.go`), `FrameSpriteEntry` (`internal/project/types.go`)

**Files**: `internal/project/types.go` (fields), `sync.go` (◀▶ buttons via `frameSpriteEntry()`), `selection.go` (frame load, `ensureFrameSprites`), `render.go` (canvas render of all entries)

## Attack Hitboxes

An attack's hitboxes live on the frames that carry them, exactly like hurtboxes: `AnimationFrame.Hitboxes []HitboxRow` with the four values X, Y, Width and Height and nothing else (no rotation, no damage multiplier). The Hitbox table in the top panel shows the frame you are editing, its `+` seeds a 32x32 box, `-` removes the selected one and the copy button clones the previous frame's boxes into an empty frame. Clicking a row opens X / Y / Width / Height in the right panel.

`HitDefs` no longer exists on `ProjectData`: a global box list could not say which frame or animation owned it, so nothing ever reached the game. Adding a frame clones the previous frame's hitboxes, and every frame switch rebuilds the hitbox table rows, so the inputs can never show the previous frame's numbers (or index past a shorter list).

The export writes them inside the frame, so they survive a save:

```go
AttackF(PhaseActive,
	S(0, 8, 0, 0, 0, 1, 1, 0.5, 0.5),
	HB(95, 61, 1.5, -32.5),
	HIT(40, 30, 60, -20),
),
```

The editor paints them on the canvas in red (the selected one filled) through the same box overlay list as hurtboxes: every entry carries a `Kind` (`RenderHurtbox`, `RenderHitbox`, `RenderCollision`) that picks the palette. Holding CapsLock and dragging the canvas moves the selected hitbox with the hurtbox math, and Ctrl-dragging a corner handle resizes it (the handle snapshots the box, so one undo covers the whole drag). The frame's rotation applies to the drawn box even though a hitbox itself has none, which is also why a selected hitbox shows no rotation handle: there is no angle to store. Picking a hitbox row clears the hurtbox row and vice versa, so the shared selection always names one box.

`HB` writes its rotation and damage multiplier only when they differ from 0 and 1 (`HB(20, 20, -4, -10, 30, 1.5)`); `HIT` always takes exactly four values. The importer accepts `S`, `HB`, `HBR` and `HIT` in any order and rejects anything else. On the game side `animation.AttackFrame` carries `Hurtboxes` and `Hitboxes`, and `Player.CurrentAttackHitboxes()` returns the playing frame's authored boxes in world space (mirrored by `FlipX`). The frames own the geometry: every attack frame pushes its authored boxes into `combat.Hitbox` through `Controller.SetGeometry` (`Hitbox.Rects` returns one rect per authored box, and falls back to the config geometry only when a frame carries none), so `Player.ActiveHitboxes()` is exactly what the editor drew. `combat.AttackConfig` keeps what the editor cannot express — damage, poise damage, stagger level, timing, stamina — and must leave `HitboxConfig.W/H/OffsetX/OffsetY` at zero: a box there would draw a second, duplicate hitbox next to the authored one. With the debug overlay on (F1) `debug.DrawHitboxes` paints that single set in red, and `debug.DrawHurtboxes` switches to the attack frame's authored hurtboxes while `State.IsAttacking`, because the movement animator is frozen on its last frame during an attack and would otherwise draw a static body box. `internal/data/characters/attacks/sayaka.go` mirrors the editor export mechanically (see "Exporting a Stage to the Game" for the transform): the frame block is copied verbatim and only `AssetKey`, the animation name and the `Sprites` paths are rewritten.

## Table Exclusive Selection

Only one table row is highlighted at a time. Clicking any row in Animation, Sprite, Hurtbox, or Hitbox tables deselects all others. This also applies to the "+ Add" hurtbox button.

No more multiple highlighted rows — each click selects exactly one row across all four tables.

**Files**: `update.go:handleTopPanelMouse()`, `app.go:addHbBtn.OnClick`

---

## Movement Data Pattern

Movement data files (`internal/data/characters/movements/*.go`) use **dot-import + constructors** for compact, readable definitions:

```go
package movements

import . "madokita-gems/internal/animation"

var SayakaMovement = Movement{
    AssetKey:       "sayaka_movement",
    DefaultOriginX: 0.506,
    DefaultOriginY: 0.5,
    Sprites: []SpriteSheetDef{
        {File: "sprites/players/sayaka_miki/sayaka_miki.png", FrameW: 256, FrameH: 256, FrameCount: 25},
    },
    Animations: map[string]MovementAnimDef{
        "idle": Anim(3, true,
            F(S(0, 0, 0, 0, 0, 1, 1, 0.506, 0.5), HB(95, 61, 1.5, -32.5), HB(54, 130, 4, 62)),
        ),
    },
}
```

| Constructor | Args | Purpose |
|-------------|------|---------|
| `Anim(fps, loop, frames...)` | FPS, loop flag, variadic frames | Wraps `MovementAnimDef` |
| `F(parts ...any)` | Mixed `S()` sprites and `HB()`/`HBR()` hurtboxes, any order | Wraps `Frame` |
| `S(spriteIdx, spriteFrameIdx, rest ...float64)` | Sheet index, sub-frame index, then offsetX, offsetY, rotation, scaleX, scaleY, originX, originY | Wraps `FrameSprite` (defaults: 0, 0, 0, 1, 1, 0.5, 0.5) |
| `HB(w, h, ox, oy, rot?, dmg?)` | Width, Height, OffsetX, OffsetY, then optional rotation and damage multiplier | Wraps `FrameHurtbox` (scale=1, rot=0, mult=1) |
| `HBR(w, h, ox, oy, rot)` | Same as `HB` with a rotation, kept as its own name for readability | `FrameHurtbox` with custom rotation |
| `HIT(w, h, ox, oy)` | Width, Height, OffsetX, OffsetY | Wraps `FrameHitbox` (attack damage box, no rotation) |
| `AttackF(phase, parts ...any)` | Phase constant + mixed `S()` sprites and `HB()`/`HBR()`/`HIT()` boxes, any order (at least one `S()`) | Wraps `AttackFrame` |
| `AttackAnim(fps, loop, wu, atk, rc, armed, armedFPS, wuF, atkF, rcF, frames...)` | FPS, loop, phase timings (incl. armed) + frame counts, variadic frames | Wraps `AttackAnimDef` |

**Rules:**
- Always use dot-import in movement/attack data files (zero logic, pure data)
- Never use struct literals — always the constructors (the `Sprites: []SpriteSheetDef{...}` block is the only allowed literal)
- Keep everything inline (no extracted vars), each frame is one `F(...)`/`AttackF(...)` call
- Multi-sprite frames: one `S(...)` per image rendered in that frame
- Attack hitboxes are authored per frame: `AttackF(PhaseActive, S(...), HIT(40, 30, 60, -20))`. The four values are width, height, offsetX, offsetY, in the same pivot-relative space as hurtboxes
- Every `F(...)` and `AttackF(...)` must declare at least one `S(...)` — the editor rejects sprite-less frames
- The editor only imports/exports this constructor format; struct-literal frames (`AttackFrame{...}`, legacy `OffsetX: [...]` fields) are rejected

## Key Patterns & Conventions

- **Error handling**: Go standard (`if err != nil { return err }`), no panics in production paths
- **Theme system**: All UI elements read `theme.Manager.Current` for colors (hot-swappable)
- **Widget lifecycle**: Widgets have `Visible`/`Enabled` flags — no reconstruction on toggle
- **Multiple modules**: Game (`madokita-gems` module) and editor (`animprite` module) are independent Go modules sharing no code
- **Native window chrome**: Both apps use the OS-native title bar/decoration (no custom-drawn chrome); the editor keeps window position/size in JSON prefs
- **Settings persistence**: Game uses SQLite; editor uses JSON (simpler, no dependency)

## Spawn System

Each `StageDef` defines `GroundY` (logical ground line Y) and optional `SpawnX`/`SpawnY` (player origin). The stage editor can place one **Start Pos** marker (`Category "start"`, no image); `NewStageDef` derives `SpawnX`/`SpawnY` from that marker. Without a marker, `SpawnY` falls back to `GroundY`. When `GameScene.SetPlayer(p)` is called, it invokes `p.SpawnAt(stage.SpawnX, spawnY, stage.GroundY)`. A marker above the ground line starts airborne; one below it is clamped to the ground line.

There is **no automatic visual foot offset**. The game must resolve ground contact through collision boxes (or place the visual floor independently); never derive `p.Y` from `FrameSize`/`DefaultOriginY`.

## Exporting a Stage to the Game

The editor exports a Go file into `animprite/tmp/wea/<stage>/<stage>.go`; importing it into `internal/game/stages/<stage>.go` is a mechanical transform, never a hand edit:

- Coordinates come from the editor (origin at the stage centre), so add `StageWidth/2` and `StageHeight/2` to every object `X`/`Y` to move them to the runtime origin (top-left). Collision boxes, hitboxes and animation offsets are object-relative and stay untouched.
- Image paths are relative to the stage directory; prefix them with `../../../assets/images/stages/<stage>/` so `runtimeStageAssetPath` (`internal/game/stages.go`) strips everything up to `/assets/` and the game resolves them against the `assets` base. The start marker keeps `File: ""` and `AnimFile` stays a bare key (`tree_moves/moves.go`), not a path.
- Keep the `package stages` + dot-import header; verify object count, that every referenced image exists, and that the previous and new exports still agree outside the region that was actually edited.

## Stage Collision

`GameScene.solidRects()` converts layout geometry into world-space collision rectangles. Blocks (`block`, `anim_block`) without authored boxes use their rendered `Width`/`Height`; props collide only when they have explicit `CollisionBoxes`; backgrounds, start markers, and `CollisionDisabled` objects are ignored. An axis-aligned box (object and box rotation at a multiple of 90°) stays one exact rectangle. A rotated box is decomposed by `stage.Object.SolidRects()` into vertical strips `SolidStripWidth` (6px) wide, each covering the rotated box's vertical extent in that slice, so a slope is walkable instead of blocking with its full axis-aligned bounding box.

The player owns a solid body derived from the standing hurtbox union (not from the moving damage hurtboxes). Ground contact is resolved at a probe (`groundProbeX` in `internal/entity/player/player.go`). Descending, the probe is the body centre pushed a quarter of the body width toward the walk direction, so three quarters of the base stand on the collision and the leading quarter may hang over the drop. Ascending *while grounded*, the probe moves to the trailing edge instead: the ground ahead is higher, so the base rides the lower ground behind it and the rise buries the front of the base instead of leaving the whole body airborne. In the air the probe never moves behind the body: a fall must meet the floor it is moving toward, because a probe over the trailing ground would let the feet drop through a step whose face the base already straddles. Never anchor a climb on the leading edge: that touches the slope with one corner and floats the rest of the base. Support, landing, small-drop snapping and step-up all resolve at that probe: the player climbs to the probe's floor, follows drops of up to 12px (`maxStepDown`) without leaving the ground, and treats only geometry rising more than 16px (`maxAutoStepHeight`) above the ground beside it (not above the feet, which ride low on a climb) as a wall. Solids behind the probe are ignored so the base may sink into a slope. The walkable-step exemptions are grounded-only: in the air nothing lifts the base, so every overlapping solid is a wall and a jump can never insert the body into a block or slide through a stair's face. A body ejected from a wall leaves through the nearer face whenever the overlap was inherited (a turn, a landing, a fall); only an overlap shallow enough for the current move to have made it follows the travel direction, so turning inside a face can never eject the body across the whole block. Landing commits across the middle half of the base — the centre and both quarter points — so turning mid-fall never drops a floor the body already hangs over; with no travel direction only the centre counts, and a base whose centre is still outside a step does not land on it, it is pushed out of the step's face and slides down beside it. The band also covers a floor the fall has already buried the feet in by up to `maxAutoStepHeight`, so the slope strips at the front of the base are ridden instead of dropped through. Horizontal movement still pushes out of walls and gravity resolves ceilings, and the same resolution (`resolveSolids`) runs right after every airborne gravity step, so a fall can never leave the body inside a block just because no input moved it that frame. Walking off a ledge clears `IsGrounded` only once less than `ledgeSupport` (2px) of the **base footprint** still overlaps a floor at feet level (`baseHasSupport`): the probe is a single point ahead of the centre, so on its own it would release with most of the body still standing on the collision. The player therefore walks to the very edge of the block before dropping, and the jump stays available there; descending a stair releases on the same rule. A walkable drop (up to `maxStepDown`) is still followed immediately at the probe, which is what keeps a stair-stepped slope smooth. Without authored block geometry the logical `StageGroundY` line remains the fallback. With the debug overlay on (F1 in game) every object's authored collision boxes are outlined in blue, the strips physics resolves against are filled in the same blue, and authored object hitboxes render in red. `StageGroundY` remains the fallback line for stages with no block geometry. Hurtboxes remain damage-only and never act as stage collision.

## Ebitengine v2 GeoM Caveat

`GeoM.Translate()` does NOT perform a proper matrix multiplication — it **only adds** to `tx`/`ty`. In contrast, `Scale()` and `Rotate()` multiply ALL matrix components **including** `tx`/`ty` (proper left-multiplication).

This means the **order of method calls matters**:

- Any `Translate` call BEFORE `Scale`/`Rotate` gets transformed by them (scaled, and ROTATED — `Rotate` rotates the accumulated translation vector too, so an offset placed before `Rotate` orbits around the origin as the sprite spins)
- Any `Translate` call AFTER `Scale`/`Rotate` stays raw (just added)

**Correct GeoM order for sprite rendering** (rotation about origin, offset NOT rotated, offset scaled by character scale and mirrored when flipped — same as the editor):

```go
op.GeoM.Translate(-originX, -originY)            // origin → scaled/rotated
op.GeoM.Scale(sx*Scale*flip, sy*Scale)           // scale about origin (before rotation)
op.GeoM.Rotate(rot)                              // rotation about origin
op.GeoM.Translate(ox*Scale*flip, oy*Scale)       // offset → manually scaled+flipped, NOT rotated
op.GeoM.Translate(p.X, p.Y)                      // world → NOT scaled
op.GeoM.Translate(-cameraX, 0)                   // camera → NOT scaled
```

Do NOT put the offset `Translate` BEFORE `Rotate` — the offset would orbit around the player (the weapon-position bug in attack active/recover phases). Do NOT reorder to the mathematically intuitive chain (e.g., `S → R → T` from left to right) — Ebitengine v2's `Translate` behavior breaks it.

**Flipped sprites**: when `FlipX`, NEGATE the rotation angle (`rot = -rot` before `Rotate`). Because `R(θ)·S(-1) = M·R(-θ)`, rotating the mirrored sprite by `+θ` produces the wrong orientation (`M·R(θ)` is the true mirror) — rotated weapon frames point the wrong way when facing left. The offset is mirrored separately (`ox = -ox`), which is already correct.

## Scene Debug & Setup Map

`cmd/game/bootstrap.go` has a `sceneSetups map[string]func() error` in `Start()` that runs pre-setup before switching to a scene. Add entries here for any scene needing asset preloading:

```go
sceneSetups := map[string]func() error{
    "game-teststage": b.setupGameScene,
    "menu-intro":     b.setupMenuAssets,
    "main-menu":      b.setupMenuAssets,
}
```

Use `SCENE=<name>` env var to skip directly to any scene. Scenes not in the map (e.g. `settings`) switch immediately with no setup.

## Audio Lifecycle

Each scene owns its audio in `Enter()`/`Exit()`. `GameScene.Enter()` stops `"menu-theme"`; menu scenes start it in their `Enter()`. No external audio management in callbacks or bootstrap — each scene self-manages.
