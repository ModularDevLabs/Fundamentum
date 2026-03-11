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
	WarningLogChannelID     string          `json:"warning_log_channel_id"`
	WarnQuarantineThreshold int             `json:"warn_quarantine_threshold"`
	WarnKickThreshold       int             `json:"warn_kick_threshold"`
	VerificationChannelID   string          `json:"verification_channel_id"`
	VerificationPhrase      string          `json:"verification_phrase"`
	UnverifiedRoleID        string          `json:"unverified_role_id"`
	VerifiedRoleID          string          `json:"verified_role_id"`
	TicketInboxChannelID    string          `json:"ticket_inbox_channel_id"`
	TicketCategoryID        string          `json:"ticket_category_id"`
	TicketSupportRoleID     string          `json:"ticket_support_role_id"`
	TicketLogChannelID      string          `json:"ticket_log_channel_id"`
	TicketOpenPhrase        string          `json:"ticket_open_phrase"`
	TicketClosePhrase       string          `json:"ticket_close_phrase"`
	TicketAutoCloseMinutes  int             `json:"ticket_auto_close_minutes"`
	AntiRaidJoinThreshold   int             `json:"anti_raid_join_threshold"`
	AntiRaidWindowSeconds   int             `json:"anti_raid_window_seconds"`
	AntiRaidCooldownMinutes int             `json:"anti_raid_cooldown_minutes"`
	AntiRaidAction          string          `json:"anti_raid_action"`
	AntiRaidAlertChannelID  string          `json:"anti_raid_alert_channel_id"`
	AnalyticsChannelID      string          `json:"analytics_channel_id"`
	AnalyticsIntervalDays   int             `json:"analytics_interval_days"`
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

type ReactionRoleRule struct {
	ID              int64     `json:"id"`
	GuildID         string    `json:"guild_id"`
	ChannelID       string    `json:"channel_id"`
	MessageID       string    `json:"message_id"`
	Emoji           string    `json:"emoji"`
	RoleID          string    `json:"role_id"`
	RemoveOnUnreact bool      `json:"remove_on_unreact"`
	CreatedAt       time.Time `json:"created_at"`
}

type WarningRow struct {
	ID          int64     `json:"id"`
	GuildID     string    `json:"guild_id"`
	UserID      string    `json:"user_id"`
	ActorUserID string    `json:"actor_user_id"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

type ScheduledMessageRow struct {
	ID              int64     `json:"id"`
	GuildID         string    `json:"guild_id"`
	ChannelID       string    `json:"channel_id"`
	Content         string    `json:"content"`
	IntervalMinutes int       `json:"interval_minutes"`
	NextRunAt       time.Time `json:"next_run_at"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TicketRow struct {
	ID            int64      `json:"id"`
	GuildID       string     `json:"guild_id"`
	ChannelID     string     `json:"channel_id"`
	CreatorUserID string     `json:"creator_user_id"`
	Subject       string     `json:"subject"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	ClosedAt      *time.Time `json:"closed_at,omitempty"`
}

type TicketMessageRow struct {
	ID           int64     `json:"id"`
	TicketID     int64     `json:"ticket_id"`
	GuildID      string    `json:"guild_id"`
	ChannelID    string    `json:"channel_id"`
	AuthorUserID string    `json:"author_user_id"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
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
	FeatureReactionRoles   = "reaction_roles"
	FeatureWarnings        = "warnings"
	FeatureScheduled       = "scheduled_messages"
	FeatureVerification    = "verification"
	FeatureTickets         = "tickets"
	FeatureAntiRaid        = "anti_raid"
	FeatureAnalytics       = "analytics"
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
			FeatureReactionRoles:   false,
			FeatureWarnings:        false,
			FeatureScheduled:       false,
			FeatureVerification:    false,
			FeatureTickets:         false,
			FeatureAntiRaid:        false,
			FeatureAnalytics:       false,
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
			"anti_raid_trigger",
		},
		AutoModBlockLinks:       true,
		AutoModBlockedWords:     []string{},
		AutoModDupWindowSec:     20,
		AutoModDupThreshold:     3,
		AutoModAction:           "delete_warn",
		WarnQuarantineThreshold: 3,
		WarnKickThreshold:       5,
		VerificationPhrase:      "!verify",
		TicketOpenPhrase:        "!ticket",
		TicketClosePhrase:       "!close",
		TicketAutoCloseMinutes:  0,
		AntiRaidJoinThreshold:   6,
		AntiRaidWindowSeconds:   30,
		AntiRaidCooldownMinutes: 10,
		AntiRaidAction:          "verification_only",
		AnalyticsIntervalDays:   7,
	}
}
