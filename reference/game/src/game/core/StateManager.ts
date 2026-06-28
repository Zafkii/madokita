import { stage1 } from "../stages/stage1"
import type { StageData } from "../stages/types"

export class StageManager {
  private static currentStage: StageData = stage1

  static getCurrentStage(): StageData {
    return this.currentStage
  }

  static setStage(stage: StageData): void {
    this.currentStage = stage
  }
}
