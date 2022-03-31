package main

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/maruel/mess/internal"
)

func (s *server) apiEndpoint(w http.ResponseWriter, r *http.Request) {
	// TODO(maruel): Cleaner code.
	if r.URL.Path == "/server/details" {
		sendJSONResponse(w, serverDetails{
			ServerVersion: serverVersion,
			BotVersion:    internal.GetBotVersion(getHost(r)),
		})
		return
	}
	if r.URL.Path == "/server/permissions" {
		sendJSONResponse(w, serverPermissions{
			DeleteBot: true,
		})
		return
	}
	// TODO(maruel): /server/token.

	cloudNow := toCloudTime(time.Now())
	if r.URL.Path == "/bots/count" {
		// TODO(maruel): Implement.
		s.tables.mu.Lock()
		count := len(s.tables.Bots)
		s.tables.mu.Unlock()
		sendJSONResponse(w, botsCount{
			Now:         cloudNow,
			Count:       int32(count),
			Quarantined: 0,
			Maintenance: 0,
			Dead:        0,
			Busy:        0,
		})
		return
	}
	if r.URL.Path == "/bots/dimensions" {
		// TODO(maruel): Arguments.
		// TODO(maruel): Implement.
		sendJSONResponse(w, botsDimensions{
			BotsDimensions: []StringListPair{},
			Now:            cloudNow,
		})
		return
	}
	if r.URL.Path == "/bots/list" {
		// TODO(maruel): Arguments.
		// TODO(maruel): Implement.
		s.tables.mu.Lock()
		items := make([]botInfo, len(s.tables.Bots))
		i := 0
		for id, b := range s.tables.Bots {
			items[i].fromDB(id, b)
			i++
		}
		s.tables.mu.Unlock()
		sendJSONResponse(w, botsList{
			Cursor:       "",
			Items:        items,
			Now:          cloudNow,
			DeathTimeout: 30,
		})
		return
	}
	// TODO(maruel): /bots/...

	if r.URL.Path == "/tasks/count" {
		// TODO(maruel): Arguments.
		// TODO(maruel): Implement.
		s.tables.mu.Lock()
		count := len(s.tables.TasksRequest)
		s.tables.mu.Unlock()
		sendJSONResponse(w, tasksCount{
			Count: int32(count),
			Now:   cloudNow,
		})
		return
	}
	if r.URL.Path == "/tasks/list" {
		// TODO(maruel): Arguments.
		// TODO(maruel): Implement.
		sendJSONResponse(w, tasksList{
			Cursor: "",
			Items:  []TaskResult{},
			Now:    cloudNow,
		})
		return
	}
	// TODO(maruel): /tasks/...

	if strings.HasPrefix(r.URL.Path, "/bot/") {
		if n := strings.SplitN(r.URL.Path[len("/bot/"):], "/", 2); len(n) == 2 {
			id := n[0]
			switch n[1] {
			case "delete":
			case "events":
				s.tables.mu.Lock()
				var events []botEvent
				if bot := s.tables.Bots[id]; bot != nil {
					events = make([]botEvent, len(bot.Events))
					for i, be := range bot.Events {
						events[i].fromDB(be)
					}
				} else {
					var be [0]botEvent
					events = be[:]
				}
				s.tables.mu.Unlock()
				sendJSONResponse(w, botEvents{
					Cursor: "",
					Items:  events,
					Now:    cloudNow,
				})
				return
			case "get":
				s.tables.mu.Lock()
				bi := botInfo{}
				if b := s.tables.Bots[id]; b != nil {
					bi.fromDB(id, b)
				}
				s.tables.mu.Unlock()
				sendJSONResponse(w, bi)
				return
			case "tasks":
				sendJSONResponse(w, botTasks{
					Cursor: "",
					Items:  []TaskResult{},
					Now:    cloudNow,
				})
				return
			case "terminate":
			}
		}
	}

	// /task
	// /queues
	// /config
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

// serverDetails reports details about the server.
type serverDetails struct {
	ServerVersion string `json:"server_version"`
	BotVersion    string `json:"bot_version"`
	// DEPRECATED: MachineProviderTemplate  string `json:"machine_provider_template"`
	// DEPRECATED: DisplayServerURLTemplate string `json:"display_server_url_template"`
	LUCIConfig      string `json:"luci_config"`
	CASViewerServer string `json:"cas_viewer_server"`
}

// serverPermissions reports the client's permissions.
type serverPermissions struct {
	DeleteBot    bool `json:"delete_bot"`
	DeleteBots   bool `json:"delete_bots"`
	TerminateBot bool `json:"terminate_bot"`
	// DEPRECATED: GetConfigs   bool `json:"get_configs"`
	// DEPRECATED: PutConfigs   bool `json:"put_configs"`
	// Cancel one single task
	CancelTask        bool `json:"cancel_task"`
	GetBootstrapToken bool `json:"get_bootstrap_token"`
	// Cancel multiple tasks at once, usually in emergencies.
	CancelTasks bool     `json:"cancel_tasks"`
	ListBots    []string `json:"list_bots"`
	ListTasks   []string `json:"list_tasks"`
}

// botsCount reports the bots count.
type botsCount struct {
	Now         cloudTime `json:"now"`
	Count       int32     `json:"count"`
	Quarantined int32     `json:"quarantined"`
	Maintenance int32     `json:"maintenance"`
	Dead        int32     `json:"dead"`
	Busy        int32     `json:"busy"`
}

type botsList struct {
	Cursor       string    `json:"cursor"`
	Items        []botInfo `json:"items"`
	Now          cloudTime `json:"now"`
	DeathTimeout int       `json:"death_timeout"`
}

// botInfo reports the bot state as known by the server.
type botInfo struct {
	BotID           string           `json:"bot_id"`
	Dimensions      []StringListPair `json:"dimensions"`
	ExternalIP      string           `json:"external_ip"`
	AuthenticatedAs string           `json:"authenticated_as"`
	FirstSeen       cloudTime        `json:"first_seen_ts"`
	IsDead          bool             `json:"is_dead"`
	LastSeen        cloudTime        `json:"last_seen_ts"`
	Quarantined     bool             `json:"quarantined"`
	MaintenanceMsg  string           `json:"maintenance_msg"`
	TaskID          string           `json:"task_id"`
	TaskName        string           `json:"task_name"`
	Version         string           `json:"version"`
	// Encoded as json since it's an arbitrary dict.
	State   string `json:"state"`
	Deleted bool   `json:"deleted"`
	// DEPRECATED: lease_id string
	// DEPRECATED: lease_expiration_ts cloudTime
	// DEPRECATED: machine_type string
	// DEPRECATED: machine_lease string
	// DEPRECATED: leased_indefinitely bool
}

func (b *botInfo) fromDB(id string, bot *Bot) {
	b.BotID = id
	// b.Dimensions
	// b.ExternalIP
	// b.AuthenticatedAs
	b.FirstSeen = toCloudTime(bot.Create)
	// b.IsDead
	b.LastSeen = toCloudTime(bot.LastSeen)
	// b.Quarantined
	// b.MaintenanceMsg
	// b.TaskID
	// b.TaskName
	b.Version = bot.Version
	// b.State
	// b.Deleted
}

type botsDimensions struct {
	BotsDimensions []StringListPair `json:"bots_dimensions"`
	Now            cloudTime        `json:"ts"`
}

type tasksCount struct {
	Count int32     `json:"count"`
	Now   cloudTime `json:"now"`
}

type tasksList struct {
	Cursor string       `json:"cursor"`
	Items  []TaskResult `json:"items"`
	Now    cloudTime    `json:"now"`
}

type botEvents struct {
	Cursor string     `json:"cursor"`
	Items  []botEvent `json:"items"`
	Now    cloudTime  `json:"now"`
}

type botEvent struct {
	Time            cloudTime        `json:"ts"`
	Event           string           `json:"event_type"`
	Message         string           `json:"message"`
	Dimensions      []StringListPair `json:"dimensions"`
	State           string           `json:"state"`
	ExternalIP      string           `json:"external_ip"`
	AuthenticatedAs string           `json:"authenticated_as"`
	Version         string           `json:"version"`
	Quarantined     bool             `json:"quarantined"`
	MaintenanceMsg  string           `json:"maintenance_msg"`
	TaskID          string           `json:"task_id"`
}

func (b *botEvent) fromDB(be *BotEvent) {
	b.Time = toCloudTime(be.Time)
	b.Event = be.Event
	b.Message = be.Message
	b.Dimensions = make([]StringListPair, 0, len(be.Dimensions))
	for k, v := range be.Dimensions {
		b.Dimensions = append(b.Dimensions, StringListPair{Key: k, Values: v})
	}
	sort.Slice(b.Dimensions, func(i, j int) bool { return b.Dimensions[i].Key < b.Dimensions[j].Key })
	b.State = be.State
	b.ExternalIP = be.ExternalIP
	b.AuthenticatedAs = be.AuthenticatedAs
	b.Version = be.Version
	b.Quarantined = be.QuarantinedMsg != ""
	b.MaintenanceMsg = be.MaintenanceMsg
	b.TaskID = be.TaskID
}

type botTasks struct {
	Cursor string       `json:"cursor"`
	Items  []TaskResult `json:"items"`
	Now    cloudTime    `json:"now"`
}

type cloudTime string

func toCloudTime(t time.Time) cloudTime {
	return cloudTime(t.UTC().Format("2006-01-02T15:04:05"))
}
