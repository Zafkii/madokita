package editor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"

	"animprite/internal/project"
)

func ExportMovement(path string, proj *project.ProjectData) error {
	var b strings.Builder

	b.WriteString("package movements\n\n")
	b.WriteString("import . \"madokita/internal/animation\"\n\n")

	fmt.Fprintf(&b, "var %s = Movement{\n", proj.AssetName)
	fmt.Fprintf(&b, "\tAssetKey:       %q,\n", proj.AssetKey)
	fmt.Fprintf(&b, "\tDefaultOriginX: %g,\n", proj.DefaultOriginX)
	fmt.Fprintf(&b, "\tDefaultOriginY: %g,\n", proj.DefaultOriginY)
	b.WriteString("\tAnimations: map[string]MovementAnimDef{\n")

	for _, anim := range proj.Animations {
		loopStr := "false"
		if anim.Loop {
			loopStr = "true"
		}
		fmt.Fprintf(&b, "\t\t%q: Anim(%d, %s,\n", anim.Name, int(anim.FPS), loopStr)
		for _, frame := range anim.Frames {
			b.WriteString("\t\t\tF(\n")
			for _, entry := range frame.Sprites {
				nb := len(entry.Hurtboxes)
				hbStart := ""
				if nb > 0 {
					hbStart = ", "
				}
				hbParts := make([]string, 0, nb)
				for _, hb := range entry.Hurtboxes {
					if hb.Rotation != 0 {
						hbParts = append(hbParts, fmt.Sprintf("HBR(%g, %g, %g, %g, %g)", hb.Width, hb.Height, hb.X, hb.Y, hb.Rotation))
					} else {
						hbParts = append(hbParts, fmt.Sprintf("HB(%g, %g, %g, %g)", hb.Width, hb.Height, hb.X, hb.Y))
					}
				}
				hbStr := strings.Join(hbParts, ", ")
				fmt.Fprintf(&b, "\t\t\t\tS(%d, %d, %g, %g, %g, %g, %g, %g, %g%s%s),\n",
					entry.SpriteIdx, entry.SpriteFrameIdx,
					entry.OffsetX, entry.OffsetY, entry.Rotation,
					entry.ScaleX, entry.ScaleY,
					entry.OriginX, entry.OriginY,
					hbStart, hbStr)
			}
			b.WriteString("\t\t\t),\n")
		}
		b.WriteString("\t\t),\n")
	}

	b.WriteString("\t},\n}\n")

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func ImportMovement(path string) (*project.ProjectData, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	var assetName string
	proj := &project.ProjectData{
		Animations: []project.AnimationRow{},
		Sprites:    []project.SpriteRow{},
	}

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
			if !ok {
				continue
			}
			assetName = vs.Names[0].Name

			for _, elt := range cl.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key := exprString(kv.Key)
				switch key {
				case "AssetKey":
					proj.AssetKey = stringLit(kv.Value)
				case "DefaultOriginX":
					proj.DefaultOriginX = floatLit(kv.Value)
				case "DefaultOriginY":
					proj.DefaultOriginY = floatLit(kv.Value)
				case "Animations":
					proj.Animations = parseAnimationsMov(kv.Value)
				}
			}
		}
	}

	if assetName == "" {
		return nil, fmt.Errorf("no Movement variable declaration found")
	}
	proj.AssetName = assetName

	proj.Sprites = buildSpriteList(proj.Animations)
	if len(proj.Sprites) == 0 {
		proj.Sprites = []project.SpriteRow{
			{Name: "Default Sprite", Width: 256, Height: 256, FrameCount: 1, CurrentIdx: 0, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
		}
	}

	return proj, nil
}

func buildSpriteList(anims []project.AnimationRow) []project.SpriteRow {
	maxIdx := -1
	for _, anim := range anims {
		for _, frame := range anim.Frames {
			for _, entry := range frame.Sprites {
				if entry.SpriteIdx > maxIdx {
					maxIdx = entry.SpriteIdx
				}
			}
		}
	}
	if maxIdx < 0 {
		return nil
	}
	sprites := make([]project.SpriteRow, maxIdx+1)
	for i := range sprites {
		sprites[i] = project.SpriteRow{
			Name:       fmt.Sprintf("Sprite %d", i),
			Width:      256,
			Height:     256,
			FrameCount: 1,
			CurrentIdx: 0,
			ScaleX:     1,
			ScaleY:     1,
			OriginX:    0.5,
			OriginY:    0.5,
		}
	}
	for _, anim := range anims {
		for _, frame := range anim.Frames {
			for _, entry := range frame.Sprites {
				if entry.SpriteIdx >= 0 && entry.SpriteIdx < len(sprites) {
					sprites[entry.SpriteIdx].OriginX = entry.OriginX
					sprites[entry.SpriteIdx].OriginY = entry.OriginY
					sprites[entry.SpriteIdx].ScaleX = entry.ScaleX
					sprites[entry.SpriteIdx].ScaleY = entry.ScaleY
				}
			}
		}
	}
	return sprites
}

func parseAnimationsMov(expr ast.Expr) []project.AnimationRow {
	cl, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var anims []project.AnimationRow
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		name := stringLit(kv.Key)
		call, ok := kv.Value.(*ast.CallExpr)
		if !ok {
			continue
		}
		anim := project.AnimationRow{Name: name}
		parseAnimCall(call, &anim)
		anims = append(anims, anim)
	}
	return anims
}

func parseAnimCall(call *ast.CallExpr, anim *project.AnimationRow) {
	if len(call.Args) < 2 {
		return
	}
	anim.FPS = float64(intLit(call.Args[0]))
	anim.Loop = boolIdent(call.Args[1])

	for i := 2; i < len(call.Args); i++ {
		fc, ok := call.Args[i].(*ast.CallExpr)
		if !ok {
			continue
		}
		fn := exprString(fc.Fun)
		if fn != "F" {
			continue
		}
		frame := project.AnimationFrame{}
		haveS := false
		for _, arg := range fc.Args {
			inner, ok := arg.(*ast.CallExpr)
			if !ok {
				continue
			}
			innerFn := exprString(inner.Fun)
			switch innerFn {
			case "S":
				entry := parseSpriteEntry(inner)
				haveS = true
				frame.Sprites = append(frame.Sprites, entry)
			case "HB", "HBR":
				if !haveS {
					entry := project.FrameSpriteEntry{
						SpriteIdx:      0,
						SpriteFrameIdx: 0,
						ScaleX:         1,
						ScaleY:         1,
						OriginX:        0.5,
						OriginY:        0.5,
					}
					if len(fc.Args) > 0 {
						if sf, ok := fc.Args[0].(*ast.BasicLit); ok && sf.Kind == token.INT {
							entry.SpriteFrameIdx = intLit(fc.Args[0])
						}
					}
					frame.Sprites = append(frame.Sprites, entry)
					haveS = true
				}
				hb := parseHurtboxCall(inner)
				if hb != nil && len(frame.Sprites) > 0 {
					last := &frame.Sprites[len(frame.Sprites)-1]
					last.Hurtboxes = append(last.Hurtboxes, *hb)
				}
			}
		}
		if !haveS {
			spriteFrame := 0
			if len(fc.Args) > 0 {
				if sf, ok := fc.Args[0].(*ast.BasicLit); ok && sf.Kind == token.INT {
					spriteFrame = intLit(fc.Args[0])
				}
			}
			frame.Sprites = append(frame.Sprites, project.FrameSpriteEntry{
				SpriteIdx:      0,
				SpriteFrameIdx: spriteFrame,
				ScaleX:         1,
				ScaleY:         1,
				OriginX:        0.5,
				OriginY:        0.5,
			})
		}
		anim.Frames = append(anim.Frames, frame)
	}
}

func parseSpriteEntry(call *ast.CallExpr) project.FrameSpriteEntry {
	entry := project.FrameSpriteEntry{
		ScaleX:  1,
		ScaleY:  1,
		OriginX: 0.5,
		OriginY: 0.5,
	}
	if len(call.Args) < 2 {
		return entry
	}
	entry.SpriteIdx = intLit(call.Args[0])
	entry.SpriteFrameIdx = intLit(call.Args[1])
	if len(call.Args) >= 3 {
		entry.OffsetX = floatLit(call.Args[2])
	}
	if len(call.Args) >= 4 {
		entry.OffsetY = floatLit(call.Args[3])
	}
	if len(call.Args) >= 5 {
		entry.Rotation = floatLit(call.Args[4])
	}
	if len(call.Args) >= 6 {
		entry.ScaleX = floatLit(call.Args[5])
	}
	if len(call.Args) >= 7 {
		entry.ScaleY = floatLit(call.Args[6])
	}
	if len(call.Args) >= 8 {
		entry.OriginX = floatLit(call.Args[7])
	}
	if len(call.Args) >= 9 {
		entry.OriginY = floatLit(call.Args[8])
	}
	for j := 9; j < len(call.Args); j++ {
		hb := parseHurtboxCall(call.Args[j])
		if hb != nil {
			entry.Hurtboxes = append(entry.Hurtboxes, *hb)
		}
	}
	return entry
}

func parseHurtboxCall(expr ast.Expr) *project.HurtboxRow {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil
	}
	fn := exprString(call.Fun)
	var hb project.HurtboxRow
	switch fn {
	case "HB":
		if len(call.Args) >= 4 {
			hb.Width = floatLit(call.Args[0])
			hb.Height = floatLit(call.Args[1])
			hb.X = floatLit(call.Args[2])
			hb.Y = floatLit(call.Args[3])
		}
	case "HBR":
		if len(call.Args) >= 5 {
			hb.Width = floatLit(call.Args[0])
			hb.Height = floatLit(call.Args[1])
			hb.X = floatLit(call.Args[2])
			hb.Y = floatLit(call.Args[3])
			hb.Rotation = floatLit(call.Args[4])
		}
	default:
		return nil
	}
	return &hb
}

func exprString(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.BasicLit:
		return v.Value
	case *ast.SelectorExpr:
		return exprString(v.X) + "." + v.Sel.Name
	default:
		return fmt.Sprintf("%T", e)
	}
}

func stringLit(e ast.Expr) string {
	bl, ok := e.(*ast.BasicLit)
	if !ok || bl.Kind != token.STRING {
		return ""
	}
	s, err := strconv.Unquote(bl.Value)
	if err != nil {
		return ""
	}
	return s
}

func floatLit(e ast.Expr) float64 {
	neg := false
	if ue, ok := e.(*ast.UnaryExpr); ok && ue.Op == token.SUB {
		neg = true
		e = ue.X
	}
	bl, ok := e.(*ast.BasicLit)
	if !ok || (bl.Kind != token.FLOAT && bl.Kind != token.INT) {
		return 0
	}
	v, _ := strconv.ParseFloat(bl.Value, 64)
	if neg {
		v = -v
	}
	return v
}

func intLit(e ast.Expr) int {
	neg := false
	if ue, ok := e.(*ast.UnaryExpr); ok && ue.Op == token.SUB {
		neg = true
		e = ue.X
	}
	bl, ok := e.(*ast.BasicLit)
	if !ok || bl.Kind != token.INT {
		return 0
	}
	v, _ := strconv.Atoi(bl.Value)
	if neg {
		v = -v
	}
	return v
}

func boolIdent(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	if !ok {
		return false
	}
	return id.Name == "true"
}
