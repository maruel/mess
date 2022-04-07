package model

import "time"

type TaskResult struct {
	SchemaVersion int            `json:"a"`
	Key           int64          `json:"b"`
	BotID         string         `json:"c"`
	Blob          TaskResultBlob `json:"d"`
}

type TaskResultBlob struct {
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

	// Internal flags.
	Killing   bool      `json:"u"`
	DeadAfter time.Time `json:"v"`
	// TODO(maruel): Stats.
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
