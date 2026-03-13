package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type BirthdaysRepo struct {
	db *sql.DB
}

func (r *BirthdaysRepo) Upsert(ctx context.Context, guildID, userID, mmdd, tz string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO birthdays(guild_id, user_id, birthday_mmdd, timezone, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET birthday_mmdd=excluded.birthday_mmdd, timezone=excluded.timezone, updated_at=excluded.updated_at`,
		guildID, userID, mmdd, tz, now, now,
	)
	return err
}

func (r *BirthdaysRepo) Delete(ctx context.Context, guildID, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM birthdays WHERE guild_id = ? AND user_id = ?`, guildID, userID)
	return err
}

func (r *BirthdaysRepo) ListByGuild(ctx context.Context, guildID string, limit int) ([]models.BirthdayRow, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := r.db.QueryContext(ctx, `SELECT guild_id, user_id, birthday_mmdd, timezone, created_at, updated_at
		FROM birthdays WHERE guild_id = ? ORDER BY birthday_mmdd ASC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.BirthdayRow, 0)
	for rows.Next() {
		var row models.BirthdayRow
		var created, updated string
		if err := rows.Scan(&row.GuildID, &row.UserID, &row.BirthdayMMDD, &row.Timezone, &created, &updated); err != nil {
			return nil, err
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		row.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *BirthdaysRepo) ListByDate(ctx context.Context, guildID, mmdd string, limit int) ([]models.BirthdayRow, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, `SELECT guild_id, user_id, birthday_mmdd, timezone, created_at, updated_at
		FROM birthdays WHERE guild_id = ? AND birthday_mmdd = ? ORDER BY user_id ASC LIMIT ?`, guildID, mmdd, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.BirthdayRow, 0)
	for rows.Next() {
		var row models.BirthdayRow
		var created, updated string
		if err := rows.Scan(&row.GuildID, &row.UserID, &row.BirthdayMMDD, &row.Timezone, &created, &updated); err != nil {
			return nil, err
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		row.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, row)
	}
	return out, rows.Err()
}
