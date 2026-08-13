package main

import (
	"strings"
	"testing"
)

func TestURLShortener_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"валидный HTTP URL", "http://example.com", false},
		{"валидный HTTPS URL", "https://google.com/search?q=test", false},
		{"невалидный URL", "not-a-url", true},
		{"пустая строка", "", true},
		{"URL без схемы", "example.com", true},
		{"URL с неподдерживаемой схемой", "ftp://example.com/file", true},
	}

	shortener := NewURLShortener()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID, err := shortener.Shorten(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(shortID) < 6 || len(shortID) > 8 {
					t.Errorf("длина короткого ID должна быть 6-8 символов, получено %d (%s)", len(shortID), shortID)
				}
			}
		})
	}
}

func TestURLShortener_Shorten_Uniqueness(t *testing.T) {
	shortener := NewURLShortener()

	ids := make(map[string]bool)
	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
		"https://example.com/4",
		"https://example.com/5",
	}

	for _, u := range urls {
		id, err := shortener.Shorten(u)
		if err != nil {
			t.Fatalf("неожиданная ошибка при сокращении %s: %v", u, err)
		}
		if ids[id] {
			t.Errorf("получен дублирующийся короткий ID: %s", id)
		}
		ids[id] = true
	}
}

func TestURLShortener_GetOriginal(t *testing.T) {
	shortener := NewURLShortener()

	originalURL := "https://example.com/very/long/path"
	shortID, err := shortener.Shorten(originalURL)
	if err != nil {
		t.Fatalf("неожиданная ошибка при сокращении: %v", err)
	}

	tests := []struct {
		name    string
		shortID string
		want    string
		wantErr bool
	}{
		{"существующий short_url", shortID, originalURL, false},
		{"несуществующий short_url", "notexist", "", true},
		{"пустой short_url", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shortener.GetOriginal(tt.shortID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("получено %q, ожидали %q", got, tt.want)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"валидный http", "http://example.com", true},
		{"валидный https", "https://example.com/path?query=1", true},
		{"пустая строка", "", false},
		{"без схемы", "example.com", false},
		{"неподдерживаемая схема", "ftp://example.com", false},
		{"мусорная строка", "not-a-url", false},
		{"только схема без хоста", "http://", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidURL(tt.url)
			if got != tt.want {
				t.Errorf("isValidURL(%q) = %v, ожидали %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestGenerateShortID(t *testing.T) {
	id := generateShortID()

	if len(id) < 6 || len(id) > 8 {
		t.Errorf("длина short ID должна быть 6-8 символов, получено %d (%s)", len(id), id)
	}

	if strings.ContainsAny(id, " \t\n") {
		t.Errorf("short ID не должен содержать пробельные символы: %q", id)
	}
}
