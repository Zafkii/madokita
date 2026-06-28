import type { SaveData } from "../SaveData"
import type { SaveMigration } from "./SaveMigration"
import { v1ToV2 } from "./v1ToV2"

const migrations: SaveMigration[] = [v1ToV2]

export function migrateSave(data: SaveData): SaveData {
  let current = structuredClone(data)
  for (const migration of migrations) {
    if (current.version < migration.version) {
      current = migration.migrate(current)
    }
  }
  return current
}
