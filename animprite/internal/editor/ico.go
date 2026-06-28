package editor

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"os"
)

type icoDirEntry struct {
	Width    uint8
	Height   uint8
	Colors   uint8
	Reserved uint8
	Planes   uint16
	BPP      uint16
	Size     uint32
	Offset   uint32
}

func LoadICO(path string, targetSize int) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hdr struct {
		_     uint16
		Type  uint16
		Count uint16
	}
	if err := binary.Read(f, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if hdr.Type != 1 || hdr.Count == 0 {
		return nil, nil
	}

	entries := make([]icoDirEntry, hdr.Count)
	if err := binary.Read(f, binary.LittleEndian, &entries); err != nil {
		return nil, err
	}

	best := pickICOEntry(entries, targetSize)
	if best == nil {
		return nil, nil
	}

	buf := make([]byte, best.Size)
	if _, err := f.ReadAt(buf, int64(best.Offset)); err != nil {
		return nil, err
	}

	if img, err := png.Decode(bytes.NewReader(buf)); err == nil {
		return img, nil
	}
	return nil, nil
}

func icoEntryWidth(e *icoDirEntry) int {
	if e.Width == 0 {
		return 256
	}
	return int(e.Width)
}

func pickICOEntry(entries []icoDirEntry, targetSize int) *icoDirEntry {
	if targetSize <= 0 {
		return &entries[len(entries)-1]
	}
	var best *icoDirEntry
	for i := range entries {
		e := &entries[i]
		w := icoEntryWidth(e)
		if best == nil {
			best = e
		}
		bw := icoEntryWidth(best)
		if w == targetSize {
			return e
		}
		if absInt(w-targetSize) < absInt(bw-targetSize) {
			best = e
		}
	}
	return best
}
