package discord

import (
	"context"
	"encoding/json"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
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
	ruleReasons, forcedAction := s.automodRuleReasons(m, settings)
	reasons = append(reasons, ruleReasons...)
	if len(reasons) == 0 {
		return
	}

	_ = s.session.ChannelMessageDelete(m.ChannelID, m.ID)
	joined := strings.Join(reasons, ", ")
	s.emitAuditEvent(m.GuildID, "automod_action", "Message removed for "+joined+" user="+m.Author.ID)

	action := settings.AutoModAction
	if forcedAction != "" {
		action = forcedAction
	}
	switch action {
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

func (s *Service) automodRuleReasons(m *discordgo.MessageCreate, settings models.GuildSettings) ([]string, string) {
	if len(settings.AutoModRules) == 0 || m == nil {
		return nil, ""
	}
	reasons := make([]string, 0, len(settings.AutoModRules))
	forcedAction := ""
	for _, rule := range settings.AutoModRules {
		if !rule.Enabled {
			continue
		}
		matched := false
		switch strings.ToLower(strings.TrimSpace(rule.Type)) {
		case "regex":
			pattern := strings.TrimSpace(rule.Pattern)
			if pattern == "" {
				continue
			}
			if re, err := regexp.Compile(pattern); err == nil && re.MatchString(m.Content) {
				matched = true
			}
		case "file_ext":
			extRaw := strings.TrimSpace(rule.Pattern)
			if extRaw == "" {
				continue
			}
			exts := strings.Split(strings.ToLower(extRaw), ",")
			extSet := map[string]struct{}{}
			for _, ext := range exts {
				norm := strings.TrimSpace(ext)
				if norm == "" {
					continue
				}
				if !strings.HasPrefix(norm, ".") {
					norm = "." + norm
				}
				extSet[norm] = struct{}{}
			}
			for _, at := range m.Attachments {
				if at == nil || at.Filename == "" {
					continue
				}
				ext := strings.ToLower(filepath.Ext(at.Filename))
				if _, ok := extSet[ext]; ok {
					matched = true
					break
				}
			}
		case "mention_spam":
			threshold := rule.Threshold
			if threshold <= 0 {
				threshold = 5
			}
			if len(m.Mentions) >= threshold {
				matched = true
			}
		case "caps_ratio":
			threshold := rule.Threshold
			if threshold <= 0 {
				if parsed, err := strconv.Atoi(strings.TrimSpace(rule.Pattern)); err == nil {
					threshold = parsed
				}
			}
			if threshold <= 0 {
				threshold = 70
			}
			letters := 0
			uppers := 0
			for _, ch := range m.Content {
				if unicode.IsLetter(ch) {
					letters++
					if unicode.IsUpper(ch) {
						uppers++
					}
				}
			}
			if letters >= 8 {
				ratio := (uppers * 100) / letters
				if ratio >= threshold {
					matched = true
				}
			}
		default:
			continue
		}
		if !matched {
			continue
		}
		reasonName := strings.TrimSpace(rule.Name)
		if reasonName == "" {
			reasonName = "advanced_rule"
		}
		reasons = append(reasons, reasonName)
		action := strings.TrimSpace(strings.ToLower(rule.Action))
		switch action {
		case "delete_quarantine":
			forcedAction = "delete_quarantine"
		case "delete_only":
			if forcedAction != "delete_quarantine" {
				forcedAction = "delete_only"
			}
		case "delete_warn":
			if forcedAction == "" {
				forcedAction = "delete_warn"
			}
		}
	}
	return reasons, forcedAction
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
