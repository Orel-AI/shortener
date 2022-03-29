package handler

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strings"
)

type ShortenerHandler struct {
	shortener *shortener.ShortenService
	baseURL   string
}
type RequestBody struct {
	URL string `json:"url"`
}

type ResponseBody struct {
	Result string `json:"result"`
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
func (h *ShortenerHandler) GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = gzr
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gzw, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer gzw.Close()
		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzw}, r)
	})
}

func NewShortenerHandler(s *shortener.ShortenService, b string) *ShortenerHandler {
	return &ShortenerHandler{s, b}
}

func (h *ShortenerHandler) GenerateShorterLinkPOSTJson(w http.ResponseWriter, r *http.Request) {
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
	result = fmt.Sprintf("%v/%v", h.baseURL, result)

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
	result = fmt.Sprintf("%v/%v", h.baseURL, result)

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
