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
		writeJSON(w, map[string]string{
			"error": "method not allowed",
		}, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{
			"error": "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	if api.authToken == "" {
		writeJSON(w, map[string]string{}, http.StatusOK)
		return
	}

	inputHash := sha256.Sum256([]byte(req.Password))
	inputToken := hex.EncodeToString(inputHash[:])

	if inputToken != api.authToken {
		writeJSON(w, map[string]string{"error": "incorrect password"}, http.StatusUnauthorized)
		return
	}

	writeJSON(w, map[string]string{"token": api.authToken}, http.StatusOK)
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
