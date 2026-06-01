package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

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

func (s *SchedulerStore) GetList(f store.TaskFilter) ([]*store.Task, error) {

	tasks := make([]*store.Task, 0, 16)
	var (
		rows *sql.Rows
		err  error
	)

	switch f.Type {
	case store.FilterByDate:
		rows, err = s.db.Query(`
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		WHERE date = $1
		ORDER BY date;`, f.Value)

		if err != nil {
			return nil, err
		}

	case store.FilterByText:
		fullPattern := "%" + f.Value + "%"
		rows, err = s.db.Query(`
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		WHERE title ILIKE $1 OR comment ILIKE $1 
		ORDER BY date`, fullPattern)

		if err != nil {
			return nil, err
		}

	case store.FilterByLimit:

		intLimit, err := strconv.ParseInt(f.Value, 10, 64)

		if err != nil {
			return nil, err
		}

		rows, err = s.db.Query(`
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		ORDER BY date
		LIMIT $1;`, intLimit)

		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("no filter specified")

	}

	defer rows.Close()
	for rows.Next() {
		t := &store.Task{}
		if err := rows.Scan(&t.ID,
			&t.Date,
			&t.Title,
			&t.Comment,
			&t.Repeat,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *SchedulerStore) GetByID(id int64) (*store.Task, error) {
	t := store.Task{}
	row := s.db.QueryRow(`
	SELECT id, date, title, comment, repeat
	FROM scheduler 
	WHERE id = $1;`, id)

	if err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *SchedulerStore) Update(task *store.Task) error {
	res, err := s.db.Exec(`UPDATE scheduler 
	SET date = $1,
	title = $2,
	comment = $3,
	repeat = $4 
	WHERE id = $5;`,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
		task.ID,
	)

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil

}

func (s *SchedulerStore) Delete(id int64) error {

	res, err := s.db.Exec(`
	DELETE FROM scheduler
	WHERE id = $1
	`, id)

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("task with id = %v was not found", id)
	}

	return nil
}
