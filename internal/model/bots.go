package model

import (
	"time"
)

type Bot struct {
	SchemaVersion int       `json:"a"`
	Key           string    `json:"b"`
	Create        time.Time `json:"c"`
	LastSeen      time.Time `json:"d"`
	Version       string    `json:"e"`
	Blob          BotBlob   `json:"f"`
}

// BotBlob contains the unindexed fields.
type BotBlob struct {
	Dimensions map[string][]string `json:"a"`
	Events     []*BotEvent         `json:"b"`
}

// TODO(maruel): Make it its own table.
func (b *Bot) AddEvent(now time.Time, event, msg string) {
	// Make a copy of the map but not the values, since they are immutable (to
	// save memory).
	dims := make(map[string][]string, len(b.Blob.Dimensions))
	for k, v := range b.Blob.Dimensions {
		dims[k] = v
	}
	b.Blob.Events = append(b.Blob.Events, &BotEvent{
		SchemaVersion: 1,
		Key:           0,
		BotID:         b.Key,
		Time:          now,
		Blob: BotEventBlob{
			Event:      event,
			Message:    msg,
			Dimensions: dims,
			Version:    b.Version,
			// TODO(maruel): Add more.
		},
	})
}

type BotEvent struct {
	SchemaVersion int          `json:"a"`
	Key           int          `json:"b"`
	BotID         string       `json:"c"`
	Time          time.Time    `json:"d"`
	Blob          BotEventBlob `json:"e"`
}

// BotEventBlob contains the unindexed fields.
type BotEventBlob struct {
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
