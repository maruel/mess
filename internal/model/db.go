package model

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// TaskSort is a way to sort tasks.
type TaskSort int

// Valid TaskSort.
const (
	TaskSortCreated TaskSort = iota
	TaskSortModified
	TaskSortCompleted
	TaskSortAbandoned
	TaskSortStarted
)

// TaskStateQuery filters on different kinds of tasks.
type TaskStateQuery int

// Valid TaskStateQuery.
const (
	TaskStateQueryPending TaskStateQuery = iota
	TaskStateQueryRunning
	TaskStateQueryPendingRunning
	TaskStateQueryCompleted
	TaskStateQueryCompletedSuccess
	TaskStateQueryCompletedFailure
	TaskStateQueryExpired
	TaskStateQueryTimedOut
	TaskStateQueryBotDied
	TaskStateQueryCanceled
	TaskStateQueryAll
	TaskStateQueryDeduped
	TaskStateQueryKilled
	TaskStateQueryNoResource
)

// Filter is a set of typical filters
type Filter struct {
	Cursor   string
	Limit    int
	Earliest time.Time
	Latest   time.Time
}

// Tables is the functions to access Swarming DB tables.
type Tables interface {
	TaskRequestGet(id int64, r *TaskRequest)
	// TaskRequestAdd adds a new TaskRequest. It is immutable so it is an error
	// to add two TaskRequest with the same key.
	TaskRequestAdd(r *TaskRequest)
	TaskRequestCount() int64
	TaskRequestSlice(f Filter) ([]TaskRequest, string)

	TaskResultGet(id int64, r *TaskResult)
	TaskResultSet(r *TaskResult)
	TaskResultCount() int64
	TaskResultSlice(botid string, f Filter, state TaskStateQuery, sort TaskSort) ([]TaskResult, string)

	BotGet(id string, b *Bot)
	BotSet(b *Bot)
	BotCount(dims map[string]string) (total, quarantined, maintenance, dead, busy int64)
	BotGetSlice(cursor string, limit int) ([]Bot, string)

	BotEventAdd(e *BotEvent)
	BotEventGetSlice(botid string, f Filter) ([]BotEvent, string)
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
