package discord

import (
	"context"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleCustomCommandMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) bool {
	if m == nil || m.Author == nil || m.Author.Bot {
		return false
	}
	if !settings.FeatureEnabled(models.FeatureCustomCommands) {
		return false
	}
	content := strings.TrimSpace(m.Content)
	if content == "" {
		return false
	}
	cmd, ok, err := s.repos.CustomCommands.FindByTrigger(ctx, m.GuildID, strings.ToLower(content))
	if err != nil {
		s.logger.Error("custom command lookup failed guild=%s err=%v", m.GuildID, err)
		return false
	}
	if !ok || strings.TrimSpace(cmd.Response) == "" {
		return false
	}
	if _, err := s.session.ChannelMessageSend(m.ChannelID, cmd.Response); err != nil {
		s.logger.Error("custom command send failed guild=%s channel=%s trigger=%s err=%v", m.GuildID, m.ChannelID, cmd.Trigger, err)
		return false
	}
	return true
}
