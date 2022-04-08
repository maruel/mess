package model

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

type taskOutputs struct {
}

func (t *taskOutputs) init() error {
	if d, err := os.Stat("output"); err == nil {
		if !d.IsDir() {
			return errors.New("output is not a directory")
		}
	} else if err := os.Mkdir("output", 0o755); err != nil {
		return err
	}
	return nil
}

func (t *taskOutputs) WriteOutput(key int64) (io.WriteCloser, error) {
	p := filepath.Join("output", strconv.FormatInt(key, 10))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	return f, err
}

func (t *taskOutputs) ReadOutput(key int64) (io.ReadCloser, error) {
	p := filepath.Join("output", strconv.FormatInt(key, 10))
	f, err := os.Open(p)
	return f, err
}
