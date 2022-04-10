package model

import (
	"encoding/json"
	"time"
)

// TaskResult is the result of running a TaskRequest.
type TaskResult struct {
	Key              int64               `json:"a"`
	SchemaVersion    int                 `json:"b"`
	BotID            string              `json:"c"`
	BotVersion       string              `json:"d"`
	BotDimensions    map[string][]string `json:"e"`
	BotIdleSince     time.Duration       `json:"f"`
	ServerVersions   []string            `json:"g"`
	CurrentTaskSlice int32               `json:"h"`
	DedupedFrom      int64               `json:"i"`
	PropertiesHash   string              `json:"j"`
	TaskOutput       TaskOutput          `json:"k"`
	ExitCode         int32               `json:"l"`
	InternalFailure  string              `json:"m"`
	State            TaskState           `json:"n"`
	Children         []int64             `json:"o"`
	Output           Digest              `json:"p"`
	CIPDClientUsed   CIPDPackage         `json:"q"`
	CIPDPins         []CIPDPackage       `json:"r"`
	ResultDB         ResultDB            `json:"s"`
	Duration         time.Duration       `json:"t"`
	Started          time.Time           `json:"u"`
	Completed        time.Time           `json:"v"`
	Abandoned        time.Time           `json:"w"`
	Modified         time.Time           `json:"x"`
	Cost             float64             `json:"y"`
	Killing          bool                `json:"z"`
	DeadAfter        time.Time           `json:"aa"`
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
		BotVersion:       t.BotVersion,
		BotDimensions:    t.BotDimensions,
		BotIdleSince:     t.BotIdleSince,
		ServerVersions:   t.ServerVersions,
		CurrentTaskSlice: t.CurrentTaskSlice,
		DedupedFrom:      t.DedupedFrom,
		PropertiesHash:   t.PropertiesHash,
		TaskOutput:       t.TaskOutput,
		ExitCode:         t.ExitCode,
		InternalFailure:  t.InternalFailure,
		State:            t.State,
		Children:         t.Children,
		Output:           t.Output,
		CIPDClientUsed:   t.CIPDClientUsed,
		CIPDPins:         t.CIPDPins,
		ResultDB:         t.ResultDB,
		Duration:         t.Duration,
		Started:          t.Started,
		Completed:        t.Completed,
		Abandoned:        t.Abandoned,
		Modified:         t.Modified,
		Cost:             t.Cost,
		Killing:          t.Killing,
		DeadAfter:        t.DeadAfter,
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
	t.BotDimensions = b.BotDimensions
	t.BotIdleSince = b.BotIdleSince
	t.ServerVersions = b.ServerVersions
	t.CurrentTaskSlice = b.CurrentTaskSlice
	t.DedupedFrom = b.DedupedFrom
	t.PropertiesHash = b.PropertiesHash
	t.TaskOutput = b.TaskOutput
	t.ExitCode = b.ExitCode
	t.InternalFailure = b.InternalFailure
	t.State = b.State
	t.Children = b.Children
	t.Output = b.Output
	t.CIPDClientUsed = b.CIPDClientUsed
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
	key           INTEGER NOT NULL,
	schemaVersion INTEGER NOT NULL,
	botID         TEXT NOT NULL,
	blob          BLOB    NOT NULL,
	PRIMARY KEY(key DESC)
) STRICT;
`

// taskResultSQLBlob contains the unindexed fields.
type taskResultSQLBlob struct {
	BotVersion       string              `json:"a"`
	BotDimensions    map[string][]string `json:"b"`
	BotIdleSince     time.Duration       `json:"c"`
	ServerVersions   []string            `json:"d"`
	CurrentTaskSlice int32               `json:"e"`
	DedupedFrom      int64               `json:"f"`
	PropertiesHash   string              `json:"g"`
	TaskOutput       TaskOutput          `json:"h"`
	ExitCode         int32               `json:"i"`
	InternalFailure  string              `json:"j"`
	State            TaskState           `json:"k"`
	Children         []int64             `json:"l"`
	Output           Digest              `json:"m"`
	CIPDClientUsed   CIPDPackage         `json:"n"`
	CIPDPins         []CIPDPackage       `json:"o"`
	ResultDB         ResultDB            `json:"p"`
	Duration         time.Duration       `json:"q"`
	Started          time.Time           `json:"r"`
	Completed        time.Time           `json:"s"`
	Abandoned        time.Time           `json:"t"`
	Modified         time.Time           `json:"u"`
	Cost             float64             `json:"v"`
	Killing          bool                `json:"w"`
	DeadAfter        time.Time           `json:"x"`
}

// TaskState is the state of the task request.
type TaskState int64

// Valid TaskState.
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

// ResultDB declares the LUCI ResultDB information.
type ResultDB struct {
	Host       string `json:"a"`
	Invocation string `json:"b"`
}

// TaskOutput stores the task's output.
type TaskOutput struct {
	Size int64 `json:"a"`
}
