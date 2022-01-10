package main

import (
	"github.com/Orel-AI/shortener.git/api/handler"
	"github.com/Orel-AI/shortener.git/storage"
	"net/http"
)

func main() {
	storage.Initialize()
	err := http.ListenAndServe(":8080", handler.ShortenerHandler{})
	if err != nil {
		panic(err)
	}
}
