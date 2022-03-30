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

type Storage interface {
	PingDB(ctx context.Context) error
	FindRecord(key string, ctx context.Context) (res string)
	AddRecord(key string, data string, ctx context.Context)
	FindAllUsersRecords(key string, baseURL string, ctx context.Context) map[string]string
	FindRecordWithUserID(key string, ctx context.Context) (res string)
}
type Dict struct {
	file     *os.File
	writer   *bufio.Writer
	fileName string
}
type DatabaseInstance struct {
	conn       *pgx.Conn
	connConfig string
}

func NewStorage(filename string, dsnString string) (Storage, error) {

	if len(dsnString) > 0 {
		storage, _ := newDatabaseConnection(dsnString)
		storage.checkExist()
		return storage, nil
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	return &Dict{
		file:     file,
		writer:   bufio.NewWriter(file),
		fileName: filename,
	}, nil
}

func newDatabaseConnection(dsn string) (*DatabaseInstance, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())
	log.Println("DB Connected!")
	return &DatabaseInstance{
		conn:       conn,
		connConfig: dsn,
	}, nil
}

func (s *Dict) AddRecord(key string, data string, ctx context.Context) {
	userId := ctx.Value("UserID").(uint64)
	userIdStr := strconv.FormatUint(userId, 10)

	s.file.Write([]byte(key + "|" + data + "|" + userIdStr + "\n"))
	s.file.Sync()
}

func (d *Dict) FindRecord(key string, ctx context.Context) (res string) {
	fileToRead, err := os.OpenFile(d.fileName, os.O_RDONLY, 0777)
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

func (d *Dict) FindRecordWithUserID(key string, ctx context.Context) (res string) {
	fileToRead, err := os.OpenFile(d.fileName, os.O_RDONLY, 0777)
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

func (d *Dict) FindAllUsersRecords(key string, baseURL string, ctx context.Context) map[string]string {
	fileToRead, err := os.OpenFile(d.fileName, os.O_RDONLY, 0777)
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

func (d *Dict) PingDB(ctx context.Context) error {
	return errors.New("there is no BD connect")
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

func (db *DatabaseInstance) PingDB(ctx context.Context) error {
	if db.connConfig == "" {
		return errors.New("there is no BD connect")
	}
	conn, err := db.reconnect()
	defer conn.Close(ctx)
	return err
}
func (db *DatabaseInstance) FindRecord(key string, ctx context.Context) (res string) {
	conn, err := db.reconnect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)
	var result string

	err = conn.QueryRow(ctx, "SELECT original_url FROM shortener.shortener "+
		"WHERE short_url  = $1;", key).Scan(&result)
	return result
}

func (db *DatabaseInstance) FindRecordWithUserID(key string, ctx context.Context) (res string) {
	conn, err := db.reconnect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)
	var result string
	userId := ctx.Value("UserID").(uint64)
	userIdStr := strconv.FormatUint(userId, 10)

	conn.QueryRow(ctx, "SELECT original_url FROM shortener.shortener "+
		"WHERE short_url  = $1 and user_id = $2;", key, userIdStr).Scan(&result)
	return result

}

func (db *DatabaseInstance) AddRecord(key string, data string, ctx context.Context) {
	userId := ctx.Value("UserID").(uint64)
	userIdStr := strconv.FormatUint(userId, 10)

	conn, err := db.reconnect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "INSERT INTO shortener.shortener "+
		"(original_url, short_url, user_id) VALUES ($1, $2, $3);", data, key, userIdStr)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (db *DatabaseInstance) FindAllUsersRecords(key string, baseURL string, ctx context.Context) map[string]string {
	userId := ctx.Value("UserID").(uint64)
	userIdStr := strconv.FormatUint(userId, 10)
	results := make(map[string]string)
	var originalURL string
	var trimShorten string

	conn, err := db.reconnect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT short_url, original_url "+
		"FROM shortener.shortener WHERE user_id = $1", userIdStr)

	for rows.Next() {
		err := rows.Scan(&trimShorten, &originalURL)
		if err != nil {
			continue
		}
		results[originalURL] = baseURL + "/" + trimShorten
	}
	return results
}

func (db *DatabaseInstance) checkExist() {
	conn, err := db.reconnect()
	if err != nil {
		log.Fatal(err)
	}
	var cnt int
	defer conn.Close(context.Background())
	_, err = conn.Exec(context.Background(), "CREATE SCHEMA IF NOT EXISTS shortener AUTHORIZATION postgres;")
	if err != nil {
		log.Fatal(err)
	}
	err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM shortener.shortener;").Scan(&cnt)
	if err != nil {
		_, err = conn.Exec(context.Background(), "CREATE TABLE shortener.shortener (user_id VARCHAR(256),"+
			" short_url VARCHAR(256), original_url VARCHAR(256) PRIMARY KEY );")
		if err != nil {
			log.Fatal(err)
		}
	}
}
