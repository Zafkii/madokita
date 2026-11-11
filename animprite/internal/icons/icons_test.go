package icons

import (
	"testing"
	"testing/fstest"

	assetIcons "animprite/assets/icons"
)

func TestRasterizeTurnsSVGIntoWhiteMask(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
		<circle cx="12" cy="12" r="10" fill="#000000"/>
	</svg>`)

	mask, err := Rasterize(svg, 32)
	if err != nil {
		t.Fatalf("Rasterize() error = %v", err)
	}
	if got := mask.Bounds().Dx(); got != 32 {
		t.Fatalf("mask width = %d, want 32", got)
	}

	r, g, b, a := mask.At(16, 16).RGBA()
	if a == 0 {
		t.Fatal("center pixel is transparent")
	}
	if r != 0xffff || g != 0xffff || b != 0xffff {
		t.Fatalf("center pixel = (%d,%d,%d), want white", r, g, b)
	}
}

func TestLoadSkipsInvalidSVGsAndKeepsValidOnes(t *testing.T) {
	fsys := fstest.MapFS{
		"valid.svg":   &fstest.MapFile{Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24"/></svg>`)},
		"broken.svg":  &fstest.MapFile{Data: []byte(`not an svg`)},
		"ignored.txt": &fstest.MapFile{Data: []byte(`ignore me`)},
	}

	set, err := Load(fsys, 16)
	if err == nil {
		t.Fatal("Load() error = nil, want invalid SVG error")
	}
	if set.masks["valid"] == nil {
		t.Fatal("valid icon was not loaded")
	}
	if set.masks["broken"] != nil {
		t.Fatal("invalid icon was loaded")
	}
	if set.masks["ignored"] != nil {
		t.Fatal("non-SVG file was loaded")
	}
}

func TestEmbeddedIconsLoad(t *testing.T) {
	set, err := Load(assetIcons.FS, 24)
	if err != nil {
		t.Fatalf("Load(embedded icons) error = %v", err)
	}
	for _, name := range []string{"hammer", "square-pen", "trash", "info"} {
		if set.masks[name] == nil {
			t.Errorf("embedded icon %q was not loaded", name)
		}
	}
}
