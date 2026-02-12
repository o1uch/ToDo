package store

// store это CRUD

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type SchedulerStore struct {
	db *sql.DB
}

func NewSchedulerStore(db *sql.DB) SchedulerStore {
	return SchedulerStore{db: db}
}

func (s *SchedulerStore) AddTask(task *Task) (int64, error) {
	var id int64

	// нужно ли перед выполнением запроса проверять, что БД доступна и открыть?
	// по идее метод AddTask() должен выполняться для БД, которую уже проинициализировали, т.е. открыли соединение и выполнили для неё NewSchedulerStore

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

func (s *SchedulerStore) GetTasks(limit int) ([]Task, error) {

	rows, err := s.db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT :limit;", sql.Named("limit", limit))

	if err != nil {
		return nil, err
	}

	tasks := make([]Task, 0, 16)

	defer rows.Close()
	for rows.Next() {
		t := Task{}
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

	if err := rows.Err(); err != nil { // для чего нужен row.Err() и когда его правильней вызывать, во время row.Next() или после?
		return nil, err
	}

	return tasks, nil
}
