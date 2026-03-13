package discord

import (
	"context"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

func (s *Service) runTicketWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runTicketAutoCloseTick(ctx)
		}
	}
}

func (s *Service) runTicketAutoCloseTick(ctx context.Context) {
	guilds, err := s.ListGuilds(ctx)
	if err != nil {
		return
	}
	for _, g := range guilds {
		settings, err := s.repos.Settings.Get(ctx, g.ID)
		if err != nil {
			continue
		}
		if !settings.FeatureEnabled(models.FeatureTickets) || settings.TicketAutoCloseMinutes <= 0 {
			continue
		}
		if settings.InMaintenanceWindow(time.Now().UTC()) {
			continue
		}
		cutoff := time.Now().UTC().Add(-time.Duration(settings.TicketAutoCloseMinutes) * time.Minute)
		rows, err := s.repos.Tickets.OpenWithLastActivityBefore(ctx, g.ID, cutoff, 100)
		if err != nil {
			continue
		}
		for _, t := range rows {
			_ = s.closeTicket(ctx, t, settings, "auto-closed for inactivity")
		}
	}
}
