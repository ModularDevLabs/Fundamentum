package web

import (
	"net/http"
	"strings"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

func (s *Server) handleDependencyCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cfg, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type checkRow struct {
		Module   string `json:"module"`
		Severity string `json:"severity"`
		Message  string `json:"message"`
	}
	rows := make([]checkRow, 0, 24)
	add := func(module, severity, message string) {
		rows = append(rows, checkRow{Module: module, Severity: severity, Message: message})
	}

	if cfg.FeatureEnabled(models.FeatureWelcomeMessages) && cfg.WelcomeChannelID == "" {
		add("welcome_messages", "error", "Enabled but welcome channel ID is empty.")
	}
	if cfg.FeatureEnabled(models.FeatureGoodbyeMessages) && cfg.GoodbyeChannelID == "" {
		add("goodbye_messages", "error", "Enabled but goodbye channel ID is empty.")
	}
	if cfg.FeatureEnabled(models.FeatureAuditLogStream) && cfg.AuditLogChannelID == "" {
		add("audit_log_stream", "error", "Enabled but audit log channel ID is empty.")
	}
	if cfg.FeatureEnabled(models.FeatureInviteTracker) && cfg.InviteLogChannelID == "" {
		add("invite_tracker", "warn", "Enabled without a dedicated invite log channel.")
	}
	if cfg.FeatureEnabled(models.FeatureVerification) {
		if cfg.VerificationChannelID == "" {
			add("verification", "error", "Enabled but verification channel ID is empty.")
		}
		if cfg.UnverifiedRoleID == "" {
			add("verification", "warn", "Enabled but unverified role ID is empty.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureTickets) {
		if cfg.TicketInboxChannelID == "" {
			add("tickets", "error", "Enabled but ticket inbox channel ID is empty.")
		}
		if cfg.TicketCategoryID == "" {
			add("tickets", "warn", "Enabled but ticket category ID is empty.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureAntiRaid) {
		if cfg.AntiRaidJoinThreshold <= 0 || cfg.AntiRaidWindowSeconds <= 0 {
			add("anti_raid", "error", "Enabled with invalid threshold/window configuration.")
		}
		if cfg.AntiRaidAction == "verification_only" && cfg.UnverifiedRoleID == "" {
			add("anti_raid", "warn", "verification_only action configured but unverified role ID is empty.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureAnalytics) && cfg.AnalyticsChannelID == "" {
		add("analytics", "warn", "Enabled without analytics channel ID.")
	}
	if cfg.FeatureEnabled(models.FeatureStarboard) {
		if cfg.StarboardChannelID == "" {
			add("starboard", "error", "Enabled but starboard channel ID is empty.")
		}
		if cfg.StarboardEmoji == "" {
			add("starboard", "error", "Enabled but starboard emoji is empty.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureLeveling) && cfg.LevelingXPPerMessage <= 0 {
		add("leveling", "error", "Enabled with invalid XP per message value.")
	}
	if cfg.FeatureEnabled(models.FeatureGiveaways) {
		if cfg.GiveawaysChannelID == "" {
			add("giveaways", "warn", "Enabled without a default giveaways channel.")
		}
		if cfg.GiveawaysReactionEmoji == "" {
			add("giveaways", "error", "Enabled but entry emoji is empty.")
		}
	}
	if cfg.FeatureEnabled(models.FeaturePolls) && cfg.PollsChannelID == "" {
		add("polls", "warn", "Enabled without default poll channel.")
	}
	if cfg.FeatureEnabled(models.FeatureSuggestions) && cfg.SuggestionsChannelID == "" {
		add("suggestions", "error", "Enabled but suggestions channel ID is empty.")
	}
	if cfg.FeatureEnabled(models.FeatureKeywordAlerts) {
		if cfg.KeywordAlertsChannelID == "" {
			add("keyword_alerts", "error", "Enabled but keyword alerts channel ID is empty.")
		}
		if len(cfg.KeywordAlertWords) == 0 {
			add("keyword_alerts", "warn", "Enabled with no keyword list configured.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureReminders) && cfg.RemindersChannelID == "" {
		add("reminders", "warn", "Enabled without default reminders channel.")
	}
	if cfg.FeatureEnabled(models.FeatureBirthdays) && cfg.BirthdaysChannelID == "" {
		add("birthdays", "warn", "Enabled without a birthday announcement channel.")
	}
	if cfg.FeatureEnabled(models.FeatureRoleProgression) && !cfg.AutoRoleProgressionEnabled {
		add("role_progression", "warn", "Feature flag enabled but auto role progression toggle is off.")
	}
	if cfg.FeatureEnabled(models.FeatureJoinScreening) {
		if !cfg.JoinScreeningEnabled {
			add("join_screening", "warn", "Feature flag enabled but join screening toggle is off.")
		}
		if cfg.JoinScreeningAccountAgeDays <= 0 {
			add("join_screening", "error", "Join screening account age days must be greater than zero.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureRaidPanic) {
		if cfg.RaidPanicDefaultMinutes <= 0 {
			add("raid_panic", "error", "Default panic duration must be greater than zero.")
		}
		if cfg.RaidPanicSlowmodeSeconds <= 0 {
			add("raid_panic", "error", "Slowmode seconds must be greater than zero.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureStreaks) {
		if !cfg.StreaksEnabled {
			add("streaks", "warn", "Feature flag enabled but streaks toggle is off.")
		}
		if cfg.StreakRewardCoins <= 0 || cfg.StreakRewardXP <= 0 {
			add("streaks", "warn", "Streak rewards should be positive for both coins and XP.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureSeasonResets) {
		if !cfg.SeasonResetsEnabled {
			add("season_resets", "warn", "Feature flag enabled but season reset toggle is off.")
		}
		if cfg.SeasonResetCadence != "monthly" && cfg.SeasonResetCadence != "quarterly" {
			add("season_resets", "error", "Cadence must be monthly or quarterly.")
		}
		if len(cfg.SeasonResetModules) == 0 {
			add("season_resets", "warn", "No modules selected; defaults will be applied.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureAccountAgeGuard) {
		if cfg.AccountAgeMinDays <= 0 {
			add("account_age_guard", "error", "Enabled with invalid minimum account age.")
		}
		if cfg.AccountAgeAction != "log_only" && cfg.AccountAgeAction != "quarantine" && cfg.AccountAgeAction != "kick" {
			add("account_age_guard", "error", "Enabled with unsupported action.")
		}
	}
	if cfg.FeatureEnabled(models.FeatureAppeals) && cfg.AppealsChannelID == "" {
		add("appeals", "warn", "Enabled without appeals intake channel.")
	}
	if cfg.RetentionDays > 0 && cfg.RetentionDays < 7 {
		add("retention", "warn", "Retention window is under 7 days; verify compliance intent.")
	}
	if len(rows) == 0 {
		add("system", "ok", "No dependency issues detected.")
	}
	writeJSON(w, map[string]any{"checks": rows})
}
