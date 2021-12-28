package main


import (
	"github.com/Orel-AI/shortener.git/api/handler"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"net/http"
)

func main() {
	shortener.InitializeMap()
	err := http.ListenAndServe(":8080", handler.ShortenerHandler{})
	if err != nil {
		panic(err)
	}
