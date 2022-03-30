package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/maruel/mess/internal"
	"github.com/maruel/mess/third_party/ui2/dist"
	"github.com/maruel/serve-dir/loghttp"
)

var started = time.Now()

type server struct {
	tables *tables
	cid    string
	l      net.Listener
}

func (s *server) start(port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s.l = l
	return nil
}

var uiFS = http.FileServer(http.FS(dist.FS))

func (s *server) serve(ctx context.Context) {
	mux := http.ServeMux{}
	// APIs.
	mux.Handle("/swarming/api/v1/bot/", http.StripPrefix("/swarming/api/v1/bot", http.HandlerFunc(s.apiBot)))
	mux.Handle("/_ah/api/swarming/v1/", http.StripPrefix("/_ah/api/swarming/v1", http.HandlerFunc(s.apiEndpoint)))
	mux.HandleFunc("/bot_code", s.apiBot)
	// UI.
	mux.Handle("/newres/", http.StripPrefix("/newres", uiFS))
	mux.HandleFunc("/bot", s.rootUIPages)
	mux.HandleFunc("/botlist", s.rootUIPages)
	mux.HandleFunc("/task", s.rootUIPages)
	mux.HandleFunc("/tasklist", s.rootUIPages)
	mux.HandleFunc("/", s.rootUIPages)
	h := &http.Server{
		BaseContext:  func(net.Listener) context.Context { return ctx },
		Handler:      &loghttp.Handler{Handler: &mux},
		ReadTimeout:  10. * time.Second,
		WriteTimeout: time.Minute,
	}
	go h.Serve(s.l)
}

func (s *server) serveUI(page string, w http.ResponseWriter, r *http.Request) {
	// TODO(maruel): Do once on startup.
	raw, _ := dist.FS.ReadFile("public_" + page + "_index.html")
	raw = bytes.ReplaceAll(raw, []byte("{{client_id}}"), []byte(s.cid))
	w.Header().Set("Content-Type", "text/html")
	http.ServeContent(w, r, "", started, bytes.NewReader(raw))
	//r.URL.Path = "/public_" + page + "_index.html"
	//uiFS.ServeHTTP(w, r)
}

func (s *server) rootUIPages(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		s.serveUI("swarming", w, r)
	} else {
		s.serveUI(r.URL.Path[1:], w, r)
	}
}

func (s *server) apiBot(w http.ResponseWriter, r *http.Request) {
	// Non-API URLs.
	h := w.Header()
	if r.URL.Path == "/server_ping" {
		h.Set("Content-Type", "text/plain")
		w.Write([]byte("Server Up"))
		return
	}
	if r.URL.Path == "/bot_code" {
		version := internal.GetBotVersion(getHost(r))
		http.Redirect(w, r, "/swarming/api/v1/bot/bot_code/"+version, http.StatusFound)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/bot_code") {
		version := internal.GetBotVersion(getHost(r))
		if r.URL.Path[len("/bot_code/"):] != version {
			// Log a warning.
			http.Redirect(w, r, "/swarming/api/v1/bot/bot_code/"+version, http.StatusFound)
			return
		}
		// TODO(maruel): Doesn't work??
		h.Set("Content-Disposition", "attachment; filename=swarming_bot.zip")
		http.ServeContent(w, r, "swarming_bot.zip", started, bytes.NewReader(internal.GetBotZIP(getHost(r))))
		return
	}

	// API URLs.
	h.Set("Content-Type", "application/json")
	if r.URL.Path == "/handshake" {
		w.Write([]byte("{}"))
		return
	}
	if r.URL.Path == "/poll" {
		w.Write([]byte("{}"))
		return
	}
	if r.URL.Path == "/event" {
		w.Write([]byte("{}"))
		return
	}
	if r.URL.Path == "/oauth_token" {
		w.Write([]byte("{}"))
		return
	}
	if r.URL.Path == "/id_token" {
		w.Write([]byte("{}"))
		return
	}
	if r.URL.Path == "/task_update" {
		w.Write([]byte("{}"))
		return
	}
	if r.URL.Path == "/task_error" {
		w.Write([]byte("{}"))
		return
	}
	w.WriteHeader(404)
	w.Write([]byte("{}"))
}

func (s *server) apiEndpoint(w http.ResponseWriter, r *http.Request) {
	// /server
	//   details
	//   token
	//   permission
	// /task
	// /tasks
	// /queues
	// /bot
	// /bots
	// /config
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(404)
	w.Write([]byte("{}"))
}

func (s *server) task(key int64) {
	r := s.tables.Requests[key]
	if r == nil {
	}
}

func getHost(r *http.Request) string {
	if r.URL.Host != "" {
		return r.URL.Host
	}
	return r.Header.Get("X-Forwarded-Host")
}
