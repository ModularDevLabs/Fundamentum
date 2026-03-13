package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) runActionWorker(ctx context.Context) {
	for {
		processedAny := false
		for {
			row, ok, err := s.repos.Actions.NextQueued(ctx)
			if err != nil {
				s.logger.Error("action poll failed: %v", err)
				break
			}
			if !ok {
				break
			}
			settings, settingsErr := s.repos.Settings.Get(ctx, row.GuildID)
			if settingsErr == nil && settings.InMaintenanceWindow(time.Now().UTC()) {
				break
			}
			processedAny = true
			if err := s.repos.Actions.UpdateStatus(ctx, row.ID, "running", ""); err != nil {
				s.logger.Error("action update failed: %v", err)
				continue
			}
			err = s.ExecuteAction(ctx, row)
			if err != nil {
				s.logger.Error("action %d failed type=%s target=%s err=%v", row.ID, row.Type, row.TargetUserID, err)
				_ = s.repos.Actions.UpdateStatus(ctx, row.ID, "failed", err.Error())
				s.emitAuditEvent(row.GuildID, "action_failed", fmt.Sprintf("Action failed: %s target=%s error=%v", row.Type, row.TargetUserID, err))
				continue
			}
			s.logger.Info("action %d success type=%s target=%s", row.ID, row.Type, row.TargetUserID)
			_ = s.repos.Actions.UpdateStatus(ctx, row.ID, "success", "")
			s.emitAuditEvent(row.GuildID, "action_success", fmt.Sprintf("Action success: %s target=%s", row.Type, row.TargetUserID))
		}

		if processedAny {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-s.actionWakeCh:
		}
	}
}

func (s *Service) NotifyActionQueued() {
	select {
	case s.actionWakeCh <- struct{}{}:
	default:
	}
}

func (s *Service) ExecuteAction(ctx context.Context, action models.ActionRow) error {
	settings, err := s.repos.Settings.Get(ctx, action.GuildID)
	if err != nil {
		return err
	}

	var payload struct {
		Reason               string   `json:"reason"`
		RoleIDs              []string `json:"role_ids"`
		RemoveAllExceptAllow bool     `json:"remove_all_except_allowlist"`
	}
	if action.PayloadJSON != "" {
		_ = json.Unmarshal([]byte(action.PayloadJSON), &payload)
	}

	member, err := s.session.GuildMember(action.GuildID, action.TargetUserID)
	if err != nil {
		return fmt.Errorf("load target member: %w", err)
	}

	isAdmin, err := s.memberIsAdmin(action.GuildID, action.TargetUserID)
	if err != nil {
		return err
	}

	if isAdmin {
		switch settings.AdminUserPolicy {
		case "refuse":
			return errors.New("target has admin permission; policy=refuse")
		case "remove_admin_roles":
			if err := s.removeAdminRoles(ctx, action.GuildID, member); err != nil {
				return err
			}
		}
	}

	switch action.Type {
	case "quarantine":
		return s.applyQuarantine(ctx, action.GuildID, member, settings, payload.Reason)
	case "kick":
		if err := s.session.GuildMemberDeleteWithReason(action.GuildID, action.TargetUserID, payload.Reason); err != nil {
			return fmt.Errorf("kick member: %w", err)
		}
		return nil
	case "remove_roles":
		return s.removeRoles(ctx, action.GuildID, member, settings, payload.RoleIDs, payload.RemoveAllExceptAllow)
	default:
		return errors.New("unknown action type")
	}
}

func (s *Service) memberIsAdmin(guildID, userID string) (bool, error) {
	if guild, err := s.session.Guild(guildID); err == nil {
		if guild.OwnerID == userID {
			return true, nil
		}
	}
	member, err := s.session.GuildMember(guildID, userID)
	if err != nil {
		return false, err
	}
	roles, err := s.session.GuildRoles(guildID)
	if err != nil {
		return false, err
	}
	perms := int64(0)
	for _, role := range roles {
		if role.ID == guildID {
			perms |= role.Permissions
		}
	}
	memberRoles := map[string]struct{}{}
	for _, id := range member.Roles {
		memberRoles[id] = struct{}{}
	}
	for _, role := range roles {
		if _, ok := memberRoles[role.ID]; ok {
			perms |= role.Permissions
		}
	}
	return (perms & discordgo.PermissionAdministrator) != 0, nil
}

func (s *Service) removeAdminRoles(ctx context.Context, guildID string, member *discordgo.Member) error {
	roles, err := s.session.GuildRoles(guildID)
	if err != nil {
		return err
	}
	adminRoles := map[string]struct{}{}
	for _, role := range roles {
		if (role.Permissions & discordgo.PermissionAdministrator) != 0 {
			adminRoles[role.ID] = struct{}{}
		}
	}
	for _, roleID := range member.Roles {
		if _, ok := adminRoles[roleID]; ok {
			if err := s.session.GuildMemberRoleRemove(guildID, member.User.ID, roleID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) applyQuarantine(ctx context.Context, guildID string, member *discordgo.Member, settings models.GuildSettings, reason string) error {
	// Per-user actions only need role/channel existence; full overwrite sweeps are guild-level provisioning.
	if err := s.EnsureQuarantineBaseAssets(ctx, guildID, settings); err != nil {
		return err
	}
	updated, err := s.repos.Settings.Get(ctx, guildID)
	if err == nil {
		settings = updated
	}
	if settings.QuarantineRoleID == "" {
		return errors.New("quarantine role not configured")
	}
	if err := s.session.GuildMemberRoleAdd(guildID, member.User.ID, settings.QuarantineRoleID); err != nil {
		return fmt.Errorf("assign quarantine role %s: %w", settings.QuarantineRoleID, err)
	}
	if err := s.removeRoles(ctx, guildID, member, settings, nil, true); err != nil {
		// Quarantine role assignment is the critical step; role pruning is best-effort.
		s.logger.Error("quarantine role-prune skipped target=%s err=%v", member.User.ID, err)
	}
	reason = strings.TrimSpace(reason)
	if settings.ReadmeChannelID != "" && reason != "" {
		_, _ = s.session.ChannelMessageSend(settings.ReadmeChannelID, fmt.Sprintf("<@%s> %s", member.User.ID, reason))
	}
	return nil
}

func (s *Service) removeRoles(ctx context.Context, guildID string, member *discordgo.Member, settings models.GuildSettings, explicit []string, removeAllExceptAllow bool) error {
	allow := map[string]struct{}{}
	for _, id := range settings.AllowlistRoleIDs {
		allow[id] = struct{}{}
	}
	if settings.QuarantineRoleID != "" {
		allow[settings.QuarantineRoleID] = struct{}{}
	}
	if removeAllExceptAllow {
		for _, roleID := range member.Roles {
			if _, ok := allow[roleID]; ok {
				continue
			}
			if err := s.session.GuildMemberRoleRemove(guildID, member.User.ID, roleID); err != nil {
				return fmt.Errorf("remove role %s: %w", roleID, err)
			}
		}
		return nil
	}

	explicitSet := map[string]struct{}{}
	for _, id := range explicit {
		explicitSet[id] = struct{}{}
	}
	for _, roleID := range member.Roles {
		if _, ok := explicitSet[roleID]; ok {
			if err := s.session.GuildMemberRoleRemove(guildID, member.User.ID, roleID); err != nil {
				return fmt.Errorf("remove role %s: %w", roleID, err)
			}
		}
	}
	return nil
}
