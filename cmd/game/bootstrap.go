package main

import (
	"fmt"
	"os"

	"madokita/internal/assets"
	"madokita/internal/audio"
	"madokita/internal/combat"
	"madokita/internal/data/characters/movements"
	"madokita/internal/entity/player"
	"madokita/internal/event"
	"madokita/internal/game"
	"madokita/internal/input"
	"madokita/internal/localization"
	"madokita/internal/menu"
	"madokita/internal/platform"
	"madokita/internal/scene"
	"madokita/internal/settings"
	"madokita/internal/fonts"
	"madokita/internal/ui"
)

func newPlayerActor() *combat.Actor {
	stats := combat.NewStats()
	return combat.NewActor("player", combat.TeamPlayer, &stats, combat.NewHurtboxSystem(nil))
}

type platformBootstrap struct {
	platform platform.API
}

func (b *platformBootstrap) ID() string        { return "platform" }
func (b *platformBootstrap) Initialize() error { return nil }
func (b *platformBootstrap) Start() error      { return nil }
func (b *platformBootstrap) Ready() error      { return nil }
func (b *platformBootstrap) Shutdown() error   { return nil }

type sceneBootstrap struct {
	sceneMgr  *scene.Manager
	cache     *ui.ImageCache
	audioMgr  *audio.AudioManager
	inputMgr  *input.Manager
	assetMgr  *assets.AssetManager
	eventBus  *event.Bus
	gameScene *game.GameScene
}

func (b *sceneBootstrap) ID() string { return "scenes" }
func (b *sceneBootstrap) Initialize() error {
	b.sceneMgr.Add("preload", menu.NewPreloadScene(b.sceneMgr, b.cache, b.audioMgr))
	b.sceneMgr.Add("menu-intro", menu.NewMenuIntroScene(b.sceneMgr, b.cache, b.audioMgr))

	mainMenu := menu.NewMainMenuScene(b.sceneMgr, b.cache, b.inputMgr, b.audioMgr)
	b.sceneMgr.Add("main-menu", mainMenu)

	b.sceneMgr.Add("settings", menu.NewSettingsScene(b.sceneMgr, b.cache, b.inputMgr, b.audioMgr))

	loadingScene := menu.NewLoadingScene(b.sceneMgr, b.cache)
	b.sceneMgr.Add("loading", loadingScene)

	b.gameScene = game.NewGameScene(b.sceneMgr, b.assetMgr, game.TestStageDef, b.audioMgr)
	b.sceneMgr.Add("game-teststage", b.gameScene)

	mainMenu.SetOnNewGame(func() {
		loadingScene.SetTarget(b.setupGameScene, "game-teststage")
		b.sceneMgr.SwitchTo("loading")
	})

	return nil
}
func (b *sceneBootstrap) Start() error {
	if err := ui.SetFontFromTTF(fonts.PxPlusIBMVGA8x14TTF, 13); err != nil {
		return fmt.Errorf("loading UI font: %w", err)
	}
	if err := ui.SetTitleFontFromTTF(fonts.PxPlusIBMVGA8x14TTF); err != nil {
		return fmt.Errorf("loading title font: %w", err)
	}

	data := settings.GetData()
	b.inputMgr.LoadBindings(data.KeyBindings)

	sceneSetups := map[string]func() error{
		"game-teststage": b.setupGameScene,
		"menu-intro":     b.setupMenuAssets,
		"main-menu":      b.setupMenuAssets,
	}

	target := os.Getenv("SCENE")
	if target == "" {
		target = "menu-intro"
	}
	if setup, ok := sceneSetups[target]; ok {
		if err := setup(); err != nil {
			return err
		}
	}
	return b.sceneMgr.SwitchTo(target)
}

func (b *sceneBootstrap) setupGameScene() error {
	if err := b.assetMgr.PreloadStage(game.TestStageDef); err != nil {
		return err
	}

	def := &movements.SayakaMovement
	sheets := make([]assets.AssetEntry, 0, len(def.Sprites))
	for _, s := range def.Sprites {
		sheets = append(sheets, assets.AssetEntry{
			Key:        def.AssetKey,
			Path:       s.File,
			FrameW:     s.FrameW,
			FrameH:     s.FrameH,
			FrameCount: s.FrameCount,
		})
	}
	if err := b.assetMgr.LoadCharacter(&assets.CharacterSheets{Key: def.AssetKey, Sheets: sheets}); err != nil {
		return err
	}

	p := player.New(0, 0, newPlayerActor(), b.inputMgr, b.eventBus)
	p.SetupAnim(def, b.assetMgr)
	p.PlayAnim("idle")
	p.State.IsControlled = true

	b.gameScene.SetPlayer(p)

	return nil
}

func (b *sceneBootstrap) setupMenuAssets() error {
	images := []string{
		"menu/madokita-title.png",
		"menu/madokita-title-top.png",
		"menu/star-title-.png",
		"menu/star-title-top.png",
		"menu/prevmenu1.png",
		"menu/prevmenu2.png",
		"menu/cosmic-effect.png",
		"images/loading.png",
	}
	if err := b.cache.Preload(images...); err != nil {
		return err
	}

	localization.Initialize("en")

	if b.audioMgr != nil {
		if err := b.audioMgr.LoadOGGLoop("menu-theme", "sounds/music/menutheme.ogg", "music"); err == nil {
			b.audioMgr.PlayLoop("menu-theme")
		}
	}

	return nil
}

func (b *sceneBootstrap) Ready() error    { return nil }
func (b *sceneBootstrap) Shutdown() error { return nil }
