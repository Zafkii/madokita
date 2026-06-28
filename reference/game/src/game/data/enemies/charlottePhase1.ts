import type { Movement } from "../../entities/base/combat/MovementTypes"

export const charlottePhase1Movement: Movement = {
  assetKey: "charlotte_phase_1",
  defaultOriginX: 0.5,
  defaultOriginY: 0.5,
  animations: {
    idle: {
      fps: 4,
      loop: true,
      frames: [
        { spriteFrames: [0], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [1], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [2], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [3], offsetX: [0], offsetY: [0], rotation: [0] },
      ],
    },
  },
}
