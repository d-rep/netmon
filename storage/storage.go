package storage

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/jmoiron/sqlx"
)

const databaseFilePath = "netmon.db"

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

CREATE VIEW IF NOT EXISTS summary AS
select
	date(created_at) as dt,
	substr('SunMonTueWedThuFriSat',strftime('%w',created_at)*3+1,3) as day_of_week,
	count(*) as total, 
	SUM(CASE success WHEN TRUE THEN 1 ELSE 0 END) AS successes,
	SUM(CASE success WHEN TRUE THEN 0 ELSE 1 END) AS failures,
	printf("%.2f", avg(duration_ms)) as average_duration_ms,
	printf("%.2f", max(duration_ms)) as worst_duration_ms
from call
group by date(created_at)
order by date(created_at) desc;
`

// model to keep history in DB
type Call struct {
	ID         uint      `json:"id" db:"id"`
	URL        string    `json:"url" db:"url"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	Status     int       `json:"status" db:"status"`   // raw HTTP status code
	Success    bool      `json:"success" db:"success"` // was HTTP call successful?
	Error      string    `json:"error" db:"error"`
	DurationMS float64   `json:"durationMs" db:"duration_ms"`
}

func (c *Call) String() string {
	return fmt.Sprintf("Call{ID:%d, URL:%s, CreatedAt:%s, Status:%d, Success:%t, DurationMS:%6.3f, Error:`%s`}", c.ID, c.URL, c.CreatedAt, c.Status, c.Success, c.DurationMS, c.Error)
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
	db, err := sqlx.Connect(driverName, dsn)
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

func GetDatabaseAndMigrate() (*Storage, error) {
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
(:url, :created_at, :status, :success, :error, :duration_ms)
;
`

func (db *Storage) SaveCall(call *Call) error {
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

const selectCall = `
select
	id,
	url,
	created_at,
	status,
	success,
	error,
	duration_ms
from call
order by created_at desc
limit %d;
`

func (db *Storage) GetRecentCalls(count uint8) ([]*Call, error) {
	var calls []*Call
	sqlRecent := fmt.Sprintf(selectCall, count)
	err := db.DB.Select(&calls, sqlRecent)
	if err != nil {
		return nil, fmt.Errorf("failure with GetRecentCalls: %w", err)
	}
	return calls, nil
}
