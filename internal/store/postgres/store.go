package postgres

import (
	"database/sql"

	"github.com/o1uch/go_final_project/internal/store"
)

type SchedulerStore struct {
	db *sql.DB
}

func NewSchedulerStore(db *sql.DB) SchedulerStore {
	return SchedulerStore{db: db}
}

func (s SchedulerStore) Create(task *store.Task) (int64, error) {
	var pk int64

	query := `INSERT INTO scheduler (date, title, comment, repeat) 
	VALUES ($1, $2, $3, $4) 
	RETURNING id`
	// RETURNING для того, чтобы вернуть идентификатор вызывающей программе
	// обратно в приложение Go.
	err := s.db.QueryRow(query, task.Date, task.Title, task.Comment, task.Repeat).Scan(&pk)

	if err != nil {
		return 0, err
	}

	return pk, nil
}
