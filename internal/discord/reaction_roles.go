package discord

import (
	"context"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) OnMessageReactionAdd(_ *discordgo.Session, evt *discordgo.MessageReactionAdd) {
	if evt == nil || evt.GuildID == "" || evt.UserID == "" {
		return
	}
	if s.session.State != nil && s.session.State.User != nil && evt.UserID == s.session.State.User.ID {
		return
	}
	s.applyReactionRole(evt.GuildID, evt.ChannelID, evt.MessageID, evt.UserID, evt.Emoji, true)
}

func (s *Service) OnMessageReactionRemove(_ *discordgo.Session, evt *discordgo.MessageReactionRemove) {
	if evt == nil || evt.GuildID == "" || evt.UserID == "" {
		return
	}
	if s.session.State != nil && s.session.State.User != nil && evt.UserID == s.session.State.User.ID {
		return
	}
	s.applyReactionRole(evt.GuildID, evt.ChannelID, evt.MessageID, evt.UserID, evt.Emoji, false)
}

func (s *Service) applyReactionRole(guildID, channelID, messageID, userID string, emoji discordgo.Emoji, isAdd bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := s.repos.Settings.Get(ctx, guildID)
	if err != nil || !settings.FeatureEnabled(models.FeatureReactionRoles) {
		return
	}
	rules, err := s.repos.ReactionRoles.ListByGuild(ctx, guildID)
	if err != nil {
		s.logger.Error("reaction rules list failed guild=%s: %v", guildID, err)
		return
	}
	target := reactionEmojiKey(emoji)
	for _, rule := range rules {
		if rule.ChannelID != channelID || rule.MessageID != messageID {
			continue
		}
		if rule.Emoji != target {
			continue
		}
		if isAdd {
			if err := s.session.GuildMemberRoleAdd(guildID, userID, rule.RoleID); err != nil {
				s.logger.Error("reaction role add failed guild=%s user=%s role=%s err=%v", guildID, userID, rule.RoleID, err)
			}
			continue
		}
		if !rule.RemoveOnUnreact {
			continue
		}
		if err := s.session.GuildMemberRoleRemove(guildID, userID, rule.RoleID); err != nil {
			s.logger.Error("reaction role remove failed guild=%s user=%s role=%s err=%v", guildID, userID, rule.RoleID, err)
		}
	}
}

func reactionEmojiKey(emoji discordgo.Emoji) string {
	if emoji.ID != "" {
		return emoji.ID
	}
	return emoji.Name
}
