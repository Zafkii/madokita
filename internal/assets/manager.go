package assets

import (
	"image/color"

	"madokita-gems/internal/animation"
	"madokita-gems/internal/data"
	"madokita-gems/internal/stage"
	"madokita-gems/internal/ui"

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
	stageImages map[string]map[string]*ebiten.Image
	stageAnims  map[string]map[string]*StageAnimation
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
		stageImages: make(map[string]map[string]*ebiten.Image),
		stageAnims:  make(map[string]map[string]*StageAnimation),
		tint:        NewTintController(),
	}
}

func (m *AssetManager) PreloadStage(def *StageDef) error {
	if _, ok := m.stageLayers[def.ID]; ok {
		return nil
	}
	images := make(map[string]*ebiten.Image)
	objectPlaceholder := ebiten.NewImage(50, 50)
	objectPlaceholder.Fill(color.RGBA{90, 90, 90, 255})
	load := func(path string, fallback *ebiten.Image) *ebiten.Image {
		if path == "" {
			return fallback
		}
		if img, ok := images[path]; ok {
			return img
		}
		img, err := m.imgCache.Load(path)
		if err != nil {
			img = fallback
		}
		images[path] = img
		m.pathRefs[path]++
		return img
	}

	var items []LayerItem
	for _, entry := range def.Images {
		img := load(entry.Path, placeholderImage(entry))
		if entry.FrameCount > 1 {
			m.frames[entry.Key] = SliceFrames(img, entry.FrameW, entry.FrameH, entry.FrameCount)
		}
		items = append(items, LayerItem{Entry: entry, Image: img})
	}
	if def.Layout != nil {
		for _, obj := range def.Layout.Objects {
			if obj.File == "" || obj.Category == stage.CategoryStart {
				continue
			}
			load(obj.File, objectPlaceholder)
		}
		usedAnimated := make(map[string]bool)
		for _, obj := range def.Layout.Objects {
			if obj.Category != "anim_prop" && obj.Category != "anim_block" && obj.AnimFile == "" {
				continue
			}
			if obj.File != "" {
				usedAnimated[obj.File] = true
			}
			if obj.AnimFile != "" {
				usedAnimated[obj.AnimFile] = true
			}
		}
		anims := make(map[string]*StageAnimation)
		for i := range def.Layout.Assets {
			asset := &def.Layout.Assets[i]
			if !usedAnimated[asset.File] && (asset.AnimFile == "" || !usedAnimated[asset.AnimFile]) {
				continue
			}
			anim := composeStageAnimation(asset, load)
			if anim == nil {
				continue
			}
			if asset.File != "" {
				anims[asset.File] = anim
			}
			if asset.AnimFile != "" {
				anims[asset.AnimFile] = anim
			}
		}
		if len(anims) > 0 {
			m.stageAnims[def.ID] = anims
		}
	}
	m.stageLayers[def.ID] = items
	m.stageImages[def.ID] = images
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

// StageObjectImage returns the texture for one placed stage object. Paths are
// relative to the game asset root (the same root ImageCache uses).
func (m *AssetManager) StageObjectImage(stageID, path string) *ebiten.Image {
	if images, ok := m.stageImages[stageID]; ok {
		return images[path]
	}
	return nil
}

// StageObjectAnimation returns the pre-composed animation for a placed stage
// object, or nil when the object is static or its animation failed to load.
func (m *AssetManager) StageObjectAnimation(stageID string, obj *stage.Object) *StageAnimation {
	if obj == nil || (obj.Category != "anim_prop" && obj.Category != "anim_block" && obj.AnimFile == "") {
		return nil
	}
	byKey := m.stageAnims[stageID]
	if byKey == nil {
		return nil
	}
	if anim := byKey[obj.File]; anim != nil {
		return anim
	}
	if obj.AnimFile != "" {
		return byKey[obj.AnimFile]
	}
	return nil
}

func (m *AssetManager) Tint() *TintController {
	return m.tint
}

func (m *AssetManager) UnloadStage(stageID string) {
	for _, item := range m.stageLayers[stageID] {
		if item.Entry.FrameCount > 1 {
			delete(m.frames, item.Entry.Key)
		}
	}
	released := make(map[*StageAnimation]bool, len(m.stageAnims[stageID]))
	for _, anim := range m.stageAnims[stageID] {
		if anim == nil || released[anim] {
			continue
		}
		released[anim] = true
		for _, img := range anim.Frames {
			if img != nil {
				img.Deallocate()
			}
		}
	}
	for path := range m.stageImages[stageID] {
		m.releasePath(path)
	}
	delete(m.stageLayers, stageID)
	delete(m.stageImages, stageID)
	delete(m.stageAnims, stageID)
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
