import type { SaveData } from "../SaveData"
import { defaultSave } from "../defaultSave"
import { mergeSave } from "../utils/mergeSave"

export function validateSave(data: unknown): SaveData {
  if (!data || typeof data !== "object") {
    return structuredClone(defaultSave)
  }

  return mergeSave(defaultSave, data as Partial<SaveData>)
}
