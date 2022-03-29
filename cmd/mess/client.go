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

func serveUI(page string, w http.ResponseWriter, r *http.Request) {
	// strings.ReplaceAll(raw, "{{client_id}}", ui_client_id)
	// http.ServeContent()
	r.URL.Path = "/public_" + page + "_index.html"
	uiFS.ServeHTTP(w, r)
}

func rootUIPages(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		serveUI("swarming", w, r)
	} else {
		serveUI(r.URL.Path[1:], w, r)
	}
}

func (s *server) serve(ctx context.Context) {
	mux := http.ServeMux{}
	// APIs.
	mux.Handle("/swarming/api/v1/bot/", http.StripPrefix("/swarming/api/v1/bot", http.HandlerFunc(s.apiBot)))
	mux.Handle("/_ah/api/swarming/v1/", http.StripPrefix("/_ah/api/swarming/v1", http.HandlerFunc(s.apiCloudEndpoint)))
	mux.HandleFunc("/bot_code", rootUIPages)
	// UI.
	mux.Handle("/newres/", http.StripPrefix("/newres", uiFS))
	mux.HandleFunc("/bot", rootUIPages)
	mux.HandleFunc("/botlist", rootUIPages)
	mux.HandleFunc("/task", rootUIPages)
	mux.HandleFunc("/tasklist", rootUIPages)
	mux.HandleFunc("/", rootUIPages)
	h := &http.Server{
		BaseContext:  func(net.Listener) context.Context { return ctx },
		Handler:      &loghttp.Handler{Handler: &mux},
		ReadTimeout:  10. * time.Second,
		WriteTimeout: time.Minute,
	}
	go h.Serve(s.l)
}

func (s *server) apiBot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/server_ping" {
		w.Write([]byte("Server Up"))
		return
	}
	if r.URL.Path == "/bot_code" {
		version := internal.GetBotVersion()
		http.Redirect(w, r, "/swarming/api/v1/bot/bot_code/"+version, http.StatusFound)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/bot_code") {
		version := internal.GetBotVersion()
		if r.URL.Path[len("/bot_code/"):] != version {
			// Log a warning.
			http.Redirect(w, r, "/swarming/api/v1/bot/bot_code/"+version, http.StatusFound)
			return
		}
		http.ServeContent(w, r, "swarming_bot.zip", started, bytes.NewReader(internal.GetBotZIP([]byte("{}"))))
		return
	}
	// /handshake
	// /poll
	// /event
	// /oauth_token
	// /id_token
	// /task_update
	// /task_error
	w.Write([]byte("TODO 1"))
}

func (s *server) apiCloudEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("TODO 2"))
}

func (s *server) task(key int64) {
	r := s.tables.Requests[key]
	if r == nil {
	}
}

func (s *server) botCode(key int64) {
	_ = internal.GetBotZIP(nil)
}
