package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
)

func mainImpl() error {
	port := flag.Int("port", 80, "HTTP port")
	flag.Parse()

	d := db{}
	if err := d.init(); err != nil {
		return err
	}
	raw, err := os.ReadFile("db.json")
	if os.IsNotExist(err) {
	} else if err != nil {
		return err
	} else {
		if err := d.load(bytes.NewReader(raw)); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	s := server{db: &d}
	if err := s.start(*port); err != nil {
		return err
	}

	go s.serve(ctx)
	<-ctx.Done()
	fmt.Printf("Terminating...\n")

	b := bytes.Buffer{}
	if err := d.save(&b); err != nil {
		// This is a big deal.
		return fmt.Errorf("data loss! %w", err)
	}
	if err := os.WriteFile("db.json", b.Bytes(), 0o644); err != nil {
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
