package store

type TaskRepository interface {
	Create(*Task) (int64, error)
	GetByID(int64) (*Task, error)
	Update(*Task) error
	Delete(int64) error
	GetList(TaskFilter) ([]*Task, error)
}
