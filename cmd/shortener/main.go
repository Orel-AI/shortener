package main

import (
	"github.com/Orel-AI/shortener.git/api/handler"
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	storage.Initialize()
	r := chi.NewRouter()
	r.Get("/{ID}", handler.LookUpOriginalLinkGET)
	r.Post("/", handler.GenerateShorterLinkPOST)
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
