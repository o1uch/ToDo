package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
)

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

	envPass := os.Getenv("TODO_PASSWORD")

	if envPass == "" {
		writeJSON(w, map[string]string{})
		return
	}

	if req.Password != envPass {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(w, map[string]string{
			"error": "incorrect password",
		})
		return
	}

	hash := sha256.Sum256([]byte(envPass))
	token := hex.EncodeToString(hash[:])

	writeJSON(w, map[string]string{
		"token": token,
	})
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		pass := os.Getenv("TODO_PASSWORD")
		if pass == "" {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		hash := sha256.Sum256([]byte(pass))
		expectedToken := hex.EncodeToString(hash[:])

		if cookie.Value != expectedToken {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
