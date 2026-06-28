import type { SaveData } from "./SaveData"

export interface SaveRepository {
  load(): Promise<string | null>
  save(data: SaveData): Promise<void>
}
