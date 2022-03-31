package main

import "time"

type StringListPair struct {
	Key    string   `json:"key"`
	Values []string `json:"value"`
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

// Store a reference to disk.
type TaskOutput struct {
	Size int64
}

type ResultDB struct {
	Host       string
	Invocation string
}

type TaskResult struct {
	Key            int64
	BotID          string
	BotVersion     string
	BotDimension   []StringListPair
	BotIdleSince   time.Duration
	ServerVersions []string
	TaskSlice      int64
	DedupedFrom    int64
	PropertiesHash string

	TaskOutput TaskOutput
	ExitCode   int64
	State      TaskState
	Children   []int64
	Output     Digest
	CIPDPins   []CIPDPackage
	ResultDB   ResultDB

	Duration  time.Duration
	Started   time.Time
	Completed time.Time
	Abandoned time.Time
	Modified  time.Time
	Cost      float64

	// Internal flags.
	Killing   bool
	DeadAfter time.Time
	// TODO(maruel): Stats.
}
