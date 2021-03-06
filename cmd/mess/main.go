package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/maruel/mess/internal/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var started = time.Now()

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown version/" + runtime.Version()
	}
	commit := ""
	timestamp := ""
	tainted := false
	for _, s := range info.Settings {
		if s.Key == "vcs.time" {
			timestamp = s.Value
		} else if s.Key == "vcs.revision" {
			commit = s.Value
		} else if s.Key == "vcs.modified" && s.Value == "true" {
			tainted = true
		}
	}
	if commit == "" || timestamp == "" {
		return "parsing error/" + runtime.Version()
	}
	s := timestamp[:16] + "-" + commit[:10]
	if tainted {
		s += "-tainted"
	}
	return s + "/" + info.GoVersion
}

func configureLog() {
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		return short + ":" + strconv.Itoa(line)
	}
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.Logger.With().Caller().Logger()
}

// watchExe watches the main executable and cancel the returned context when
// touched.
//
// The assumption is that the executable is run as a systemd service unit or
// something similar that will restart the service.
func watchExe(ctx context.Context) (context.Context, func(), error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, nil, err
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}
	if err := w.Add(exe); err != nil {
		w.Close()
		return nil, nil, err
	}
	log.Debug().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Str("file", exe).Msg("file watching")
	ctx2, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case e := <-w.Events:
			log.Warn().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Interface("ev", e).Str("file", exe).Msg("file watching event")
		case err2 := <-w.Errors:
			log.Warn().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Err(err2).Str("file", exe).Msg("file watching error")
		case <-ctx2.Done():
		}
		w.Close()
		cancel()
	}()
	return ctx2, cancel, err
}

func mainImpl() error {
	configureLog()
	port := flag.Int("port", 7899, "HTTP port for the web server to listen to")
	local := flag.Bool("local", false, "Bind local, allow everyone to be admin; useful for local testing the UI")
	cid := flag.String("cid", "", "Google OAuth2 Client ID")
	usr := flag.String("usr", "", "Comma separated users allowed access")

	flag.Parse()

	if *cid == "" {
		fmt.Printf("Warning: you should pass -cid\n")
		fmt.Printf("\n")
		fmt.Printf("- Visit https://console.cloud.google.com/apis/credentials\n")
		fmt.Printf("- Expose this server to the internet, preferably fronted with Caddy.\n")
		fmt.Printf("- Create ID client Oauth 2.0 for a Web Application.\n")
		fmt.Printf("- Javascript Origin: https://<domain.com\n")
		fmt.Printf("- Redirection: https://<domain.com/oauth2callback\n")
		fmt.Printf("\n")
	} else if b, _ := regexp.MatchString(`^[a-z0-9\-]{10,}\.apps\.googleusercontent\.com$`, *cid); !b {
		fmt.Printf("Warning: the client id passed to -cid doesn't like the expected form.\n")
		fmt.Printf("\n")
		fmt.Printf("It should look like 1111-aaaaaaaaaaa.apps.googleusercontent.com\n")
		fmt.Printf("\n")
	}

	allowed := map[string]struct{}{}
	for _, u := range strings.Split(*usr, ",") {
		allowed[u] = struct{}{}
	}
	if len(allowed) == 0 {
		fmt.Printf("Warning: No user is allowed access.\n")
		fmt.Printf("\n")
		fmt.Printf("Use the -usr flag to allow Google Accounts.\n")
		fmt.Printf("\n")
	}

	outputs, err := model.NewTaskOutputs("outputs")
	if err != nil {
		return err
	}
	// Use one of sqlite or json DB backend.
	d, err := model.NewDBSqlite3("mess.db")
	//d, err := model.NewDBJSON("db.json.zst")
	if err != nil {
		return err
	}

	ctx, cancel, err := watchExe(context.Background())
	if err != nil {
		return err
	}
	defer cancel()
	log.Info().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Msg("Loaded DB")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	wg := sync.WaitGroup{}
	ver := getVersion()
	s := server{
		local:     *local,
		version:   ver,
		cid:       *cid,
		allowed:   allowed,
		tables:    d,
		outputs:   outputs,
		authCache: map[string]*userInfo{},
	}
	s.sched.init(d)
	wg.Add(1)
	go func() {
		s.sched.loop(ctx)
		wg.Done()
	}()
	if err := s.start(*port); err != nil {
		cancel()
		wg.Wait()
		return err
	}
	wg.Add(1)
	go func() {
		s.serve(ctx)
		wg.Done()
	}()

	log.Info().Dur("ms", time.Since(started).Round(time.Millisecond)/10).
		Str("port", s.l.Addr().String()).
		Str("version", ver).Msg("Listening")
	done := ctx.Done()
	wg.Add(1)
	go func() {
		// Intentionally save too often initially.
		for t := time.NewTimer(5 * time.Second); ; {
			select {
			case <-done:
				t.Stop()
				wg.Done()
				return
			case <-t.C:
				_ = d.Snapshot()
			}
		}
	}()
	wg.Add(1)
	go func() {
		outputs.Loop(ctx, 10000, 6*time.Minute)
		wg.Done()
	}()

	<-done
	stopping := time.Now()
	log.Info().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Msg("Terminating")
	// Do not save until everything completed.
	wg.Wait()
	err = d.Close()
	if err != nil {
		// This is a big deal.
		err = fmt.Errorf("data loss! %w", err)
	}
	log.Info().Dur("ms", time.Since(stopping).Round(time.Millisecond)/10).Msg("Saved DB")
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "mess: %v\n", err)
		os.Exit(1)
	}
}
