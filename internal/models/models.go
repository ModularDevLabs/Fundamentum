package models

import (
	"strconv"
	"strings"
	"time"
)

type AutoModRule struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Type      string `json:"type"`
	Pattern   string `json:"pattern"`
	Threshold int    `json:"threshold"`
	Action    string `json:"action"`
}

type GuildSettings struct {
	GuildID                     string              `json:"guild_id"`
	InactiveDays                int                 `json:"inactive_days"`
	BackfillDays                int                 `json:"backfill_days"`
	QuarantineRoleID            string              `json:"quarantine_role_id"`
	ReadmeChannelID             string              `json:"readme_channel_id"`
	AllowlistRoleIDs            []string            `json:"allowlist_role_ids"`
	AdminUserPolicy             string              `json:"admin_user_policy"`
	BackfillConcurrency         int                 `json:"backfill_concurrency"`
	BackfillIncludeTypes        []string            `json:"backfill_include_types"`
	SafeQuarantineMode          bool                `json:"safe_quarantine_mode"`
	ActionDryRun                bool                `json:"action_dry_run"`
	ActionRequireConfirm        bool                `json:"action_require_confirm"`
	ActionTwoPersonApproval     bool                `json:"action_two_person_approval"`
	DashboardRolePolicies       map[string][]string `json:"dashboard_role_policies"`
	ModuleChannelScopes         map[string][]string `json:"module_channel_scopes"`
	RetentionDays               int                 `json:"retention_days"`
	RetentionArchiveBeforePurge bool                `json:"retention_archive_before_purge"`
	IncidentModeEnabled         bool                `json:"incident_mode_enabled"`
	IncidentModeReason          string              `json:"incident_mode_reason"`
	IncidentModeEndsAt          string              `json:"incident_mode_ends_at"`
	ImmutableAuditTrail         bool                `json:"immutable_audit_trail"`
	MaintenanceWindowEnabled    bool                `json:"maintenance_window_enabled"`
	MaintenanceWindowStart      string              `json:"maintenance_window_start"`
	MaintenanceWindowEnd        string              `json:"maintenance_window_end"`
	ReviewQueueEnabled          bool                `json:"review_queue_enabled"`
	ModSummaryChannelID         string              `json:"mod_summary_channel_id"`
	ModSummaryIntervalHours     int                 `json:"mod_summary_interval_hours"`
	AutoThreadEnabled           bool                `json:"auto_thread_enabled"`
	AutoThreadChannelID         string              `json:"auto_thread_channel_id"`
	AutoThreadKeywords          []string            `json:"auto_thread_keywords"`
	VoiceRewardsEnabled         bool                `json:"voice_rewards_enabled"`
	VoiceRewardCoinsPerMinute   int                 `json:"voice_reward_coins_per_minute"`
	VoiceRewardXPPerMinute      int                 `json:"voice_reward_xp_per_minute"`
	ConfessionsEnabled          bool                `json:"confessions_enabled"`
	ConfessionsChannelID        string              `json:"confessions_channel_id"`
	ConfessionsRequireReview    bool                `json:"confessions_require_review"`
	Web3WhaleAlertsEnabled      bool                `json:"web3_whale_alerts_enabled"`
	Web3WhaleMinTradeUSD        int                 `json:"web3_whale_min_trade_usd"`
	Web3PriceAlertsEnabled      bool                `json:"web3_price_alerts_enabled"`
	Web3PriceAlertPumpPct       int                 `json:"web3_price_alert_pump_pct"`
	Web3PriceAlertDumpPct       int                 `json:"web3_price_alert_dump_pct"`
	Web3HealthChecksEnabled     bool                `json:"web3_health_checks_enabled"`
	Web3HealthMinLiquidityUSD   int                 `json:"web3_health_min_liquidity_usd"`
	Web3MiniTAEnabled           bool                `json:"web3_mini_ta_enabled"`
	Web3TrendSignalsEnabled     bool                `json:"web3_trend_signals_enabled"`
	Web3RugRiskEnabled          bool                `json:"web3_rug_risk_enabled"`
	Web3HolderViewEnabled       bool                `json:"web3_holder_view_enabled"`
	Web3WalletWatchEnabled      bool                `json:"web3_wallet_watch_enabled"`
	Web3WalletWatchlist         []string            `json:"web3_wallet_watchlist"`
	Web3ConfidenceScoreEnabled  bool                `json:"web3_confidence_score_enabled"`
	Web3CommandsEnabled         bool                `json:"web3_commands_enabled"`
	Web3AntiSpamEnabled         bool                `json:"web3_anti_spam_enabled"`
	Web3PerTokenCooldownSec     int                 `json:"web3_per_token_cooldown_sec"`
	BirthdaysEnabled            bool                `json:"birthdays_enabled"`
	BirthdaysChannelID          string              `json:"birthdays_channel_id"`
	AutoRoleProgressionEnabled  bool                `json:"auto_role_progression_enabled"`
	JoinScreeningEnabled        bool                `json:"join_screening_enabled"`
	JoinScreeningLogChannelID   string              `json:"join_screening_log_channel_id"`
	JoinScreeningAccountAgeDays int                 `json:"join_screening_account_age_days"`
	JoinScreeningRequireAvatar  bool                `json:"join_screening_require_avatar"`
	RaidPanicEnabled            bool                `json:"raid_panic_enabled"`
	RaidPanicDefaultMinutes     int                 `json:"raid_panic_default_minutes"`
	RaidPanicSlowmodeSeconds    int                 `json:"raid_panic_slowmode_seconds"`
	StreaksEnabled              bool                `json:"streaks_enabled"`
	StreakRewardCoins           int                 `json:"streak_reward_coins"`
	StreakRewardXP              int                 `json:"streak_reward_xp"`
	SeasonResetsEnabled         bool                `json:"season_resets_enabled"`
	SeasonResetCadence          string              `json:"season_reset_cadence"`
	SeasonResetNextRunAt        string              `json:"season_reset_next_run_at"`
	SeasonResetModules          []string            `json:"season_reset_modules"`
	FeatureFlags                map[string]bool     `json:"feature_flags"`
	WelcomeChannelID            string              `json:"welcome_channel_id"`
	WelcomeMessage              string              `json:"welcome_message"`
	GoodbyeChannelID            string              `json:"goodbye_channel_id"`
	GoodbyeMessage              string              `json:"goodbye_message"`
	AuditLogChannelID           string              `json:"audit_log_channel_id"`
	AuditLogEventTypes          []string            `json:"audit_log_event_types"`
	InviteLogChannelID          string              `json:"invite_log_channel_id"`
	AutoModBlockLinks           bool                `json:"automod_block_links"`
	AutoModBlockedWords         []string            `json:"automod_blocked_words"`
	AutoModDupWindowSec         int                 `json:"automod_dup_window_sec"`
	AutoModDupThreshold         int                 `json:"automod_dup_threshold"`
	AutoModAction               string              `json:"automod_action"`
	AutoModIgnoreChannelIDs     []string            `json:"automod_ignore_channel_ids"`
	AutoModIgnoreRoleIDs        []string            `json:"automod_ignore_role_ids"`
	AutoModRules                []AutoModRule       `json:"automod_rules"`
	WarningLogChannelID         string              `json:"warning_log_channel_id"`
	WarnQuarantineThreshold     int                 `json:"warn_quarantine_threshold"`
	WarnKickThreshold           int                 `json:"warn_kick_threshold"`
	VerificationChannelID       string              `json:"verification_channel_id"`
	VerificationPhrase          string              `json:"verification_phrase"`
	UnverifiedRoleID            string              `json:"unverified_role_id"`
	VerifiedRoleID              string              `json:"verified_role_id"`
	TicketInboxChannelID        string              `json:"ticket_inbox_channel_id"`
	TicketCategoryID            string              `json:"ticket_category_id"`
	TicketSupportRoleID         string              `json:"ticket_support_role_id"`
	TicketLogChannelID          string              `json:"ticket_log_channel_id"`
	TicketOpenPhrase            string              `json:"ticket_open_phrase"`
	TicketClosePhrase           string              `json:"ticket_close_phrase"`
	TicketAutoCloseMinutes      int                 `json:"ticket_auto_close_minutes"`
	AntiRaidJoinThreshold       int                 `json:"anti_raid_join_threshold"`
	AntiRaidWindowSeconds       int                 `json:"anti_raid_window_seconds"`
	AntiRaidCooldownMinutes     int                 `json:"anti_raid_cooldown_minutes"`
	AntiRaidAction              string              `json:"anti_raid_action"`
	AntiRaidAlertChannelID      string              `json:"anti_raid_alert_channel_id"`
	AnalyticsChannelID          string              `json:"analytics_channel_id"`
	AnalyticsIntervalDays       int                 `json:"analytics_interval_days"`
	StarboardChannelID          string              `json:"starboard_channel_id"`
	StarboardEmoji              string              `json:"starboard_emoji"`
	StarboardThreshold          int                 `json:"starboard_threshold"`
	LevelingChannelID           string              `json:"leveling_channel_id"`
	LevelingXPPerMessage        int                 `json:"leveling_xp_per_message"`
	LevelingCooldownSeconds     int                 `json:"leveling_cooldown_seconds"`
	LevelingCurve               string              `json:"leveling_curve"`
	LevelingXPBase              int                 `json:"leveling_xp_base"`
	GiveawaysChannelID          string              `json:"giveaways_channel_id"`
	GiveawaysReactionEmoji      string              `json:"giveaways_reaction_emoji"`
	PollsChannelID              string              `json:"polls_channel_id"`
	SuggestionsChannelID        string              `json:"suggestions_channel_id"`
	SuggestionsLogChannelID     string              `json:"suggestions_log_channel_id"`
	KeywordAlertsChannelID      string              `json:"keyword_alerts_channel_id"`
	KeywordAlertWords           []string            `json:"keyword_alert_words"`
	AFKSetPhrase                string              `json:"afk_set_phrase"`
	RemindersChannelID          string              `json:"reminders_channel_id"`
	AccountAgeMinDays           int                 `json:"account_age_min_days"`
	AccountAgeAction            string              `json:"account_age_action"`
	AccountAgeLogChannelID      string              `json:"account_age_log_channel_id"`
	NotesLogChannelID           string              `json:"notes_log_channel_id"`
	AppealsChannelID            string              `json:"appeals_channel_id"`
	AppealsLogChannelID         string              `json:"appeals_log_channel_id"`
	AppealsOpenPhrase           string              `json:"appeals_open_phrase"`
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
	GroupKey        string    `json:"group_key"`
	MaxSelect       int       `json:"max_select"`
	MinSelect       int       `json:"min_select"`
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

type AppealRow struct {
	ID         int64      `json:"id"`
	GuildID    string     `json:"guild_id"`
	UserID     string     `json:"user_id"`
	Reason     string     `json:"reason"`
	Status     string     `json:"status"`
	Resolution string     `json:"resolution"`
	ReviewedBy string     `json:"reviewed_by"`
	CreatedAt  time.Time  `json:"created_at"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
}

type CustomCommandRow struct {
	ID        int64     `json:"id"`
	GuildID   string    `json:"guild_id"`
	Trigger   string    `json:"trigger"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
}

type StarboardEntryRow struct {
	ID               int64      `json:"id"`
	GuildID          string     `json:"guild_id"`
	SourceChannelID  string     `json:"source_channel_id"`
	SourceMessageID  string     `json:"source_message_id"`
	StarboardChannel string     `json:"starboard_channel_id"`
	StarboardMessage string     `json:"starboard_message_id"`
	StarCount        int        `json:"star_count"`
	LastUpdatedAt    time.Time  `json:"last_updated_at"`
	PostedAt         *time.Time `json:"posted_at,omitempty"`
}

type MemberLevelRow struct {
	GuildID  string    `json:"guild_id"`
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	XP       int       `json:"xp"`
	Level    int       `json:"level"`
	LastXPAt time.Time `json:"last_xp_at"`
}

type GiveawayRow struct {
	ID          int64     `json:"id"`
	GuildID     string    `json:"guild_id"`
	ChannelID   string    `json:"channel_id"`
	MessageID   string    `json:"message_id"`
	Prize       string    `json:"prize"`
	WinnerCount int       `json:"winner_count"`
	EndsAt      time.Time `json:"ends_at"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	EntryCount  int       `json:"entry_count"`
}

type PollRow struct {
	ID         int64      `json:"id"`
	GuildID    string     `json:"guild_id"`
	ChannelID  string     `json:"channel_id"`
	MessageID  string     `json:"message_id"`
	Question   string     `json:"question"`
	Options    []string   `json:"options"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ClosedAt   *time.Time `json:"closed_at,omitempty"`
	TotalVotes int        `json:"total_votes"`
}

type SuggestionRow struct {
	ID           int64      `json:"id"`
	GuildID      string     `json:"guild_id"`
	UserID       string     `json:"user_id"`
	Content      string     `json:"content"`
	MessageID    string     `json:"message_id"`
	ChannelID    string     `json:"channel_id"`
	Status       string     `json:"status"`
	DecisionBy   string     `json:"decision_by"`
	DecisionNote string     `json:"decision_note"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type AFKStatusRow struct {
	GuildID   string    `json:"guild_id"`
	UserID    string    `json:"user_id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type ReminderRow struct {
	ID        int64     `json:"id"`
	GuildID   string    `json:"guild_id"`
	ChannelID string    `json:"channel_id"`
	Content   string    `json:"content"`
	RunAt     time.Time `json:"run_at"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type BirthdayRow struct {
	GuildID      string    `json:"guild_id"`
	UserID       string    `json:"user_id"`
	BirthdayMMDD string    `json:"birthday_mmdd"`
	Timezone     string    `json:"timezone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RoleProgressionRuleRow struct {
	ID        int64     `json:"id"`
	GuildID   string    `json:"guild_id"`
	Metric    string    `json:"metric"`
	Threshold int       `json:"threshold"`
	RoleID    string    `json:"role_id"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type JoinScreeningRow struct {
	ID               int64      `json:"id"`
	GuildID          string     `json:"guild_id"`
	UserID           string     `json:"user_id"`
	Username         string     `json:"username"`
	AccountCreatedAt *time.Time `json:"account_created_at,omitempty"`
	Reason           string     `json:"reason"`
	Status           string     `json:"status"`
	ReviewedBy       string     `json:"reviewed_by"`
	CreatedAt        time.Time  `json:"created_at"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
}

type StreakRow struct {
	GuildID        string    `json:"guild_id"`
	UserID         string    `json:"user_id"`
	CurrentStreak  int       `json:"current_streak"`
	BestStreak     int       `json:"best_streak"`
	LastActiveDate string    `json:"last_active_date"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SeasonResetRunRow struct {
	ID           int64            `json:"id"`
	GuildID      string           `json:"guild_id"`
	TriggeredBy  string           `json:"triggered_by"`
	Modules      []string         `json:"modules"`
	AffectedRows map[string]int64 `json:"affected_rows"`
	Status       string           `json:"status"`
	Error        string           `json:"error"`
	StartedAt    time.Time        `json:"started_at"`
	CompletedAt  time.Time        `json:"completed_at"`
}

type MemberNoteRow struct {
	ID         int64      `json:"id"`
	GuildID    string     `json:"guild_id"`
	UserID     string     `json:"user_id"`
	AuthorID   string     `json:"author_id"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

type GuildInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WebhookIntegrationRow struct {
	ID        int64     `json:"id"`
	GuildID   string    `json:"guild_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Enabled   bool      `json:"enabled"`
	LastError string    `json:"last_error"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditTrailRow struct {
	ID         int64     `json:"id"`
	GuildID    string    `json:"guild_id"`
	EventType  string    `json:"event_type"`
	Message    string    `json:"message"`
	Payload    string    `json:"payload"`
	PrevHash   string    `json:"prev_hash"`
	EventHash  string    `json:"event_hash"`
	RecordedAt time.Time `json:"recorded_at"`
}

type Web3FirstScanRow struct {
	GuildID            string    `json:"guild_id"`
	AssetKey           string    `json:"asset_key"`
	AssetType          string    `json:"asset_type"`
	DisplaySymbol      string    `json:"display_symbol"`
	DisplayName        string    `json:"display_name"`
	FirstScannerUserID string    `json:"first_scanner_user_id"`
	FirstScannerName   string    `json:"first_scanner_name"`
	FirstPriceUSD      float64   `json:"first_price_usd"`
	FirstScannedAt     time.Time `json:"first_scanned_at"`
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
	FeatureStarboard       = "starboard"
	FeatureLeveling        = "leveling"
	FeatureGiveaways       = "giveaways"
	FeaturePolls           = "polls"
	FeatureSuggestions     = "suggestions"
	FeatureKeywordAlerts   = "keyword_alerts"
	FeatureAFK             = "afk"
	FeatureReminders       = "reminders"
	FeatureAccountAgeGuard = "account_age_guard"
	FeatureMemberNotes     = "member_notes"
	FeatureAppeals         = "appeals"
	FeatureCustomCommands  = "custom_commands"
	FeatureBirthdays       = "birthdays"
	FeatureRoleProgression = "role_progression"
	FeatureJoinScreening   = "join_screening"
	FeatureRaidPanic       = "raid_panic"
	FeatureStreaks         = "streaks"
	FeatureSeasonResets    = "season_resets"
	FeatureReputation      = "reputation"
	FeatureEconomy         = "economy"
	FeatureAchievements    = "achievements"
	FeatureTrivia          = "trivia"
	FeatureCalendar        = "calendar"
	FeatureConfessions     = "confessions"
	FeatureWeb3Intel       = "web3_intel"
)

func (s GuildSettings) FeatureEnabled(flag string) bool {
	if s.FeatureFlags == nil {
		return false
	}
	return s.FeatureFlags[flag]
}

func (s GuildSettings) FeatureAllowedInChannel(flag, channelID string) bool {
	if channelID == "" {
		return true
	}
	if s.ModuleChannelScopes == nil {
		return true
	}
	scope := s.ModuleChannelScopes[flag]
	if len(scope) == 0 {
		return true
	}
	for _, channel := range scope {
		if channel == channelID {
			return true
		}
	}
	return false
}

func (s GuildSettings) InMaintenanceWindow(now time.Time) bool {
	if !s.MaintenanceWindowEnabled {
		return false
	}
	startMin, okStart := parseHHMM(s.MaintenanceWindowStart)
	endMin, okEnd := parseHHMM(s.MaintenanceWindowEnd)
	if !okStart || !okEnd {
		return false
	}
	currentMin := now.UTC().Hour()*60 + now.UTC().Minute()
	if startMin == endMin {
		return true
	}
	if startMin < endMin {
		return currentMin >= startMin && currentMin < endMin
	}
	return currentMin >= startMin || currentMin < endMin
}

func parseHHMM(raw string) (int, bool) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 2 {
		return 0, false
	}
	hh, err := strconv.Atoi(parts[0])
	if err != nil || hh < 0 || hh > 23 {
		return 0, false
	}
	mm, err := strconv.Atoi(parts[1])
	if err != nil || mm < 0 || mm > 59 {
		return 0, false
	}
	return hh*60 + mm, true
}

func DefaultGuildSettings(guildID string) GuildSettings {
	return GuildSettings{
		GuildID:                     guildID,
		InactiveDays:                180,
		BackfillDays:                60,
		AdminUserPolicy:             "refuse",
		BackfillConcurrency:         2,
		SafeQuarantineMode:          false,
		ActionDryRun:                false,
		ActionRequireConfirm:        true,
		ActionTwoPersonApproval:     false,
		DashboardRolePolicies:       map[string][]string{},
		ModuleChannelScopes:         map[string][]string{},
		RetentionDays:               0,
		RetentionArchiveBeforePurge: true,
		IncidentModeEnabled:         false,
		IncidentModeReason:          "",
		IncidentModeEndsAt:          "",
		ImmutableAuditTrail:         false,
		MaintenanceWindowEnabled:    false,
		MaintenanceWindowStart:      "02:00",
		MaintenanceWindowEnd:        "03:00",
		ReviewQueueEnabled:          false,
		ModSummaryChannelID:         "",
		ModSummaryIntervalHours:     24,
		AutoThreadEnabled:           false,
		AutoThreadChannelID:         "",
		AutoThreadKeywords:          []string{"?", "help", "issue"},
		VoiceRewardsEnabled:         false,
		VoiceRewardCoinsPerMinute:   1,
		VoiceRewardXPPerMinute:      2,
		ConfessionsEnabled:          false,
		ConfessionsChannelID:        "",
		ConfessionsRequireReview:    true,
		Web3WhaleAlertsEnabled:      false,
		Web3WhaleMinTradeUSD:        25000,
		Web3PriceAlertsEnabled:      false,
		Web3PriceAlertPumpPct:       25,
		Web3PriceAlertDumpPct:       25,
		Web3HealthChecksEnabled:     true,
		Web3HealthMinLiquidityUSD:   20000,
		Web3MiniTAEnabled:           false,
		Web3TrendSignalsEnabled:     true,
		Web3RugRiskEnabled:          true,
		Web3HolderViewEnabled:       true,
		Web3WalletWatchEnabled:      false,
		Web3WalletWatchlist:         []string{},
		Web3ConfidenceScoreEnabled:  true,
		Web3CommandsEnabled:         true,
		Web3AntiSpamEnabled:         true,
		Web3PerTokenCooldownSec:     30,
		BirthdaysEnabled:            false,
		BirthdaysChannelID:          "",
		AutoRoleProgressionEnabled:  false,
		JoinScreeningEnabled:        false,
		JoinScreeningLogChannelID:   "",
		JoinScreeningAccountAgeDays: 7,
		JoinScreeningRequireAvatar:  false,
		RaidPanicEnabled:            false,
		RaidPanicDefaultMinutes:     30,
		RaidPanicSlowmodeSeconds:    10,
		StreaksEnabled:              false,
		StreakRewardCoins:           5,
		StreakRewardXP:              10,
		SeasonResetsEnabled:         false,
		SeasonResetCadence:          "monthly",
		SeasonResetNextRunAt:        "",
		SeasonResetModules:          []string{"leveling", "economy", "trivia"},
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
			FeatureStarboard:       false,
			FeatureLeveling:        false,
			FeatureGiveaways:       false,
			FeaturePolls:           false,
			FeatureSuggestions:     false,
			FeatureKeywordAlerts:   false,
			FeatureAFK:             false,
			FeatureReminders:       false,
			FeatureAccountAgeGuard: false,
			FeatureMemberNotes:     false,
			FeatureAppeals:         false,
			FeatureCustomCommands:  false,
			FeatureBirthdays:       false,
			FeatureRoleProgression: false,
			FeatureJoinScreening:   false,
			FeatureRaidPanic:       false,
			FeatureStreaks:         false,
			FeatureSeasonResets:    false,
			FeatureReputation:      false,
			FeatureEconomy:         false,
			FeatureAchievements:    false,
			FeatureTrivia:          false,
			FeatureCalendar:        false,
			FeatureConfessions:     false,
			FeatureWeb3Intel:       false,
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
		AutoModRules:            []AutoModRule{},
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
		StarboardEmoji:          "⭐",
		StarboardThreshold:      3,
		LevelingXPPerMessage:    10,
		LevelingCooldownSeconds: 60,
		LevelingCurve:           "quadratic",
		LevelingXPBase:          100,
		GiveawaysReactionEmoji:  "🎉",
		AFKSetPhrase:            "!afk",
		AccountAgeMinDays:       7,
		AccountAgeAction:        "log_only",
		AppealsOpenPhrase:       "!appeal",
	}
}
