package model

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
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

// TaskOutputs is a good enough task outputs manager.
//
// It uses a files backed store.
//
// TODO: Implement compression.
type TaskOutputs struct {
	root    string
	mu      sync.Mutex
	handles map[int64]*output
}

type output struct {
	mu   sync.Mutex
	f    *os.File
	buf  []byte
	err  error
	last time.Time
}

// NewTaskOutputs returns an initialized TaskOutputs.
func NewTaskOutputs(root string) (*TaskOutputs, error) {
	t := &TaskOutputs{
		root:    root,
		handles: map[int64]*output{},
	}
	if d, err := os.Stat(t.root); err == nil {
		if !d.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", t.root)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	} else if err := os.Mkdir(t.root, 0o755); err != nil {
		return nil, err
	}
	return t, nil
}

// SetOutput sets the output for a task at the specified offset.
func (t *TaskOutputs) SetOutput(key, offset int64, content []byte) error {
	o := t.getLocked(key, offset, false)
	if o.err == nil {
		_, o.err = o.f.Write(content)
	}
	o.mu.Unlock()
	return o.err
}

// ReadOutput reads the task output from a file at the specified offset.
func (t *TaskOutputs) ReadOutput(key, offset int64, max int) ([]byte, error) {
	o := t.getLocked(key, offset, true)
	d := 0
	if o.err == nil {
		if len(o.buf) < max {
			o.buf = make([]byte, max)
		}
		d, o.err = o.f.Read(o.buf)
	}
	o.mu.Unlock()
	return o.buf[:d], o.err
}

// Loop should be run to lazily close file handles.
func (t *TaskOutputs) Loop(ctx context.Context, max int, cutoff time.Duration) {
	done := ctx.Done()
	for jitter := 0; ; jitter = (jitter + 1) % 6 {
		select {
		case now := <-time.After(time.Minute + time.Duration(jitter)*time.Second):
			old := now.Add(-cutoff)
			t.mu.Lock()
			for k, o := range t.handles {
				if old.After(o.last) {
					if o.f != nil {
						o.f.Close()
					}
					delete(t.handles, k)
				}
			}
			for len(t.handles) > max {
				// Close random files.
				for k, o := range t.handles {
					if o.f != nil {
						o.f.Close()
					}
					delete(t.handles, k)
				}
			}
			t.mu.Unlock()
		case <-done:
			return
		}
	}
}

func (t *TaskOutputs) getLocked(key, offset int64, forRead bool) *output {
	t.mu.Lock()
	o := t.handles[key]
	if o == nil {
		o = &output{}
		t.handles[key] = o
	}
	o.mu.Lock()
	t.mu.Unlock()

	if o.err == nil && !os.IsNotExist(o.err) {
		if o.f == nil {
			p := filepath.Join(t.root, strconv.FormatInt(key, 10))
			flag := os.O_RDWR
			if !forRead {
				flag |= os.O_CREATE
			}
			o.f, o.err = os.OpenFile(p, flag, 0o644)
		}
		if o.err == nil {
			// TODO(maruel): I'm not sure if we acn seek past file's end. Need to
			// unit test.
			if _, o.err = o.f.Seek(offset, io.SeekStart); o.err == nil {
				o.last = time.Now()
			}
		}
	}
	return o
}
