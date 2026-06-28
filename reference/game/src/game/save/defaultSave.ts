import type { SaveData } from "./SaveData"

export const defaultSave: SaveData = {
  version: 2,

  currentLocation: {
    stageId: "stage1",
    players: {
      sayaka: { x: 2000, y: 400 },
    },
  },

  progress: {
    stagesUnlocked: ["stage1"],

    charactersUnlocked: ["madoka"],

    endingsUnlocked: [],
  },

  stages: {
    stage1: { unlocked: true, completed: false },
  },

  player: {
    upgrades: {},

    unlockedSkills: [],
  },

  characters: {
    madoka: {
      unlockedSkills: [],
    },
  },

  lastSavedAt: "",
}
