package discord

import (
	"context"
	"time"
)

func (s *Service) runRetentionWorker(ctx context.Context) {
	timer := time.NewTimer(2 * time.Minute)
	defer timer.Stop()

	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.runRetentionPass(ctx)
		case <-ticker.C:
			s.runRetentionPass(ctx)
		}
	}
}

func (s *Service) runRetentionPass(ctx context.Context) {
	guildIDs, err := s.repos.Settings.ListGuildIDs(ctx)
	if err != nil {
		s.logger.Error("retention list guilds failed: %v", err)
		return
	}
	now := time.Now().UTC()
	for _, guildID := range guildIDs {
		cfg, err := s.repos.Settings.Get(ctx, guildID)
		if err != nil {
			s.logger.Error("retention load settings failed guild=%s: %v", guildID, err)
			continue
		}
		if cfg.RetentionDays <= 0 {
			continue
		}
		cutoff := now.Add(-time.Duration(cfg.RetentionDays) * 24 * time.Hour)
		preview, err := s.repos.Retention.CountGuildBefore(ctx, guildID, cutoff)
		if err != nil {
			s.logger.Error("retention preview failed guild=%s: %v", guildID, err)
			continue
		}
		if cfg.RetentionArchiveBeforePurge {
			if err := s.repos.Retention.RecordArchiveEvent(ctx, guildID, cutoff, preview); err != nil {
				s.logger.Error("retention archive event failed guild=%s: %v", guildID, err)
			}
		}
		if _, err := s.repos.Retention.PurgeGuildBefore(ctx, guildID, cutoff); err != nil {
			s.logger.Error("retention purge failed guild=%s: %v", guildID, err)
			continue
		}
	}
}
