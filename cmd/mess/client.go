package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type server struct {
	l net.Listener
}

func (s *server) start(port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s.l = l
	return nil
}

func (s *server) serve(ctx context.Context) {
	h := &http.Server{
		BaseContext:  func(net.Listener) context.Context { return ctx },
		Handler:      s,
		ReadTimeout:  10. * time.Second,
		WriteTimeout: time.Minute,
	}
	go h.Serve(s.l)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (s *server) task() {
}
