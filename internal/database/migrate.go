package database

import "database/sql"

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		sql: `
CREATE TABLE IF NOT EXISTS meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS profile (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS save_slots (
    slot_name    TEXT PRIMARY KEY,
    save_version INTEGER NOT NULL DEFAULT 1,
    save_data    TEXT NOT NULL,
    updated_at   TEXT NOT NULL DEFAULT (datetime('now'))
);
`,
	},
}

func migrate() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentVer int
	err = tx.QueryRow("SELECT COALESCE(MAX(version), 0) FROM meta WHERE key = 'schema_version'").Scan(&currentVer)
	if err != nil && err != sql.ErrNoRows {
		currentVer = 0
	}

	for _, m := range migrations {
		if m.version > currentVer {
			if _, err := tx.Exec(m.sql); err != nil {
				return err
			}
			if _, err := tx.Exec("INSERT OR REPLACE INTO meta (key, value) VALUES ('schema_version', ?)", m.version); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
