package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/klauspost/compress/zstd"
)

type tables struct {
	// Single thread the whole thing. It's going to be a problem eventually.
	mu sync.Mutex

	// Immutable after creation.
	TasksRequest map[int64]*TaskRequest

	TasksResult map[int64]*TaskResult

	Bots map[string]*Bot
}

func (t *tables) init() error {
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

func (t *tables) writeOutput(key int64) (io.WriteCloser, error) {
	p := filepath.Join("output", strconv.FormatInt(key, 10))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	return f, err
}

func (t *tables) readOutput(key int64) (io.ReadCloser, error) {
	p := filepath.Join("output", strconv.FormatInt(key, 10))
	f, err := os.Open(p)
	return f, err
}

type db struct {
	tables tables
}

func (d *db) load() error {
	if err := d.tables.init(); err != nil {
		return err
	}

	// When the process starts, there isn't much memory use yet. So buffer the
	// whole file in memory and decompress all at once. It will be garbage
	// collected quickly.
	//
	// Use zstd since we can about shutdown / startup performance. We may want to
	// use a more effective encoding that json.
	src, err := os.ReadFile("db.json.zst")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	dec, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	if err != nil {
		return err
	}
	raw, err := dec.DecodeAll(src, nil)
	if err != nil {
		return err
	}
	return d.loadFrom(bytes.NewReader(raw))
}

func (d *db) save() error {
	// It's probably faster to buffer all in memory and write as one shot. It
	// will likely use more memory and could be problematic over heavy memory
	// usage. Stream for now to reduce risks.
	f, err := os.OpenFile("db.json.new.zst", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	enc, err := zstd.NewWriter(f)
	if err == nil {
		err = d.saveTo(enc)
	}
	if err2 := enc.Close(); err == nil {
		err = err2
	}
	if err2 := f.Close(); err == nil {
		err = err2
	}
	if err == nil {
		// Only overwrite if saving worked.
		err = os.Rename("db.json.new.zst", "db.json.zst")
	}
	return err
}

func (d *db) loadFrom(r io.Reader) error {
	j := json.NewDecoder(r)
	j.DisallowUnknownFields()
	j.UseNumber()
	d.tables.mu.Lock()
	if err := j.Decode(&d.tables); err != nil {
		return err
	}
	d.tables.mu.Unlock()
	// TODO(maruel): Validate.
	return nil
}

func (d *db) saveTo(w io.Writer) error {
	j := json.NewEncoder(w)
	j.SetEscapeHTML(false)
	d.tables.mu.Lock()
	err := j.Encode(&d.tables)
	d.tables.mu.Unlock()
	return err
}
