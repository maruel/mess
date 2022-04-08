package model

import (
	"io"
	"time"
)

// Tables is the functions to access Swarming DB tables.
type Tables interface {
	TaskRequestGet(id int64, r *TaskRequest)
	// TaskRequestAdd adds a new TaskRequest. It is immutable so it is an error
	// to add two TaskRequest with the same key.
	TaskRequestAdd(r *TaskRequest)
	TaskRequestCount() int

	TaskResultGet(id int64, r *TaskResult)
	TaskResultSet(r *TaskResult)
	TaskResultCount() int

	BotGet(id string, b *Bot)
	BotSet(b *Bot)
	BotCount() int
	BotGetSlice(cursor string, limit int) ([]Bot, string)

	BotEventAdd(e *BotEvent)
	BotEventGetSlice(id, cursor string, limit int, earliest, latest time.Time) ([]BotEvent, string)
}

// DB is a database backend.
type DB interface {
	Tables
	io.Closer
	// Snapshot ensures there's a copy on disk in case of a crash.
	Snapshot() error
}
