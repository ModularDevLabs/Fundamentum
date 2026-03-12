package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleKeywordAlerts(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.FeatureEnabled(models.FeatureKeywordAlerts) || strings.TrimSpace(settings.KeywordAlertsChannelID) == "" {
		return
	}
	body := strings.ToLower(strings.TrimSpace(m.Content))
	if body == "" {
		return
	}
	matched := ""
	for _, w := range settings.KeywordAlertWords {
		word := strings.ToLower(strings.TrimSpace(w))
		if word == "" {
			continue
		}
		if strings.Contains(body, word) {
			matched = word
			break
		}
	}
	if matched == "" {
		return
	}
	jumpURL := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", m.GuildID, m.ChannelID, m.ID)
	alert := fmt.Sprintf("Keyword alert `%s` by <@%s> in <#%s>\n%s", matched, m.Author.ID, m.ChannelID, jumpURL)
	_, _ = s.session.ChannelMessageSend(settings.KeywordAlertsChannelID, alert)
}
