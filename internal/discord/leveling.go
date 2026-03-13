package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleLevelingMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.FeatureEnabled(models.FeatureLeveling) {
		return
	}
	addXP := settings.LevelingXPPerMessage
	if addXP <= 0 {
		addXP = 10
	}
	cooldown := settings.LevelingCooldownSeconds
	if cooldown <= 0 {
		cooldown = 60
	}
	curve := settings.LevelingCurve
	if curve == "" {
		curve = "quadratic"
	}
	base := settings.LevelingXPBase
	if base <= 0 {
		base = 100
	}

	row, leveledUp, err := s.repos.Leveling.AddXPIfDue(ctx, m.GuildID, m.Author.ID, m.Author.Username, addXP, cooldown, curve, base)
	if err != nil {
		s.logger.Error("leveling update failed guild=%s user=%s err=%v", m.GuildID, m.Author.ID, err)
		return
	}
	if !leveledUp {
		return
	}

	msg := fmt.Sprintf("%s leveled up to **%d** (%d XP).", m.Author.Mention(), row.Level, row.XP)
	announceChannel := strings.TrimSpace(settings.LevelingChannelID)
	if announceChannel == "" {
		announceChannel = m.ChannelID
	}
	_, _ = s.session.ChannelMessageSend(announceChannel, msg)
}
