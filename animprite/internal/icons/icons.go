package icons

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"path"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// RasterSize is the source resolution used for embedded SVGs. Icons are
// rasterized once and then scaled down when drawn, which keeps their edges
// smooth without relying on Ebitengine to understand SVG.
const RasterSize = 48

// Set holds rasterized SVG icons keyed by their file name without extension.
type Set struct {
	masks  map[string]image.Image
	images map[string]*ebiten.Image
}

// Load rasterizes every SVG in fsys into an icon set. Invalid files are
// skipped and reported through the returned error so one bad asset does not
// hide the rest of the set.
func Load(fsys fs.FS, size int) (*Set, error) {
	if size <= 0 {
		size = RasterSize
	}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return &Set{masks: map[string]image.Image{}, images: map[string]*ebiten.Image{}}, err
	}

	set := &Set{
		masks:  make(map[string]image.Image, len(entries)),
		images: make(map[string]*ebiten.Image, len(entries)),
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(path.Ext(entry.Name()), ".svg") {
			continue
		}
		data, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", entry.Name(), err))
			continue
		}
		mask, err := Rasterize(data, size)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", entry.Name(), err))
			continue
		}
		key := strings.TrimSuffix(entry.Name(), path.Ext(entry.Name()))
		set.masks[key] = mask
	}
	return set, errors.Join(errs...)
}

// Rasterize converts an SVG document into a white alpha mask. The original
// colors are discarded so callers can tint the icon with the active theme.
func Rasterize(data []byte, size int) (image.Image, error) {
	if size <= 0 {
		size = RasterSize
	}

	// oksvg does not resolve the CSS currentColor keyword, which Lucide uses
	// for small filled dots. Since the icon is turned into an alpha mask
	// anyway, replace it with an opaque color before parsing.
	data = bytes.ReplaceAll(data, []byte("currentColor"), []byte("#000000"))
	icon, err := oksvg.ReadIconStream(bytes.NewReader(data), oksvg.WarnErrorMode)
	if err != nil {
		return nil, err
	}

	icon.SetTarget(0, 0, float64(size), float64(size))
	rgba := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, rgba, rgba.Bounds())
	raster := rasterx.NewDasher(size, size, scanner)
	icon.Draw(raster, 1)

	mask := image.NewRGBA(rgba.Bounds())
	visible := false
	for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
		for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
			_, _, _, alpha := rgba.At(x, y).RGBA()
			if alpha == 0 {
				continue
			}
			visible = true
			mask.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: uint8(alpha >> 8)})
		}
	}
	if !visible {
		return nil, errors.New("svg renders no visible pixels")
	}
	return mask, nil
}

// Has reports whether name is available in the set.
func (s *Set) Has(name string) bool {
	if s == nil {
		return false
	}
	return s.masks[name] != nil
}

// Image returns the Ebitengine image for name, creating it lazily because
// Ebitengine images can only be created after the game has started.
func (s *Set) Image(name string) *ebiten.Image {
	if s == nil {
		return nil
	}
	if img := s.images[name]; img != nil {
		return img
	}
	mask := s.masks[name]
	if mask == nil {
		return nil
	}
	img := ebiten.NewImageFromImage(mask)
	s.images[name] = img
	return img
}

// Draw paints a centered icon in the given cell, tinted with clr. It reports
// whether the icon existed.
func (s *Set) Draw(screen *ebiten.Image, name string, x, y, w, h int, clr color.Color) bool {
	img := s.Image(name)
	if img == nil || w <= 0 || h <= 0 {
		return false
	}

	bounds := img.Bounds()
	scale := min(float64(w)/float64(bounds.Dx()), float64(h)/float64(bounds.Dy()))
	dw := float64(bounds.Dx()) * scale
	dh := float64(bounds.Dy()) * scale

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x)+(float64(w)-dw)/2, float64(y)+(float64(h)-dh)/2)
	op.Filter = ebiten.FilterLinear
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(img, op)
	return true
}
