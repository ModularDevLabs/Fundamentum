package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleJoinScreeningOnJoin(ctx context.Context, m *discordgo.GuildMemberAdd, settings models.GuildSettings) {
	if m == nil || m.User == nil || m.GuildID == "" {
		return
	}
	if !settings.JoinScreeningEnabled || !settings.FeatureEnabled(models.FeatureJoinScreening) {
		return
	}

	reasons := make([]string, 0, 2)
	ageDays := settings.JoinScreeningAccountAgeDays
	if ageDays <= 0 {
		ageDays = 7
	}
	createdAt, err := discordgo.SnowflakeTimestamp(m.User.ID)
	if err != nil {
		createdAt = time.Now().UTC()
	}
	if time.Since(createdAt) < time.Duration(ageDays)*24*time.Hour {
		reasons = append(reasons, fmt.Sprintf("account younger than %d days", ageDays))
	}
	if settings.JoinScreeningRequireAvatar && strings.TrimSpace(m.User.Avatar) == "" {
		reasons = append(reasons, "no avatar")
	}
	if len(reasons) == 0 {
		return
	}

	username := m.User.Username
	if username == "" {
		username = m.User.ID
	}
	reason := strings.Join(reasons, "; ")
	_, _ = s.repos.JoinScreening.Create(ctx, models.JoinScreeningRow{
		GuildID:          m.GuildID,
		UserID:           m.User.ID,
		Username:         username,
		AccountCreatedAt: &createdAt,
		Reason:           reason,
		Status:           "pending",
	})
	if settings.JoinScreeningLogChannelID != "" {
		_, _ = s.session.ChannelMessageSend(settings.JoinScreeningLogChannelID,
			fmt.Sprintf("🔎 Join screening pending: <@%s> (%s). Reason: %s", m.User.ID, username, reason))
	}
}
