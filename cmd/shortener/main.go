package main

import (
	"log"
	"shortener/internal/app"
	"shortener/internal/app/handlers/service/repository"
	"shortener/internal/config"
	"sync"
)

func main() {
	config, err := config.InitConfig()
	if err != nil {
		log.Fatal("Ошибка загрузки конфига", err)
	}

	storage, err := app.InitStorage(config)
	if err != nil {
		log.Fatal("Ошибка создания хранилища", err)
	}

	deleteChan := make(chan repository.DeleteRequest, 100)
	var wg sync.WaitGroup

	if err := app.Run(config, storage, deleteChan, &wg); err != nil {
		log.Fatal("Ошибка старта сервера", err)
	}

	wg.Wait()
	close(deleteChan)

}
