package main

import (
	"encoding/json"
	"io"
)

type db struct {
	Requests map[int64]*TaskRequest
	//Results  map[int64]*TaskResult
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
