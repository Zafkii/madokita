package menu

import (
	"fmt"
	"madokita/internal/settings"
	"math"
)

type navTab int

const (
	navProfile navTab = iota
	navDisplay
	navControllers
	navVolume
	navSystem
	navCount
)

type rightContent int

const (
	rcNone rightContent = iota
	rcResolution
	rcFPS
	rcMode
	rcBindings
	rcProfile
	rcVolumeBar
)

type colFocus int

const (
	focusNav colFocus = iota
	focusMiddle
	focusRight
)

func (s *SettingsScene) middleItems() []string {
	switch s.selNav {
	case navDisplay:
		data := settings.GetData()
		mode := "Windowed"
		if data.Fullscreen {
			mode = "Fullscreen"
		}
		return []string{
			fmt.Sprintf("Resolution: %dx%d", data.Resolution.Width, data.Resolution.Height),
			fmt.Sprintf("FPS Limit: %d", data.FPSLimit),
			fmt.Sprintf("Mode: %s", mode),
		}
	case navControllers:
		return []string{"UI", "Game", "System"}
	case navVolume:
		data := settings.GetData()
		return []string{
			fmt.Sprintf("General: %d%%", int(math.Round(data.VolumeGeneral*100))),
			fmt.Sprintf("Music: %d%%", int(math.Round(data.VolumeMusic*100))),
			fmt.Sprintf("Effects: %d%%", int(math.Round(data.VolumeEffects*100))),
		}
	case navSystem:
		return nil
	default:
		return nil
	}
}

func (s *SettingsScene) setNav(n navTab) {
	s.selNav = n
	s.selRight = 0
	s.colFocus = focusNav
	s.capturing = false
	switch n {
	case navProfile:
		s.selMiddle = -1
		s.rightType = rcProfile
	case navDisplay:
		s.selMiddle = 0
		s.buildRight()
	case navControllers:
		s.selMiddle = 0
		s.buildRight()
	case navVolume:
		s.selMiddle = 0
		s.buildRight()
	case navSystem:
		s.selMiddle = -1
		s.rightType = rcNone
	default:
		s.selMiddle = -1
		s.rightType = rcNone
	}
}

func (s *SettingsScene) toggleMiddle(idx int) {
	items := s.middleItems()
	if idx < 0 || idx >= len(items) {
		s.selMiddle = -1
		s.rightType = rcNone
		return
	}
	s.selMiddle = idx
	s.capturing = false
	s.buildRight()
}

func (s *SettingsScene) buildRight() {
	s.selRight = 0
	switch s.selNav {
	case navProfile:
		s.rightType = rcProfile
	case navDisplay:
		switch s.selMiddle {
		case 0:
			s.rightType = rcResolution
			current := fmt.Sprintf("%dx%d", settings.GetData().Resolution.Width, settings.GetData().Resolution.Height)
			for i, r := range resPresets {
				if r == current {
					s.selRight = i
					break
				}
			}
		case 1:
			s.rightType = rcFPS
			cur := fmt.Sprintf("%d", settings.GetData().FPSLimit)
			for i, f := range fpsPresets {
				if f == cur {
					s.selRight = i
					break
				}
			}
		case 2:
			s.rightType = rcMode
			if settings.GetData().Fullscreen {
				s.selRight = 1
			}
		default:
			s.rightType = rcNone
		}
	case navControllers:
		if s.selMiddle == 0 || s.selMiddle == 1 || s.selMiddle == 2 {
			s.rightType = rcBindings
		} else {
			s.rightType = rcNone
		}
	case navVolume:
		s.rightType = rcVolumeBar
		data := settings.GetData()
		var vol float64
		switch s.selMiddle {
		case 0:
			vol = data.VolumeGeneral
		case 1:
			vol = data.VolumeMusic
		case 2:
			vol = data.VolumeEffects
		default:
			vol = 0.8
		}
		s.selRight = int(math.Round(vol / 0.05))
	case navSystem:
		s.rightType = rcNone
	default:
		s.rightType = rcNone
	}
}

func (s *SettingsScene) rightOptionLabels() []string {
	switch s.rightType {
	case rcResolution:
		return resPresets
	case rcFPS:
		return fpsPresets
	case rcMode:
		return []string{"Windowed", "Fullscreen"}
	case rcProfile:
		return []string{"Name", "Level", "Location", "Completion", "Fav Char"}
	}
	return nil
}

func (s *SettingsScene) applyVolumeFromSelRight() {
	vol := float64(s.selRight) * 0.05
	switch s.selMiddle {
	case 0:
		settings.SetVolumeGeneral(vol)
	case 1:
		settings.SetVolumeMusic(vol)
	case 2:
		settings.SetVolumeEffects(vol)
	}
}

func (s *SettingsScene) selectRightOption(optIdx int) {
	switch s.rightType {
	case rcResolution:
		presets := []struct{ w, h int }{{1920, 1080}, {1600, 900}, {1280, 720}, {854, 480}}
		if optIdx >= 0 && optIdx < len(presets) {
			settings.SetResolution(presets[optIdx].w, presets[optIdx].h)
		}
	case rcFPS:
		vals := []int{30, 60, 120}
		if optIdx >= 0 && optIdx < len(vals) {
			settings.SetFPSLimit(vals[optIdx])
		}
	case rcMode:
		settings.SetFullscreen(optIdx == 1)
	}
}

func (s *SettingsScene) currentBindings() []bindingEntry {
	if s.selNav == navControllers && s.selMiddle == 2 {
		return systemBindings
	}
	if s.selMiddle == 0 {
		return uiBindings
	}
	return gameBindings
}
