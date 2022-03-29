package storage

import (
	"bufio"
	"context"
	"log"
	"os"
	"strconv"
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
	userId := ctx.Value("UserID").(uint64)
	userIdStr := strconv.FormatUint(userId, 10)

	s.file.Write([]byte(key + "|" + data + "|" + userIdStr + "\n"))
	s.file.Sync()
}

func (s *Storage) FindRecord(key string, ctx context.Context) (res string) {
	fileToRead, err := os.OpenFile(s.fileName, os.O_RDONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer fileToRead.Close()

	scanner := bufio.NewScanner(fileToRead)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), key) {
			line := scanner.Text()
			line = line[strings.Index(line, "|")+1 : strings.LastIndex(line, "|")]
			line = strings.ReplaceAll(line, "\n", "")
			return line
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ""
}

func (s *Storage) FindRecordWithUserID(key string, ctx context.Context) (res string) {
	fileToRead, err := os.OpenFile(s.fileName, os.O_RDONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer fileToRead.Close()

	userId := ctx.Value("UserID").(uint64)
	UserID := strconv.FormatUint(userId, 10)

	scanner := bufio.NewScanner(fileToRead)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), key) {
			if strings.Contains(scanner.Text(), UserID) {
				line := scanner.Text()
				line = line[strings.Index(line, "|")+1 : strings.LastIndex(line, "|")]
				line = strings.ReplaceAll(line, "\n", "")
				return line
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ""
}

func (s *Storage) FindAllUsersRecords(key string, baseURL string, ctx context.Context) map[string]string {
	fileToRead, err := os.OpenFile(s.fileName, os.O_RDONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer fileToRead.Close()

	scanner := bufio.NewScanner(fileToRead)
	results := make(map[string]string)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), key) {
			line := scanner.Text()
			results[line[strings.Index(line, "|")+1:strings.LastIndex(line, "|")]] =
				baseURL + line[:strings.Index(line, "|")]
		}
	}
	return results
}
