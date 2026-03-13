package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type AchievementsRepo struct {
	db *sql.DB
}

type AchievementRow struct {
	BadgeKey  string `json:"badge_key"`
	BadgeName string `json:"badge_name"`
	AwardedAt string `json:"awarded_at"`
	MetaJSON  string `json:"meta_json"`
}

func (r *AchievementsRepo) AwardIfMissing(ctx context.Context, guildID, userID, key, name string, meta map[string]any) error {
	metaRaw, _ := json.Marshal(meta)
	_, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO achievements(guild_id, user_id, badge_key, badge_name, awarded_at, meta_json)
		VALUES(?, ?, ?, ?, ?, ?)`,
		guildID, userID, key, name, time.Now().UTC().Format(time.RFC3339), string(metaRaw),
	)
	return err
}

func (r *AchievementsRepo) ListByUser(ctx context.Context, guildID, userID string, limit int) ([]AchievementRow, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `SELECT badge_key, badge_name, awarded_at, meta_json
		FROM achievements WHERE guild_id = ? AND user_id = ? ORDER BY awarded_at DESC LIMIT ?`, guildID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AchievementRow, 0, limit)
	for rows.Next() {
		var row AchievementRow
		if err := rows.Scan(&row.BadgeKey, &row.BadgeName, &row.AwardedAt, &row.MetaJSON); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
