package main

import (
	"github.com/Orel-AI/shortener.git/api/handler"
	"github.com/Orel-AI/shortener.git/config"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {

	envs := config.NewConfig()
	store, err := storage.NewStorage(envs.FileStoragePath, envs.DSNString)
	if err != nil {
		log.Fatal(err)
	}
	service := shortener.NewShortenService(store)
	shortenerHandler := handler.NewShortenerHandler(service, envs.BaseURL, envs.SecretString, envs.CookieName)
	r := chi.NewRouter()
	r.Use(shortenerHandler.GzipHandle)
	r.Use(shortenerHandler.AuthHandler)
	r.Use(middleware.Logger)
	r.Get("/{ID}", shortenerHandler.LookUpOriginalLinkGET)
	r.Get("/api/user/urls", shortenerHandler.LookUpUsersRequest)
	r.Get("/ping", shortenerHandler.PingDBByRequest)
	r.Post("/", shortenerHandler.GenerateShorterLinkPOST)
	r.Post("/api/shorten", shortenerHandler.GenerateShorterLinkPOSTJson)
	r.Post("/api/shorten/batch", shortenerHandler.GenerateShorterLinkPOSTBatch)

	err = http.ListenAndServe(envs.AddressToServe, r)
	if err != nil {
		log.Fatal(err)
	}
}
