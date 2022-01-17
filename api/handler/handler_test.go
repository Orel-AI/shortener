package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Orel-AI/shortener.git/service/shortener"
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestShortenerHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code         int
		requestLink  string
		responseLink string
	}
	tests := []struct {
		method string
		target string
		name   string
		want   want
	}{
		{
			name:   "Success POST shortener test",
			target: "/",
			method: http.MethodPost,
			want: want{
				code:         201,
				requestLink:  "https://ya.ru",
				responseLink: "http://localhost:8080/MTA0OTY4",
			},
		},
		{
			name:   "Fail POST shortener twest",
			target: "http://localhost:8080",
			method: http.MethodPost,
			want: want{
				code:         400,
				requestLink:  "invalidlink",
				responseLink: "invalidlink is not correct URL\n",
			},
		},
		{
			name:   "Success GET shortener test",
			target: "http://localhost:8080/MTA0OTY4",
			method: http.MethodGet,
			want: want{
				code:         307,
				responseLink: "https://ya.ru",
			},
		},
		{
			name:   "Success POST Json shortener test",
			target: "http://localhost:8080/api/shorten",
			method: http.MethodPost,
			want: want{
				code:         201,
				requestLink:  "https://ya.ru",
				responseLink: "http://localhost:8080/MTA0OTY4",
			},
		},
	}

	store := storage.NewStorage()
	service := shortener.NewShortenService(store)
	shortenerHandler := NewShortenerHandler(service)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.method, tt.target, strings.NewReader(tt.want.requestLink))
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			request.URL, err = url.Parse(tt.target)
			if err != nil {
				t.Fatal(err)
			}
			if tt.method == http.MethodPost && tt.target == "http://localhost:8080/api/shorten" {
				type RequestBody struct {
					URL string `json:"url"`
				}
				type ResponseBody struct {
					Result string `json:"result"`
				}
				reqBody := RequestBody{URL: tt.want.requestLink}
				requestBody, err := json.Marshal(reqBody)
				if err != nil {
					log.Fatal(err)
				}

				request, err = http.NewRequest(tt.method, tt.target, bytes.NewBuffer(requestBody))
				if err != nil {
					log.Fatal(err)
				}

				request.Header.Add("Content-Type", "application/json")

				shortenerHandler.GenerateShorterLinkPOSTJson(response, request)

				body, err := io.ReadAll(response.Body)
				if err != nil {
					t.Fatal(err)
				}

				resBody := ResponseBody{}
				err = json.Unmarshal(body, &resBody)
				if err != nil {
					log.Fatal(err)
				}

				assert.Equal(t, tt.want.code, response.Code)
				assert.Equal(t, tt.want.responseLink, resBody.Result)

			} else if tt.method == http.MethodPost {
				shortenerHandler.GenerateShorterLinkPOST(response, request)

				body, err := io.ReadAll(response.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, tt.want.code, response.Code)
				assert.Equal(t, tt.want.responseLink, string(body))
			}
			if tt.method == http.MethodGet {
				store.AddRecord(strings.TrimPrefix(request.URL.Path, "/"), tt.want.responseLink, context.Background())
				shortenerHandler.LookUpOriginalLinkGET(response, request)
				result := response.Result()
				defer result.Body.Close()

				assert.Equal(t, tt.want.code, result.StatusCode)
				assert.Equal(t, tt.want.responseLink, result.Header.Get("Location"))
			}
		})
	}
}
