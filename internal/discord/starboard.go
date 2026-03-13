package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleStarboardReaction(ctx context.Context, evtGuildID, evtChannelID, evtMessageID string, emoji discordgo.Emoji) {
	settings, err := s.repos.Settings.Get(ctx, evtGuildID)
	if err != nil || !settings.FeatureEnabled(models.FeatureStarboard) {
		return
	}
	if !settings.FeatureAllowedInChannel(models.FeatureStarboard, evtChannelID) {
		return
	}
	if settings.StarboardChannelID == "" {
		return
	}
	if evtChannelID == settings.StarboardChannelID {
		return
	}
	threshold := settings.StarboardThreshold
	if threshold <= 0 {
		threshold = 3
	}
	targetEmoji := strings.TrimSpace(settings.StarboardEmoji)
	if targetEmoji == "" {
		targetEmoji = "⭐"
	}
	if reactionEmojiKey(emoji) != targetEmoji {
		return
	}

	msg, err := s.session.ChannelMessage(evtChannelID, evtMessageID)
	if err != nil || msg == nil {
		return
	}
	if msg.Author != nil && msg.Author.Bot {
		return
	}

	count := 0
	for _, react := range msg.Reactions {
		if react == nil {
			continue
		}
		if react.Emoji == nil || reactionEmojiKey(*react.Emoji) != targetEmoji {
			continue
		}
		count = react.Count
		break
	}

	entry, exists, err := s.repos.Starboard.GetBySource(ctx, evtGuildID, evtChannelID, evtMessageID)
	if err != nil {
		s.logger.Error("starboard lookup failed guild=%s channel=%s message=%s err=%v", evtGuildID, evtChannelID, evtMessageID, err)
		return
	}
	if count < threshold && !exists {
		return
	}

	content := buildStarboardContent(evtGuildID, evtChannelID, evtMessageID, msg, targetEmoji, count)
	if !exists && count >= threshold {
		posted, err := s.session.ChannelMessageSend(settings.StarboardChannelID, content)
		if err != nil || posted == nil {
			s.logger.Error("starboard post failed guild=%s channel=%s message=%s err=%v", evtGuildID, evtChannelID, evtMessageID, err)
			return
		}
		now := time.Now().UTC()
		_ = s.repos.Starboard.Upsert(ctx, models.StarboardEntryRow{
			GuildID:          evtGuildID,
			SourceChannelID:  evtChannelID,
			SourceMessageID:  evtMessageID,
			StarboardChannel: settings.StarboardChannelID,
			StarboardMessage: posted.ID,
			StarCount:        count,
			PostedAt:         &now,
		})
		return
	}
	if !exists {
		return
	}

	if entry.StarboardChannel == "" || entry.StarboardMessage == "" {
		entry.StarboardChannel = settings.StarboardChannelID
	}
	if entry.StarboardChannel == settings.StarboardChannelID && entry.StarboardMessage != "" {
		_, _ = s.session.ChannelMessageEdit(entry.StarboardChannel, entry.StarboardMessage, content)
	}
	entry.StarCount = count
	now := time.Now().UTC()
	entry.PostedAt = &now
	_ = s.repos.Starboard.Upsert(ctx, entry)
}

func buildStarboardContent(guildID, channelID, messageID string, msg *discordgo.Message, emoji string, count int) string {
	author := "unknown"
	if msg != nil && msg.Author != nil && msg.Author.Username != "" {
		author = msg.Author.Username
	}
	body := ""
	if msg != nil {
		body = strings.TrimSpace(msg.Content)
	}
	if body == "" {
		body = "(no text content)"
	}
	if len(body) > 220 {
		body = body[:220] + "..."
	}
	jumpURL := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, messageID)
	return fmt.Sprintf("%s **%d** in <#%s> by **%s**\n%s\n%s", emoji, count, channelID, author, body, jumpURL)
}
