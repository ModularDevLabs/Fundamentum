package db

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type ActivityRepo struct {
	db *sql.DB
}

func (r *ActivityRepo) UpsertActivity(ctx context.Context, guildID, userID, channelID string, ts time.Time, username, globalName, displayName string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO activity(
		guild_id, user_id, last_message_at, last_channel_id, username, global_name, display_name
	) VALUES(?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(guild_id, user_id) DO UPDATE SET
		last_message_at=excluded.last_message_at,
		last_channel_id=excluded.last_channel_id,
		username=excluded.username,
		global_name=excluded.global_name,
		display_name=excluded.display_name`,
		guildID, userID, ts.UTC().Format(time.RFC3339), channelID, username, globalName, displayName,
	)
	return err
}

func (r *ActivityRepo) UpsertActivityIfStale(ctx context.Context, guildID, userID, channelID string, ts time.Time, username, globalName, displayName string, cutoff time.Time) (bool, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO activity(
		guild_id, user_id, last_message_at, last_channel_id, username, global_name, display_name
	) VALUES(?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(guild_id, user_id) DO UPDATE SET
		last_message_at=excluded.last_message_at,
		last_channel_id=excluded.last_channel_id,
		username=excluded.username,
		global_name=excluded.global_name,
		display_name=excluded.display_name
	WHERE last_message_at < ?`,
		guildID, userID, ts.UTC().Format(time.RFC3339), channelID, username, globalName, displayName,
		cutoff.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *ActivityRepo) ListMembers(ctx context.Context, guildID string, limit, offset int, search string) ([]models.MemberRow, error) {
	query := `SELECT guild_id, user_id, last_message_at, last_channel_id, username, global_name, display_name
		FROM activity WHERE guild_id = ?`
	args := []any{guildID}

	if search != "" {
		query += " AND (user_id LIKE ? OR username LIKE ? OR global_name LIKE ? OR display_name LIKE ?)"
		like := "%" + strings.ToLower(search) + "%"
		args = append(args, like, like, like, like)
	}

	query += " ORDER BY last_message_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.MemberRow, 0)
	for rows.Next() {
		var row models.MemberRow
		var lastMessage string
		if err := rows.Scan(&row.GuildID, &row.UserID, &lastMessage, &row.LastChannelID, &row.Username, &row.GlobalName, &row.DisplayName); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, lastMessage)
		if err == nil {
			row.LastMessageAt = &t
		}
		out = append(out, row)
	}

	return out, rows.Err()
}

func (r *ActivityRepo) GetMember(ctx context.Context, guildID, userID string) (models.MemberRow, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT guild_id, user_id, last_message_at, last_channel_id, username, global_name, display_name
		FROM activity WHERE guild_id = ? AND user_id = ?`, guildID, userID)

	var out models.MemberRow
	var lastMessage string
	if err := row.Scan(&out.GuildID, &out.UserID, &lastMessage, &out.LastChannelID, &out.Username, &out.GlobalName, &out.DisplayName); err != nil {
		if err == sql.ErrNoRows {
			return models.MemberRow{}, false, nil
		}
		return models.MemberRow{}, false, err
	}
	if t, err := time.Parse(time.RFC3339, lastMessage); err == nil {
		out.LastMessageAt = &t
	}
	return out, true, nil
}

func (r *ActivityRepo) ActiveUsersSince(ctx context.Context, guildID string, since time.Time) (map[string]struct{}, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT user_id FROM activity WHERE guild_id = ? AND last_message_at >= ?`,
		guildID, since.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]struct{})
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		out[userID] = struct{}{}
	}
	return out, rows.Err()
}

func (r *ActivityRepo) ListMembersAll(ctx context.Context, guildID string) ([]models.MemberRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT guild_id, user_id, last_message_at, last_channel_id, username, global_name, display_name
		FROM activity WHERE guild_id = ?`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.MemberRow, 0)
	for rows.Next() {
		var row models.MemberRow
		var lastMessage string
		if err := rows.Scan(&row.GuildID, &row.UserID, &lastMessage, &row.LastChannelID, &row.Username, &row.GlobalName, &row.DisplayName); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, lastMessage)
		if err == nil {
			row.LastMessageAt = &t
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *ActivityRepo) CountTracked(ctx context.Context, guildID string) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity WHERE guild_id = ?`, guildID)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *ActivityRepo) CountInactiveBefore(ctx context.Context, guildID string, cutoff time.Time) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity WHERE guild_id = ? AND last_message_at < ?`,
		guildID, cutoff.UTC().Format(time.RFC3339),
	)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
