package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type GiveawaysRepo struct {
	db *sql.DB
}

func (r *GiveawaysRepo) Create(ctx context.Context, row models.GiveawayRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO giveaways(
		guild_id, channel_id, message_id, prize, winner_count, ends_at, status, created_at
	) VALUES(?, ?, ?, ?, ?, ?, 'open', ?)`,
		row.GuildID, row.ChannelID, row.MessageID, row.Prize, row.WinnerCount, row.EndsAt.UTC().Format(time.RFC3339), now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *GiveawaysRepo) AttachMessageID(ctx context.Context, id int64, messageID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE giveaways SET message_id = ? WHERE id = ?`, messageID, id)
	return err
}

func (r *GiveawaysRepo) ListByGuild(ctx context.Context, guildID string, limit int) ([]models.GiveawayRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `SELECT g.id, g.guild_id, g.channel_id, g.message_id, g.prize, g.winner_count, g.ends_at, g.status, g.created_at,
		(SELECT COUNT(*) FROM giveaway_entries e WHERE e.giveaway_id = g.id) AS entry_count
		FROM giveaways g
		WHERE g.guild_id = ?
		ORDER BY g.created_at DESC
		LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.GiveawayRow, 0)
	for rows.Next() {
		var row models.GiveawayRow
		var endsAt, created string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.ChannelID, &row.MessageID, &row.Prize, &row.WinnerCount, &endsAt, &row.Status, &created, &row.EntryCount); err != nil {
			return nil, err
		}
		row.EndsAt, _ = time.Parse(time.RFC3339, endsAt)
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *GiveawaysRepo) FindOpenByMessage(ctx context.Context, guildID, channelID, messageID string) (models.GiveawayRow, bool, error) {
	var row models.GiveawayRow
	var endsAt, created string
	err := r.db.QueryRowContext(ctx, `SELECT id, guild_id, channel_id, message_id, prize, winner_count, ends_at, status, created_at
		FROM giveaways
		WHERE guild_id = ? AND channel_id = ? AND message_id = ? AND status = 'open'
		LIMIT 1`, guildID, channelID, messageID).Scan(
		&row.ID, &row.GuildID, &row.ChannelID, &row.MessageID, &row.Prize, &row.WinnerCount, &endsAt, &row.Status, &created,
	)
	if err == sql.ErrNoRows {
		return models.GiveawayRow{}, false, nil
	}
	if err != nil {
		return models.GiveawayRow{}, false, err
	}
	row.EndsAt, _ = time.Parse(time.RFC3339, endsAt)
	row.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return row, true, nil
}

func (r *GiveawaysRepo) AddEntry(ctx context.Context, giveawayID int64, userID string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO giveaway_entries(giveaway_id, user_id, created_at)
		VALUES(?, ?, ?)
		ON CONFLICT(giveaway_id, user_id) DO NOTHING`, giveawayID, userID, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (r *GiveawaysRepo) ListEntries(ctx context.Context, giveawayID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT user_id FROM giveaway_entries WHERE giveaway_id = ? ORDER BY created_at ASC`, giveawayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		out = append(out, userID)
	}
	return out, rows.Err()
}

func (r *GiveawaysRepo) MarkEnded(ctx context.Context, giveawayID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE giveaways SET status = 'ended' WHERE id = ?`, giveawayID)
	return err
}

func (r *GiveawaysRepo) GetByID(ctx context.Context, guildID string, giveawayID int64) (models.GiveawayRow, bool, error) {
	var row models.GiveawayRow
	var endsAt, created string
	err := r.db.QueryRowContext(ctx, `SELECT id, guild_id, channel_id, message_id, prize, winner_count, ends_at, status, created_at,
		(SELECT COUNT(*) FROM giveaway_entries e WHERE e.giveaway_id = giveaways.id) AS entry_count
		FROM giveaways WHERE guild_id = ? AND id = ? LIMIT 1`, guildID, giveawayID).Scan(
		&row.ID, &row.GuildID, &row.ChannelID, &row.MessageID, &row.Prize, &row.WinnerCount, &endsAt, &row.Status, &created, &row.EntryCount,
	)
	if err == sql.ErrNoRows {
		return models.GiveawayRow{}, false, nil
	}
	if err != nil {
		return models.GiveawayRow{}, false, err
	}
	row.EndsAt, _ = time.Parse(time.RFC3339, endsAt)
	row.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return row, true, nil
}
