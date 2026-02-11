package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/o1uch/go_final_project/internal/repeat"
	"github.com/o1uch/go_final_project/internal/store"
)

type API struct {
	store *store.SchedulerStore
}

func Init(store *store.SchedulerStore) {
	api := API{store: store}
	http.HandleFunc("/api/nextdate", api.nextDayHandler)
	http.HandleFunc("/api/task", api.MainTaskHandler)
}

func (api *API) nextDayHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		err := fmt.Sprintf("method %s, is not allowed", r.Method)
		http.Error(w, err, http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.URL.Query().Get("now")
	now := time.Now()

	if nowStr != "" {
		var err error
		now, err = time.Parse(repeat.DateLayout, nowStr)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	}

	dstart := r.URL.Query().Get("date")

	if dstart == "" {
		err := "the \"date\" parameter is not specified"
		http.Error(w, err, http.StatusBadRequest)
		return
	}

	repeatValue := r.URL.Query().Get("repeat")

	nextDate, err := repeat.NextDate(now, dstart, repeatValue)

	if err != nil {
		http.Error(w, "Error in calculating the nextDate: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(nextDate))

}

func (api *API) MainTaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		api.addTaskHandler(w, r)
	default:
		http.Error(w, "The post method should be used to create a task", http.StatusMethodNotAllowed)
		return
	}
}

func trimExtraSpaces(t *store.Task) {
	t.Title = strings.TrimSpace(t.Title)
	t.Comment = strings.TrimSpace(t.Comment)
	t.Date = strings.TrimSpace(t.Date)
}

func (api *API) addTaskHandler(w http.ResponseWriter, r *http.Request) {

	newTask := store.Task{}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &newTask)

	if err != nil {
		http.Error(w, "Invalid JOSN: "+err.Error(), http.StatusBadRequest)
		return
	}

	trimExtraSpaces(&newTask)

	if newTask.Title == "" {
		http.Error(w, "error: "+"the issue title is not specified", http.StatusBadRequest)
		return
	}

	now := time.Now()
	nowStr := now.Format("20060102")

	if newTask.Date == "" {
		newTask.Date = nowStr
	}

	date, err := time.Parse("20060102", newTask.Date)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if date.Before(now) {

		if newTask.Repeat == "" {
			newTask.Date = nowStr
		} else {
			nextDate, err := repeat.NextDate(now, newTask.Date, newTask.Repeat)

			if err != nil {
				http.Error(w, "error: "+err.Error(), http.StatusBadRequest)
				return
			}
			newTask.Date = nextDate
		}

	}

	lastID, err := api.store.AddTask(&newTask)

	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusBadRequest)
	}

	resp := fmt.Sprintf("\"id\":\"%d\"", lastID)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte(resp))

}
