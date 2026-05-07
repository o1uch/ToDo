package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/o1uch/go_final_project/internal/repeat"
	"github.com/o1uch/go_final_project/internal/service"
	"github.com/o1uch/go_final_project/internal/store"
)

type API struct {
	service   *service.Service
	authToken string
}

type TasksResp struct {
	Tasks []*store.Task `json:"tasks"`
}

func Init(svc *service.Service) {
	api := &API{service: svc}
	api.initAuthToken()
	http.HandleFunc("/api/signin", api.SignInHandler)

	http.HandleFunc("/api/nextdate", api.NextDayHandler)
	http.HandleFunc("/api/task", api.auth(api.MainTaskHandler))
	http.HandleFunc("/api/tasks", api.auth(api.GetTasksHandler))
	http.HandleFunc("/api/task/done", api.auth(api.DoneTaskHandler))
}

func (api *API) MainTaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		api.addTaskHandler(w, r)
	case http.MethodGet:
		api.GetTaskByIDHandler(w, r)
	case http.MethodPut:
		api.UpdateTaskHandler(w, r)
	case http.MethodDelete:
		api.DeleteTaskByID(w, r)
	default:
		writeJSON(w, map[string]string{
			"error": "The post method should be used to create a task",
		},
			http.StatusMethodNotAllowed)
		return
	}
}

func (api *API) NextDayHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		err := fmt.Sprintf("method %s, is not allowed", r.Method)
		writeJSON(w, map[string]string{
			"error": err,
		},
			http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.URL.Query().Get("now")
	dstart := r.URL.Query().Get("date")
	repeatValue := r.URL.Query().Get("repeat")

	nextDate, err := api.service.NextDate(nowStr, dstart, repeatValue)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": "error in calculating the nextDate: " + err.Error(),
		},
			http.StatusBadRequest)
		return
	}

	w.Write([]byte(nextDate))

}

func (api *API) addTaskHandler(w http.ResponseWriter, r *http.Request) {

	newTask := store.Task{}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &newTask)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": "invalid JSON: " + err.Error(),
		}, http.StatusBadRequest)
		return
	}

	id, err := api.service.AddTask(&newTask)

	if err != nil {
		status := http.StatusInternalServerError

		switch {
		case errors.Is(err, service.ErrEmptyTitle):
			status = http.StatusBadRequest

		case errors.Is(err, service.ErrDateParse):
			status = http.StatusBadRequest

		case errors.Is(err, service.ErrNextDate):
			status = http.StatusBadRequest
		}

		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, status)
		return

	}

	writeJSON(w, map[string]int64{
		"id": id,
	}, http.StatusOK)

}

func (api *API) GetTasksHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		writeJSON(w, map[string]string{
			"error": "the method used is not supported",
		}, http.StatusMethodNotAllowed)
		return
	}

	searchPattern := r.URL.Query().Get("search")

	TaskList, err := api.service.GetTasks(searchPattern)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, TasksResp{Tasks: TaskList}, http.StatusOK)

}

func (api *API) GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")

	task, err := api.service.GetTaskByID(idStr)

	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			status = http.StatusNotFound

		case errors.Is(err, service.ErrInvalidID),
			errors.Is(err, service.ErrEmptyID):
			status = http.StatusBadRequest
		}

		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, status)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

func (api *API) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {

	changedTask := &store.Task{}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, changedTask)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": "invalid JSON: " + err.Error(),
		}, http.StatusBadRequest)
		return
	}

	err = api.service.ChangeTaskByID(changedTask)

	if err != nil {
		status := http.StatusBadRequest

		switch {
		case errors.Is(err, service.ErrUpdateTask):
			status = http.StatusInternalServerError
		}

		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, status)
		return
	}

	writeJSON(w, map[string]any{}, http.StatusOK)
}

func (api *API) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		err := fmt.Sprintf("method %s, is not allowed", r.Method)
		writeJSON(w, map[string]string{
			"error": err,
		}, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")

	if idStr == "" {
		writeJSON(w, map[string]string{
			"error": "Id is not set",
		}, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, map[string]string{
			"error": "invalid ID",
		}, http.StatusBadRequest)
		return
	}

	task, err := api.store.GetByID(id)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, http.StatusNotFound)
		return
	}

	if task.Repeat == "" {
		err := api.store.Delete(task.ID)

		if err != nil {
			writeJSON(w, map[string]string{
				"error": err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]any{}, http.StatusOK)
		return
	}

	t := time.Now()

	now := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	nextDate, err := repeat.NextDate(now, task.Date, task.Repeat)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, http.StatusBadRequest)
		return
	}

	task.Date = nextDate

	err = api.store.Update(task)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{}, http.StatusOK)

}

func (api *API) DeleteTaskByID(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		err := fmt.Sprintf("method %s, is not allowed", r.Method)
		writeJSON(w, map[string]string{
			"error": err,
		}, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")

	if idStr == "" {
		writeJSON(w, map[string]string{
			"error": "Id is not set",
		}, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, map[string]string{
			"error": "invalid ID",
		}, http.StatusBadRequest)
		return
	}

	err = api.store.Delete(id)

	if err != nil {
		writeJSON(w, map[string]string{
			"error": err.Error(),
		}, http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{}, http.StatusOK)
}
