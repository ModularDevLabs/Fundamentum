package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (s *Service) OnVoiceStateUpdate(_ *discordgo.Session, evt *discordgo.VoiceStateUpdate) {
	if evt == nil || evt.GuildID == "" || evt.UserID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	settings, err := s.repos.Settings.Get(ctx, evt.GuildID)
	if err != nil || !settings.VoiceRewardsEnabled {
		return
	}
	key := fmt.Sprintf("%s:%s", evt.GuildID, evt.UserID)
	now := time.Now().UTC()

	s.voiceMu.Lock()
	start, had := s.voiceJoined[key]
	if evt.ChannelID != "" {
		if !had {
			s.voiceJoined[key] = now
			s.voiceMu.Unlock()
			return
		}
		s.voiceMu.Unlock()
		return
	}
	if had {
		delete(s.voiceJoined, key)
	}
	s.voiceMu.Unlock()
	if !had {
		return
	}
	minutes := int(now.Sub(start).Minutes())
	if minutes <= 0 {
		return
	}
	if settings.VoiceRewardCoinsPerMinute > 0 {
		_ = s.repos.Economy.AddBalance(ctx, evt.GuildID, evt.UserID, minutes*settings.VoiceRewardCoinsPerMinute)
	}
	if settings.VoiceRewardXPPerMinute > 0 {
		username := evt.UserID
		if evt.Member != nil && evt.Member.User != nil && evt.Member.User.Username != "" {
			username = evt.Member.User.Username
		}
		_, _, _ = s.repos.Leveling.AddXPIfDue(ctx, evt.GuildID, evt.UserID, username, minutes*settings.VoiceRewardXPPerMinute, 0, settings.LevelingCurve, settings.LevelingXPBase)
	}
}
