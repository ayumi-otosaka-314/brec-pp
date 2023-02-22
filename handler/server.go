package handler

import (
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	logger     *zap.Logger
	listenAddr string
	mux        *http.ServeMux
}

func NewServer(logger *zap.Logger, listenAddr string, mux *http.ServeMux) *Server {
	return &Server{
		logger:     logger,
		listenAddr: listenAddr,
		mux:        mux,
	}
}

func (s *Server) Serve() {
	s.logger.Info("starting server", zap.String("listenAddress", s.listenAddr))
	s.logger.Warn("shutting down server", zap.Error(http.ListenAndServe(s.listenAddr, s.mux)))
}
