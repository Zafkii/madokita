import type { SaveData } from "../SaveData"
import type { SaveRepository } from "../SaveRepository"

export class ElectronSaveRepository implements SaveRepository {
  async load(): Promise<string | null> {
    return (await window.electronAPI.loadSave()) ?? null
  }

  async save(data: SaveData): Promise<void> {
    await window.electronAPI.saveGame(JSON.stringify(data))
  }
}
