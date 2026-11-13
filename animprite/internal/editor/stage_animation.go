package editor

import (
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"animprite/internal/project"

	"github.com/hajimehoshi/ebiten/v2"
)

// stageAnimation is the pre-composed idle animation of one animated stage
// asset. Frames are independent images so the canvas can draw the current
// frame with the same transform used by static stage objects.
type stageAnimation struct {
	Frames []*ebiten.Image
	FPS    float64
	Loop   bool
}

func (a *EditorApp) stageAnimationKey(asset project.StageAsset) string {
	return a.stageImageKey(asset.AnimFile) + "|" + asset.AnimName + "|" + a.stageImageKey(asset.File)
}

func (a *EditorApp) stageAnimationForAsset(asset project.StageAsset) *stageAnimation {
	if !project.StageCategoryIsAnimated(asset.Category) || asset.AnimFile == "" {
		return nil
	}
	if a.st.animations == nil {
		a.st.animations = make(map[string]*stageAnimation)
	}
	key := a.stageAnimationKey(asset)
	if anim, ok := a.st.animations[key]; ok {
		return anim
	}

	fail := func() *stageAnimation {
		a.st.animations[key] = &stageAnimation{}
		return a.st.animations[key]
	}

	mv, err := ImportMovement(asset.AnimFile)
	if err != nil {
		return fail()
	}
	animRow := pickStageAnimation(mv.Animations, asset.AnimName)
	if animRow == nil || len(animRow.Frames) == 0 {
		return fail()
	}

	defs := append([]project.SpriteRow(nil), mv.Sprites...)
	if len(defs) == 0 {
		defs = []project.SpriteRow{{
			File:       asset.File,
			Width:      asset.Width,
			Height:     asset.Height,
			FrameCount: asset.FrameCount,
		}}
	}

	dir := filepath.Dir(asset.AnimFile)
	sheets := make([]*ebiten.Image, len(defs))
	for i, def := range defs {
		file := def.File
		if file == "" {
			file = asset.File
		}
		path := resolveStagePath(dir, file)
		img, loadErr := a.loadStageImage(path)
		if loadErr != nil {
			return fail()
		}
		sheets[i] = img
	}

	frameW, frameH := stageAnimationSize(defs, *animRow)
	if frameW <= 0 || frameH <= 0 {
		if len(defs) > 0 && defs[0].Width > 0 && defs[0].Height > 0 {
			frameW, frameH = defs[0].Width, defs[0].Height
		} else if len(sheets) > 0 && sheets[0] != nil {
			b := sheets[0].Bounds()
			frameW, frameH = b.Dx(), b.Dy()
		}
	}
	if frameW <= 0 || frameH <= 0 {
		return fail()
	}

	anim := &stageAnimation{FPS: animRow.FPS, Loop: animRow.Loop}
	anim.Frames = make([]*ebiten.Image, 0, len(animRow.Frames))
	for _, frame := range animRow.Frames {
		img := ebiten.NewImage(frameW, frameH)
		drawStageAnimationFrame(img, sheets, defs, frame)
		anim.Frames = append(anim.Frames, img)
	}
	a.st.animations[key] = anim
	return anim
}

func (a *EditorApp) stageAnimationForObject(obj *project.StageObject) *stageAnimation {
	if obj == nil || !project.StageCategoryIsAnimated(obj.Category) {
		return nil
	}
	asset := project.StageAsset{
		Category: obj.Category,
		File:     obj.File,
		AnimFile: obj.AnimFile,
		AnimName: obj.AnimName,
	}
	if found := stageAssetByFile(a.proj.Stage.Assets, obj.File); found != nil {
		if asset.AnimFile == "" {
			asset.AnimFile = found.AnimFile
		}
		if asset.AnimName == "" {
			asset.AnimName = found.AnimName
		}
		if asset.Width <= 0 {
			asset.Width = found.Width
		}
		if asset.Height <= 0 {
			asset.Height = found.Height
		}
	}
	if asset.File == "" && asset.AnimFile != "" {
		if found := stageAssetByAnimFile(a.proj.Stage.Assets, asset.AnimFile); found != nil {
			asset.File = found.File
			asset.Width = found.Width
			asset.Height = found.Height
			if asset.AnimName == "" {
				asset.AnimName = found.AnimName
			}
		}
	}
	return a.stageAnimationForAsset(asset)
}

func stageAssetByAnimFile(assets []project.StageAsset, animFile string) *project.StageAsset {
	if animFile == "" {
		return nil
	}
	clean := filepath.Clean(animFile)
	for i := range assets {
		if assets[i].AnimFile == "" {
			continue
		}
		if filepath.Clean(assets[i].AnimFile) == clean {
			return &assets[i]
		}
	}
	return nil
}

func (a *EditorApp) stageAnimationFrame(anim *stageAnimation) *ebiten.Image {
	if anim == nil || len(anim.Frames) == 0 {
		return nil
	}
	fps := anim.FPS
	if fps <= 0 {
		fps = 1
	}
	idx := int(a.st.animClock * fps)
	if anim.Loop {
		idx %= len(anim.Frames)
	} else if idx >= len(anim.Frames) {
		idx = len(anim.Frames) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return anim.Frames[idx]
}

func (a *EditorApp) advanceStageAnimations() {
	if !a.isStageMode() {
		a.st.animClock = 0
		return
	}
	tps := ebiten.ActualTPS()
	if tps <= 0 {
		tps = 60
	}
	a.st.animClock += 1 / tps
}

func (a *EditorApp) clearStageAnimations() {
	for key, anim := range a.st.animations {
		if anim != nil {
			for _, img := range anim.Frames {
				if img != nil {
					img.Deallocate()
				}
			}
		}
		delete(a.st.animations, key)
	}
	a.st.animations = make(map[string]*stageAnimation)
}

func (a *EditorApp) stageAssetThumb(asset project.StageAsset) *ebiten.Image {
	if project.StageCategoryIsAnimated(asset.Category) {
		if anim := a.stageAnimationForAsset(asset); anim != nil {
			return a.stageAnimationFrame(anim)
		}
		return nil
	}
	return a.stageImage(asset.File)
}

// animatedCategoryHint classifies an animated asset from its display name,
// spritesheet basename, and movement folder/file. It avoids leaking unrelated
// parent directories into the block/prop heuristic.
func animatedCategoryHint(name, file, animFile string) string {
	parts := []string{name, filepath.Base(file)}
	if animFile != "" {
		parts = append(parts, filepath.Base(filepath.Dir(animFile)), filepath.Base(animFile))
	}
	return project.StageCategoryForAnimatedFile(strings.Join(parts, " "))
}

func stageAssetByFile(assets []project.StageAsset, file string) *project.StageAsset {
	if file == "" {
		return nil
	}
	clean := filepath.Clean(file)
	for i := range assets {
		if assets[i].File == "" {
			continue
		}
		if filepath.Clean(assets[i].File) == clean {
			return &assets[i]
		}
	}
	return nil
}

// scanAnimatedStageAsset looks for a movement definition inside a stage
// subfolder (for example tree_moves/moves.go) and turns its idle animation
// into a stage asset.
func scanAnimatedStageAsset(dir string) (project.StageAsset, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return project.StageAsset{}, false
	}

	var goFiles, pngFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(entry.Name())) {
		case ".go":
			goFiles = append(goFiles, entry.Name())
		case ".png":
			pngFiles = append(pngFiles, entry.Name())
		}
	}
	sort.Strings(goFiles)
	sort.Strings(pngFiles)

	for _, goName := range goFiles {
		movementPath := filepath.Join(dir, goName)
		if abs, absErr := filepath.Abs(movementPath); absErr == nil {
			movementPath = filepath.Clean(abs)
		}
		mv, importErr := ImportMovement(movementPath)
		if importErr != nil {
			continue
		}
		anim := pickStageAnimation(mv.Animations, "")
		if anim == nil || len(anim.Frames) == 0 {
			continue
		}

		spriteFile := ""
		if len(mv.Sprites) > 0 {
			spriteFile = mv.Sprites[0].File
		}
		if spriteFile == "" && len(pngFiles) > 0 {
			spriteFile = pngFiles[0]
		}
		if spriteFile == "" {
			continue
		}
		spritePath := resolveStagePath(dir, spriteFile)
		if abs, absErr := filepath.Abs(spritePath); absErr == nil {
			spritePath = filepath.Clean(abs)
		}

		f, openErr := os.Open(spritePath)
		if openErr != nil {
			continue
		}
		cfg, decodeErr := png.DecodeConfig(f)
		f.Close()
		if decodeErr != nil {
			continue
		}

		defs := mv.Sprites
		if len(defs) == 0 && len(pngFiles) > 0 {
			defs = []project.SpriteRow{{
				File:   pngFiles[0],
				Width:  cfg.Width,
				Height: cfg.Height,
			}}
		}
		w, h := stageAnimationSize(defs, *anim)
		if w <= 0 || h <= 0 {
			w, h = cfg.Width, cfg.Height
			if len(defs) > 0 && defs[0].Width > 0 && defs[0].Height > 0 {
				w, h = defs[0].Width, defs[0].Height
			}
		}
		if w <= 0 || h <= 0 {
			continue
		}

		name := animatedStageAssetName(filepath.Base(dir), spritePath)
		category := project.StageCategoryForAnimatedFile(strings.Join([]string{
			filepath.Base(dir),
			filepath.Base(spritePath),
			anim.Name,
			mv.AssetName,
			mv.AssetKey,
		}, " "))

		return project.StageAsset{
			Category:   category,
			Name:       name,
			File:       spritePath,
			Width:      w,
			Height:     h,
			FrameCount: len(anim.Frames),
			AnimFile:   movementPath,
			AnimName:   anim.Name,
		}, true
	}

	return project.StageAsset{}, false
}

func animatedStageAssetName(dirName, spriteFile string) string {
	name := dirName
	for _, suffix := range []string{"_moves", "-moves", "_anim", "-anim", "_animation", "-animation"} {
		name = strings.TrimSuffix(name, suffix)
	}
	if name != "" {
		return name
	}
	return strings.TrimSuffix(filepath.Base(spriteFile), filepath.Ext(spriteFile))
}

func pickStageAnimation(anims []project.AnimationRow, preferred string) *project.AnimationRow {
	if len(anims) == 0 {
		return nil
	}
	if preferred != "" {
		for i := range anims {
			if anims[i].Name == preferred && len(anims[i].Frames) > 0 {
				return &anims[i]
			}
		}
	}
	var first *project.AnimationRow
	for i := range anims {
		if len(anims[i].Frames) == 0 {
			continue
		}
		if first == nil {
			first = &anims[i]
		}
		if strings.EqualFold(anims[i].Name, "idle") {
			return &anims[i]
		}
	}
	return first
}

// stageAnimationSize returns the smallest symmetric frame canvas that fits
// every sprite in the animation. The movement origin stays at the canvas
// center, so the stage object's X/Y remains the visual center.
func stageAnimationSize(defs []project.SpriteRow, anim project.AnimationRow) (int, int) {
	if len(defs) == 0 {
		return 0, 0
	}
	halfW, halfH := 0.0, 0.0
	for _, frame := range anim.Frames {
		for _, entry := range frame.Sprites {
			if entry.SpriteIdx < 0 || entry.SpriteIdx >= len(defs) {
				continue
			}
			def := defs[entry.SpriteIdx]
			fw := float64(def.Width)
			fh := float64(def.Height)
			if fw <= 0 || fh <= 0 {
				continue
			}
			originX := entry.OriginX * fw
			originY := entry.OriginY * fh
			rot := entry.Rotation * math.Pi / 180
			cos := math.Cos(rot)
			sin := math.Sin(rot)
			corners := [4][2]float64{
				{-originX, -originY},
				{fw - originX, -originY},
				{fw - originX, fh - originY},
				{-originX, fh - originY},
			}
			for _, corner := range corners {
				x := corner[0] * entry.ScaleX
				y := corner[1] * entry.ScaleY
				rx := x*cos - y*sin + entry.OffsetX
				ry := x*sin + y*cos + entry.OffsetY
				halfW = math.Max(halfW, math.Abs(rx))
				halfH = math.Max(halfH, math.Abs(ry))
			}
		}
	}
	w := int(math.Ceil(halfW * 2))
	h := int(math.Ceil(halfH * 2))
	if w <= 0 || h <= 0 {
		return 0, 0
	}
	return w, h
}

func drawStageAnimationFrame(dst *ebiten.Image, sheets []*ebiten.Image, defs []project.SpriteRow, frame project.AnimationFrame) {
	cx := float64(dst.Bounds().Dx()) / 2
	cy := float64(dst.Bounds().Dy()) / 2
	for _, entry := range frame.Sprites {
		if entry.SpriteIdx < 0 || entry.SpriteIdx >= len(defs) || entry.SpriteIdx >= len(sheets) {
			continue
		}
		sheet := sheets[entry.SpriteIdx]
		def := defs[entry.SpriteIdx]
		if sheet == nil || def.Width <= 0 || def.Height <= 0 {
			continue
		}
		fw, fh := def.Width, def.Height
		cols := sheet.Bounds().Dx() / fw
		if cols < 1 {
			cols = 1
		}
		idx := entry.SpriteFrameIdx
		if idx < 0 {
			idx = 0
		}
		x := (idx % cols) * fw
		y := (idx / cols) * fh
		if x+fw > sheet.Bounds().Dx() || y+fh > sheet.Bounds().Dy() {
			continue
		}
		sub, ok := sheet.SubImage(image.Rect(x, y, x+fw, y+fh)).(*ebiten.Image)
		if !ok {
			continue
		}

		originX := entry.OriginX * float64(fw)
		originY := entry.OriginY * float64(fh)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-originX, -originY)
		op.GeoM.Scale(entry.ScaleX, entry.ScaleY)
		op.GeoM.Rotate(entry.Rotation * math.Pi / 180)
		op.GeoM.Translate(cx+entry.OffsetX, cy+entry.OffsetY)
		dst.DrawImage(sub, op)
	}
}
