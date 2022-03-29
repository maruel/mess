package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

type db struct {
	// Immutable after creation.
	Requests map[int64]*TaskRequest
	Results  map[int64]*TaskResult
}

func (d *db) init() error {
	if d, err := os.Stat("output"); err == nil {
		if !d.IsDir() {
			return errors.New("output is not a directory")
		}
	} else if err := os.Mkdir("output", 0755); err != nil {
		return err
	}
	d.Requests = map[int64]*TaskRequest{}
	d.Results = map[int64]*TaskResult{}
	return nil
}

func (d *db) load(r io.Reader) error {
	j := json.NewDecoder(r)
	j.DisallowUnknownFields()
	j.UseNumber()
	if err := j.Decode(d); err != nil {
		return err
	}
	// TODO(maruel): Validate.
	return nil
}

func (d *db) save(w io.Writer) error {
	j := json.NewEncoder(w)
	j.SetEscapeHTML(false)
	return j.Encode(d)
}

func (d *db) writeOutput(key int64) (io.WriteCloser, error) {
	p := filepath.Join("output", strconv.FormatInt(key, 10))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	return f, err
}

func (d *db) readOutput(key int64) (io.ReadCloser, error) {
	p := filepath.Join("output", strconv.FormatInt(key, 10))
	f, err := os.Open(p)
	return f, err
}
