import type { SaveData } from "../SaveData"
import type { SaveRepository } from "../SaveRepository"

const SAVE_KEY = "madokita-save"

export class BrowserSaveRepository implements SaveRepository {
  async load(): Promise<string | null> {
    return localStorage.getItem(SAVE_KEY)
  }

  async save(data: SaveData): Promise<void> {
    localStorage.setItem(SAVE_KEY, JSON.stringify(data))
  }
}
