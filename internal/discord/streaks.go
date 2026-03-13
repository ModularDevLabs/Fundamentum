package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleStreakMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.GuildID == "" {
		return
	}
	if !settings.StreaksEnabled || !settings.FeatureEnabled(models.FeatureStreaks) {
		return
	}
	day := time.Now().UTC().Format("2006-01-02")
	row, newDay, err := s.repos.Streaks.UpsertDailyActivity(ctx, m.GuildID, m.Author.ID, day)
	if err != nil || !newDay {
		return
	}
	if settings.StreakRewardCoins > 0 {
		_ = s.repos.Economy.AddBalance(ctx, m.GuildID, m.Author.ID, settings.StreakRewardCoins)
	}
	if settings.StreakRewardXP > 0 {
		_, _, _ = s.repos.Leveling.AddXPIfDue(ctx, m.GuildID, m.Author.ID, m.Author.Username, settings.StreakRewardXP, 0, settings.LevelingCurve, settings.LevelingXPBase)
	}
	if row.CurrentStreak > 1 && row.CurrentStreak%7 == 0 {
		_, _ = s.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔥 %s reached a %d-day activity streak!", m.Author.Mention(), row.CurrentStreak))
	}
}
