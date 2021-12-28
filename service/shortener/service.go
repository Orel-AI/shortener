package shortener

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/url"
)

var linkMap map[string]string

func InitializeMap() {
	linkMap = make(map[string]string)
}

func GetShortLink(link string) (string, error) {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		return "", errors.New(link + " is not correct URL")
	}

	encodedString := GenerateShortLink(link)

	value, found := linkMap[encodedString]
	if found && value == link {
		return "http://localhost:8080/" + encodedString, nil
	} else if found && value != link {
		return "", errors.New("shortener for link: " + link + " is already in DB")
	} else {
		linkMap[encodedString] = link
		return "http://localhost:8080/" + encodedString, nil
	}
}

func GetOriginalLink(linkID string) (string, error) {
	value, found := linkMap[linkID]
	if found {
		return value, nil
	} else {
		return "", errors.New("no link with such LinkID")
	}
}

func sha256Of(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
}

func GenerateShortLink(initialLink string) string {
	urlHashBytes := sha256Of(initialLink)
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	finalString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", generatedNumber)))
	return finalString[:8]
}
