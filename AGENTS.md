# Madokita — AI Project Guide

## Overview

Two-in-one Go project: **Madokita** (fighting game) + **animprite** (animation editor).

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
│   │   ├── fonts/     # Embedded NotoSansNerdFont TTF
│   │   ├── project/   # Domain model: AnimationRow, SpriteRow, ProjectData, DeepCopy (zero Ebitengine dep)
│   │   ├── theme/     # Dark/Light palette system (29 color fields)
│   │   ├── ui/        # Editor UI widgets (Button, Dropdown, TextInput, Slider, Table)
│   │   └── windrag/   # Window dragging (same pattern as game)
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

---

## Animation System (`internal/animation/`)

- **Animator**: frame-based playback with FPS timing
- **Types**: `Frame`, `Movement`, `Attack`, `MovementAnimDef`, `AttackAnimDef`
- Frame phases: windup/active/recover/idle (for attacks)
- **MultiSpriteAnimator**: N-sprite playback support

---

## Editor (`animprite/`)

### Architecture

- `EditorApp` implements `ebiten.Game` (Update/Draw/Layout)
- **Immediate-mode UI**: widgets painted every frame with vector primitives
- No widget hierarchy — flat dispatch in `Update()` (split across 3 files: `update.go` orchestrator, `update_window.go` chrome, `update_canvas.go` interaction), draw order in `Draw()`
- **`internal/project/`** — domain model (`ProjectData`, `AnimationRow`, `SpriteRow`, etc.) in a zero-dependency package, testable without Ebitengine

### Layout

```
+----------------------------------------------------------+
| Title Bar (28px)                              [_][□][X]  |
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
| Window chrome (drag/resize/minimize/maximize/close) | ✅ Done           |
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
| **Open/Save project files**                         | ❌ Pending        |
| **Undo/Redo**                                       | ❌ Pending        |

## Per-Frame Sprite Frame Selection

Each `AnimationFrame` stores both `SpriteIdx` (which sprite resource) and `SpriteFrameIdx` (which sub-frame of the spritesheet). The ◀▶ buttons in the sprite table, when the right panel is in animation frame mode, write to the active `AnimationFrame` instead of the global `SpriteRow.CurrentIdx`. New animation frames inherit `SpriteFrameIdx` from the previous frame.

When switching frames via `loadCurrentFrameProps()`, `SpriteRow.CurrentIdx` is synced from `AnimationFrame.SpriteFrameIdx` so the sprite table display matches.

**Structs**: `AnimationFrame.SpriteFrameIdx` (`internal/project/types.go`), `AnimationFrame.SpriteIdx` (`internal/project/types.go`)

**Files**: `internal/project/types.go` (field), `app.go` (nav buttons, new-frame inheritance), `draw.go` (render), `update.go` (sync on frame load)

## Table Exclusive Selection

Only one table row is highlighted at a time. Clicking any row in Animation, Sprite, Hurtbox, or Hitbox tables deselects all others. This also applies to the "+ Add" hurtbox button.

No more multiple highlighted rows — each click selects exactly one row across all four tables.

**Files**: `update.go:handleTopPanelMouse()`, `app.go:addHbBtn.OnClick`

---

## Key Patterns & Conventions

- **Error handling**: Go standard (`if err != nil { return err }`), no panics in production paths
- **Theme system**: All UI elements read `theme.Manager.Current` for colors (hot-swappable)
- **Widget lifecycle**: Widgets have `Visible`/`Enabled` flags — no reconstruction on toggle
- **Multiple modules**: Game (`madokita` module) and editor (`animprite` module) are independent Go modules sharing no code
- **Window dragging**: Same pattern duplicated in both modules (win32 API for cursor position)
- **Settings persistence**: Game uses SQLite; editor uses JSON (simpler, no dependency)

## Spawn System

Each `StageDef` defines `GroundY` (ground line Y) and `SpawnX` (horizontal spawn point). When `GameScene.SetPlayer(p)` is called, it automatically invokes `p.Spawn(stage.SpawnX, stage.GroundY)`. The `Player.Spawn()` method computes `footOffset = FrameSize * (1 - DefaultOriginY) * Scale` and positions the player so feet touch the ground.

**No manual Y calculation needed.** Never use `p.Y = GroundY - (FrameSize/2)*Scale` — that formula is wrong for characters with a non-0.5 DefaultOriginY. Always let `Spawn()` handle it.

## Ebitengine v2 GeoM Caveat

`GeoM.Translate()` does NOT perform a proper matrix multiplication — it **only adds** to `tx`/`ty`. In contrast, `Scale()` multiplies ALL matrix components including `tx`/`ty`.

This means the **order of method calls matters**:

- Any `Translate` call BEFORE `Scale` gets multiplied by the scale factor
- Any `Translate` call AFTER `Scale` stays unscaled (just added)

**Correct GeoM order for sprite rendering:**

```go
op.GeoM.Translate(-originX, -originY)    // origin → scaled
op.GeoM.Translate(ox, oy)                // per-frame offset → scaled
op.GeoM.Rotate(rot)                      // rotation about origin
op.GeoM.Scale(sx*Scale*flip, sy*Scale)   // scale everything above
op.GeoM.Translate(p.X, p.Y)              // world → NOT scaled
op.GeoM.Translate(-cameraX, 0)           // camera → NOT scaled
```

Do NOT reorder to the mathematically intuitive chain (e.g., `S → R → T` from left to right) — Ebitengine v2's `Translate` behavior breaks it.

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
