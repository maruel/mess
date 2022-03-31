package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"time"
)

var started = time.Now()

func mainImpl() error {
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

	fmt.Printf("Started in %dms.\n", time.Since(started).Round(time.Millisecond)/time.Millisecond)
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
	fmt.Printf("Terminating...\n")
	// Do not save until everything completed.
	wg.Wait()
	if err := d.save(); err != nil {
		// This is a big deal.
		return fmt.Errorf("data loss! %w", err)
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "mess: %v\n", err)
		os.Exit(1)
	}
}
