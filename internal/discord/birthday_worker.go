package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Service) runBirthdayWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	sentToday := map[string]struct{}{}

	for {
		now := time.Now().UTC()
		mmdd := now.Format("01-02")
		dayKey := now.Format("2006-01-02")

		guilds, err := s.ListGuilds(ctx)
		if err != nil {
			continue
		}
		for _, g := range guilds {
			settings, err := s.repos.Settings.Get(ctx, g.ID)
			if err != nil || !settings.FeatureEnabled(models.FeatureBirthdays) || settings.BirthdaysChannelID == "" {
				continue
			}
			if settings.InMaintenanceWindow(now) {
				continue
			}
			rows, err := s.repos.Birthdays.ListByDate(ctx, g.ID, mmdd, 200)
			if err != nil {
				continue
			}
			for _, row := range rows {
				key := fmt.Sprintf("%s:%s:%s", g.ID, row.UserID, dayKey)
				if _, ok := sentToday[key]; ok {
					continue
				}
				msg := fmt.Sprintf("🎂 Happy birthday <@%s>! Wishing you an awesome day.", row.UserID)
				if _, err := s.session.ChannelMessageSend(settings.BirthdaysChannelID, msg); err != nil {
					s.logger.Error("birthday send failed guild=%s user=%s channel=%s err=%v", g.ID, row.UserID, settings.BirthdaysChannelID, err)
					continue
				}
				sentToday[key] = struct{}{}
			}
		}

		// Keep the dedupe map bounded to recent dates.
		cutoff := now.AddDate(0, 0, -2).Format("2006-01-02")
		for key := range sentToday {
			if len(key) < 10 {
				delete(sentToday, key)
				continue
			}
			if key[len(key)-10:] < cutoff {
				delete(sentToday, key)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
