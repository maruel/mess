package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/maruel/mess/internal"
	"github.com/maruel/mess/internal/model"
	"github.com/maruel/mess/messapi"
)

func (s *server) apiEndpoint(w http.ResponseWriter, r *http.Request) {
	// TODO(maruel): Cleaner code.
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
	// TODO(maruel): /server/token.

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
	// TODO(maruel): /bots/...

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
	// TODO(maruel): /tasks/...

	if strings.HasPrefix(r.URL.Path, "/bot/") {
		if n := strings.SplitN(r.URL.Path[len("/bot/"):], "/", 2); len(n) == 2 {
			id := n[0]
			switch n[1] {
			case "delete":
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
			}
		}
	}

	// /task
	// /queues
	// /config
	sendJSONResponse(w, errorStatus{status: 404, err: errUnknownAPI})
}
