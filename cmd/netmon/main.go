package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	exitFail         = 1
	databaseFilePath = "netmon.db"
)

const schema = `
CREATE TABLE IF NOT EXISTS call (
  id integer PRIMARY KEY,
  created_at datetime DEFAULT current_timestamp,
  url text,
  status integer,
  success boolean,
  error text
);
`

// model to keep history in DB
type Call struct {
	ID        uint      `json:"id" db:"id"`
	URL       string    `json:"url" db:"url"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	Status    uint      `json:"status" db:"status"`
	Success   bool      `json:"success" db:"success"`
	Error     string    `json:"error" db:"error"`
}

func isUrlUp(url string) (int, error) {
	resp, err := http.Head(url)
	if err != nil {
		// happens on "connection refused"
		return 0, err
	}
	// a HEAD should not have a response body to close
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("could not close response body: %s", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		statusText := http.StatusText(resp.StatusCode)
		// a HEAD should not have a response body to read (content always empty)
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, fmt.Errorf("could not call URL \"%s\", got status %d: %s, and error reading response body: %w", url, resp.StatusCode, statusText, err)
		}
		return resp.StatusCode, fmt.Errorf("could not call URL \"%s\", got status %d: %s, content: \"%s\"", url, resp.StatusCode, statusText, content)
	}
	return resp.StatusCode, nil
}

var urls = []string{
	"https://www.cloudflare.com/",
	"https://www.google.com/",
	"https://www.amazon.com/",
	"https://www.fastly.com/",
}

type Storage struct {
	DB *sqlx.DB
}

func getDatabase() (*Storage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return &Storage{}, err
	}
	dsn := path.Join(home, databaseFilePath)
	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return &Storage{}, err
	}
	return &Storage{
		DB: db,
	}, nil
}

func (db *Storage) applyMigrations() error {
	_, err := db.DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("could not apply schema migrations to database: %w", err)
	}
	return nil
}

const sqlInsert = `
INSERT INTO call
(url, created_at, status, success, error)
VALUES
(:url, :created_at, :status, :success, :error)
;
`

func (db *Storage) record(call *Call) error {
	result, err := db.DB.NamedExec(sqlInsert, call)
	if err != nil {
		return fmt.Errorf("could not insert new Call record into database: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("created new Call record but could not get last insert ID: %w", err)
	}
	call.ID = uint(id)
	return nil
}

func run(args []string, _ io.Writer) error {
	db, err := getDatabase()
	if err != nil {
		return err
	}
	err = db.applyMigrations()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		url = flags.String("url", "", "which URL to use when checking if internet connection is working?")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if len(*url) != 0 {
		urls = []string{*url}
	}

	for _, url := range urls {
		statusCode, headErr := isUrlUp(url)
		call := &Call{
			URL:       url,
			Status:    uint(statusCode),
			Success:   headErr == nil,
			CreatedAt: time.Now(),
		}
		if !call.Success {
			call.Error = headErr.Error()
		}
		err = db.record(call)
		if err != nil {
			return err
		}
		if headErr != nil {
			fmt.Printf("%s is down! Status %d, %s\n", url, statusCode, headErr)
			continue
		}
		fmt.Printf("%s is up\n", url)
	}
	return nil
}

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}
