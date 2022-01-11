package handler

import (
	"context"
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/stretchr/testify/assert"
	"io"
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
	}

	storage.Initialize()

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

			if tt.method == http.MethodPost {
				GenerateShorterLinkPOST(response, request)

				body, err := io.ReadAll(response.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, tt.want.code, response.Code)
				assert.Equal(t, tt.want.responseLink, string(body))
			}
			if tt.method == http.MethodGet {
				storage.AddRecord(strings.TrimPrefix(request.URL.Path, "/"), tt.want.responseLink, context.Background())
				LookUpOriginalLinkGET(response, request)
				result := response.Result()
				defer result.Body.Close()

				assert.Equal(t, tt.want.code, result.StatusCode)
				assert.Equal(t, tt.want.responseLink, result.Header.Get("Location"))
			}
		})
	}
}
