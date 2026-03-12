package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleAccountAgeGuardOnJoin(ctx context.Context, evt *discordgo.GuildMemberAdd, settings models.GuildSettings) {
	if evt == nil || evt.User == nil || evt.GuildID == "" {
		return
	}
	if !settings.FeatureEnabled(models.FeatureAccountAgeGuard) {
		return
	}
	minDays := settings.AccountAgeMinDays
	if minDays <= 0 {
		minDays = 7
	}
	createdAt, err := discordgo.SnowflakeTimestamp(evt.User.ID)
	if err != nil {
		return
	}
	age := time.Since(createdAt)
	if age >= time.Duration(minDays)*24*time.Hour {
		return
	}

	reason := fmt.Sprintf("Account age %.1f days is below minimum %d days", age.Hours()/24, minDays)
	logMsg := fmt.Sprintf("Account-age guard triggered for %s (%s): %s", evt.User.Username, evt.User.ID, reason)
	if strings.TrimSpace(settings.AccountAgeLogChannelID) != "" {
		_, _ = s.session.ChannelMessageSend(settings.AccountAgeLogChannelID, logMsg)
	}

	action := settings.AccountAgeAction
	switch action {
	case "kick", "quarantine":
		payload, _ := json.Marshal(map[string]any{
			"reason":      reason,
			"target_name": evt.User.Username,
		})
		_, err := s.repos.Actions.Enqueue(ctx, models.ActionRow{
			GuildID:      evt.GuildID,
			ActorUserID:  "account_age_guard",
			TargetUserID: evt.User.ID,
			Type:         action,
			PayloadJSON:  string(payload),
		})
		if err == nil {
			s.NotifyActionQueued()
		}
	default:
	}
}
