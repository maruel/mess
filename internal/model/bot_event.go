package model

import (
	"encoding/json"
	"time"
)

type BotEvent struct {
	Key             int                 `json:"a"`
	SchemaVersion   int                 `json:"b"`
	BotID           string              `json:"c"`
	Time            time.Time           `json:"d"`
	Event           string              `json:"e"`
	Message         string              `json:"f"`
	Dimensions      map[string][]string `json:"g"`
	State           string              `json:"h"`
	ExternalIP      string              `json:"i"`
	AuthenticatedAs string              `json:"j"`
	Version         string              `json:"k"`
	QuarantinedMsg  string              `json:"l"`
	MaintenanceMsg  string              `json:"m"`
	TaskID          string              `json:"n"`
}

type botEventSQL struct {
	key           int
	schemaVersion int
	botID         string
	time          int64
	blob          []byte
}

func (b *botEventSQL) fields() []interface{} {
	return []interface{}{
		&b.key,
		&b.schemaVersion,
		&b.botID,
		&b.time,
		&b.blob,
	}
}

func (b *botEventSQL) from(d *BotEvent) {
	b.key = d.Key
	b.schemaVersion = d.SchemaVersion
	b.botID = d.BotID
	b.time = d.Time.UnixMicro()
	s := botEventSQLBlob{
		Event:           d.Event,
		Message:         d.Message,
		Dimensions:      d.Dimensions,
		State:           d.State,
		ExternalIP:      d.ExternalIP,
		AuthenticatedAs: d.AuthenticatedAs,
		Version:         d.Version,
		QuarantinedMsg:  d.QuarantinedMsg,
		MaintenanceMsg:  d.MaintenanceMsg,
		TaskID:          d.TaskID,
	}
	var err error
	b.blob, err = json.Marshal(&s)
	if err != nil {
		panic("internal error: " + err.Error())
	}
}

func (b *botEventSQL) to(d *BotEvent) {
	d.Key = b.key
	d.SchemaVersion = b.schemaVersion
	d.BotID = b.botID
	d.Time = time.UnixMicro(b.time).UTC()
	s := botEventSQLBlob{}
	if err := json.Unmarshal(b.blob, &s); err != nil {
		panic("internal error: " + err.Error())
	}
	d.Event = s.Event
	d.Message = s.Message
	d.Dimensions = s.Dimensions
	d.State = s.State
	d.ExternalIP = s.ExternalIP
	d.AuthenticatedAs = s.AuthenticatedAs
	d.Version = s.Version
	d.QuarantinedMsg = s.QuarantinedMsg
	d.MaintenanceMsg = s.MaintenanceMsg
	d.TaskID = s.TaskID
}

// See:
// - https://sqlite.org/lang_createtable.html#rowids_and_the_integer_primary_key
// - https://sqlite.org/datatype3.html
// BLOB
const schemaBotEvent = `
CREATE TABLE IF NOT EXISTS BotEvent (
	key           INTEGER PRIMARY KEY,
	schemaVersion INTEGER NOT NULL,
	botID         TEXT    NOT NULL,
	time          INTEGER NOT NULL,
	blob          BLOB    NOT NULL
) STRICT;
`

// botEventSQLBlob contains the unindexed fields.
type botEventSQLBlob struct {
	Event           string              `json:"a"`
	Message         string              `json:"b"`
	Dimensions      map[string][]string `json:"c"`
	State           string              `json:"d"`
	ExternalIP      string              `json:"e"`
	AuthenticatedAs string              `json:"f"`
	Version         string              `json:"g"`
	QuarantinedMsg  string              `json:"h"`
	MaintenanceMsg  string              `json:"i"`
	TaskID          string              `json:"j"`
}
