import queries from "../../database/queries/queries.json"

import { getDatabase } from "../../database/db"

export class GameSaveRepository {
  private static readonly DEFAULT_SLOT = "slot_1"

  static async initialize(): Promise<void> {
    try {
      const db = getDatabase()

      await db.exec(queries.saveSlots.createTable)
    } catch (error) {
      console.error("[GameSaveRepository] Failed to initialize:", error)
    }
  }

  static async save(data: string, version: number): Promise<void> {
    try {
      const db = getDatabase()

      await db.run(queries.saveSlots.set, this.DEFAULT_SLOT, version, data)
    } catch (error) {
      console.error("[GameSaveRepository] Failed to save:", error)
    }
  }

  static async load(): Promise<string | null> {
    try {
      const db = getDatabase()

      const row = await db.get<{ save_data: string }>(
        queries.saveSlots.get,
        this.DEFAULT_SLOT,
      )

      return row?.save_data ?? null
    } catch (error) {
      console.error("[GameSaveRepository] Failed to load:", error)

      return null
    }
  }
}