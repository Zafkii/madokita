export type LoadedImage = {
  name: string
  image: HTMLImageElement
  width: number
  height: number
}

export function createFilePicker(
  label: string,
  accept: string,
  onLoad: (result: LoadedImage) => void,
): HTMLElement {
  const container = document.createElement("div")
  container.style.display = "flex"
  container.style.alignItems = "center"
  container.style.gap = "8px"

  const labelEl = document.createElement("label")
  labelEl.textContent = label
  labelEl.style.fontSize = "13px"
  labelEl.style.color = "var(--label-color,#aaa)"
  labelEl.style.whiteSpace = "nowrap"

  const input = document.createElement("input")
  input.type = "file"
  input.accept = accept
  input.style.display = "none"

  const btn = document.createElement("button")
  btn.textContent = "Browse..."
  btn.style.cssText =
    "padding:4px 12px;background:#2a2a4a;color:var(--btn-color,#e0e0e0);border:1px solid #444;border-radius:4px;cursor:pointer;font-size:13px;"

  const nameEl = document.createElement("span")
  nameEl.textContent = "File"
  nameEl.style.fontSize = "12px"
  nameEl.style.color = "var(--label-color,#888)"
  nameEl.style.overflow = "hidden"
  nameEl.style.textOverflow = "ellipsis"
  nameEl.style.maxWidth = "180px"
  nameEl.style.whiteSpace = "nowrap"

  btn.addEventListener("click", () => input.click())

  input.addEventListener("change", () => {
    const file = input.files?.[0]
    if (!file) return

    const img = new Image()
    img.onload = () => {
      onLoad({
        name: file.name,
        image: img,
        width: img.width,
        height: img.height,
      })
    }
    img.onerror = () => {
      nameEl.textContent = "Error loading file"
    }
    img.src = URL.createObjectURL(file)
    input.value = ""
  })

  container.appendChild(labelEl)
  container.appendChild(btn)
  container.appendChild(input)
  container.appendChild(nameEl)

  return container
}

export function loadImageFromUrl(url: string): Promise<LoadedImage> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => {
      const name = url.split("/").pop() ?? url
      resolve({ name, image: img, width: img.width, height: img.height })
    }
    img.onerror = reject
    img.src = url
  })
}
