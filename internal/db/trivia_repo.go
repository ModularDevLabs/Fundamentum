package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TriviaRepo struct {
	db *sql.DB
}

type TriviaScoreRow struct {
	UserID string `json:"user_id"`
	Score  int    `json:"score"`
}

func (r *TriviaRepo) AddScore(ctx context.Context, guildID, userID string, delta int) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO trivia_scores(guild_id, user_id, score, updated_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET score=trivia_scores.score + excluded.score, updated_at=excluded.updated_at`,
		guildID, userID, delta, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (r *TriviaRepo) Leaderboard(ctx context.Context, guildID string, limit int) ([]TriviaScoreRow, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `SELECT user_id, score FROM trivia_scores WHERE guild_id = ? ORDER BY score DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]TriviaScoreRow, 0, limit)
	for rows.Next() {
		var row TriviaScoreRow
		if err := rows.Scan(&row.UserID, &row.Score); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *TriviaRepo) ResetScores(ctx context.Context, guildID string) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM trivia_scores WHERE guild_id = ?`, guildID)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("trivia reset rows affected: %w", err)
	}
	return n, nil
}
