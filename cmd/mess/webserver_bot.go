package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/maruel/mess/internal"
	"github.com/maruel/mess/internal/model"
	"github.com/maruel/mess/messapi"
	"github.com/rs/zerolog/log"
)

func (s *server) apiBot(w http.ResponseWriter, r *http.Request) {
	if !isLocal(r) {
		sendJSONResponse(w, errorStatus{status: 403})
		return
	}

	// Non-API URLs.
	h := w.Header()
	if r.URL.Path == "/server_ping" {
		h.Set("Content-Type", "text/plain")
		w.Write([]byte("Server Up"))
		return
	}
	if r.URL.Path == "/bot_code" {
		version := internal.GetBotVersion(r)
		http.Redirect(w, r, "/swarming/api/v1/bot/bot_code/"+version, http.StatusFound)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/bot_code") {
		version := internal.GetBotVersion(r)
		if r.URL.Path[len("/bot_code/"):] != version {
			// It happens...
			http.Redirect(w, r, "/swarming/api/v1/bot/bot_code/"+version, http.StatusFound)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		h.Set("Cache-Control", "public, max-age=3600")
		h.Set("Content-Type", "application/octet-stream")
		h.Set("Content-Disposition", "attachment; filename=\"swarming_bot.zip\"")
		http.ServeContent(w, r, "swarming_bot.zip", started, bytes.NewReader(internal.GetBotZIP(r)))
		return
	}

	// All other endpoints are bot APIs expecting a JSON response.
	id := r.Header.Get("X-Luci-Swarming-Bot-ID")
	now := time.Now()
	// canary, _ := r.Cookie("GOOGAPPUID")

	if r.Method != "POST" {
		sendJSONResponse(w, errorStatus{status: http.StatusMethodNotAllowed})
		return
	}

	// Read it all first to ensure there's not a connection error. The data
	// should fit memory.
	// TODO(maruel): limit data size.
	raw, err := ioutil.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		sendJSONResponse(w, errorStatus{status: 400, err: err})
		return
	}

	br := botRequest{}
	j := json.NewDecoder(bytes.NewReader(raw))
	j.DisallowUnknownFields()
	j.UseNumber()
	if err = j.Decode(&br); err != nil {
		log.Ctx(r.Context()).Error().Str("err", err.Error()).Msg("failed to decode bot request")
		sendJSONResponse(w, errorStatus{status: 400, err: err})
		return
	}

	if id == "" && r.URL.Path == "/handshake" {
		// handshake doesn't set the header yet. We should fix.
		if d := br.Dimensions["id"]; len(d) == 1 {
			id = d[0]
		}
	}
	if id == "" {
		sendJSONResponse(w, errorStatus{status: 400, err: errors.New("missing bot id HTTP header")})
		return
	}

	bot := model.Bot{Key: id, Created: now}
	s.tables.BotGet(id, &bot)
	bot.LastSeen = now
	bot.Version = br.Version
	bot.Dimensions = br.Dimensions
	bot.ExternalIP = getRemoteIP(r)
	if s, err := json.Marshal(br.State); err == nil {
		bot.State = s
	} else {
		bot.State, _ = json.Marshal(map[string]string{"quarantined": "invalid state: " + err.Error()})
	}
	s.tables.BotSet(&bot)

	// API URLs.
	if r.URL.Path == "/handshake" {
		e := model.BotEvent{}
		e.InitFrom(&bot, now, "handshake", "")
		s.tables.BotEventAdd(&e)

		data := botHandshake{
			BotVersion:         internal.GetBotVersion(r),
			BotConfigRev:       "??",
			BotConfigName:      "bot_config.py",
			ServerVersion:      s.version,
			BotGroupCfgVersion: "??",
			BotGroupCfg: botGroupCfg{
				// Inject server side dimensions.
				Dimensions: []messapi.StringListPair{},
			},
		}
		// Inject data.BotConfig, data.BotConfigRev, data.BotConfigName
		sendJSONResponse(w, data)
		s.tables.BotSet(&bot)
		return
	}

	if r.URL.Path == "/poll" {
		s.apiBotPoll(w, r, now, id, &bot)
		s.tables.BotSet(&bot)
		return
	}
	if r.URL.Path == "/event" {
		e := model.BotEvent{}
		e.InitFrom(&bot, now, br.Event, br.Message)
		s.tables.BotEventAdd(&e)
		sendJSONResponse(w, map[string]string{})
		s.tables.BotSet(&bot)
		return
	}
	if r.URL.Path == "/oauth_token" {
		// "account_id"
		// "id"
		// "scopes"
		// "task_id"
		sendJSONResponse(w, map[string]string{})
		s.tables.BotSet(&bot)
		return
	}
	if r.URL.Path == "/id_token" {
		// "account_id"
		// "id"
		// "audience"
		// "task_id"
		sendJSONResponse(w, map[string]string{})
		s.tables.BotSet(&bot)
		return
	}
	if r.URL.Path == "/task_update" {
		e := model.BotEvent{}
		e.InitFrom(&bot, now, "task_update", br.Message)
		s.tables.BotEventAdd(&e)
		sendJSONResponse(w, map[string]string{})
		s.tables.BotSet(&bot)
		return
	}
	if r.URL.Path == "/task_error" {
		e := model.BotEvent{}
		e.InitFrom(&bot, now, "task_error", br.Message)
		s.tables.BotEventAdd(&e)
		sendJSONResponse(w, map[string]string{})
		s.tables.BotSet(&bot)
		return
	}
	log.Ctx(r.Context()).Error().Msg("Unknown bot request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiBotPoll(w http.ResponseWriter, r *http.Request, now time.Time, id string, bot *model.Bot) {
	// In practice it would be the command sent.
	// bot.AddEvent(now, "poll", "")
	bp := botPoll{}
	if version := internal.GetBotVersion(r); bot.Version != version {
		bp.Cmd = "update"
		bp.Version = version
		sendJSONResponse(w, bp)
		return
	}

	task := s.sched.poll(r.Context(), bot)
	if task != nil {
		bp.Cmd = "run"
		bp.Manifest.fromRequest(task)
		bp.Manifest.BotID = bot.Key
		bp.Manifest.BotAuthenticatedAs = bot.AuthenticatedAs
		sendJSONResponse(w, bp)
	}
	// TODO(maruel): bot_restart, terminate.

	// TODO(maruel): When sleep, do long (2 minutes?) hanging poll instead.
	bp.Cmd = "sleep"
	bp.Duration = 10
	sendJSONResponse(w, bp)
}

// botRequest is the JSON HTTP POST content. Depending on different endpoints,
// different values are used. This should be cleaned up.
type botRequest struct {
	Token string `json:"tok"`
	//BotID       string                 `json:"bot_id"`
	Dimensions  map[string][]string    `json:"dimensions"`
	RequestUUID string                 `json:"request_uuid"`
	State       map[string]interface{} `json:"state"`
	Version     string                 `json:"version"`
	Event       string                 `json:"event"`
	Message     string                 `json:"message"`
}

type botHandshake struct {
	BotVersion         string      `json:"bot_version"`
	BotConfigRev       string      `json:"bot_config_rev"`
	BotConfigName      string      `json:"bot_config_name"`
	ServerVersion      string      `json:"server_version"`
	BotGroupCfgVersion string      `json:"bot_group_cfg_version"`
	BotGroupCfg        botGroupCfg `json:"bot_group_cfg"`
}

type botGroupCfg struct {
	Dimensions []messapi.StringListPair `json:"dimensions"`
}

type botPoll struct {
	Cmd string `json:"cmd"`

	// Cmd == "bot_restart"
	Message string `json:"message"`

	// Cmd == "run"
	Manifest botPollManifest `json:"manifest"`

	// Cmd == "sleep"
	Duration    int  `json:"duration"`
	Quarantined bool `json:"quarantined"`

	// Cmd == "terminate"
	TaskID string `json:"task_id"`

	// Cmd == "update"
	Version string `json:"version"`
}

type botPollManifest struct {
	BotID              string                 `json:"bot_id"`
	BotAuthenticatedAs string                 `json:"bot_authenticated_as"`
	Caches             []botPollCache         `json:"caches"`
	CIPDInput          []botPollCIPDInput     `json:"cipd_input"`
	Command            []string               `json:"command"`
	Containment        botPollContainment     `json:"containment"`
	Dimensions         []messapi.StringPair   `json:"dimensions"`
	Env                []messapi.StringPair   `json:"env"`
	EnvPrefixes        []messapi.StringPair   `json:"env_prefixes"`
	GracePeriod        int                    `json:"grace_period"`
	HardTimeout        int                    `json:"hard_timeout"`
	Host               string                 `json:"host"`
	IOTimeout          int                    `json:"io_timeout"`
	SecretBytes        string                 `json:"secret_bytes"` // base64 encoded
	CASInputRoot       botPollCASInputRoot    `json:"cas_input_root"`
	Outputs            []string               `json:"outputs"`
	Realm              botPollRealm           `json:"realm"`
	RelativeWD         string                 `json:"relative_cwd"`
	ResultDB           botPollResultDB        `json:"resultdb"`
	ServiceAccounts    botPollServiceAccounts `json:"service_accounts"`
	TaskID             model.TaskID           `json:"task_id"`
}

func (b *botPollManifest) fromRequest(t *model.TaskRequest) {
	slice := t.TaskSlices[0]
	//b.Caches = slice.Properties.Caches
	//b.CIPDInput = slice.Properties.CIPDInput
	b.Command = slice.Properties.Command
	//b.Containment = slice.Properties.Containment
	// ...
	b.TaskID = model.ToTaskID(t.Key)
}

type botPollCache struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Hint int64  `json:"hint"`
}

type botPollCIPDInput struct {
	ClientPackage map[string]string   `json:"client_package"`
	Packages      []map[string]string `json:"packages"`
	Server        string              `json:"server"`
}

type botPollContainment struct {
	ContainmentType string `json:"containment_type"`
}

type botPollCASInputRoot struct {
	CASInstance string       `json:"cas_instance"`
	Digest      model.Digest `json:"digest"`
}

type botPollRealm struct {
	Name string `json:"name"`
}

type botPollResultDB struct {
	Host              string                    `json:"hostname"`
	CurrentInvocation botPollResultDBInvocation `json:"current_invocation"`
}

type botPollResultDBInvocation struct {
	Name        string `json:"name"`
	UpdateToken string `json:"update_token"`
}

type botPollServiceAccounts struct {
	// The values are one of "none", "bot" or an email address. When an email
	// address is specified, it is assumed to be a Google Cloud IAM service
	// account. The bot uses /oauth_token API to grab a token.

	System string `json:"system"`
	Task   string `json:"task"`
}
