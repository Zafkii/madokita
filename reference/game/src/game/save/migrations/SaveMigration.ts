import type { SaveData } from "../SaveData"

export interface SaveMigration {
  version: number

  migrate(data: SaveData): SaveData
}
