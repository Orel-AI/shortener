package main

import (
	"flag"
	"github.com/Orel-AI/shortener.git/api/handler"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
)

type env struct {
	addressToServe  string
	baseURL         string
	fileStoragePath string
}

func main() {

	address := flag.String("a", os.Getenv("SERVER_ADDRESS"), "address to start up server")
	baseURL := flag.String("b", os.Getenv("BASE_URL"), "part of shorten link")
	filePath := flag.String("f", os.Getenv("FILE_STORAGE_PATH"), "path for storage file")
	flag.Parse()
	envs := env{*address, *baseURL, *filePath}

	if len(envs.addressToServe) == 0 {
		envs.addressToServe = "localhost:8080"
	}

	if len(envs.fileStoragePath) == 0 {
		envs.fileStoragePath = "storage.txt"
	}

	if len(envs.baseURL) == 0 {
		envs.baseURL = "http://localhost:8080/"
	}

	store, err := storage.NewStorage(envs.fileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	service := shortener.NewShortenService(store)
	shortenerHandler := handler.NewShortenerHandler(service, envs.baseURL)

	r := chi.NewRouter()
	r.Get("/{ID}", shortenerHandler.LookUpOriginalLinkGET)
	r.Post("/", shortenerHandler.GenerateShorterLinkPOST)
	r.Post("/api/shorten", shortenerHandler.GenerateShorterLinkPOSTJson)

	err = http.ListenAndServe(envs.addressToServe, r)
	if err != nil {
		log.Fatal(err)
	}
}
