package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type SettingsRepo struct {
	db *sql.DB
}

func (r *SettingsRepo) Get(ctx context.Context, guildID string) (models.GuildSettings, error) {
	row := r.db.QueryRowContext(ctx, `SELECT config_json FROM guild_settings WHERE guild_id = ?`, guildID)
	var raw string
	if err := row.Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return models.DefaultGuildSettings(guildID), nil
		}
		return models.GuildSettings{}, err
	}

	var cfg models.GuildSettings
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return models.GuildSettings{}, err
	}
	if cfg.GuildID == "" {
		cfg.GuildID = guildID
	}
	return applyGuildSettingDefaults(cfg), nil
}

func (r *SettingsRepo) Upsert(ctx context.Context, cfg models.GuildSettings) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO guild_settings(guild_id, config_json, updated_at)
	VALUES(?, ?, ?)
	ON CONFLICT(guild_id) DO UPDATE SET config_json=excluded.config_json, updated_at=excluded.updated_at`,
		cfg.GuildID, string(data), time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (r *SettingsRepo) EnsureDefaults(ctx context.Context, guildID string) (models.GuildSettings, error) {
	cfg, err := r.Get(ctx, guildID)
	if err != nil {
		return models.GuildSettings{}, err
	}

	if cfg.GuildID == "" {
		cfg = models.DefaultGuildSettings(guildID)
	}
	normalized := applyGuildSettingDefaults(cfg)
	if !reflect.DeepEqual(cfg, normalized) {
		cfg = normalized
		if err := r.Upsert(ctx, cfg); err != nil {
			return models.GuildSettings{}, err
		}
	}

	return cfg, nil
}

func (r *SettingsRepo) ListGuildIDs(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT guild_id FROM guild_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, 32)
	for rows.Next() {
		var guildID string
		if err := rows.Scan(&guildID); err != nil {
			return nil, err
		}
		if guildID == "" {
			continue
		}
		out = append(out, guildID)
	}
	return out, rows.Err()
}

func applyGuildSettingDefaults(cfg models.GuildSettings) models.GuildSettings {
	def := models.DefaultGuildSettings(cfg.GuildID)
	if cfg.FeatureFlags == nil {
		cfg.FeatureFlags = map[string]bool{}
	}
	if cfg.DashboardRolePolicies == nil {
		cfg.DashboardRolePolicies = map[string][]string{}
	}
	if cfg.ModuleChannelScopes == nil {
		cfg.ModuleChannelScopes = map[string][]string{}
	}
	for k, v := range def.FeatureFlags {
		if _, ok := cfg.FeatureFlags[k]; !ok {
			cfg.FeatureFlags[k] = v
		}
	}
	if cfg.WelcomeMessage == "" {
		cfg.WelcomeMessage = def.WelcomeMessage
	}
	if cfg.GoodbyeMessage == "" {
		cfg.GoodbyeMessage = def.GoodbyeMessage
	}
	if len(cfg.AuditLogEventTypes) == 0 {
		cfg.AuditLogEventTypes = append([]string{}, def.AuditLogEventTypes...)
	} else {
		seen := map[string]struct{}{}
		for _, t := range cfg.AuditLogEventTypes {
			seen[t] = struct{}{}
		}
		for _, t := range def.AuditLogEventTypes {
			if _, ok := seen[t]; ok {
				continue
			}
			cfg.AuditLogEventTypes = append(cfg.AuditLogEventTypes, t)
		}
	}
	if cfg.AutoModDupWindowSec <= 0 {
		cfg.AutoModDupWindowSec = def.AutoModDupWindowSec
	}
	if cfg.AutoModDupThreshold <= 0 {
		cfg.AutoModDupThreshold = def.AutoModDupThreshold
	}
	if cfg.AutoModAction == "" {
		cfg.AutoModAction = def.AutoModAction
	}
	if cfg.AutoModRules == nil {
		cfg.AutoModRules = []models.AutoModRule{}
	}
	if cfg.WarnQuarantineThreshold <= 0 {
		cfg.WarnQuarantineThreshold = def.WarnQuarantineThreshold
	}
	if cfg.WarnKickThreshold <= 0 {
		cfg.WarnKickThreshold = def.WarnKickThreshold
	}
	if cfg.VerificationPhrase == "" {
		cfg.VerificationPhrase = def.VerificationPhrase
	}
	if cfg.TicketOpenPhrase == "" {
		cfg.TicketOpenPhrase = def.TicketOpenPhrase
	}
	if cfg.TicketClosePhrase == "" {
		cfg.TicketClosePhrase = def.TicketClosePhrase
	}
	if cfg.TicketAutoCloseMinutes < 0 {
		cfg.TicketAutoCloseMinutes = 0
	}
	if cfg.AntiRaidJoinThreshold <= 0 {
		cfg.AntiRaidJoinThreshold = def.AntiRaidJoinThreshold
	}
	if cfg.AntiRaidWindowSeconds <= 0 {
		cfg.AntiRaidWindowSeconds = def.AntiRaidWindowSeconds
	}
	if cfg.AntiRaidCooldownMinutes <= 0 {
		cfg.AntiRaidCooldownMinutes = def.AntiRaidCooldownMinutes
	}
	if cfg.AntiRaidAction == "" {
		cfg.AntiRaidAction = def.AntiRaidAction
	}
	if cfg.AnalyticsIntervalDays <= 0 {
		cfg.AnalyticsIntervalDays = def.AnalyticsIntervalDays
	}
	if cfg.StarboardEmoji == "" {
		cfg.StarboardEmoji = def.StarboardEmoji
	}
	if cfg.StarboardThreshold <= 0 {
		cfg.StarboardThreshold = def.StarboardThreshold
	}
	if cfg.LevelingXPPerMessage <= 0 {
		cfg.LevelingXPPerMessage = def.LevelingXPPerMessage
	}
	if cfg.LevelingCooldownSeconds <= 0 {
		cfg.LevelingCooldownSeconds = def.LevelingCooldownSeconds
	}
	if cfg.LevelingCurve == "" {
		cfg.LevelingCurve = def.LevelingCurve
	}
	if cfg.LevelingCurve != "linear" && cfg.LevelingCurve != "quadratic" {
		cfg.LevelingCurve = def.LevelingCurve
	}
	if cfg.LevelingXPBase <= 0 {
		cfg.LevelingXPBase = def.LevelingXPBase
	}
	if cfg.GiveawaysReactionEmoji == "" {
		cfg.GiveawaysReactionEmoji = def.GiveawaysReactionEmoji
	}
	if cfg.KeywordAlertWords == nil {
		cfg.KeywordAlertWords = []string{}
	}
	if cfg.AFKSetPhrase == "" {
		cfg.AFKSetPhrase = def.AFKSetPhrase
	}
	if cfg.AccountAgeMinDays <= 0 {
		cfg.AccountAgeMinDays = def.AccountAgeMinDays
	}
	if cfg.AccountAgeAction == "" {
		cfg.AccountAgeAction = def.AccountAgeAction
	}
	if cfg.AppealsOpenPhrase == "" {
		cfg.AppealsOpenPhrase = def.AppealsOpenPhrase
	}
	if cfg.RetentionDays < 0 {
		cfg.RetentionDays = 0
	}
	if !cfg.RetentionArchiveBeforePurge && cfg.RetentionDays == 0 {
		cfg.RetentionArchiveBeforePurge = def.RetentionArchiveBeforePurge
	}
	if cfg.IncidentModeEnabled && cfg.IncidentModeEndsAt != "" {
		if t, err := time.Parse(time.RFC3339, cfg.IncidentModeEndsAt); err == nil && time.Now().UTC().After(t) {
			cfg.IncidentModeEnabled = false
		}
	}
	if cfg.MaintenanceWindowStart == "" {
		cfg.MaintenanceWindowStart = def.MaintenanceWindowStart
	}
	if cfg.MaintenanceWindowEnd == "" {
		cfg.MaintenanceWindowEnd = def.MaintenanceWindowEnd
	}
	if cfg.ModSummaryIntervalHours <= 0 {
		cfg.ModSummaryIntervalHours = def.ModSummaryIntervalHours
	}
	if cfg.AutoThreadKeywords == nil {
		cfg.AutoThreadKeywords = append([]string{}, def.AutoThreadKeywords...)
	}
	if cfg.VoiceRewardCoinsPerMinute <= 0 {
		cfg.VoiceRewardCoinsPerMinute = def.VoiceRewardCoinsPerMinute
	}
	if cfg.VoiceRewardXPPerMinute <= 0 {
		cfg.VoiceRewardXPPerMinute = def.VoiceRewardXPPerMinute
	}
	if cfg.JoinScreeningAccountAgeDays <= 0 {
		cfg.JoinScreeningAccountAgeDays = def.JoinScreeningAccountAgeDays
	}
	if cfg.RaidPanicDefaultMinutes <= 0 {
		cfg.RaidPanicDefaultMinutes = def.RaidPanicDefaultMinutes
	}
	if cfg.RaidPanicSlowmodeSeconds <= 0 {
		cfg.RaidPanicSlowmodeSeconds = def.RaidPanicSlowmodeSeconds
	}
	if cfg.StreakRewardCoins <= 0 {
		cfg.StreakRewardCoins = def.StreakRewardCoins
	}
	if cfg.StreakRewardXP <= 0 {
		cfg.StreakRewardXP = def.StreakRewardXP
	}
	if cfg.SeasonResetCadence != "monthly" && cfg.SeasonResetCadence != "quarterly" {
		cfg.SeasonResetCadence = def.SeasonResetCadence
	}
	filteredSeasonModules := make([]string, 0, len(cfg.SeasonResetModules))
	seenSeasonModules := map[string]struct{}{}
	for _, raw := range cfg.SeasonResetModules {
		module := strings.ToLower(strings.TrimSpace(raw))
		if module != "leveling" && module != "economy" && module != "trivia" {
			continue
		}
		if _, ok := seenSeasonModules[module]; ok {
			continue
		}
		seenSeasonModules[module] = struct{}{}
		filteredSeasonModules = append(filteredSeasonModules, module)
	}
	if len(filteredSeasonModules) == 0 {
		cfg.SeasonResetModules = append([]string{}, def.SeasonResetModules...)
	} else {
		cfg.SeasonResetModules = filteredSeasonModules
	}
	return cfg
}
