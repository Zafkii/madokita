package editor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"animprite/internal/project"
)

func ExportAttack(path string, proj *project.ProjectData) error {
	var b strings.Builder

	b.WriteString("package attacks\n\n")
	b.WriteString("import . \"madokita/internal/animation\"\n\n")

	fmt.Fprintf(&b, "var %s = Attack{\n", proj.AssetName)
	fmt.Fprintf(&b, "\tAssetKey:       %q,\n", proj.AssetKey)
	fmt.Fprintf(&b, "\tDefaultOriginX: %g,\n", proj.DefaultOriginX)
	fmt.Fprintf(&b, "\tDefaultOriginY: %g,\n", proj.DefaultOriginY)
	b.WriteString("\tAnimations: map[string]AttackAnimDef{\n")

	for _, anim := range proj.Animations {
		loopStr := "false"
		if anim.Loop {
			loopStr = "true"
		}

		wuFrames, atkFrames, rcFrames := countPhaseFrames(anim.Frames)

		fmt.Fprintf(&b, "\t\t%q: AttackAnim(%g, %s, %g, %g, %g, %d, %d, %d,\n",
			anim.Name, anim.FPS, loopStr,
			anim.Windup, anim.Active, anim.Recover,
			wuFrames, atkFrames, rcFrames)

		for _, frame := range anim.Frames {
			phaseStr := phaseName(frame.Phase)

			if len(frame.Sprites) == 1 {
				s := frame.Sprites[0]
			if s.OffsetX == 0 && s.OffsetY == 0 && s.Rotation == 0 && s.ScaleX == 1 && s.ScaleY == 1 {
				fmt.Fprintf(&b, "\t\t\tAttackF(%d, %s),\n", s.SpriteFrameIdx, phaseStr)
			} else {
				b.WriteString("\t\t\t{\n")
				fmt.Fprintf(&b, "\t\t\t\tSpriteFrames: []int{%d},\n", s.SpriteFrameIdx)
				fmt.Fprintf(&b, "\t\t\t\tOffsetX:      []float64{%g},\n", s.OffsetX)
				fmt.Fprintf(&b, "\t\t\t\tOffsetY:      []float64{%g},\n", s.OffsetY)
				fmt.Fprintf(&b, "\t\t\t\tRotation:     []float64{%g},\n", s.Rotation)
				fmt.Fprintf(&b, "\t\t\t\tScaleX:       []float64{%g},\n", s.ScaleX)
				fmt.Fprintf(&b, "\t\t\t\tScaleY:       []float64{%g},\n", s.ScaleY)
				fmt.Fprintf(&b, "\t\t\t\tPhase:        phasePtr(%s),\n", phaseStr)
				b.WriteString("\t\t\t},\n")
			}
		} else {
			b.WriteString("\t\t\t{\n")
			b.WriteString("\t\t\t\tSpriteFrames: []int{")
			for i, s := range frame.Sprites {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%d", s.SpriteFrameIdx)
			}
			b.WriteString("},\n")
			b.WriteString("\t\t\t\tOffsetX: []float64{")
			for i, s := range frame.Sprites {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%g", s.OffsetX)
			}
			b.WriteString("},\n")
			b.WriteString("\t\t\t\tOffsetY: []float64{")
			for i, s := range frame.Sprites {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%g", s.OffsetY)
			}
			b.WriteString("},\n")
			b.WriteString("\t\t\t\tRotation: []float64{")
			for i, s := range frame.Sprites {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%g", s.Rotation)
			}
			b.WriteString("},\n")
			b.WriteString("\t\t\t\tScaleX: []float64{")
			for i, s := range frame.Sprites {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%g", s.ScaleX)
			}
			b.WriteString("},\n")
			b.WriteString("\t\t\t\tScaleY: []float64{")
			for i, s := range frame.Sprites {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%g", s.ScaleY)
			}
			b.WriteString("},\n")
			fmt.Fprintf(&b, "\t\t\t\tPhase:        phasePtr(%s),\n", phaseStr)
			b.WriteString("\t\t\t},\n")
		}
		}

		b.WriteString("\t\t),\n")
	}

	b.WriteString("\t},\n}\n")

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func countPhaseFrames(frames []project.AnimationFrame) (wu, atk, rc int) {
	for _, f := range frames {
		switch f.Phase {
		case project.PhaseWindup:
			wu++
		case project.PhaseActive:
			atk++
		case project.PhaseRecover:
			rc++
		}
	}
	return
}

func phaseName(phase project.FramePhase) string {
	switch phase {
	case project.PhaseWindup:
		return "PhaseWindup"
	case project.PhaseActive:
		return "PhaseActive"
	case project.PhaseRecover:
		return "PhaseRecover"
	case project.PhaseArmed:
		return "PhaseArmed"
	default:
		return "PhaseWindup"
	}
}

func ImportAttack(path string) (*project.ProjectData, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	var assetName string
	proj := &project.ProjectData{
		AssetName:      assetName,
		Animations:     []project.AnimationRow{},
		Sprites:        []project.SpriteRow{},
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
					proj.Animations = parseAttackAnimations(kv.Value)
				}
			}
		}
	}

	if assetName == "" {
		return nil, fmt.Errorf("no Attack variable declaration found")
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

func parseAttackAnimations(expr ast.Expr) []project.AnimationRow {
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
		parseAttackAnimCall(call, &anim)
		anims = append(anims, anim)
	}
	return anims
}

func parseAttackAnimCall(call *ast.CallExpr, anim *project.AnimationRow) {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "AttackAnim" {
		return
	}

	args := call.Args
	if len(args) < 8 {
		return
	}

	anim.FPS = floatLit(args[0])
	anim.Loop = boolIdent(args[1])
	anim.Windup = floatLit(args[2])
	anim.Active = floatLit(args[3])
	anim.Recover = floatLit(args[4])

	frameArgsStart := 8
	for i := frameArgsStart; i < len(args); i++ {
		frame := project.AnimationFrame{}
		innerCall, ok := args[i].(*ast.CallExpr)
		if ok {
			ident, ok := innerCall.Fun.(*ast.Ident)
			if ok && ident.Name == "AttackF" {
				frame = parseAttackFFrame(innerCall)
				anim.Frames = append(anim.Frames, frame)
				continue
			}
		}
		cl, ok := args[i].(*ast.CompositeLit)
		if ok {
			frame = parseAttackStructFrame(cl)
			anim.Frames = append(anim.Frames, frame)
		}
	}
}

func parseAttackFFrame(call *ast.CallExpr) project.AnimationFrame {
	frame := project.AnimationFrame{}
	if len(call.Args) < 2 {
		return frame
	}
	spriteFrame := intLit(call.Args[0])
	phaseIdent, ok := call.Args[1].(*ast.Ident)
	if !ok {
		return frame
	}
	frame.Phase = phaseFromIdent(phaseIdent.Name)
	frame.Sprites = append(frame.Sprites, project.FrameSpriteEntry{
		SpriteIdx:      0,
		SpriteFrameIdx: spriteFrame,
		ScaleX:         1,
		ScaleY:         1,
		OriginX:        0.5,
		OriginY:        0.5,
	})
	return frame
}

func parseAttackStructFrame(cl *ast.CompositeLit) project.AnimationFrame {
	frame := project.AnimationFrame{}
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := exprString(kv.Key)
		switch key {
		case "SpriteFrames":
			if cl2, ok := kv.Value.(*ast.CompositeLit); ok {
				for _, e := range cl2.Elts {
					frame.Sprites = append(frame.Sprites, project.FrameSpriteEntry{
						SpriteFrameIdx: intLit(e),
						ScaleX:         1,
						ScaleY:         1,
						OriginX:        0.5,
						OriginY:        0.5,
					})
				}
			}
		case "OffsetX":
			if cl2, ok := kv.Value.(*ast.CompositeLit); ok {
				for i, e := range cl2.Elts {
					if i < len(frame.Sprites) {
						frame.Sprites[i].OffsetX = floatLit(e)
					}
				}
			}
		case "OffsetY":
			if cl2, ok := kv.Value.(*ast.CompositeLit); ok {
				for i, e := range cl2.Elts {
					if i < len(frame.Sprites) {
						frame.Sprites[i].OffsetY = floatLit(e)
					}
				}
			}
		case "Rotation":
			if cl2, ok := kv.Value.(*ast.CompositeLit); ok {
				for i, e := range cl2.Elts {
					if i < len(frame.Sprites) {
						frame.Sprites[i].Rotation = floatLit(e)
					}
				}
			}
		case "ScaleX":
			if cl2, ok := kv.Value.(*ast.CompositeLit); ok {
				for i, e := range cl2.Elts {
					if i < len(frame.Sprites) {
						frame.Sprites[i].ScaleX = floatLit(e)
					}
				}
			}
		case "ScaleY":
			if cl2, ok := kv.Value.(*ast.CompositeLit); ok {
				for i, e := range cl2.Elts {
					if i < len(frame.Sprites) {
						frame.Sprites[i].ScaleY = floatLit(e)
					}
				}
			}
		case "Phase":
			if ce, ok := kv.Value.(*ast.CallExpr); ok {
				if len(ce.Args) > 0 {
					if ident, ok := ce.Args[0].(*ast.Ident); ok {
						frame.Phase = phaseFromIdent(ident.Name)
					}
				}
			}
		}
	}
	return frame
}

func phaseFromIdent(name string) project.FramePhase {
	short := strings.TrimPrefix(name, "Phase")
	switch short {
	case "Windup":
		return project.PhaseWindup
	case "Active":
		return project.PhaseActive
	case "Recover":
		return project.PhaseRecover
	case "Armed":
		return project.PhaseArmed
	default:
		return project.PhaseWindup
	}
}


