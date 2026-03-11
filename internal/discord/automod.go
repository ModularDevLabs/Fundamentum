package discord

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleAutoMod(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.FeatureEnabled(models.FeatureAutoMod) {
		return
	}
	if stringInSlice(m.ChannelID, settings.AutoModIgnoreChannelIDs) {
		return
	}
	for _, roleID := range settings.AutoModIgnoreRoleIDs {
		if memberHasRole(m.Member, roleID) {
			return
		}
	}
	reasons := s.automodReasons(m, settings)
	if len(reasons) == 0 {
		return
	}

	_ = s.session.ChannelMessageDelete(m.ChannelID, m.ID)
	joined := strings.Join(reasons, ", ")
	s.emitAuditEvent(m.GuildID, "automod_action", "Message removed for "+joined+" user="+m.Author.ID)

	switch settings.AutoModAction {
	case "delete_only":
		return
	case "delete_quarantine":
		payload, _ := json.Marshal(map[string]any{"reason": "AutoMod: " + joined})
		row := models.ActionRow{
			GuildID:      m.GuildID,
			ActorUserID:  "automod",
			TargetUserID: m.Author.ID,
			Type:         "quarantine",
			PayloadJSON:  string(payload),
		}
		if _, err := s.repos.Actions.Enqueue(ctx, row); err == nil {
			s.NotifyActionQueued()
		}
	default:
		_, _ = s.session.ChannelMessageSend(m.ChannelID, "AutoMod removed a message from "+m.Author.Mention()+" ("+joined+").")
	}
}

func (s *Service) automodReasons(m *discordgo.MessageCreate, settings models.GuildSettings) []string {
	reasons := make([]string, 0, 3)
	content := strings.ToLower(strings.TrimSpace(m.Content))
	if content == "" {
		return reasons
	}
	if settings.AutoModBlockLinks && containsLink(content) {
		reasons = append(reasons, "link")
	}
	for _, w := range settings.AutoModBlockedWords {
		word := strings.ToLower(strings.TrimSpace(w))
		if word == "" {
			continue
		}
		if strings.Contains(content, word) {
			reasons = append(reasons, "blocked_word")
			break
		}
	}
	if s.isDuplicateSpam(m.GuildID, m.Author.ID, content, settings.AutoModDupWindowSec, settings.AutoModDupThreshold) {
		reasons = append(reasons, "duplicate_spam")
	}
	return reasons
}

func containsLink(content string) bool {
	return strings.Contains(content, "http://") || strings.Contains(content, "https://") || strings.Contains(content, "discord.gg/")
}

func (s *Service) isDuplicateSpam(guildID, userID, content string, windowSec, threshold int) bool {
	if windowSec <= 0 || threshold <= 1 {
		return false
	}
	key := guildID + ":" + userID + ":" + content
	cutoff := time.Now().Add(-time.Duration(windowSec) * time.Second)

	s.automodMu.Lock()
	defer s.automodMu.Unlock()

	existing := s.automodSeen[key]
	kept := make([]time.Time, 0, len(existing)+1)
	for _, ts := range existing {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	kept = append(kept, time.Now())
	s.automodSeen[key] = kept
	return len(kept) >= threshold
}

func stringInSlice(value string, items []string) bool {
	for _, it := range items {
		if strings.TrimSpace(it) == value {
			return true
		}
	}
	return false
}

func memberHasRole(member *discordgo.Member, roleID string) bool {
	if member == nil || roleID == "" {
		return false
	}
	for _, r := range member.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}
