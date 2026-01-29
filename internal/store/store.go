package store

// store это CRUD

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Scheduler struct {
	Id      int
	Date    string
	Title   string
	Comment string
	Repeat  string
}

type SchedulerStore struct {
	db *sql.DB
}

func NewSchedulerStore(db *sql.DB) SchedulerStore {
	return SchedulerStore{db: db}
}
