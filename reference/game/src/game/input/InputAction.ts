export const INPUT_ACTIONS = {
  MOVE_LEFT: "MOVE_LEFT",
  MOVE_RIGHT: "MOVE_RIGHT",
  JUMP: "JUMP",
  DODGE: "DODGE",
  ATTACK: "ATTACK",
  SKILL: "SKILL",
  TRANSFORM: "TRANSFORM",
  MENU_UP: "MENU_UP",
  MENU_DOWN: "MENU_DOWN",
  MENU_CONFIRM: "MENU_CONFIRM",
  MENU_BACK: "MENU_BACK",
} as const

export type InputAction = (typeof INPUT_ACTIONS)[keyof typeof INPUT_ACTIONS]
