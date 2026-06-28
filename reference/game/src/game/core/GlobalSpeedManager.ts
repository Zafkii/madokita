export class GlobalSpeedManager {
  private static speed = 1

  static getSpeed(): number {
    return this.speed
  }

  static setSpeed(value: number): void {
    this.speed = value
  }

  static increase(amount: number): void {
    this.speed += amount
  }

  static reset(): void {
    this.speed = 1
  }
}
