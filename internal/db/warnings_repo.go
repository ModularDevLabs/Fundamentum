package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type WarningsRepo struct {
	db *sql.DB
}

func (r *WarningsRepo) Create(ctx context.Context, row models.WarningRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO warnings(
		guild_id, user_id, actor_user_id, reason, created_at
	) VALUES(?, ?, ?, ?, ?)`, row.GuildID, row.UserID, row.ActorUserID, row.Reason, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *WarningsRepo) CountByUser(ctx context.Context, guildID, userID string) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM warnings WHERE guild_id = ? AND user_id = ?`, guildID, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *WarningsRepo) ListByGuild(ctx context.Context, guildID string, limit int) ([]models.WarningRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, user_id, actor_user_id, reason, created_at
		FROM warnings WHERE guild_id = ?
		ORDER BY created_at DESC
		LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.WarningRow, 0)
	for rows.Next() {
		var item models.WarningRow
		var created string
		if err := rows.Scan(&item.ID, &item.GuildID, &item.UserID, &item.ActorUserID, &item.Reason, &created); err != nil {
			return nil, err
		}
		item.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *WarningsRepo) CountSince(ctx context.Context, guildID string, since time.Time) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM warnings WHERE guild_id = ? AND created_at >= ?`,
		guildID, since.UTC().Format(time.RFC3339),
	)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
