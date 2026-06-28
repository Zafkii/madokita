package audio

import (
	"bytes"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/jfreymuth/oggvorbis"
)

const sampleRate = 44100
const bytesPerFrame = 4

type trackedReader struct {
	src io.ReadSeeker
	pos int64
	mu  sync.Mutex
}

func (r *trackedReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, err := r.src.Read(p)
	r.pos += int64(n)
	return n, err
}

func (r *trackedReader) Seek(offset int64, whence int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	newPos, err := r.src.Seek(offset, whence)
	if err == nil {
		r.pos = newPos
	}
	return newPos, err
}

func (r *trackedReader) currentPos() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	bytesPerSecond := int64(sampleRate * bytesPerFrame)
	sec := float64(r.pos) / float64(bytesPerSecond)
	return time.Duration(sec * float64(time.Second))
}

type loopReader struct {
	data []byte
	pos  int64
	mu   sync.Mutex
}

func (r *loopReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	total := 0
	for total < len(p) {
		n := copy(p[total:], r.data[r.pos:])
		total += n
		r.pos += int64(n)
		if r.pos >= int64(len(r.data)) {
			r.pos = 0
		}
	}
	return total, nil
}

func (r *loopReader) Seek(offset int64, whence int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	dataLen := int64(len(r.data))
	switch whence {
	case io.SeekStart:
		r.pos = offset % dataLen
	case io.SeekCurrent:
		r.pos = (r.pos + offset) % dataLen
	case io.SeekEnd:
		r.pos = (dataLen + offset) % dataLen
	}
	if r.pos < 0 {
		r.pos += dataLen
	}
	return r.pos, nil
}

func (r *loopReader) currentPos() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	bytesPerSecond := int64(sampleRate * bytesPerFrame)
	sec := float64(r.pos) / float64(bytesPerSecond)
	return time.Duration(sec * float64(time.Second))
}

type audioEntry struct {
	data     []byte
	player   *oto.Player
	tr       *trackedReader
	lr       *loopReader
	looping  bool
	duration time.Duration
}

type AudioManager struct {
	ctx            *oto.Context
	entries        map[string]*audioEntry
	basePath       string
	channelVolumes map[string]float64
	entryChannels  map[string]string
}

func NewAudioManager(basePath string, bufferSizeMs int) *AudioManager {
	if env := os.Getenv("MADOKITA_AUDIO_BUFFER_MS"); env != "" {
		if ms, err := strconv.Atoi(env); err == nil && ms >= 10 && ms <= 500 {
			bufferSizeMs = ms
		}
	}
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   time.Duration(bufferSizeMs) * time.Millisecond,
	}
	ctx, ch, err := oto.NewContext(op)
	if err != nil {
		panic(err)
	}
	<-ch
	return &AudioManager{
		ctx:            ctx,
		entries:        make(map[string]*audioEntry),
		basePath:       basePath,
		channelVolumes: make(map[string]float64),
		entryChannels:  make(map[string]string),
	}
}

func (a *AudioManager) decodeOGG(fullPath string) ([]byte, error) {
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := oggvorbis.NewReader(f)
	if err != nil {
		return nil, err
	}

	var allF32 []float32
	buf := make([]float32, 8192)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			allF32 = append(allF32, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	data := make([]byte, len(allF32)*2)
	for i, s := range allF32 {
		s = max(-1, min(1, s))
		val := int16(s * 32767)
		data[i*2] = byte(val)
		data[i*2+1] = byte(val >> 8)
	}
	return data, nil
}

func (a *AudioManager) LoadOGG(name, relPath, channel string) error {
	full := filepath.Join(a.basePath, relPath)
	data, err := a.decodeOGG(full)
	if err != nil {
		return err
	}
	dur := time.Duration(float64(len(data)) / float64(sampleRate*bytesPerFrame) * float64(time.Second))
	a.entries[name] = &audioEntry{data: data, duration: dur}
	if channel != "" {
		a.entryChannels[name] = channel
	}
	return nil
}

func (a *AudioManager) LoadOGGLoop(name, relPath, channel string) error {
	if err := a.LoadOGG(name, relPath, channel); err != nil {
		return err
	}
	a.entries[name].looping = true
	return nil
}

func (a *AudioManager) Play(name string) {
	e, ok := a.entries[name]
	if !ok {
		return
	}
	if e.player != nil {
		e.player.Pause()
	}
	if e.looping {
		e.lr = &loopReader{data: e.data}
		e.player = a.ctx.NewPlayer(e.lr)
	} else {
		e.tr = &trackedReader{src: bytes.NewReader(e.data)}
		e.player = a.ctx.NewPlayer(e.tr)
	}
	if vol := a.channelVolumeFor(name); vol >= 0 {
		e.player.SetVolume(vol)
	}
	e.player.Play()
}

func (a *AudioManager) PlayLoop(name string) {
	e, ok := a.entries[name]
	if !ok {
		return
	}
	if e.player != nil {
		e.player.Pause()
	}
	e.lr = &loopReader{data: e.data}
	e.player = a.ctx.NewPlayer(e.lr)
	if vol := a.channelVolumeFor(name); vol >= 0 {
		e.player.SetVolume(vol)
	}
	e.player.Play()
}

func (a *AudioManager) channelVolumeFor(name string) float64 {
	ch, hasCh := a.entryChannels[name]
	if !hasCh {
		return -1
	}
	vol, ok := a.channelVolumes[ch]
	if !ok {
		vol = 1.0
	}
	return vol
}

func (a *AudioManager) Stop(name string) {
	if e, ok := a.entries[name]; ok && e.player != nil {
		e.player.Pause()
	}
}

func (a *AudioManager) StopAll() {
	for _, e := range a.entries {
		if e.player != nil {
			e.player.Pause()
		}
	}
}

func (a *AudioManager) SetVolume(name string, vol float64) {
	if e, ok := a.entries[name]; ok && e.player != nil {
		e.player.SetVolume(math.Max(0, math.Min(1, vol)))
	}
}

func (a *AudioManager) SetChannelVolume(channel string, vol float64) {
	vol = math.Max(0, math.Min(1, vol))
	a.channelVolumes[channel] = vol
	for name, ch := range a.entryChannels {
		if ch == channel {
			if e, ok := a.entries[name]; ok && e.player != nil {
				e.player.SetVolume(vol)
			}
		}
	}
}

func (a *AudioManager) GetChannelVolume(channel string) float64 {
	vol, ok := a.channelVolumes[channel]
	if !ok {
		return 1.0
	}
	return vol
}

func (a *AudioManager) IsPlaying(name string) bool {
	e, ok := a.entries[name]
	if !ok || e.player == nil {
		return false
	}
	return e.player.IsPlaying()
}

func (a *AudioManager) GetPosition(name string) (time.Duration, bool) {
	e, ok := a.entries[name]
	if !ok || e.player == nil {
		return 0, false
	}
	if e.looping && e.lr != nil {
		return e.lr.currentPos(), true
	}
	if e.tr != nil {
		return e.tr.currentPos(), true
	}
	return 0, false
}

func (a *AudioManager) GetDuration(name string) (time.Duration, bool) {
	e, ok := a.entries[name]
	if !ok {
		return 0, false
	}
	return e.duration, true
}

func (a *AudioManager) Update() {}
