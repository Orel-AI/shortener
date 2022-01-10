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

func GetShortLink(link string, ctx context.Context) (string, error) {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		return "", errors.New(link + " is not correct URL")
	}

	encodedString := GenerateShortLink(link, ctx)

	value := storage.FindRecord(encodedString, ctx)
	if value == link {
		return encodedString, nil
	} else if value != "" && value != link {
		return "", errors.New("Shortener for link: " + link + " is already in DB")
	} else {
		storage.AddRecord(encodedString, link, ctx)
		return encodedString, nil
	}
}

func GetOriginalLink(linkID string, ctx context.Context) (string, error) {
	value := storage.FindRecord(linkID, ctx)
	if value != "" {
		return value, nil
	} else {
		return "", errors.New("no link with such LinkId")
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
