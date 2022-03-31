package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/maruel/mess/third_party/ui2/dist"
	"github.com/maruel/serve-dir/loghttp"
)

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
	// See webserver_bot.go
	mux.Handle("/swarming/api/v1/bot/", http.StripPrefix("/swarming/api/v1/bot", http.HandlerFunc(s.apiBot)))
	// See webserver_client.go
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

// API helpers.

var errUnknownAPI = errors.New("unknown API")

type apiFunc func() interface{}

type errorStatus struct {
	status int
	err    error
}

// sendJSONResponse sends a JSON response, handling errors.
func sendJSONResponse(w http.ResponseWriter, res interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err, ok := res.(error); ok {
		w.WriteHeader(500)
		res = map[string]string{"error": err.Error()}
	} else if err, ok := res.(errorStatus); ok {
		w.WriteHeader(err.status)
		s := ""
		if err.err == nil {
			s = http.StatusText(http.StatusMethodNotAllowed)
		} else {
			s = err.err.Error()
		}
		res = map[string]string{"error": s}
	}
	raw, _ := json.Marshal(res)
	w.Write(raw)
}

func getHost(r *http.Request) string {
	if r.URL.Host != "" {
		return r.URL.Host
	}
	return r.Header.Get("X-Forwarded-Host")
}

const serverVersion = "v0.0.1"
