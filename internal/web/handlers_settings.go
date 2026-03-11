package web

import (
	"encoding/json"
	"net/http"
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
			AppealsChannelID        string          `json:"appeals_channel_id"`
			AppealsLogChannelID     string          `json:"appeals_log_channel_id"`
			AppealsOpenPhrase       string          `json:"appeals_open_phrase"`
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
