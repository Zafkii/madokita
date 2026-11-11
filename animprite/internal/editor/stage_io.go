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
		fmt.Fprintf(&b, "\t\t{Name: %q, File: %q, Category: %q, Width: %d, Height: %d, FrameCount: %d},\n",
			asset.Name, relativeSpritePath(dir, asset.File), asset.Category,
			asset.Width, asset.Height, asset.FrameCount)
	}
	b.WriteString("\t},\n")
	b.WriteString("\tObjects: []Object{\n")
	for _, obj := range stage.Objects {
		fmt.Fprintf(&b, "\t\t{Layer: %q, Category: %q, File: %q, X: %g, Y: %g, Width: %g, Height: %g, Rotation: %g, Opacity: %g},\n",
			obj.Layer, obj.Category, relativeSpritePath(dir, obj.File),
			obj.X, obj.Y, obj.Width, obj.Height, obj.Rotation, obj.Opacity)
	}
	b.WriteString("\t},\n}\n")

	return os.WriteFile(path, []byte(b.String()), 0644)
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
				Width:   project.DefaultStageWidth,
				Height:  project.DefaultStageHeight,
			}
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
				case "Height":
					stage.Height = floatLit(kv.Value)
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
			}
		}
		if asset.FrameCount <= 0 {
			asset.FrameCount = 1
		}
		if asset.Category == "" {
			asset.Category = project.StageCategoryForFile(asset.File)
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
			}
		}
		if !haveOpacity {
			obj.Opacity = 1
		}
		if obj.Category == "" {
			obj.Category = project.StageCategoryForFile(obj.File)
		}
		objects = append(objects, obj)
	}
	return objects
}

// applyStageObjectDefaults fills object sizes from the matching asset when a
// hand-written config omits Width/Height.
func applyStageObjectDefaults(stage *project.StageData) {
	byFile := make(map[string]project.StageAsset, len(stage.Assets))
	for _, asset := range stage.Assets {
		byFile[filepath.Clean(asset.File)] = asset
	}
	for i := range stage.Objects {
		obj := &stage.Objects[i]
		asset, ok := byFile[filepath.Clean(obj.File)]
		if !ok {
			continue
		}
		if obj.Width <= 0 && asset.Width > 0 {
			obj.Width = float64(asset.Width)
		}
		if obj.Height <= 0 && asset.Height > 0 {
			obj.Height = float64(asset.Height)
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
