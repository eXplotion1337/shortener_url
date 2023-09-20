package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"shortener/internal/config"
	"strings"
	"sync"
)

type FileStorage struct {
	filename string
	addData  sync.Mutex
}

func NewFileStorage(filename string) *FileStorage {
	return &FileStorage{
		filename: filename,
	}
}

func (fs *FileStorage) GetLongURL(id string) (string, bool, error) {
	InMemoryCollection.Mutex.Lock()
	defer InMemoryCollection.Mutex.Unlock()

	for _, v := range InMemoryCollection.ObjectURL {
		if strings.EqualFold(v.ID, id) {
			return v.LongURL, v.Flag, nil
		}
	}

	return "", false, fmt.Errorf("URL not found")
}

func (fs *FileStorage) SaveURL(longURL *InMemoryStorage) (shortURL string, err error) {
	fs.addData.Lock()
	defer fs.addData.Unlock()

	file, err := os.OpenFile(fs.filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	if fileInfo.Size() == 0 {
		_, err = file.Write([]byte("{}"))
		if err != nil {
			return "", err
		}
	}

	jsonData, err := os.ReadFile(fs.filename)
	if err != nil {
		return "", err
	}

	var obj JSON
	if err = json.Unmarshal(jsonData, &obj); err != nil {
		return "", err
	}

	InMemoryCollection.Mutex.Lock()
	defer InMemoryCollection.Mutex.Unlock()

	obj.ObjectURL = append(obj.ObjectURL, *longURL)
	InMemoryCollection.ObjectURL = append(InMemoryCollection.ObjectURL, *longURL)

	if jsonData, err = json.Marshal(&obj); err != nil {
		return "", err
	}
	if _, err = file.WriteAt(jsonData, 0); err != nil {
		return "", err
	}

	return "", nil
}

func (fs *FileStorage) DeleteURL(ids []string, user string) error {
	fs.addData.Lock()
	defer fs.addData.Unlock()

	// Открываем файл для чтения и записи
	file, err := os.OpenFile(fs.filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Читаем содержимое файла
	jsonData, err := os.ReadFile(fs.filename)
	if err != nil {
		return err
	}

	var obj JSON
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		return err
	}

	deleted := false

	for i, url := range obj.ObjectURL {
		for _, id := range ids {
			if url.ID == id && url.UserID == user {
				obj.ObjectURL[i].Flag = true
				deleted = true
				break
			}
		}
	}

	if !deleted {
		return errors.New("URL-ы не найдены для удаления")
	}

	updatedData, err := json.Marshal(&obj)
	if err != nil {
		return err
	}

	if _, err := file.WriteAt(updatedData, 0); err != nil {
		return err
	}

	return nil
}

func (fs *FileStorage) Ping(config *config.Config) error {
	_, err := os.Open(config.StoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Файл %s не существует\n", config.StoragePath)
		} else {
			log.Printf("Ошибка при проверке файла: %v\n", err)
		}
		return err
	}

	log.Printf("Файл %s существует\n", config.StoragePath)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Ошибка при закрытии файла: %v\n", err)
		}
	}()

	return nil
}

func ReadJSONFile(filepath string) error {

	jsonFile, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonFile, &InMemoryCollection)
	if err != nil {
		return err
	}

	return nil
}

func CreateFileIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// получаем директорию, где должен быть файл
		dir := filepath.Dir(path)

		// создаем все директории в пути к файлу
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// создаем сам файл
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}
