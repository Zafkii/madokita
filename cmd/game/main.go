package main

import (
	"errors"
	"image"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"madokita-gems/internal/assets"
	"madokita-gems/internal/audio"
	"madokita-gems/internal/combat"
	"madokita-gems/internal/database"
	"madokita-gems/internal/ecs"
	"madokita-gems/internal/engine"
	"madokita-gems/internal/event"
	"madokita-gems/internal/input"
	"madokita-gems/internal/platform"
	"madokita-gems/internal/save"
	"madokita-gems/internal/scene"
	"madokita-gems/internal/settings"
	"madokita-gems/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

var ErrWindowClose = errors.New("window close requested")

const (
	screenWidth  = 1280
	screenHeight = 720
	minWindowW   = 854
)

type GameApp struct {
	runtime   *engine.Runtime
	ecsWorld  *ecs.World
	eventBus  *event.Bus
	inputMgr  *input.Manager
	sceneMgr  *scene.Manager
	platform  platform.API
	collision *combat.CollisionSystem
	target    *combat.TargetSystem
	cache     *ui.ImageCache
	assetMgr  *assets.AssetManager
	audioMgr  *audio.AudioManager

	// OS window frame (title bar + borders) deltas, measured once in Layout
	// so the client area can be sized exactly to the game resolution.
	frameW, frameH  int
	frameCalibrated bool

	// Client size of the drawable area, measured in Layout. The single source
	// of truth for the windowed snap (see snapWindowAspect in window.go).
	lastClientW, lastClientH int

	// Timestamp of the previous Update, for real dt in the game loop.
	lastTick time.Time

	pendingWindow bool
	pendingData   settings.Data
}

func NewGameApp() *GameApp {
	cache := ui.NewImageCache("assets")
	assetMgr := assets.NewAssetManager(cache)

	if err := database.Init(); err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	settings.Initialize(settings.NewSQLiteRepo())
	settings.TryMigrateFromFile()

	bufferMs := settings.GetAudioBufferMs()
	if isVM() && bufferMs == 20 {
		bufferMs = 100
	}
	log.Printf("audio: buffer size %dms (VM=%v)", bufferMs, isVM())
	audioMgr := audio.NewAudioManager("assets", bufferMs)

	app := &GameApp{
		runtime:   engine.NewRuntime(),
		ecsWorld:  ecs.NewWorld(),
		eventBus:  event.NewBus(),
		inputMgr:  input.NewManager(),
		sceneMgr:  scene.NewManager(screenWidth, screenHeight),
		platform:  platform.NewDesktop(),
		collision: combat.NewCollisionSystem(),
		target:    combat.NewTargetSystem(),
		cache:     cache,
		assetMgr:  assetMgr,
		audioMgr:  audioMgr,
	}
	app.lastTick = time.Now()

	app.runtime.Container().Register("events", app.eventBus)
	app.runtime.Container().Register("input", app.inputMgr)
	app.runtime.Container().Register("platform", app.platform)

	d0 := settings.GetData()
	audioMgr.SetChannelVolume("general", d0.VolumeGeneral)
	audioMgr.SetChannelVolume("music", d0.VolumeMusic)
	audioMgr.SetChannelVolume("effects", d0.VolumeEffects)

	settings.SetOnApply(func(d settings.Data) {
		app.pendingData = d
		app.pendingWindow = true
		audioMgr.SetChannelVolume("general", d.VolumeGeneral)
		audioMgr.SetChannelVolume("music", d.VolumeMusic)
		audioMgr.SetChannelVolume("effects", d.VolumeEffects)
	})
	// Defer the OS close so the window position can be persisted in Update
	// when ebiten.IsWindowBeingClosed fires, before returning ErrWindowClose.
	ebiten.SetWindowClosingHandled(true)
	save.Initialize(save.NewSQLiteRepo())

	app.runtime.Systems().Add(&platformBootstrap{platform: app.platform})
	app.runtime.Systems().Add(&sceneBootstrap{
		sceneMgr: app.sceneMgr,
		cache:    cache,
		audioMgr: audioMgr,
		inputMgr: app.inputMgr,
		assetMgr: assetMgr,
		eventBus: app.eventBus,
	})

	return app
}

func isBrokenVmwgfx() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if _, err := os.Stat("/sys/module/vmwgfx"); os.IsNotExist(err) {
		return false
	}
	for _, path := range []string{
		"/sys/devices/virtual/dmi/id/product_name",
		"/sys/devices/virtual/dmi/id/sys_vendor",
	} {
		if d, err := os.ReadFile(path); err == nil && strings.Contains(strings.ToLower(string(d)), "virtualbox") {
			return true
		}
	}
	return false
}

func isVM() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if data, err := os.ReadFile("/sys/hypervisor/type"); err == nil {
		t := strings.TrimSpace(strings.ToLower(string(data)))
		if t == "kvm" || t == "xen" || t == "hyperv" || t == "acrn" {
			return true
		}
	}
	dmiChecks := []struct {
		path   string
		tokens []string
	}{
		{"/sys/devices/virtual/dmi/id/product_name", []string{"kvm", "qemu", "virtualbox", "vmware", "virtual machine"}},
		{"/sys/devices/virtual/dmi/id/sys_vendor", []string{"kvm", "qemu", "virtualbox", "vmware", "microsoft corporation", "xen"}},
	}
	for _, check := range dmiChecks {
		if data, err := os.ReadFile(check.path); err == nil {
			lower := strings.ToLower(string(data))
			for _, token := range check.tokens {
				if strings.Contains(lower, token) {
					return true
				}
			}
		}
	}
	return false
}

func main() {
	if isBrokenVmwgfx() {
		log.Println("vmwgfx: broken VirtualBox driver detected, forcing software rendering")
		os.Setenv("LIBGL_ALWAYS_SOFTWARE", "1")
	}

	app := NewGameApp()

	if err := app.runtime.Initialize(); err != nil {
		log.Fatalf("runtime init failed: %v", err)
	}
	defer app.runtime.Shutdown()
	defer database.Close()

	d := settings.GetData()
	app.setClientSize(d.Resolution.Width, d.Resolution.Height)
	ebiten.SetWindowSizeLimits(minWindowW, minWindowW*9/16, -1, -1)
	if d.WindowX != 0 || d.WindowY != 0 {
		ebiten.SetWindowPosition(d.WindowX, d.WindowY)
	}
	ebiten.SetMaxTPS(d.FPSLimit)
	if d.Fullscreen {
		if mon := ebiten.Monitor(); mon != nil {
			mw, mh := mon.Size()
			scale := mon.DeviceScaleFactor()
			ebiten.SetWindowSize(int(float64(mw)*scale), int(float64(mh)*scale))
		}
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
		ebiten.SetFullscreen(true)
	} else {
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	}
	ebiten.SetWindowTitle("Madokita Gems")

	if img, err := assets.LoadICO("assets/madokita-gems.ico", 0); err == nil && img != nil {
		ebiten.SetWindowIcon([]image.Image{img})
	}

	if err := ebiten.RunGame(app); err != nil && !errors.Is(err, ErrWindowClose) {
		log.Fatal(err)
	}
}
