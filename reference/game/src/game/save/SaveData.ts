export interface SaveData {
  version: number

  currentLocation: {
    stageId: string
    players: Record<string, { x: number; y: number }>
  }

  progress: {
    stagesUnlocked: string[]
    charactersUnlocked: string[]
    endingsUnlocked: string[]
  }

  stages: Record<string, StageProgress>

  player: {
    upgrades: Record<string, number>
    unlockedSkills: string[]
  }

  characters: Record<string, CharacterSaveData>

  lastSavedAt: string
}

export interface StageProgress {
  unlocked: boolean
  completed: boolean
}

export interface CharacterSaveData {
  unlockedSkills: string[]
}
