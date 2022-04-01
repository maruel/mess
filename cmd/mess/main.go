package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var started = time.Now()

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

func mainImpl() error {
	configureLog()
	port := flag.Int("port", 7899, "HTTP port")
	cid := flag.String("cid", "", "Google OAuth2 Client ID")
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
	d := db{}
	if err := d.load(); err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	wg.Add(1)
	go func() {
		<-c
		cancel()
		wg.Done()
	}()

	s := server{tables: &d.tables, cid: *cid}
	if err := s.start(*port); err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		s.serve(ctx)
		wg.Done()
	}()

	log.Info().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Msg("Loaded DB")
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
				d.save()
			}
		}
	}()

	<-done
	stopping := time.Now()
	log.Info().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Msg("Terminating")
	// Do not save until everything completed.
	wg.Wait()
	err := d.save()
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
