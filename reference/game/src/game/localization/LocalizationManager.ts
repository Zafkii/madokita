export class LocalizationManager {
  private static translations: Record<string, any> = {}

  static async load(language: string): Promise<void> {
    const response = await fetch(`/assets/locals/${language}.json`)
    if (!response.ok) {
      throw new Error(`Failed to load language: ${language}`)
    }

    this.translations = await response.json()
  }

  static get(path: string): string {
    const keys = path.split(".")
    let current: any = this.translations
    for (const key of keys) {
      current = current?.[key]
      if (current === undefined) {
        return path
      }
    }
    return typeof current === "string" ? current : path
  }
}
