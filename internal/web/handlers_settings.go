package web

import (
	"encoding/json"
	"net/http"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		cfg, err := s.repos.Settings.Get(r.Context(), guildID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, cfg)
	case http.MethodPut:
		var cfg struct {
			InactiveDays                int                  `json:"inactive_days"`
			BackfillDays                int                  `json:"backfill_days"`
			QuarantineRoleID            string               `json:"quarantine_role_id"`
			ReadmeChannelID             string               `json:"readme_channel_id"`
			AllowlistRoleIDs            []string             `json:"allowlist_role_ids"`
			AdminUserPolicy             string               `json:"admin_user_policy"`
			BackfillConcurrency         int                  `json:"backfill_concurrency"`
			BackfillIncludeTypes        []string             `json:"backfill_include_types"`
			SafeQuarantineMode          bool                 `json:"safe_quarantine_mode"`
			ActionDryRun                bool                 `json:"action_dry_run"`
			ActionRequireConfirm        bool                 `json:"action_require_confirm"`
			ActionTwoPersonApproval     bool                 `json:"action_two_person_approval"`
			DashboardRolePolicies       map[string][]string  `json:"dashboard_role_policies"`
			ModuleChannelScopes         map[string][]string  `json:"module_channel_scopes"`
			RetentionDays               int                  `json:"retention_days"`
			RetentionArchiveBeforePurge bool                 `json:"retention_archive_before_purge"`
			IncidentModeEnabled         bool                 `json:"incident_mode_enabled"`
			IncidentModeReason          string               `json:"incident_mode_reason"`
			IncidentModeEndsAt          string               `json:"incident_mode_ends_at"`
			ImmutableAuditTrail         bool                 `json:"immutable_audit_trail"`
			MaintenanceWindowEnabled    bool                 `json:"maintenance_window_enabled"`
			MaintenanceWindowStart      string               `json:"maintenance_window_start"`
			MaintenanceWindowEnd        string               `json:"maintenance_window_end"`
			ReviewQueueEnabled          bool                 `json:"review_queue_enabled"`
			ModSummaryChannelID         string               `json:"mod_summary_channel_id"`
			ModSummaryIntervalHours     int                  `json:"mod_summary_interval_hours"`
			AutoThreadEnabled           bool                 `json:"auto_thread_enabled"`
			AutoThreadChannelID         string               `json:"auto_thread_channel_id"`
			AutoThreadKeywords          []string             `json:"auto_thread_keywords"`
			VoiceRewardsEnabled         bool                 `json:"voice_rewards_enabled"`
			VoiceRewardCoinsPerMinute   int                  `json:"voice_reward_coins_per_minute"`
			VoiceRewardXPPerMinute      int                  `json:"voice_reward_xp_per_minute"`
			ConfessionsEnabled          bool                 `json:"confessions_enabled"`
			ConfessionsChannelID        string               `json:"confessions_channel_id"`
			ConfessionsRequireReview    bool                 `json:"confessions_require_review"`
			Web3WhaleAlertsEnabled      bool                 `json:"web3_whale_alerts_enabled"`
			Web3WhaleMinTradeUSD        int                  `json:"web3_whale_min_trade_usd"`
			Web3PriceAlertsEnabled      bool                 `json:"web3_price_alerts_enabled"`
			Web3PriceAlertPumpPct       int                  `json:"web3_price_alert_pump_pct"`
			Web3PriceAlertDumpPct       int                  `json:"web3_price_alert_dump_pct"`
			Web3HealthChecksEnabled     bool                 `json:"web3_health_checks_enabled"`
			Web3HealthMinLiquidityUSD   int                  `json:"web3_health_min_liquidity_usd"`
			Web3MiniTAEnabled           bool                 `json:"web3_mini_ta_enabled"`
			Web3TrendSignalsEnabled     bool                 `json:"web3_trend_signals_enabled"`
			Web3RugRiskEnabled          bool                 `json:"web3_rug_risk_enabled"`
			Web3HolderViewEnabled       bool                 `json:"web3_holder_view_enabled"`
			Web3WalletWatchEnabled      bool                 `json:"web3_wallet_watch_enabled"`
			Web3WalletWatchlist         []string             `json:"web3_wallet_watchlist"`
			Web3ConfidenceScoreEnabled  bool                 `json:"web3_confidence_score_enabled"`
			Web3CommandsEnabled         bool                 `json:"web3_commands_enabled"`
			Web3AntiSpamEnabled         bool                 `json:"web3_anti_spam_enabled"`
			Web3PerTokenCooldownSec     int                  `json:"web3_per_token_cooldown_sec"`
			BirthdaysEnabled            bool                 `json:"birthdays_enabled"`
			BirthdaysChannelID          string               `json:"birthdays_channel_id"`
			AutoRoleProgressionEnabled  bool                 `json:"auto_role_progression_enabled"`
			JoinScreeningEnabled        bool                 `json:"join_screening_enabled"`
			JoinScreeningLogChannelID   string               `json:"join_screening_log_channel_id"`
			JoinScreeningAccountAgeDays int                  `json:"join_screening_account_age_days"`
			JoinScreeningRequireAvatar  bool                 `json:"join_screening_require_avatar"`
			RaidPanicEnabled            bool                 `json:"raid_panic_enabled"`
			RaidPanicDefaultMinutes     int                  `json:"raid_panic_default_minutes"`
			RaidPanicSlowmodeSeconds    int                  `json:"raid_panic_slowmode_seconds"`
			StreaksEnabled              bool                 `json:"streaks_enabled"`
			StreakRewardCoins           int                  `json:"streak_reward_coins"`
			StreakRewardXP              int                  `json:"streak_reward_xp"`
			SeasonResetsEnabled         bool                 `json:"season_resets_enabled"`
			SeasonResetCadence          string               `json:"season_reset_cadence"`
			SeasonResetNextRunAt        string               `json:"season_reset_next_run_at"`
			SeasonResetModules          []string             `json:"season_reset_modules"`
			FeatureFlags                map[string]bool      `json:"feature_flags"`
			WelcomeChannelID            string               `json:"welcome_channel_id"`
			WelcomeMessage              string               `json:"welcome_message"`
			GoodbyeChannelID            string               `json:"goodbye_channel_id"`
			GoodbyeMessage              string               `json:"goodbye_message"`
			AuditLogChannelID           string               `json:"audit_log_channel_id"`
			AuditLogEventTypes          []string             `json:"audit_log_event_types"`
			InviteLogChannelID          string               `json:"invite_log_channel_id"`
			AutoModBlockLinks           bool                 `json:"automod_block_links"`
			AutoModBlockedWords         []string             `json:"automod_blocked_words"`
			AutoModDupWindowSec         int                  `json:"automod_dup_window_sec"`
			AutoModDupThreshold         int                  `json:"automod_dup_threshold"`
			AutoModAction               string               `json:"automod_action"`
			AutoModIgnoreChannelIDs     []string             `json:"automod_ignore_channel_ids"`
			AutoModIgnoreRoleIDs        []string             `json:"automod_ignore_role_ids"`
			AutoModRules                []models.AutoModRule `json:"automod_rules"`
			WarningLogChannelID         string               `json:"warning_log_channel_id"`
			WarnQuarantineThreshold     int                  `json:"warn_quarantine_threshold"`
			WarnKickThreshold           int                  `json:"warn_kick_threshold"`
			VerificationChannelID       string               `json:"verification_channel_id"`
			VerificationPhrase          string               `json:"verification_phrase"`
			UnverifiedRoleID            string               `json:"unverified_role_id"`
			VerifiedRoleID              string               `json:"verified_role_id"`
			TicketInboxChannelID        string               `json:"ticket_inbox_channel_id"`
			TicketCategoryID            string               `json:"ticket_category_id"`
			TicketSupportRoleID         string               `json:"ticket_support_role_id"`
			TicketLogChannelID          string               `json:"ticket_log_channel_id"`
			TicketOpenPhrase            string               `json:"ticket_open_phrase"`
			TicketClosePhrase           string               `json:"ticket_close_phrase"`
			TicketAutoCloseMinutes      int                  `json:"ticket_auto_close_minutes"`
			AntiRaidJoinThreshold       int                  `json:"anti_raid_join_threshold"`
			AntiRaidWindowSeconds       int                  `json:"anti_raid_window_seconds"`
			AntiRaidCooldownMinutes     int                  `json:"anti_raid_cooldown_minutes"`
			AntiRaidAction              string               `json:"anti_raid_action"`
			AntiRaidAlertChannelID      string               `json:"anti_raid_alert_channel_id"`
			AnalyticsChannelID          string               `json:"analytics_channel_id"`
			AnalyticsIntervalDays       int                  `json:"analytics_interval_days"`
			StarboardChannelID          string               `json:"starboard_channel_id"`
			StarboardEmoji              string               `json:"starboard_emoji"`
			StarboardThreshold          int                  `json:"starboard_threshold"`
			LevelingChannelID           string               `json:"leveling_channel_id"`
			LevelingXPPerMessage        int                  `json:"leveling_xp_per_message"`
			LevelingCooldownSeconds     int                  `json:"leveling_cooldown_seconds"`
			LevelingCurve               string               `json:"leveling_curve"`
			LevelingXPBase              int                  `json:"leveling_xp_base"`
			GiveawaysChannelID          string               `json:"giveaways_channel_id"`
			GiveawaysReactionEmoji      string               `json:"giveaways_reaction_emoji"`
			PollsChannelID              string               `json:"polls_channel_id"`
			SuggestionsChannelID        string               `json:"suggestions_channel_id"`
			SuggestionsLogChannelID     string               `json:"suggestions_log_channel_id"`
			KeywordAlertsChannelID      string               `json:"keyword_alerts_channel_id"`
			KeywordAlertWords           []string             `json:"keyword_alert_words"`
			AFKSetPhrase                string               `json:"afk_set_phrase"`
			RemindersChannelID          string               `json:"reminders_channel_id"`
			AccountAgeMinDays           int                  `json:"account_age_min_days"`
			AccountAgeAction            string               `json:"account_age_action"`
			AccountAgeLogChannelID      string               `json:"account_age_log_channel_id"`
			NotesLogChannelID           string               `json:"notes_log_channel_id"`
			AppealsChannelID            string               `json:"appeals_channel_id"`
			AppealsLogChannelID         string               `json:"appeals_log_channel_id"`
			AppealsOpenPhrase           string               `json:"appeals_open_phrase"`
		}
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		current, err := s.repos.Settings.Get(r.Context(), guildID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		current.InactiveDays = cfg.InactiveDays
		current.BackfillDays = cfg.BackfillDays
		current.QuarantineRoleID = cfg.QuarantineRoleID
		current.ReadmeChannelID = cfg.ReadmeChannelID
		current.AllowlistRoleIDs = cfg.AllowlistRoleIDs
		current.AdminUserPolicy = cfg.AdminUserPolicy
		current.BackfillConcurrency = cfg.BackfillConcurrency
		current.BackfillIncludeTypes = cfg.BackfillIncludeTypes
		current.SafeQuarantineMode = cfg.SafeQuarantineMode
		current.ActionDryRun = cfg.ActionDryRun
		current.ActionRequireConfirm = cfg.ActionRequireConfirm
		current.ActionTwoPersonApproval = cfg.ActionTwoPersonApproval
		current.DashboardRolePolicies = cfg.DashboardRolePolicies
		current.ModuleChannelScopes = cfg.ModuleChannelScopes
		current.RetentionDays = cfg.RetentionDays
		current.RetentionArchiveBeforePurge = cfg.RetentionArchiveBeforePurge
		current.IncidentModeEnabled = cfg.IncidentModeEnabled
		current.IncidentModeReason = cfg.IncidentModeReason
		current.IncidentModeEndsAt = cfg.IncidentModeEndsAt
		current.ImmutableAuditTrail = cfg.ImmutableAuditTrail
		current.MaintenanceWindowEnabled = cfg.MaintenanceWindowEnabled
		current.MaintenanceWindowStart = cfg.MaintenanceWindowStart
		current.MaintenanceWindowEnd = cfg.MaintenanceWindowEnd
		current.ReviewQueueEnabled = cfg.ReviewQueueEnabled
		current.ModSummaryChannelID = cfg.ModSummaryChannelID
		current.ModSummaryIntervalHours = cfg.ModSummaryIntervalHours
		current.AutoThreadEnabled = cfg.AutoThreadEnabled
		current.AutoThreadChannelID = cfg.AutoThreadChannelID
		current.AutoThreadKeywords = cfg.AutoThreadKeywords
		current.VoiceRewardsEnabled = cfg.VoiceRewardsEnabled
		current.VoiceRewardCoinsPerMinute = cfg.VoiceRewardCoinsPerMinute
		current.VoiceRewardXPPerMinute = cfg.VoiceRewardXPPerMinute
		current.ConfessionsEnabled = cfg.ConfessionsEnabled
		current.ConfessionsChannelID = cfg.ConfessionsChannelID
		current.ConfessionsRequireReview = cfg.ConfessionsRequireReview
		current.Web3WhaleAlertsEnabled = cfg.Web3WhaleAlertsEnabled
		current.Web3WhaleMinTradeUSD = cfg.Web3WhaleMinTradeUSD
		current.Web3PriceAlertsEnabled = cfg.Web3PriceAlertsEnabled
		current.Web3PriceAlertPumpPct = cfg.Web3PriceAlertPumpPct
		current.Web3PriceAlertDumpPct = cfg.Web3PriceAlertDumpPct
		current.Web3HealthChecksEnabled = cfg.Web3HealthChecksEnabled
		current.Web3HealthMinLiquidityUSD = cfg.Web3HealthMinLiquidityUSD
		current.Web3MiniTAEnabled = cfg.Web3MiniTAEnabled
		current.Web3TrendSignalsEnabled = cfg.Web3TrendSignalsEnabled
		current.Web3RugRiskEnabled = cfg.Web3RugRiskEnabled
		current.Web3HolderViewEnabled = cfg.Web3HolderViewEnabled
		current.Web3WalletWatchEnabled = cfg.Web3WalletWatchEnabled
		current.Web3WalletWatchlist = cfg.Web3WalletWatchlist
		current.Web3ConfidenceScoreEnabled = cfg.Web3ConfidenceScoreEnabled
		current.Web3CommandsEnabled = cfg.Web3CommandsEnabled
		current.Web3AntiSpamEnabled = cfg.Web3AntiSpamEnabled
		current.Web3PerTokenCooldownSec = cfg.Web3PerTokenCooldownSec
		current.BirthdaysEnabled = cfg.BirthdaysEnabled
		current.BirthdaysChannelID = cfg.BirthdaysChannelID
		current.AutoRoleProgressionEnabled = cfg.AutoRoleProgressionEnabled
		current.JoinScreeningEnabled = cfg.JoinScreeningEnabled
		current.JoinScreeningLogChannelID = cfg.JoinScreeningLogChannelID
		current.JoinScreeningAccountAgeDays = cfg.JoinScreeningAccountAgeDays
		current.JoinScreeningRequireAvatar = cfg.JoinScreeningRequireAvatar
		current.RaidPanicEnabled = cfg.RaidPanicEnabled
		current.RaidPanicDefaultMinutes = cfg.RaidPanicDefaultMinutes
		current.RaidPanicSlowmodeSeconds = cfg.RaidPanicSlowmodeSeconds
		current.StreaksEnabled = cfg.StreaksEnabled
		current.StreakRewardCoins = cfg.StreakRewardCoins
		current.StreakRewardXP = cfg.StreakRewardXP
		current.SeasonResetsEnabled = cfg.SeasonResetsEnabled
		current.SeasonResetCadence = cfg.SeasonResetCadence
		current.SeasonResetNextRunAt = cfg.SeasonResetNextRunAt
		current.SeasonResetModules = cfg.SeasonResetModules
		current.FeatureFlags = cfg.FeatureFlags
		current.WelcomeChannelID = cfg.WelcomeChannelID
		current.WelcomeMessage = cfg.WelcomeMessage
		current.GoodbyeChannelID = cfg.GoodbyeChannelID
		current.GoodbyeMessage = cfg.GoodbyeMessage
		current.AuditLogChannelID = cfg.AuditLogChannelID
		current.AuditLogEventTypes = cfg.AuditLogEventTypes
		current.InviteLogChannelID = cfg.InviteLogChannelID
		current.AutoModBlockLinks = cfg.AutoModBlockLinks
		current.AutoModBlockedWords = cfg.AutoModBlockedWords
		current.AutoModDupWindowSec = cfg.AutoModDupWindowSec
		current.AutoModDupThreshold = cfg.AutoModDupThreshold
		current.AutoModAction = cfg.AutoModAction
		current.AutoModIgnoreChannelIDs = cfg.AutoModIgnoreChannelIDs
		current.AutoModIgnoreRoleIDs = cfg.AutoModIgnoreRoleIDs
		current.AutoModRules = cfg.AutoModRules
		current.WarningLogChannelID = cfg.WarningLogChannelID
		current.WarnQuarantineThreshold = cfg.WarnQuarantineThreshold
		current.WarnKickThreshold = cfg.WarnKickThreshold
		current.VerificationChannelID = cfg.VerificationChannelID
		current.VerificationPhrase = cfg.VerificationPhrase
		current.UnverifiedRoleID = cfg.UnverifiedRoleID
		current.VerifiedRoleID = cfg.VerifiedRoleID
		current.TicketInboxChannelID = cfg.TicketInboxChannelID
		current.TicketCategoryID = cfg.TicketCategoryID
		current.TicketSupportRoleID = cfg.TicketSupportRoleID
		current.TicketLogChannelID = cfg.TicketLogChannelID
		current.TicketOpenPhrase = cfg.TicketOpenPhrase
		current.TicketClosePhrase = cfg.TicketClosePhrase
		current.TicketAutoCloseMinutes = cfg.TicketAutoCloseMinutes
		current.AntiRaidJoinThreshold = cfg.AntiRaidJoinThreshold
		current.AntiRaidWindowSeconds = cfg.AntiRaidWindowSeconds
		current.AntiRaidCooldownMinutes = cfg.AntiRaidCooldownMinutes
		current.AntiRaidAction = cfg.AntiRaidAction
		current.AntiRaidAlertChannelID = cfg.AntiRaidAlertChannelID
		current.AnalyticsChannelID = cfg.AnalyticsChannelID
		current.AnalyticsIntervalDays = cfg.AnalyticsIntervalDays
		current.StarboardChannelID = cfg.StarboardChannelID
		current.StarboardEmoji = cfg.StarboardEmoji
		current.StarboardThreshold = cfg.StarboardThreshold
		current.LevelingChannelID = cfg.LevelingChannelID
		current.LevelingXPPerMessage = cfg.LevelingXPPerMessage
		current.LevelingCooldownSeconds = cfg.LevelingCooldownSeconds
		current.LevelingCurve = cfg.LevelingCurve
		current.LevelingXPBase = cfg.LevelingXPBase
		current.GiveawaysChannelID = cfg.GiveawaysChannelID
		current.GiveawaysReactionEmoji = cfg.GiveawaysReactionEmoji
		current.PollsChannelID = cfg.PollsChannelID
		current.SuggestionsChannelID = cfg.SuggestionsChannelID
		current.SuggestionsLogChannelID = cfg.SuggestionsLogChannelID
		current.KeywordAlertsChannelID = cfg.KeywordAlertsChannelID
		current.KeywordAlertWords = cfg.KeywordAlertWords
		current.AFKSetPhrase = cfg.AFKSetPhrase
		current.RemindersChannelID = cfg.RemindersChannelID
		current.AccountAgeMinDays = cfg.AccountAgeMinDays
		current.AccountAgeAction = cfg.AccountAgeAction
		current.AccountAgeLogChannelID = cfg.AccountAgeLogChannelID
		current.NotesLogChannelID = cfg.NotesLogChannelID
		current.AppealsChannelID = cfg.AppealsChannelID
		current.AppealsLogChannelID = cfg.AppealsLogChannelID
		current.AppealsOpenPhrase = cfg.AppealsOpenPhrase

		if err := s.repos.Settings.Upsert(r.Context(), current); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, current)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGuilds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guilds, err := s.discord.ListGuilds(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, guilds)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleBackfillStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	settings, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id, err := s.discord.StartBackfill(r.Context(), guildID, settings.BackfillDays)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"job_id": id})
}

func (s *Server) handleBackfillStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.discord.BackfillStatus())
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
