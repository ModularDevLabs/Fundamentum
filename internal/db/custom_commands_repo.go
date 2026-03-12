package db

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type CustomCommandsRepo struct {
	db *sql.DB
}

func (r *CustomCommandsRepo) ListByGuild(ctx context.Context, guildID string) ([]models.CustomCommandRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, trigger, response, created_at
		FROM custom_commands
		WHERE guild_id = ?
		ORDER BY trigger ASC`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.CustomCommandRow, 0)
	for rows.Next() {
		var row models.CustomCommandRow
		var created string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.Trigger, &row.Response, &created); err != nil {
			return nil, err
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *CustomCommandsRepo) Create(ctx context.Context, row models.CustomCommandRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO custom_commands(
		guild_id, trigger, response, created_at
	) VALUES(?, ?, ?, ?)`,
		row.GuildID, strings.TrimSpace(strings.ToLower(row.Trigger)), strings.TrimSpace(row.Response), now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CustomCommandsRepo) Delete(ctx context.Context, guildID string, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM custom_commands WHERE guild_id = ? AND id = ?`, guildID, id)
	return err
}

func (r *CustomCommandsRepo) DeleteAllByGuild(ctx context.Context, guildID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM custom_commands WHERE guild_id = ?`, guildID)
	return err
}

func (r *CustomCommandsRepo) FindByTrigger(ctx context.Context, guildID, trigger string) (models.CustomCommandRow, bool, error) {
	var row models.CustomCommandRow
	var created string
	err := r.db.QueryRowContext(ctx, `SELECT id, guild_id, trigger, response, created_at
		FROM custom_commands
		WHERE guild_id = ? AND trigger = ? LIMIT 1`,
		guildID, strings.TrimSpace(strings.ToLower(trigger)),
	).Scan(&row.ID, &row.GuildID, &row.Trigger, &row.Response, &created)
	if err == sql.ErrNoRows {
		return models.CustomCommandRow{}, false, nil
	}
	if err != nil {
		return models.CustomCommandRow{}, false, err
	}
	row.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return row, true, nil
}
