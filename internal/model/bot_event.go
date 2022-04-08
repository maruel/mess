package model

import (
	"encoding/json"
	"time"
)

// BotEvent is an event on a bot.
type BotEvent struct {
	Key           int64  `json:"a"`
	SchemaVersion int    `json:"b"`
	BotID         string `json:"c"`
	// Information about the event.
	Time    time.Time `json:"d"`
	Event   string    `json:"e"`
	Message string    `json:"f"`
	// Information copied for the bot.
	Version         string              `json:"g"`
	AuthenticatedAs string              `json:"h"`
	Dimensions      map[string][]string `json:"i"`
	State           []byte              `json:"j"`
	ExternalIP      string              `json:"k"`
	TaskID          int64               `json:"l"`
	QuarantinedMsg  string              `json:"m"`
	MaintenanceMsg  string              `json:"n"`
}

// InitFrom initializes a BotEvent from a bot.
func (e *BotEvent) InitFrom(b *Bot, now time.Time, event, msg string) {
	e.Key = 0
	e.SchemaVersion = 1
	e.BotID = b.Key
	e.Time = now
	e.Event = event
	e.Message = msg
	e.Version = b.Version
	e.AuthenticatedAs = b.AuthenticatedAs
	// Make a copy of the map but not the values, since they are immutable (to
	// save memory).
	e.Dimensions = make(map[string][]string, len(b.Dimensions))
	for k, v := range b.Dimensions {
		e.Dimensions[k] = v
	}
	e.State = b.State
	e.ExternalIP = b.ExternalIP
	e.TaskID = b.TaskID
	e.QuarantinedMsg = b.QuarantinedMsg
	e.MaintenanceMsg = b.MaintenanceMsg
}

type botEventSQL struct {
	key           int64
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
		Version:         d.Version,
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
	d.Version = s.Version
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
const schemaBotEvent = `
CREATE TABLE IF NOT EXISTS BotEvent (
	key           INTEGER NOT NULL,
	schemaVersion INTEGER NOT NULL,
	botID         TEXT    NOT NULL,
	time          INTEGER NOT NULL,
	blob          BLOB    NOT NULL,
	PRIMARY KEY(key DESC)
) STRICT;
`

// botEventSQLBlob contains the unindexed fields.
type botEventSQLBlob struct {
	Event           string              `json:"a"`
	Message         string              `json:"b"`
	Version         string              `json:"c"`
	AuthenticatedAs string              `json:"d"`
	Dimensions      map[string][]string `json:"e"`
	State           []byte              `json:"f"`
	ExternalIP      string              `json:"g"`
	TaskID          int64               `json:"h"`
	QuarantinedMsg  string              `json:"i"`
	MaintenanceMsg  string              `json:"j"`
}
