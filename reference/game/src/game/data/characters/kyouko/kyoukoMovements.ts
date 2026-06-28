import type { Movement } from "../../../entities/base/combat/MovementTypes"

export const kyoukoMovement: Movement = {
  assetKey: "kyouko",
  defaultOriginX: 0.5,
  defaultOriginY: 0.5,
  animations: {
    idle: {
      fps: 3,
      loop: true,
      frames: [
        {
          spriteFrames: [0],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [95, 61, 1.5, -32.5, 1, 1, 0, 1],
            [54, 130, 4, 62, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [1],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [95, 61, 1.5, -32.5, 1, 1, 0, 1],
            [54, 130, 4, 62, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [2],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [95, 61, 1.5, -32.5, 1, 1, 0, 1],
            [54, 130, 4, 62, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [3],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [95, 61, 1.5, -32.5, 1, 1, 0, 1],
            [54, 130, 4, 62, 1, 1, 0, 1],
          ],
        },
      ],
    },
    walk: {
      fps: 10,
      loop: true,
      frames: [
        {
          spriteFrames: [4],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [100, 57, 1, -32.5, 1, 1, 0, 1],
            [52, 130, 1, 61, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [5],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [100, 57, 1, -32.5, 1, 1, 0, 1],
            [52, 130, 1, 61, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [6],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [100, 57, 1, -32.5, 1, 1, 0, 1],
            [52, 130, 1, 61, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [7],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [100, 57, 1, -32.5, 1, 1, 0, 1],
            [52, 130, 1, 61, 1, 1, 0, 1],
          ],
        },
      ],
    },
    jump: {
      fps: 7,
      loop: false,
      frames: [
        {
          spriteFrames: [17],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [92, 63, 0, -34.5, 1, 1, 0, 1],
            [52, 103, -2, 49.5, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [18],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [92, 63, 0, -34.5, 1, 1, 0, 1],
            [52, 128, -2, 62, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [19],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [92, 63, 0, -34.5, 1, 1, 0, 1],
            [52, 128, -2, 62, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [20],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [92, 63, 0, -34.5, 1, 1, 0, 1],
            [52, 128, -2, 62, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [21],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [92, 63, 0, -34.5, 1, 1, 0, 1],
            [52, 128, -2, 62, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [22],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [92, 63, 0, -34.5, 1, 1, 0, 1],
            [52, 128, -2, 62, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [23],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [92, 63, 0, -34.5, 1, 1, 0, 1],
            [52, 107, -2, 51.5, 1, 1, 0, 1],
          ],
        },
      ],
    },
    dodge: {
      fps: 1,
      loop: false,
      frames: [
        {
          spriteFrames: [24],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [89, 74, 5, -11, 1, 1, 0, 1],
            [40, 102, -35.1, 69.65, 1, 1, 35, 1],
          ],
        },
      ],
    },
    skill: {
      fps: 8,
      loop: false,
      frames: [
        {
          spriteFrames: [8],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [96, 56, 1, -34, 1, 1, 0, 1],
            [52, 133, 8, 60.5, 1, 1, 0, 1],
          ],
        },
      ],
    },
    burst: {
      fps: 7,
      loop: false,
      frames: [
        {
          spriteFrames: [9],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [93, 61, -3.5, -34.5, 1, 1, 0, 1],
            [35, 131, -1.5, 59.5, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [10],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [93, 61, -3.5, -34.5, 1, 1, 0, 1],
            [35, 131, -1.5, 59.5, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [11],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [93, 61, -3.5, -34.5, 1, 1, 0, 1],
            [35, 131, -1.5, 59.5, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [12],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [93, 61, -3.5, -34.5, 1, 1, 0, 1],
            [35, 131, -1.5, 59.5, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [13],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [93, 61, -3.5, -34.5, 1, 1, 0, 1],
            [35, 131, -1.5, 59.5, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [14],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [93, 61, -3.5, -34.5, 1, 1, 0, 1],
            [35, 131, -1.5, 59.5, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [15],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [93, 61, -3.5, -34.5, 1, 1, 0, 1],
            [35, 131, -1.5, 59.5, 1, 1, 0, 1],
          ],
        },
        {
          spriteFrames: [16],
          offsetX: [0],
          offsetY: [0],
          rotation: [0],
          hurtboxes: [
            [93, 61, -3.5, -34.5, 1, 1, 0, 1],
            [35, 131, -1.5, 59.5, 1, 1, 0, 1],
          ],
        },
      ],
    },
  },
}
