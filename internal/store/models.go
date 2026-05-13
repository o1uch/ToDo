package store

// main type

type Task struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// реализация структуры для фильтра на store.List

type FilterType int

const (
	FilterByText FilterType = iota
	FilterByDate
	FilterByLimit
)

type TaskFilter struct {
	Value string
	Type  FilterType
}
