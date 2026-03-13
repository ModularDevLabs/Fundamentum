package discord

import (
	"context"
	"errors"
	"fmt"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) EnsureQuarantineAssets(ctx context.Context, guildID string, settings models.GuildSettings) error {
	_, _, err := s.ensureQuarantineBaseAssets(ctx, guildID, settings)
	if err != nil {
		return err
	}
	if settings.SafeQuarantineMode {
		// Safe mode: do not change overwrites across the guild.
		return nil
	}
	if err := s.applyQuarantineOverwrites(guildID, settings.QuarantineRoleID, settings.ReadmeChannelID); err != nil {
		return fmt.Errorf("apply quarantine overwrites: %w", err)
	}
	return nil
}

func (s *Service) EnsureQuarantineBaseAssets(ctx context.Context, guildID string, settings models.GuildSettings) error {
	_, _, err := s.ensureQuarantineBaseAssets(ctx, guildID, settings)
	return err
}

func (s *Service) ensureQuarantineBaseAssets(ctx context.Context, guildID string, settings models.GuildSettings) (string, string, error) {
	roleID, err := s.ensureQuarantineRole(guildID, settings)
	if err != nil {
		return "", "", fmt.Errorf("ensure quarantine role: %w", err)
	}
	if settings.QuarantineRoleID == "" && roleID != "" {
		settings.QuarantineRoleID = roleID
		_ = s.repos.Settings.Upsert(ctx, settings)
	}

	channelID, err := s.ensureReadmeChannel(guildID, settings)
	if err != nil {
		return "", "", fmt.Errorf("ensure readme channel: %w", err)
	}
	if settings.ReadmeChannelID == "" && channelID != "" {
		settings.ReadmeChannelID = channelID
		_ = s.repos.Settings.Upsert(ctx, settings)
	}

	if settings.QuarantineRoleID == "" || settings.ReadmeChannelID == "" {
		return "", "", errors.New("missing quarantine role or readme channel")
	}

	if err := s.ensureReadmeOverwrite(settings.ReadmeChannelID, settings.QuarantineRoleID); err != nil {
		// Readme overwrite is helpful but should not block quarantine assignment.
		s.logger.Error("readme overwrite skipped channel=%s role=%s err=%v", settings.ReadmeChannelID, settings.QuarantineRoleID, err)
	}
	return settings.QuarantineRoleID, settings.ReadmeChannelID, nil
}

func (s *Service) ensureQuarantineRole(guildID string, settings models.GuildSettings) (string, error) {
	if settings.QuarantineRoleID != "" {
		return settings.QuarantineRoleID, nil
	}
	roles, err := s.session.GuildRoles(guildID)
	if err != nil {
		return "", err
	}
	for _, role := range roles {
		if role.Name == "Quarantined" {
			return role.ID, nil
		}
	}
	role, err := s.session.GuildRoleCreate(guildID, &discordgo.RoleParams{
		Name: "Quarantined",
	})
	if err != nil {
		return "", err
	}
	if role != nil {
		return role.ID, nil
	}
	return "", nil
}

func (s *Service) ensureReadmeChannel(guildID string, settings models.GuildSettings) (string, error) {
	if settings.ReadmeChannelID != "" {
		return settings.ReadmeChannelID, nil
	}
	channels, err := s.session.GuildChannels(guildID)
	if err != nil {
		return "", err
	}
	for _, ch := range channels {
		if ch.Name == "quarantine-readme" {
			return ch.ID, nil
		}
	}
	channel, err := s.session.GuildChannelCreate(guildID, "quarantine-readme", discordgo.ChannelTypeGuildText)
	if err != nil {
		return "", err
	}
	return channel.ID, nil
}

func (s *Service) ensureReadmeOverwrite(channelID, roleID string) error {
	allow := int64(discordgo.PermissionViewChannel | discordgo.PermissionReadMessageHistory)
	deny := int64(discordgo.PermissionSendMessages)
	return s.session.ChannelPermissionSet(channelID, roleID, discordgo.PermissionOverwriteTypeRole, allow, deny)
}

func (s *Service) applyQuarantineOverwrites(guildID, roleID, readmeChannelID string) error {
	channels, err := s.session.GuildChannels(guildID)
	if err != nil {
		return err
	}
	deny := int64(discordgo.PermissionViewChannel)
	failed := 0
	for _, ch := range channels {
		if ch.ID == readmeChannelID {
			continue
		}
		if !shouldDenyViewInChannelType(ch.Type) {
			continue
		}
		if err := s.session.ChannelPermissionSet(ch.ID, roleID, discordgo.PermissionOverwriteTypeRole, 0, deny); err != nil {
			failed++
			s.logger.Error("quarantine overwrite skipped channel=%s (%s): %v", ch.Name, ch.ID, err)
			continue
		}
	}
	if failed > 0 {
		s.logger.Error("quarantine overwrite completed with %d skipped channels due to access errors", failed)
	}
	return nil
}

func shouldDenyViewInChannelType(t discordgo.ChannelType) bool {
	switch t {
	case discordgo.ChannelTypeGuildCategory:
		return true
	case discordgo.ChannelTypeGuildText:
		return true
	case discordgo.ChannelTypeGuildNews:
		return true
	case discordgo.ChannelTypeGuildForum:
		return true
	case discordgo.ChannelTypeGuildVoice:
		return true
	case discordgo.ChannelTypeGuildStageVoice:
		return true
	default:
		return false
	}
}
