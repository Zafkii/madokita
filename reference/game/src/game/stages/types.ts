export interface StageData {
  id: number

  name: string

  background: string

  playableCharacter: string

  allies: string[]

  enemies: string[]

  baseSpeed: number

  nextStage?: number
}
