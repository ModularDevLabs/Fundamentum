package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type StreaksRepo struct {
	db *sql.DB
}

func (r *StreaksRepo) UpsertDailyActivity(ctx context.Context, guildID, userID, day string) (models.StreakRow, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT current_streak, best_streak, last_active_date FROM member_streaks WHERE guild_id = ? AND user_id = ?`, guildID, userID)
	var current, best int
	var lastDate string
	newDay := false
	switch err := row.Scan(&current, &best, &lastDate); err {
	case nil:
		if lastDate == day {
			return models.StreakRow{
				GuildID:        guildID,
				UserID:         userID,
				CurrentStreak:  current,
				BestStreak:     best,
				LastActiveDate: lastDate,
				UpdatedAt:      time.Now().UTC(),
			}, false, nil
		}
		prev, _ := time.Parse("2006-01-02", lastDate)
		target, _ := time.Parse("2006-01-02", day)
		if target.Sub(prev) == 24*time.Hour {
			current++
		} else {
			current = 1
		}
		if current > best {
			best = current
		}
		newDay = true
	case sql.ErrNoRows:
		current = 1
		best = 1
		newDay = true
	default:
		return models.StreakRow{}, false, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO member_streaks(guild_id, user_id, current_streak, best_streak, last_active_date, updated_at)
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET current_streak=excluded.current_streak, best_streak=excluded.best_streak, last_active_date=excluded.last_active_date, updated_at=excluded.updated_at`,
		guildID, userID, current, best, day, now,
	)
	if err != nil {
		return models.StreakRow{}, false, err
	}
	updatedAt, _ := time.Parse(time.RFC3339, now)
	return models.StreakRow{
		GuildID:        guildID,
		UserID:         userID,
		CurrentStreak:  current,
		BestStreak:     best,
		LastActiveDate: day,
		UpdatedAt:      updatedAt,
	}, newDay, nil
}

func (r *StreaksRepo) Leaderboard(ctx context.Context, guildID string, limit int) ([]models.StreakRow, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `SELECT guild_id, user_id, current_streak, best_streak, last_active_date, updated_at
		FROM member_streaks WHERE guild_id = ? ORDER BY current_streak DESC, best_streak DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.StreakRow, 0)
	for rows.Next() {
		var row models.StreakRow
		var updated string
		if err := rows.Scan(&row.GuildID, &row.UserID, &row.CurrentStreak, &row.BestStreak, &row.LastActiveDate, &updated); err != nil {
			return nil, err
		}
		row.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *StreaksRepo) GetUser(ctx context.Context, guildID, userID string) (models.StreakRow, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT guild_id, user_id, current_streak, best_streak, last_active_date, updated_at
		FROM member_streaks WHERE guild_id = ? AND user_id = ?`, guildID, userID)
	var out models.StreakRow
	var updated string
	if err := row.Scan(&out.GuildID, &out.UserID, &out.CurrentStreak, &out.BestStreak, &out.LastActiveDate, &updated); err != nil {
		if err == sql.ErrNoRows {
			return models.StreakRow{}, false, nil
		}
		return models.StreakRow{}, false, err
	}
	out.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return out, true, nil
}
