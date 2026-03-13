package web

import (
	"context"
	"net/http"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/db"
	"github.com/ModularDevLabs/Fundamentum/internal/discord"
)

type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
}

type EventLogger interface {
	RecentEvents(limit int) []string
}

type Server struct {
	bindAddr  string
	adminPass string
	repos     *db.Repositories
	discord   *discord.Service
	logger    Logger

	httpServer *http.Server
}

func NewServer(bindAddr, adminPass string, repos *db.Repositories, discordSvc *discord.Service, logger Logger) *Server {
	return &Server{
		bindAddr:  bindAddr,
		adminPass: adminPass,
		repos:     repos,
		discord:   discordSvc,
		logger:    logger,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:              s.bindAddr,
		Handler:           s.loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	s.logger.Info("web server listening on %s", s.bindAddr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
