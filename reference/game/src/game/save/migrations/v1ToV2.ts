import type { SaveData } from "../SaveData"
import type { SaveMigration } from "./SaveMigration"

export const v1ToV2: SaveMigration = {
  version: 2,
  migrate(data: SaveData): SaveData {
    const v1 = data as SaveData & {
      currentLocation?: unknown
      stages?: unknown
      lastSavedAt?: unknown
    }

    const stages: Record<string, { unlocked: boolean; completed: boolean }> = {}
    for (const stageId of v1.progress.stagesUnlocked) {
      stages[stageId] = { unlocked: true, completed: false }
    }

    return {
      ...v1,
      version: 2,
      currentLocation: {
        stageId: "stage1",
        players: {},
      },
      stages,
      lastSavedAt: "",
    }
  },
}