package handler

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ShortenerHandler struct {
	Shortener    *shortener.ShortenService
	baseURL      string
	secretString string
	cookieName   string
}

type RequestBody struct {
	URL string `json:"url"`
}

type ResponseBody struct {
	Result string `json:"result"`
}

type MapOriginalShorten struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
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
			gzippedOutput, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = gzippedOutput
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gzipW, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer gzipW.Close()
		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzipW}, r)
	})
}

func (h *ShortenerHandler) AuthMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(h.cookieName)
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := h.decodeAuthCookie(cookie)
		if err != nil {
			cookie, id, err = h.generateAuthCookie()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.SetCookie(w, cookie)
		}
		log.Println(id)
		ctx := context.WithValue(r.Context(), "UserID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (h *ShortenerHandler) decodeAuthCookie(cookie *http.Cookie) (uint64, error) {
	if cookie == nil {
		return 0, http.ErrNoCookie
	}

	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return 0, err
	}

	id := binary.BigEndian.Uint64(data[:8])

	hm := hmac.New(sha256.New, []byte(h.secretString))
	hm.Write(data[:8])
	sign := hm.Sum(nil)
	if hmac.Equal(data[8:], sign) {
		return id, nil
	}
	return 0, http.ErrNoCookie
}

func (h *ShortenerHandler) generateAuthCookie() (*http.Cookie, uint64, error) {
	id := make([]byte, 8)

	_, err := rand.Read(id)
	if err != nil {
		return nil, 0, err
	}

	hm := hmac.New(sha256.New, []byte(h.secretString))
	hm.Write(id)
	sign := hex.EncodeToString(append(id, hm.Sum(nil)...))

	return &http.Cookie{
			Name:  h.cookieName,
			Value: sign,
		},
		binary.BigEndian.Uint64(id),
		nil
}

func NewShortenerHandler(s *shortener.ShortenService, b string, secretString string, cookieName string) *ShortenerHandler {
	return &ShortenerHandler{s, b, secretString, cookieName}
}

func (h *ShortenerHandler) GenerateShorterLinkPOSTJson(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type header is not valid", http.StatusBadRequest)
		return
	}

	reqBody := RequestBody{}

	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(reqBody.URL)
	result, err := h.Shortener.GetShortLink(reqBody.URL, ctx)
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
	ctx := r.Context()
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
	result, err := h.Shortener.GetShortLink(string(body), ctx)
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
	originalLink, err := h.Shortener.GetOriginalLink(ID, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", originalLink)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *ShortenerHandler) LookUpUsersRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := ctx.Value("UserID").(uint64)
	UserID := strconv.FormatUint(userId, 10)

	searchResult, err := h.Shortener.GetUsersLinks(UserID, h.baseURL, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	var response []MapOriginalShorten
	for key, value := range searchResult {
		response = append(response,
			MapOriginalShorten{OriginalURL: key, ShortURL: value})
	}

	resJSON, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(resJSON))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}

func (h *ShortenerHandler) PingDBByRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("ya tut")
	err := h.Shortener.Storage.PingDB(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Println("ya tut 2")
	w.WriteHeader(http.StatusOK)
	return
}
