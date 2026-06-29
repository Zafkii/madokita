package project

func DeepCopy(src *ProjectData) ProjectData {
	dst := ProjectData{
		AssetName:      src.AssetName,
		AssetKey:       src.AssetKey,
		DefaultOriginX: src.DefaultOriginX,
		DefaultOriginY: src.DefaultOriginY,
		Animations:     make([]AnimationRow, len(src.Animations)),
		Sprites:        make([]SpriteRow, len(src.Sprites)),
		HitDefs:        append([]HitboxRow(nil), src.HitDefs...),
	}
	for i := range src.Animations {
		dst.Animations[i] = src.Animations[i]
		dst.Animations[i].Frames = make([]AnimationFrame, len(src.Animations[i].Frames))
		for j := range src.Animations[i].Frames {
			dst.Animations[i].Frames[j] = src.Animations[i].Frames[j]
			srcF := src.Animations[i].Frames[j]
			dst.Animations[i].Frames[j].Sprites = make([]FrameSpriteEntry, len(srcF.Sprites))
			copy(dst.Animations[i].Frames[j].Sprites, srcF.Sprites)
			for k := range srcF.Sprites {
				dst.Animations[i].Frames[j].Sprites[k].Hurtboxes = append([]HurtboxRow(nil), srcF.Sprites[k].Hurtboxes...)
			}
		}
	}
	copy(dst.Sprites, src.Sprites)
	return dst
}
