package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type PollsRepo struct {
	db *sql.DB
}

func (r *PollsRepo) Create(ctx context.Context, row models.PollRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	optionsJSON, _ := json.Marshal(row.Options)
	res, err := r.db.ExecContext(ctx, `INSERT INTO polls(
		guild_id, channel_id, message_id, question, options_json, status, created_at, closed_at
	) VALUES(?, ?, ?, ?, ?, 'open', ?, NULL)`,
		row.GuildID, row.ChannelID, row.MessageID, row.Question, string(optionsJSON), now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *PollsRepo) AttachMessageID(ctx context.Context, id int64, messageID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE polls SET message_id = ? WHERE id = ?`, messageID, id)
	return err
}

func (r *PollsRepo) ListByGuild(ctx context.Context, guildID string, limit int) ([]models.PollRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, channel_id, message_id, question, options_json, status, created_at, closed_at
		FROM polls WHERE guild_id = ? ORDER BY created_at DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.PollRow, 0)
	for rows.Next() {
		var row models.PollRow
		var optionsJSON string
		var created string
		var closed sql.NullString
		if err := rows.Scan(&row.ID, &row.GuildID, &row.ChannelID, &row.MessageID, &row.Question, &optionsJSON, &row.Status, &created, &closed); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(optionsJSON), &row.Options)
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		if closed.Valid {
			t, _ := time.Parse(time.RFC3339, closed.String)
			row.ClosedAt = &t
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *PollsRepo) GetByID(ctx context.Context, guildID string, id int64) (models.PollRow, bool, error) {
	var row models.PollRow
	var optionsJSON string
	var created string
	var closed sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id, guild_id, channel_id, message_id, question, options_json, status, created_at, closed_at
		FROM polls WHERE guild_id = ? AND id = ? LIMIT 1`, guildID, id).
		Scan(&row.ID, &row.GuildID, &row.ChannelID, &row.MessageID, &row.Question, &optionsJSON, &row.Status, &created, &closed)
	if err == sql.ErrNoRows {
		return models.PollRow{}, false, nil
	}
	if err != nil {
		return models.PollRow{}, false, err
	}
	_ = json.Unmarshal([]byte(optionsJSON), &row.Options)
	row.CreatedAt, _ = time.Parse(time.RFC3339, created)
	if closed.Valid {
		t, _ := time.Parse(time.RFC3339, closed.String)
		row.ClosedAt = &t
	}
	return row, true, nil
}

func (r *PollsRepo) MarkClosed(ctx context.Context, guildID string, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE polls SET status = 'closed', closed_at = ? WHERE guild_id = ? AND id = ?`,
		time.Now().UTC().Format(time.RFC3339), guildID, id)
	return err
}
