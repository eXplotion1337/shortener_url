package repository

import "sync"

type Rez struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

var mu sync.Mutex

func FindURL(input string) ([]Rez, error) {
	mu.Lock()
	defer mu.Unlock()
	result := make([]Rez, 0, len(InMemoryCollection.ObjectURL))

	if len(InMemoryCollection.ObjectURL) > 0 {
		for _, record := range InMemoryCollection.ObjectURL {
			if record.UserID == input {
				prom := Rez{LongURL: record.LongURL, ShortURL: record.ShortURL}
				result = append(result, prom)
			}
		}
	}

	return result, nil
}
