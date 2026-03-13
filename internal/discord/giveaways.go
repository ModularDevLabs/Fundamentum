package discord

import (
	"context"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Service) handleGiveawayReaction(ctx context.Context, guildID, channelID, messageID, userID, emoji string) {
	settings, err := s.repos.Settings.Get(ctx, guildID)
	if err != nil || !settings.FeatureEnabled(models.FeatureGiveaways) {
		return
	}
	if !settings.FeatureAllowedInChannel(models.FeatureGiveaways, channelID) {
		return
	}
	targetEmoji := strings.TrimSpace(settings.GiveawaysReactionEmoji)
	if targetEmoji == "" {
		targetEmoji = "🎉"
	}
	if strings.TrimSpace(emoji) != targetEmoji {
		return
	}
	giveaway, found, err := s.repos.Giveaways.FindOpenByMessage(ctx, guildID, channelID, messageID)
	if err != nil || !found {
		return
	}
	if time.Now().UTC().After(giveaway.EndsAt) {
		return
	}
	if err := s.repos.Giveaways.AddEntry(ctx, giveaway.ID, userID); err != nil {
		s.logger.Error("giveaway add entry failed giveaway=%d user=%s err=%v", giveaway.ID, userID, err)
	}
}
