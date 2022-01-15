package handler

import (
	"context"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strings"
)

type ShortenerHandler struct {
	shortener *shortener.ShortenService
}

func NewShortenerHandler(s *shortener.ShortenService) *ShortenerHandler {
	return &ShortenerHandler{s}
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

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(result))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
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
