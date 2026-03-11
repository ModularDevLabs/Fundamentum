package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type ScheduledMessagesRepo struct {
	db *sql.DB
}

func (r *ScheduledMessagesRepo) ListByGuild(ctx context.Context, guildID string) ([]models.ScheduledMessageRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, channel_id, content, interval_minutes, next_run_at, enabled, created_at, updated_at
		FROM scheduled_messages
		WHERE guild_id = ?
		ORDER BY id DESC`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.ScheduledMessageRow, 0)
	for rows.Next() {
		var item models.ScheduledMessageRow
		var nextRun, created, updated string
		var enabledInt int
		if err := rows.Scan(&item.ID, &item.GuildID, &item.ChannelID, &item.Content, &item.IntervalMinutes, &nextRun, &enabledInt, &created, &updated); err != nil {
			return nil, err
		}
		item.Enabled = enabledInt == 1
		item.NextRunAt, _ = time.Parse(time.RFC3339, nextRun)
		item.CreatedAt, _ = time.Parse(time.RFC3339, created)
		item.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ScheduledMessagesRepo) Create(ctx context.Context, row models.ScheduledMessageRow) (int64, error) {
	now := time.Now().UTC()
	enabled := 0
	if row.Enabled {
		enabled = 1
	}
	res, err := r.db.ExecContext(ctx, `INSERT INTO scheduled_messages(
		guild_id, channel_id, content, interval_minutes, next_run_at, enabled, created_at, updated_at
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		row.GuildID,
		row.ChannelID,
		row.Content,
		row.IntervalMinutes,
		row.NextRunAt.UTC().Format(time.RFC3339),
		enabled,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ScheduledMessagesRepo) Delete(ctx context.Context, guildID string, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM scheduled_messages WHERE guild_id = ? AND id = ?`, guildID, id)
	return err
}

func (r *ScheduledMessagesRepo) Due(ctx context.Context, now time.Time, limit int) ([]models.ScheduledMessageRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, channel_id, content, interval_minutes, next_run_at, enabled, created_at, updated_at
		FROM scheduled_messages
		WHERE enabled = 1 AND next_run_at <= ?
		ORDER BY next_run_at ASC
		LIMIT ?`, now.UTC().Format(time.RFC3339), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.ScheduledMessageRow, 0)
	for rows.Next() {
		var item models.ScheduledMessageRow
		var nextRun, created, updated string
		var enabledInt int
		if err := rows.Scan(&item.ID, &item.GuildID, &item.ChannelID, &item.Content, &item.IntervalMinutes, &nextRun, &enabledInt, &created, &updated); err != nil {
			return nil, err
		}
		item.Enabled = enabledInt == 1
		item.NextRunAt, _ = time.Parse(time.RFC3339, nextRun)
		item.CreatedAt, _ = time.Parse(time.RFC3339, created)
		item.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ScheduledMessagesRepo) MarkRan(ctx context.Context, id int64, next time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE scheduled_messages SET next_run_at = ?, updated_at = ? WHERE id = ?`,
		next.UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}
