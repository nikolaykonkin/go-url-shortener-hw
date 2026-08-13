package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"sync"
)

// URLShortener хранит соответствие короткий_идентификатор -> оригинальный URL.
type URLShortener struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]string),
	}
}

var (
	// ErrInvalidURL возвращается, если переданная строка не является
	// корректным HTTP/HTTPS адресом.
	ErrInvalidURL = errors.New("невалидный URL")
	// ErrShortNotFound возвращается, если по короткому идентификатору
	// не найден оригинальный URL.
	ErrShortNotFound = errors.New("короткий идентификатор не найден")
)

// Shorten создаёт короткий идентификатор для URL.
func (us *URLShortener) Shorten(originalURL string) (string, error) {
	if !isValidURL(originalURL) {
		return "", ErrInvalidURL
	}

	us.mu.Lock()
	defer us.mu.Unlock()

	var shortID string
	for {
		shortID = generateShortID()
		if _, exists := us.urls[shortID]; !exists {
			break
		}
	}

	us.urls[shortID] = originalURL
	return shortID, nil
}

// GetOriginal возвращает оригинальный URL по короткому идентификатору.
func (us *URLShortener) GetOriginal(shortID string) (string, error) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	originalURL, exists := us.urls[shortID]
	if !exists {
		return "", ErrShortNotFound
	}
	return originalURL, nil
}

// generateShortID генерирует случайный короткий идентификатор длиной 6-8 символов.
func generateShortID() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read практически никогда не возвращает ошибку
		// на поддерживаемых платформах, но на случай сбоя ОС
		// возвращаем детерминированный fallback, а не паникуем.
		return "000000"
	}

	encoded := base64.RawURLEncoding.EncodeToString(b)
	if len(encoded) > 8 {
		encoded = encoded[:8]
	}
	return encoded
}

// isValidURL проверяет, что строка является корректным HTTP/HTTPS адресом с хостом.
func isValidURL(str string) bool {
	if str == "" {
		return false
	}

	u, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	return true
}
