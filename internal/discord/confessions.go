package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleConfessionMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.ConfessionsEnabled || strings.TrimSpace(settings.ConfessionsChannelID) == "" {
		return
	}
	if m.ChannelID != strings.TrimSpace(settings.ConfessionsChannelID) {
		return
	}
	content := strings.TrimSpace(m.Content)
	if content == "" {
		return
	}
	status := "pending"
	if !settings.ConfessionsRequireReview {
		status = "posted"
	}
	id, err := s.repos.Confessions.Create(ctx, m.GuildID, m.Author.ID, content, status)
	if err != nil {
		return
	}
	if settings.ConfessionsRequireReview {
		_, _ = s.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Confession #%d submitted for review.", id))
		return
	}
	msg, err := s.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Anonymous confession #%d:\n%s", id, content))
	if err == nil && msg != nil {
		_ = s.repos.Confessions.UpdateStatus(ctx, id, "posted", msg.ID)
	}
}
