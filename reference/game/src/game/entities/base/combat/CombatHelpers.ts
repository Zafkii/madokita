import type { HurtboxData } from "./MovementTypes"
import type { GameEventBus } from "../../../events/GameEventBus"

// ── Hurtbox resolution ──────────────────────────────────────────────────────

export interface HasFrameHurtboxes {
  getCurrentFrameHurtboxes(): HurtboxData[] | null
  getDefaultHurtboxes(): HurtboxData[] | undefined
}

export function resolveHurtboxData(
  primary: HasFrameHurtboxes | null | undefined,
  fallback: HasFrameHurtboxes | null | undefined,
): HurtboxData[] | null {
  const fromPrimary = primary?.getCurrentFrameHurtboxes()
  if (fromPrimary) return fromPrimary

  const fromFallback = fallback?.getCurrentFrameHurtboxes()
  if (fromFallback) return fromFallback

  return fallback?.getDefaultHurtboxes() ?? null
}

// ── EventBus subscriptions ──────────────────────────────────────────────────

export interface CombatEventHandlers {
  onStaggerStart: (isFlinch: boolean) => void
  onStaggerEnd: () => void
  onFlinchEnd: () => void
  onDeath: () => void
}

export function subscribeToCombatEvents(
  eventBus: GameEventBus,
  actorId: string,
  handlers: CombatEventHandlers,
): (() => void)[] {
  return [
    eventBus.on("stagger-start", (payload) => {
      if (payload.actorId === actorId) {
        handlers.onStaggerStart(payload.isFlinch)
      }
    }),

    eventBus.on("stagger-end", (payload) => {
      if (payload.actorId === actorId) {
        handlers.onStaggerEnd()
      }
    }),

    eventBus.on("flinch-end", (payload) => {
      if (payload.actorId === actorId) {
        handlers.onFlinchEnd()
      }
    }),

    eventBus.on("actor-died", (payload) => {
      if (payload.actorId === actorId) {
        handlers.onDeath()
      }
    }),
  ]
}
