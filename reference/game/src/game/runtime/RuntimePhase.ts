export const RuntimePhase = {
  PRE_INIT: "PRE_INIT",
  INIT: "INIT",
  POST_INIT: "POST_INIT",
  READY: "READY",
  RUNNING: "RUNNING",
  SHUTDOWN: "SHUTDOWN",
} as const

export type RuntimePhase = (typeof RuntimePhase)[keyof typeof RuntimePhase]
