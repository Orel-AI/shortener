package handler

import (
	"context"
	"encoding/json"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strings"
)

type ShortenerHandler struct {
	shortener *shortener.ShortenService
}

func NewShortenerHandler(s *shortener.ShortenService) *ShortenerHandler {
	return &ShortenerHandler{s}
}

func (h *ShortenerHandler) GenerateShorterLinkPOSTJson(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		URL string `json:"url"`
	}

	type ResponseBody struct {
		Result string `json:"result"`
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		return
	}

	reqBody := RequestBody{}

	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		log.Fatal(err)
	}

	result, err := h.shortener.GetShortLink(reqBody.URL, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result = "http://localhost:8080/" + result

	resBody := ResponseBody{Result: result}
	resJSON, err := json.Marshal(resBody)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(resJSON))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}

func (h *ShortenerHandler) GenerateShorterLinkPOST(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if string(body) == "" {
		http.Error(w, "link not provided in request's body", http.StatusBadRequest)
		return
	}
	result, err := h.shortener.GetShortLink(string(body), ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result = "http://localhost:8080/" + result

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(result))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *ShortenerHandler) LookUpOriginalLinkGET(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ID := chi.URLParam(r, "ID")
	if ID == "" && strings.TrimPrefix(r.URL.Path, "/") == "" {
		http.Error(w, "ID of link is missed", http.StatusBadRequest)
		return
	} else {
		ID = strings.TrimPrefix(r.URL.Path, "/")
	}
	originalLink, err := h.shortener.GetOriginalLink(ID, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", originalLink)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
