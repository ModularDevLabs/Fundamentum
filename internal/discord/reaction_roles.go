package discord

import (
	"context"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) OnMessageReactionAdd(_ *discordgo.Session, evt *discordgo.MessageReactionAdd) {
	if evt == nil || evt.GuildID == "" || evt.UserID == "" {
		return
	}
	if s.session.State != nil && s.session.State.User != nil && evt.UserID == s.session.State.User.ID {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.applyReactionRole(evt.GuildID, evt.ChannelID, evt.MessageID, evt.UserID, evt.Emoji, true)
	s.handleStarboardReaction(ctx, evt.GuildID, evt.ChannelID, evt.MessageID, evt.Emoji)
	s.handleGiveawayReaction(ctx, evt.GuildID, evt.ChannelID, evt.MessageID, evt.UserID, reactionEmojiKey(evt.Emoji))
}

func (s *Service) OnMessageReactionRemove(_ *discordgo.Session, evt *discordgo.MessageReactionRemove) {
	if evt == nil || evt.GuildID == "" || evt.UserID == "" {
		return
	}
	if s.session.State != nil && s.session.State.User != nil && evt.UserID == s.session.State.User.ID {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.applyReactionRole(evt.GuildID, evt.ChannelID, evt.MessageID, evt.UserID, evt.Emoji, false)
	s.handleStarboardReaction(ctx, evt.GuildID, evt.ChannelID, evt.MessageID, evt.Emoji)
}

func (s *Service) applyReactionRole(guildID, channelID, messageID, userID string, emoji discordgo.Emoji, isAdd bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := s.repos.Settings.Get(ctx, guildID)
	if err != nil || !settings.FeatureEnabled(models.FeatureReactionRoles) {
		return
	}
	if !settings.FeatureAllowedInChannel(models.FeatureReactionRoles, channelID) {
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
		member, err := s.session.GuildMember(guildID, userID)
		if err != nil || member == nil {
			continue
		}
		roleSet := map[string]struct{}{}
		for _, rid := range member.Roles {
			roleSet[rid] = struct{}{}
		}
		if isAdd {
			if _, has := roleSet[rule.RoleID]; has {
				continue
			}
			if strings.TrimSpace(rule.GroupKey) != "" && rule.MaxSelect > 0 {
				groupRules := make([]models.ReactionRoleRule, 0)
				for _, candidate := range rules {
					if candidate.ChannelID == channelID && candidate.MessageID == messageID && candidate.GroupKey == rule.GroupKey {
						groupRules = append(groupRules, candidate)
					}
				}
				selected := make([]string, 0)
				for _, g := range groupRules {
					if _, has := roleSet[g.RoleID]; has {
						selected = append(selected, g.RoleID)
					}
				}
				for len(selected) >= rule.MaxSelect {
					victim := selected[0]
					selected = selected[1:]
					if err := s.session.GuildMemberRoleRemove(guildID, userID, victim); err == nil {
						delete(roleSet, victim)
					}
				}
			}
			if err := s.session.GuildMemberRoleAdd(guildID, userID, rule.RoleID); err != nil {
				s.logger.Error("reaction role add failed guild=%s user=%s role=%s err=%v", guildID, userID, rule.RoleID, err)
			}
			continue
		}
		if !rule.RemoveOnUnreact {
			continue
		}
		if strings.TrimSpace(rule.GroupKey) != "" && rule.MinSelect > 0 {
			current := 0
			for _, candidate := range rules {
				if candidate.ChannelID == channelID && candidate.MessageID == messageID && candidate.GroupKey == rule.GroupKey {
					if _, has := roleSet[candidate.RoleID]; has {
						current++
					}
				}
			}
			if current <= rule.MinSelect {
				continue
			}
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
