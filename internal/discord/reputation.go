package discord

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var repUserIDPattern = regexp.MustCompile(`\d{5,}`)

func (s *Service) handleReputationCommand(ctx context.Context, m *discordgo.MessageCreate) bool {
	if m == nil || m.Author == nil || m.Author.Bot {
		return false
	}
	body := strings.TrimSpace(m.Content)
	if body == "" {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(body), "+rep") && !strings.HasPrefix(strings.ToLower(body), "-rep") {
		return false
	}
	delta := 1
	if strings.HasPrefix(strings.ToLower(body), "-rep") {
		delta = -1
	}
	targetID := ""
	matches := repUserIDPattern.FindAllString(body, -1)
	for _, candidate := range matches {
		if candidate != m.Author.ID {
			targetID = candidate
			break
		}
	}
	if targetID == "" {
		_, _ = s.session.ChannelMessageSend(m.ChannelID, "Usage: `+rep @user` or `-rep @user`")
		return true
	}
	last, found, err := s.repos.Reputation.LastGivenAt(ctx, m.GuildID, m.Author.ID, targetID)
	if err != nil {
		return true
	}
	if found && time.Since(last) < 12*time.Hour {
		_, _ = s.session.ChannelMessageSend(m.ChannelID, "Reputation cooldown is active for this member pair (12h).")
		return true
	}
	if err := s.repos.Reputation.AddDelta(ctx, m.GuildID, m.Author.ID, targetID, delta); err != nil {
		return true
	}
	sign := "+"
	if delta < 0 {
		sign = "-"
	}
	_, _ = s.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Recorded %s1 rep for <@%s>.", sign, targetID))
	return true
}
