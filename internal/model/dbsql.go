package model

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

type sqlDB struct {
	db *sql.DB
}

// NewDBSqlite3 creates or opens a sqlite3 DB.
func NewDBSqlite3(p string) (DB, error) {
	s := &sqlDB{}
	//c, err := sqlx.Connect("sqlite3", p)
	c, err := sql.Open("sqlite3", p)
	if err != nil {
		return nil, err
	}
	s.db = c

	// Make sure the tables are setup*
	_, err = s.db.Exec(schemaTaskRequest)
	if err != nil {
		s.db.Close()
		return nil, err
	}
	return s, nil
}

func (s *sqlDB) Snapshot() error {
	// Unnecessary.
	return nil
}

func (s *sqlDB) Close() error {
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *sqlDB) TaskRequestGet(id int64, r *TaskRequest) {
	r2 := TaskRequest{}
	row := s.db.QueryRow("SELECT * FROM TaskRequest WHERE Key = ?", id)
	if err := row.Scan(r2.fields()...); err == sql.ErrNoRows {
		return
	} else if err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return
	}
	if err := json.Unmarshal(r2.raw, &r2.Blob); err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return
	}
	r2.raw = nil
	*r = r2
}

func (s *sqlDB) TaskRequestSet(r *TaskRequest) {
	var err error
	if r.raw, err = json.Marshal(&r.Blob); err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		r.raw = nil
		return
	}
	stmt := "INSERT INTO TaskRequest (Key, SchemaVersion, Created, Priority, ParentTask, Tags, raw) VALUES ($1, $2)"
	_, err = s.db.Exec(stmt, r.fields()...)
	if err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return
	}
	r.raw = nil
}

func (s *sqlDB) TaskRequestCount() int {
	row := s.db.QueryRow("SELECT COUNT(*) FROM TaskRequest")
	count := 0
	if err := row.Scan(&count); err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return 0
	}
	return count
}

func (s *sqlDB) TaskResultGet(id int64, r *TaskResult) { panic("TODO") }
func (s *sqlDB) TaskResultSet(r *TaskResult)           { panic("TODO") }
func (s *sqlDB) TaskResultCount() int                  { panic("TODO") }
func (s *sqlDB) BotGet(id string, b *Bot)              { panic("TODO") }
func (s *sqlDB) BotSet(b *Bot)                         { panic("TODO") }
func (s *sqlDB) BotCount() int                         { panic("TODO") }
func (s *sqlDB) BotGetAll(b []Bot) []Bot               { panic("TODO") }
