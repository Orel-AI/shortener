package shortener

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Orel-AI/shortener.git/storage"
	"log"
	"math/big"
	"net/url"
)

type ShortenService struct {
	Storage storage.Storage
}

type Job struct {
	BaseURL string
	UserID  string
}

func NewShortenService(storage storage.Storage) *ShortenService {
	return &ShortenService{storage}
}

func (s *ShortenService) GetShortLink(link string, userID string, ctx context.Context) (string, bool, error) {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		return "", false, errors.New(link + " is not correct URL")
	}

	encodedString := GenerateShortLink(link, ctx)

	value, err := s.Storage.FindRecord(encodedString, ctx)
	if !errors.Is(err, storage.ErrRecordIsDeleted) && err != nil {
		log.Fatal(err)
	}
	if value == link {
		log.Println("I have found short link for this url : " + value)
		return encodedString, true, nil
	} else {
		s.Storage.AddRecord(encodedString, link, userID, ctx)
		return encodedString, false, nil
	}
}

func (s *ShortenService) GetOriginalLink(linkID string, ctx context.Context) (string, error) {
	value, err := s.Storage.FindRecord(linkID, ctx)
	if err != nil {
		return "", err
	}
	if value != "" {
		return value, nil
	}

	return "", errors.New("no link with such LinkId")
}

func (s *ShortenService) GetUsersLinks(UserID string, baseURL string, ctx context.Context) (map[string]string, error) {
	res := s.Storage.FindAllUsersRecords(UserID, baseURL, ctx)
	if len(res) != 0 {
		return res, nil
	}

	return res, errors.New("no records with such UserID")
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

func (s *ShortenService) ChangeFlagToDeleteWorkPool(urlIDs []string, userID string) {

	jobCh := make(chan *Job)

	for i := 1; i <= len(urlIDs); i++ {
		go func() {
			for job := range jobCh {
				s.Storage.SetDeleteFlag(job.BaseURL, job.UserID)
			}
		}()
	}

	for _, element := range urlIDs {
		job := &Job{BaseURL: element, UserID: userID}
		jobCh <- job
	}
}
