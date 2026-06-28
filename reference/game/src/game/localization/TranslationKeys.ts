import { LocalizationManager } from "./LocalizationManager"

type TranslationTree = {
  [key: string]: TranslationTree | string
}

function createTranslationTree<T extends TranslationTree>(
  tree: T,
  parent = "",
): T {
  const result: any = {}
  for (const key in tree) {
    const value = tree[key]
    const path = parent ? `${parent}.${key}` : key
    if (typeof value === "object" && value !== null) {
      result[key] = createTranslationTree(value as TranslationTree, path)
      continue
    }

    Object.defineProperty(result, key, {
      get() {
        return LocalizationManager.get(path)
      },
      enumerable: true,
    })
  }

  return result
}

export const TranslationKeys = createTranslationTree({
  MENU: {
    TOUCH_TO_START: "",
    NEW_GAME: "",
    CONTINUE: "",
    SETTINGS: "",
    EXIT: "",
    DISPLAY: {
      RESOLUTION: "",
      FULLSCREEN: "",
    },
  },
})
