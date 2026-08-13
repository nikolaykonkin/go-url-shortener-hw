package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// shortenHandler обрабатывает POST /shorten: принимает JSON с полем url,
// валидирует и сокращает его, возвращает short_url и original_url.
func shortenHandler(us *URLShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
			return
		}

		var req shortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "некорректный JSON")
			return
		}

		shortID, err := us.Shorten(req.URL)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp := shortenResponse{
			ShortURL:    shortID,
			OriginalURL: req.URL,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

// redirectHandler обрабатывает GET /{short_url}: ищет оригинальный URL
// по короткому идентификатору и делает HTTP 302 редирект, либо 404,
// если идентификатор не найден.
func redirectHandler(us *URLShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
			return
		}

		shortID := strings.TrimPrefix(r.URL.Path, "/")
		if shortID == "" {
			writeError(w, http.StatusNotFound, "короткий идентификатор не указан")
			return
		}

		originalURL, err := us.GetOriginal(shortID)
		if err != nil {
			writeError(w, http.StatusNotFound, "url не найден")
			return
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}

func main() {
	us := NewURLShortener()

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", shortenHandler(us))
	mux.HandleFunc("/", redirectHandler(us))

	log.Println("Сервер запущен на :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
