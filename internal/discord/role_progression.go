package discord

import (
	"context"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type RoleProgressionSyncResult struct {
	Evaluated int `json:"evaluated"`
	Added     int `json:"added"`
	Removed   int `json:"removed"`
}

func (s *Service) SyncRoleProgressionForUser(ctx context.Context, guildID, userID string) (RoleProgressionSyncResult, error) {
	return s.syncRoleProgressionForUser(ctx, guildID, userID, true)
}

func (s *Service) syncRoleProgressionForUser(ctx context.Context, guildID, userID string, force bool) (RoleProgressionSyncResult, error) {
	if guildID == "" || userID == "" {
		return RoleProgressionSyncResult{}, nil
	}
	settings, err := s.repos.Settings.Get(ctx, guildID)
	if err != nil {
		return RoleProgressionSyncResult{}, err
	}
	if !settings.AutoRoleProgressionEnabled || !settings.FeatureEnabled(models.FeatureRoleProgression) {
		return RoleProgressionSyncResult{}, nil
	}
	if !force && !s.shouldRunRoleProgression(guildID, userID) {
		return RoleProgressionSyncResult{}, nil
	}

	rules, err := s.repos.RoleProgression.ListByGuild(ctx, guildID)
	if err != nil {
		return RoleProgressionSyncResult{}, err
	}
	if len(rules) == 0 {
		return RoleProgressionSyncResult{}, nil
	}

	levelRow, _, _ := s.repos.Leveling.GetMember(ctx, guildID, userID)
	repTotal, _ := s.repos.Reputation.TotalForUser(ctx, guildID, userID)
	balance, _ := s.repos.Economy.GetBalance(ctx, guildID, userID)

	member, err := s.session.GuildMember(guildID, userID)
	if err != nil || member == nil {
		return RoleProgressionSyncResult{}, err
	}
	currentRoles := map[string]struct{}{}
	for _, rid := range member.Roles {
		currentRoles[rid] = struct{}{}
	}

	result := RoleProgressionSyncResult{}
	for _, rule := range rules {
		if !rule.Enabled || strings.TrimSpace(rule.RoleID) == "" {
			continue
		}
		result.Evaluated++
		value := 0
		switch rule.Metric {
		case "level":
			value = levelRow.Level
		case "reputation":
			value = repTotal
		case "economy":
			value = balance
		default:
			continue
		}
		_, hasRole := currentRoles[rule.RoleID]
		shouldHave := value >= rule.Threshold
		if shouldHave && !hasRole {
			if err := s.session.GuildMemberRoleAdd(guildID, userID, rule.RoleID); err == nil {
				result.Added++
				currentRoles[rule.RoleID] = struct{}{}
			}
			continue
		}
		if !shouldHave && hasRole {
			if err := s.session.GuildMemberRoleRemove(guildID, userID, rule.RoleID); err == nil {
				result.Removed++
				delete(currentRoles, rule.RoleID)
			}
		}
	}
	return result, nil
}

func (s *Service) shouldRunRoleProgression(guildID, userID string) bool {
	key := guildID + ":" + userID
	now := time.Now().UTC()
	s.progressionMu.Lock()
	defer s.progressionMu.Unlock()
	last, ok := s.progressionLast[key]
	if ok && now.Sub(last) < 2*time.Minute {
		return false
	}
	s.progressionLast[key] = now
	return true
}
