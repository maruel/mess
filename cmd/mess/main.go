package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
)

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
	}
	d := db{}
	if err := d.load(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	s := server{tables: &d.tables, cid: *cid}
	if err := s.start(*port); err != nil {
		return err
	}

	go s.serve(ctx)
	<-ctx.Done()
	fmt.Printf("Terminating...\n")

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
