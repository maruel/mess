package model

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type sqlDB struct {
	db *sql.DB
}

func NewDBSqlite3() (DB, error) {
	s := &sqlDB{}
	//c, err := sqlx.Connect("sqlite3", "mess.db")
	c, err := sql.Open("sqlite3", "mess.db")
	if err != nil {
		return nil, err
	}
	s.db = c
	/*
		// Query all the tables.
		if err = s.db.Select(&s.tables.TasksRequest, "SELECT * FROM TasksRequest"); err != nil {
			s.db.Close()
			return nil, err
		}
		if err = s.db.Select(&s.tables.TasksResult, "SELECT * FROM TasksResult"); err != nil {
			s.db.Close()
			return nil, err
		}
		if err = s.db.Select(&s.tables.Bots, "SELECT * FROM Bots"); err != nil {
			s.db.Close()
			return nil, err
		}
	*/
	return s, nil
}

func (s *sqlDB) Snapshot() error {
	return nil
}

func (s *sqlDB) Close() error {
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *sqlDB) TaskRequestGet(id int64, r *TaskRequest) { panic("TODO") }
func (s *sqlDB) TaskRequestSet(r *TaskRequest)           { panic("TODO") }
func (s *sqlDB) TaskRequestCount() int                   { panic("TODO") }
func (s *sqlDB) TaskResultGet(id int64, r *TaskResult)   { panic("TODO") }
func (s *sqlDB) TaskResultSet(r *TaskResult)             { panic("TODO") }
func (s *sqlDB) TaskResultCount() int                    { panic("TODO") }
func (s *sqlDB) BotGet(id string, b *Bot)                { panic("TODO") }
func (s *sqlDB) BotSet(b *Bot)                           { panic("TODO") }
func (s *sqlDB) BotCount() int                           { panic("TODO") }
func (s *sqlDB) BotGetAll(b []Bot) []Bot                 { panic("TODO") }
