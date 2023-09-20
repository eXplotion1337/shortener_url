package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log"
	"shortener/internal/config"
)

type DatabaseStorage struct {
	db *sql.DB
}

func NewDatabaseStorage(db *sql.DB) *DatabaseStorage {
	return &DatabaseStorage{
		db: db,
	}
}

func (ds *DatabaseStorage) SaveURL(item *InMemoryStorage) (string, error) {
	insertQuery := `
		INSERT INTO urls (id, long_url, short_url, user_id, flag)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (long_url) DO NOTHING
	`

	getShortURL := `
		SELECT short_url FROM urls WHERE long_url = $1
	`

	_, err := ds.db.Exec(insertQuery, item.ID, item.LongURL, item.ShortURL, item.UserID, item.Flag)
	if err != nil {
		return "", err
	}

	var shortURL string
	err = ds.db.QueryRow(getShortURL, item.LongURL).Scan(&shortURL)
	if err != nil {
		return "", err
	}

	if shortURL != item.ShortURL {
		return shortURL, nil
	}

	return "", nil
}

func (ds *DatabaseStorage) GetLongURL(id string) (string, bool, error) {

	var longURL string
	var flag bool

	selectQuery := `
		SELECT long_url, flag  FROM urls WHERE id = $1
	`

	err := ds.db.QueryRow(selectQuery, id).Scan(&longURL, &flag)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, fmt.Errorf("URL not found")
		}
		return "", false, err
	}

	return longURL, flag, nil
}

func (ds *DatabaseStorage) DeleteURL(ids []string, user string) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
        UPDATE urls
        SET flag = true
        WHERE user_id = $1 AND id = ANY($2)
    `

	_, err := ds.db.Exec(query, user, pq.Array(ids))
	if err != nil {
		log.Printf("ошибка изменения флага %s", err)
		return err
	}
	return nil
}

func (ds *DatabaseStorage) Ping(config *config.Config) error {

	err := ds.db.Ping()
	if err != nil {
		fmt.Println(err)
		//http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return err
	}
	return nil
}

func createURLsTable(db *sql.DB) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS urls (
			id VARCHAR(36) PRIMARY KEY,
			long_url TEXT UNIQUE NOT NULL,
			short_url VARCHAR(100) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
		    flag BOOLEAN NOT NULL
		)
	`

	_, err := db.Exec(createTableQuery)
	return err
}

func CheckBD(databaseDSN string) error {
	if databaseDSN == "" {
		log.Println("DATABASE_DSN environment variable is not set")
		return errors.New("DATABASE_DSN environment variable is not set")
	}

	// Открытие соединения с базой данных
	db, err := sql.Open("postgres", databaseDSN)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	err = createURLsTable(db)
	if err != nil {
		log.Printf("Нет доступа к БД: %s", err)
		return err
	}

	return nil
}
