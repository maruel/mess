package messapi

import (
	"sort"
	"time"

	"github.com/maruel/mess/internal/model"
)

// BotsCount is /bots/count.
type BotsCount struct {
	Now         Time  `json:"now"`
	Count       int32 `json:"count"`
	Quarantined int32 `json:"quarantined"`
	Maintenance int32 `json:"maintenance"`
	Dead        int32 `json:"dead"`
	Busy        int32 `json:"busy"`
}

// BotsDimensions is /bots/dimensions.
type BotsDimensions struct {
	BotsDimensions []StringListPair `json:"bots_dimensions"`
	Now            Time             `json:"ts"`
}

// BotsList is /bots/list.
type BotsList struct {
	Cursor       string `json:"cursor"`
	Items        []Bot  `json:"items"`
	Now          Time   `json:"now"`
	DeathTimeout int    `json:"death_timeout"`
}

// Bot reports the bot state as known by the server.
type Bot struct {
	BotID           string           `json:"bot_id"`
	Dimensions      []StringListPair `json:"dimensions"`
	ExternalIP      string           `json:"external_ip"`
	AuthenticatedAs string           `json:"authenticated_as"`
	FirstSeen       Time             `json:"first_seen_ts"`
	IsDead          bool             `json:"is_dead"`
	LastSeen        Time             `json:"last_seen_ts"`
	Quarantined     bool             `json:"quarantined"`
	MaintenanceMsg  string           `json:"maintenance_msg"`
	TaskID          model.TaskID     `json:"task_id"`
	TaskName        string           `json:"task_name"`
	Version         string           `json:"version"`
	// Encoded as json since it's an arbitrary dict.
	State   string `json:"state"`
	Deleted bool   `json:"deleted"`
	// DEPRECATED: lease_id string
	// DEPRECATED: lease_expiration_ts Time
	// DEPRECATED: machine_type string
	// DEPRECATED: machine_lease string
	// DEPRECATED: leased_indefinitely bool
}

// FromDB converts the model to the API.
func (b *Bot) FromDB(m *model.Bot) {
	b.BotID = m.Key
	b.Dimensions = make([]StringListPair, len(m.Dimensions))
	i := 0
	for k, v := range m.Dimensions {
		b.Dimensions[i].Key = k
		b.Dimensions[i].Values = v
		i++
	}
	sort.Slice(b.Dimensions, func(i, j int) bool { return b.Dimensions[i].Key < b.Dimensions[j].Key })
	b.ExternalIP = m.ExternalIP
	b.AuthenticatedAs = m.AuthenticatedAs
	b.FirstSeen = CloudTime(m.Created)
	// TODO(maruel): timer.
	b.IsDead = time.Since(m.LastSeen) > 10*time.Minute
	b.LastSeen = CloudTime(m.LastSeen)
	b.Quarantined = m.QuarantinedMsg != ""
	b.MaintenanceMsg = m.MaintenanceMsg
	b.TaskID = model.ToTaskID(m.TaskID)
	// b.TaskName
	b.Version = m.Version
	b.State = string(m.State)
	// b.Deleted
}

// BotEvents is events that a bot produced.
type BotEvents struct {
	Cursor string     `json:"cursor"`
	Items  []BotEvent `json:"items"`
	Now    Time       `json:"now"`
}

// BotEvent is one event that a bot produced.
type BotEvent struct {
	Time            Time             `json:"ts"`
	Event           string           `json:"event_type"`
	Message         string           `json:"message"`
	Dimensions      []StringListPair `json:"dimensions"`
	State           string           `json:"state"`
	ExternalIP      string           `json:"external_ip"`
	AuthenticatedAs string           `json:"authenticated_as"`
	Version         string           `json:"version"`
	Quarantined     bool             `json:"quarantined"`
	MaintenanceMsg  string           `json:"maintenance_msg"`
	TaskID          model.TaskID     `json:"task_id"`
}

// FromDB converts the model to the API.
func (b *BotEvent) FromDB(m *model.BotEvent) {
	b.Time = CloudTime(m.Time)
	b.Event = m.Event
	b.Message = m.Message
	b.Dimensions = make([]StringListPair, 0, len(m.Dimensions))
	for k, v := range m.Dimensions {
		b.Dimensions = append(b.Dimensions, StringListPair{Key: k, Values: v})
	}
	sort.Slice(b.Dimensions, func(i, j int) bool { return b.Dimensions[i].Key < b.Dimensions[j].Key })
	b.State = string(m.State)
	b.ExternalIP = m.ExternalIP
	b.AuthenticatedAs = m.AuthenticatedAs
	b.Version = m.Version
	b.Quarantined = m.QuarantinedMsg != ""
	b.MaintenanceMsg = m.MaintenanceMsg
	b.TaskID = model.ToTaskID(m.TaskID)
}

// BotTasks is /bot/tasks
type BotTasks struct {
	Cursor string       `json:"cursor"`
	Items  []TaskResult `json:"items"`
	Now    Time         `json:"now"`
}
