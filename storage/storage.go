package storage

import (
	"bufio"
	context "context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Orel-AI/shortener.git/config"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"strings"
)

type Storage interface {
	PingDB(ctx context.Context) error
	FindRecord(key string, ctx context.Context) (res string, err error)
	AddRecord(key string, data string, userID string, ctx context.Context)
	FindAllUsersRecords(key string, baseURL string, ctx context.Context) map[string]string
	FindRecordWithUserID(key string, userID string, ctx context.Context) (res string)
	SetDeleteFlag(baseURL string, userID string)
}
type Dict struct {
	file     *os.File
	writer   *bufio.Writer
	fileName string
}
type DatabaseInstance struct {
	conn       *pgxpool.Pool
	connConfig string
	db         *sql.DB
}

var (
	ErrRecordIsDeleted = errors.New("record is deleted")
)

func NewStorage(env config.Env) (Storage, error) {

	if len(env.DSNString) > 0 {
		storage, err := newDatabaseConnection(env.DSNString)
		if err != nil {
			log.Fatal(err)
		}
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
	conn, err := pgxpool.Connect(context.Background(), dsn)
	//conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}

	//defer conn.Close(context.Background())
	log.Println("DB Connected!")
	return &DatabaseInstance{
		conn:       conn,
		connConfig: dsn,
	}, nil
}

func (d *Dict) AddRecord(key string, data string, userID string, ctx context.Context) {

	d.file.Write([]byte(key + "|" + data + "|" + userID + "\n"))
	d.file.Sync()
}

func (d *Dict) FindRecord(key string, ctx context.Context) (res string, err error) {
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
			return line, nil
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return "", nil
}

func (d *Dict) FindRecordWithUserID(key string, userID string, ctx context.Context) (res string) {
	fileToRead, err := os.OpenFile(d.fileName, os.O_RDONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer fileToRead.Close()

	scanner := bufio.NewScanner(fileToRead)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), key) {
			if strings.Contains(scanner.Text(), userID) {
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
	err := db.conn.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (db *DatabaseInstance) FindRecord(key string, ctx context.Context) (res string, err error) {

	var result string
	var count int
	err = db.conn.QueryRow(ctx, "SELECT count(original_url) FROM shortener.shortener "+
		"WHERE short_url  = $1 and deleted = '1'", key).Scan(&count)
	if err != nil {
		return "", err
	}
	if count == 1 {
		err := ErrRecordIsDeleted
		return "", err
	}

	err = db.conn.QueryRow(ctx, "SELECT original_url FROM shortener.shortener "+
		"WHERE short_url  = $1;", key).Scan(&result)
	if err != pgx.ErrNoRows && err != nil {
		log.Fatal(err)
	}
	return result, nil
}

func (db *DatabaseInstance) FindRecordWithUserID(key string, userID string, ctx context.Context) (res string) {

	var result string

	err := db.conn.QueryRow(ctx, "SELECT original_url FROM shortener.shortener "+
		"WHERE short_url  = $1 and user_id = $2;", key, userID).Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result

}

func (db *DatabaseInstance) AddRecord(key string, data string, userID string, ctx context.Context) {
	_, err := db.conn.Exec(ctx, "INSERT INTO shortener.shortener "+
		"(original_url, short_url, user_id) VALUES ($1, $2, $3);", data, key, userID)
	if err != nil {
		log.Fatal(err)
	}
}

func (db *DatabaseInstance) FindAllUsersRecords(key string, baseURL string, ctx context.Context) map[string]string {

	results := make(map[string]string)
	var originalURL string
	var trimShorten string

	rows, _ := db.conn.Query(ctx, "SELECT short_url, original_url "+
		"FROM shortener.shortener WHERE user_id = $1", key)

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
	var cnt int

	_, err := db.conn.Exec(context.Background(), "CREATE SCHEMA IF NOT EXISTS shortener AUTHORIZATION postgres;")
	if err != nil {
		log.Fatal(err)
	}
	err = db.conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM shortener.shortener;").Scan(&cnt)
	if err != nil {
		_, err = db.conn.Exec(context.Background(), "CREATE TABLE shortener.shortener (user_id VARCHAR(256),"+
			" short_url VARCHAR(256), original_url VARCHAR(256) PRIMARY KEY, DELETED VARCHAR(1) );")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (db *DatabaseInstance) SetDeleteFlag(baseURL string, userID string) {

	ctx := context.Background()

	_, err := db.conn.Exec(ctx, "UPDATE shortener.shortener set deleted = '1'"+
		" where short_url = $1 and user_id = $2;", baseURL, userID)
	if err != nil {
		log.Fatal(err)
	}
}

func (d *Dict) SetDeleteFlag(baseURL string, userID string) {
}
