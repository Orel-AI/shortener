package main

import (
	"github.com/Orel-AI/shortener.git/api/handler"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
)

func main() {
	addressToServe := os.Getenv("SERVER_ADDRESS")
	if len(addressToServe) == 0 {
		addressToServe = "localhost:8080"
	}

	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	if len(fileStoragePath) == 0 {
		fileStoragePath = "storage.txt"
	}

	store, err := storage.NewStorage(fileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	service := shortener.NewShortenService(store)
	shortenerHandler := handler.NewShortenerHandler(service)

	r := chi.NewRouter()
	r.Get("/{ID}", shortenerHandler.LookUpOriginalLinkGET)
	r.Post("/", shortenerHandler.GenerateShorterLinkPOST)
	r.Post("/api/shorten", shortenerHandler.GenerateShorterLinkPOSTJson)

	err = http.ListenAndServe(addressToServe, r)
	if err != nil {
		log.Fatal(err)
	}
}
