package service

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/o1uch/go_final_project/internal/repeat"
	"github.com/o1uch/go_final_project/internal/store"
)

const (
	defaultTaskLimit = 30
)

var (
	ErrEmptyTitle      = errors.New("title is required")
	ErrDateParse       = errors.New("error parsing a string into a date")
	ErrEmptyDate       = errors.New("date is required")
	ErrInvalidNow      = errors.New("invalid now format")
	ErrNextDate        = errors.New("error calculating the nextDate function")
	ErrCreateTask      = errors.New("task creation error")
	ErrGettingTaskList = errors.New("error getting task list")
	ErrEmptyID         = errors.New("ID is required")
	ErrInvalidID       = errors.New("invalid ID format")
	ErrTaskNotFound    = errors.New("task not found")
	ErrGettingTask     = errors.New("error getting task")
)

type Service struct {
	repo store.TaskRepository
}

func NewService(repo store.TaskRepository) *Service {
	return &Service{repo: repo}
}

func trimExtraSpaces(t *store.Task) {
	t.Title = strings.TrimSpace(t.Title)
	t.Comment = strings.TrimSpace(t.Comment)
	t.Date = strings.TrimSpace(t.Date)
	t.Repeat = strings.TrimSpace(t.Repeat)
}

func (s *Service) NextDate(nowStr string, dstart string, repeatValue string) (string, error) {
	now := time.Now()

	if nowStr != "" {
		var err error
		now, err = time.Parse(repeat.DateLayout, nowStr)

		if err != nil {
			return "", errors.Join(ErrInvalidNow, err)
		}
	}

	if dstart == "" {
		return "", ErrEmptyDate
	}

	nextDate, err := repeat.NextDate(now, dstart, repeatValue)

	if err != nil {
		return "", err
	}

	return nextDate, nil

}

func (s *Service) AddTask(task *store.Task) (int64, error) {

	trimExtraSpaces(task)

	if task.Title == "" {
		return 0, ErrEmptyTitle
	}

	t := time.Now()
	now := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	nowStr := now.Format(repeat.DateLayout)

	if task.Date == "" {
		task.Date = nowStr
	}

	date, err := time.Parse(repeat.DateLayout, task.Date)

	if err != nil {
		return 0, errors.Join(ErrDateParse, err)
	}

	if date.Before(now) {
		if task.Repeat == "" {
			task.Date = nowStr
		} else {
			nextDate, err := repeat.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return 0, errors.Join(ErrNextDate, err)
			}
			task.Date = nextDate
		}
	}

	id, err := s.repo.Create(task)

	if err != nil {
		return 0, errors.Join(ErrCreateTask, err)
	}

	return id, nil
}

func (s *Service) GetTasks(searchPattern string) ([]*store.Task, error) {

	filter := store.TaskFilter{}

	if searchPattern != "" {
		date, err := time.Parse("02.01.2006", searchPattern)
		if err != nil {
			filter.Type = store.FilterByText
			filter.Value = searchPattern
		} else {
			filter.Type = store.FilterByDate
			strDate := date.Format(repeat.DateLayout)
			filter.Value = strDate
		}
	} else {
		filter.Type = store.FilterByLimit
		strLimit := strconv.Itoa(defaultTaskLimit)
		filter.Value = strLimit
	}

	tasks, err := s.repo.GetList(filter)
	if err != nil {
		return nil, errors.Join(ErrGettingTaskList, err)
	}

	if tasks == nil {
		tasks = []*store.Task{}
	}

	return tasks, nil
}

func (s *Service) GetTaskByID(idStr string) (*store.Task, error) {

	if idStr == "" {
		return nil, ErrEmptyID
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidID
	}

	task, err := s.repo.GetByID(id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}

		return nil, errors.Join(ErrGettingTask, err)
	}

	return task, nil
}
