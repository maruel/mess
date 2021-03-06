package model

import (
	"encoding/json"
	"time"
)

// DeadAfter is the amount of time after which a bot is considered dead.
const DeadAfter = 10 * time.Minute

// Bot represents a bot as known by the server.
type Bot struct {
	Key             string              `json:"a,omitempty"`
	SchemaVersion   int                 `json:"b,omitempty"`
	Created         time.Time           `json:"c,omitempty"`
	LastSeen        time.Time           `json:"d,omitempty"`
	Version         string              `json:"e,omitempty"`
	Deleted         bool                `json:"f,omitempty"`
	Dead            bool                `json:"g,omitempty"`
	QuarantinedMsg  string              `json:"h,omitempty"`
	MaintenanceMsg  string              `json:"i,omitempty"`
	TaskID          int64               `json:"j,omitempty"`
	AuthenticatedAs string              `json:"k,omitempty"`
	Dimensions      map[string][]string `json:"l,omitempty"`
	State           []byte              `json:"m,omitempty"`
	ExternalIP      string              `json:"n,omitempty"`
}

type botSQL struct {
	key            string
	schemaVersion  int
	created        int64
	lastSeen       int64
	version        string
	deleted        bool
	dead           bool
	quarantinedMsg string
	maintenanceMsg string
	taskID         int64
	blob           []byte
}

func (b *botSQL) fields() []interface{} {
	return []interface{}{
		&b.key,
		&b.schemaVersion,
		&b.created,
		&b.lastSeen,
		&b.version,
		&b.deleted,
		&b.dead,
		&b.quarantinedMsg,
		&b.maintenanceMsg,
		&b.taskID,
		&b.blob,
	}
}

func (b *botSQL) from(d *Bot) {
	b.key = d.Key
	b.schemaVersion = d.SchemaVersion
	b.created = d.Created.UnixMicro()
	b.lastSeen = d.LastSeen.UnixMicro()
	b.version = d.Version
	b.deleted = d.Deleted
	b.dead = d.Dead
	b.quarantinedMsg = d.QuarantinedMsg
	b.maintenanceMsg = d.MaintenanceMsg
	b.taskID = d.TaskID
	s := botSQLBlob{
		AuthenticatedAs: d.AuthenticatedAs,
		Dimensions:      d.Dimensions,
		State:           d.State,
		ExternalIP:      d.ExternalIP,
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
	d.Deleted = b.deleted
	d.Dead = b.dead
	d.QuarantinedMsg = b.quarantinedMsg
	d.MaintenanceMsg = b.maintenanceMsg
	d.TaskID = b.taskID
	s := botSQLBlob{}
	if err := json.Unmarshal(b.blob, &s); err != nil {
		panic("internal error: " + err.Error())
	}
	d.AuthenticatedAs = s.AuthenticatedAs
	d.Dimensions = s.Dimensions
	d.State = s.State
	d.ExternalIP = s.ExternalIP
}

// See:
// - https://sqlite.org/lang_createtable.html#rowids_and_the_integer_primary_key
// - https://sqlite.org/datatype3.html
// BLOB
const schemaBot = `
CREATE TABLE IF NOT EXISTS Bot (
	key            TEXT    NOT NULL,
	schemaVersion  INTEGER NOT NULL,
	created        INTEGER NOT NULL,
	lastSeen       INTEGER NOT NULL,
	version        TEXT,
	deleted        INTEGER NOT NULL,
	dead           INTEGER NOT NULL,
	quarantinedMsg TEXT,
	maintenanceMsg TEXT,
	taskID         INTEGER,
	blob           BLOB    NOT NULL,
	PRIMARY KEY(key ASC)
) STRICT;
`

// botSQLBlob contains the unindexed fields.
type botSQLBlob struct {
	AuthenticatedAs string              `json:"a,omitempty"`
	Dimensions      map[string][]string `json:"b,omitempty"`
	State           []byte              `json:"c,omitempty"`
	ExternalIP      string              `json:"d,omitempty"`
}
