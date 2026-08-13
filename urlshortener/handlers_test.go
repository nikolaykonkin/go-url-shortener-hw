package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{
			name:       "успешное сокращение",
			method:     http.MethodPost,
			body:       `{"url": "https://example.com/very/long/path"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "некорректный JSON",
			method:     http.MethodPost,
			body:       `{invalid json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "невалидный URL",
			method:     http.MethodPost,
			body:       `{"url": "not-a-url"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "пустой URL",
			method:     http.MethodPost,
			body:       `{"url": ""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "неподдерживаемый метод",
			method:     http.MethodGet,
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := NewURLShortener()
			handler := shortenHandler(us)

			req := httptest.NewRequest(tt.method, "/shorten", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			handler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("статус = %d, ожидали %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var resp shortenResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("не удалось распарсить ответ: %v", err)
				}
				if len(resp.ShortURL) < 6 || len(resp.ShortURL) > 8 {
					t.Errorf("short_url должен быть 6-8 символов, получено: %s", resp.ShortURL)
				}
				if resp.OriginalURL == "" {
					t.Errorf("original_url не должен быть пустым")
				}
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	us := NewURLShortener()
	originalURL := "https://example.com/very/long/path"
	shortID, err := us.Shorten(originalURL)
	if err != nil {
		t.Fatalf("не удалось подготовить тестовые данные: %v", err)
	}

	tests := []struct {
		name         string
		method       string
		path         string
		wantStatus   int
		wantLocation string
	}{
		{
			name:         "успешный редирект",
			method:       http.MethodGet,
			path:         "/" + shortID,
			wantStatus:   http.StatusFound,
			wantLocation: originalURL,
		},
		{
			name:       "несуществующий short_url",
			method:     http.MethodGet,
			path:       "/doesnotexist",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "неподдерживаемый метод",
			method:     http.MethodPost,
			path:       "/" + shortID,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
            name:       "пустой путь",
            method:     http.MethodGet,
            path:       "/",
            wantStatus: http.StatusNotFound,
        },
	}

	handler := redirectHandler(us)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("статус = %d, ожидали %d", rec.Code, tt.wantStatus)
			}

			if tt.wantLocation != "" {
				gotLocation := rec.Header().Get("Location")
				if gotLocation != tt.wantLocation {
					t.Errorf("Location = %q, ожидали %q", gotLocation, tt.wantLocation)
				}
			}
		})
	}
}
