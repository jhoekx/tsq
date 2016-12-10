package tsq

import (
	"bytes"
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type SQLiteStore struct {
	path string
	db   *sql.DB
}

func NewSQLiteStore() JobStore {
	return &SQLiteStore{path: "./tsq.sqlite3"}
}

func CreateJobDB(db *sql.DB) (err error) {
	_, err = db.Exec(`create table Job (
		uuid text not null primary key,
		name text not null,
		status text not null,
		arguments text,
		result text,
		created datetime not null,
		updated datetime not null
	)`)
	return
}

func (s *SQLiteStore) Start() (err error) {
	s.db, err = sql.Open("sqlite3", s.path)
	if err != nil {
		return
	}
	err = s.db.Ping()
	if err != nil {
		return
	}
	migrations := NewMigrations(s.db)
	migrations.Register("V1__001_CreateJobDB", CreateJobDB)
	err = migrations.Run()
	return
}

func (s *SQLiteStore) Stop() {
	s.db.Close()
}

func (s *SQLiteStore) Store(job *Job) (err error) {
	arguments, err := encode(job.Arguments)
	if err != nil {
		return
	}
	result, err := encode(job.Result)
	if err != nil {
		return
	}
	_, err = s.db.Exec(`insert into Job (uuid, name, status, arguments, result, created, updated)
			            values (?, ?, ?, ?, ?, ?, ?)`,
		job.UUID, job.Name, job.Status, toNullString(arguments), toNullString(result), job.Created, job.Updated)
	return
}

func (s *SQLiteStore) GetJobs() (jobs []*Job, err error) {
	rows, err := s.db.Query("select uuid, name, status, arguments, result, created, updated from Job order by created desc")
	if err != nil {
		return
	}
	defer rows.Close()

	jobs = make([]*Job, 0)
	for rows.Next() {
		job, readErr := readJob(rows)
		if readErr != nil {
			return nil, readErr
		}
		jobs = append(jobs, &job)
	}
	return
}

func (s *SQLiteStore) GetJob(uuid string) (*Job, error) {
	job, err := readJob(s.db.QueryRow("select uuid, name, status, arguments, result, created, updated from Job where uuid = ?", uuid))
	return &job, err
}

func (s *SQLiteStore) SetStatus(uuid string, status string, updated time.Time) (err error) {
	_, err = s.db.Exec("update Job set status = ?, updated = ? where uuid = ?", status, updated, uuid)
	return
}

func (s *SQLiteStore) SetResult(uuid string, result interface{}) (err error) {
	value, err := encode(result)
	if err != nil {
		return
	}
	_, err = s.db.Exec("update Job set result = ? where uuid = ?", toNullString(value), uuid)
	return
}

type SQLRow interface {
	Scan(...interface{}) error
}

func readJob(row SQLRow) (job Job, err error) {
	job = Job{}
	var arguments sql.NullString
	var result sql.NullString
	err = row.Scan(&job.UUID, &job.Name, &job.Status, &arguments, &result, &job.Created, &job.Updated)
	if err != nil {
		return
	}
	job.Arguments, err = decode(arguments)
	if err != nil {
		return
	}
	job.Result, err = decode(result)
	if err != nil {
		return
	}
	return
}

func encode(thing interface{}) (result string, err error) {
	if thing == nil {
		return
	}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(thing)
	result = buf.String()
	return
}

func decode(result sql.NullString) (thing interface{}, err error) {
	if !result.Valid {
		thing = ""
		return
	}
	err = json.NewDecoder(bytes.NewBufferString(result.String)).Decode(&thing)
	if err != nil {
		return
	}
	if thing == nil {
		thing = ""
	}
	return
}

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
