package main

import (
	"compress/gzip"
	"github.com/Orel-AI/shortener.git/api/handler"
	"github.com/Orel-AI/shortener.git/config"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func main() {

	envs := config.NewConfig()
	store, err := storage.NewStorage(envs.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	service := shortener.NewShortenService(store)
	shortenerHandler := handler.NewShortenerHandler(service, envs.BaseURL)

	r := chi.NewRouter()
	r.Get("/{ID}", shortenerHandler.LookUpOriginalLinkGET)
	r.Post("/", shortenerHandler.GenerateShorterLinkPOST)
	r.Post("/api/shorten", shortenerHandler.GenerateShorterLinkPOSTJson)

	err = http.ListenAndServe(envs.AddressToServe, gzipHandle(r))
	if err != nil {
		log.Fatal(err)
	}
}
