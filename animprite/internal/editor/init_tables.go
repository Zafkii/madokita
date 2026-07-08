package editor

import (
	"fmt"

	"animprite/internal/project"
	"animprite/internal/theme"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func (a *EditorApp) initTables() {
	th := a.th

	a.animTable = ui.NewTable("Animation", []ui.TableColumn{
		{Label: "Name", Width: 0},
		{Label: "Frame", Width: 0},
	}, 4, th)
	a.spriteTable = ui.NewTable("Sprite", []ui.TableColumn{
		{Label: "Grade", Width: 0},
		{Label: "File", Width: 0},
		{Label: "Width", Width: 0},
		{Label: "Height", Width: 0},
		{Label: "Frame", Width: 0},
	}, 4, th)
	a.hurtboxTable = ui.NewTable("Hurtbox", []ui.TableColumn{
		{Label: "No.", Width: 0},
		{Label: "X", Width: 0},
		{Label: "Y", Width: 0},
		{Label: "Width", Width: 0},
		{Label: "Height", Width: 0},
		{Label: "dmg-mult", Width: 0},
	}, 4, th)
	a.hitboxTable = ui.NewTable("Hitbox", []ui.TableColumn{
		{Label: "No.", Width: 0},
		{Label: "Width", Width: 0},
		{Label: "Height", Width: 0},
	}, 4, th)

	// ── Animation table buttons ──
	addAnimBtn := ui.NewButton(0, 0, 40, topPanelBtnH, "+ Add", th)
	addAnimBtn.BtnType = ui.BtnBlue
	addAnimBtn.OnClick = func() {
		a.saveSnapshot()
		n := len(a.proj.Animations)
		var defName string
		switch {
		case n == 0 && a.mode == modeAttack:
			defName = "slash"
		case n == 0:
			defName = "idle"
		case a.mode == modeAttack:
			defName = fmt.Sprintf("atk %d", n+1)
		default:
			defName = fmt.Sprintf("move %d", n+1)
		}
		a.proj.Animations = append(a.proj.Animations, project.AnimationRow{
			Name: defName, CurrentIdx: 0,
			Frames: []project.AnimationFrame{{Sprites: a.defaultFrameSprites()}},
			FPS:    14,
		})
		a.syncAnimBtns()
		a.syncLayout()
	}
	delAnimBtn := ui.NewButton(0, 0, 40, topPanelBtnH, "- Del", th)
	delAnimBtn.BtnType = ui.BtnRed
	delAnimBtn.OnClick = func() {
		a.saveSnapshot()
		if idx := a.animTable.SelectedIdx; idx >= 0 && idx < len(a.proj.Animations) {
			a.proj.Animations = append(a.proj.Animations[:idx], a.proj.Animations[idx+1:]...)
			a.syncAnimBtns()
			a.syncLayout()
		}
	}
	a.animTable.AddBtn = addAnimBtn
	a.animTable.RemoveBtn = delAnimBtn

	a.animTable.DrawRow = func(screen *ebiten.Image, p theme.Palette, tx, rowY, cw int, xs []int, idx int) {
		row := &a.proj.Animations[idx]
		a.animNameInputs[idx].X = tx + xs[0]
		a.animNameInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.animNameInputs[idx].W = xs[1] - xs[0] - 2
		a.animNameInputs[idx].H = topPanelBtnH
		a.animNameInputs[idx].Draw(screen)

		frameInputW := 24
		navGroupW := a.animFramePrevBtns[idx].W + 2 + frameInputW + 2 + ui.TextWidth("/99") + 2 +
			a.animFrameNextBtns[idx].W + btnGap + a.animAddFrameBtns[idx].W + btnGap + a.animRemoveFrameBtns[idx].W
		frameColW := (tx + cw - tableColPad) - (tx + xs[1])
		frameX := tx + xs[1] + (frameColW-navGroupW)/2

		frameNavY := rowY + (topTableRowH-topPanelBtnH)/2
		fx := frameX
		a.animFramePrevBtns[idx].X = fx
		a.animFramePrevBtns[idx].Y = frameNavY
		a.animFramePrevBtns[idx].Draw(screen)

		fx += a.animFramePrevBtns[idx].W + 2
		inp := a.animFrameInputs[idx]
		inp.X = fx
		inp.Y = frameNavY
		inp.W = frameInputW
		inp.H = topPanelBtnH
		inp.SetNumeric(float64(row.CurrentIdx + 1))
		inp.Max = float64(len(row.Frames))
		inp.Draw(screen)

		fx += inp.W + 2
		slashLabel := fmt.Sprintf("/%d ", len(row.Frames))
		ui.DrawText(screen, slashLabel, fx, rowY+3, 1, p.TextPrimary)

		fx += ui.TextWidth(slashLabel) + 2
		a.animFrameNextBtns[idx].X = fx
		a.animFrameNextBtns[idx].Y = frameNavY
		a.animFrameNextBtns[idx].Draw(screen)

		fx += a.animFrameNextBtns[idx].W + btnGap
		a.animAddFrameBtns[idx].X = fx
		a.animAddFrameBtns[idx].Y = frameNavY
		a.animAddFrameBtns[idx].Draw(screen)

		fx += a.animAddFrameBtns[idx].W + btnGap
		a.animRemoveFrameBtns[idx].X = fx
		a.animRemoveFrameBtns[idx].Y = frameNavY
		a.animRemoveFrameBtns[idx].Draw(screen)
	}

	// ── Sprite table buttons ──
	addSpriteBtn := ui.NewButton(0, 0, 40, topPanelBtnH, "+ Add", th)
	addSpriteBtn.BtnType = ui.BtnBlue
	addSpriteBtn.OnClick = func() {
		a.saveSnapshot()
		newIdx := len(a.proj.Sprites)
		a.proj.Sprites = append(a.proj.Sprites, project.SpriteRow{
			Name:       fmt.Sprintf("sprite %d", newIdx+1),
			File:       "",
			Width:      256,
			Height:     256,
			FrameCount: 1,
			CurrentIdx: 0,
			ScaleX:     1,
			ScaleY:     1,
			OriginX:    0.5,
			OriginY:    0.5,
		})
		entry := project.FrameSpriteEntry{
			SpriteIdx: newIdx,
			OriginX:   0.5, OriginY: 0.5,
			ScaleX: 1, ScaleY: 1,
		}
		for ai := range a.proj.Animations {
			for fi := range a.proj.Animations[ai].Frames {
				a.proj.Animations[ai].Frames[fi].Sprites = append(a.proj.Animations[ai].Frames[fi].Sprites, entry)
			}
		}
		a.syncSpriteBtns()
		a.syncLayout()
	}
	delSpriteBtn := ui.NewButton(0, 0, 40, topPanelBtnH, "- Del", th)
	delSpriteBtn.BtnType = ui.BtnRed
	delSpriteBtn.OnClick = func() {
		a.saveSnapshot()
		if idx := a.spriteTable.SelectedIdx; idx >= 0 && idx < len(a.proj.Sprites) {
			if old, ok := a.loadedSprites[idx]; ok {
				old.Deallocate()
				delete(a.loadedSprites, idx)
			}
			for i := idx + 1; i < len(a.proj.Sprites); i++ {
				if img, ok := a.loadedSprites[i]; ok {
					a.loadedSprites[i-1] = img
				} else {
					delete(a.loadedSprites, i-1)
				}
				delete(a.loadedSprites, i)
			}
			a.proj.Sprites = append(a.proj.Sprites[:idx], a.proj.Sprites[idx+1:]...)
			a.syncSpriteBtns()
			a.syncLayout()
		}
	}
	a.spriteTable.AddBtn = addSpriteBtn
	a.spriteTable.RemoveBtn = delSpriteBtn

	a.spriteTable.DrawRow = func(screen *ebiten.Image, p theme.Palette, tx, rowY, cw int, xs []int, idx int) {
		row := &a.proj.Sprites[idx]

		ui.DrawText(screen, row.Name, tx+xs[0]+2, rowY+4, 1, p.TextPrimary)

		fileX := tx + xs[1]
		frameNavY := rowY + (topTableRowH-topPanelBtnH)/2
		a.spriteBrowseBtns[idx].X = fileX
		a.spriteBrowseBtns[idx].Y = frameNavY
		a.spriteBrowseBtns[idx].Draw(screen)
		if row.File != "" {
			fileTextX := fileX + a.spriteBrowseBtns[idx].W + 2
			fileTextW := (tx + xs[2] - 2) - fileTextX
			if fileTextW > 4 {
				ui.DrawText(screen, ui.TruncateText(row.File, fileTextW), fileTextX, rowY+3, 1, p.TextPrimary)
			}
		}

		iw := xs[3] - xs[2] - 2
		if iw < 20 {
			iw = 20
		}
		a.spriteWidthInputs[idx].X = tx + xs[2]
		a.spriteWidthInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.spriteWidthInputs[idx].W = iw
		a.spriteWidthInputs[idx].Draw(screen)

		ih := xs[4] - xs[3] - 2
		if ih < 20 {
			ih = 20
		}
		a.spriteHeightInputs[idx].X = tx + xs[3]
		a.spriteHeightInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.spriteHeightInputs[idx].W = ih
		a.spriteHeightInputs[idx].Draw(screen)

		fx := tx + xs[4]
		a.spriteFramePrevBtns[idx].X = fx
		a.spriteFramePrevBtns[idx].Y = frameNavY
		a.spriteFramePrevBtns[idx].Draw(screen)

		fx += a.spriteFramePrevBtns[idx].W + 2
		frameLabel := fmt.Sprintf(" %d/%d ", row.CurrentIdx+1, row.FrameCount)
		ui.DrawText(screen, frameLabel, fx, rowY+3, 1, p.TextPrimary)

		fx += ui.TextWidth(frameLabel) + 2
		a.spriteFrameNextBtns[idx].X = fx
		a.spriteFrameNextBtns[idx].Y = frameNavY
		a.spriteFrameNextBtns[idx].Draw(screen)
	}

	// ── Hurtbox table buttons ──
	addHbBtn := ui.NewButton(0, 0, 40, topPanelBtnH, "+ Add", th)
	addHbBtn.BtnType = ui.BtnBlue
	addHbBtn.OnClick = func() {
		a.flushInputsToData()
		a.saveSnapshot()
		animIdx := a.animTable.SelectedIdx
		if animIdx < 0 {
			animIdx = a.hurtboxAnimCtx
		}
		if animIdx < 0 || animIdx >= len(a.proj.Animations) {
			return
		}
		anim := &a.proj.Animations[animIdx]
		defaultHb := project.HurtboxRow{Width: 32, Height: 32}
		for fi := range anim.Frames {
			for si := range anim.Frames[fi].Sprites {
				anim.Frames[fi].Sprites[si].Hurtboxes = append(
					anim.Frames[fi].Sprites[si].Hurtboxes, defaultHb)
			}
		}
		entry := a.currentFrameSpriteEntry()
		newIdx := 0
		if entry != nil {
			newIdx = len(entry.Hurtboxes) - 1
		}
		a.navigateToHurtbox(newIdx)
	}
	delHbBtn := ui.NewButton(0, 0, 40, topPanelBtnH, "- Del", th)
	delHbBtn.BtnType = ui.BtnRed
	delHbBtn.OnClick = func() {
		a.flushInputsToData()
		a.saveSnapshot()
		idx := a.hurtboxTable.SelectedIdx
		if idx < 0 {
			return
		}
		animIdx := a.animTable.SelectedIdx
		if animIdx < 0 {
			animIdx = a.hurtboxAnimCtx
		}
		if animIdx < 0 || animIdx >= len(a.proj.Animations) {
			return
		}
		anim := &a.proj.Animations[animIdx]
		for fi := range anim.Frames {
			for si := range anim.Frames[fi].Sprites {
				hbs := anim.Frames[fi].Sprites[si].Hurtboxes
				if idx < len(hbs) {
					anim.Frames[fi].Sprites[si].Hurtboxes = append(hbs[:idx], hbs[idx+1:]...)
				}
			}
		}
		a.navigateToHurtbox(a.hurtboxTable.SelectedIdx)
	}
	copyHbBtn := ui.NewButton(0, 0, 110, topPanelBtnH, "\U000F04AE Copy Prev", th)
	copyHbBtn.OnClick = func() {
		a.saveSnapshot()
		animIdx := a.animTable.SelectedIdx
		if animIdx < 0 {
			animIdx = a.hurtboxAnimCtx
		}
		if animIdx < 0 || animIdx >= len(a.proj.Animations) {
			return
		}
		anim := &a.proj.Animations[animIdx]
		if anim.CurrentIdx < 0 || anim.CurrentIdx >= len(anim.Frames) {
			return
		}
		entry := a.currentFrameSpriteEntry()
		if entry == nil {
			return
		}
		if len(entry.Hurtboxes) > 0 {
			return
		}
		prevIdx := anim.CurrentIdx - 1
		if prevIdx < 0 || prevIdx >= len(anim.Frames) {
			return
		}
		prev := anim.Frames[prevIdx]
		prevEntry := a.frameSpriteEntry(&prev, a.spriteEditIdx)
		if prevEntry == nil || len(prevEntry.Hurtboxes) == 0 {
			return
		}
		entry.Hurtboxes = make([]project.HurtboxRow, len(prevEntry.Hurtboxes))
		copy(entry.Hurtboxes, prevEntry.Hurtboxes)
		a.syncHurtboxBtns()
		a.syncLayout()
	}
	a.hurtboxTable.AddBtn = addHbBtn
	a.hurtboxTable.RemoveBtn = delHbBtn
	a.hurtboxTable.ExtraBtns = []*ui.Button{copyHbBtn}

	a.hurtboxTable.DrawRow = func(screen *ebiten.Image, p theme.Palette, tx, rowY, cw int, xs []int, idx int) {
		ui.DrawText(screen, fmt.Sprintf("%d", idx+1), tx+xs[0], rowY+3, 1, p.TextPrimary)
		iw := xs[2] - xs[1] - 2
		if iw < 20 {
			iw = 20
		}
		a.hurtboxXInputs[idx].X = tx + xs[1]
		a.hurtboxXInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.hurtboxXInputs[idx].W = iw
		a.hurtboxXInputs[idx].H = topPanelBtnH
		a.hurtboxXInputs[idx].Draw(screen)

		ih := xs[3] - xs[2] - 2
		if ih < 20 {
			ih = 20
		}
		a.hurtboxYInputs[idx].X = tx + xs[2]
		a.hurtboxYInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.hurtboxYInputs[idx].W = ih
		a.hurtboxYInputs[idx].H = topPanelBtnH
		a.hurtboxYInputs[idx].Draw(screen)

		iw2 := xs[4] - xs[3] - 2
		if iw2 < 20 {
			iw2 = 20
		}
		a.hurtboxWidthInputs[idx].X = tx + xs[3]
		a.hurtboxWidthInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.hurtboxWidthInputs[idx].W = iw2
		a.hurtboxWidthInputs[idx].H = topPanelBtnH
		a.hurtboxWidthInputs[idx].Draw(screen)

		ih2 := xs[5] - xs[4] - 2
		if ih2 < 20 {
			ih2 = 20
		}
		a.hurtboxHeightInputs[idx].X = tx + xs[4]
		a.hurtboxHeightInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.hurtboxHeightInputs[idx].W = ih2
		a.hurtboxHeightInputs[idx].H = topPanelBtnH
		a.hurtboxHeightInputs[idx].Draw(screen)

		dm := tx + xs[5]
		a.hurtboxDmgMultInputs[idx].X = dm
		a.hurtboxDmgMultInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.hurtboxDmgMultInputs[idx].W = (tx + cw - tableColPad) - dm - 2
		a.hurtboxDmgMultInputs[idx].H = topPanelBtnH
		a.hurtboxDmgMultInputs[idx].Draw(screen)
	}

	// ── Hitbox table buttons ──
	addHtBtn := ui.NewButton(0, 0, 40, topPanelBtnH, "+ Add", th)
	addHtBtn.BtnType = ui.BtnBlue
	addHtBtn.OnClick = func() {
		a.saveSnapshot()
		a.proj.HitDefs = append(a.proj.HitDefs, project.HitboxRow{})
		a.syncHitboxBtns()
		a.syncLayout()
	}
	delHtBtn := ui.NewButton(0, 0, 40, topPanelBtnH, "- Del", th)
	delHtBtn.BtnType = ui.BtnRed
	delHtBtn.OnClick = func() {
		a.saveSnapshot()
		if idx := a.hitboxTable.SelectedIdx; idx >= 0 && idx < len(a.proj.HitDefs) {
			a.proj.HitDefs = append(a.proj.HitDefs[:idx], a.proj.HitDefs[idx+1:]...)
			a.syncHitboxBtns()
			a.syncLayout()
		}
	}
	copyHtBtn := ui.NewButton(0, 0, 110, topPanelBtnH, "\U000F04AE Copy Prev", th)
	copyHtBtn.OnClick = func() {
		a.saveSnapshot()
		if len(a.proj.HitDefs) >= 2 {
			last := len(a.proj.HitDefs) - 1
			a.proj.HitDefs[last] = a.proj.HitDefs[last-1]
		}
	}
	a.hitboxTable.AddBtn = addHtBtn
	a.hitboxTable.RemoveBtn = delHtBtn
	a.hitboxTable.ExtraBtns = []*ui.Button{copyHtBtn}

	a.hitboxTable.DrawRow = func(screen *ebiten.Image, p theme.Palette, tx, rowY, cw int, xs []int, idx int) {
		ui.DrawText(screen, fmt.Sprintf("%d", idx+1), tx+xs[0], rowY+3, 1, p.TextPrimary)
		iw := xs[2] - xs[1] - 2
		if iw < 20 {
			iw = 20
		}
		a.hitboxWidthInputs[idx].X = tx + xs[1]
		a.hitboxWidthInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.hitboxWidthInputs[idx].W = iw
		a.hitboxWidthInputs[idx].H = topPanelBtnH
		a.hitboxWidthInputs[idx].Draw(screen)

		ih := (tx + cw - tableColPad) - (tx + xs[2]) - 2
		if ih < 20 {
			ih = 20
		}
		a.hitboxHeightInputs[idx].X = tx + xs[2]
		a.hitboxHeightInputs[idx].Y = rowY + (topTableRowH-topPanelBtnH)/2
		a.hitboxHeightInputs[idx].W = ih
		a.hitboxHeightInputs[idx].H = topPanelBtnH
		a.hitboxHeightInputs[idx].Draw(screen)
	}

	a.syncAnimBtns()
	a.syncSpriteBtns()
	a.syncHurtboxBtns()
	a.syncHitboxBtns()

	if len(a.proj.Animations) > 0 {
		a.animTable.SelectedIdx = 0
	}
}
