package repository

import (
	"testing"
)

func TestSaveAndGetLongURL(t *testing.T) {
	storage := &JSON{}
	longURL := "https://example.com"
	id := "testID"
	urlData := &InMemoryStorage{
		ID:       id,
		LongURL:  longURL,
		ShortURL: "",
		UserID:   "1",
	}

	_, err := storage.SaveURL(urlData)
	if err != nil {
		t.Errorf("Ошибка при сохранении URL: %v", err)
	}

	retrievedURL, _, err := storage.GetLongURL(id)
	if err != nil {
		t.Errorf("Ошибка при получении длинного URL: %v", err)
	}

	if retrievedURL != longURL {
		t.Errorf("Ожидался длинный URL: %s, но получили: %s", longURL, retrievedURL)
	}
}

func TestGetLongURLNotFound(t *testing.T) {
	storage := &JSON{}
	id := "nonExistentID"

	retrievedURL, _, err := storage.GetLongURL(id)
	if err != nil {
		t.Errorf("Ошибка при получении длинного URL: %v", err)
	}

	if retrievedURL != "" {
		t.Errorf("Ожидалась пустая строка для ненайденного ID, но получили: %s", retrievedURL)
	}
}
