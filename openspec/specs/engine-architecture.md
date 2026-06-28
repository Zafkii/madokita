# Engine Architecture — Madokita

## Overview

The game engine follows a **phase-driven lifecycle** with layered bootstrap systems, an ECS-influenced entity model, scene-based state management, and a simple string-keyed DI container.

## Phases

The `Runtime` in `internal/engine/runtime.go` owns a six-phase state machine:

| Phase | Description |
|-------|-------------|
| `PhasePreInit` | Initial state after construction |
| `PhaseInit` | Systems are registered |
| `PhasePostInit` | Systems.InitializeAll() completed |
| `PhaseReady` | Systems.StartAll() completed |
| `PhaseRunning` | Systems.ReadyAll() completed — game loop active |
| `PhaseShutdown` | Shutdown requested |

The three-phase startup pipeline runs sequentially:
1. `SystemManager.InitializeAll()` — allocate resources, load assets
2. `SystemManager.StartAll()` — begin background work
3. `SystemManager.ReadyAll()` — signal subsystems ready

**Invariant**: Each phase MUST complete before the next begins. `ShutdownAll()` iterates in reverse order.

## System Interface

`internal/engine/system.go`

Every subsystem implements `System`:

```go
type System interface {
    ID() string
    Initialize() error
    Start() error
    Ready() error
    Shutdown() error
}
```

`SystemManager` registers systems via `Add()`, runs them through each lifecycle phase in registration order (reverse for shutdown).

## DI Container

`internal/engine/container.go`

A flat `map[string]any` service locator:

- `Register(name, svc)` — stores service
- `Get(name)` — retrieves service (caller type-asserts)

**Registered services** (from `cmd/game/bootstrap.go`):

| Key | Type |
|-----|------|
| `"events"` | `*event.Bus` |
| `"input"` | `*input.Manager` |
| `"platform"` | `platform.API` |

## Event Bus

`internal/event/bus.go`

Typed pub/sub with auto-unsubscribe:

- `On(event string, handler Handler) func()` — subscribe, returns unsubscribe closure
- `Emit(event string, payload any)` — dispatch to all handlers
- `Off(event string, id uint64)` — explicit unsubscribe
- `Clear()` — remove all listeners

Handler IDs use `atomic.AddUint64` for thread-safe generation.

## Scene Manager

`internal/scene/manager.go`

Manages game states as named `Interface` implementations:

```go
type Interface interface {
    Update(dt float64) error
    Draw(screen *ebiten.Image)
    Enter() error
    Exit() error
    Pause()
    Resume()
    IsActive() bool
}
```

**Key behaviors**:

- Fixed 1280x720 render target scaled to window via `GeoM.Scale`
- `SwitchTo(name)` — instant switch
- `SwitchWithFade(name, duration)` — cross-fade: fade-out → scene.Exit() → swap → scene.Enter() → fade-in
- Registered scenes: `"preload"`, `"menu-intro"`, `"main-menu"`, `"settings"`

## Platform Abstraction

`internal/platform/api.go`

```go
type API interface {
    IsDesktop() bool
    PlatformName() string
    CloseGame()
    SetResolution(w, h int) error
    ToggleFullscreen() error
}
```

Current implementation: `Desktop` — stub that returns nil/empty (desktop-only).

## Game Loop

`cmd/game/main.go`

1. `NewGameApp()` constructs all singletons, registers systems
2. `runtime.Initialize()` runs the 3-phase pipeline
3. A background goroutine ticks `sceneMgr.Update(1.0/60.0)` at 60Hz
4. `ebiten.RunGame(app)` starts Ebitengine's own loop
5. `defer runtime.Shutdown()` tears down on exit

## Bootstrap Systems

`internal/engine/bootstrap.go`

Two `System` implementations wire up the domain:

- `platformBootstrap` — owns `platform.API`
- `sceneBootstrap` — owns scene manager, image cache, audio manager, input manager; registers scenes during `Initialize()`, switches to first scene during `Start()`

## Stage & Spawn System

`internal/assets/registry.go` — `StageDef` defines each stage's visual layers and metadata:

```go
type StageDef struct {
    ID         string
    Name       string
    Images     []AssetEntry
    Characters []string
    Enemies    []string
    BGM        string
    GroundY    float64   // ground line Y coordinate
    SpawnX     float64   // player horizontal spawn point
}
```

`internal/entity/player/player.go` — `Player.Spawn()` computes the correct grounded position:

```go
func (p *Player) Spawn(x, stageGroundY float64) {
    p.X = x
    p.StageGroundY = stageGroundY
    originY := p.AnimDef.DefaultOriginY // falls back to 0.5
    footOffset := FrameSize * (1 - originY) * p.Scale
    p.Y = stageGroundY - footOffset
    p.Movement.IsGrounded = true
}
```

`internal/game/scene.go` — `GameScene.SetPlayer(p)` automatically calls `p.Spawn(stage.SpawnX, stage.GroundY)`. Bootstrap code must NOT manually compute Y — the formula `GroundY - (FrameSize/2)*Scale` is wrong for non-0.5 DefaultOriginY characters.
