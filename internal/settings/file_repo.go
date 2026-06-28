package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FileRepo struct {
	path    string
	cached  Data
}

func NewFileRepo(path string) *FileRepo {
	return &FileRepo{path: path}
}

func (r *FileRepo) Load() (Data, error) {
	raw, err := os.ReadFile(r.path)
	if err != nil {
		return Data{}, err
	}
	var d Data
	if err := json.Unmarshal(raw, &d); err != nil {
		return Data{}, err
	}
	r.cached = d
	return d, nil
}

func (r *FileRepo) Save(d Data) error {
	r.cached = d
	if err := os.MkdirAll(filepath.Dir(r.path), 0755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, raw, 0644)
}
