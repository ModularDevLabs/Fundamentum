package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) refreshInviteCache(guildID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := s.repos.Settings.Get(ctx, guildID)
	if err != nil || !settings.FeatureEnabled(models.FeatureInviteTracker) {
		return
	}
	invites, err := s.session.GuildInvites(guildID)
	if err != nil {
		s.logger.Error("invite cache refresh failed guild=%s: %v", guildID, err)
		return
	}
	s.invitesMu.Lock()
	defer s.invitesMu.Unlock()
	s.invitesCache[guildID] = mapInviteUses(invites)
}

func (s *Service) handleInviteTracking(guildID string, user *discordgo.User) {
	if user == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := s.repos.Settings.Get(ctx, guildID)
	if err != nil || !settings.FeatureEnabled(models.FeatureInviteTracker) || settings.InviteLogChannelID == "" {
		return
	}
	current, err := s.session.GuildInvites(guildID)
	if err != nil {
		s.logger.Error("invite tracker fetch failed guild=%s user=%s err=%v", guildID, user.ID, err)
		return
	}

	s.invitesMu.Lock()
	previous := s.invitesCache[guildID]
	currentUses := mapInviteUses(current)
	s.invitesCache[guildID] = currentUses
	s.invitesMu.Unlock()

	code, inviter, uses, ok := findUsedInvite(previous, current)
	msg := fmt.Sprintf("New member: %s (%s) joined.", user.Username, user.ID)
	if ok {
		inviterName := "Unknown"
		inviterID := "unknown"
		if inviter != nil {
			inviterName = inviter.Username
			inviterID = inviter.ID
		}
		msg = fmt.Sprintf("New member: %s (%s) joined via invite `%s` by %s (%s), uses=%d.", user.Username, user.ID, code, inviterName, inviterID, uses)
	} else if len(previous) == 0 {
		msg = fmt.Sprintf("New member: %s (%s) joined (invite tracker warming cache; source unknown).", user.Username, user.ID)
	}

	if _, err := s.session.ChannelMessageSend(settings.InviteLogChannelID, msg); err != nil {
		s.logger.Error("invite tracker send failed guild=%s channel=%s err=%v", guildID, settings.InviteLogChannelID, err)
	}
}

func (s *Service) OnInviteCreate(_ *discordgo.Session, evt *discordgo.InviteCreate) {
	if evt == nil || evt.GuildID == "" {
		return
	}
	s.refreshInviteCache(evt.GuildID)
}

func (s *Service) OnInviteDelete(_ *discordgo.Session, evt *discordgo.InviteDelete) {
	if evt == nil || evt.GuildID == "" {
		return
	}
	s.refreshInviteCache(evt.GuildID)
}

func mapInviteUses(invites []*discordgo.Invite) map[string]int {
	out := map[string]int{}
	for _, inv := range invites {
		if inv == nil || inv.Code == "" {
			continue
		}
		out[inv.Code] = inv.Uses
	}
	return out
}

func findUsedInvite(previous map[string]int, current []*discordgo.Invite) (string, *discordgo.User, int, bool) {
	for _, inv := range current {
		if inv == nil || inv.Code == "" {
			continue
		}
		if inv.Uses > previous[inv.Code] {
			return inv.Code, inv.Inviter, inv.Uses, true
		}
	}
	return "", nil, 0, false
}
