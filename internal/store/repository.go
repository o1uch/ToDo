package store

type TaskRepository interface {
	Create(*Task) (int64, error)         // old name = AddTask(task *Task) (int64, error)
	GetByID(int64) (*Task, error)        // old name = GetTaskByID(id string) (*Task, error)
	Update(*Task) error                  // Update(task *Task) error
	Delete(int64) error                  // Delete(id string) error
	GetList(TaskFilter) ([]*Task, error) // old name GetTasks(limit int) ([]*Task, error) + FindTask(limit int, pattern string, f TaskFilter) ([]*Task, error)
}
