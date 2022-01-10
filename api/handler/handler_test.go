package handler

import (
	"github.com/Orel-AI/shortener.git/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
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
			name:   "Fail POST shortener test",
			target: "/",
			method: http.MethodPost,
			want: want{
				code:         400,
				requestLink:  "invalidlink",
				responseLink: "invalidlink is not correct URL\n",
			},
		},
		{
			name:   "Success GET shortener test",
			target: "/MTA0OTY4",
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
			request := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.want.requestLink))
			w := httptest.NewRecorder()
			ShortenerHandler{}.ServeHTTP(w, request)
			r := w.Result()
			_ = r.Close
			if tt.method == http.MethodPost {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.want.code, r.StatusCode)
				assert.Equal(t, tt.want.responseLink, string(body))
			}
			if tt.method == http.MethodGet {
				assert.Equal(t, tt.want.code, r.StatusCode)
				assert.Equal(t, tt.want.responseLink, r.Header.Get("Location"))
			}
		})
	}
}
