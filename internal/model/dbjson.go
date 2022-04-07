package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/klauspost/compress/zstd"
)

type rawTables struct {
	// Single thread the whole thing. It's going to be a problem eventually.
	mu sync.Mutex

	// Immutable after creation.
	TasksRequest map[int64]*TaskRequest

	TasksResult map[int64]*TaskResult

	Bots map[string]*Bot
}

func (t *rawTables) TaskRequestGet(id int64, r *TaskRequest) {
	// No need for deep copy since TaskRequest are immutable.
	t.mu.Lock()
	*r = *t.TasksRequest[id]
	t.mu.Unlock()
}

func (t *rawTables) TaskRequestSet(r *TaskRequest) {
	t.mu.Lock()
	if t.TasksRequest[r.Key] == nil {
		v := &TaskRequest{}
		// TODO(maruel): Deep copy slices. :(
		*v = *r
		t.TasksRequest[r.Key] = v
	} else {
		panic("task requests are immutable")
	}
	t.mu.Unlock()
}

func (t *rawTables) TaskRequestCount() int {
	t.mu.Lock()
	l := len(t.TasksRequest)
	t.mu.Unlock()
	return l
}

func (t *rawTables) TaskResultGet(id int64, r *TaskResult) {
	t.mu.Lock()
	// TODO(maruel): Deep copy slices. :(
	*r = *t.TasksResult[id]
	t.mu.Unlock()
}

func (t *rawTables) TaskResultSet(r *TaskResult) {
	t.mu.Lock()
	if t.TasksResult[r.Key] == nil {
		v := &TaskResult{}
		// TODO(maruel): Deep copy slices. :(
		*v = *r
		t.TasksResult[r.Key] = v
	} else {
		// TODO(maruel): Deep copy slices. :(
		*t.TasksResult[r.Key] = *r
	}
	t.mu.Unlock()
}

func (t *rawTables) TaskResultCount() int {
	t.mu.Lock()
	l := len(t.TasksResult)
	t.mu.Unlock()
	return l
}

func (t *rawTables) BotGet(id string, b *Bot) {
	t.mu.Lock()
	// TODO(maruel): Deep copy slices. :(
	*b = *t.Bots[id]
	t.mu.Unlock()
}

func (t *rawTables) BotSet(b *Bot) {
	t.mu.Lock()
	if t.Bots[b.Key] == nil {
		v := &Bot{}
		// TODO(maruel): Deep copy slices. :(
		*v = *b
		t.Bots[b.Key] = v
	} else {
		// TODO(maruel): Deep copy slices. :(
		*t.Bots[b.Key] = *b
	}
	t.mu.Unlock()
}

func (t *rawTables) BotCount() int {
	t.mu.Lock()
	l := len(t.Bots)
	t.mu.Unlock()
	return l
}

func (t *rawTables) BotGetAll(b []Bot) []Bot {
	t.mu.Lock()
	if len(b) < len(t.Bots) {
		b = make([]Bot, len(t.Bots))
	}
	i := 0
	for _, v := range t.Bots {
		// TODO(maruel): Deep copy slices. :(
		b[i] = *v
		i++
	}
	t.mu.Unlock()
	return b
}

func (t *rawTables) init() error {
	if d, err := os.Stat("output"); err == nil {
		if !d.IsDir() {
			return errors.New("output is not a directory")
		}
	} else if err := os.Mkdir("output", 0755); err != nil {
		return err
	}
	t.TasksRequest = map[int64]*TaskRequest{}
	t.TasksResult = map[int64]*TaskResult{}
	t.Bots = map[string]*Bot{}
	return nil
}

func (t *rawTables) WriteOutput(key int64) (io.WriteCloser, error) {
	p := filepath.Join("output", strconv.FormatInt(key, 10))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	return f, err
}

func (t *rawTables) ReadOutput(key int64) (io.ReadCloser, error) {
	p := filepath.Join("output", strconv.FormatInt(key, 10))
	f, err := os.Open(p)
	return f, err
}

type jsonDriver struct {
	rawTables
	p string
}

// NewDBJSON opens file p.
//
// Use "db.json.zst".
func NewDBJSON(p string) (DB, error) {
	j := &jsonDriver{p: p}
	if err := j.rawTables.init(); err != nil {
		return nil, err
	}

	// When the process starts, there isn't much memory use yet. So buffer the
	// whole file in memory and decompress all at once. It will be garbage
	// collected quickly.
	//
	// Use zstd since we can about shutdown / startup performance. We may want to
	// use a more effective encoding that json.
	src, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return j, nil
	}
	if err != nil {
		return nil, err
	}
	dec, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	if err != nil {
		return nil, err
	}
	raw, err := dec.DecodeAll(src, nil)
	if err != nil {
		return nil, err
	}
	if err := j.loadFrom(bytes.NewReader(raw)); err != nil {
		return nil, err
	}
	return j, nil
}

func (j *jsonDriver) Snapshot() error {
	return j.Close()
}

func (j *jsonDriver) Close() error {
	// It's probably faster to buffer all in memory and write as one shot. It
	// will likely use more memory and could be problematic over heavy memory
	// usage. Stream for now to reduce risks.
	d := filepath.Dir(j.p)
	b := filepath.Base(j.p)
	if i := strings.Index(b, "."); i != -1 {
		b = b[:i] + ".new" + b[i:]
	} else {
		b = b + ".new"
	}
	n := filepath.Join(d, b)
	f, err := os.OpenFile(n, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	enc, err := zstd.NewWriter(f)
	if err == nil {
		err = j.saveTo(enc)
	}
	if err2 := enc.Close(); err == nil {
		err = err2
	}
	if err2 := f.Close(); err == nil {
		err = err2
	}
	if err == nil {
		// Only overwrite if saving worked.
		err = os.Rename(n, j.p)
	}
	return err
}

func (j *jsonDriver) loadFrom(r io.Reader) error {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	d.UseNumber()
	j.rawTables.mu.Lock()
	if err := d.Decode(&j.rawTables); err != nil {
		return err
	}
	j.rawTables.mu.Unlock()
	// TODO(maruel): Validate.
	return nil
}

func (j *jsonDriver) saveTo(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)
	j.rawTables.mu.Lock()
	err := e.Encode(&j.rawTables)
	j.rawTables.mu.Unlock()
	return err
}
