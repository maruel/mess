package model

import (
	"encoding/json"
	"time"
)

type Bot struct {
	Key             string              `json:"a"`
	SchemaVersion   int                 `json:"b"`
	Created         time.Time           `json:"c"`
	LastSeen        time.Time           `json:"d"`
	Version         string              `json:"e"`
	AuthenticatedAs string              `json:"f"`
	Dimensions      map[string][]string `json:"g"`
	State           []byte              `json:"h"`
	ExternalIP      string              `json:"i"`
	TaskID          int                 `json:"j"`
	QuarantinedMsg  string              `json:"k"`
	MaintenanceMsg  string              `json:"l"`
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
	s := botSQLBlob{
		AuthenticatedAs: d.AuthenticatedAs,
		Dimensions:      d.Dimensions,
		State:           d.State,
		ExternalIP:      d.ExternalIP,
		TaskID:          d.TaskID,
		QuarantinedMsg:  d.QuarantinedMsg,
		MaintenanceMsg:  d.MaintenanceMsg,
	}
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
	d.AuthenticatedAs = s.AuthenticatedAs
	d.Dimensions = s.Dimensions
	d.State = s.State
	d.ExternalIP = s.ExternalIP
	d.TaskID = s.TaskID
	d.QuarantinedMsg = s.QuarantinedMsg
	d.MaintenanceMsg = s.MaintenanceMsg
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
	AuthenticatedAs string              `json:"a"`
	Dimensions      map[string][]string `json:"b"`
	State           []byte              `json:"c"`
	ExternalIP      string              `json:"d"`
	TaskID          int                 `json:"e"`
	QuarantinedMsg  string              `json:"f"`
	MaintenanceMsg  string              `json:"g"`
}
