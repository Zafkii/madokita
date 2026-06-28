import type { Movement } from "../../types/MovementTypes"
import type { MovementEntry } from "./movementTypes"
import {
  normalizeExportPath,
} from "../shared/importExportUtils"

export { pathToUrl, downloadFile } from "../shared/importExportUtils"

export function parseMovementImportText(text: string): Movement | null {
  try {
    let searchStart = 0
    while (true) {
      const startIdx = text.indexOf("{", searchStart)
      if (startIdx === -1) {
        console.warn("[parseMovementImportText] no '{' found from index", searchStart)
        return null
      }

      let depth = 0
      let endIdx = -1
      for (let i = startIdx; i < text.length; i++) {
        if (text[i] === "{") depth++
        else if (text[i] === "}") {
          depth--
          if (depth === 0) {
            endIdx = i + 1
            break
          }
        }
      }
      if (endIdx === -1) {
        console.warn("[parseMovementImportText] unmatched '{' at index", startIdx)
        return null
      }

      const cleaned = text.slice(startIdx, endIdx).trim()
        try {
        const obj = new Function("return (" + cleaned + ")")()
        if (obj && typeof obj === "object" && obj.assetKey && obj.animations) {
          console.log("[parseMovementImportText] found Movement object, assetKey:", obj.assetKey, "animations:", Object.keys(obj.animations))
          return obj as Movement
        }
        console.log("[parseMovementImportText] block at", startIdx, "has no assetKey/animations — skipping")
      } catch (e) {
        console.debug("[parseMovementImportText] eval failed for block near index", searchStart, e)
      }

      searchStart = endIdx
    }
  } catch (e) {
    console.error("[parseMovementImportText] unexpected error", e)
    return null
  }
}

export { parseLoadImagePaths as parseMovementLoadImagePaths } from "../shared/importExportUtils"

export function exportMovements(
  entries: MovementEntry[],
  spritePaths: Record<string, string>,
): string {
  const lines: string[] = []
  lines.push(`// Movement Editor`)
  lines.push(
    `import type { Movement } from "../../entities/base/combat/MovementTypes"`,
  )
  lines.push(``)

  for (let w = 0; w < entries.length; w++) {
    const entry = entries[w]
    const def = entry.def
    const varName = `movementDef${w > 0 ? w : ""}`

    lines.push(`const ${varName}: Movement = {`)
    lines.push(`  assetKey: "${def.assetKey}",`)
    if (def.defaultOriginX !== undefined) {
      lines.push(`  defaultOriginX: ${def.defaultOriginX},`)
    }
    if (def.defaultOriginY !== undefined) {
      lines.push(`  defaultOriginY: ${def.defaultOriginY},`)
    }
    if (def.defaultHurtboxes) {
      lines.push(`  defaultHurtboxes: ${JSON.stringify(def.defaultHurtboxes)},`)
    }
    lines.push(`  animations: {`)

    const animNames = Object.keys(def.animations)
    for (let a = 0; a < animNames.length; a++) {
      const animName = animNames[a]
      const anim = def.animations[animName]
      lines.push(`    ${animName}: {`)
      if (anim.fps !== undefined) {
        lines.push(`      fps: ${anim.fps},`)
      }
      if (anim.loop !== undefined) {
        lines.push(`      loop: ${anim.loop},`)
      }
      lines.push(`      frames: [`)

      for (const frame of anim.frames) {
        const n = frame.spriteFrames.length
        const arr = (vals: number[] | undefined) => {
          if (!vals) return `[${Array(n).fill("1").join(", ")}]`
          while (vals.length < n) vals.push(0)
          return `[${vals.join(", ")}]`
        }
        lines.push(`        {`)
        lines.push(`          spriteFrames: ${arr(frame.spriteFrames)},`)
        lines.push(`          offsetX: ${arr(frame.offsetX)},`)
        lines.push(`          offsetY: ${arr(frame.offsetY)},`)
        lines.push(`          rotation: ${arr(frame.rotation)},`)
        if (frame.scaleX && frame.scaleX.some(v => v !== 1)) lines.push(`          scaleX: ${arr(frame.scaleX)},`)
        if (frame.scaleY && frame.scaleY.some(v => v !== 1)) lines.push(`          scaleY: ${arr(frame.scaleY)},`)
        if (frame.hurtboxes) lines.push(`          hurtboxes: ${JSON.stringify(frame.hurtboxes)},`)
        lines.push(`        },`)
      }

      lines.push(`      ],`)
      lines.push(`    },`)
    }

    lines.push(`  },`)
    lines.push(`}`)

    const charPath = spritePaths["char"]
    if (charPath) {
      lines.push(
        `// this.load.image("char", "${normalizeExportPath(charPath)}")`,
      )
    }
    for (const [key, path] of Object.entries(spritePaths)) {
      if (key === "char") continue
      lines.push(
        `// this.load.image("${key}", "${normalizeExportPath(path)}")`,
      )
    }

    lines.push(``)
    lines.push(
      `export const ${entry.name.replace(/[^a-zA-Z0-9_]/g, "_")} = ${varName}`,
    )
    lines.push(``)
  }

  return lines.join("\n")
}
