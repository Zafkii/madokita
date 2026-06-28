import type { SaveData } from "./SaveData"
import type { SaveRepository } from "./SaveRepository"
import { migrateSave } from "./migrations/migrateSave"
import { mergeSave } from "./utils/mergeSave"
import { validateSave } from "./validators/validateSave"

export class SaveManager {
  private static repository: SaveRepository
  private static data: SaveData
  static async initialize(repository: SaveRepository): Promise<void> {
    this.repository = repository
    const raw = await repository.load()
    let parsed: unknown = null
    if (raw) {
      try {
        parsed = JSON.parse(raw)
      } catch {
        parsed = null
      }
    }
    const validated = validateSave(parsed)
    const migrated = migrateSave(validated)
    this.data = migrated
  }

  static getData(): Readonly<SaveData> {
    return this.data
  }

  static async save(): Promise<void> {
    await this.repository.save(this.data)
  }

  static async update(data: Partial<SaveData>): Promise<void> {
    this.data = mergeSave(this.data, data)
    await this.save()
  }
}
