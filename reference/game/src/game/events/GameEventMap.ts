export interface GameEventMap {
  "enemy-defeated": {
    enemyId: string
  }

  "achievement-unlocked": {
    achievementId: string
  }

  "player-damaged": {
    amount: number
  }

  "attack-started": {
    attackId: string
  }

  "hit-landed": {
    attackerId: string
    defenderId: string
    damage: number
    poiseDamage: number
    brokePoise: boolean
    killed: boolean
  }

  "stagger-start": {
    actorId: string
    duration: number
    isFlinch: boolean
  }

  "stagger-end": {
    actorId: string
  }

  "flinch-end": {
    actorId: string
  }

  "actor-died": {
    actorId: string
    team: string
  }
}
