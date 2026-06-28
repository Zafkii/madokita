package main

import (
	"context"
	"errors"
	"image"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"madokita/internal/assets"
	"madokita/internal/audio"
	"madokita/internal/combat"
	"madokita/internal/database"
	"madokita/internal/ecs"
	"madokita/internal/engine"
	"madokita/internal/event"
	"madokita/internal/input"
	"madokita/internal/platform"
	"madokita/internal/save"
	"madokita/internal/scene"
	"madokita/internal/settings"
	"madokita/internal/ui"
	"madokita/internal/windrag"

	"github.com/hajimehoshi/ebiten/v2"
)

var ErrWindowClose = errors.New("window close requested")

type titleBarBtn int

const (
	btnNone titleBarBtn = iota
	btnMinimize
	btnMaximize
	btnClose
)

const (
	screenWidth       = 1280
	screenHeight      = 720
	titleBarPhysH     = 30
	btnPhysW          = 30
	btnPhysPad        = 8
	minWindowW        = 854
	resizeEdgeThick   = 6
)

type resizeEdge int

const (
	edgeNone resizeEdge = iota
	edgeLeft
	edgeRight
	edgeTop
	edgeBottom
	edgeTopLeft
	edgeTopRight
	edgeBottomLeft
	edgeBottomRight
)

type resizeInfo struct {
	active  bool
	edge    resizeEdge
	startSX int
	startSY int
	initW   int
	initH   int
	initX   int
	initY   int
}

const (
	rsIdle = iota
	rsPending
	rsApply
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

	dragMgr      *windrag.DragManager
	prevLeftBtn  bool
	titleBarImg  *ebiten.Image
	titleBarW    int

	clickTimer   time.Time
	lastClickMX  int
	lastClickMY  int
	dragPending  bool
	gameWidth     int
	gameHeight    int
	outsideWidth  int
	outsideHeight int
	barLogicH     int
	btnLogicW     int
	resizing      resizeInfo

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	pendingWindow bool
	pendingData   settings.Data

	prevW, prevH    int
	restoreState    int
	restoreW, restoreH int

	hoveredBtn titleBarBtn
	titleLogo  *ebiten.Image

	prevTitleFontSize float64
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
		dragMgr:   &windrag.DragManager{},
		barLogicH: 1,
	}
	app.ctx, app.cancel = context.WithCancel(context.Background())

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
	ebiten.SetWindowDecorated(false)
	save.Initialize(save.NewSQLiteRepo())

	if logoImg, err := assets.LoadICO("assets/madokita.ico", 16); err == nil && logoImg != nil {
		app.titleLogo = ebiten.NewImageFromImage(logoImg)
	}

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

func (g *GameApp) runGameLoop(ctx context.Context) {
	g.wg.Add(1)
	defer g.wg.Done()

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if g.runtime.Phase() == engine.PhaseRunning {
				g.sceneMgr.Update(1.0 / 60.0)
			}
			g.inputMgr.Update()
		}
	}
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

	go app.runGameLoop(app.ctx)

	d := settings.GetData()
	ebiten.SetWindowSize(d.Resolution.Width, d.Resolution.Height)
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
		ebiten.SetWindowDecorated(false)
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
		ebiten.SetFullscreen(true)
	}
	ebiten.SetWindowTitle("Madokita")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if img, err := assets.LoadICO("assets/madokita.ico", 0); err == nil && img != nil {
		ebiten.SetWindowIcon([]image.Image{img})
	}

	if err := ebiten.RunGame(app); err != nil && !errors.Is(err, ErrWindowClose) {
		log.Fatal(err)
	}

	app.cancel()
	app.wg.Wait()
}
