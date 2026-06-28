package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func Path() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfg, "madokita")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "madokita.sav"), nil
}

func Init() error {
	p, err := Path()
	if err != nil {
		return err
	}
	conn, err := sql.Open("sqlite", p)
	if err != nil {
		return err
	}
	conn.SetMaxOpenConns(1)
	db = conn
	return migrate()
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func DB() *sql.DB {
	return db
}
