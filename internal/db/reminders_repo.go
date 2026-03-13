package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type RemindersRepo struct {
	db *sql.DB
}

func (r *RemindersRepo) Create(ctx context.Context, row models.ReminderRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO reminders(
		guild_id, channel_id, content, run_at, status, created_at
	) VALUES(?, ?, ?, ?, 'queued', ?)`,
		row.GuildID, row.ChannelID, row.Content, row.RunAt.UTC().Format(time.RFC3339), now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *RemindersRepo) ListByGuild(ctx context.Context, guildID string, limit int) ([]models.ReminderRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, channel_id, content, run_at, status, created_at
		FROM reminders WHERE guild_id = ? ORDER BY created_at DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.ReminderRow, 0)
	for rows.Next() {
		var row models.ReminderRow
		var runAt, created string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.ChannelID, &row.Content, &runAt, &row.Status, &created); err != nil {
			return nil, err
		}
		row.RunAt, _ = time.Parse(time.RFC3339, runAt)
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *RemindersRepo) ListDue(ctx context.Context, now time.Time, limit int) ([]models.ReminderRow, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, channel_id, content, run_at, status, created_at
		FROM reminders
		WHERE status = 'queued' AND run_at <= ?
		ORDER BY run_at ASC
		LIMIT ?`, now.UTC().Format(time.RFC3339), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.ReminderRow, 0)
	for rows.Next() {
		var row models.ReminderRow
		var runAt, created string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.ChannelID, &row.Content, &runAt, &row.Status, &created); err != nil {
			return nil, err
		}
		row.RunAt, _ = time.Parse(time.RFC3339, runAt)
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *RemindersRepo) MarkSent(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE reminders SET status = 'sent' WHERE id = ?`, id)
	return err
}
