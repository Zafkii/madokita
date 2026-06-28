# Animation Editor — animprite

## Overview

Standalone Ebitengine-based animation editor for the Madokita game. Independent Go module with its own `go.mod` and `cmd/editor` entry point.

## Architecture

- `EditorApp` implements `ebiten.Game` (Update/Draw/Layout)
- **Immediate-mode UI**: every widget painted each frame via vector primitives
- No widget tree or layout system — `syncLayout()` recalculates positions on window resize
- Theme-aware via `theme.Manager` (Dark/Light toggle, hot-swappable)
- **`internal/project/`** — domain model (`ProjectData`, `AnimationRow`, `SpriteRow`, etc.) in a zero-dependency package, testable without Ebitengine. Includes `DeepCopy()` for undo/redo snapshots.
- `Update()` is split across 3 files: `update.go` (orchestrator, ~57 lines of dispatch), `update_window.go` (window chrome: resize, drag, cursor), `update_canvas.go` (canvas: wheel, mouse, handles, highlight)

## Widget Library

`animprite/internal/ui/`

| Widget      | File          | Description                                                                                                                         |
| ----------- | ------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `Button`    | `button.go`   | Color-coded types (Green/Red/Blue/Orange/Default), hover overlay, visible/enabled flags                                             |
| `Dropdown`  | `dropdown.go` | Label + popup list, multi-line display text, separate Draw/DrawPopup calls                                                          |
| `TextInput` | `input.go`    | Label + input field, numeric mode (clamp min/max/step), blinking cursor, Tab navigation                                             |
| `Slider`    | `slider.go`   | Track + draggable thumb, configurable min/max/step, real-time OnChange                                                              |
| `Table`     | `table.go`    | Title row (add/remove/extra buttons), column headers, scrollable data rows, selection highlight, scrollbar, custom DrawRow callback |

### Text Rendering

`label.go` — global functions (not struct):

- `SetFontFromTTF(ttfData, size)` — parse and set font
- `DrawText`, `DrawTextCentered`, `DrawTextRight` — cached offscreen buffers per string
- `TruncateText(str, maxWidth)` — add "..." overflow

## Camera

`animprite/internal/camera/camera.go`

- Pan: `Pan(dx, dy)` — offset X/Y
- Zoom: `ZoomAt(amount, cx, cy)` — zoom toward cursor point
- Transforms: `WorldToCanvas(wx, wy)`, `CanvasToWorld(cx, cy)`
- Range: [0.05, 20], default 1.0
- `Reset()` — back to (0,0), zoom 1.0

## Canvas

`animprite/internal/canvas/canvas.go`

Viewport rendering with:

- Zoom-aware grid (50px spacing, hidden when < 4px)
- Origin crosshair (16px scaled, min 4px)
- Dashed boundary rect (1280x720 default)
- Offscreen buffer (`drawBuf`) recreated on viewport change

## Theme

`animprite/internal/theme/theme.go`

- `Palette` struct with 29 color fields (canvas, panels, text, buttons, inputs, scrollbar)
- Two presets: `Dark` and `Light`
- `Manager` with `Toggle()` — immediate switch, no widget rebuild needed

## Current Editor Layout

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

## Implementation Status

### Completed (Visual Layer)

- Window chrome: drag, resize (6px edges), minimize/maximize/close with hover states
- Window position/size persistence (JSON in `UserConfigDir/madokita/animprite.json`)
- Camera: pan (left drag), zoom (scroll wheel toward cursor), reset
- Canvas: zoom-aware grid, origin crosshair, dashed boundary
- Theme: Dark/Light toggle with 29-color palette
- Mode toggle: Movement / Attack (changes right panel inputs)
- Data tables: Animation, Sprite, Hurtbox, Hitbox with Add/Remove/Scroll/Select and per-frame navigation buttons
- Right panel: 4 sections with TextInputs (15 fields), Phase dropdown, Preview slider
- Tab/Shift+Tab navigation between inputs
- Numeric input with clamping on Enter
- Status bar: zoom percentage, temporary messages, Reset View button

### Pending (Functionality)

- Sprite loading from disk (Browse button is placeholder)
- Hitbox/Hurtbox rendering and interaction on canvas
- Animation preview playback (Play/Loop/Speed UI exists but no runtime)
- Project Open/Save (buttons have no OnClick)
- Undo/Redo system
- Timeline or frame strip visualization
