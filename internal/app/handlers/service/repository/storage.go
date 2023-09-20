package repository

import (
	"errors"
	"fmt"
	"shortener/internal/config"
	"strings"
	"sync"
)

type InMemoryStorage struct {
	ID       string `json:"id"`
	LongURL  string `json:"longURL"`
	ShortURL string `json:"short_url"`
	UserID   string `json:"userID"`
	Flag     bool   `json:"flag"`
}

type JSON struct {
	sync.Mutex
	ObjectURL []InMemoryStorage
}

var InMemoryCollection JSON

type DeleteRequest struct {
	UserID string   // Идентификатор пользователя
	URLs   []string // Список URL для удаления
}

type Storage interface {
	SaveURL(longURL *InMemoryStorage) (sortURL string, err error)
	GetLongURL(id string) (longURL string, flag bool, err error)
	DeleteURL(ids []string, user string) error
	Ping(config *config.Config) error
}

func (in *JSON) SaveURL(longURL *InMemoryStorage) (sortURL string, err error) {
	in.Lock()
	defer in.Unlock()
	InMemoryCollection.ObjectURL = append(InMemoryCollection.ObjectURL, *longURL)
	return "", nil
}

func (in *JSON) GetLongURL(id string) (longURL string, flag bool, err error) {
	in.Lock()
	defer in.Unlock()
	fmt.Println(InMemoryCollection.ObjectURL, id)
	if len(InMemoryCollection.ObjectURL) > 0 {
		for _, v := range InMemoryCollection.ObjectURL {
			if strings.EqualFold(v.ID, id) {
				return v.LongURL, v.Flag, nil
			}
		}
	}
	return "", false, nil
}

func (in *JSON) DeleteURL(ids []string, user string) error {
	in.Lock()
	defer in.Unlock()

	if len(InMemoryCollection.ObjectURL) == 0 {
		return errors.New("коллекция пуста")
	}

	deleted := false

	for i, v := range InMemoryCollection.ObjectURL {
		for _, id := range ids {
			if strings.EqualFold(v.ID, id) && v.UserID == user {
				InMemoryCollection.ObjectURL[i].Flag = true
				deleted = true
				break
			}
		}
	}

	if !deleted {
		return errors.New("URL-ы не найдены для удаления")
	}
	return nil
}

func (in *JSON) Ping(config *config.Config) error {
	return nil
}

func DeleteHandler(storage Storage, deleteChan chan DeleteRequest, wg *sync.WaitGroup) {
	for req := range deleteChan {
		userID := req.UserID
		urlsToDelete := req.URLs
		wg.Add(1)

		go func() {
			defer wg.Done()
			err := storage.DeleteURL(urlsToDelete, userID)
			if err != nil {
				fmt.Printf("Ошибка при удалении URL %s: %v\n", urlsToDelete, err)
			}

		}()
	}
}
