package storage

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
	"os"
	"strconv"
	"strings"
)

type Storage struct {
	file     *os.File
	writer   *bufio.Writer
	fileName string
	Database DatabaseInstance
}
type DatabaseInstance struct {
	conn       *pgx.Conn
	connConfig string
}

func NewStorage(filename string, dsnString string) (*Storage, error) {
	var database = DatabaseInstance{}

	if len(dsnString) > 0 {
		conn, err := pgx.Connect(context.Background(), dsnString)
		if err != nil {
			log.Fatal(err)
		}
		database = DatabaseInstance{
			conn:       conn,
			connConfig: dsnString,
		}
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	return &Storage{
		file:     file,
		writer:   bufio.NewWriter(file),
		fileName: filename,
		Database: database,
	}, nil
}

func (db *DatabaseInstance) reconnect() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), db.connConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connection to database1: %v", err)
	}

	if err = conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("couldn't ping postgre database: %v", err)
	}

	return conn, err
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
				baseURL + "/" + line[:strings.Index(line, "|")]
		}
	}
	return results
}

func (s *Storage) PingDB(ctx context.Context) error {
	if s.Database.connConfig == "" {
		return errors.New("there is no BD connect")
	}
	conn, err := s.Database.reconnect()
	defer conn.Close(ctx)
	return err
}
