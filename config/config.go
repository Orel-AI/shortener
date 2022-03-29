package config

import (
	"flag"
	"os"
)

type Env struct {
	AddressToServe  string
	BaseURL         string
	FileStoragePath string
	DSNString       string
}

func NewConfig() Env {
	address := flag.String("a", os.Getenv("SERVER_ADDRESS"), "address to start up server")
	baseURL := flag.String("b", os.Getenv("BASE_URL"), "part of shorten link")
	filePath := flag.String("f", os.Getenv("FILE_STORAGE_PATH"), "path for storage file")
	dsnString := flag.String("d", os.Getenv("DATABASE_DSN"), "dsn to connect PostgreSQL")
	flag.Parse()
	envs := Env{*address, *baseURL, *filePath, *dsnString}
	if len(envs.AddressToServe) == 0 {
		envs.AddressToServe = "localhost:8080"
	}

	if len(envs.FileStoragePath) == 0 {
		envs.FileStoragePath = "storage.txt"
	}

	if len(envs.BaseURL) == 0 {
		envs.BaseURL = "http://localhost:8080"
	}

	return envs
}
