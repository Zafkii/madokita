package settings

import (
	"database/sql"
	"encoding/json"
	"madokita/internal/database"
	"os"
)

type SQLiteRepo struct{}

func NewSQLiteRepo() *SQLiteRepo {
	return &SQLiteRepo{}
}

func (r *SQLiteRepo) db() *sql.DB {
	return database.DB()
}

func (r *SQLiteRepo) Load() (Data, error) {
	var d Data
	row := r.db().QueryRow("SELECT value FROM meta WHERE key = 'settings'")
	var raw string
	if err := row.Scan(&raw); err != nil {
		return Data{}, err
	}
	// Detect missing volume fields (pre-migration data) and fill defaults
	var rawMap map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &rawMap); err != nil {
		return Data{}, err
	}
	def := DefaultData()
	if _, ok := rawMap["volumeGeneral"]; !ok {
		rawMap["volumeGeneral"] = def.VolumeGeneral
	}
	if _, ok := rawMap["volumeMusic"]; !ok {
		rawMap["volumeMusic"] = def.VolumeMusic
	}
	if _, ok := rawMap["volumeEffects"]; !ok {
		rawMap["volumeEffects"] = def.VolumeEffects
	}
	fixed, _ := json.Marshal(rawMap)
	if err := json.Unmarshal(fixed, &d); err != nil {
		return Data{}, err
	}
	if d.KeyBindings == nil {
		d.KeyBindings = make(map[string]int)
	}
	return d, nil
}

func (r *SQLiteRepo) Save(d Data) error {
	raw, err := json.Marshal(d)
	if err != nil {
		return err
	}
	_, err = r.db().Exec("INSERT OR REPLACE INTO meta (key, value) VALUES ('settings', ?)", string(raw))
	return err
}

func HasSettings() bool {
	var count int
	err := database.DB().QueryRow("SELECT COUNT(*) FROM meta WHERE key = 'settings'").Scan(&count)
	return err == nil && count > 0
}

func MigrateFromFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var d Data
	if err := json.Unmarshal(raw, &d); err != nil {
		return err
	}
	if d.KeyBindings == nil {
		d.KeyBindings = make(map[string]int)
	}
	repo := NewSQLiteRepo()
	return repo.Save(d)
}
