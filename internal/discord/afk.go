package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleAFKMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.FeatureEnabled(models.FeatureAFK) {
		return
	}
	phrase := strings.TrimSpace(settings.AFKSetPhrase)
	if phrase == "" {
		phrase = "!afk"
	}
	content := strings.TrimSpace(m.Content)

	if strings.HasPrefix(strings.ToLower(content), strings.ToLower(phrase)) {
		reason := strings.TrimSpace(content[len(phrase):])
		if reason == "" {
			reason = "AFK"
		}
		_ = s.repos.AFK.Set(ctx, m.GuildID, m.Author.ID, reason)
		_, _ = s.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s AFK set: %s", m.Author.Mention(), reason))
		return
	}

	if _, found, _ := s.repos.AFK.Get(ctx, m.GuildID, m.Author.ID); found {
		_ = s.repos.AFK.Clear(ctx, m.GuildID, m.Author.ID)
		_, _ = s.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s welcome back, AFK cleared.", m.Author.Mention()))
	}

	if len(m.Mentions) == 0 {
		return
	}
	for _, mentioned := range m.Mentions {
		if mentioned == nil || mentioned.ID == "" || mentioned.Bot {
			continue
		}
		afk, found, err := s.repos.AFK.Get(ctx, m.GuildID, mentioned.ID)
		if err != nil || !found {
			continue
		}
		_, _ = s.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s is AFK: %s", mentioned.Mention(), afk.Reason))
	}
}
