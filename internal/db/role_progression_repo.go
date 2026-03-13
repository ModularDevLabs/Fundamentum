package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type RoleProgressionRepo struct {
	db *sql.DB
}

func (r *RoleProgressionRepo) ListByGuild(ctx context.Context, guildID string) ([]models.RoleProgressionRuleRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, metric, threshold, role_id, enabled, created_at
		FROM role_progression_rules WHERE guild_id = ? ORDER BY metric ASC, threshold ASC, id ASC`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.RoleProgressionRuleRow, 0)
	for rows.Next() {
		var row models.RoleProgressionRuleRow
		var enabledInt int
		var created string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.Metric, &row.Threshold, &row.RoleID, &enabledInt, &created); err != nil {
			return nil, err
		}
		row.Enabled = enabledInt == 1
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *RoleProgressionRepo) Create(ctx context.Context, row models.RoleProgressionRuleRow) (int64, error) {
	enabled := 0
	if row.Enabled {
		enabled = 1
	}
	res, err := r.db.ExecContext(ctx, `INSERT INTO role_progression_rules(guild_id, metric, threshold, role_id, enabled, created_at)
		VALUES(?, ?, ?, ?, ?, ?)`,
		row.GuildID, row.Metric, row.Threshold, row.RoleID, enabled, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *RoleProgressionRepo) Delete(ctx context.Context, guildID string, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM role_progression_rules WHERE guild_id = ? AND id = ?`, guildID, id)
	return err
}
