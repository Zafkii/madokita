export type Theme = "dark" | "light"

interface ThemeColors {
  canvasBg: string
  toolbarBg: string
  toolbarBorderBottom: string
  toolbarColor: string
  statusbarBg: string
  statusbarBorderTop: string
  statusbarColor: string
  panelBg: string
  panelBorderLeft: string
  gridLine: string
  gridAxis: string
  boundaryStroke: string
  bgBorderStroke: string
  handleFill: string
  handleStroke: string
  handleDash: string
  crosshairStroke: string
  textPrimary: string
  textMuted: string
  textAccent: string
  textOnColor: string
  textInverse: string
  labelColor: string
  buttonText: string
  buttonSwitchBg: string
  rowSelected: string
  fontSm: string
  fontMd: string
  fontLg: string
  radiusSm: string
  radiusMd: string
  radiusLg: string
  inputPad: string
  inputPadSm: string
  btnPadSm: string
  btnPadMd: string
  gapSm: string
  gapMd: string
  gapLg: string
  cellPad: string
  navPad: string
  sectionPad: string
}

const DARK: ThemeColors = {
  canvasBg: "#000000",
  toolbarBg: "#142c55", //bg too top
  toolbarBorderBottom: "#333333",
  toolbarColor: "#ffbcfe",
  statusbarBg: "#171740",
  statusbarBorderTop: "#333",
  statusbarColor: "#888",
  panelBg: "#173135", //bg tool side
  panelBorderLeft: "#333",
  gridLine: "rgba(255, 255, 255, 0.2)",
  gridAxis: "rgba(92, 176, 212, 0.68)",
  boundaryStroke: "rgba(0,200,255,0.35)",
  bgBorderStroke: "rgba(255,255,100,0.3)",
  handleFill: "rgba(255, 255, 255, 0.85)",
  handleStroke: "rgba(0,200,255,0.9)",
  handleDash: "rgba(0,200,255,0.25)",
  crosshairStroke: "rgba(0,200,255,0.8)",
  textPrimary: "#e0e0e0", //prob unique value
  textMuted: "#d8ff2a", //prob unique value
  textAccent: "rgb(108, 187, 247)", //prob unique value
  textOnColor: "#cccccc", //prob unique value
  textInverse: "#eeeeee", //prob unique value
  labelColor: "#a0ffe1",
  buttonText: "#f6ffbd",
  buttonSwitchBg: "#009b9b",
  rowSelected: "rgba(251, 255, 0, 0.43)",
  fontSm: "12px",
  fontMd: "12px",
  fontLg: "13px",
  radiusSm: "3px",
  radiusMd: "3px",
  radiusLg: "4px",
  inputPad: "4px 6px",
  inputPadSm: "1px 2px",
  btnPadSm: "2px 8px",
  btnPadMd: "3px 14px",
  gapSm: "4px",
  gapMd: "8px",
  gapLg: "12px",
  cellPad: "1px 4px",
  navPad: "0 4px",
  sectionPad: "3px 4px",
}

const LIGHT: ThemeColors = {
  canvasBg: "#328da5",
  toolbarBg: "#00537a",
  toolbarBorderBottom: "#000000",
  toolbarColor: "#ff9af3",
  statusbarBg: "#001954",
  statusbarBorderTop: "#696969",
  statusbarColor: "#000000",
  panelBg: "#007670",
  panelBorderLeft: "#000000",
  gridLine: "rgba(36, 23, 68, 0.31)",
  gridAxis: "rgba(0, 235, 98, 0.67)",
  boundaryStroke: "rgba(106, 255, 0, 0.82)",
  bgBorderStroke: "rgba(255,255,100,0.3)",
  handleFill: "rgba(255, 255, 255, 0.47)",
  handleStroke: "rgba(0, 200, 255, 0.54)",
  handleDash: "rgba(199, 199, 199, 0.46)",
  crosshairStroke: "rgba(0,200,255,0.8)",
  textPrimary: "#ffffff",
  textMuted: "#ee00be",
  textAccent: "#7a7a7a",
  textOnColor: "#ffffff",
  textInverse: "#333333",
  labelColor: "#95faff",
  buttonText: "#a7ffe3",
  buttonSwitchBg: "#000000",
  rowSelected: "rgba(0, 255, 247, 0.34)",
  fontSm: "12px",
  fontMd: "12px",
  fontLg: "13px",
  radiusSm: "3px",
  radiusMd: "3px",
  radiusLg: "4px",
  inputPad: "4px 6px",
  inputPadSm: "1px 2px",
  btnPadSm: "2px 8px",
  btnPadMd: "3px 14px",
  gapSm: "4px",
  gapMd: "8px",
  gapLg: "12px",
  cellPad: "1px 4px",
  navPad: "0 4px",
  sectionPad: "3px 4px",
}

export class ThemeManager {
  private _current: Theme = "light"
  private _c: ThemeColors = LIGHT

  get current(): Theme {
    return this._current
  }

  get isLight(): boolean {
    return this._current === "light"
  }

  setTheme(theme: Theme): void {
    this._current = theme
    this._c = theme === "dark" ? DARK : LIGHT
  }

  get canvasBg(): string {
    return this._c.canvasBg
  }
  get toolbarBg(): string {
    return this._c.toolbarBg
  }
  get toolbarBorderBottom(): string {
    return this._c.toolbarBorderBottom
  }
  get toolbarColor(): string {
    return this._c.toolbarColor
  }
  get statusbarBg(): string {
    return this._c.statusbarBg
  }
  get statusbarBorderTop(): string {
    return this._c.statusbarBorderTop
  }
  get statusbarColor(): string {
    return this._c.statusbarColor
  }
  get panelBg(): string {
    return this._c.panelBg
  }
  get panelBorderLeft(): string {
    return this._c.panelBorderLeft
  }
  get gridLine(): string {
    return this._c.gridLine
  }
  get gridAxis(): string {
    return this._c.gridAxis
  }
  get boundaryStroke(): string {
    return this._c.boundaryStroke
  }
  get bgBorderStroke(): string {
    return this._c.bgBorderStroke
  }
  get handleFill(): string {
    return this._c.handleFill
  }
  get handleStroke(): string {
    return this._c.handleStroke
  }
  get handleDash(): string {
    return this._c.handleDash
  }
  get crosshairStroke(): string {
    return this._c.crosshairStroke
  }
  get textPrimary(): string {
    return this._c.textPrimary
  }
  get textMuted(): string {
    return this._c.textMuted
  }
  get textAccent(): string {
    return this._c.textAccent
  }
  get textOnColor(): string {
    return this._c.textOnColor
  }
  get textInverse(): string {
    return this._c.textInverse
  }
  get labelColor(): string {
    return this._c.labelColor
  }
  get buttonText(): string {
    return this._c.buttonText
  }
  get buttonSwitchBg(): string {
    return this._c.buttonSwitchBg
  }
  get rowSelected(): string {
    return this._c.rowSelected
  }
  get fontSm(): string {
    return this._c.fontSm
  }
  get fontMd(): string {
    return this._c.fontMd
  }
  get fontLg(): string {
    return this._c.fontLg
  }
  get radiusSm(): string {
    return this._c.radiusSm
  }
  get radiusMd(): string {
    return this._c.radiusMd
  }
  get radiusLg(): string {
    return this._c.radiusLg
  }
  get inputPad(): string {
    return this._c.inputPad
  }
  get inputPadSm(): string {
    return this._c.inputPadSm
  }
  get btnPadSm(): string {
    return this._c.btnPadSm
  }
  get btnPadMd(): string {
    return this._c.btnPadMd
  }
  get gapSm(): string {
    return this._c.gapSm
  }
  get gapMd(): string {
    return this._c.gapMd
  }
  get gapLg(): string {
    return this._c.gapLg
  }
  get cellPad(): string {
    return this._c.cellPad
  }
  get navPad(): string {
    return this._c.navPad
  }
  get sectionPad(): string {
    return this._c.sectionPad
  }
}
