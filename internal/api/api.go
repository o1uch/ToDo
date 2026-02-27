package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/o1uch/go_final_project/internal/repeat"
	"github.com/o1uch/go_final_project/internal/store"
)

const (
	defaultTaskLimit = 30
)

type API struct {
	store     *store.SchedulerStore
	authToken string
}

type TasksResp struct {
	Tasks []*store.Task `json:"tasks"`
}

func Init(store *store.SchedulerStore) {
	api := &API{store: store}
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
		api.GetTaskByID(w, r)
	case http.MethodPut:
		api.ChangeTaskByID(w, r)
	case http.MethodDelete:
		api.DeleteTaskByID(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, map[string]string{
			"error": "The post method should be used to create a task",
		})
		return
	}
}

func (api *API) NextDayHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := fmt.Sprintf("method %s, is not allowed", r.Method)
		writeJSON(w, map[string]string{
			"error": err,
		})
		return
	}

	nowStr := r.URL.Query().Get("now")
	now := time.Now()

	if nowStr != "" {
		var err error
		now, err = time.Parse(repeat.DateLayout, nowStr)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]string{
				"error": err.Error(),
			})
			return
		}

	}

	dstart := r.URL.Query().Get("date")

	if dstart == "" {
		err := "the \"date\" parameter is not specified"
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": err,
		})
		return
	}

	repeatValue := r.URL.Query().Get("repeat")

	nextDate, err := repeat.NextDate(now, dstart, repeatValue)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "error in calculating the nextDate: " + err.Error(),
		})
		return
	}

	w.Write([]byte(nextDate))

}

func (api *API) addTaskHandler(w http.ResponseWriter, r *http.Request) {

	newTask := store.Task{}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	err = json.Unmarshal(body, &newTask)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "invalid JSON: " + err.Error(),
		})
		return
	}

	trimExtraSpaces(&newTask)

	if newTask.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "the \"title\" field is empty",
		})
		return
	}

	t := time.Now()
	now := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	nowStr := now.Format("20060102")

	if newTask.Date == "" {
		newTask.Date = nowStr
	}

	date, err := time.Parse("20060102", newTask.Date)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if date.Before(now) {

		if newTask.Repeat == "" {
			newTask.Date = nowStr
		} else {
			nextDate, err := repeat.NextDate(now, newTask.Date, newTask.Repeat)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(w, map[string]string{
					"error": err.Error(),
				})
				return
			}
			newTask.Date = nextDate
		}

	}

	lastID, err := api.store.AddTask(&newTask)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, map[string]int64{
		"id": lastID,
	})

}

func (api *API) GetTasksHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, map[string]string{
			"error": "the method used is not supported",
		})
		return
	}

	searchPattern := r.URL.Query().Get("search")
	if searchPattern != "" {
		filter := store.TaskFilter{}
		date, err := time.Parse("02.01.2006", searchPattern)

		if err != nil {
			filter.ByText = true
			Resp, err := api.store.FindTask(defaultTaskLimit, searchPattern, filter)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeJSON(w, map[string]string{
					"error": err.Error(),
				})
				return
			}

			writeJSON(w, TasksResp{Tasks: Resp})
			return
		}

		filter.ByDate = true
		strDate := date.Format(repeat.DateLayout)
		Resp, err := api.store.FindTask(defaultTaskLimit, strDate, filter)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, TasksResp{Tasks: Resp})
		return

	}

	Resp, err := api.store.GetTasks(defaultTaskLimit)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if Resp == nil {
		Resp = []*store.Task{}
	}

	writeJSON(w, TasksResp{Tasks: Resp})

}

func (api *API) GetTaskByID(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "the \"id\" field is empty",
		})
		return
	}

	task, err := api.store.GetTaskByID(id)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]string{
				"error": "task not found",
			})
			return

		} else {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{
				"error": err.Error(),
			})
			return
		}
	}
	writeJSON(w, task)
}

func (api *API) ChangeTaskByID(w http.ResponseWriter, r *http.Request) {

	changeTask := store.Task{}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	err = json.Unmarshal(body, &changeTask)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "invalid JSON: " + err.Error(),
		})
		return
	}

	trimExtraSpaces(&changeTask)

	if changeTask.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "the \"id\" field is empty",
		})
		return
	}

	if changeTask.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "the \"title\" field is empty",
		})
		return
	}

	t := time.Now()
	now := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	nowStr := now.Format("20060102")

	if changeTask.Date == "" {
		changeTask.Date = nowStr
	}

	date, err := time.Parse("20060102", changeTask.Date)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if date.Before(now) {

		if changeTask.Repeat == "" {
			changeTask.Date = nowStr
		} else {
			nextDate, err := repeat.NextDate(now, changeTask.Date, changeTask.Repeat)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(w, map[string]string{
					"error": err.Error(),
				})
				return
			}
			changeTask.Date = nextDate
		}

	}

	err = api.store.Update(&changeTask)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, map[string]any{})

}

func (api *API) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := fmt.Sprintf("method %s, is not allowed", r.Method)
		writeJSON(w, map[string]string{
			"error": err,
		})
		return
	}

	id := r.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "Id is not set",
		})
		return
	}

	task, err := api.store.GetTaskByID(id)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if task.Repeat == "" {
		err := api.store.Delete(task.ID)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, map[string]any{})
		return
	}

	t := time.Now()

	now := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	nextDate, err := repeat.NextDate(now, task.Date, task.Repeat)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	task.Date = nextDate

	err = api.store.Update(task)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, map[string]any{})

}

func (api *API) DeleteTaskByID(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := fmt.Sprintf("method %s, is not allowed", r.Method)
		writeJSON(w, map[string]string{
			"error": err,
		})
		return
	}

	id := r.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "Id is not set",
		})
		return
	}

	err := api.store.Delete(id)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, map[string]any{})
}
