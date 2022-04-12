package storage

import (
	"bufio"
	context "context"
	"errors"
	"fmt"
	"github.com/Orel-AI/shortener.git/config"
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
	SetDeleteFlag(baseURL []string, UserId string)
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

func NewStorage(env config.Env) (Storage, error) {

	if len(env.DSNString) > 0 {
		storage, _ := newDatabaseConnection(env.DSNString)
		storage.checkExist()
		return storage, nil
	}

	file, err := os.OpenFile(env.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	return &Dict{
		file:     file,
		writer:   bufio.NewWriter(file),
		fileName: env.FileStoragePath,
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

func (d *Dict) AddRecord(key string, data string, ctx context.Context) {
	userID := ctx.Value("UserID").(uint64)
	userIDStr := strconv.FormatUint(userID, 10)

	d.file.Write([]byte(key + "|" + data + "|" + userIDStr + "\n"))
	d.file.Sync()
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

	userID := ctx.Value("UserID").(uint64)
	UserIDStr := strconv.FormatUint(userID, 10)

	scanner := bufio.NewScanner(fileToRead)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), key) {
			if strings.Contains(scanner.Text(), UserIDStr) {
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
	if err != nil {
		log.Fatal(err)
	}

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
	var count int
	err = conn.QueryRow(ctx, "SELECT count(original_url) FROM shortener.shortener "+
		"WHERE short_url  = $1 and deleted = '1'", key).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	if count == 1 {
		return "deleted"
	}

	err = conn.QueryRow(ctx, "SELECT original_url FROM shortener.shortener "+
		"WHERE short_url  = $1;", key).Scan(&result)
	if err != pgx.ErrNoRows && err != nil {
		log.Fatal(err)
	}
	return result
}

func (db *DatabaseInstance) FindRecordWithUserID(key string, ctx context.Context) (res string) {
	conn, err := db.reconnect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)
	var result string
	userID := ctx.Value("UserID").(uint64)
	userIDStr := strconv.FormatUint(userID, 10)

	err = conn.QueryRow(ctx, "SELECT original_url FROM shortener.shortener "+
		"WHERE short_url  = $1 and user_id = $2;", key, userIDStr).Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result

}

func (db *DatabaseInstance) AddRecord(key string, data string, ctx context.Context) {
	userID := ctx.Value("UserID").(uint64)
	userIDStr := strconv.FormatUint(userID, 10)

	conn, err := db.reconnect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "INSERT INTO shortener.shortener "+
		"(original_url, short_url, user_id) VALUES ($1, $2, $3);", data, key, userIDStr)
	if err != nil {
		log.Fatal(err)
	}
}

func (db *DatabaseInstance) FindAllUsersRecords(key string, baseURL string, ctx context.Context) map[string]string {
	userID := ctx.Value("UserID").(uint64)
	userIDStr := strconv.FormatUint(userID, 10)
	results := make(map[string]string)
	var originalURL string
	var trimShorten string

	conn, err := db.reconnect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	rows, _ := conn.Query(ctx, "SELECT short_url, original_url "+
		"FROM shortener.shortener WHERE user_id = $1", userIDStr)

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
			" short_url VARCHAR(256), original_url VARCHAR(256) PRIMARY KEY, DELETED VARCHAR(1) );")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (db *DatabaseInstance) SetDeleteFlag(baseURLs []string, userID string) {

	conn, err := db.reconnect()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	_, err = conn.Exec(ctx, "UPDATE shortener.shortener set deleted = '1'"+
		" where short_url = ANY($1) and user_id = $2;", baseURLs, userID)
	if err != nil {
		log.Fatal(err)
	}
}

func (d *Dict) SetDeleteFlag(baseURLs []string, UserId string) {
}
