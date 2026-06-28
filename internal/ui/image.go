package ui

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

type ImageCache struct {
	images   map[string]*ebiten.Image
	basePath string
}

func NewImageCache(basePath string) *ImageCache {
	return &ImageCache{
		images:   make(map[string]*ebiten.Image),
		basePath: basePath,
	}
}

func (c *ImageCache) Load(relPath string) (*ebiten.Image, error) {
	if img, ok := c.images[relPath]; ok {
		return img, nil
	}
	full := filepath.Join(c.basePath, relPath)
	f, err := os.Open(full)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	eimg := ebiten.NewImageFromImage(img)
	c.images[relPath] = eimg
	return eimg, nil
}

func (c *ImageCache) Get(relPath string) *ebiten.Image {
	return c.images[relPath]
}

func (c *ImageCache) Preload(paths ...string) error {
	for _, p := range paths {
		if _, err := c.Load(p); err != nil {
			return err
		}
	}
	return nil
}

func (c *ImageCache) Remove(path string) {
	delete(c.images, path)
}

func (c *ImageCache) Keys() []string {
	keys := make([]string, 0, len(c.images))
	for k := range c.images {
		keys = append(keys, k)
	}
	return keys
}

func (c *ImageCache) Clear() {
	c.images = make(map[string]*ebiten.Image)
}
