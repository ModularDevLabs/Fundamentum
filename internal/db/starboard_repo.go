package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type StarboardRepo struct {
	db *sql.DB
}

func (r *StarboardRepo) GetBySource(ctx context.Context, guildID, sourceChannelID, sourceMessageID string) (models.StarboardEntryRow, bool, error) {
	var row models.StarboardEntryRow
	var updated string
	var posted sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id, guild_id, source_channel_id, source_message_id, starboard_channel_id, starboard_message_id, star_count, last_updated_at, posted_at
		FROM starboard_entries
		WHERE guild_id = ? AND source_channel_id = ? AND source_message_id = ? LIMIT 1`,
		guildID, sourceChannelID, sourceMessageID,
	).Scan(&row.ID, &row.GuildID, &row.SourceChannelID, &row.SourceMessageID, &row.StarboardChannel, &row.StarboardMessage, &row.StarCount, &updated, &posted)
	if err == sql.ErrNoRows {
		return models.StarboardEntryRow{}, false, nil
	}
	if err != nil {
		return models.StarboardEntryRow{}, false, err
	}
	row.LastUpdatedAt, _ = time.Parse(time.RFC3339, updated)
	if posted.Valid {
		t, _ := time.Parse(time.RFC3339, posted.String)
		row.PostedAt = &t
	}
	return row, true, nil
}

func (r *StarboardRepo) Upsert(ctx context.Context, row models.StarboardEntryRow) error {
	now := time.Now().UTC().Format(time.RFC3339)
	postedAt := sql.NullString{}
	if row.PostedAt != nil {
		postedAt = sql.NullString{String: row.PostedAt.UTC().Format(time.RFC3339), Valid: true}
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO starboard_entries(
		guild_id, source_channel_id, source_message_id, starboard_channel_id, starboard_message_id, star_count, last_updated_at, posted_at
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(guild_id, source_channel_id, source_message_id) DO UPDATE SET
		starboard_channel_id=excluded.starboard_channel_id,
		starboard_message_id=excluded.starboard_message_id,
		star_count=excluded.star_count,
		last_updated_at=excluded.last_updated_at,
		posted_at=excluded.posted_at`,
		row.GuildID, row.SourceChannelID, row.SourceMessageID, row.StarboardChannel, row.StarboardMessage, row.StarCount, now, postedAt,
	)
	return err
}
