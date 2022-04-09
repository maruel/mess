package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
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
		log.Ctx(r.Context()).Error().Msg("TODO: implement luci config")
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
		// TODO(maruel): There is not ACL yet.
		sendJSONResponse(w, messapi.ServerPermissionsResponse{
			DeleteBot:         true,
			DeleteBots:        true,
			TerminateBot:      true,
			CancelTask:        true,
			GetBootstrapToken: true,
			CancelTasks:       true,
			ListBots:          []string{}, // Depends on realm.
			ListTasks:         []string{},
		})
		return
	}
	if r.URL.Path == "/server/token" {
		log.Ctx(r.Context()).Error().Msg("TODO: implement server token")
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
		req := messapi.BotsCountRequest{
			Dimensions: r.Form["dimensions"],
		}
		dims := make(map[string]string, len(req.Dimensions))
		for _, s := range req.Dimensions {
			p := strings.SplitN(s, ":", 2)
			if len(p) != 2 {
				sendJSONResponse(w, errorStatus{status: 404, err: errors.New("bad dimensions format")})
			}
			dims[p[0]] = p[1]
		}
		total, quarantined, maintenance, dead, busy := s.tables.BotCount(dims)
		sendJSONResponse(w, messapi.BotsCountResponse{
			Now:         cloudNow,
			Count:       total,
			Quarantined: quarantined,
			Maintenance: maintenance,
			Dead:        dead,
			Busy:        busy,
		})
		return
	}
	if r.URL.Path == "/bots/dimensions" {
		req := messapi.BotsDimensionsRequest{
			Pool: r.Form["pool"],
		}
		if len(req.Pool) == 1 && req.Pool[0] == "" {
			req.Pool = nil
		}
		if len(req.Pool) != 0 {
			log.Ctx(r.Context()).Error().Interface("pool", req.Pool).Msg("TODO: implement bot pool")
		}
		// TODO(maruel): It has to be made more performant; O(n^3).
		objs, _ := s.tables.BotGetSlice("", 1000)
		dims := map[string][]string{}
		for i := range objs {
			for k, botvals := range objs[i].Dimensions {
				for _, botval := range botvals {
					found := false
					for _, v := range dims[k] {
						if botval == v {
							found = true
							break
						}
					}
					if !found {
						dims[k] = append(dims[k], botval)
					}
				}
			}
		}
		items := make([]messapi.StringListPair, 0, len(dims))
		for k, vals := range dims {
			sort.Strings(vals)
			items = append(items, messapi.StringListPair{Key: k, Values: vals})
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
		sendJSONResponse(w, messapi.BotsDimensionsResponse{
			BotsDimensions: items,
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
		log.Ctx(r.Context()).Error().Msg("TODO: implement filters")
		objs, cursor := s.tables.BotGetSlice(req.Cursor, int(req.Limit))
		items := make([]messapi.Bot, len(objs))
		for i := range objs {
			items[i].FromDB(&objs[i])
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
		if t.Limit == 0 {
			t.Limit = 100
		}
		log.Ctx(r.Context()).Error().Msg("TODO: implement mass cancel")
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
		log.Ctx(r.Context()).Error().Msg("TODO: filters end, start, state, tags")
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
		req := messapi.TasksGetStateRequest{
			TaskID: r.Form["task_id"],
		}
		out := make([]messapi.TaskState, len(req.TaskID))
		res := model.TaskResult{}
		for i, tid := range req.TaskID {
			// TODO(maruel): Be more efficient.
			id := model.FromTaskID(model.TaskID(tid))
			if id == 0 {
				out[i] = messapi.BotDied
			} else {
				s.tables.TaskResultGet(id, &res)
				out[i] = messapi.FromDBTaskState(res.State)
			}
		}
		sendJSONResponse(w, messapi.TasksGetStateResponse{
			States: out,
		})
		return
	}
	if r.URL.Path == "/tasks/list" {
		if !isMethodJSON(w, r, "GET") {
			return
		}
		req := messapi.TasksListRequest{
			Limit:                   messapi.ToInt64(r.FormValue("limit"), 200),
			Cursor:                  r.FormValue("cursor"),
			End:                     messapi.ToTime(r.FormValue("end")),
			Start:                   messapi.ToTime(r.FormValue("start")),
			State:                   r.FormValue("state"),
			Tags:                    r.Form["tags"],
			Sort:                    r.FormValue("sort"),
			IncludePerformanceStats: r.FormValue("include_performance_stats") == "",
		}
		log.Ctx(r.Context()).Error().Msg("TODO: implement State, Tags, Sort, Perf")
		f := model.Filter{
			Cursor:   req.Cursor,
			Limit:    int(req.Limit),
			Earliest: req.Start,
			Latest:   req.End,
		}
		objs, cursor := s.tables.TaskResultSlice("", f, model.TaskStateQueryAll, model.TaskSortCreated)
		items := make([]messapi.TaskResult, len(objs))
		for i := range objs {
			items[i].FromDB(&objs[i])
		}
		sendJSONResponse(w, messapi.TasksListResponse{
			Cursor: cursor,
			Items:  items,
			Now:    cloudNow,
		})
		return
	}
	if r.URL.Path == "/tasks/new" {
		t := messapi.TasksNewRequest{}
		if !readPOSTJSON(w, r, &t) {
			return
		}
		m := model.TaskRequest{}
		t.ToDB(&m)
		s.tables.TaskRequestAdd(&m)
		n := s.sched.enqueue(r.Context(), &m)
		s.tables.TaskResultSet(n)
		resp := messapi.TasksNewResponse{TaskID: model.ToTaskID(m.Key)}
		resp.Request.FromDB(&m)
		resp.Result.FromDB(n)
		sendJSONResponse(w, resp)
		return
	}
	if r.URL.Path == "/tasks/requests" {
		if !isMethodJSON(w, r, "GET") {
			return
		}
		req := messapi.TasksRequestsRequest{
			Limit:                   messapi.ToInt64(r.FormValue("limit"), 200),
			Cursor:                  r.FormValue("cursor"),
			End:                     messapi.ToTime(r.FormValue("end")),
			Start:                   messapi.ToTime(r.FormValue("start")),
			State:                   r.FormValue("state"),
			Tags:                    r.Form["tags"],
			Sort:                    r.FormValue("sort"),
			IncludePerformanceStats: r.FormValue("include_performance_stats") == "",
		}
		log.Ctx(r.Context()).Error().Msg("TODO: state, tags, sort, perf")
		f := model.Filter{
			Cursor:   req.Cursor,
			Limit:    int(req.Limit),
			Earliest: req.Start,
			Latest:   req.End,
		}
		objs, cursor := s.tables.TaskRequestSlice(f)
		items := make([]messapi.TaskRequest, len(objs))
		for i := range objs {
			items[i].FromDB(&objs[i])
		}
		sendJSONResponse(w, messapi.TasksRequestsResponse{
			Cursor: cursor,
			Items:  items,
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
			bot := model.Bot{}
			s.tables.BotGet(id, &bot)
			if !bot.Dead && time.Now().After(bot.LastSeen.Add(model.DeadAfter)) {
				bot.Dead = true
			}
			canDelete := !bot.Deleted && bot.Dead
			if canDelete {
				bot.Deleted = true
			}
			sendJSONResponse(w, messapi.BotDeleteResponse{
				Deleted: canDelete,
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
			f := model.Filter{
				Cursor:   req.Cursor,
				Limit:    int(req.Limit),
				Earliest: req.Start,
				Latest:   req.End,
			}
			objs, cursor := s.tables.BotEventGetSlice(id, f)
			items := make([]messapi.BotEvent, len(objs))
			for i := range objs {
				items[i].FromDB(&objs[i])
			}
			sendJSONResponse(w, messapi.BotEventsResponse{
				Cursor: cursor,
				Items:  items,
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
			req := messapi.BotTasksRequest{
				Limit:                   messapi.ToInt64(r.FormValue("limit"), 200),
				Cursor:                  r.FormValue("cursor"),
				End:                     messapi.ToTime(r.FormValue("end")),
				Start:                   messapi.ToTime(r.FormValue("start")),
				State:                   r.FormValue("state"),
				Sort:                    r.FormValue("sort"),
				IncludePerformanceStats: r.FormValue("include_performance_stats") == "",
			}
			log.Ctx(r.Context()).Error().Msg("TODO: State and Sort")
			f := model.Filter{
				Cursor:   req.Cursor,
				Limit:    int(req.Limit),
				Earliest: req.Start,
				Latest:   req.End,
			}
			s.tables.TaskResultSlice(id, f, model.TaskStateQueryAll, model.TaskSortCreated)
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
			log.Ctx(r.Context()).Error().Msg("TODO: implement terminate bot")
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
		id := model.FromTaskID(model.TaskID(n[0]))
		if id == 0 {
			sendJSONResponse(w, errorStatus{status: 400, err: errors.New("bad taskid")})
			return
		}
		switch n[1] {
		case "cancel":
			// It's a POST but with nothing in it.
			if !readPOSTJSON(w, r, &struct{}{}) {
				return
			}
			t := model.TaskResult{}
			s.tables.TaskResultGet(id, &t)
			log.Ctx(r.Context()).Error().Msg("TODO: implement cancel")
			sendJSONResponse(w, messapi.TaskCancelResponse{
				Ok:         false,
				WasRunning: t.State == model.Running,
			})
			return
		case "request":
			if !isMethodJSON(w, r, "GET") {
				return
			}
			t := model.TaskRequest{}
			s.tables.TaskRequestGet(id, &t)
			resp := messapi.TaskRequestResponse{}
			resp.FromDB(&t)
			sendJSONResponse(w, resp)
			return
		case "result":
			if !isMethodJSON(w, r, "GET") {
				return
			}
			_ = messapi.TaskResultRequest{
				IncludePerformanceStats: r.FormValue("include_performance_stats") == "",
			}
			t := model.TaskResult{}
			s.tables.TaskResultGet(id, &t)
			resp := messapi.TaskResultResponse{}
			resp.FromDB(&t)
			sendJSONResponse(w, resp)
			return
		case "stdout":
			if !isMethodJSON(w, r, "GET") {
				return
			}
			_ = messapi.TaskStdoutRequest{
				Offset: messapi.ToInt64(r.FormValue("offset"), 0),
				Length: messapi.ToInt64(r.FormValue("length"), 16*1000*1024),
			}
			log.Ctx(r.Context()).Error().Msg("TODO: implement stdout")
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
		log.Ctx(r.Context()).Error().Msg("TODO: implement query task queues")
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
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	d.UseNumber()
	if err := d.Decode(v); err != nil {
		r.Body.Close()
		sendJSONResponse(w, errorStatus{status: 400, err: err})
		return false
	}
	r.Body.Close()
	return true
}
