package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/o1uch/go_final_project/internal/store"
	_ "modernc.org/sqlite"
)

type SchedulerStore struct {
	db *sql.DB
}

func NewSchedulerStore(db *sql.DB) SchedulerStore {
	return SchedulerStore{db: db}
}

func (s *SchedulerStore) Create(task *store.Task) (int64, error) {
	var id int64

	res, err := s.db.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES(:date,
	:title,:comment,:repeat)`,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))

	if err != nil {
		return 0, err
	}

	id, err = res.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil

}

func (s *SchedulerStore) GetList(f store.TaskFilter) ([]*store.Task, error) {
	tasks := make([]*store.Task, 0, 16)
	var rows *sql.Rows
	var err error

	switch f.Type {
	case FilterByDate:
		rows, err = s.db.Query(`
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		WHERE date = :date 
		ORDER BY date;`, sql.Named("date", f.Value))

		if err != nil {
			return nil, err
		}

	case FilterByText:
		fullPattern := "%" + f.Value + "%"
		rows, err = s.db.Query(`
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		WHERE title like :substr OR comment like :substr 
		ORDER BY date`, sql.Named("substr", fullPattern))

		if err != nil {
			return nil, err
		}

	case FilterByLimit:
		rows, err = s.db.Query(`
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		ORDER BY date 
		LIMIT :limit;`, sql.Named("limit", f.Value))

		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("no filter specified")
	}

	defer rows.Close()
	for rows.Next() {
		t := &Task{}
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

func (s *SchedulerStore) GetByID(id int64) (*Task, error) {
	t := Task{}
	row := s.db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id;`, sql.Named("id", id))

	if err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *SchedulerStore) Update(task *Task) error {
	res, err := s.db.Exec(`UPDATE scheduler 
	SET date = :date,
	title = :title,
	comment = :comment,
	repeat = :repeat 
	WHERE id = :id;`,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.ID))

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
	WHERE id = :id
	`, sql.Named("id", id))

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
