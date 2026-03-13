package discord

import (
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleAutoThreadHelper(m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.AutoThreadEnabled || strings.TrimSpace(settings.AutoThreadChannelID) == "" {
		return
	}
	if m.ChannelID != strings.TrimSpace(settings.AutoThreadChannelID) {
		return
	}
	body := strings.ToLower(strings.TrimSpace(m.Content))
	if body == "" {
		return
	}
	matched := false
	for _, kw := range settings.AutoThreadKeywords {
		token := strings.ToLower(strings.TrimSpace(kw))
		if token == "" {
			continue
		}
		if strings.Contains(body, token) {
			matched = true
			break
		}
	}
	if !matched {
		return
	}
	threadName := "help-" + m.Author.Username
	if len(threadName) > 80 {
		threadName = threadName[:80]
	}
	_, err := s.session.MessageThreadStartComplex(m.ChannelID, m.ID, &discordgo.ThreadStart{
		Name:                threadName,
		AutoArchiveDuration: 1440,
		Type:                discordgo.ChannelTypeGuildPublicThread,
	})
	if err != nil {
		s.logger.Error("auto-thread create failed guild=%s channel=%s message=%s err=%v", m.GuildID, m.ChannelID, m.ID, err)
	}
}
