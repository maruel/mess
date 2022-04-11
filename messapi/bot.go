package messapi

import (
	"time"

	"github.com/maruel/mess/internal/model"
)

// BotsCountRequest is /bots/count (GET).
type BotsCountRequest struct {
	// Dimensions must be a list of 'key:value' strings to filter the returned
	// list of bots on.
	Dimensions []string
}

// BotsCountResponse is /bots/count (GET).
type BotsCountResponse struct {
	Now         Time  `json:"now,omitempty"`
	Count       int64 `json:"count,omitempty"`
	Quarantined int64 `json:"quarantined,omitempty"`
	Maintenance int64 `json:"maintenance,omitempty"`
	Dead        int64 `json:"dead,omitempty"`
	Busy        int64 `json:"busy,omitempty"`
}

// BotsDimensionsRequest is /bots/dimensions (GET).
type BotsDimensionsRequest struct {
	Pool []string
}

// BotsDimensionsResponse is /bots/dimensions (GET).
type BotsDimensionsResponse struct {
	BotsDimensions []StringListPair `json:"bots_dimensions,omitempty"`
	Now            Time             `json:"ts,omitempty"`
}

// BotsListRequest is /bots/list (GET).
type BotsListRequest struct {
	Limit  int64
	Cursor string
	// Dimensions must be a list of 'key:value' strings to filter the returned
	// list of bots on.
	Dimensions    []string
	Quarantined   ThreeState
	InMaintenance ThreeState
	IsDead        ThreeState
	IsBusy        ThreeState
}

// BotsListResponse is /bots/list (GET).
type BotsListResponse struct {
	Cursor       string `json:"cursor,omitempty"`
	Items        []Bot  `json:"items,omitempty"`
	Now          Time   `json:"now,omitempty"`
	DeathTimeout int64  `json:"death_timeout,omitempty"`
}

// BotDeleteResponse is /bot/<id>/delete (POST).
type BotDeleteResponse struct {
	Deleted bool `json:"deleted,omitempty"`
}

// BotEventsRequest is /bot/<id>/events (GET).
type BotEventsRequest struct {
	Limit  int64
	Cursor string
	End    time.Time
	Start  time.Time
}

// BotEventsResponse is /bot/<id>/events (GET).
type BotEventsResponse struct {
	Cursor string     `json:"cursor,omitempty"`
	Items  []BotEvent `json:"items,omitempty"`
	Now    Time       `json:"now,omitempty"`
}

// BotGetResponse is /bot/<id>/get (GET).
type BotGetResponse = Bot

// BotTasksRequest is /bot/<id>/tasks (GET).
type BotTasksRequest struct {
	Limit                   int64
	Cursor                  string
	End                     time.Time
	Start                   time.Time
	State                   string // TaskStateQuery default=ALL
	Sort                    string
	IncludePerformanceStats bool
}

// BotTasksResponse is /bot/<id>/tasks (GET).
type BotTasksResponse struct {
	Cursor string       `json:"cursor,omitempty"`
	Items  []TaskResult `json:"items,omitempty"`
	Now    Time         `json:"now,omitempty"`
}

// BotTerminateResponse is /bot/<id>/terminate (POST).
type BotTerminateResponse struct {
	TaskID model.TaskID `json:"task_id,omitempty"`
}

//

// Bot reports the bot state as known by the server.
type Bot struct {
	BotID           string           `json:"bot_id,omitempty"`
	Dimensions      []StringListPair `json:"dimensions,omitempty"`
	ExternalIP      string           `json:"external_ip,omitempty"`
	AuthenticatedAs string           `json:"authenticated_as,omitempty"`
	FirstSeen       Time             `json:"first_seen_ts,omitempty"`
	IsDead          bool             `json:"is_dead,omitempty"`
	LastSeen        Time             `json:"last_seen_ts,omitempty"`
	Quarantined     bool             `json:"quarantined,omitempty"`
	MaintenanceMsg  string           `json:"maintenance_msg,omitempty"`
	TaskID          model.TaskID     `json:"task_id,omitempty"`
	TaskName        string           `json:"task_name,omitempty"`
	Version         string           `json:"version,omitempty"`
	// Encoded as json since it's an arbitrary dict.
	State   string `json:"state,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
	// DEPRECATED: lease_id string
	// DEPRECATED: lease_expiration_ts Time
	// DEPRECATED: machine_type string
	// DEPRECATED: machine_lease string
	// DEPRECATED: leased_indefinitely bool
}

// FromDB converts the model to the API.
func (b *Bot) FromDB(m *model.Bot) {
	b.BotID = m.Key
	b.Dimensions = ToStringListPairs(m.Dimensions)
	b.ExternalIP = m.ExternalIP
	b.AuthenticatedAs = m.AuthenticatedAs
	b.FirstSeen = CloudTime(m.Created)
	b.IsDead = m.Dead
	b.LastSeen = CloudTime(m.LastSeen)
	b.Quarantined = m.QuarantinedMsg != ""
	b.MaintenanceMsg = m.MaintenanceMsg
	b.TaskID = model.ToTaskID(m.TaskID)
	// TODO(maruel): b.TaskName
	b.Version = m.Version
	b.State = string(m.State)
	b.Deleted = b.Deleted
}

// BotEvents is events that a bot produced.
type BotEvents struct {
	Cursor string     `json:"cursor,omitempty"`
	Items  []BotEvent `json:"items,omitempty"`
	Now    Time       `json:"now,omitempty"`
}

// BotEvent is one event that a bot produced.
type BotEvent struct {
	Time            Time             `json:"ts,omitempty"`
	Event           string           `json:"event_type,omitempty"`
	Message         string           `json:"message,omitempty"`
	Dimensions      []StringListPair `json:"dimensions,omitempty"`
	State           string           `json:"state,omitempty"`
	ExternalIP      string           `json:"external_ip,omitempty"`
	AuthenticatedAs string           `json:"authenticated_as,omitempty"`
	Version         string           `json:"version,omitempty"`
	Quarantined     bool             `json:"quarantined,omitempty"`
	MaintenanceMsg  string           `json:"maintenance_msg,omitempty"`
	TaskID          model.TaskID     `json:"task_id,omitempty"`
}

// FromDB converts the model to the API.
func (b *BotEvent) FromDB(m *model.BotEvent) {
	b.Time = CloudTime(m.Time)
	b.Event = m.Event
	b.Message = m.Message
	b.Dimensions = ToStringListPairs(m.Dimensions)
	b.State = string(m.State)
	b.ExternalIP = m.ExternalIP
	b.AuthenticatedAs = m.AuthenticatedAs
	b.Version = m.Version
	b.Quarantined = m.QuarantinedMsg != ""
	b.MaintenanceMsg = m.MaintenanceMsg
	b.TaskID = model.ToTaskID(m.TaskID)
}
