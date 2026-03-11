package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type TicketsRepo struct {
	db *sql.DB
}

func (r *TicketsRepo) Create(ctx context.Context, row models.TicketRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO tickets(
		guild_id, channel_id, creator_user_id, subject, status, created_at, closed_at
	) VALUES(?, ?, ?, ?, ?, ?, NULL)`,
		row.GuildID, row.ChannelID, row.CreatorUserID, row.Subject, "open", now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TicketsRepo) ListByGuild(ctx context.Context, guildID string, status string, limit int) ([]models.TicketRow, error) {
	query := `SELECT id, guild_id, channel_id, creator_user_id, subject, status, created_at, closed_at
		FROM tickets WHERE guild_id = ?`
	args := []any{guildID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.TicketRow, 0)
	for rows.Next() {
		var item models.TicketRow
		var created string
		var closed sql.NullString
		if err := rows.Scan(&item.ID, &item.GuildID, &item.ChannelID, &item.CreatorUserID, &item.Subject, &item.Status, &created, &closed); err != nil {
			return nil, err
		}
		item.CreatedAt, _ = time.Parse(time.RFC3339, created)
		if closed.Valid {
			t, _ := time.Parse(time.RFC3339, closed.String)
			item.ClosedAt = &t
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *TicketsRepo) GetByChannel(ctx context.Context, guildID, channelID string) (models.TicketRow, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, guild_id, channel_id, creator_user_id, subject, status, created_at, closed_at
		FROM tickets WHERE guild_id = ? AND channel_id = ? ORDER BY id DESC LIMIT 1`, guildID, channelID)
	var item models.TicketRow
	var created string
	var closed sql.NullString
	if err := row.Scan(&item.ID, &item.GuildID, &item.ChannelID, &item.CreatorUserID, &item.Subject, &item.Status, &created, &closed); err != nil {
		if err == sql.ErrNoRows {
			return models.TicketRow{}, false, nil
		}
		return models.TicketRow{}, false, err
	}
	item.CreatedAt, _ = time.Parse(time.RFC3339, created)
	if closed.Valid {
		t, _ := time.Parse(time.RFC3339, closed.String)
		item.ClosedAt = &t
	}
	return item, true, nil
}

func (r *TicketsRepo) Close(ctx context.Context, guildID string, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tickets SET status='closed', closed_at=?, created_at=created_at WHERE guild_id = ? AND id = ?`,
		time.Now().UTC().Format(time.RFC3339), guildID, id,
	)
	return err
}

func (r *TicketsRepo) GetByID(ctx context.Context, guildID string, id int64) (models.TicketRow, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, guild_id, channel_id, creator_user_id, subject, status, created_at, closed_at
		FROM tickets WHERE guild_id = ? AND id = ?`, guildID, id)
	var item models.TicketRow
	var created string
	var closed sql.NullString
	if err := row.Scan(&item.ID, &item.GuildID, &item.ChannelID, &item.CreatorUserID, &item.Subject, &item.Status, &created, &closed); err != nil {
		if err == sql.ErrNoRows {
			return models.TicketRow{}, false, nil
		}
		return models.TicketRow{}, false, err
	}
	item.CreatedAt, _ = time.Parse(time.RFC3339, created)
	if closed.Valid {
		t, _ := time.Parse(time.RFC3339, closed.String)
		item.ClosedAt = &t
	}
	return item, true, nil
}
