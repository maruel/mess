package messapi

import (
	"math"
	"strconv"
	"strings"
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

// StringPair is a key value item.
type StringPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// StringListPair is a key values item.
type StringListPair struct {
	Key    string   `json:"key"`
	Values []string `json:"value"`
}

// ThreeState is an optional value.
type ThreeState int

// Valid ThreeState.
const (
	ThreeStateFalse ThreeState = 1
	ThreeStateTrue  ThreeState = 2
	ThreeStateNone  ThreeState = 3
)

// ToThreeState parses a string passed as a HTTP GET query argument.
func ToThreeState(v string) ThreeState {
	switch strings.ToLower(v) {
	case "true", "2":
		return ThreeStateTrue
	case "false", "1":
		return ThreeStateFalse
	default:
		return ThreeStateNone
	}
}

// ToInt64 parses a string passed as a HTTP GET query argument.
func ToInt64(v string, def int64) int64 {
	if v == "" {
		return def
	}
	i, err := strconv.ParseInt(v, 64, 10)
	if err != nil {
		return def
	}
	return i
}

// ToTime parses a string passed as a HTTP GET query argument.
func ToTime(v string) time.Time {
	if v == "" {
		return time.Time{}
	}
	i, err := strconv.ParseFloat(v, 64)
	if err != nil || i == 0 {
		return time.Time{}
	}
	sec, dec := math.Modf(i)
	return time.Unix(int64(sec), int64(dec*(1e9)))
}
