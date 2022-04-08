package model

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type sqlDB struct {
	db *sql.DB
}

// NewDBSqlite3 creates or opens a sqlite3 DB.
func NewDBSqlite3(p string) (DB, error) {
	s := &sqlDB{}
	c, err := sql.Open("sqlite3", "file:"+p)
	if err != nil {
		return nil, err
	}
	s.db = c

	// TODO(maruel): Figure out reasonable cache size.
	// https://sqlite.org/pragma.html#pragma_case_sensitive_like
	_, err = s.db.Exec("PRAGMA case_sensitive_like = TRUE;")
	if err != nil {
		s.db.Close()
		return nil, err
	}

	// Make sure the tables are setup.
	for _, stmt := range []string{schemaTaskRequest, schemaTaskResult, schemaBot, schemaBotEvent} {
		if _, err = s.db.Exec(stmt); err != nil {
			s.db.Close()
			return nil, err
		}
	}
	return s, nil
}

func (s *sqlDB) Snapshot() error {
	// https://sqlite.org/pragma.html#pragma_auto_vacuum
	// TODO(maruel): VACUUM but not too often.
	return nil
}

func (s *sqlDB) Close() error {
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *sqlDB) TaskRequestGet(id int64, r *TaskRequest) {
	r2 := taskRequestSQL{}
	row := s.db.QueryRow("SELECT * FROM TaskRequest WHERE key = ?", id)
	if err := row.Scan(r2.fields()...); err == sql.ErrNoRows {
		return
	} else if err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return
	}
	r2.to(r)
}

func (s *sqlDB) TaskRequestSet(r *TaskRequest) {
	r2 := taskRequestSQL{}
	r2.from(r)
	stmt := "INSERT INTO TaskRequest (key, schemaVersion, created, priority, parentTask, tags, blob) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	if _, err := s.db.Exec(stmt, r2.fields()...); err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return
	}
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

func (s *sqlDB) TaskResultGet(id int64, r *TaskResult) {
	r2 := taskResultSQL{}
	row := s.db.QueryRow("SELECT * FROM TaskResult WHERE key = ?", id)
	if err := row.Scan(r2.fields()...); err == sql.ErrNoRows {
		return
	} else if err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return
	}
	r2.to(r)
}

func (s *sqlDB) TaskResultSet(r *TaskResult) {
	r2 := taskResultSQL{}
	r2.from(r)
	stmt := "INSERT OR REPLACE INTO TaskResult (key, schemaVersion, botID, blob) VALUES ($1, $2, $3, $4)"
	if _, err := s.db.Exec(stmt, r2.fields()...); err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return
	}
}

func (s *sqlDB) TaskResultCount() int {
	row := s.db.QueryRow("SELECT COUNT(*) FROM TaskResult")
	count := 0
	if err := row.Scan(&count); err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return 0
	}
	return count
}

func (s *sqlDB) BotGet(id string, b *Bot) {
	b2 := botSQL{}
	row := s.db.QueryRow("SELECT * FROM Bot WHERE key = ?", id)
	if err := row.Scan(b2.fields()...); err == sql.ErrNoRows {
		return
	} else if err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return
	}
	b2.to(b)
}

func (s *sqlDB) BotSet(b *Bot) {
	b2 := botSQL{}
	b2.from(b)
	stmt := "INSERT OR REPLACE INTO Bot (key, schemaVersion, created, lastSeen, version, blob) VALUES ($1, $2, $3, $4, $5, $6)"
	if _, err := s.db.Exec(stmt, b2.fields()...); err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return
	}
}

func (s *sqlDB) BotCount() int {
	row := s.db.QueryRow("SELECT COUNT(*) FROM Bot")
	count := 0
	if err := row.Scan(&count); err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
		return 0
	}
	return count
}

func (s *sqlDB) BotGetAll(all []Bot) []Bot {
	rows, err := s.db.Query("SELECT * FROM Bot ORDER BY key")
	if err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
	}
	b2 := botSQL{}
	b := Bot{}
	for rows.Next() {
		if err := rows.Scan(b2.fields()...); err != nil {
			// TODO(maruel): Surface error? Delete entity?
			panic(err)
		}
		b2.to(&b)
		all = append(all, b)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		// TODO(maruel): Surface error? Delete entity?
		panic(err)
	}
	rows.Close()
	return all
}
