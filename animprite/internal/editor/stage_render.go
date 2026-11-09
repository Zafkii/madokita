package editor

import (
	"math"

	"animprite/internal/canvas"
	"animprite/internal/project"
)

func (a *EditorApp) buildStageRenders() {
	stage := &a.proj.Stage
	renders := make([]canvas.SpriteRender, 0, len(stage.Objects))

	// Backgrounds render first; blocks render on top, matching hit-testing.
	for _, wantBlock := range []bool{false, true} {
		for i := range stage.Objects {
			obj := &stage.Objects[i]
			isBlock := obj.Category == project.StageCategoryBlock
			if isBlock != wantBlock {
				continue
			}
			renders = append(renders, a.stageObjectRender(obj, obj.Opacity))
		}
	}

	if a.st.dragAsset >= 0 && a.st.draggingFromPanel && a.st.dragAsset < len(stage.Assets) {
		asset := stage.Assets[a.st.dragAsset]
		w, h := a.stageDefaultSize(asset)
		ghost := project.StageObject{
			Category: asset.Category,
			File:     asset.File,
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
	a.canvas.SetHurtboxRenders(nil)

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
