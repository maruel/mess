package main

import (
	"time"
)

type Bot struct {
	Create     time.Time
	LastSeen   time.Time
	Dimensions map[string][]string
	Version    string
	// TODO(maruel): Keep compressed in memory?
	Events []*BotEvent
}

func (b *Bot) addEvent(now time.Time, event, msg string) {
	// Make a copy of the map but not the values, since they are immutable (to
	// save memory).
	dims := make(map[string][]string, len(b.Dimensions))
	for k, v := range b.Dimensions {
		dims[k] = v
	}
	b.Events = append(b.Events, &BotEvent{
		Time:       now,
		Event:      event,
		Message:    msg,
		Dimensions: dims,
		Version:    b.Version,
		// TODO(maruel): Add more.
	})
}

type BotEvent struct {
	Time            time.Time
	Event           string
	Message         string
	Dimensions      map[string][]string
	State           string
	ExternalIP      string
	AuthenticatedAs string
	Version         string
	QuarantinedMsg  string
	MaintenanceMsg  string
	TaskID          string
}
