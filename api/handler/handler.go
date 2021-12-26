package handler

import (
	"github.com/Orel-AI/shortener.git/service/shortener"
	"io"
	"net/http"
	"strings"
)

type ShortenerHandler struct {
	Request []byte
}

func (h ShortenerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	switch r.Method {

	case "POST":
		if path == "" {
			body, err := io.ReadAll(r.Body)

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			result, err := shortener.GetShortLink(string(body))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusCreated)
			_, err = w.Write([]byte(result))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			return
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

	case "GET":
		originalLink, err := shortener.GetOriginalLink(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Add("Location", originalLink)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return

	default:
		return
	}
}
