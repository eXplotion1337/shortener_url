package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"net/url"
	"shortener/internal"
	"shortener/internal/app/handlers/service/repository"
	"shortener/internal/app/middleware"
	"shortener/internal/config"
	"strings"
	"sync"
)

type ShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func PostAddURL(w http.ResponseWriter, r *http.Request, config *config.Config, storage repository.Storage) {

	var userID string

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		fmt.Println("userID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return

	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	sit, err := url.ParseRequestURI(string(body))
	if err != nil {
		fmt.Println("URL is not valid", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s := sit.String()
	sitr, err := url.PathUnescape(s)
	if err != nil {
		fmt.Println(err, sitr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := internal.GenerateRandomString(10)
	shortURL := config.BaseURL + "/" + id

	newItem := repository.InMemoryStorage{
		ID:       id,
		LongURL:  sitr,
		ShortURL: shortURL,
		UserID:   userID,
		Flag:     false,
	}

	short, err := storage.SaveURL(&newItem)
	if err != nil {
		log.Println("Ошибка сохранения url", err)
	}
	if short != "" {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(short))
	} else {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}

}

func GetByID(w http.ResponseWriter, r *http.Request, config *config.Config, storage repository.Storage) {
	id := chi.URLParam(r, "id")

	long, flag, _ := storage.GetLongURL(id)
	Location := strings.TrimSpace(long)
	w.Header().Set("Location", long)

	if Location != "" && !flag {
		http.Redirect(w, r, Location, http.StatusTemporaryRedirect)
		return
	} else if Location != "" && flag {
		w.WriteHeader(http.StatusGone)
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func PostAPIShorten(w http.ResponseWriter, r *http.Request, config *config.Config, storage repository.Storage) {
	var requestData struct {
		URL string `json:"url"`
	}

	var userID string

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		fmt.Println("userID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Декодируем тело запроса
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.URL == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sit, err := url.ParseRequestURI(requestData.URL)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s := sit.String()
	sitr, err := url.PathUnescape(s)
	if err != nil {
		fmt.Println(err, sitr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Назначаем случайный id
	id := internal.GenerateRandomString(10)
	respoID := config.BaseURL + "/" + id

	newItem := repository.InMemoryStorage{
		ID:       id,
		LongURL:  requestData.URL,
		ShortURL: respoID,
		UserID:   userID,
		Flag:     false,
	}

	storage.SaveURL(&newItem)

	short, _ := storage.SaveURL(&newItem)
	if short != "" {
		response := map[string]string{"result": short}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
	} else {
		response := map[string]string{"result": respoID}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}

}

func GetUrlsHandler(w http.ResponseWriter, r *http.Request) {
	var userID string

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		fmt.Println("userID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	urls, err := repository.FindURL(userID)
	if err != nil {
		http.Error(w, "Не удалось получить список URL пользователя", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		http.Error(w, "Нет данных", http.StatusNoContent)
		return
	}

	jsonResult, err := json.Marshal(urls)
	if err != nil {
		http.Error(w, "Ошибка сериализации JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}

func PingDB(w http.ResponseWriter, r *http.Request, config *config.Config, storage repository.Storage) {
	err := storage.Ping(config)
	if err != nil {
		log.Printf("ошибка storage.Ping: %s", err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PostBatch(w http.ResponseWriter, r *http.Request, config *config.Config, storage repository.Storage) {
	var requests []ShortenRequest

	defer r.Body.Close()

	var userID string
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		fmt.Println("userID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	var responses []ShortenResponse

	for _, req := range requests {

		id := internal.GenerateRandomString(10)
		short := config.BaseURL + "/" + id

		responses = append(responses, ShortenResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      short,
		})

		newURL := repository.InMemoryStorage{
			ID:       id,
			LongURL:  req.OriginalURL,
			ShortURL: short,
			UserID:   userID,
		}

		if _, err := storage.SaveURL(&newURL); err != nil {
			log.Printf("ссылка %s есть в базе", newURL.LongURL)

		}

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func Delete(w http.ResponseWriter, r *http.Request, storage repository.Storage, deleteChan chan<- repository.DeleteRequest, wg *sync.WaitGroup) {
	var urlsToDelete []string
	err := json.NewDecoder(r.Body).Decode(&urlsToDelete)
	if err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		fmt.Println("userID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	deleteChan <- repository.DeleteRequest{UserID: userID, URLs: urlsToDelete}

	w.WriteHeader(http.StatusAccepted)

}
