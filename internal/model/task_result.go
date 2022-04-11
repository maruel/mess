package model

import (
	"encoding/json"
	"time"
)

// TaskResult is the result of running a TaskRequest.
type TaskResult struct {
	Key              int64               `json:"a,omitempty"`
	SchemaVersion    int                 `json:"b,omitempty"`
	BotID            string              `json:"c,omitempty"`
	BotVersion       string              `json:"d,omitempty"`
	BotDimensions    map[string][]string `json:"e,omitempty"`
	BotIdleSince     time.Duration       `json:"f,omitempty"`
	ServerVersions   []string            `json:"g,omitempty"`
	CurrentTaskSlice int32               `json:"h,omitempty"`
	DedupedFrom      int64               `json:"i,omitempty"`
	PropertiesHash   string              `json:"j,omitempty"`
	TaskOutput       TaskOutput          `json:"k,omitempty"`
	ExitCode         int32               `json:"l,omitempty"`
	InternalFailure  string              `json:"m,omitempty"`
	State            TaskState           `json:"n,omitempty"`
	Children         []int64             `json:"o,omitempty"`
	Output           Digest              `json:"p,omitempty"`
	CIPDClientUsed   CIPDPackage         `json:"q,omitempty"`
	CIPDPins         []CIPDPackage       `json:"r,omitempty"`
	ResultDB         ResultDB            `json:"s,omitempty"`
	Duration         time.Duration       `json:"t,omitempty"`
	Started          time.Time           `json:"u,omitempty"`
	Completed        time.Time           `json:"v,omitempty"`
	Abandoned        time.Time           `json:"w,omitempty"`
	Modified         time.Time           `json:"x,omitempty"`
	Cost             float64             `json:"y,omitempty"`
	Killing          bool                `json:"z,omitempty"`
	DeadAfter        time.Time           `json:"aa,omitempty"`
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
	BotVersion       string              `json:"a,omitempty"`
	BotDimensions    map[string][]string `json:"b,omitempty"`
	BotIdleSince     time.Duration       `json:"c,omitempty"`
	ServerVersions   []string            `json:"d,omitempty"`
	CurrentTaskSlice int32               `json:"e,omitempty"`
	DedupedFrom      int64               `json:"f,omitempty"`
	PropertiesHash   string              `json:"g,omitempty"`
	TaskOutput       TaskOutput          `json:"h,omitempty"`
	ExitCode         int32               `json:"i,omitempty"`
	InternalFailure  string              `json:"j,omitempty"`
	State            TaskState           `json:"k,omitempty"`
	Children         []int64             `json:"l,omitempty"`
	Output           Digest              `json:"m,omitempty"`
	CIPDClientUsed   CIPDPackage         `json:"n,omitempty"`
	CIPDPins         []CIPDPackage       `json:"o,omitempty"`
	ResultDB         ResultDB            `json:"p,omitempty"`
	Duration         time.Duration       `json:"q,omitempty"`
	Started          time.Time           `json:"r,omitempty"`
	Completed        time.Time           `json:"s,omitempty"`
	Abandoned        time.Time           `json:"t,omitempty"`
	Modified         time.Time           `json:"u,omitempty"`
	Cost             float64             `json:"v,omitempty"`
	Killing          bool                `json:"w,omitempty"`
	DeadAfter        time.Time           `json:"x,omitempty"`
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
	Host       string `json:"a,omitempty"`
	Invocation string `json:"b,omitempty"`
}

// TaskOutput stores the task's output.
type TaskOutput struct {
	Size int64 `json:"a,omitempty"`
}
