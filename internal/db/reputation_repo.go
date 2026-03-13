package db

import (
	"context"
	"database/sql"
	"time"
)

type ReputationRepo struct {
	db *sql.DB
}

type ReputationLeaderboardRow struct {
	UserID string `json:"user_id"`
	Score  int    `json:"score"`
}

func (r *ReputationRepo) LastGivenAt(ctx context.Context, guildID, fromUserID, toUserID string) (time.Time, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT last_given_at FROM reputation_points WHERE guild_id = ? AND from_user_id = ? AND to_user_id = ?`,
		guildID, fromUserID, toUserID)
	var raw string
	if err := row.Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}
	ts, _ := time.Parse(time.RFC3339, raw)
	return ts, true, nil
}

func (r *ReputationRepo) AddDelta(ctx context.Context, guildID, fromUserID, toUserID string, delta int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO reputation_points(guild_id, from_user_id, to_user_id, score, last_given_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(guild_id, from_user_id, to_user_id) DO UPDATE SET
			score = reputation_points.score + excluded.score,
			last_given_at = excluded.last_given_at`,
		guildID, fromUserID, toUserID, delta, now,
	)
	return err
}

func (r *ReputationRepo) Leaderboard(ctx context.Context, guildID string, limit int) ([]ReputationLeaderboardRow, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `SELECT to_user_id, COALESCE(SUM(score), 0) AS total
		FROM reputation_points WHERE guild_id = ? GROUP BY to_user_id ORDER BY total DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ReputationLeaderboardRow, 0, limit)
	for rows.Next() {
		var row ReputationLeaderboardRow
		if err := rows.Scan(&row.UserID, &row.Score); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *ReputationRepo) TotalForUser(ctx context.Context, guildID, userID string) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(score), 0) FROM reputation_points WHERE guild_id = ? AND to_user_id = ?`, guildID, userID)
	var total int
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
