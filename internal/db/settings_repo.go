package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
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

func applyGuildSettingDefaults(cfg models.GuildSettings) models.GuildSettings {
	def := models.DefaultGuildSettings(cfg.GuildID)
	if cfg.FeatureFlags == nil {
		cfg.FeatureFlags = map[string]bool{}
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
	return cfg
}
