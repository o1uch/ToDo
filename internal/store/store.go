package store

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskFilter struct {
	ByText bool
	ByDate bool
}

type SchedulerStore struct {
	db *sql.DB
}

func NewSchedulerStore(db *sql.DB) SchedulerStore {
	return SchedulerStore{db: db}
}

func (s *SchedulerStore) AddTask(task *Task) (int64, error) {
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

func (s *SchedulerStore) GetTasks(limit int) ([]*Task, error) {

	rows, err := s.db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT :limit;", sql.Named("limit", limit))

	if err != nil {
		return nil, err
	}

	tasks := make([]*Task, 0, limit)

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

func (s *SchedulerStore) FindTask(limit int, pattern string, f TaskFilter) ([]*Task, error) {
	tasks := make([]*Task, 0, 16)
	var rows *sql.Rows
	var err error

	switch {
	case f.ByDate:
		rows, err = s.db.Query(`
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		WHERE date = :date 
		ORDER BY date 
		LIMIT :limit;`, sql.Named("date", pattern), sql.Named("limit", limit))

		if err != nil {
			return nil, err
		}

	case f.ByText:
		fullPattern := "%" + pattern + "%"
		rows, err = s.db.Query(`
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		WHERE title like :substr OR comment like :substr 
		ORDER BY date 
		LIMIT :limit;`, sql.Named("substr", fullPattern), sql.Named("limit", limit))

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

func (s *SchedulerStore) GetTaskByID(id string) (*Task, error) {
	t := Task{}
	row := s.db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id;`, sql.Named("id", id))

	if err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *SchedulerStore) UpdateTaskByID(task *Task) error {
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

func (s *SchedulerStore) DeleteTaskByID(id string) error {

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
