package scene

import "github.com/hajimehoshi/ebiten/v2"

type Interface interface {
	Update(dt float64) error
	Draw(screen *ebiten.Image)
	Enter() error
	Exit() error
	Pause()
	Resume()
	IsActive() bool
}

type OverlayDrawer interface {
	DrawOverlay(screen *ebiten.Image)
}
