package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/maruel/mess/internal/model"
	"github.com/maruel/mess/third_party/ui2/dist"
	"github.com/maruel/panicparse/v2/stack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type server struct {
	// Immutable
	local   bool
	version string
	cid     string
	allowed map[string]struct{}

	tables model.Tables
	sched  scheduler
	l      net.Listener

	mu        sync.Mutex
	authCache map[string]*userInfo
}

func (s *server) start(port int) error {
	var err error
	suffix := fmt.Sprintf(":%d", port)
	if s.local {
		// See https://github.com/golang/go/issues/22826 and example workaround at
		// https://go-review.googlesource.com/c/crypto/+/42496
		if s.l, err = net.Listen("tcp", "127.0.0.1"+suffix); err != nil {
			s.l, err = net.Listen("tcp", "[::1]"+suffix)
		}
	} else {
		s.l, err = net.Listen("tcp", suffix)
	}
	return err
}

var uiFS = http.FileServer(http.FS(dist.FS))

// countingWriter wraps a http.ResponseWriter.
type countingWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (c *countingWriter) WriteHeader(code int) {
	if c.status == -1 {
		c.ResponseWriter.WriteHeader(code)
		c.status = code
	}
}

func (c *countingWriter) Write(buf []byte) (int, error) {
	if c.status == -1 {
		c.status = 200
	}
	n, err := c.ResponseWriter.Write(buf)
	c.length += n
	return n, err
}

func (c *countingWriter) Unwrap() http.ResponseWriter {
	return c.ResponseWriter
}

func (c *countingWriter) Flush() {
	c.ResponseWriter.(http.Flusher).Flush()
}

var reqID uint64

// wrapLog wraps logging with zerolog.
//
// Reduce function calls by embedding a lot of the features of zerolog/hlog
// inline.
func wrapLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		cw := countingWriter{ResponseWriter: w, status: -1}
		// Create a copy of the logger (including internal context slice)
		// to prevent data race when using UpdateContext.
		l := log.With().Logger()
		r = r.WithContext(l.WithContext(r.Context()))
		p := r.URL.Path
		m := r.Method
		ip := r.RemoteAddr
		ip2 := getRemoteIP(r)
		ref := r.Header.Get("Referer")
		ua := r.Header.Get("User-Agent")
		if bot := r.Header.Get("X-Luci-Swarming-Bot-ID"); bot != "" {
			l.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("bot", bot)
			})
		}
		// Instead of UUID, use a monotonically increasing request id. Use
		// something smarter once needed.
		rid := atomic.AddUint64(&reqID, 1)
		l.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Uint64("rid", rid)
		})
		defer func() {
			var line *zerolog.Event
			msg := ""
			if v := recover(); v != nil {
				msg = "panic"
				if err, ok := v.(error); ok {
					line = l.Error().Err(err)
				} else {
					line = l.Error().Str("recovered", fmt.Sprintf("%v", v))
				}
				// TODO(maruel): Send an monitoring alert.
				rawStack := debug.Stack()
				st, _, err := stack.ScanSnapshot(bytes.NewReader(rawStack), ioutil.Discard, stack.DefaultOpts())
				if err != nil || len(st.Goroutines) != 1 {
					line = line.Str("stack", string(rawStack))
				} else {
					// Calculate alignment.
					srcLen := 0
					pkgLen := 0
					for _, line := range st.Goroutines[0].Stack.Calls {
						if l := len(fmt.Sprintf("%s:%d", line.SrcName, line.Line)); l > srcLen {
							srcLen = l
						}
						if l := len(filepath.Base(line.Func.ImportPath)); l > pkgLen {
							pkgLen = l
						}
					}
					buf := bytes.Buffer{}
					for _, line := range st.Goroutines[0].Stack.Calls {
						fmt.Fprintf(
							&buf,
							"    %-*s %-*s %s(%s)\n",
							pkgLen, line.Func.DirName, srcLen,
							fmt.Sprintf("%s:%d", line.SrcName, line.Line),
							line.Func.Name, &line.Args)
					}
					if st.Goroutines[0].Stack.Elided {
						io.WriteString(&buf, "    (...)\n")
					}
					line = line.Str("parsed", buf.String())
				}
			} else {
				line = l.Info()
			}
			line = line.Int("s", cw.status).Int("l", cw.length).
				Dur("ms", time.Since(start).Round(time.Millisecond/10)).
				Str("p", p).
				Str("m", m).
				Str("ip", ip)
			if !strings.HasPrefix(ip, ip2) {
				// If the IP address was forwarded by a transparent HTTP proxy.
				line = line.Str("ip2", ip2)
			}
			if ua != "" {
				line = line.Str("ua", ua)
			}
			if ref != "" {
				line = line.Str("ref", ref)
			}
			line.Msg(msg)
		}()
		h.ServeHTTP(&cw, r)
	})
}

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

	w := &http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Handler:     wrapLog(&mux),
		ReadTimeout: time.Minute,
		IdleTimeout: 10 * time.Minute,
		// TODO(maruel): TLSConfig: tlscfg,
	}
	go w.Serve(s.l)
}

func (s *server) serveUI(page string, w http.ResponseWriter, r *http.Request) {
	// TODO(maruel): Do once on startup. No need to do it repeatedly.
	raw, _ := dist.FS.ReadFile("public_" + page + "_index.html")
	raw = bytes.ReplaceAll(raw, []byte("{{client_id}}"), []byte(s.cid))
	w.Header().Set("Content-Type", "text/html")
	http.ServeContent(w, r, "", started, bytes.NewReader(raw))
	// Simple version not replacing content:
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
var errWrongMethod = errors.New("method Not Allowed")

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

// getRemoteIP returns the IP address, handling transparent HTTP proxy.
//
// Note that this can be faked, so it cannot be used for authentication.
func getRemoteIP(r *http.Request) string {
	//if s := r.Header.Get("Forwarded"); s != "" {
	//	return s
	//}
	if s := r.Header.Get("X-Forwarded-IP"); s != "" {
		return s
	}
	if s := r.Header.Get("X-Forwarded-For"); s != "" {
		return s
	}
	return r.RemoteAddr
}

// isLocal returns if the HTTP request comes from a local IP address.
func isLocal(r *http.Request) bool {
	// TODO(maruel): Very cheezy way to only allow local bots for now.
	// TODO(maruel): Allow local IPv6.
	if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1:") {
		return false
	}
	// Check for local transparent HTTP proxy.
	if !strings.HasPrefix(getRemoteIP(r), "127.0.0.1") {
		return false
	}
	return true
}
