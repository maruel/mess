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
	ctx := r.Context()

	// Non-API URLs.
	h := w.Header()
	if r.URL.Path == "/server_ping" {
		h.Set("Content-Type", "text/plain")
		w.Write([]byte("Server Up"))
		return
	}
	if r.URL.Path == "/bot_code" {
		version := internal.GetBotVersion(ctx, getURL(r))
		http.Redirect(w, r, "/swarming/api/v1/bot/bot_code/"+version, http.StatusFound)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/bot_code") {
		version := internal.GetBotVersion(ctx, getURL(r))
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
		http.ServeContent(w, r, "swarming_bot.zip", started, bytes.NewReader(internal.GetBotZIP(ctx, getURL(r))))
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

	bcr := botCommonRequest{}
	// Ignore extra keys. Will be processed below.
	if err = json.Unmarshal(raw, &bcr); err != nil {
		log.Ctx(ctx).Error().Str("err", err.Error()).Msg("failed to decode bot request")
		sendJSONResponse(w, errorStatus{status: 400, err: err})
		return
	}

	if id == "" && r.URL.Path == "/handshake" {
		// handshake doesn't set the header yet. We should fix!
		if len(bcr.Dimensions) != 0 {
			if d := bcr.Dimensions["id"]; len(d) == 1 {
				id = d[0]
			}
		}
	}
	if id == "" {
		sendJSONResponse(w, errorStatus{status: 400, err: errors.New("missing bot id HTTP header")})
		return
	}

	bot := model.Bot{Key: id, Created: now}
	s.tables.BotGet(id, &bot)
	bot.LastSeen = now
	if bcr.Version != "" {
		bot.Version = bcr.Version
	}
	if len(bcr.Dimensions) != 0 {
		bot.Dimensions = bcr.Dimensions
	}
	bot.ExternalIP = getRemoteIP(r)
	if len(bcr.State) != 0 {
		if s, err := json.Marshal(bcr.State); err == nil {
			bot.State = s
		} else {
			bot.State, _ = json.Marshal(map[string]string{"quarantined": "invalid state: " + err.Error()})
		}
	}
	s.tables.BotSet(&bot)

	// API URLs.
	if r.URL.Path == "/handshake" {
		e := model.BotEvent{}
		e.InitFrom(&bot, now, "handshake", "")
		s.tables.BotEventAdd(&e)
		bhr := botHandshakeRequest{}
		if err := decodeJSONStrict(raw, &bhr); err != nil {
			panic(err)
			sendJSONResponse(w, errorStatus{status: 400, err: err})
			return
		}
		data := botHandshakeResponse{
			BotVersion:         internal.GetBotVersion(ctx, getURL(r)),
			BotConfigRev:       "??",
			BotConfigName:      "bot_config.py",
			ServerVersion:      s.version,
			BotGroupCfgVersion: "??",
			BotGroupCfg: botGroupCfg{
				Dimensions: []messapi.StringListPair{},
			},
		}
		// TODO(maruel): Inject server-side bot config and dimensions.
		sendJSONResponse(w, data)
		return
	}

	if r.URL.Path == "/poll" {
		s.apiBotPoll(w, r, now, id, &bot, raw)
		return
	}
	if r.URL.Path == "/event" {
		ber := botEventRequest{}
		if err := decodeJSONStrict(raw, &ber); err != nil {
			panic(err)
			sendJSONResponse(w, errorStatus{status: 400, err: err})
			return
		}
		e := model.BotEvent{}
		e.InitFrom(&bot, now, ber.Event, ber.Message)
		s.tables.BotEventAdd(&e)
		sendJSONResponse(w, map[string]string{})
		return
	}
	if r.URL.Path == "/oauth_token" {
		bor := botOAuthTokenRequest{}
		if err := decodeJSONStrict(raw, &bor); err != nil {
			panic(err)
			sendJSONResponse(w, errorStatus{status: 400, err: err})
			return
		}
		sendJSONResponse(w, botOAuthTokenResponse{})
		return
	}
	if r.URL.Path == "/id_token" {
		bir := botIDTokenRequest{}
		if err := decodeJSONStrict(raw, &bir); err != nil {
			panic(err)
			sendJSONResponse(w, errorStatus{status: 400, err: err})
			return
		}
		sendJSONResponse(w, botIDTokenResponse{})
		return
	}
	if r.URL.Path == "/task_update" || strings.HasPrefix(r.URL.Path, "/task_update/") {
		// The bot has an inconsistency where it may use two kinds of URLs. task_id
		// is always passed as a POST argument to use this.
		btr := botTaskUpdateRequest{}
		if err := decodeJSONStrict(raw, &btr); err != nil {
			panic(err)
			sendJSONResponse(w, errorStatus{status: 400, err: err})
			return
		}
		/* Only when state changes.
		e := model.BotEvent{}
		e.InitFrom(&bot, now, "task_update", btr.Message)
		s.tables.BotEventAdd(&e)
		*/
		sendJSONResponse(w, botTaskUpdateResponse{Ok: true})
		return
	}
	if r.URL.Path == "/task_error" || strings.HasPrefix(r.URL.Path, "/task_error/") {
		btr := botTaskErrorRequest{}
		if err := decodeJSONStrict(raw, &btr); err != nil {
			panic(err)
			sendJSONResponse(w, errorStatus{status: 400, err: err})
			return
		}
		e := model.BotEvent{}
		e.InitFrom(&bot, now, "task_error", btr.Message)
		s.tables.BotEventAdd(&e)
		sendJSONResponse(w, map[string]string{})
		return
	}
	log.Ctx(ctx).Error().Msg("Unknown bot request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiBotPoll(w http.ResponseWriter, r *http.Request, now time.Time, id string, bot *model.Bot, raw []byte) {
	bpr := botPollRequest{}
	if err := decodeJSONStrict(raw, &bpr); err != nil {
		panic(err)
		sendJSONResponse(w, errorStatus{status: 400, err: err})
		return
	}
	// In practice it would be the command sent.
	// bot.AddEvent(now, "poll", "")
	bp := botPollResponse{}
	ctx := r.Context()
	if version := internal.GetBotVersion(ctx, getURL(r)); bot.Version != version {
		bp.Cmd = "update"
		bp.Version = version
		sendJSONResponse(w, bp)
		return
	}

	task := s.sched.poll(ctx, bot)
	if task != nil {
		bp.Cmd = "run"
		bp.Manifest.fromRequest(task, 0)
		bp.Manifest.BotID = bot.Key
		bp.Manifest.BotAuthenticatedAs = bot.AuthenticatedAs
		bp.Manifest.Host = getURL(r)
		sendJSONResponse(w, bp)
	}
	// TODO(maruel): bot_restart, terminate.
	bp.Cmd = "sleep"
	bp.Duration = 10
	sendJSONResponse(w, bp)
}

// botCommonRequest is the JSON HTTP POST content for most requests under
// /swarming/api/v1/bot/.
type botCommonRequest struct {
	//Token   string `json:"tok"`
	Dimensions map[string][]string    `json:"dimensions"`
	State      map[string]interface{} `json:"state"`
	Version    string                 `json:"version"`
}

// botHandshakeRequest is arguments for /swarming/api/v1/bot/handshake.
type botHandshakeRequest struct {
	botCommonRequest
}

// botHandshakeResponse is response to /swarming/api/v1/bot/handshake.
type botHandshakeResponse struct {
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

// botPollRequest is arguments for /swarming/api/v1/bot/poll.
type botPollRequest struct {
	botCommonRequest
	RequestUUID string `json:"request_uuid"`
}

// botPollResponse is response to /swarming/api/v1/bot/poll.
type botPollResponse struct {
	Cmd string `json:"cmd"`

	// Cmd == "bot_restart"
	Message string `json:"message,omitempty"`

	// Cmd == "run"
	Manifest botPollManifest `json:"manifest,omitempty"`

	// Cmd == "sleep"
	Duration    int  `json:"duration,omitempty"`
	Quarantined bool `json:"quarantined,omitempty"`

	// Cmd == "terminate"
	TaskID string `json:"task_id,omitempty"`

	// Cmd == "update"
	Version string `json:"version,omitempty"`
}

type botPollManifest struct {
	BotID              string                   `json:"bot_id,omitempty"`
	BotAuthenticatedAs string                   `json:"bot_authenticated_as,omitempty"`
	Caches             []botPollCache           `json:"caches,omitempty"`
	CIPDInput          botPollCIPDInput         `json:"cipd_input,omitempty"`
	Command            []string                 `json:"command,omitempty"`
	Containment        botPollContainment       `json:"containment,omitempty"`
	Dimensions         []messapi.StringPair     `json:"dimensions,omitempty"`
	Env                []messapi.StringPair     `json:"env,omitempty"`
	EnvPrefixes        []messapi.StringListPair `json:"env_prefixes,omitempty"`
	GracePeriod        int64                    `json:"grace_period,omitempty"`
	HardTimeout        int64                    `json:"hard_timeout,omitempty"`
	Host               string                   `json:"host,omitempty"`
	IOTimeout          int64                    `json:"io_timeout,omitempty"`
	SecretBytes        string                   `json:"secret_bytes,omitempty"` // base64 encoded
	CASInputRoot       botPollCASInputRoot      `json:"cas_input_root,omitempty"`
	Outputs            []string                 `json:"outputs,omitempty"`
	Realm              botPollRealm             `json:"realm,omitempty"`
	RelativeWD         string                   `json:"relative_cwd,omitempty"`
	ResultDB           botPollResultDB          `json:"resultdb,omitempty"`
	ServiceAccounts    botPollServiceAccounts   `json:"service_accounts,omitempty"`
	TaskID             model.TaskID             `json:"task_id,omitempty"`
}

func (b *botPollManifest) fromRequest(t *model.TaskRequest, slice int) {
	p := &t.TaskSlices[slice].Properties
	b.Caches = make([]botPollCache, len(p.Caches))
	for i := range p.Caches {
		b.Caches[i].Name = p.Caches[i].Name
		b.Caches[i].Path = p.Caches[i].Path
		b.Caches[i].Hint = 0 // TODO(maruel): Calculate.
	}
	b.CIPDInput.ClientPackage.fromDB(&p.CIPDClient)
	b.CIPDInput.Packages = make([]botCIPDPackage, len(p.CIPDPackages))
	for i := range p.CIPDPackages {
		b.CIPDInput.Packages[i].fromDB(&p.CIPDPackages[i])
	}
	b.Command = p.Command
	// TODO(maruel): b.Containment
	b.Dimensions = messapi.ToStringPairs(p.Dimensions)
	b.Env = messapi.ToStringPairs(p.Env)
	b.EnvPrefixes = messapi.ToStringListPairs(p.EnvPrefixes)
	b.GracePeriod = int64(p.GracePeriod / time.Second)
	b.HardTimeout = int64(p.HardTimeout / time.Second)
	b.IOTimeout = int64(p.IOTimeout / time.Second)
	// TODO(maruel): SecretBytes = // base64
	b.CASInputRoot.CASInstance = p.CASHost
	b.CASInputRoot.Digest = p.Input
	b.Outputs = p.Outputs
	b.Realm.Name = t.Realm
	b.RelativeWD = p.RelativeWD
	b.ResultDB.Host = "" // TODO(maruel): Add
	b.ResultDB.CurrentInvocation.Name = ""
	b.ResultDB.CurrentInvocation.UpdateToken = ""
	b.ServiceAccounts.System = "none" // TODO(maruel): impersonation
	b.ServiceAccounts.Task = "none"
	b.TaskID = model.ToTaskID(t.Key)
}

type botPollCache struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
	Hint int64  `json:"hint,omitempty"`
}

type botCIPDPackage struct {
	PkgName string `json:"package_name",omitempty`
	Version string `json:"version,omitempty"`
	Path    string `json:"path,omitempty"`
}

func (b *botCIPDPackage) fromDB(m *model.CIPDPackage) {
	b.PkgName = m.PkgName
	b.Version = m.Version
	b.Path = m.Path
}

type botPollCIPDInput struct {
	ClientPackage botCIPDPackage   `json:"client_package,omitempty"`
	Packages      []botCIPDPackage `json:"packages,omitempty"`
	Server        string           `json:"server,omitempty"`
}

type botPollContainment struct {
	ContainmentType string `json:"containment_type,omitempty"`
}

type botPollCASInputRoot struct {
	CASInstance string       `json:"cas_instance,omitempty"`
	Digest      model.Digest `json:"digest,omitempty"`
}

type botPollRealm struct {
	Name string `json:"name,omitempty"`
}

type botPollResultDB struct {
	Host              string                    `json:"hostname,omitempty"`
	CurrentInvocation botPollResultDBInvocation `json:"current_invocation,omitempty"`
}

type botPollResultDBInvocation struct {
	Name        string `json:"name,omitempty"`
	UpdateToken string `json:"update_token,omitempty"`
}

type botPollServiceAccounts struct {
	// The values are one of "none", "bot" or an email address. When an email
	// address is specified, it is assumed to be a Google Cloud IAM service
	// account. The bot uses /oauth_token API to grab a token.

	System string `json:"system,omitempty"`
	Task   string `json:"task,omitempty"`
}

// botEventRequest is arguments for /swarming/api/v1/bot/event.
type botEventRequest struct {
	botCommonRequest
	Event   string `json:"event"`
	Message string `json:"message"`
}

// botOAuthTokenRequest is arguments for /swarming/api/v1/bot/oauth_token.
type botOAuthTokenRequest struct {
	botCommonRequest
	AccountID string `json:"account_id"`
	Audience  string `json:"audience"`
	TaskID    string `json:"task_id"`
}

// botOAuthTokenResponse is response to /swarming/api/v1/bot/oauth_token.
type botOAuthTokenResponse struct {
	ServiceAccount string `json:"service_account"`
	IDToken        string `json:"id_token"`
	ExpiryEpoch    int64  `json:"expiry"`
}

// botIDTokenRequest is arguments for /swarming/api/v1/bot/id_token.
type botIDTokenRequest struct {
	botCommonRequest
	AccountID string `json:"account_id"`
	Audience  string `json:"audience"`
	TaskID    string `json:"task_id"`
}

// botIDTokenResponse is response to /swarming/api/v1/bot/id_token.
type botIDTokenResponse struct {
	ServiceAccount string `json:"service_account"`
	IDToken        string `json:"id_token"`
	ExpiryEpoch    int64  `json:"expiry"`
}

// botTaskUpdateRequest is arguments for /swarming/api/v1/bot/task_update.
type botTaskUpdateRequest struct {
	botCommonRequest
	BotOverheadSecs  float64     `json:"bot_overhead"`
	CacheTrimStats   interface{} `json:"cache_trim_stats"`
	CASOutputRoot    interface{} `json:"cas_output_root"`
	CIPDPins         interface{} `json:"cipd_pins"`
	CIPDStats        interface{} `json:"cipd_stats"`
	CleanupStats     interface{} `json:"cleanup_stats"`
	CostUSD          float64     `json:"cost_usd"`
	DurationSecs     float64     `json:"duration"`
	ExitCode         int32       `json:"exit_code"`
	HardTimeout      bool        `json:"hard_timeout"`
	ID               string      `json:"id"`
	IOTimeout        bool        `json:"io_timeout"`
	IsolatedStats    interface{} `json:"isolated_stats"`
	NamedCachesStats interface{} `json:"named_caches_stats"`
	Output           []byte      `json:"output"`
	OutputChunkStart int64       `json:"output_chunk_start"`
	TaskID           string      `json:"task_id"`
}

// botTaskUpdateResponse is arguments for /swarming/api/v1/bot/task_update.
type botTaskUpdateResponse struct {
	MustStop bool `json:"must_stop"`
	Ok       bool `json:"ok"`
}

// botTaskErrorRequest is arguments for /swarming/api/v1/bot/task_error.
type botTaskErrorRequest struct {
	botCommonRequest
	TaskID  string `json:"task_id"`
	Message string `json:"message"`
}
