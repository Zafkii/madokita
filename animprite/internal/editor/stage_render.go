package editor

import (
	"math"

	"animprite/internal/canvas"
	"animprite/internal/project"
)

const stageInactiveLayerOpacity = 0.25

func (a *EditorApp) buildStageRenders() {
	stage := &a.proj.Stage
	if len(stage.Layers) == 0 {
		stage.NormalizeLayers()
	}
	renders := make([]canvas.SpriteRender, 0, len(stage.Objects))

	// Layers render bottom to top. Inside one layer backgrounds render first,
	// static props and then animated props above them, static blocks and then
	// animated blocks on top. Hit-testing inside that layer walks this order
	// backwards.
	for _, layer := range stage.Layers {
		for _, wantCategory := range []string{
			project.StageCategoryBG,
			project.StageCategoryProp,
			project.StageCategoryAnimProp,
			project.StageCategoryBlock,
			project.StageCategoryAnimBlock,
		} {
			for i := range stage.Objects {
				obj := &stage.Objects[i]
				if stage.LayerNameFor(*obj) != layer.Name {
					continue
				}
				if obj.Category != wantCategory {
					continue
				}
				renders = append(renders, a.stageObjectRender(obj, a.stageObjectOpacity(obj)))
			}
		}
	}

	if a.st.dragAsset >= 0 && a.st.draggingFromPanel && a.st.dragAsset < len(stage.Assets) {
		asset := stage.Assets[a.st.dragAsset]
		w, h := a.stageDefaultSize(asset)
		ghost := project.StageObject{
			Category: asset.Category,
			File:     asset.File,
			AnimFile: asset.AnimFile,
			AnimName: asset.AnimName,
			X:        a.st.ghostWX,
			Y:        a.st.ghostWY,
			Width:    w,
			Height:   h,
			Opacity:  0.65,
		}
		renders = append(renders, a.stageObjectRender(&ghost, 0.65))
	}

	if len(renders) == 0 {
		renders = nil
	}
	a.canvas.SetSpriteRenders(renders)
	a.canvas.SetHurtboxRenders(a.stageCollisionRenders())

	extras := make([]canvas.SelectionRect, 0, a.stageSelectionCount())
	for _, idx := range a.stageSelectedIndices() {
		if idx == a.st.selectedObj {
			continue
		}
		obj := &stage.Objects[idx]
		extras = append(extras, canvas.SelectionRect{
			X:        obj.X - obj.Width/2,
			Y:        obj.Y - obj.Height/2,
			W:        obj.Width,
			H:        obj.Height,
			Rotation: obj.Rotation * math.Pi / 180,
			Visible:  true,
		})
	}
	a.canvas.SetExtraSelections(extras)

	if obj := a.selectedStageObject(); obj != nil {
		a.canvas.SetSelectionRect(canvas.SelectionRect{
			X:        obj.X - obj.Width/2,
			Y:        obj.Y - obj.Height/2,
			W:        obj.Width,
			H:        obj.Height,
			Rotation: obj.Rotation * math.Pi / 180,
			Visible:  true,
		})
		return
	}
	a.canvas.SetSelectionRect(canvas.SelectionRect{Visible: false})
}

func (a *EditorApp) stageObjectOpacity(obj *project.StageObject) float64 {
	opacity := obj.Opacity
	if a.st.activeLayer != "" && a.proj.Stage.LayerNameFor(*obj) != a.st.activeLayer {
		opacity *= stageInactiveLayerOpacity
	}
	return clamp(opacity, 0, 1)
}

func (a *EditorApp) stageObjectRender(obj *project.StageObject, opacity float64) canvas.SpriteRender {
	w := obj.Width
	h := obj.Height
	if w <= 0 {
		w = project.DefaultBlockSize
	}
	if h <= 0 {
		h = project.DefaultBlockSize
	}

	r := canvas.SpriteRender{
		OffsetX:  obj.X,
		OffsetY:  obj.Y,
		Rotation: obj.Rotation * math.Pi / 180,
		OriginX:  0.5,
		OriginY:  0.5,
		Opacity:  clamp(opacity, 0, 1),
	}

	img := a.stageImage(obj.File)
	if project.StageCategoryIsAnimated(obj.Category) {
		img = nil
		if anim := a.stageAnimationForObject(obj); anim != nil {
			img = a.stageAnimationFrame(anim)
		}
	}
	if img != nil {
		iw := float64(img.Bounds().Dx())
		ih := float64(img.Bounds().Dy())
		r.Image = img
		if iw > 0 {
			r.ScaleX = w / iw
		}
		if ih > 0 {
			r.ScaleY = h / ih
		}
		return r
	}

	r.ScaleX = 1
	r.ScaleY = 1
	r.PlaceholderW = w
	r.PlaceholderH = h
	return r
}

// stageBoxRender places an authored box in world space using the object's
// center and rotation. Box X/Y are offsets from that center.
func stageBoxRender(obj *project.StageObject, box project.StageBox, kind canvas.RenderKind, selected bool) canvas.HurtboxRender {
	rot := obj.Rotation * math.Pi / 180
	cos := math.Cos(rot)
	sin := math.Sin(rot)
	return canvas.HurtboxRender{
		OffsetX:     obj.X + box.X*cos - box.Y*sin,
		OffsetY:     obj.Y + box.X*sin + box.Y*cos,
		WorldWidth:  box.Width,
		WorldHeight: box.Height,
		Rotation:    rot + box.Rotation*math.Pi/180,
		Selected:    selected,
		Kind:        kind,
	}
}

// stageCollisionRenders builds the collision and hitbox overlays shown when
// the Layers panel's show-collisions toggle is on. Disabled blocks produce no
// overlays; blocks without authored collision fall back to their block size.
func (a *EditorApp) stageCollisionRenders() []canvas.HurtboxRender {
	if !a.st.showCollisions {
		return nil
	}
	stage := &a.proj.Stage
	var out []canvas.HurtboxRender
	for i := range stage.Objects {
		obj := &stage.Objects[i]
		collisions := obj.EffectiveCollisionBoxes()
		hitboxes := obj.EffectiveHitboxes()
		if len(collisions) == 0 && len(hitboxes) == 0 {
			continue
		}

		selectedObj := a.st.collisionEdit && i == a.st.selectedObj
		activeIdx := -1
		if a.st.collisionTable != nil {
			activeIdx = a.st.collisionTable.SelectedIdx
		}
		for j, box := range collisions {
			selected := selectedObj && a.st.collisionTab == "collision" && j == activeIdx
			out = append(out, stageBoxRender(obj, box, canvas.RenderCollision, selected))
		}
		for j, box := range hitboxes {
			selected := selectedObj && a.st.collisionTab == "hitbox" && j == activeIdx
			out = append(out, stageBoxRender(obj, box, canvas.RenderHitbox, selected))
		}
	}
	return out
}
