package shortener

type URL struct {
	ID       string `json:"id"`
	LongURL  string `json:"longURL"`
	ShortURL string `json:"short_url"`
	UserID   string `json:"Authorization"`
}
