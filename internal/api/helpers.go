package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/o1uch/go_final_project/internal/store"
)

func trimExtraSpaces(t *store.Task) {
	t.Title = strings.TrimSpace(t.Title)
	t.Comment = strings.TrimSpace(t.Comment)
	t.Date = strings.TrimSpace(t.Date)
	t.Repeat = strings.TrimSpace(t.Repeat)
}

func writeJSON(w http.ResponseWriter, data any) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
