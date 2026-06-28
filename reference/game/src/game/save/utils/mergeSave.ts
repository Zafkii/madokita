type PlainObject = Record<string, unknown>

function isObject(value: unknown): value is PlainObject {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}

export function mergeSave<T>(target: T, source: Partial<T>): T {
  const output = structuredClone(target) as PlainObject

  for (const key in source) {
    const sourceValue = source[key]

    const targetValue = output[key]

    if (sourceValue === undefined) {
      continue
    }

    if (isObject(targetValue) && isObject(sourceValue)) {
      output[key] = mergeSave(
        targetValue,
        sourceValue,
      ) as unknown as PlainObject

      continue
    }

    output[key] = structuredClone(sourceValue)
  }

  return output as T
}
