package assets

import (
	"image/color"

	"madokita/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

type LayerItem struct {
	Entry AssetEntry
	Image *ebiten.Image
}

type AssetManager struct {
	imgCache    *ui.ImageCache
	frames      map[string][]*ebiten.Image
	charFrames  map[string][][]*ebiten.Image
	stageLayers map[string][]LayerItem
	tint        *TintController
}

func NewAssetManager(cache *ui.ImageCache) *AssetManager {
	return &AssetManager{
		imgCache:    cache,
		frames:      make(map[string][]*ebiten.Image),
		charFrames:  make(map[string][][]*ebiten.Image),
		stageLayers: make(map[string][]LayerItem),
		tint:        NewTintController(),
	}
}

func (m *AssetManager) PreloadStage(stage *StageDef) error {
	var items []LayerItem
	for _, entry := range stage.Images {
		img, err := m.imgCache.Load(entry.Path)
		if err != nil {
			img = placeholderImage(entry)
		}
		if entry.FrameCount > 1 {
			m.frames[entry.Key] = SliceFrames(img, entry.FrameW, entry.FrameH, entry.FrameCount)
		}
		items = append(items, LayerItem{Entry: entry, Image: img})
	}
	m.stageLayers[stage.ID] = items
	return nil
}

func (m *AssetManager) LoadCharacter(char *CharacterSheets) error {
	var sheets [][]*ebiten.Image
	for _, entry := range char.Sheets {
		img, err := m.imgCache.Load(entry.Path)
		if err != nil {
			img = placeholderImage(entry)
		}
		frames := SliceFrames(img, entry.FrameW, entry.FrameH, entry.FrameCount)
		sheets = append(sheets, frames)
	}
	m.charFrames[char.Key] = sheets
	return nil
}

func (m *AssetManager) GetFrame(charKey string, sheetIdx, frameIdx int) *ebiten.Image {
	sheets, ok := m.charFrames[charKey]
	if !ok {
		return nil
	}
	if sheetIdx < 0 || sheetIdx >= len(sheets) {
		return nil
	}
	if frameIdx < 0 || frameIdx >= len(sheets[sheetIdx]) {
		return nil
	}
	return sheets[sheetIdx][frameIdx]
}

func (m *AssetManager) GetFrames(key string) []*ebiten.Image {
	return m.frames[key]
}

func (m *AssetManager) StageLayers(stageID string) []LayerItem {
	return m.stageLayers[stageID]
}

func (m *AssetManager) Tint() *TintController {
	return m.tint
}

func (m *AssetManager) UnloadStage(stageID string) {
	delete(m.stageLayers, stageID)
}

func placeholderImage(entry AssetEntry) *ebiten.Image {
	w := entry.FrameW
	if w <= 0 {
		w = 1280
	}
	h := entry.FrameH
	if h <= 0 {
		h = 720
	}
	img := ebiten.NewImage(w, h)
	img.Fill(layerColor(entry.Group))
	return img
}

func layerColor(g LayerGroup) color.Color {
	switch g {
	case LayerSky:
		return color.RGBA{100, 150, 255, 255}
	case LayerClouds:
		return color.RGBA{200, 220, 255, 255}
	case LayerMountainsFar:
		return color.RGBA{120, 140, 180, 255}
	case LayerMountainsNear:
		return color.RGBA{80, 120, 80, 255}
	case LayerStructFar:
		return color.RGBA{140, 100, 80, 255}
	case LayerStructNear:
		return color.RGBA{100, 70, 50, 255}
	case LayerFloor:
		return color.RGBA{60, 140, 60, 255}
	case LayerCharBack:
		return color.RGBA{255, 100, 100, 255}
	case LayerCharMain:
		return color.RGBA{255, 50, 50, 255}
	case LayerCharFront:
		return color.RGBA{200, 0, 0, 255}
	case LayerOverlay:
		return color.RGBA{100, 200, 100, 255}
	case LayerEffect:
		return color.RGBA{255, 255, 0, 255}
	default:
		return color.RGBA{255, 0, 255, 255}
	}
}
