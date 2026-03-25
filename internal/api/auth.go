package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
)

func (api *API) initAuthToken() {
	pass := os.Getenv("TODO_PASSWORD")
	if pass == "" {
		api.authToken = ""
		return
	}
	hash := sha256.Sum256([]byte(pass))
	api.authToken = hex.EncodeToString(hash[:])
}

func (api *API) SignInHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{
			"error": "invalid JSON",
		})
		return
	}

	if api.authToken == "" {
		writeJSON(w, map[string]string{})
		return
	}

	inputHash := sha256.Sum256([]byte(req.Password))
	inputToken := hex.EncodeToString(inputHash[:])

	if inputToken != api.authToken {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(w, map[string]string{"error": "incorrect password"})
		return
	}

	writeJSON(w, map[string]string{"token": api.authToken})
}

func (api *API) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if api.authToken == "" {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil || cookie.Value != api.authToken {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
