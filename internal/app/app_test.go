package app

import (
	"net/http"
	"net/http/httptest"
	"shortener/internal/app/handlers"
	"shortener/internal/app/handlers/service/repository"
	"shortener/internal/config"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRun(t *testing.T) {
	// Создаем фейковый конфиг и хранилище
	fakeConfig := &config.Config{
		ServerAddr:  "127.0.0.1:8080",
		BaseURL:     "http://localhost",
		StoragePath: "path/to/storage",
		DataBaseDSN: "database_dsn",
		TypeStorage: "memory",
	}
	fakeStorage := &repository.JSON{} // Замените на фейковое хранилище
	deleteChan := make(chan repository.DeleteRequest, 100)
	var wg sync.WaitGroup

	// Создаем фейковый маршрутизатор
	r := chi.NewRouter()

	// Заменяем Post и Get обработчики на фейковые
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostAddURL(w, r, fakeConfig, fakeStorage)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetByID(w, r, fakeConfig, fakeStorage)
	})

	// Создаем тестовый сервер
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// Запускаем функцию Run в фоновом режиме
	go func() {
		err := Run(fakeConfig, fakeStorage, deleteChan, &wg)
		if err != nil {
			t.Errorf("Ошибка при запуске сервера: %v", err)
		}
		wg.Wait()
		close(deleteChan)
	}()

	// Выполняем GET-запрос к серверу (замените "your-id" на реальный ID)
	response, err := http.Get(testServer.URL + "/123")
	if err != nil {
		t.Errorf("Ошибка при выполнении GET-запроса: %v", err)
	}
	defer response.Body.Close()

	// Проверяем код ответа
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, но получили %d", http.StatusBadRequest, response.StatusCode)
	}

}
