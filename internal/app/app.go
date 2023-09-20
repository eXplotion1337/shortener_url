package app

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"sync"

	"shortener/internal/app/handlers"
	"shortener/internal/app/handlers/service/repository"
	"shortener/internal/app/middleware"
	"shortener/internal/config"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func Run(config *config.Config, storage repository.Storage, deleteChan chan repository.DeleteRequest, wg *sync.WaitGroup) error {
	r := chi.NewRouter()
	r.Use(middleware.GZipMiddleware)
	r.Use(middleware.SetUserIDCookie)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostAddURL(w, r, config, storage)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetByID(w, r, config, storage)
	})

	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostAPIShorten(w, r, config, storage)
	})

	r.Get("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetUrlsHandler(w, r)
	})

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		handlers.PingDB(w, r, config, storage)
	})

	r.Post("/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostBatch(w, r, config, storage)
	})

	r.Delete("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		handlers.Delete(w, r, storage, deleteChan, wg)
	})

	log.Printf("Сервер запущен на %s", config.ServerAddr)
	log.Printf("Base URL  %s", config.BaseURL)
	log.Printf("Файл для сохранения данных расположен %s", config.StoragePath)
	log.Printf("База данных  %s", config.DataBaseDSN)
	log.Printf("Хранение данных реализовано через  %s", config.TypeStorage)

	go repository.DeleteHandler(storage, deleteChan, wg)

	return http.ListenAndServe(config.ServerAddr, r)
}

func InitStorage(conf *config.Config) (repository.Storage, error) {
	var storage repository.Storage

	switch conf.TypeStorage {
	case "In-memoryStorage":
		storage = &repository.JSON{}

	case "FileStorage":
		storage = repository.NewFileStorage(conf.StoragePath)

		err := repository.CreateFileIfNotExists(conf.StoragePath)
		if err != nil {
			log.Println("Ошибка создания файла", err)
			return nil, err
		}

		err = repository.ReadJSONFile(conf.StoragePath)
		if err != nil {
			log.Println("Ошибка чтения файла", err)
		}

	case "DataBaseStorage":
		db, err := sql.Open("postgres", conf.DataBaseDSN)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		storage = repository.NewDatabaseStorage(db)
		err = repository.CheckBD(conf.DataBaseDSN)
		if err != nil {
			log.Println("Ошибка соединения с БД", err)
			return nil, err
		}

	default:
		return nil, errors.New("не удалось инициализировать хранилище")

	}

	return storage, nil
}
