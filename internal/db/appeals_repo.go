package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type AppealsRepo struct {
	db *sql.DB
}

func (r *AppealsRepo) Create(ctx context.Context, row models.AppealRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO appeals(
		guild_id, user_id, reason, status, resolution, reviewed_by, created_at, reviewed_at
	) VALUES(?, ?, ?, 'open', '', '', ?, NULL)`,
		row.GuildID, row.UserID, row.Reason, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *AppealsRepo) ListByGuild(ctx context.Context, guildID, status string, limit int) ([]models.AppealRow, error) {
	query := `SELECT id, guild_id, user_id, reason, status, resolution, reviewed_by, created_at, reviewed_at
		FROM appeals WHERE guild_id = ?`
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
	out := make([]models.AppealRow, 0)
	for rows.Next() {
		var item models.AppealRow
		var created string
		var reviewed sql.NullString
		if err := rows.Scan(&item.ID, &item.GuildID, &item.UserID, &item.Reason, &item.Status, &item.Resolution, &item.ReviewedBy, &created, &reviewed); err != nil {
			return nil, err
		}
		item.CreatedAt, _ = time.Parse(time.RFC3339, created)
		if reviewed.Valid {
			t, _ := time.Parse(time.RFC3339, reviewed.String)
			item.ReviewedAt = &t
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *AppealsRepo) Resolve(ctx context.Context, guildID string, id int64, actor, resolution string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE appeals
		SET status='resolved', resolution=?, reviewed_by=?, reviewed_at=?
		WHERE guild_id = ? AND id = ?`,
		resolution, actor, time.Now().UTC().Format(time.RFC3339), guildID, id,
	)
	return err
}
