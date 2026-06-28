import type { HurtboxData, Movement } from "../../../entities/base/combat/MovementTypes"

const butterflyHurtboxes: HurtboxData[] = [
  [50, 80, 0, 10, 1, 1, 0, 1],
]

export const mobButterflyMovement: Movement = {
  assetKey: "mobbutterfly",
  defaultOriginX: 0.5,
  defaultOriginY: 0.5,
  defaultHurtboxes: butterflyHurtboxes,
  animations: {
    idle: {
      fps: 2,
      loop: true,
      frames: [
        { spriteFrames: [0], offsetX: [0], offsetY: [0], rotation: [0] },
      ],
    },
    walk: {
      fps: 2,
      loop: true,
      frames: [
        { spriteFrames: [1], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [2], offsetX: [0], offsetY: [0], rotation: [0] },
      ],
    },
    windup: {
      fps: 8,
      loop: false,
      frames: [
        { spriteFrames: [3], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [4], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [5], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [6], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [7], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [8], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [9], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [10], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [11], offsetX: [0], offsetY: [0], rotation: [0] },
      ],
    },
    attack: {
      fps: 3,
      loop: false,
      frames: [
        { spriteFrames: [12], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [13], offsetX: [0], offsetY: [0], rotation: [0] },
      ],
    },
    recover: {
      fps: 3,
      loop: false,
      frames: [
        { spriteFrames: [14], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [15], offsetX: [0], offsetY: [0], rotation: [0] },
        { spriteFrames: [16], offsetX: [0], offsetY: [0], rotation: [0] },
      ],
    },
  },
}
export const MobButterflyData = {
  visionRange: 660,
  attackRange: 60,
  retreatDistance: 150,
  moveSpeed: 120,

  hp: 30,
}
