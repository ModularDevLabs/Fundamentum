package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type AFKRepo struct {
	db *sql.DB
}

func (r *AFKRepo) Set(ctx context.Context, guildID, userID, reason string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO afk_status(guild_id, user_id, reason, created_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET reason=excluded.reason, created_at=excluded.created_at`,
		guildID, userID, reason, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (r *AFKRepo) Clear(ctx context.Context, guildID, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM afk_status WHERE guild_id = ? AND user_id = ?`, guildID, userID)
	return err
}

func (r *AFKRepo) Get(ctx context.Context, guildID, userID string) (models.AFKStatusRow, bool, error) {
	var row models.AFKStatusRow
	var created string
	err := r.db.QueryRowContext(ctx, `SELECT guild_id, user_id, reason, created_at FROM afk_status WHERE guild_id = ? AND user_id = ?`,
		guildID, userID).Scan(&row.GuildID, &row.UserID, &row.Reason, &created)
	if err == sql.ErrNoRows {
		return models.AFKStatusRow{}, false, nil
	}
	if err != nil {
		return models.AFKStatusRow{}, false, err
	}
	row.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return row, true, nil
}
