package discord

import (
	"context"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Service) runScheduledWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runScheduledTick(ctx)
		}
	}
}

func (s *Service) runScheduledTick(ctx context.Context) {
	rows, err := s.repos.Scheduled.Due(ctx, time.Now().UTC(), 50)
	if err != nil {
		s.logger.Error("scheduled poll failed: %v", err)
		return
	}
	for _, row := range rows {
		settings, err := s.repos.Settings.Get(ctx, row.GuildID)
		if err != nil || !settings.FeatureEnabled(models.FeatureScheduled) {
			continue
		}
		if settings.InMaintenanceWindow(time.Now().UTC()) {
			continue
		}
		content := strings.TrimSpace(row.Content)
		if content != "" {
			if _, err := s.session.ChannelMessageSend(row.ChannelID, content); err != nil {
				s.logger.Error("scheduled send failed guild=%s channel=%s id=%d err=%v", row.GuildID, row.ChannelID, row.ID, err)
			}
		}
		next := time.Now().UTC().Add(time.Duration(row.IntervalMinutes) * time.Minute)
		if row.IntervalMinutes < 1 {
			next = time.Now().UTC().Add(1 * time.Minute)
		}
		_ = s.repos.Scheduled.MarkRan(ctx, row.ID, next)
	}
}
