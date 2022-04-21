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
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

type key int

const (
	keyPrincipalID key = iota
)

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
func GzipMiddleware(next http.Handler) http.Handler {
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

func (h *ShortenerHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(h.cookieName)
		log.Println("Cookie found by name: ", cookie)
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := h.decodeCookie(cookie)
		if err != nil {
			cookie, id, err = h.generateCookie()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.SetCookie(w, cookie)
		}
		log.Println("UserID: ", id)
		ctx := context.WithValue(r.Context(), keyPrincipalID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (h *ShortenerHandler) decodeCookie(cookie *http.Cookie) (uint64, error) {
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

func (h *ShortenerHandler) generateCookie() (*http.Cookie, uint64, error) {
	id := make([]byte, 8)

	_, err := rand.Read(id)
	if err != nil {
		return nil, 0, err
	}

	hm := hmac.New(sha256.New, []byte(h.secretString))
	hm.Write(id)
	sign := hex.EncodeToString(append(id, hm.Sum(nil)...))

	return &http.Cookie{
			Name:   h.cookieName,
			Value:  sign,
			Path:   "/",
			Secure: false,
		},
		binary.BigEndian.Uint64(id),
		nil
}

func NewShortenerHandler(s *shortener.ShortenService, b string, secretString string, cookieName string) *ShortenerHandler {
	return &ShortenerHandler{s, b, secretString, cookieName}
}

func (h *ShortenerHandler) GenerateShorterLinkPOSTJson(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := ctx.Value(keyPrincipalID).(uint64)
	userIDStr := strconv.FormatUint(userID, 10)

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
	result, isExisted, err := h.Shortener.GetShortLink(reqBody.URL, userIDStr, ctx)
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
	if isExisted {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	_, err = w.Write([]byte(resJSON))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}

func (h *ShortenerHandler) GenerateShorterLinkPOST(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	userID := ctx.Value(keyPrincipalID).(uint64)
	userIDStr := strconv.FormatUint(userID, 10)

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
	result, isExisted, err := h.Shortener.GetShortLink(string(body), userIDStr, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result = fmt.Sprintf("%v/%v", h.baseURL, result)

	w.Header().Set("Content-Type", "text/plain")
	if isExisted {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	_, err = w.Write([]byte(result))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *ShortenerHandler) LookUpOriginalLinkGET(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//userID := ctx.Value(keyPrincipalID).(uint64)
	//userIDStr := strconv.FormatUint(userID, 10)

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
	if originalLink == "deleted" {
		w.WriteHeader(http.StatusGone)
		return
	}
	w.Header().Add("Location", originalLink)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *ShortenerHandler) LookUpUsersRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(keyPrincipalID).(uint64)
	userIDStr := strconv.FormatUint(userID, 10)

	searchResult, err := h.Shortener.GetUsersLinks(userIDStr, h.baseURL, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	var response []MapOriginalShorten
	for key, value := range searchResult {
		response = append(response,
			MapOriginalShorten{ShortURL: value, OriginalURL: key})
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
	err := h.Shortener.Storage.PingDB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ShortenerHandler) GenerateShorterLinkPOSTBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := ctx.Value(keyPrincipalID).(uint64)
	userIDStr := strconv.FormatUint(userID, 10)

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

	var reqBody []BatchRequest
	var resBody []BatchResponse

	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(reqBody); i++ {
		result, _, err := h.Shortener.GetShortLink(reqBody[i].OriginalURL, userIDStr, ctx)
		if err != nil {
			continue
		}
		result = fmt.Sprintf("%v/%v", h.baseURL, result)
		resBody = append(resBody, BatchResponse{
			CorrelationID: reqBody[i].CorrelationID,
			ShortURL:      result,
		})

	}

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

func (h *ShortenerHandler) BatchDeleteLinks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := ctx.Value(keyPrincipalID).(uint64)
	userIDStr := strconv.FormatUint(userID, 10)

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bodyStr := string(body)
	bodyStr = strings.ReplaceAll(bodyStr, "[", "")
	bodyStr = strings.ReplaceAll(bodyStr, "]", "")
	bodyStr = strings.ReplaceAll(bodyStr, "\"", "")
	linksToDelete := strings.Split(bodyStr, ",")
	for _, link := range linksToDelete {
		log.Println(link)
	}

	go func() {
		h.Shortener.Storage.SetDeleteFlag(linksToDelete, userIDStr)
	}()
	log.Println(linksToDelete)

	w.WriteHeader(http.StatusAccepted)
}
