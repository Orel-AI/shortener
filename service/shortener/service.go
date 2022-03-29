package shortener

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Orel-AI/shortener.git/storage"
	"math/big"
	"net/url"
)

type ShortenService struct {
	Storage *storage.Storage
}

func NewShortenService(storage *storage.Storage) *ShortenService {
	return &ShortenService{storage}
}

func (s *ShortenService) GetShortLink(link string, ctx context.Context) (string, error) {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		return "", errors.New(link + " is not correct URL")
	}

	encodedString := GenerateShortLink(link, ctx)

	value := s.Storage.FindRecordWithUserID(encodedString, ctx)
	if value == link {
		return encodedString, nil
	} else {
		s.Storage.AddRecord(encodedString, link, ctx)
		return encodedString, nil
	}
}

func (s *ShortenService) GetOriginalLink(linkID string, ctx context.Context) (string, error) {
	value := s.Storage.FindRecord(linkID, ctx)
	if value != "" {
		return value, nil
	} else {
		return "", errors.New("no link with such LinkId")
	}
}

func (s *ShortenService) GetUsersLinks(UserID string, baseURL string, ctx context.Context) (map[string]string, error) {
	res := s.Storage.FindAllUsersRecords(UserID, baseURL, ctx)
	if len(res) != 0 {
		return res, nil
	} else {
		return res, errors.New("no records with such UserID")
	}
}

func sha256Of(input string, ctx context.Context) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
}

func GenerateShortLink(initialLink string, ctx context.Context) string {
	urlHashBytes := sha256Of(initialLink, ctx)
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	finalString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", generatedNumber)))
	return finalString[:8]
}
