package models

import "time"

type GuildSettings struct {
	GuildID                 string          `json:"guild_id"`
	InactiveDays            int             `json:"inactive_days"`
	BackfillDays            int             `json:"backfill_days"`
	QuarantineRoleID        string          `json:"quarantine_role_id"`
	ReadmeChannelID         string          `json:"readme_channel_id"`
	AllowlistRoleIDs        []string        `json:"allowlist_role_ids"`
	AdminUserPolicy         string          `json:"admin_user_policy"`
	BackfillConcurrency     int             `json:"backfill_concurrency"`
	BackfillIncludeTypes    []string        `json:"backfill_include_types"`
	SafeQuarantineMode      bool            `json:"safe_quarantine_mode"`
	FeatureFlags            map[string]bool `json:"feature_flags"`
	WelcomeChannelID        string          `json:"welcome_channel_id"`
	WelcomeMessage          string          `json:"welcome_message"`
	GoodbyeChannelID        string          `json:"goodbye_channel_id"`
	GoodbyeMessage          string          `json:"goodbye_message"`
	AuditLogChannelID       string          `json:"audit_log_channel_id"`
	AuditLogEventTypes      []string        `json:"audit_log_event_types"`
	InviteLogChannelID      string          `json:"invite_log_channel_id"`
	AutoModBlockLinks       bool            `json:"automod_block_links"`
	AutoModBlockedWords     []string        `json:"automod_blocked_words"`
	AutoModDupWindowSec     int             `json:"automod_dup_window_sec"`
	AutoModDupThreshold     int             `json:"automod_dup_threshold"`
	AutoModAction           string          `json:"automod_action"`
	AutoModIgnoreChannelIDs []string        `json:"automod_ignore_channel_ids"`
	AutoModIgnoreRoleIDs    []string        `json:"automod_ignore_role_ids"`
}

type MemberRow struct {
	GuildID       string     `json:"guild_id"`
	UserID        string     `json:"user_id"`
	Username      string     `json:"username"`
	GlobalName    string     `json:"global_name"`
	DisplayName   string     `json:"display_name"`
	LastMessageAt *time.Time `json:"last_message_at"`
	LastChannelID string     `json:"last_channel_id"`
	Status        string     `json:"status"`
	Quarantined   bool       `json:"quarantined"`
}

type ActionRow struct {
	ID           int64     `json:"id"`
	GuildID      string    `json:"guild_id"`
	ActorUserID  string    `json:"actor_user_id"`
	TargetUserID string    `json:"target_user_id"`
	TargetName   string    `json:"target_name"`
	Type         string    `json:"type"`
	PayloadJSON  string    `json:"payload_json"`
	Status       string    `json:"status"`
	Error        string    `json:"error"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GuildInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const (
	FeatureWelcomeMessages = "welcome_messages"
	FeatureGoodbyeMessages = "goodbye_messages"
	FeatureAuditLogStream  = "audit_log_stream"
	FeatureInviteTracker   = "invite_tracker"
	FeatureAutoMod         = "automod"
)

func (s GuildSettings) FeatureEnabled(flag string) bool {
	if s.FeatureFlags == nil {
		return false
	}
	return s.FeatureFlags[flag]
}

func DefaultGuildSettings(guildID string) GuildSettings {
	return GuildSettings{
		GuildID:             guildID,
		InactiveDays:        180,
		BackfillDays:        60,
		AdminUserPolicy:     "refuse",
		BackfillConcurrency: 2,
		SafeQuarantineMode:  false,
		FeatureFlags: map[string]bool{
			FeatureWelcomeMessages: false,
			FeatureGoodbyeMessages: false,
			FeatureAuditLogStream:  false,
			FeatureInviteTracker:   false,
			FeatureAutoMod:         false,
		},
		WelcomeMessage: "Welcome {user} to {server}.",
		GoodbyeMessage: "Goodbye {user}.",
		AuditLogEventTypes: []string{
			"ban_add",
			"ban_remove",
			"role_create",
			"role_update",
			"role_delete",
			"channel_create",
			"channel_update",
			"channel_delete",
			"action_success",
			"action_failed",
			"automod_action",
		},
		AutoModBlockLinks:   true,
		AutoModBlockedWords: []string{},
		AutoModDupWindowSec: 20,
		AutoModDupThreshold: 3,
		AutoModAction:       "delete_warn",
	}
}
