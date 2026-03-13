package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

func (s *Service) runAnalyticsWorker(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	lastSent := map[string]time.Time{}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			guilds, err := s.ListGuilds(ctx)
			if err != nil {
				continue
			}
			now := time.Now().UTC()
			for _, g := range guilds {
				settings, err := s.repos.Settings.Get(ctx, g.ID)
				if err != nil || !settings.FeatureEnabled(models.FeatureAnalytics) || settings.AnalyticsChannelID == "" {
					continue
				}
				if settings.InMaintenanceWindow(now) {
					continue
				}
				interval := time.Duration(settings.AnalyticsIntervalDays) * 24 * time.Hour
				if interval <= 0 {
					interval = 7 * 24 * time.Hour
				}
				if last, ok := lastSent[g.ID]; ok && now.Sub(last) < interval {
					continue
				}
				if err := s.sendAnalyticsSummary(ctx, g.ID, settings, now.Add(-interval), now); err != nil {
					s.logger.Error("analytics send failed guild=%s err=%v", g.ID, err)
					continue
				}
				lastSent[g.ID] = now
			}
		}
	}
}

func (s *Service) sendAnalyticsSummary(ctx context.Context, guildID string, settings models.GuildSettings, since, until time.Time) error {
	tracked, err := s.repos.Activity.CountTracked(ctx, guildID)
	if err != nil {
		return err
	}
	inactive, err := s.repos.Activity.CountInactiveBefore(ctx, guildID, until.AddDate(0, 0, -settings.InactiveDays))
	if err != nil {
		return err
	}
	warnings, err := s.repos.Warnings.CountSince(ctx, guildID, since)
	if err != nil {
		return err
	}
	actions, err := s.repos.Actions.CountSince(ctx, guildID, since, "")
	if err != nil {
		return err
	}
	failedActions, err := s.repos.Actions.CountSince(ctx, guildID, since, "failed")
	if err != nil {
		return err
	}
	ticketsOpened, err := s.repos.Tickets.CountCreatedSince(ctx, guildID, since)
	if err != nil {
		return err
	}
	ticketsOpen, err := s.repos.Tickets.CountByStatus(ctx, guildID, "open")
	if err != nil {
		return err
	}
	msg := fmt.Sprintf(
		"Analytics summary (%s to %s)\nTracked: %d\nInactive: %d\nWarnings: %d\nActions: %d (failed=%d)\nTickets opened: %d\nTickets open: %d",
		since.Format("2006-01-02"),
		until.Format("2006-01-02"),
		tracked, inactive, warnings, actions, failedActions, ticketsOpened, ticketsOpen,
	)
	_, err = s.session.ChannelMessageSend(settings.AnalyticsChannelID, msg)
	return err
}
