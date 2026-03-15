package web

import (
	"context"
	"net/http"
	"sync"
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
	bindAddr             string
	adminPass            string
	dashboardRoleSecrets map[string]string
	sessionTTL           time.Duration
	allowLegacyBearer    bool
	authProxyEnabled     bool
	authProxySecret      string
	authProxyUserHeader  string
	authProxyRoleHeader  string
	repos                *db.Repositories
	discord              *discord.Service
	logger               Logger

	httpServer *http.Server
	loginMu    sync.Mutex
	loginState map[string]loginAttemptState
}

func NewServer(bindAddr, adminPass string, dashboardRoleSecrets map[string]string, sessionTTL time.Duration, allowLegacyBearer bool, authProxyEnabled bool, authProxySecret, authProxyUserHeader, authProxyRoleHeader string, repos *db.Repositories, discordSvc *discord.Service, logger Logger) *Server {
	if dashboardRoleSecrets == nil {
		dashboardRoleSecrets = map[string]string{}
	}
	if sessionTTL <= 0 {
		sessionTTL = 8 * time.Hour
	}
	return &Server{
		bindAddr:             bindAddr,
		adminPass:            adminPass,
		dashboardRoleSecrets: dashboardRoleSecrets,
		sessionTTL:           sessionTTL,
		allowLegacyBearer:    allowLegacyBearer,
		authProxyEnabled:     authProxyEnabled,
		authProxySecret:      authProxySecret,
		authProxyUserHeader:  authProxyUserHeader,
		authProxyRoleHeader:  authProxyRoleHeader,
		repos:                repos,
		discord:              discordSvc,
		logger:               logger,
		loginState:           map[string]loginAttemptState{},
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:              s.bindAddr,
		Handler:           s.loggingMiddleware(s.securityHeadersMiddleware(mux)),
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
