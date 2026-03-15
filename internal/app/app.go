package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/db"
	"github.com/ModularDevLabs/Fundamentum/internal/discord"
	"github.com/ModularDevLabs/Fundamentum/internal/web"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	cfg    ProcessConfig
	logger *Logger

	db      *sql.DB
	repos   *db.Repositories
	discord *discord.Service
	web     *web.Server
}

func New(cfg ProcessConfig) (*App, error) {
	logger := NewLogger(cfg.LogLevel)

	dbConn, err := db.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(dbConn); err != nil {
		return nil, err
	}

	repos := db.NewRepositories(dbConn)
	if err := ensureDashboardUsers(context.Background(), repos, cfg.AdminPassword, cfg.DashboardRoleSecret); err != nil {
		return nil, err
	}

	discordSvc, err := discord.NewService(cfg.DiscordToken, repos, logger)
	if err != nil {
		return nil, err
	}

	webServer := web.NewServer(cfg.BindAddr, cfg.AdminPassword, cfg.DashboardRoleSecret, cfg.DashboardSessionTTL, cfg.AllowLegacyBearer, cfg.AuthProxyEnabled, cfg.AuthProxySecret, cfg.AuthProxyUserHeader, cfg.AuthProxyRoleHeader, repos, discordSvc, logger)

	return &App{
		cfg:     cfg,
		logger:  logger,
		db:      dbConn,
		repos:   repos,
		discord: discordSvc,
		web:     webServer,
	}, nil
}

func ensureDashboardUsers(ctx context.Context, repos *db.Repositories, adminPassword string, roleSecrets map[string]string) error {
	adminHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := repos.DashboardAuth.UpsertUser(ctx, db.DashboardUserRow{
		Username:     "admin",
		PasswordHash: string(adminHash),
		Role:         "admin",
		Enabled:      true,
	}); err != nil {
		return err
	}
	for role, secret := range roleSecrets {
		role = strings.ToLower(strings.TrimSpace(role))
		secret = strings.TrimSpace(secret)
		if role == "" || role == "admin" || secret == "" {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if err := repos.DashboardAuth.UpsertUser(ctx, db.DashboardUserRow{
			Username:     role,
			PasswordHash: string(hash),
			Role:         role,
			Enabled:      true,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.discord.Open(); err != nil {
		return err
	}
	defer a.discord.Close()

	a.discord.StartWorkers(ctx)

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- a.web.Start()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.web.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-srvErr:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
