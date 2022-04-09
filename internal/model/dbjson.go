package model

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

type rawTables struct {
	// Single thread the whole thing. It's going to be a problem eventually.
	mu sync.Mutex

	// Immutable after creation.
	TasksRequest map[int64]*TaskRequest

	TasksResult map[int64]*TaskResult

	Bots map[string]*Bot

	BotEvents      map[string][]*BotEvent
	nextBotEventID int64
}

func (t *rawTables) TaskRequestGet(id int64, r *TaskRequest) {
	// No need for deep copy since TaskRequest are immutable.
	t.mu.Lock()
	d := t.TasksRequest[id]
	if d != nil {
		*r = *d
	}
	t.mu.Unlock()
}

func (t *rawTables) TaskRequestAdd(r *TaskRequest) {
	if r.Key != 0 {
		panic("do not set key")
	}
	t.mu.Lock()
	r.Key = int64(len(t.TasksRequest)) + 1
	if t.TasksRequest[r.Key] != nil {
		panic("task requests are immutable")
	}
	v := &TaskRequest{}
	// TODO(maruel): Deep copy slices. :(
	*v = *r
	t.TasksRequest[r.Key] = v
	t.mu.Unlock()
}

func (t *rawTables) TaskRequestCount() int64 {
	t.mu.Lock()
	l := len(t.TasksRequest)
	t.mu.Unlock()
	return int64(l)
}

func (t *rawTables) TaskRequestSlice(f Filter) ([]TaskRequest, string) {
	if f.Cursor != "" || !f.Earliest.IsZero() || !f.Latest.IsZero() {
		panic("implement filters")
	}
	if f.Limit == 0 {
		panic("set limit")
	}
	t.mu.Lock()
	l := len(t.TasksRequest)
	if l > f.Limit {
		l = f.Limit
	}
	out := make([]TaskRequest, 0, l)
	// TODO(maruel): Copy in order.
	for _, v := range t.TasksRequest {
		// TODO(maruel): Deep copy slices. :(
		out = append(out, *v)
	}
	t.mu.Unlock()
	return out, ""
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

func (t *rawTables) TaskResultCount() int64 {
	t.mu.Lock()
	l := len(t.TasksResult)
	t.mu.Unlock()
	return int64(l)
}

func (t *rawTables) BotGet(id string, b *Bot) {
	t.mu.Lock()
	// TODO(maruel): Deep copy slices. :(
	d := t.Bots[id]
	if d != nil {
		*b = *d
	}
	t.mu.Unlock()
}

func (t *rawTables) TaskResultSlice(botid string, f Filter, state TaskStateQuery, sort TaskSort) ([]TaskResult, string) {
	if f.Cursor != "" || !f.Earliest.IsZero() || !f.Latest.IsZero() || state != TaskStateQueryAll || sort != TaskSortCreated {
		panic("implement filters")
	}
	if f.Limit == 0 {
		panic("set limit")
	}
	t.mu.Lock()
	l := len(t.TasksResult)
	if l > f.Limit {
		l = f.Limit
	}
	out := make([]TaskResult, 0, l)
	// TODO(maruel): Copy in order.
	for _, v := range t.TasksResult {
		// TODO(maruel): Deep copy slices. :(
		out = append(out, *v)
	}
	t.mu.Unlock()
	return out, ""
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

func (t *rawTables) BotCount(dims map[string]string) (total, quarantined, maintenance, dead, busy int64) {
	cutoff := time.Now().Add(-DeadAfter)
	if len(dims) != 0 {
		panic("implement filters")
	}
	t.mu.Lock()
	for _, b := range t.Bots {
		if b.Deleted {
			continue
		}
		total++
		if b.QuarantinedMsg != "" {
			quarantined++
		}
		if b.MaintenanceMsg != "" {
			maintenance++
		}
		if b.LastSeen.After(cutoff) {
			dead++
		}
		if b.TaskID != 0 {
			busy++
		}
	}
	t.mu.Unlock()
	return
}

func (t *rawTables) BotGetSlice(cursor string, limit int) ([]Bot, string) {
	if cursor != "" {
		panic("implement cursor")
	}
	if limit == 0 {
		panic("set limit")
	}
	t.mu.Lock()
	l := len(t.Bots) / 2
	if l > limit {
		l = limit
	}
	b := make([]Bot, 0, l)
	for _, v := range t.Bots {
		if v.Deleted {
			continue
		}
		// TODO(maruel): Deep copy slices. :(
		b = append(b, *v)
		if len(b) == l {
			break
		}
	}
	t.mu.Unlock()
	return b, ""
}

func (t *rawTables) BotEventAdd(e *BotEvent) {
	if e.Key != 0 {
		panic("do not set key")
	}
	t.mu.Lock()
	t.nextBotEventID++
	e.Key = t.nextBotEventID
	t.BotEvents[e.BotID] = append(t.BotEvents[e.BotID], e)
	t.mu.Unlock()
}

func (t *rawTables) BotEventGetSlice(botid string, f Filter) ([]BotEvent, string) {
	if f.Cursor != "" || !f.Earliest.IsZero() || !f.Latest.IsZero() {
		panic("implement filters")
	}
	t.mu.Lock()
	be := t.BotEvents[botid]
	l := len(be)
	if l > f.Limit {
		l = f.Limit
	}
	b := make([]BotEvent, 0, l)
	// Copy in reverse order.
	for i := len(be) - 1; i >= 0; i-- {
		// TODO(maruel): Deep copy slices. :(
		b = append(b, *be[i])
	}
	t.mu.Unlock()
	return b, ""
}

func (t *rawTables) init() error {
	t.TasksRequest = map[int64]*TaskRequest{}
	t.TasksResult = map[int64]*TaskResult{}
	t.Bots = map[string]*Bot{}
	t.BotEvents = map[string][]*BotEvent{}
	return nil
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
	j.nextBotEventID = 0
	for _, val := range j.BotEvents {
		j.nextBotEventID += int64(len(val))
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
