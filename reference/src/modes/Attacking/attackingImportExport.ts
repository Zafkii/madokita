import type { Attack } from "../../types/AttackTypes"
import type { AttackingEntry } from "./attackingTypes"
import { normalizeExportPath } from "../shared/importExportUtils"

export { pathToUrl, downloadFile, parseLoadImagePaths } from "../shared/importExportUtils"

export function parseImportText(text: string): Attack | null {
  try {
    const lines = text.split("\n")

    let objectStartLine = -1

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim()
      if (
        line.includes("export const") &&
        line.includes(": Attack") &&
        line.includes("= {")
      ) {
        objectStartLine = i
        break
      }
    }

    if (objectStartLine === -1) {
      console.warn("[parseImportText] no 'export const ...: Attack = {' line found")
      return null
    }

    let fullText = lines.slice(objectStartLine).join("\n")
    const startIdx = fullText.indexOf("{")
    if (startIdx === -1) {
      console.warn("[parseImportText] no '{' after object start line")
      return null
    }

    let depth = 0
    let endIdx = -1
    for (let i = startIdx; i < fullText.length; i++) {
      if (fullText[i] === "{") depth++
      else if (fullText[i] === "}") {
        depth--
        if (depth === 0) {
          endIdx = i + 1
          break
        }
      }
    }

    if (endIdx === -1) {
      console.warn("[parseImportText] unmatched '{' starting at", startIdx)
      return null
    }

    const cleaned = fullText.slice(startIdx, endIdx).trim()

    try {
      const obj = new Function("return (" + cleaned + ")")()
      if (!obj) {
        console.warn("[parseImportText] eval returned null/undefined")
        return null
      }
      if (typeof obj !== "object") {
        console.warn("[parseImportText] eval result not an object", typeof obj)
        return null
      }
      if (!obj.assetKey) {
        console.warn("[parseImportText] parsed object missing assetKey")
        return null
      }
      if (!obj.animations) {
        console.warn("[parseImportText] parsed object missing animations")
        return null
      }
      console.log("[parseImportText] found Attack object, assetKey:", obj.assetKey, "animations:", Object.keys(obj.animations))
      return obj as Attack
    } catch (e) {
      console.error("[parseImportText] eval failed", e)
      return null
    }
  } catch (e) {
    console.error("[parseImportText] unexpected error", e)
    return null
  }
}

export function exportAttacks(
  attacks: AttackingEntry[],
  spritePaths: Record<string, string>,
): string {
  const lines: string[] = []
  lines.push(`// Attacking Editor`)
  lines.push(
    `import type { Attack } from "../../entities/base/combat/AttackTypes"`,
  )
  lines.push(``)

  for (let w = 0; w < attacks.length; w++) {
    const entry = attacks[w]
    const assetKey = entry.def.assetKey
    const defName = `${assetKey}Def`

    for (const [key, path] of Object.entries(spritePaths)) {
      if (path) {
        lines.push(
          `// this.load.image("${key}", "${normalizeExportPath(path)}")`,
        )
      }
    }

    lines.push(`export const ${defName}: Attack = {`)
    lines.push(`  assetKey: "${assetKey}",`)
    lines.push(`  defaultOriginX: ${entry.def.defaultOriginX ?? 0.5},`)
    lines.push(`  defaultOriginY: ${entry.def.defaultOriginY ?? 0.5},`)
    if (entry.def.defaultHurtboxes) {
      lines.push(
        `  defaultHurtboxes: ${JSON.stringify(entry.def.defaultHurtboxes)},`,
      )
    }
    lines.push(`  animations: {`)

    const animKeys = Object.keys(entry.def.animations)
    for (let a = 0; a < animKeys.length; a++) {
      const animName = animKeys[a]
      const anim = entry.def.animations[animName]
      lines.push(`    ${animName}: {`)
      if (anim.guardFps !== undefined)
        lines.push(`      guardFps: ${anim.guardFps},`)

      if (anim.windup !== undefined) lines.push(`      windup: ${anim.windup},`)
      if (anim.activeTime !== undefined)
        lines.push(`      activeTime: ${anim.activeTime},`)
      if (anim.recover !== undefined)
        lines.push(`      recover: ${anim.recover},`)
      if (anim.guard !== undefined) lines.push(`      guard: ${anim.guard},`)
      lines.push(`      frames: [`)
      for (const f of anim.frames) {
        const sf = JSON.stringify(f.spriteFrames ?? [0])
        const ox = JSON.stringify(f.offsetX ?? [0])
        const oy = JSON.stringify(f.offsetY ?? [0])
        const rot = JSON.stringify(f.rotation ?? [0])
        const parts = [
          `        { spriteFrames: ${sf}, offsetX: ${ox}, offsetY: ${oy}, rotation: ${rot}`,
        ]
        if (f.scaleX && f.scaleX.some((v) => v !== 1)) {
          parts.push(`, scaleX: ${JSON.stringify(f.scaleX)}`)
        }
        if (f.scaleY && f.scaleY.some((v) => v !== 1)) {
          parts.push(`, scaleY: ${JSON.stringify(f.scaleY)}`)
        }
        if (f.hurtboxes) {
          parts.push(`, hurtboxes: ${JSON.stringify(f.hurtboxes)}`)
        }
        parts.push(`, phase: "${f.phase ?? "atk"}" },`)
        lines.push(parts.join(""))
      }
      lines.push(`      ],`)
      lines.push(`    },`)
    }

    lines.push(`  },`)
    lines.push(`}`)
    lines.push(``)
  }

  return lines.join("\n")
}
