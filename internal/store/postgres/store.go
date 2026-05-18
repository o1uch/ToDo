package postgres

import "database/sql"

type SchedulerStore struct {
	db *sql.DB
}

func NewSchedulerStore(db *sql.DB) SchedulerStore {
	return SchedulerStore{db: db}
}
