package messapi

import (
	"time"
)

// Generic stuff.

// Time is a time that is not strictly correctly formatted as ISO8601 because
// it has the trailing Z trimmed off.
type Time string

// CloudTime converts a time object into the formatted string.
func CloudTime(t time.Time) Time {
	return Time(t.UTC().Format("2006-01-02T15:04:05"))
}

type StringPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type StringListPair struct {
	Key    string   `json:"key"`
	Values []string `json:"value"`
}
