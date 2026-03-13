package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleAppealMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.FeatureEnabled(models.FeatureAppeals) || settings.AppealsChannelID == "" {
		return
	}
	if m.ChannelID != settings.AppealsChannelID {
		return
	}
	phrase := strings.TrimSpace(settings.AppealsOpenPhrase)
	if phrase == "" {
		phrase = "!appeal"
	}
	content := strings.TrimSpace(m.Content)
	if !strings.HasPrefix(strings.ToLower(content), strings.ToLower(phrase)) {
		return
	}
	reason := strings.TrimSpace(content[len(phrase):])
	if reason == "" {
		reason = "no reason provided"
	}
	id, err := s.repos.Appeals.Create(ctx, models.AppealRow{
		GuildID:   m.GuildID,
		UserID:    m.Author.ID,
		Reason:    reason,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		s.logger.Error("appeal create failed guild=%s user=%s err=%v", m.GuildID, m.Author.ID, err)
		return
	}
	_ = s.session.ChannelMessageDelete(m.ChannelID, m.ID)
	_, _ = s.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> appeal submitted (id=%d).", m.Author.ID, id))
	if settings.AppealsLogChannelID != "" {
		_, _ = s.session.ChannelMessageSend(settings.AppealsLogChannelID, fmt.Sprintf("Appeal #%d submitted by <@%s>: %s", id, m.Author.ID, reason))
	}
}
