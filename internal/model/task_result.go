package model

import (
	"encoding/json"
	"time"
)

type TaskResult struct {
	Key            int64               `json:"a"`
	SchemaVersion  int                 `json:"b"`
	BotID          string              `json:"c"`
	BotVersion     string              `json:"d"`
	BotDimension   map[string][]string `json:"e"`
	BotIdleSince   time.Duration       `json:"f"`
	ServerVersions []string            `json:"g"`
	TaskSlice      int64               `json:"h"`
	DedupedFrom    int64               `json:"i"`
	PropertiesHash string              `json:"j"`
	TaskOutput     TaskOutput          `json:"k"`
	ExitCode       int64               `json:"l"`
	State          TaskState           `json:"m"`
	Children       []int64             `json:"n"`
	Output         Digest              `json:"o"`
	CIPDPins       []CIPDPackage       `json:"p"`
	ResultDB       ResultDB            `json:"q"`
	Duration       time.Duration       `json:"r"`
	Started        time.Time           `json:"s"`
	Completed      time.Time           `json:"t"`
	Abandoned      time.Time           `json:"u"`
	Modified       time.Time           `json:"v"`
	Cost           float64             `json:"w"`
	Killing        bool                `json:"x"`
	DeadAfter      time.Time           `json:"y"`
}

type taskResultSQL struct {
	key           int64
	schemaVersion int
	botID         string
	blob          []byte
}

func (r *taskResultSQL) fields() []interface{} {
	return []interface{}{
		&r.key,
		&r.schemaVersion,
		&r.botID,
		&r.blob,
	}
}

func (r *taskResultSQL) from(t *TaskResult) {
	r.key = t.Key
	r.schemaVersion = t.SchemaVersion
	r.botID = t.BotID
	b := taskResultSQLBlob{
		BotVersion:     t.BotVersion,
		BotDimension:   t.BotDimension,
		BotIdleSince:   t.BotIdleSince,
		ServerVersions: t.ServerVersions,
		TaskSlice:      t.TaskSlice,
		DedupedFrom:    t.DedupedFrom,
		PropertiesHash: t.PropertiesHash,
		TaskOutput:     t.TaskOutput,
		ExitCode:       t.ExitCode,
		State:          t.State,
		Children:       t.Children,
		Output:         t.Output,
		CIPDPins:       t.CIPDPins,
		ResultDB:       t.ResultDB,
		Duration:       t.Duration,
		Started:        t.Started,
		Completed:      t.Completed,
		Abandoned:      t.Abandoned,
		Modified:       t.Modified,
		Cost:           t.Cost,
		Killing:        t.Killing,
		DeadAfter:      t.DeadAfter,
	}
	var err error
	r.blob, err = json.Marshal(&b)
	if err != nil {
		panic("internal error: " + err.Error())
	}
}

func (r *taskResultSQL) to(t *TaskResult) {
	t.Key = r.key
	t.SchemaVersion = r.schemaVersion
	t.BotID = r.botID
	b := taskResultSQLBlob{}
	if err := json.Unmarshal(r.blob, &b); err != nil {
		panic("internal error: " + err.Error())
	}
	t.BotVersion = b.BotVersion
	t.BotDimension = b.BotDimension
	t.BotIdleSince = b.BotIdleSince
	t.ServerVersions = b.ServerVersions
	t.TaskSlice = b.TaskSlice
	t.DedupedFrom = b.DedupedFrom
	t.PropertiesHash = b.PropertiesHash
	t.TaskOutput = b.TaskOutput
	t.ExitCode = b.ExitCode
	t.State = b.State
	t.Children = b.Children
	t.Output = b.Output
	t.CIPDPins = b.CIPDPins
	t.ResultDB = b.ResultDB
	t.Duration = b.Duration
	t.Started = b.Started
	t.Completed = b.Completed
	t.Abandoned = b.Abandoned
	t.Modified = b.Modified
	t.Cost = b.Cost
	t.Killing = b.Killing
	t.DeadAfter = b.DeadAfter
}

// See:
// - https://sqlite.org/lang_createtable.html#rowids_and_the_integer_primary_key
// - https://sqlite.org/datatype3.html
// BLOB
const schemaTaskResult = `
CREATE TABLE IF NOT EXISTS TaskResult (
	key           INTEGER PRIMARY KEY,
	schemaVersion INTEGER NOT NULL,
	botID         TEXT NOT NULL,
	blob          BLOB    NOT NULL
) STRICT;
`

// taskResultSQLBlob contains the unindexed fields.
type taskResultSQLBlob struct {
	BotVersion     string              `json:"a"`
	BotDimension   map[string][]string `json:"b"`
	BotIdleSince   time.Duration       `json:"c"`
	ServerVersions []string            `json:"d"`
	TaskSlice      int64               `json:"e"`
	DedupedFrom    int64               `json:"f"`
	PropertiesHash string              `json:"g"`
	TaskOutput     TaskOutput          `json:"h"`
	ExitCode       int64               `json:"i"`
	State          TaskState           `json:"j"`
	Children       []int64             `json:"k"`
	Output         Digest              `json:"l"`
	CIPDPins       []CIPDPackage       `json:"m"`
	ResultDB       ResultDB            `json:"n"`
	Duration       time.Duration       `json:"o"`
	Started        time.Time           `json:"p"`
	Completed      time.Time           `json:"q"`
	Abandoned      time.Time           `json:"r"`
	Modified       time.Time           `json:"s"`
	Cost           float64             `json:"t"`
	Killing        bool                `json:"u"`
	DeadAfter      time.Time           `json:"v"`
}

type TaskState int64

const (
	Running TaskState = iota
	Pending
	Expired
	Timedout
	BotDied
	Canceled
	Completed
	Killed
	NoResource
)

type ResultDB struct {
	Host       string `json:"a"`
	Invocation string `json:"b"`
}

// Store a reference to disk.
type TaskOutput struct {
	Size int64 `json:"a"`
}
