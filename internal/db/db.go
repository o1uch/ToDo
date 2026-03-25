package db

// пакет инициализации БД

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func createSchedulerSchema(db *sql.DB) error {

	createSchema := `CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL DEFAULT "",
	title VARCHAR(256) NOT NULL DEFAULT "",
	comment TEXT NOT NULL DEFAULT "",
	repeat VARCHAR(128) NOT NULL DEFAULT "");

CREATE INDEX IF NOT EXISTS  date_index ON scheduler (date);`

	_, err := db.Exec(createSchema)

	if err != nil {
		return err
	}

	return nil
}

func Init(dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbFile)

	if err != nil {
		return nil, fmt.Errorf("error opening the database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error checking the availability of the database: %w", err)
	}

	if err := createSchedulerSchema(db); err != nil {
		return nil, fmt.Errorf("error initializing the scheduler schema: %w", err)
	}

	return db, nil
}
