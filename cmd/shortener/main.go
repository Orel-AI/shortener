package main

import (
	"github.com/Orel-AI/shortener.git/api/handler"
	"github.com/Orel-AI/shortener.git/config"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {

	envs := config.NewConfig()
	store, err := storage.NewStorage(envs.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	service := shortener.NewShortenService(store)
	shortenerHandler := handler.NewShortenerHandler(service, envs.BaseURL)

	r := chi.NewRouter()
	r.Use(shortenerHandler.GzipHandle)
	r.Get("/{ID}", shortenerHandler.LookUpOriginalLinkGET)
	r.Post("/", shortenerHandler.GenerateShorterLinkPOST)
	r.Post("/api/shorten", shortenerHandler.GenerateShorterLinkPOSTJson)

	err = http.ListenAndServe(envs.AddressToServe, r)
	if err != nil {
		log.Fatal(err)
	}
}

func selectRoute() {

}
