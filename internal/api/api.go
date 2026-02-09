package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/o1uch/go_final_project/internal/repeat"
)

func Init() {
	http.HandleFunc("/api/nextdate", nextDayHandler)
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {

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
