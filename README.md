# go-url-shortener-hw
URL shortener HTTP service in Go with in-memory storage, URL validation, table-driven unit tests, and httptest-based handler tests (>80% coverage).


# Домашнее задание к занятию "Разработка тестируемого приложения" - Конкин Николай


### Инструкция по выполнению домашнего задания

   1. Сделайте `fork` данного репозитория к себе в Github и переименуйте его по названию или номеру занятия, например, https://github.com/имя-вашего-репозитория/git-hw или  https://github.com/имя-вашего-репозитория/7-1-ansible-hw).
   2. Выполните клонирование данного репозитория к себе на ПК с помощью команды `git clone`.
   3. Выполните домашнее задание и заполните у себя локально этот файл README.md:
      - впишите вверху название занятия и вашу фамилию и имя
      - в каждом задании добавьте решение в требуемом виде (текст/код/скриншоты/ссылка)
      - для корректного добавления скриншотов воспользуйтесь [инструкцией "Как вставить скриншот в шаблон с решением](https://github.com/netology-code/sys-pattern-homework/blob/main/screen-instruction.md)
      - при оформлении используйте возможности языка разметки md (коротко об этом можно посмотреть в [инструкции  по MarkDown](https://github.com/netology-code/sys-pattern-homework/blob/main/md-instruction.md))
   4. После завершения работы над домашним заданием сделайте коммит (`git commit -m "comment"`) и отправьте его на Github (`git push origin`);
   5. В личном кабинете прикрепите и отправьте ссылку на решение в виде md-файла в вашем Github.
   6. Любые вопросы по выполнению заданий спрашивайте в разделе “Вопросы по заданию” в личном кабинете.

Желаем успехов в выполнении домашнего задания!

### Дополнительные материалы, которые могут быть полезны для выполнения задания

1. [Руководство по оформлению Markdown файлов](https://gist.github.com/Jekins/2bf2d0638163f1294637#Code)

---

### Задание 1

Разработать HTTP-сервис для сокращения URL-адресов с полным покрытием unit-тестами и тестами HTTP-обработчиков: эндпоинт `POST /shorten` для сокращения URL и `GET /{short_url}` для редиректа на оригинальный адрес, хранение данных в памяти, валидация входных URL и обработка ошибок.

---

Структура проекта:

```
├── README.md
├── img/                          // скриншоты для README
│   ├── test-verbose.png
│   ├── test-coverage.png
│   └── coverage-html.png
└── urlshortener/
    ├── go.mod
    ├── main.go                   // HTTP-сервер и маршруты
    ├── shortener.go              // Бизнес-логика сокращения URL
    ├── shortener_test.go         // Unit-тесты бизнес-логики (table-driven)
    └── handlers_test.go          // Тесты HTTP-обработчиков (httptest)
```

---

Этапы выполнения:

1. **Бизнес-логика (`shortener.go`).** Реализован тип `URLShortener` с полем `urls map[string]string` и `sync.RWMutex` для потокобезопасного доступа. Метод `Shorten` валидирует URL через `isValidURL`, генерирует уникальный короткий идентификатор через `generateShortID` (с проверкой на коллизии в цикле) и сохраняет пару в map под записью (`Lock`). Метод `GetOriginal` ищет URL по идентификатору под чтением (`RLock`) и возвращает `ErrShortNotFound`, если ключ отсутствует.
2. **Генерация идентификатора.** `generateShortID` использует `crypto/rand` для получения 6 случайных байт и кодирует их в `base64.RawURLEncoding` (URL-safe алфавит, без паддинга), обрезая результат до 8 символов — итоговая длина укладывается в требуемые 6–8 символов.
3. **Валидация URL.** `isValidURL` проверяет, что строка не пустая, парсится через `url.ParseRequestURI`, имеет схему `http` или `https` и непустой хост — так отсекаются как мусорные строки (`not-a-url`), так и URL без схемы (`example.com`) или с неподдерживаемой схемой (`ftp://...`).
4. **HTTP-обработчики (`main.go`).** `shortenHandler` разбирает JSON-тело запроса, вызывает `Shorten` и возвращает `{"short_url": ..., "original_url": ...}` с кодом 200, либо ошибку с кодом 400 (некорректный JSON или невалидный URL) или 405 (не тот метод). `redirectHandler` достаёт короткий идентификатор из пути запроса, вызывает `GetOriginal` и делает `http.Redirect` с кодом 302, либо возвращает 404, если идентификатор не найден.
5. **Unit-тесты бизнес-логики (`shortener_test.go`).** Table-driven тесты для `Shorten` (валидные/невалидные URL, пустая строка, URL без схемы, неподдерживаемая схема), отдельный тест на уникальность генерируемых идентификаторов при массовом сокращении, table-driven тест для `GetOriginal` (существующий/несуществующий/пустой short_url) и для `isValidURL`. Каждый кейс оформлен через `t.Run` с говорящим названием.
6. **Тесты HTTP-обработчиков (`handlers_test.go`).** С помощью `httptest.NewRequest` и `httptest.NewRecorder` протестированы оба обработчика: успешные сценарии (сокращение URL, редирект) и ошибочные (некорректный JSON, невалидный URL, пустой URL, несуществующий short_url, неподдерживаемый HTTP-метод) — каждый как отдельный под-тест через `t.Run`.
7. **Проверка покрытия.** Команда `go test -coverprofile=coverage.out ./...` и `go test -cover ./...` используются для контроля, что бизнес-логика (`shortener.go`) покрыта тестами не менее чем на 80%.

---

```go
// shortener.go
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
```

---

```go
// main.go
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
```

---

```go
// shortener_test.go
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
```

---

```go
// handlers_test.go
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
```

---

Все тесты пройдены успешно (`go test -v ./...`):

![Вывод go test -v, часть 1](https://github.com/nikolaykonkin/go-url-shortener-hw/blob/main/img/test-verbose-1.png)

![Вывод go test -v, часть 2](https://github.com/nikolaykonkin/go-url-shortener-hw/blob/main/img/test-verbose-2.png)

Результат `go test -cover ./...`:

![Вывод go test -cover](https://github.com/nikolaykonkin/go-url-shortener-hw/blob/main/img/test-coverage.png)

HTML-отчёт покрытия (`go tool cover -html=coverage.out`) — бизнес-логика в `shortener.go` покрыта на 94.3%:

![HTML-отчёт покрытия shortener.go, часть 1](https://github.com/nikolaykonkin/go-url-shortener-hw/blob/main/img/coverage-html-shortener-1.png)

![HTML-отчёт покрытия shortener.go, часть 2](https://github.com/nikolaykonkin/go-url-shortener-hw/blob/main/img/coverage-html-shortener-2.png)

HTML-отчёт покрытия для `main.go` (HTTP-обработчики и точка входа):

![HTML-отчёт покрытия main.go, часть 1](https://github.com/nikolaykonkin/go-url-shortener-hw/blob/main/img/coverage-html-main-1.png)

![HTML-отчёт покрытия main.go, часть 2](https://github.com/nikolaykonkin/go-url-shortener-hw/blob/main/img/coverage-html-main-2.png)

Пример команд для проверки (выполняются внутри папки `urlshortener/`, так как именно там лежит `go.mod`):

```bash
cd urlshortener
go test -v ./...
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

Как выполнены ключевые требования задания:

| Требование | Реализация |
|---|---|
| `POST /shorten` | `shortenHandler` — валидация, сокращение, JSON-ответ, коды 200/400/405 |
| `GET /{short_url}` | `redirectHandler` — поиск оригинала, `http.Redirect` (302), код 404 при отсутствии |
| Хранение в памяти | `map[string]string` внутри `URLShortener`, защищён `sync.RWMutex` |
| Валидация URL | `isValidURL` — проверка схемы (`http`/`https`) и наличия хоста |
| Обработка ошибок | Некорректный JSON, невалидный URL, отсутствующий short_url — отдельные коды и сообщения |
| Table-driven tests | `TestURLShortener_Shorten`, `TestURLShortener_GetOriginal`, `TestIsValidURL`, `TestShortenHandler`, `TestRedirectHandler` |
| `httptest` | `handlers_test.go` — `httptest.NewRequest` + `httptest.NewRecorder` |
| `t.Run` | Используется во всех тестах для организации под-тестов |
| Покрытие ≥ 80% для бизнес-логики | Проверяется командой `go test -coverprofile=coverage.out` (см. результат выше) |
