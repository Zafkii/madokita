package settings

type Repository interface {
	Load() (Data, error)
	Save(settings Data) error
}

type Data struct {
	Resolution     Resolution      `json:"resolution"`
	Fullscreen     bool            `json:"fullscreen"`
	Language       string          `json:"language"`
	FPSLimit       int             `json:"fpsLimit"`
	KeyBindings    map[string]int  `json:"keyBindings"`
	VolumeGeneral  float64         `json:"volumeGeneral"`
	VolumeMusic    float64         `json:"volumeMusic"`
	VolumeEffects  float64         `json:"volumeEffects"`
	WindowX        int             `json:"windowX"`
	WindowY        int             `json:"windowY"`
	AudioBufferMs  int             `json:"audioBufferMs"`
}

type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func DefaultData() Data {
	return Data{
		Resolution:     Resolution{Width: 1280, Height: 720},
		Fullscreen:     false,
		Language:       "en",
		FPSLimit:       60,
		KeyBindings:    make(map[string]int),
		VolumeGeneral:  0.7,
		VolumeMusic:    0.7,
		VolumeEffects:  0.7,
		AudioBufferMs:  20,
	}
}

var (
	global  Data
	repo    Repository
	onApply func(Data)
)

func Initialize(r Repository) error {
	repo = r
	data, err := repo.Load()
	if err != nil {
		data = DefaultData()
	}
	if data.KeyBindings == nil {
		data.KeyBindings = make(map[string]int)
	}
	if data.AudioBufferMs <= 0 {
		data.AudioBufferMs = DefaultData().AudioBufferMs
	}
	global = data
	return nil
}

func GetData() Data { return global }

func GetResolution() (int, int) { return global.Resolution.Width, global.Resolution.Height }
func GetLanguage() string       { return global.Language }
func IsFullscreen() bool        { return global.Fullscreen }
func GetFPSLimit() int          { return global.FPSLimit }

func GetAudioBufferMs() int {
	if global.AudioBufferMs <= 0 {
		return DefaultData().AudioBufferMs
	}
	return global.AudioBufferMs
}

func SetAudioBufferMs(ms int) {
	if ms < 10 {
		ms = 10
	}
	if ms > 500 {
		ms = 500
	}
	global.AudioBufferMs = ms
	save()
}

func SetResolution(w, h int) {
	global.Resolution = Resolution{Width: w, Height: h}
	save()
	apply()
}

func SetFullscreen(v bool) {
	global.Fullscreen = v
	save()
	apply()
}

func SetWindowPosition(x, y int) {
	global.WindowX = x
	global.WindowY = y
	save()
}

func SetLanguage(lang string) {
	global.Language = lang
	save()
}

func SetFPSLimit(fps int) {
	global.FPSLimit = fps
	save()
	apply()
}

func SetKeyBinding(actionName string, key int) {
	global.KeyBindings[actionName] = key
	save()
}

func GetVolumeGeneral() float64  { return global.VolumeGeneral }
func GetVolumeMusic() float64    { return global.VolumeMusic }
func GetVolumeEffects() float64  { return global.VolumeEffects }

func SetVolumeGeneral(v float64) {
	global.VolumeGeneral = clampVol(v)
	save()
	apply()
}

func SetVolumeMusic(v float64) {
	global.VolumeMusic = clampVol(v)
	save()
	apply()
}

func SetVolumeEffects(v float64) {
	global.VolumeEffects = clampVol(v)
	save()
	apply()
}

func clampVol(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func GetKeyBinding(actionName string) (int, bool) {
	v, ok := global.KeyBindings[actionName]
	return v, ok
}

func ResetDisplay() {
	d := DefaultData()
	global.Resolution = d.Resolution
	global.Fullscreen = d.Fullscreen
	global.FPSLimit = d.FPSLimit
	save()
	apply()
}

func ResetControllers() {
	global.KeyBindings = make(map[string]int)
	save()
}

func ResetVolume() {
	d := DefaultData()
	global.VolumeGeneral = d.VolumeGeneral
	global.VolumeMusic = d.VolumeMusic
	global.VolumeEffects = d.VolumeEffects
	save()
	apply()
}

func ResetSystem() {
	d := DefaultData()
	global.Language = d.Language
	save()
}

func SetOnApply(fn func(Data)) {
	onApply = fn
}

func save() {
	if repo != nil {
		repo.Save(global)
	}
}

func apply() {
	if onApply != nil {
		onApply(global)
	}
}
