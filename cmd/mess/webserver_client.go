package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/maruel/mess/internal"
	"github.com/maruel/mess/internal/model"
	"github.com/maruel/mess/messapi"
	"github.com/rs/zerolog/log"
)

func (s *server) apiEndpoint(w http.ResponseWriter, r *http.Request) {
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
	// TODO(maruel): /config
	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiEndpointServer(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/server/details" {
		sendJSONResponse(w, messapi.ServerDetails{
			ServerVersion: serverVersion,
			BotVersion:    internal.GetBotVersion(r),
		})
		return
	}
	if r.URL.Path == "/server/permissions" {
		sendJSONResponse(w, messapi.ServerPermissions{
			DeleteBot: true,
		})
		return
	}
	if r.URL.Path == "/server/token" {
		log.Ctx(r.Context()).Error().Msg("TODO")
	}

	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiEndpointBots(w http.ResponseWriter, r *http.Request) {
	cloudNow := messapi.CloudTime(time.Now())
	if r.URL.Path == "/bots/count" {
		// TODO(maruel): Implement.
		count := s.tables.BotCount()
		sendJSONResponse(w, messapi.BotsCount{
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
		sendJSONResponse(w, messapi.BotsDimensions{
			BotsDimensions: []messapi.StringListPair{},
			Now:            cloudNow,
		})
		return
	}
	if r.URL.Path == "/bots/list" {
		// TODO(maruel): Arguments.
		// TODO(maruel): Implement.
		bots, cursor := s.tables.BotGetSlice("", 100)
		items := make([]messapi.Bot, len(bots))
		for i := range bots {
			items[i].FromDB(&bots[i])
		}
		sendJSONResponse(w, messapi.BotsList{
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
		log.Ctx(r.Context()).Error().Msg("TODO")
	}
	if r.URL.Path == "/tasks/count" {
		// TODO(maruel): Arguments.
		// TODO(maruel): Implement.
		count := s.tables.TaskRequestCount()
		sendJSONResponse(w, messapi.TasksCount{
			Count: int32(count),
			Now:   cloudNow,
		})
		return
	}
	if r.URL.Path == "/tasks/get_states" {
		log.Ctx(r.Context()).Error().Msg("TODO")
	}
	if r.URL.Path == "/tasks/list" {
		// TODO(maruel): Arguments.
		// TODO(maruel): Implement.
		sendJSONResponse(w, messapi.TasksList{
			Cursor: "",
			Items:  []messapi.TaskResult{},
			Now:    cloudNow,
		})
		return
	}
	if r.URL.Path == "/tasks/new" {
		log.Ctx(r.Context()).Error().Msg("TODO")
	}
	if r.URL.Path == "/tasks/requests" {
		log.Ctx(r.Context()).Error().Msg("TODO")
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
			log.Ctx(r.Context()).Error().Msg("TODO")
		case "events":
			all, cursor := s.tables.BotEventGetSlice(id, "", 100, time.Now(), time.Now())
			events := make([]messapi.BotEvent, len(all))
			for i := range all {
				events[i].FromDB(&all[i])
			}
			sendJSONResponse(w, messapi.BotEvents{
				Cursor: cursor,
				Items:  events,
				Now:    cloudNow,
			})
			return
		case "get":
			bi := messapi.Bot{}
			bot := model.Bot{}
			s.tables.BotGet(id, &bot)
			bi.FromDB(&bot)
			sendJSONResponse(w, bi)
			return
		case "tasks":
			sendJSONResponse(w, messapi.BotTasks{
				Cursor: "",
				Items:  []messapi.TaskResult{},
				Now:    cloudNow,
			})
			return
		case "terminate":
			log.Ctx(r.Context()).Error().Msg("TODO")
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
			log.Ctx(r.Context()).Error().Msg("TODO")
		case "request":
			log.Ctx(r.Context()).Error().Msg("TODO")
		case "result":
			log.Ctx(r.Context()).Error().Msg("TODO")
		case "stdout":
			log.Ctx(r.Context()).Error().Msg("TODO")
		}
	}

	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}

func (s *server) apiEndpointQueues(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/queues/list" {
		log.Ctx(r.Context()).Error().Msg("TODO")
	}

	log.Ctx(r.Context()).Warn().Msg("Unknown client request")
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}
