package platform

type API interface {
	IsDesktop() bool
	PlatformName() string
	CloseGame()
	SetResolution(w, h int) error
	ToggleFullscreen() error
}

type Desktop struct{}

func NewDesktop() *Desktop {
	return &Desktop{}
}

func (d *Desktop) IsDesktop() bool              { return true }
func (d *Desktop) PlatformName() string         { return "desktop" }
func (d *Desktop) CloseGame()                   {}
func (d *Desktop) SetResolution(w, h int) error { return nil }
func (d *Desktop) ToggleFullscreen() error      { return nil }
