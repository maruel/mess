package main

import (
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

	s := server{tables: &d.tables}
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
