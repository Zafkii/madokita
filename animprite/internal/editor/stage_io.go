package editor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"animprite/internal/project"
)

// ExportStage writes a stage as a Go data file in the same style as the
// movement and attack files. Image paths are stored relative to the file so
// the stage folder stays portable.
func ExportStage(path string, stage project.StageData) error {
	stage.NormalizeLayers()
	var b strings.Builder

	b.WriteString("package stages\n\n")
	b.WriteString("import . \"madokita-gems/internal/stage\"\n\n")

	fmt.Fprintf(&b, "var %s = Stage{\n", stageVarName(stage.Name))
	fmt.Fprintf(&b, "\tName:   %q,\n", stage.Name)
	fmt.Fprintf(&b, "\tWidth:  %g,\n", stage.Width)
	fmt.Fprintf(&b, "\tHeight: %g,\n", stage.Height)
	b.WriteString("\tLayers: []Layer{\n")
	for _, layer := range stage.Layers {
		fmt.Fprintf(&b, "\t\t{Name: %q},\n", layer.Name)
	}
	b.WriteString("\t},\n")
	b.WriteString("\tAssets: []Asset{\n")
	dir := filepath.Dir(path)
	for _, asset := range stage.Assets {
		fmt.Fprintf(&b, "\t\t{Name: %q, File: %q, Category: %q, Width: %d, Height: %d, FrameCount: %d",
			asset.Name, relativeSpritePath(dir, asset.File), asset.Category,
			asset.Width, asset.Height, asset.FrameCount)
		if asset.AnimFile != "" {
			fmt.Fprintf(&b, ", AnimFile: %q", relativeSpritePath(dir, asset.AnimFile))
		}
		if asset.AnimName != "" {
			fmt.Fprintf(&b, ", AnimName: %q", asset.AnimName)
		}
		b.WriteString(stageAssetAnimationLiteral(asset, dir))
		b.WriteString("},\n")
	}
	b.WriteString("\t},\n")
	b.WriteString("\tObjects: []Object{\n")
	for _, obj := range stage.Objects {
		fmt.Fprintf(&b, "\t\t{Layer: %q, Category: %q, File: %q, X: %g, Y: %g, Width: %g, Height: %g, Rotation: %g, Opacity: %g",
			obj.Layer, obj.Category, relativeSpritePath(dir, obj.File),
			obj.X, obj.Y, obj.Width, obj.Height, obj.Rotation, obj.Opacity)
		if obj.AnimFile != "" {
			fmt.Fprintf(&b, ", AnimFile: %q", relativeSpritePath(dir, obj.AnimFile))
		}
		if obj.AnimName != "" {
			fmt.Fprintf(&b, ", AnimName: %q", obj.AnimName)
		}
		if obj.FlipX {
			b.WriteString(", FlipX: true")
		}
		if obj.CollisionDisabled {
			b.WriteString(", CollisionDisabled: true")
		}
		if len(obj.CollisionBoxes) > 0 {
			fmt.Fprintf(&b, ", CollisionBoxes: %s", stageBoxLiteral(obj.CollisionBoxes))
		}
		if len(obj.Hitboxes) > 0 {
			fmt.Fprintf(&b, ", Hitboxes: %s", stageBoxLiteral(obj.Hitboxes))
		}
		b.WriteString("},\n")
	}
	b.WriteString("\t},\n}\n")

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func stageBoxLiteral(boxes []project.StageBox) string {
	var b strings.Builder
	b.WriteString("[]Box{")
	for i, box := range boxes {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "{X: %g, Y: %g, Width: %g, Height: %g, Rotation: %g}",
			box.X, box.Y, box.Width, box.Height, box.Rotation)
	}
	b.WriteString("}")
	return b.String()
}

// stageAssetAnimationLiteral exports the movement animation that belongs to an
// animated asset. The runtime game cannot parse the editor's movement source,
// so the stage file carries the composited frame description directly. It
// returns an empty string when the movement cannot be loaded; the asset keeps
// AnimFile/AnimName and the editor can still reopen it.
func stageAssetAnimationLiteral(asset project.StageAsset, dir string) string {
	if asset.AnimFile == "" || !project.StageCategoryIsAnimated(asset.Category) {
		return ""
	}
	mv, err := ImportMovement(asset.AnimFile)
	if err != nil {
		return ""
	}
	anim := pickStageAnimation(mv.Animations, asset.AnimName)
	if anim == nil || len(anim.Frames) == 0 || len(mv.Sprites) == 0 {
		return ""
	}
	frameW, frameH := stageAnimationSize(mv.Sprites, *anim)
	if frameW <= 0 || frameH <= 0 {
		frameW, frameH = asset.Width, asset.Height
	}
	if frameW <= 0 || frameH <= 0 {
		return ""
	}

	animDir := filepath.Dir(asset.AnimFile)
	var b strings.Builder
	b.WriteString(", Sprites: []SpriteSheet{")
	for i, def := range mv.Sprites {
		if i > 0 {
			b.WriteString(", ")
		}
		sheetPath := resolveStagePath(animDir, def.File)
		if sheetPath == "" {
			sheetPath = asset.File
		}
		fmt.Fprintf(&b, "{File: %q, FrameW: %d, FrameH: %d, FrameCount: %d}",
			relativeSpritePath(dir, sheetPath), def.Width, def.Height, def.FrameCount)
	}
	fmt.Fprintf(&b, "}, Animation: &Animation{Name: %q, FPS: %g, Loop: %t, FrameW: %d, FrameH: %d, Frames: []AnimFrame{",
		anim.Name, anim.FPS, anim.Loop, frameW, frameH)
	for i, frame := range anim.Frames {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("{Sprites: []AnimSprite{")
		for j, entry := range frame.Sprites {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(stageAnimSpriteLiteral(entry))
		}
		b.WriteString("}}")
	}
	b.WriteString("}}")
	return b.String()
}

func stageAnimSpriteLiteral(entry project.FrameSpriteEntry) string {
	// Write every field: a zero-valued ScaleX/ScaleY or OriginX/OriginY in the
	// generated literal would hide or shift the sprite.
	return fmt.Sprintf(
		"{SpriteIdx: %d, SpriteFrameIdx: %d, OffsetX: %g, OffsetY: %g, Rotation: %g, ScaleX: %g, ScaleY: %g, OriginX: %g, OriginY: %g}",
		entry.SpriteIdx, entry.SpriteFrameIdx, entry.OffsetX, entry.OffsetY,
		entry.Rotation, entry.ScaleX, entry.ScaleY, entry.OriginX, entry.OriginY,
	)
}

// ImportStage reads a Go stage data file produced by ExportStage (or written
// by hand in the same shape) and resolves image paths relative to the file.
func ImportStage(path string) (project.StageData, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return project.StageData{}, fmt.Errorf("parse: %w", err)
	}

	dir := filepath.Dir(path)
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
				continue
			}
			cl, ok := vs.Values[0].(*ast.CompositeLit)
			if !ok || !isStageType(cl.Type) {
				continue
			}

			stage := project.StageData{
				Version: project.StageVersion,
				Name:    vs.Names[0].Name,
			}
			hasWidth, hasHeight := false, false
			for _, elt := range cl.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				switch exprString(kv.Key) {
				case "Name":
					stage.Name = stringLit(kv.Value)
				case "Width":
					stage.Width = floatLit(kv.Value)
					hasWidth = stage.Width > 0
				case "Height":
					stage.Height = floatLit(kv.Value)
					hasHeight = stage.Height > 0
				case "Layers":
					stage.Layers = parseStageLayers(kv.Value)
				case "Assets":
					stage.Assets = parseStageAssets(kv.Value, dir)
				case "Objects":
					stage.Objects = parseStageObjects(kv.Value, dir)
				}
			}
			if stage.Name == "" {
				stage.Name = vs.Names[0].Name
			}
			if !hasWidth || !hasHeight {
				width, height := stage.Width, stage.Height
				if stage.ApplySizeFromFirstBackground() {
					if hasWidth {
						stage.Width = width
					}
					if hasHeight {
						stage.Height = height
					}
				}
			}
			if stage.Width <= 0 {
				stage.Width = project.DefaultStageWidth
			}
			if stage.Height <= 0 {
				stage.Height = project.DefaultStageHeight
			}
			stage.NormalizeLayers()
			applyStageObjectDefaults(&stage)
			return stage, nil
		}
	}

	return project.StageData{}, fmt.Errorf("no Stage variable declaration found")
}

func isStageType(e ast.Expr) bool {
	name := exprString(e)
	return name == "Stage" || name == "StageDef" ||
		strings.HasSuffix(name, ".Stage") || strings.HasSuffix(name, ".StageDef")
}

func parseStageLayers(expr ast.Expr) []project.StageLayer {
	cl, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var layers []project.StageLayer
	for _, elt := range cl.Elts {
		lit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		layer := project.StageLayer{}
		for _, kv := range lit.Elts {
			kve, ok := kv.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if exprString(kve.Key) == "Name" {
				layer.Name = stringLit(kve.Value)
			}
		}
		if layer.Name != "" {
			layers = append(layers, layer)
		}
	}
	return layers
}

func parseStageAssets(expr ast.Expr, dir string) []project.StageAsset {
	cl, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var assets []project.StageAsset
	for _, elt := range cl.Elts {
		lit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		asset := project.StageAsset{FrameCount: 1}
		for _, kv := range lit.Elts {
			kve, ok := kv.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			switch exprString(kve.Key) {
			case "Name":
				asset.Name = stringLit(kve.Value)
			case "File":
				asset.File = resolveStagePath(dir, stringLit(kve.Value))
			case "Category":
				asset.Category = stringLit(kve.Value)
			case "Width":
				asset.Width = intLit(kve.Value)
			case "Height":
				asset.Height = intLit(kve.Value)
			case "FrameCount":
				asset.FrameCount = intLit(kve.Value)
			case "AnimFile":
				asset.AnimFile = resolveStagePath(dir, stringLit(kve.Value))
			case "AnimName":
				asset.AnimName = stringLit(kve.Value)
			}
		}
		if asset.FrameCount <= 0 {
			asset.FrameCount = 1
		}
		if asset.Category == "" {
			if asset.AnimFile != "" {
				asset.Category = animatedCategoryHint(asset.Name, asset.File, asset.AnimFile)
			} else {
				asset.Category = project.StageCategoryForFile(asset.File)
			}
		}
		if asset.Name == "" {
			asset.Name = strings.TrimSuffix(filepath.Base(asset.File), filepath.Ext(asset.File))
		}
		assets = append(assets, asset)
	}
	return assets
}

func parseStageObjects(expr ast.Expr, dir string) []project.StageObject {
	cl, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var objects []project.StageObject
	for _, elt := range cl.Elts {
		lit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		obj := project.StageObject{Opacity: 1}
		haveOpacity := false
		for _, kv := range lit.Elts {
			kve, ok := kv.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			switch exprString(kve.Key) {
			case "Layer":
				obj.Layer = stringLit(kve.Value)
			case "Category":
				obj.Category = stringLit(kve.Value)
			case "File":
				obj.File = resolveStagePath(dir, stringLit(kve.Value))
			case "X":
				obj.X = floatLit(kve.Value)
			case "Y":
				obj.Y = floatLit(kve.Value)
			case "Width":
				obj.Width = floatLit(kve.Value)
			case "Height":
				obj.Height = floatLit(kve.Value)
			case "Rotation":
				obj.Rotation = floatLit(kve.Value)
			case "Opacity":
				obj.Opacity = floatLit(kve.Value)
				haveOpacity = true
			case "AnimFile":
				obj.AnimFile = resolveStagePath(dir, stringLit(kve.Value))
			case "AnimName":
				obj.AnimName = stringLit(kve.Value)
			case "FlipX":
				obj.FlipX = exprString(kve.Value) == "true"
			case "CollisionDisabled":
				obj.CollisionDisabled = exprString(kve.Value) == "true"
			case "CollisionBoxes":
				obj.CollisionBoxes = parseStageBoxes(kve.Value)
			case "Hitboxes":
				obj.Hitboxes = parseStageBoxes(kve.Value)
			}
		}
		if !haveOpacity {
			obj.Opacity = 1
		}
		if obj.Category == "" {
			if obj.AnimFile != "" {
				obj.Category = animatedCategoryHint("", obj.File, obj.AnimFile)
			} else {
				obj.Category = project.StageCategoryForFile(obj.File)
			}
		}
		objects = append(objects, obj)
	}
	return objects
}

func parseStageBoxes(expr ast.Expr) []project.StageBox {
	cl, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var boxes []project.StageBox
	for _, elt := range cl.Elts {
		lit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		box := project.StageBox{}
		for _, kv := range lit.Elts {
			kve, ok := kv.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			switch exprString(kve.Key) {
			case "X":
				box.X = floatLit(kve.Value)
			case "Y":
				box.Y = floatLit(kve.Value)
			case "Width":
				box.Width = floatLit(kve.Value)
			case "Height":
				box.Height = floatLit(kve.Value)
			case "Rotation":
				box.Rotation = floatLit(kve.Value)
			}
		}
		boxes = append(boxes, box)
	}
	return boxes
}

// applyStageObjectDefaults fills object sizes from the matching asset when a
// hand-written config omits Width/Height.
func applyStageObjectDefaults(stage *project.StageData) {
	byFile := make(map[string]project.StageAsset, len(stage.Assets))
	byAnim := make(map[string]project.StageAsset, len(stage.Assets))
	for _, asset := range stage.Assets {
		if asset.File != "" {
			byFile[filepath.Clean(asset.File)] = asset
		}
		if asset.AnimFile != "" {
			byAnim[filepath.Clean(asset.AnimFile)] = asset
		}
	}
	for i := range stage.Objects {
		obj := &stage.Objects[i]
		var asset project.StageAsset
		var ok bool
		if obj.File != "" {
			asset, ok = byFile[filepath.Clean(obj.File)]
		}
		if !ok && obj.AnimFile != "" {
			asset, ok = byAnim[filepath.Clean(obj.AnimFile)]
		}
		if !ok {
			continue
		}
		if obj.File == "" {
			obj.File = asset.File
		}
		if obj.Width <= 0 && asset.Width > 0 {
			obj.Width = float64(asset.Width)
		}
		if obj.Height <= 0 && asset.Height > 0 {
			obj.Height = float64(asset.Height)
		}
		if obj.AnimFile == "" {
			obj.AnimFile = asset.AnimFile
		}
		if obj.AnimName == "" {
			obj.AnimName = asset.AnimName
		}
	}
}

// stageVarName turns a stage name into a valid Go identifier.
func stageVarName(name string) string {
	var b strings.Builder
	upper := true
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			if upper {
				b.WriteRune(r - 'a' + 'A')
			} else {
				b.WriteRune(r)
			}
			upper = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
			upper = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			upper = false
		default:
			upper = true
		}
	}
	out := b.String()
	if out == "" {
		return "Stage1"
	}
	if out[0] >= '0' && out[0] <= '9' {
		out = "Stage" + out
	}
	if out == "Stage" {
		out = "Stage1"
	}
	return out
}
