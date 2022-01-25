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
	store := storage.NewStorage()
	service := shortener.NewShortenService(store)
	shortenerHandler := handler.NewShortenerHandler(service)
	r := chi.NewRouter()
	r.Get("/{ID}", shortenerHandler.LookUpOriginalLinkGET)
	r.Post("/", shortenerHandler.GenerateShorterLinkPOST)
	r.Post("/api/shorten", shortenerHandler.GenerateShorterLinkPOSTJson)

	err := http.ListenAndServe(os.Getenv("SERVER_ADDRESS"), r)
	if err != nil {
		log.Fatal(err)
	}
}
