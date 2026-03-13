package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Service) runRemindersWorker(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		due, err := s.repos.Reminders.ListDue(ctx, time.Now().UTC(), 100)
		if err != nil {
			s.logger.Error("reminders due list failed: %v", err)
			continue
		}
		for _, item := range due {
			settings, err := s.repos.Settings.Get(ctx, item.GuildID)
			if err != nil || !settings.FeatureEnabled(models.FeatureReminders) {
				continue
			}
			if settings.InMaintenanceWindow(time.Now().UTC()) {
				continue
			}
			channelID := strings.TrimSpace(item.ChannelID)
			if channelID == "" {
				channelID = strings.TrimSpace(settings.RemindersChannelID)
			}
			if channelID == "" {
				continue
			}
			if _, err := s.session.ChannelMessageSend(channelID, fmt.Sprintf("⏰ Reminder: %s", item.Content)); err != nil {
				s.logger.Error("reminder send failed id=%d guild=%s channel=%s err=%v", item.ID, item.GuildID, channelID, err)
				continue
			}
			_ = s.repos.Reminders.MarkSent(ctx, item.ID)
		}
	}
}
