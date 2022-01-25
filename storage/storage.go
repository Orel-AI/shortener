package storage

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
)

type Storage struct {
	file     *os.File
	writer   *bufio.Writer
	fileName string
}

func NewStorage(filename string) (*Storage, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &Storage{
		file:     file,
		writer:   bufio.NewWriter(file),
		fileName: filename,
	}, nil
}

func (s *Storage) AddRecord(key string, data string, ctx context.Context) {
	s.file.Write([]byte(key + "|" + data + "\n"))
	s.file.Sync()
}

func (s *Storage) FindRecord(key string, ctx context.Context) (res string) {
	fileToRead, err := os.OpenFile(s.fileName, os.O_RDONLY, 0777)
	defer fileToRead.Close()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(fileToRead)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), key) {
			line := scanner.Text()
			line = line[strings.Index(line, "|")+1:]
			line = strings.ReplaceAll(line, "\n", "")
			return line
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ""
}
