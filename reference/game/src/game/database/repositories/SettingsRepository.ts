import queries from "../queries/queries.json"

import { getDatabase } from "../db"

export class SettingsRepository {
  static async initialize(): Promise<void> {
    try {
      const db = getDatabase()

      await db.exec(queries.settings.createTable)
    } catch (error) {
      console.error("[SettingsRepository] Failed to initialize:", error)
    }
  }

  static async setSetting(key: string, value: string): Promise<void> {
    try {
      const db = getDatabase()

      await db.run(queries.settings.set, key, value)
    } catch (error) {
      console.error("[SettingsRepository] Failed to set setting:", error)
    }
  }

  static async getSetting(key: string): Promise<string | null> {
    try {
      const db = getDatabase()

      const row = await db.get<{ value: string }>(queries.settings.get, key)

      return row?.value ?? null
    } catch (error) {
      console.error("[SettingsRepository] Failed to get setting:", error)

      return null
    }
  }
  static async getAllSettings(): Promise<Record<string, string>> {
    try {
      const db = getDatabase()

      const rows = await db.all<
        {
          key: string
          value: string
        }[]
      >("SELECT key, value FROM settings")

      const settings: Record<string, string> = {}

      for (const row of rows) {
        settings[row.key] = row.value
      }

      return settings
    } catch (error) {
      console.error("[SettingsRepository] Failed to get all settings:", error)

      return {}
    }
  }
}
