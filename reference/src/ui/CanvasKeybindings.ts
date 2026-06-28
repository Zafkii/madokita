export type MouseModifier = "ctrl" | "shift" | "alt"

export interface KeyBinding {
  key: string
  ctrl?: boolean
  shift?: boolean
}

export const canvasKeybindings = {
  panModifier: "ctrl" as MouseModifier,
  panWithRightClick: true,
  handleDragModifier: "alt" as MouseModifier,

  scaleModifier: "shift" as MouseModifier,
  rotateModifier: "alt" as MouseModifier,
  zoomModifier: "ctrl" as MouseModifier,

  undo: { key: "z", ctrl: true } as KeyBinding,
  redo: { key: "y", ctrl: true } as KeyBinding,
  redoAlt: { key: "z", ctrl: true, shift: true } as KeyBinding,
}

export function matchesModifier(
  e: { ctrlKey: boolean; metaKey: boolean; shiftKey: boolean; altKey: boolean },
  mod: MouseModifier | null,
): boolean {
  if (mod === "ctrl") return e.ctrlKey || e.metaKey
  if (mod === "shift") return e.shiftKey
  if (mod === "alt") return e.altKey
  return false
}

export function matchesBinding(e: KeyboardEvent, binding: KeyBinding): boolean {
  if (e.key !== binding.key) return false
  if (binding.ctrl && !(e.ctrlKey || e.metaKey)) return false
  if (binding.shift !== undefined && e.shiftKey !== binding.shift) return false
  return true
}
