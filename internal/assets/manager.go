package assets

import (
	"image/color"

	"madokita/internal/animation"
	"madokita/internal/data"
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
	charPaths   map[string][]string
	pathRefs    map[string]int
	stageLayers map[string][]LayerItem
	tint        *TintController
}

func NewAssetManager(cache *ui.ImageCache) *AssetManager {
	return &AssetManager{
		imgCache:    cache,
		frames:      make(map[string][]*ebiten.Image),
		charFrames:  make(map[string][][]*ebiten.Image),
		charPaths:   make(map[string][]string),
		pathRefs:    make(map[string]int),
		stageLayers: make(map[string][]LayerItem),
		tint:        NewTintController(),
	}
}

func (m *AssetManager) PreloadStage(stage *StageDef) error {
	if _, ok := m.stageLayers[stage.ID]; ok {
		return nil
	}
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
		m.pathRefs[entry.Path]++
	}
	m.stageLayers[stage.ID] = items
	return nil
}

func (m *AssetManager) LoadCharacter(char *CharacterSheets) error {
	if _, ok := m.charFrames[char.Key]; ok {
		return nil
	}
	var sheets [][]*ebiten.Image
	paths := make([]string, 0, len(char.Sheets))
	for _, entry := range char.Sheets {
		img, err := m.imgCache.Load(entry.Path)
		if err != nil {
			img = placeholderImage(entry)
		}
		frames := SliceFrames(img, entry.FrameW, entry.FrameH, entry.FrameCount)
		sheets = append(sheets, frames)
		paths = append(paths, entry.Path)
		m.pathRefs[entry.Path]++
	}
	m.charFrames[char.Key] = sheets
	m.charPaths[char.Key] = paths
	return nil
}

// PreloadCharacter loads every spritesheet a character needs for a match:
// its primary movement animation plus its attack animation, if any.
func (m *AssetManager) PreloadCharacter(cd *data.CharacterData) error {
	if cd == nil || len(cd.Animations) == 0 {
		return nil
	}
	def := &cd.Animations[0]
	if err := m.LoadCharacter(characterSheets(def.AssetKey, def.Sprites)); err != nil {
		return err
	}
	if cd.Attack != nil {
		return m.LoadCharacter(characterSheets(cd.Attack.AssetKey, cd.Attack.Sprites))
	}
	return nil
}

// UnloadCharacter releases the frame slices and source textures of a
// character, returning memory between matches on low-end machines.
func (m *AssetManager) UnloadCharacter(charKey string) {
	delete(m.charFrames, charKey)
	for _, p := range m.charPaths[charKey] {
		m.releasePath(p)
	}
	delete(m.charPaths, charKey)
}

func characterSheets(assetKey string, sprites []animation.SpriteSheetDef) *CharacterSheets {
	sheets := make([]AssetEntry, 0, len(sprites))
	for _, s := range sprites {
		sheets = append(sheets, AssetEntry{
			Key:        assetKey,
			Path:       s.File,
			FrameW:     s.FrameW,
			FrameH:     s.FrameH,
			FrameCount: s.FrameCount,
		})
	}
	return &CharacterSheets{Key: assetKey, Sheets: sheets}
}

// releasePath drops one reference to a cached texture and frees it from the
// ImageCache once no character or stage uses it anymore.
func (m *AssetManager) releasePath(path string) {
	m.pathRefs[path]--
	if m.pathRefs[path] <= 0 {
		delete(m.pathRefs, path)
		m.imgCache.Remove(path)
	}
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
	for _, item := range m.stageLayers[stageID] {
		if item.Entry.FrameCount > 1 {
			delete(m.frames, item.Entry.Key)
		}
		m.releasePath(item.Entry.Path)
	}
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
