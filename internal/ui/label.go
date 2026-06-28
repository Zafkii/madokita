package ui

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
)

var DefaultFace font.Face = basicfont.Face7x13
var SymbolFace font.Face = basicfont.Face7x13
var faceMu sync.Mutex

var baseTTF *opentype.Font
var baseSymbolTTF *opentype.Font
var fontCache = map[float64]font.Face{}
var symbolFontCache = map[float64]font.Face{}
var fontMu sync.Mutex

func SetFontFromTTF(data []byte, size float64) error {
	tt, err := opentype.Parse(data)
	if err != nil {
		return err
	}
	faceMu.Lock()
	defer faceMu.Unlock()
	f, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return err
	}
	DefaultFace = f
	baseTTF = tt
	for k, v := range textBufCache {
		v.Deallocate()
		delete(textBufCache, k)
	}
	fontMu.Lock()
	for k := range fontCache {
		delete(fontCache, k)
	}
	fontMu.Unlock()
	log.Printf("font: loaded TTF face (size=%.1f)", size)
	return nil
}

func SetSymbolFontFromTTF(data []byte, size float64) error {
	tt, err := opentype.Parse(data)
	if err != nil {
		return err
	}
	faceMu.Lock()
	defer faceMu.Unlock()
	f, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return err
	}
	SymbolFace = f
	baseSymbolTTF = tt
	for k, v := range textBufCache {
		v.Deallocate()
		delete(textBufCache, k)
	}
	fontMu.Lock()
	for k := range symbolFontCache {
		delete(symbolFontCache, k)
	}
	fontMu.Unlock()
	log.Printf("font: loaded symbol TTF face (size=%.1f)", size)
	return nil
}

func getFace(pixels float64) font.Face {
	size := math.Round(pixels)
	fontMu.Lock()
	defer fontMu.Unlock()
	if f, ok := fontCache[size]; ok {
		return f
	}
	if baseTTF == nil {
		return DefaultFace
	}
	f, err := opentype.NewFace(baseTTF, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return DefaultFace
	}
	fontCache[size] = f
	return f
}

func getSymbolFace(pixels float64) font.Face {
	size := math.Round(pixels)
	fontMu.Lock()
	defer fontMu.Unlock()
	if f, ok := symbolFontCache[size]; ok {
		return f
	}
	if baseSymbolTTF == nil {
		return SymbolFace
	}
	f, err := opentype.NewFace(baseSymbolTTF, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return SymbolFace
	}
	symbolFontCache[size] = f
	return f
}

var textBufCache = map[string]*ebiten.Image{}

func TextWidth(str string) int {
	b := text.BoundString(DefaultFace, str)
	return b.Dx()
}

func TextHeight() int {
	return DefaultFace.Metrics().Height.Ceil()
}

func drawTextFace(screen *ebiten.Image, str string, x, y int, face font.Face, size float64, clr color.Color) {
	bounds := text.BoundString(face, str)
	w := bounds.Dx()
	h := bounds.Dy()
	if w == 0 || h == 0 {
		return
	}

	key := fmt.Sprintf("%s-%.0f", str, size)
	offscreen, ok := textBufCache[key]
	if !ok || offscreen.Bounds().Dx() != w || offscreen.Bounds().Dy() != h {
		if offscreen != nil {
			offscreen.Deallocate()
		}
		offscreen = ebiten.NewImage(w, h)
		textBufCache[key] = offscreen
	}

	offscreen.Clear()
	text.Draw(offscreen, str, face, -bounds.Min.X, -bounds.Min.Y, clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	_, _, _, a := clr.RGBA()
	op.ColorScale.ScaleAlpha(float32(a) / 65535)
	screen.DrawImage(offscreen, op)
}

func DrawText(screen *ebiten.Image, str string, x, y int, scale float64, clr color.Color) {
	size := math.Round(13.0 * scale)
	face := getFace(size)
	drawTextFace(screen, str, x, y, face, size, clr)
}

func DrawTextCentered(screen *ebiten.Image, str string, cx, cy int, scale float64, clr color.Color) {
	size := math.Round(13.0 * scale)
	face := getFace(size)
	bounds := text.BoundString(face, str)
	x := cx - bounds.Dx()/2
	y := cy - bounds.Dy()/2
	drawTextFace(screen, str, x, y, face, size, clr)
}

func DrawTextRight(screen *ebiten.Image, str string, rx, y int, scale float64, clr color.Color) {
	size := math.Round(13.0 * scale)
	face := getFace(size)
	bounds := text.BoundString(face, str)
	x := rx - bounds.Dx()
	drawTextFace(screen, str, x, y, face, size, clr)
}

func SymbolTextWidth(str string) int {
	b := text.BoundString(SymbolFace, str)
	return b.Dx()
}

func drawSymbolTextFace(screen *ebiten.Image, str string, x, y int, face font.Face, size float64, clr color.Color) {
	bounds := text.BoundString(face, str)
	w := bounds.Dx()
	h := bounds.Dy()
	if w == 0 || h == 0 {
		return
	}

	key := "__sym__" + fmt.Sprintf("%s-%.0f", str, size)
	offscreen, ok := textBufCache[key]
	if !ok || offscreen.Bounds().Dx() != w || offscreen.Bounds().Dy() != h {
		if offscreen != nil {
			offscreen.Deallocate()
		}
		offscreen = ebiten.NewImage(w, h)
		textBufCache[key] = offscreen
	}

	offscreen.Clear()
	text.Draw(offscreen, str, face, -bounds.Min.X, -bounds.Min.Y, clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	_, _, _, a := clr.RGBA()
	op.ColorScale.ScaleAlpha(float32(a) / 65535)
	screen.DrawImage(offscreen, op)
}

func DrawSymbolText(screen *ebiten.Image, str string, x, y int, scale float64, clr color.Color) {
	size := math.Round(13.0 * scale)
	face := getSymbolFace(size)
	drawSymbolTextFace(screen, str, x, y, face, size, clr)
}

func DrawSymbolTextCentered(screen *ebiten.Image, str string, cx, cy int, scale float64, clr color.Color) {
	size := math.Round(13.0 * scale)
	face := getSymbolFace(size)
	bounds := text.BoundString(face, str)
	x := cx - bounds.Dx()/2
	y := cy - bounds.Dy()/2
	drawSymbolTextFace(screen, str, x, y, face, size, clr)
}

var TitleFace font.Face = basicfont.Face7x13
var titleTTF *opentype.Font
var titleFontMu sync.Mutex
var curTitleFontSize float64

func SetTitleFontFromTTF(data []byte) error {
	tt, err := opentype.Parse(data)
	if err != nil {
		return err
	}
	titleFontMu.Lock()
	defer titleFontMu.Unlock()
	titleTTF = tt
	curTitleFontSize = 0
	return nil
}

func SetTitleFontPixelSize(pixels float64) {
	size := math.Round(pixels)
	if size < 6 {
		size = 6
	}
	if size > 72 {
		size = 72
	}
	titleFontMu.Lock()
	defer titleFontMu.Unlock()
	if size == curTitleFontSize {
		return
	}
	curTitleFontSize = size
	if titleTTF == nil {
		TitleFace = basicfont.Face7x13
		return
	}
	f, err := opentype.NewFace(titleTTF, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		TitleFace = basicfont.Face7x13
		return
	}
	TitleFace = f
}

func TitleTextWidth(str string) int {
	return text.BoundString(TitleFace, str).Dx()
}

func TitleTextHeight() int {
	return TitleFace.Metrics().Height.Ceil()
}

func TitleAscent() int {
	return TitleFace.Metrics().Ascent.Ceil()
}

func DrawTitleText(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	text.Draw(screen, str, TitleFace, x, y, clr)
}

func FillRect(screen *ebiten.Image, x, y, w, h int, clr color.Color) {
	if w <= 0 || h <= 0 {
		return
	}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), clr, false)
}
