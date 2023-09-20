package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type gzipResponseWriter struct {
	*gzip.Writer
	http.ResponseWriter
}

func (grw gzipResponseWriter) Write(p []byte) (int, error) {
	return grw.Writer.Write(p)
}

func (grw gzipResponseWriter) Header() http.Header {
	return grw.ResponseWriter.Header()
}

func (grw gzipResponseWriter) WriteHeader(statusCode int) {
	grw.ResponseWriter.WriteHeader(statusCode)
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if gzip is accepted
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// create gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// set headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// wrap response writer and serve
		gw := gzipResponseWriter{gz, w}
		next.ServeHTTP(gw, r)
	})

}

func UngzipData(data []byte) (string, error) {
	isCompressed := false
	if len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b {
		isCompressed = true
	}

	if isCompressed {
		// Распаковываем данные из GZIP
		gzipReader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return "", err
		}

		uncompressedData := bytes.Buffer{}
		_, err = uncompressedData.ReadFrom(gzipReader)
		if err != nil {
			return "", err
		}

		err = gzipReader.Close()
		if err != nil {
			return "", err
		}

		return uncompressedData.String(), nil
	}
	// Если данные не сжаты, то просто конвертируем их в строку
	return string(data), nil

}

func GZipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			r.Body, _ = gzip.NewReader(r.Body)
		}

		next.ServeHTTP(w, r)
	})
}

type contextKey string

const UserIDKey contextKey = "userID"

func SetUserIDCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		if errors.Is(err, http.ErrNoCookie) || len(cookie.Value) == 0 {
			uid := uuid.New()

			// Создаем симметрично подписанную куку и устанавливаем ее в браузере пользователя
			http.SetCookie(w, &http.Cookie{
				Name:  "Authorization",
				Value: uid.String(),
			})

			// Добавляем user_id в контекст запроса
			ctx := context.WithValue(r.Context(), UserIDKey, uid.String())
			r = r.WithContext(ctx)
		} else {
			userID := cookie.Value
			_, err := uuid.Parse(userID)
			if err != nil {
				http.Error(w, "Некорректный идентификатор пользователя", http.StatusUnauthorized)
				return
			}

			// Добавляем user_id в контекст запроса
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			r = r.WithContext(ctx)
		}

		// Продолжаем обработку запроса
		next.ServeHTTP(w, r)
	})
}
