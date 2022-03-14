package main

import (
	"compress/gzip"
	"github.com/Orel-AI/shortener.git/api/handler"
	"github.com/Orel-AI/shortener.git/config"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

var defaultCompressibleContentTypes = []string{
	"text/html",
	"text/css",
	"text/plain",
	"text/javascript",
	"application/javascript",
	"application/x-javascript",
	"application/json",
	"application/atom+xml",
	"application/rss+xml",
	"image/svg+xml",
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
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(gzip.DefaultCompression, defaultCompressibleContentTypes...))
	r.Get("/{ID}", shortenerHandler.LookUpOriginalLinkGET)
	r.Post("/", shortenerHandler.GenerateShorterLinkPOST)
	r.Post("/api/shorten", shortenerHandler.GenerateShorterLinkPOSTJson)

	err = http.ListenAndServe(envs.AddressToServe, r)
	if err != nil {
		log.Fatal(err)
	}
}
