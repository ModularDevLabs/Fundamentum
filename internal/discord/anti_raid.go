package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Service) antiRaidRecordJoin(guildID string, now time.Time, windowSec int) int {
	cutoff := now.Add(-time.Duration(windowSec) * time.Second)
	s.raidMu.Lock()
	defer s.raidMu.Unlock()
	existing := s.raidJoins[guildID]
	kept := make([]time.Time, 0, len(existing)+1)
	for _, ts := range existing {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	kept = append(kept, now)
	s.raidJoins[guildID] = kept
	return len(kept)
}

func (s *Service) antiRaidActivate(guildID string, now time.Time, cooldownMinutes int) {
	s.raidMu.Lock()
	defer s.raidMu.Unlock()
	s.raidUntil[guildID] = now.Add(time.Duration(cooldownMinutes) * time.Minute)
}

func (s *Service) antiRaidActive(guildID string, now time.Time) bool {
	s.raidMu.Lock()
	defer s.raidMu.Unlock()
	until, ok := s.raidUntil[guildID]
	return ok && now.Before(until)
}

func (s *Service) handleAntiRaidOnJoin(ctx context.Context, guildID, userID string, settings models.GuildSettings) {
	if !settings.FeatureEnabled(models.FeatureAntiRaid) {
		return
	}
	now := time.Now().UTC()
	count := s.antiRaidRecordJoin(guildID, now, settings.AntiRaidWindowSeconds)
	if count >= settings.AntiRaidJoinThreshold {
		s.antiRaidActivate(guildID, now, settings.AntiRaidCooldownMinutes)
		s.emitAuditEvent(guildID, "anti_raid_trigger", fmt.Sprintf("Join spike detected: %d joins in %ds", count, settings.AntiRaidWindowSeconds))
		if settings.AntiRaidAlertChannelID != "" {
			_, _ = s.session.ChannelMessageSend(settings.AntiRaidAlertChannelID, fmt.Sprintf("Anti-raid activated: %d joins in %d seconds. Cooldown %d minutes.", count, settings.AntiRaidWindowSeconds, settings.AntiRaidCooldownMinutes))
		}
	}
	if !s.antiRaidActive(guildID, now) {
		return
	}
	switch settings.AntiRaidAction {
	case "quarantine":
		row := models.ActionRow{
			GuildID:      guildID,
			ActorUserID:  "anti_raid",
			TargetUserID: userID,
			Type:         "quarantine",
			PayloadJSON:  `{"reason":"Anti-raid protection triggered"}`,
		}
		if _, err := s.repos.Actions.Enqueue(ctx, row); err == nil {
			s.NotifyActionQueued()
		}
	default:
		if settings.UnverifiedRoleID != "" {
			_ = s.session.GuildMemberRoleAdd(guildID, userID, settings.UnverifiedRoleID)
		}
	}
}
