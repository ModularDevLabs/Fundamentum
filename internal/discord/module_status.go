package discord

import (
	"context"
	"errors"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

type InviteTrackerStatus struct {
	HasManageGuild bool   `json:"has_manage_guild"`
	HasGuild       bool   `json:"has_guild"`
	BotUserID      string `json:"bot_user_id"`
}

type ModulePermissionState struct {
	HasAll              bool     `json:"has_all"`
	RequiredPermissions []string `json:"required_permissions"`
	MissingPermissions  []string `json:"missing_permissions"`
}

type ModulePermissionStatus struct {
	HasGuild  bool                             `json:"has_guild"`
	BotUserID string                           `json:"bot_user_id"`
	Modules   map[string]ModulePermissionState `json:"modules"`
}

type permissionRequirement struct {
	Name string
	Bit  int64
}

var modulePermissionRequirements = map[string][]permissionRequirement{
	models.FeatureWelcomeMessages: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureGoodbyeMessages: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureAuditLogStream: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureInviteTracker: {
		{Name: "Manage Server", Bit: discordgo.PermissionManageServer},
	},
	models.FeatureAutoMod: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Read Message History", Bit: discordgo.PermissionReadMessageHistory},
		{Name: "Manage Messages", Bit: discordgo.PermissionManageMessages},
	},
	models.FeatureReactionRoles: {
		{Name: "Manage Roles", Bit: discordgo.PermissionManageRoles},
	},
	models.FeatureRoleProgression: {
		{Name: "Manage Roles", Bit: discordgo.PermissionManageRoles},
	},
	models.FeatureJoinScreening: {
		{Name: "Kick Members", Bit: discordgo.PermissionKickMembers},
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureWarnings: {
		{Name: "Manage Roles", Bit: discordgo.PermissionManageRoles},
		{Name: "Kick Members", Bit: discordgo.PermissionKickMembers},
	},
	models.FeatureScheduled: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureVerification: {
		{Name: "Manage Roles", Bit: discordgo.PermissionManageRoles},
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureTickets: {
		{Name: "Manage Channels", Bit: discordgo.PermissionManageChannels},
		{Name: "Manage Roles", Bit: discordgo.PermissionManageRoles},
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureAntiRaid: {
		{Name: "Manage Roles", Bit: discordgo.PermissionManageRoles},
		{Name: "Kick Members", Bit: discordgo.PermissionKickMembers},
	},
	models.FeatureAnalytics: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureStarboard: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureLeveling: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureGiveaways: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
		{Name: "Add Reactions", Bit: discordgo.PermissionAddReactions},
	},
	models.FeaturePolls: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
		{Name: "Add Reactions", Bit: discordgo.PermissionAddReactions},
	},
	models.FeatureSuggestions: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
		{Name: "Add Reactions", Bit: discordgo.PermissionAddReactions},
	},
	models.FeatureKeywordAlerts: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureAFK: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureReminders: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureAccountAgeGuard: {
		{Name: "Manage Roles", Bit: discordgo.PermissionManageRoles},
		{Name: "Kick Members", Bit: discordgo.PermissionKickMembers},
	},
	models.FeatureMemberNotes: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureAppeals: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
		{Name: "Manage Messages", Bit: discordgo.PermissionManageMessages},
	},
	models.FeatureCustomCommands: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
	models.FeatureBirthdays: {
		{Name: "View Channel", Bit: discordgo.PermissionViewChannel},
		{Name: "Send Messages", Bit: discordgo.PermissionSendMessages},
	},
}

func (s *Service) GetInviteTrackerStatus(_ context.Context, guildID string) (InviteTrackerStatus, error) {
	if s.session == nil || s.session.State == nil || s.session.State.User == nil {
		return InviteTrackerStatus{}, errors.New("discord session not ready")
	}
	botUserID := s.session.State.User.ID
	if botUserID == "" {
		return InviteTrackerStatus{}, errors.New("bot user id unavailable")
	}
	member, err := s.session.GuildMember(guildID, botUserID)
	if err != nil {
		return InviteTrackerStatus{}, err
	}
	roles, err := s.session.GuildRoles(guildID)
	if err != nil {
		return InviteTrackerStatus{}, err
	}
	perms := memberPermissions(guildID, member, roles)
	hasManageGuild := (perms&discordgo.PermissionManageServer) != 0 || (perms&discordgo.PermissionAdministrator) != 0
	return InviteTrackerStatus{
		HasManageGuild: hasManageGuild,
		HasGuild:       true,
		BotUserID:      botUserID,
	}, nil
}

func (s *Service) GetModulePermissionStatus(_ context.Context, guildID string) (ModulePermissionStatus, error) {
	if s.session == nil || s.session.State == nil || s.session.State.User == nil {
		return ModulePermissionStatus{}, errors.New("discord session not ready")
	}
	botUserID := s.session.State.User.ID
	if botUserID == "" {
		return ModulePermissionStatus{}, errors.New("bot user id unavailable")
	}
	member, err := s.session.GuildMember(guildID, botUserID)
	if err != nil {
		return ModulePermissionStatus{}, err
	}
	roles, err := s.session.GuildRoles(guildID)
	if err != nil {
		return ModulePermissionStatus{}, err
	}
	perms := memberPermissions(guildID, member, roles)
	modules := map[string]ModulePermissionState{}
	for module, reqs := range modulePermissionRequirements {
		required := make([]string, 0, len(reqs))
		missing := make([]string, 0, len(reqs))
		for _, req := range reqs {
			required = append(required, req.Name)
			if (perms & req.Bit) == 0 {
				missing = append(missing, req.Name)
			}
		}
		modules[module] = ModulePermissionState{
			HasAll:              len(missing) == 0,
			RequiredPermissions: required,
			MissingPermissions:  missing,
		}
	}
	return ModulePermissionStatus{
		HasGuild:  true,
		BotUserID: botUserID,
		Modules:   modules,
	}, nil
}

func memberPermissions(guildID string, member *discordgo.Member, roles []*discordgo.Role) int64 {
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
	return perms
}
