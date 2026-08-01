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
	b.WriteString("\tSprites: []SpriteSheetDef{\n")
	for _, s := range proj.Sprites {
		fmt.Fprintf(&b, "\t\t{File: %q, FrameW: %d, FrameH: %d, FrameCount: %d},\n",
			s.File, s.Width, s.Height, s.FrameCount)
	}
	b.WriteString("\t},\n")
	b.WriteString("\tAnimations: map[string]AttackAnimDef{\n")

	for _, anim := range proj.Animations {
		loopStr := "false"
		if anim.Loop {
			loopStr = "true"
		}

		wuFrames, atkFrames, rcFrames := countPhaseFrames(anim.Frames)

		fmt.Fprintf(&b, "\t\t%q: AttackAnim(%g, %s, %g, %g, %g, %g, %g, %d, %d, %d,\n",
			anim.Name, anim.FPS, loopStr,
			anim.Windup, anim.Active, anim.Recover, anim.Armed, anim.ArmedFPS,
			wuFrames, atkFrames, rcFrames)

		for _, frame := range anim.Frames {
			phaseStr := phaseName(frame.Phase)
			fmt.Fprintf(&b, "\t\t\tAttackF(%s,\n", phaseStr)
			for _, s := range frame.Sprites {
				fmt.Fprintf(&b, "\t\t\t\tS(%d, %d, %g, %g, %g, %g, %g, %g, %g),\n",
					s.SpriteIdx, s.SpriteFrameIdx,
					s.OffsetX, s.OffsetY, s.Rotation, s.ScaleX, s.ScaleY,
					s.OriginX, s.OriginY)
			}
			b.WriteString("\t\t\t),\n")
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
				case "Sprites":
					proj.Sprites = parseSpriteSheets(kv.Value)
				case "Animations":
					anims, err := parseAttackAnimations(kv.Value)
					if err != nil {
						return nil, err
					}
					proj.Animations = anims
				}
			}
		}
	}

	if assetName == "" {
		return nil, fmt.Errorf("no Attack variable declaration found")
	}
	proj.AssetName = assetName

	if len(proj.Sprites) > 0 {
		applySpriteEntryProps(proj.Sprites, proj.Animations)
	} else {
		proj.Sprites = buildSpriteList(proj.Animations)
	}
	if len(proj.Sprites) == 0 {
		proj.Sprites = []project.SpriteRow{
			{Name: "Default Sprite", Width: 256, Height: 256, FrameCount: 1, CurrentIdx: 0, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
		}
	}

	return proj, nil
}

func parseAttackAnimations(expr ast.Expr) ([]project.AnimationRow, error) {
	cl, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("Animations: expected map literal")
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
			return nil, fmt.Errorf("animation %q: expected AttackAnim(...) call", name)
		}
		anim := project.AnimationRow{Name: name}
		if err := parseAttackAnimCall(call, &anim); err != nil {
			return nil, fmt.Errorf("animation %q: %w", name, err)
		}
		anims = append(anims, anim)
	}
	return anims, nil
}

func parseAttackAnimCall(call *ast.CallExpr, anim *project.AnimationRow) error {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "AttackAnim" {
		return fmt.Errorf("expected AttackAnim(...) call")
	}
	args := call.Args
	if len(args) < 10 {
		return fmt.Errorf("AttackAnim requires fps, loop, windup, active, recover, armed, armedFPS, windupFrames, activeFrames, recoverFrames and frames...")
	}

	anim.FPS = floatLit(args[0])
	anim.Loop = boolIdent(args[1])
	anim.Windup = floatLit(args[2])
	anim.Active = floatLit(args[3])
	anim.Recover = floatLit(args[4])
	anim.Armed = floatLit(args[5])
	anim.ArmedFPS = floatLit(args[6])

	for i := 10; i < len(args); i++ {
		fc, ok := args[i].(*ast.CallExpr)
		if !ok {
			return fmt.Errorf("unsupported frame: only AttackF(...) is allowed")
		}
		if ident, ok := fc.Fun.(*ast.Ident); !ok || ident.Name != "AttackF" {
			return fmt.Errorf("unsupported frame: only AttackF(...) is allowed, got %s", exprString(fc.Fun))
		}
		frame, err := parseAttackFFrame(fc)
		if err != nil {
			return err
		}
		anim.Frames = append(anim.Frames, frame)
	}
	return nil
}

func parseAttackFFrame(call *ast.CallExpr) (project.AnimationFrame, error) {
	var frame project.AnimationFrame
	if len(call.Args) < 1 {
		return frame, fmt.Errorf("AttackF requires a phase constant and at least one S(...) sprite")
	}
	phaseIdent, ok := call.Args[0].(*ast.Ident)
	if !ok {
		return frame, fmt.Errorf("AttackF first argument must be a phase constant")
	}
	frame.Phase = phaseFromIdent(phaseIdent.Name)
	for i := 1; i < len(call.Args); i++ {
		inner, ok := call.Args[i].(*ast.CallExpr)
		if !ok || exprString(inner.Fun) != "S" {
			return frame, fmt.Errorf("AttackF only accepts S(...) sprites")
		}
		frame.Sprites = append(frame.Sprites, parseSpriteEntry(inner))
	}
	if len(frame.Sprites) == 0 {
		return frame, fmt.Errorf("AttackF requires at least one S(...) sprite")
	}
	return frame, nil
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


