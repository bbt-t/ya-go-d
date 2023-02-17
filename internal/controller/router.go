package controller

import (
	"context"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func NewHTTPServer(address string, handler http.Handler) *Server {
	/*
		New http-server.
	*/
	return &Server{
		httpServer: &http.Server{
			Addr:    address,
			Handler: handler,
		},
	}
}

func (s *Server) UP() error {
	/*
		http-server start.
	*/
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	/*
		http-server stop.
	*/
	return s.httpServer.Shutdown(ctx)
}
