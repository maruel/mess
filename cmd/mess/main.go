package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
)

func mainImpl() error {
	d := db{}
	raw, err := os.ReadFile("db.json")
	if err != nil {
		return err
	}
	if err := d.load(bytes.NewReader(raw)); err != nil {
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

	port := 51234
	s := server{}
	if err := s.start(port); err != nil {
		return err
	}

	go s.serve(ctx)
	<-ctx.Done()
	//s.Shutdown(context.Background())

	b := bytes.Buffer{}
	if err := d.save(&b); err != nil {
		// This is a big deal.
		return err
	}
	return os.WriteFile("db.json", b.Bytes(), 0o644)
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "mess: %v\n", err)
		os.Exit(1)
	}
}
