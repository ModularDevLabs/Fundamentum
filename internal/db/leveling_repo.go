package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type LevelingRepo struct {
	db *sql.DB
}

func XPForLevel(level int, curve string, base int) int {
	if level <= 0 {
		return 0
	}
	if base <= 0 {
		base = 100
	}
	switch curve {
	case "linear":
		return level * base
	default:
		return level * level * base
	}
}

func LevelForXP(xp int, curve string, base int) int {
	level := 0
	for XPForLevel(level+1, curve, base) <= xp {
		level++
	}
	return level
}

func (r *LevelingRepo) AddXPIfDue(ctx context.Context, guildID, userID, username string, addXP, cooldownSec int, curve string, base int) (models.MemberLevelRow, bool, error) {
	now := time.Now().UTC()
	row, found, err := r.GetMember(ctx, guildID, userID)
	if err != nil {
		return models.MemberLevelRow{}, false, err
	}
	if !found {
		row = models.MemberLevelRow{
			GuildID:  guildID,
			UserID:   userID,
			Username: username,
			XP:       0,
			Level:    0,
			LastXPAt: time.Time{},
		}
	}

	if cooldownSec > 0 && !row.LastXPAt.IsZero() && now.Sub(row.LastXPAt) < time.Duration(cooldownSec)*time.Second {
		return row, false, nil
	}

	prevLevel := row.Level
	row.XP += addXP
	if row.XP < 0 {
		row.XP = 0
	}
	row.Level = LevelForXP(row.XP, curve, base)
	row.LastXPAt = now
	row.Username = username

	_, err = r.db.ExecContext(ctx, `INSERT INTO member_levels(guild_id, user_id, username, xp, level, last_xp_at)
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET
			username=excluded.username,
			xp=excluded.xp,
			level=excluded.level,
			last_xp_at=excluded.last_xp_at`,
		row.GuildID, row.UserID, row.Username, row.XP, row.Level, row.LastXPAt.Format(time.RFC3339),
	)
	if err != nil {
		return models.MemberLevelRow{}, false, err
	}

	return row, row.Level > prevLevel, nil
}

func (r *LevelingRepo) GetMember(ctx context.Context, guildID, userID string) (models.MemberLevelRow, bool, error) {
	var row models.MemberLevelRow
	var last string
	err := r.db.QueryRowContext(ctx, `SELECT guild_id, user_id, username, xp, level, last_xp_at
		FROM member_levels WHERE guild_id = ? AND user_id = ?`,
		guildID, userID,
	).Scan(&row.GuildID, &row.UserID, &row.Username, &row.XP, &row.Level, &last)
	if err == sql.ErrNoRows {
		return models.MemberLevelRow{}, false, nil
	}
	if err != nil {
		return models.MemberLevelRow{}, false, err
	}
	row.LastXPAt, _ = time.Parse(time.RFC3339, last)
	return row, true, nil
}

func (r *LevelingRepo) TopByGuild(ctx context.Context, guildID string, limit int) ([]models.MemberLevelRow, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `SELECT guild_id, user_id, username, xp, level, last_xp_at
		FROM member_levels
		WHERE guild_id = ?
		ORDER BY xp DESC, last_xp_at ASC
		LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.MemberLevelRow, 0)
	for rows.Next() {
		var row models.MemberLevelRow
		var last string
		if err := rows.Scan(&row.GuildID, &row.UserID, &row.Username, &row.XP, &row.Level, &last); err != nil {
			return nil, err
		}
		row.LastXPAt, _ = time.Parse(time.RFC3339, last)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *LevelingRepo) ResetGuild(ctx context.Context, guildID string) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM member_levels WHERE guild_id = ?`, guildID)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("leveling reset rows affected: %w", err)
	}
	return n, nil
}
