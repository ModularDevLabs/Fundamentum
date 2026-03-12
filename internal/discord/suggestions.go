package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleSuggestionMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.FeatureEnabled(models.FeatureSuggestions) || settings.SuggestionsChannelID == "" {
		return
	}
	if m.ChannelID != settings.SuggestionsChannelID {
		return
	}
	content := strings.TrimSpace(m.Content)
	if content == "" {
		return
	}
	embed := fmt.Sprintf("💡 **Suggestion from %s**\n%s", m.Author.Mention(), content)
	msg, err := s.session.ChannelMessageSend(m.ChannelID, embed)
	if err != nil || msg == nil {
		return
	}
	_ = s.session.MessageReactionAdd(m.ChannelID, msg.ID, "👍")
	_ = s.session.MessageReactionAdd(m.ChannelID, msg.ID, "👎")
	_, _ = s.repos.Suggestions.Create(ctx, models.SuggestionRow{
		GuildID:   m.GuildID,
		UserID:    m.Author.ID,
		Content:   content,
		MessageID: msg.ID,
		ChannelID: m.ChannelID,
	})
	_ = s.session.ChannelMessageDelete(m.ChannelID, m.ID)
}
