package save

import (
	"database/sql"
	"encoding/json"
	"madokita/internal/database"
)

type SQLiteRepo struct{}

func NewSQLiteRepo() *SQLiteRepo {
	return &SQLiteRepo{}
}

func (r *SQLiteRepo) db() *sql.DB {
	return database.DB()
}

func (r *SQLiteRepo) LoadProfile() (Profile, error) {
	var p Profile
	row := r.db().QueryRow("SELECT value FROM profile WHERE key = 'profile'")
	var raw string
	if err := row.Scan(&raw); err != nil {
		return Profile{}, err
	}
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return Profile{}, err
	}
	return p, nil
}

func (r *SQLiteRepo) SaveProfile(p Profile) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = r.db().Exec("INSERT OR REPLACE INTO profile (key, value) VALUES ('profile', ?)", string(raw))
	return err
}

func (r *SQLiteRepo) ListSlots() ([]SlotInfo, error) {
	rows, err := r.db().Query("SELECT slot_name, save_version, updated_at FROM save_slots ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []SlotInfo
	for rows.Next() {
		var s SlotInfo
		if err := rows.Scan(&s.Name, &s.Version, &s.UpdatedAt); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

func (r *SQLiteRepo) LoadSlot(slot string) (Data, error) {
	var d Data
	row := r.db().QueryRow("SELECT save_data FROM save_slots WHERE slot_name = ?", slot)
	var raw string
	if err := row.Scan(&raw); err != nil {
		return Data{}, err
	}
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return Data{}, err
	}
	if d.Player.Upgrades == nil {
		d.Player.Upgrades = make(map[string]int)
	}
	return d, nil
}

func (r *SQLiteRepo) SaveSlot(slot string, data Data) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = r.db().Exec(
		`INSERT INTO save_slots (slot_name, save_version, save_data, updated_at)
		 VALUES (?, 1, ?, datetime('now'))
		 ON CONFLICT(slot_name) DO UPDATE SET
		 save_version = excluded.save_version,
		 save_data = excluded.save_data,
		 updated_at = datetime('now')`,
		slot, string(raw),
	)
	return err
}

func (r *SQLiteRepo) DeleteSlot(slot string) error {
	_, err := r.db().Exec("DELETE FROM save_slots WHERE slot_name = ?", slot)
	return err
}
