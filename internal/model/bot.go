package model

import (
	"encoding/json"
	"time"
)

type Bot struct {
	Key           string              `json:"a"`
	SchemaVersion int                 `json:"b"`
	Created       time.Time           `json:"c"`
	LastSeen      time.Time           `json:"d"`
	Version       string              `json:"e"`
	Dimensions    map[string][]string `json:"f"`
}

func (b *Bot) AddEvent(now time.Time, event, msg string, e *BotEvent) {
	e.SchemaVersion = 1
	e.Key = 0
	e.BotID = b.Key
	e.Time = now
	e.Event = event
	e.Message = msg
	// Make a copy of the map but not the values, since they are immutable (to
	// save memory).
	e.Dimensions = make(map[string][]string, len(b.Dimensions))
	for k, v := range b.Dimensions {
		e.Dimensions[k] = v
	}
	e.Version = b.Version
}

type botSQL struct {
	key           string
	schemaVersion int
	created       int64
	lastSeen      int64
	version       string
	blob          []byte
}

func (b *botSQL) fields() []interface{} {
	return []interface{}{
		&b.key,
		&b.schemaVersion,
		&b.created,
		&b.lastSeen,
		&b.version,
		&b.blob,
	}
}

func (b *botSQL) from(d *Bot) {
	b.key = d.Key
	b.schemaVersion = d.SchemaVersion
	b.created = d.Created.UnixMicro()
	b.lastSeen = d.LastSeen.UnixMicro()
	b.version = d.Version
	s := botSQLBlob{Dimensions: d.Dimensions}
	var err error
	b.blob, err = json.Marshal(&s)
	if err != nil {
		panic("internal error: " + err.Error())
	}
}

func (b *botSQL) to(d *Bot) {
	d.Key = b.key
	d.SchemaVersion = b.schemaVersion
	d.Created = time.UnixMicro(b.created).UTC()
	d.LastSeen = time.UnixMicro(b.lastSeen).UTC()
	d.Version = b.version
	s := botSQLBlob{}
	if err := json.Unmarshal(b.blob, &s); err != nil {
		panic("internal error: " + err.Error())
	}
	d.Dimensions = s.Dimensions
}

// See:
// - https://sqlite.org/lang_createtable.html#rowids_and_the_integer_primary_key
// - https://sqlite.org/datatype3.html
// BLOB
const schemaBot = `
CREATE TABLE IF NOT EXISTS Bot (
	key           TEXT    PRIMARY KEY,
	schemaVersion INTEGER NOT NULL,
	created       INTEGER NOT NULL,
	lastSeen      INTEGER NOT NULL,
	version       TEXT,
	blob          BLOB    NOT NULL
) STRICT;
`

// botSQLBlob contains the unindexed fields.
type botSQLBlob struct {
	Dimensions map[string][]string `json:"a"`
}
