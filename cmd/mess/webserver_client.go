package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/maruel/mess/internal"
	"github.com/maruel/mess/internal/model"
	"github.com/maruel/mess/messapi"
	"github.com/rs/zerolog/log"
)

func (s *server) apiEndpoint(w http.ResponseWriter, r *http.Request) {
	// Always parse URL query parameters.
	newValues, err := url.ParseQuery(r.URL.RawQuery)
	if newValues == nil {
		newValues = url.Values{}
	}
	r.Form = newValues
	if err != nil {
		sendJSONResponse(w, errorStatus{status: 400, err: err})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/server/") {
		s.apiEndpointServer(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/bots/") {
		s.apiEndpointBots(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") {
		s.apiEndpointTasks(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/bot/") {
		s.apiEndpointBot(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/task/") {
		s.apiEndpointTask(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/queues/") {
		s.apiEndpointQueues(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/config/") {
		log.Ctx(r.Context()).Error().Msg("TODO")
	}
	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiEndpointServer(w http.ResponseWriter, r *http.Request) {
	// All server APIs are GET.
	if !isMethodJSON(w, r, "GET") {
		return
	}
	if r.URL.Path == "/server/details" {
		sendJSONResponse(w, messapi.ServerDetailsResponse{
			ServerVersion: serverVersion,
			BotVersion:    internal.GetBotVersion(r),
		})
		return
	}
	if r.URL.Path == "/server/permissions" {
		_ = messapi.ServerPermissionsRequest{
			BotID:  r.FormValue("bot_id"),
			TaskID: model.TaskID(r.FormValue("task_id")),
			Tags:   r.Form["tags"],
		}
		sendJSONResponse(w, messapi.ServerPermissionsResponse{
			DeleteBot: true,
		})
		return
	}
	if r.URL.Path == "/server/token" {
		log.Ctx(r.Context()).Error().Msg("TODO")
	}

	// Intentionally not implementing get_bootstrap and get_bot_config.
	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiEndpointBots(w http.ResponseWriter, r *http.Request) {
	// All bots APIs are GET.
	if !isMethodJSON(w, r, "GET") {
		return
	}
	cloudNow := messapi.CloudTime(time.Now())
	if r.URL.Path == "/bots/count" {
		_ = messapi.BotsCountRequest{
			Dimensions: r.Form["dimensions"],
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		count := s.tables.BotCount()
		sendJSONResponse(w, messapi.BotsCountResponse{
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
		_ = messapi.BotsDimensionsRequest{
			Pool: r.Form["pool"],
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		sendJSONResponse(w, messapi.BotsDimensionsResponse{
			BotsDimensions: []messapi.StringListPair{},
			Now:            cloudNow,
		})
		return
	}
	if r.URL.Path == "/bots/list" {
		req := messapi.BotsListRequest{
			Limit:         messapi.ToInt64(r.FormValue("limit"), 200),
			Cursor:        r.FormValue("cursor"),
			Dimensions:    r.Form["dimensions"],
			Quarantined:   messapi.ToThreeState(r.FormValue("quarantined")),
			InMaintenance: messapi.ToThreeState(r.FormValue("in_maintenance")),
			IsDead:        messapi.ToThreeState(r.FormValue("is_dead")),
			IsBusy:        messapi.ToThreeState(r.FormValue("is_busy")),
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		bots, cursor := s.tables.BotGetSlice(req.Cursor, req.Limit)
		items := make([]messapi.Bot, len(bots))
		for i := range bots {
			items[i].FromDB(&bots[i])
		}
		sendJSONResponse(w, messapi.BotsListResponse{
			Cursor:       cursor,
			Items:        items,
			Now:          cloudNow,
			DeathTimeout: 30,
		})
		return
	}

	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiEndpointTasks(w http.ResponseWriter, r *http.Request) {
	cloudNow := messapi.CloudTime(time.Now())
	if r.URL.Path == "/tasks/cancel" {
		t := messapi.TasksCancelRequest{}
		if !readPOSTJSON(w, r, &t) {
			return
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		sendJSONResponse(w, messapi.TasksCancelResponse{
			Cursor:  "",
			Now:     cloudNow,
			Matched: 0,
		})
		return
	}
	if r.URL.Path == "/tasks/count" {
		if !isMethodJSON(w, r, "GET") {
			return
		}
		_ = messapi.TasksCountRequest{
			End:   messapi.ToTime(r.FormValue("end")),
			Start: messapi.ToTime(r.FormValue("start")),
			State: r.FormValue("state"),
			Tags:  r.Form["tags"],
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		count := s.tables.TaskRequestCount()
		sendJSONResponse(w, messapi.TasksCountResponse{
			Count: int32(count),
			Now:   cloudNow,
		})
		return
	}
	if r.URL.Path == "/tasks/get_states" {
		if !isMethodJSON(w, r, "GET") {
			return
		}
		_ = messapi.TasksGetStateRequest{
			TaskID: r.Form["task_id"],
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		sendJSONResponse(w, messapi.TasksGetStateResponse{})
		return
	}
	if r.URL.Path == "/tasks/list" {
		if !isMethodJSON(w, r, "GET") {
			return
		}
		_ = messapi.TasksListRequest{
			Limit:                   messapi.ToInt64(r.FormValue("limit"), 200),
			Cursor:                  r.FormValue("cursor"),
			End:                     messapi.ToTime(r.FormValue("end")),
			Start:                   messapi.ToTime(r.FormValue("start")),
			State:                   r.FormValue("state"),
			Tags:                    r.Form["tags"],
			Sort:                    r.FormValue("sort"),
			IncludePerformanceStats: r.FormValue("include_performance_stats") == "",
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		sendJSONResponse(w, messapi.TasksListResponse{
			Cursor: "",
			Items:  []messapi.TaskResult{},
			Now:    cloudNow,
		})
		return
	}
	if r.URL.Path == "/tasks/new" {
		t := messapi.TasksNewRequest{}
		if !readPOSTJSON(w, r, &t) {
			return
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		sendJSONResponse(w, messapi.TasksNewResponse{})
		return
	}
	if r.URL.Path == "/tasks/requests" {
		if !isMethodJSON(w, r, "GET") {
			return
		}
		_ = messapi.TasksRequestsRequest{
			Limit:                   messapi.ToInt64(r.FormValue("limit"), 200),
			Cursor:                  r.FormValue("cursor"),
			End:                     messapi.ToTime(r.FormValue("end")),
			Start:                   messapi.ToTime(r.FormValue("start")),
			State:                   r.FormValue("state"),
			Tags:                    r.Form["tags"],
			Sort:                    r.FormValue("sort"),
			IncludePerformanceStats: r.FormValue("include_performance_stats") == "",
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		sendJSONResponse(w, messapi.TasksRequestsResponse{
			Cursor: "",
			Items:  []messapi.TaskRequest{},
			Now:    cloudNow,
		})
		return
	}

	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiEndpointBot(w http.ResponseWriter, r *http.Request) {
	cloudNow := messapi.CloudTime(time.Now())
	if n := strings.SplitN(r.URL.Path[len("/bot/"):], "/", 2); len(n) == 2 {
		id := n[0]
		switch n[1] {
		case "delete":
			// It's a POST but with nothing in it.
			if !readPOSTJSON(w, r, &struct{}{}) {
				return
			}
			// TODO(maruel): bot.Deleted = true
			log.Ctx(r.Context()).Error().Msg("TODO")
			sendJSONResponse(w, messapi.BotDeleteResponse{
				Deleted: true,
			})
			return
		case "events":
			if !isMethodJSON(w, r, "GET") {
				return
			}
			req := messapi.BotEventsRequest{
				Limit:  messapi.ToInt64(r.FormValue("limit"), 200),
				Cursor: r.FormValue("cursor"),
				End:    messapi.ToTime(r.FormValue("end")),
				Start:  messapi.ToTime(r.FormValue("start")),
			}
			all, cursor := s.tables.BotEventGetSlice(id, req.Cursor, req.Limit, req.Start, req.End)
			events := make([]messapi.BotEvent, len(all))
			for i := range all {
				events[i].FromDB(&all[i])
			}
			sendJSONResponse(w, messapi.BotEventsResponse{
				Cursor: cursor,
				Items:  events,
				Now:    cloudNow,
			})
			return
		case "get":
			if !isMethodJSON(w, r, "GET") {
				return
			}
			bi := messapi.BotGetResponse{}
			bot := model.Bot{}
			s.tables.BotGet(id, &bot)
			bi.FromDB(&bot)
			sendJSONResponse(w, bi)
			return
		case "tasks":
			if !isMethodJSON(w, r, "GET") {
				return
			}
			_ = messapi.BotTasksRequest{
				Limit:                   messapi.ToInt64(r.FormValue("limit"), 200),
				Cursor:                  r.FormValue("cursor"),
				End:                     messapi.ToTime(r.FormValue("end")),
				Start:                   messapi.ToTime(r.FormValue("start")),
				State:                   r.FormValue("state"),
				Sort:                    r.FormValue("sort"),
				IncludePerformanceStats: r.FormValue("include_performance_stats") == "",
			}
			log.Ctx(r.Context()).Error().Msg("TODO")
			sendJSONResponse(w, messapi.BotTasksResponse{
				Cursor: "",
				Items:  []messapi.TaskResult{},
				Now:    cloudNow,
			})
			return
		case "terminate":
			// It's a POST but with nothing in it.
			if !readPOSTJSON(w, r, &struct{}{}) {
				return
			}
			log.Ctx(r.Context()).Error().Msg("TODO")
			sendJSONResponse(w, messapi.BotTerminateResponse{
				TaskID: "",
			})
			return
		}
	}

	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiEndpointTask(w http.ResponseWriter, r *http.Request) {
	if n := strings.SplitN(r.URL.Path[len("/task/"):], "/", 2); len(n) == 2 {
		//id := n[0]
		switch n[1] {
		case "cancel":
			// It's a POST but with nothing in it.
			if !readPOSTJSON(w, r, &struct{}{}) {
				return
			}
			log.Ctx(r.Context()).Error().Msg("TODO")
			sendJSONResponse(w, messapi.TaskCancelResponse{
				Ok:         false,
				WasRunning: false,
			})
			return
		case "request":
			if !isMethodJSON(w, r, "GET") {
				return
			}
			log.Ctx(r.Context()).Error().Msg("TODO")
			sendJSONResponse(w, messapi.TaskRequestResponse{})
			return
		case "result":
			if !isMethodJSON(w, r, "GET") {
				return
			}
			_ = messapi.TaskResultRequest{
				IncludePerformanceStats: r.FormValue("include_performance_stats") == "",
			}
			log.Ctx(r.Context()).Error().Msg("TODO")
			sendJSONResponse(w, messapi.TaskResultResponse{})
			return
		case "stdout":
			if !isMethodJSON(w, r, "GET") {
				return
			}
			_ = messapi.TaskStdoutRequest{
				Offset: messapi.ToInt64(r.FormValue("offset"), 0),
				Length: messapi.ToInt64(r.FormValue("length"), 16*1000*1024),
			}
			log.Ctx(r.Context()).Error().Msg("TODO")
			sendJSONResponse(w, messapi.TaskStdoutResponse{
				Output: "",
				State:  messapi.Pending,
			})
			return
		}
	}

	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiEndpointQueues(w http.ResponseWriter, r *http.Request) {
	// All queues APIs are GET.
	if !isMethodJSON(w, r, "GET") {
		return
	}
	if r.URL.Path == "/queues/list" {
		_ = messapi.TaskQueuesListRequest{
			Limit:  messapi.ToInt64(r.FormValue("limit"), 200),
			Cursor: r.FormValue("cursor"),
		}
		log.Ctx(r.Context()).Error().Msg("TODO")
		sendJSONResponse(w, messapi.TaskQueuesListResponse{})
		return
	}

	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func isMethodJSON(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		sendJSONResponse(w, errorStatus{status: 405, err: errWrongMethod})
		return false
	}
	return true
}

func readPOSTJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if r.Method != "POST" {
		sendJSONResponse(w, errorStatus{status: 405, err: errWrongMethod})
		return false
	}
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		r.Body.Close()
		sendJSONResponse(w, errorStatus{status: 400, err: err})
		return false
	}
	r.Body.Close()
	return true
}
