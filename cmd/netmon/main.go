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
  error text,
  duration_ms decimal(10,3)
);
`

// model to keep history in DB
type Call struct {
	ID         uint      `json:"id" db:"id"`
	URL        string    `json:"url" db:"url"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	Status     int       `json:"status" db:"status"`   // raw HTTP status code
	Success    bool      `json:"success" db:"success"` // was HTTP call successful?
	Error      string    `json:"error" db:"error"`
	DurationMS float64   `json:"duration" db:"duration"`
}

func (c *Call) String() string {
	return fmt.Sprintf("Call{ID:%d, URL:%s, CreatedAt:%s, Status:%d, Success:%t, Error:`%s`}", c.ID, c.URL, c.CreatedAt, c.Status, c.Success, c.Error)
}

func isUrlUp(url string) *Call {
	call := &Call{
		URL:       url,
		Success:   false,
		CreatedAt: time.Now(),
	}
	start := time.Now()
	resp, err := http.Head(url)
	call.DurationMS = getMillisecondsSince(start)
	if err != nil {
		// happens on "connection refused"
		call.Error = err.Error()
		return call
	}
	// a HEAD should not have a response body to close
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("could not close response body: %s", err)
		}
	}()

	call.Status = resp.StatusCode
	call.Success = resp.StatusCode == http.StatusOK
	if !call.Success {
		statusText := http.StatusText(resp.StatusCode)
		// a HEAD should not have a response body to read (content always empty)
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			call.Error = fmt.Errorf("failed reading response body: %w", err).Error()
		} else {
			call.Error = fmt.Errorf("HTTP %s, Content: \"%s\"", statusText, content).Error()
		}
	}
	return call
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

func getDatabaseAndMigrate() (*Storage, error) {
	db, err := getDatabase()
	if err != nil {
		return nil, err
	}
	err = db.applyMigrations()
	if err != nil {
		return nil, err
	}
	return db, nil
}

const sqlInsert = `
INSERT INTO call
(url, created_at, status, success, error, duration_ms)
VALUES
(:url, :created_at, :status, :success, :error, :duration)
;
`

func (db *Storage) saveResult(call *Call) error {
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

func (db *Storage) callAndSaveResult(url string) (*Call, error) {
	call := isUrlUp(url)
	err := db.saveResult(call)
	if err != nil {
		return call, fmt.Errorf("could not save Call result: %v: %w", call, err)
	}
	return call, nil
}

func run(args []string, _ io.Writer) error {
	db, err := getDatabaseAndMigrate()
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
		call, err := db.callAndSaveResult(url)
		if err != nil {
			// failed to save Call, but keep going to display results to user
			fmt.Println(err)
		}
		if !call.Success {
			fmt.Printf("%s is down! %v\n", url, call)
			continue
		}
		fmt.Printf("%s is up\n", url)
	}
	return nil
}

func getMillisecondsSince(start time.Time) float64 {
	duration := time.Since(start)
	// https://stackoverflow.com/a/41503910
	ms := float64(duration) / float64(time.Millisecond)
	return ms
}

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}
