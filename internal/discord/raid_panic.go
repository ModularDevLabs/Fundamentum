package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/db"
	"github.com/bwmarrin/discordgo"
)

type RaidPanicResult struct {
	LockdownID      int64 `json:"lockdown_id"`
	ChannelsUpdated int   `json:"channels_updated"`
}

func (s *Service) ActivateRaidPanic(ctx context.Context, guildID, actor string, durationMinutes, slowmodeSeconds int) (RaidPanicResult, error) {
	if guildID == "" {
		return RaidPanicResult{}, fmt.Errorf("missing guild id")
	}
	if durationMinutes <= 0 {
		durationMinutes = 30
	}
	if slowmodeSeconds <= 0 {
		slowmodeSeconds = 10
	}
	s.raidPanicMu.Lock()
	defer s.raidPanicMu.Unlock()

	if active, ok, err := s.repos.RaidPanic.ActiveLockdownByGuild(ctx, guildID); err != nil {
		return RaidPanicResult{}, err
	} else if ok {
		return RaidPanicResult{}, fmt.Errorf("raid panic already active (id=%d)", active.ID)
	}

	lockID, err := s.repos.RaidPanic.CreateLockdown(ctx, guildID, slowmodeSeconds, actor, time.Now().UTC().Add(time.Duration(durationMinutes)*time.Minute))
	if err != nil {
		return RaidPanicResult{}, err
	}

	channels, err := s.session.GuildChannels(guildID)
	if err != nil {
		return RaidPanicResult{}, err
	}
	updated := 0
	for _, ch := range channels {
		if ch == nil {
			continue
		}
		if ch.Type != discordgo.ChannelTypeGuildText && ch.Type != discordgo.ChannelTypeGuildNews {
			continue
		}
		_ = s.repos.RaidPanic.AddChannelState(ctx, db.RaidPanicChannelStateRow{
			LockdownID:              lockID,
			GuildID:                 guildID,
			ChannelID:               ch.ID,
			PreviousSlowmodeSeconds: ch.RateLimitPerUser,
		})
		if ch.RateLimitPerUser == slowmodeSeconds {
			continue
		}
		val := slowmodeSeconds
		if _, err := s.session.ChannelEditComplex(ch.ID, &discordgo.ChannelEdit{
			RateLimitPerUser: &val,
		}); err == nil {
			updated++
		}
	}
	return RaidPanicResult{LockdownID: lockID, ChannelsUpdated: updated}, nil
}

func (s *Service) DeactivateRaidPanic(ctx context.Context, guildID, reason string) (RaidPanicResult, error) {
	s.raidPanicMu.Lock()
	defer s.raidPanicMu.Unlock()

	active, ok, err := s.repos.RaidPanic.ActiveLockdownByGuild(ctx, guildID)
	if err != nil {
		return RaidPanicResult{}, err
	}
	if !ok {
		return RaidPanicResult{}, nil
	}
	return s.endRaidPanicLockdown(ctx, active, reason)
}

func (s *Service) RaidPanicStatus(ctx context.Context, guildID string) (db.RaidPanicLockdownRow, bool, error) {
	return s.repos.RaidPanic.ActiveLockdownByGuild(ctx, guildID)
}

func (s *Service) runRaidPanicWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		rows, err := s.repos.RaidPanic.ListDueActiveLockdowns(ctx, time.Now().UTC(), 20)
		if err != nil {
			continue
		}
		for _, row := range rows {
			s.raidPanicMu.Lock()
			_, _ = s.endRaidPanicLockdown(ctx, row, "auto-expired")
			s.raidPanicMu.Unlock()
		}
	}
}

func (s *Service) endRaidPanicLockdown(ctx context.Context, active db.RaidPanicLockdownRow, reason string) (RaidPanicResult, error) {
	states, err := s.repos.RaidPanic.ListChannelStates(ctx, active.ID)
	if err != nil {
		return RaidPanicResult{}, err
	}
	restored := 0
	for _, state := range states {
		val := state.PreviousSlowmodeSeconds
		if _, err := s.session.ChannelEditComplex(state.ChannelID, &discordgo.ChannelEdit{
			RateLimitPerUser: &val,
		}); err == nil {
			restored++
		}
	}
	_ = s.repos.RaidPanic.EndLockdown(ctx, active.ID, reason)
	return RaidPanicResult{LockdownID: active.ID, ChannelsUpdated: restored}, nil
}
