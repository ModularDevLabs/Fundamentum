package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type SuggestionsRepo struct {
	db *sql.DB
}

func (r *SuggestionsRepo) Create(ctx context.Context, row models.SuggestionRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO suggestions(
		guild_id, user_id, content, message_id, channel_id, status, decision_by, decision_note, created_at, updated_at
	) VALUES(?, ?, ?, ?, ?, 'open', '', '', ?, NULL)`,
		row.GuildID, row.UserID, row.Content, row.MessageID, row.ChannelID, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *SuggestionsRepo) ListByGuild(ctx context.Context, guildID, status string, limit int) ([]models.SuggestionRow, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, guild_id, user_id, content, message_id, channel_id, status, decision_by, decision_note, created_at, updated_at
		FROM suggestions WHERE guild_id = ?`
	args := []any{guildID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.SuggestionRow, 0)
	for rows.Next() {
		var row models.SuggestionRow
		var created string
		var updated sql.NullString
		if err := rows.Scan(&row.ID, &row.GuildID, &row.UserID, &row.Content, &row.MessageID, &row.ChannelID, &row.Status, &row.DecisionBy, &row.DecisionNote, &created, &updated); err != nil {
			return nil, err
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		if updated.Valid {
			t, _ := time.Parse(time.RFC3339, updated.String)
			row.UpdatedAt = &t
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *SuggestionsRepo) Decide(ctx context.Context, guildID string, id int64, status, actor, note string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE suggestions
		SET status=?, decision_by=?, decision_note=?, updated_at=?
		WHERE guild_id=? AND id=?`,
		status, actor, note, time.Now().UTC().Format(time.RFC3339), guildID, id,
	)
	return err
}
