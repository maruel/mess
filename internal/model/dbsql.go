package model

import (
	"database/sql"
	"sync"

	// Force the sqlite3 driver to be registered.
	_ "github.com/mattn/go-sqlite3"
)

type sqlDB struct {
	db             *sql.DB
	mu             sync.Mutex
	lastTaskID     int64
	lastBotEventID int64
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
	s.db.QueryRow("SELECT key FROM TaskRequest ORDER BY key DESC").Scan(&s.lastTaskID)
	s.db.QueryRow("SELECT key FROM BotEvent ORDER BY key DESC").Scan(&s.lastBotEventID)
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
		panic(err)
		return
	}
	r2.to(r)
}

func (s *sqlDB) TaskRequestAdd(r *TaskRequest) {
	if r.Key != 0 {
		panic("do not set key")
	}
	r2 := taskRequestSQL{}
	r2.from(r)
	s.mu.Lock()
	s.lastTaskID++
	r2.key = s.lastTaskID
	r.Key = r2.key
	s.mu.Unlock()
	stmt := "INSERT INTO TaskRequest (key, schemaVersion, created, priority, parentTask, tags, blob) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	if _, err := s.db.Exec(stmt, r2.fields()...); err != nil {
		panic(err)
		return
	}
}

func (s *sqlDB) TaskRequestCount() int64 {
	row := s.db.QueryRow("SELECT COUNT(*) FROM TaskRequest")
	count := int64(0)
	if err := row.Scan(&count); err != nil {
		panic(err)
		return 0
	}
	return count
}
func (s *sqlDB) TaskRequestSlice(f Filter) ([]TaskRequest, string) {
	if f.Cursor != "" || !f.Earliest.IsZero() || !f.Latest.IsZero() {
		panic("implement filters")
	}
	if f.Limit == 0 {
		panic("set limit")
	}
	rows, err := s.db.Query("SELECT * FROM TaskRequest ORDER BY key DESC LIMIT ?", f.Limit)
	if err != nil {
		panic(err)
	}
	var all []TaskRequest
	d := taskRequestSQL{}
	r := TaskRequest{}
	for rows.Next() {
		if err := rows.Scan(d.fields()...); err != nil {
			panic(err)
		}
		d.to(&r)
		all = append(all, r)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	rows.Close()
	return all, ""
}

func (s *sqlDB) TaskResultGet(id int64, r *TaskResult) {
	r2 := taskResultSQL{}
	row := s.db.QueryRow("SELECT * FROM TaskResult WHERE key = ?", id)
	if err := row.Scan(r2.fields()...); err == sql.ErrNoRows {
		return
	} else if err != nil {
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
		panic(err)
		return
	}
}

func (s *sqlDB) TaskResultCount() int64 {
	row := s.db.QueryRow("SELECT COUNT(*) FROM TaskResult")
	count := int64(0)
	if err := row.Scan(&count); err != nil {
		panic(err)
		return 0
	}
	return count
}

func (s *sqlDB) TaskResultSlice(botid string, f Filter, state TaskStateQuery, sort TaskSort) ([]TaskResult, string) {
	if f.Cursor != "" || !f.Earliest.IsZero() || !f.Latest.IsZero() || state != TaskStateQueryAll || sort != TaskSortCreated {
		// TODO(maruel): Implement.
		//panic("implement filters")
	}
	if f.Limit == 0 {
		panic("set limit")
	}
	var rows *sql.Rows
	var err error
	if botid != "" {
		rows, err = s.db.Query("SELECT * FROM TaskResult WHERE botID = ? ORDER BY key DESC LIMIT ?", botid, f.Limit)
	} else {
		rows, err = s.db.Query("SELECT * FROM TaskResult ORDER BY key DESC LIMIT ?", f.Limit)
	}
	if err != nil {
		panic(err)
	}
	var all []TaskResult
	d := taskResultSQL{}
	r := TaskResult{}
	for rows.Next() {
		if err := rows.Scan(d.fields()...); err != nil {
			panic(err)
		}
		d.to(&r)
		all = append(all, r)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	rows.Close()
	return all, ""
}

func (s *sqlDB) BotGet(id string, b *Bot) {
	b2 := botSQL{}
	row := s.db.QueryRow("SELECT * FROM Bot WHERE key = ?", id)
	if err := row.Scan(b2.fields()...); err == sql.ErrNoRows {
		return
	} else if err != nil {
		panic(err)
		return
	}
	b2.to(b)
}

func (s *sqlDB) BotSet(b *Bot) {
	b2 := botSQL{}
	b2.from(b)
	stmt := "INSERT OR REPLACE INTO Bot " +
		"(key, schemaVersion, created, lastSeen, version, deleted, dead, quarantinedMsg, maintenanceMsg, taskID, blob) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)"
	if _, err := s.db.Exec(stmt, b2.fields()...); err != nil {
		panic(err)
		return
	}
}

func (s *sqlDB) BotCount(dims map[string]string) (total, quarantined, maintenance, dead, busy int64) {
	if len(dims) != 0 {
		panic("implement dims")
	}
	row := s.db.QueryRow("SELECT COUNT(*), COUNT(quarantinedMsg), COUNT(maintenanceMsg), COUNT(dead), COUNT(taskID) FROM Bot WHERE NOT deleted")
	if err := row.Scan(&total, &quarantined, &maintenance, &dead, &busy); err != nil {
		panic(err)
		return 0, 0, 0, 0, 0
	}
	return
}

func (s *sqlDB) BotGetSlice(cursor string, limit int) ([]Bot, string) {
	if cursor != "" {
		panic("implement cursor")
	}
	if limit == 0 {
		panic("set limit")
	}
	rows, err := s.db.Query("SELECT * FROM Bot WHERE NOT DELETED ORDER BY key LIMIT ?", limit)
	if err != nil {
		panic(err)
	}
	var all []Bot
	d := botSQL{}
	b := Bot{}
	for rows.Next() {
		if err := rows.Scan(d.fields()...); err != nil {
			panic(err)
		}
		d.to(&b)
		all = append(all, b)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	rows.Close()
	return all, ""
}

func (s *sqlDB) BotEventAdd(e *BotEvent) {
	if e.Key != 0 {
		panic("do not set key")
	}
	e2 := botEventSQL{}
	e2.from(e)
	s.mu.Lock()
	s.lastBotEventID++
	e2.key = s.lastBotEventID
	e.Key = e2.key
	s.mu.Unlock()
	stmt := "INSERT INTO BotEvent (key, schemaVersion, botID, time, blob) VALUES ($1, $2, $3, $4, $5)"
	if _, err := s.db.Exec(stmt, e2.fields()...); err != nil {
		panic(err)
		return
	}
}

func (s *sqlDB) BotEventGetSlice(botid string, f Filter) ([]BotEvent, string) {
	if f.Cursor != "" || !f.Earliest.IsZero() || !f.Latest.IsZero() {
		panic("implement filters")
	}
	if f.Limit == 0 {
		panic("set limit")
	}
	rows, err := s.db.Query("SELECT * FROM BotEvent WHERE botID = ? ORDER BY key DESC LIMIT ?", botid, f.Limit)
	if err != nil {
		panic(err)
	}
	var all []BotEvent
	d := botEventSQL{}
	b := BotEvent{}
	for rows.Next() {
		if err := rows.Scan(d.fields()...); err != nil {
			panic(err)
		}
		d.to(&b)
		all = append(all, b)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	rows.Close()
	return all, ""
}
