package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type RetentionRepo struct {
	db *sql.DB
}

func (r *RetentionRepo) PurgeGuildBefore(ctx context.Context, guildID string, cutoff time.Time) (map[string]int64, error) {
	cutoffTS := cutoff.UTC().Format(time.RFC3339)
	type plan struct {
		name  string
		query string
	}
	plans := []plan{
		{name: "warnings", query: `DELETE FROM warnings WHERE guild_id = ? AND created_at < ?`},
		{name: "ticket_messages", query: `DELETE FROM ticket_messages WHERE guild_id = ? AND created_at < ?`},
		{name: "appeals", query: `DELETE FROM appeals WHERE guild_id = ? AND created_at < ?`},
		{name: "suggestions", query: `DELETE FROM suggestions WHERE guild_id = ? AND created_at < ?`},
		{name: "member_notes", query: `DELETE FROM member_notes WHERE guild_id = ? AND created_at < ?`},
		{name: "reminders", query: `DELETE FROM reminders WHERE guild_id = ? AND created_at < ?`},
		{name: "actions", query: `DELETE FROM actions WHERE guild_id = ? AND created_at < ?`},
	}
	out := make(map[string]int64, len(plans))
	for _, item := range plans {
		res, err := r.db.ExecContext(ctx, item.query, guildID, cutoffTS)
		if err != nil {
			return nil, err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return nil, err
		}
		out[item.name] = affected
	}
	return out, nil
}

func (r *RetentionRepo) CountGuildBefore(ctx context.Context, guildID string, cutoff time.Time) (map[string]int64, error) {
	cutoffTS := cutoff.UTC().Format(time.RFC3339)
	type plan struct {
		name  string
		query string
	}
	plans := []plan{
		{name: "warnings", query: `SELECT COUNT(*) FROM warnings WHERE guild_id = ? AND created_at < ?`},
		{name: "ticket_messages", query: `SELECT COUNT(*) FROM ticket_messages WHERE guild_id = ? AND created_at < ?`},
		{name: "appeals", query: `SELECT COUNT(*) FROM appeals WHERE guild_id = ? AND created_at < ?`},
		{name: "suggestions", query: `SELECT COUNT(*) FROM suggestions WHERE guild_id = ? AND created_at < ?`},
		{name: "member_notes", query: `SELECT COUNT(*) FROM member_notes WHERE guild_id = ? AND created_at < ?`},
		{name: "reminders", query: `SELECT COUNT(*) FROM reminders WHERE guild_id = ? AND created_at < ?`},
		{name: "actions", query: `SELECT COUNT(*) FROM actions WHERE guild_id = ? AND created_at < ?`},
	}
	out := make(map[string]int64, len(plans))
	for _, item := range plans {
		var count int64
		if err := r.db.QueryRowContext(ctx, item.query, guildID, cutoffTS).Scan(&count); err != nil {
			return nil, err
		}
		out[item.name] = count
	}
	return out, nil
}

func (r *RetentionRepo) RecordArchiveEvent(ctx context.Context, guildID string, cutoff time.Time, rows map[string]int64) error {
	if rows == nil {
		rows = map[string]int64{}
	}
	payload, err := json.Marshal(map[string]any{
		"scope":      "retention_preview",
		"cutoff":     cutoff.UTC().Format(time.RFC3339),
		"row_counts": rows,
	})
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = r.db.ExecContext(ctx, `INSERT INTO actions(guild_id, actor_user_id, target_user_id, type, payload_json, status, error, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, 'success', '', ?, ?)`,
		guildID, "system:retention", "system:retention", "retention_archive", string(payload), now, now,
	)
	return err
}
