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

func (r *TicketsRepo) AppendMessage(ctx context.Context, msg models.TicketMessageRow) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO ticket_messages(
		ticket_id, guild_id, channel_id, author_user_id, content, created_at
	) VALUES(?, ?, ?, ?, ?, ?)`,
		msg.TicketID, msg.GuildID, msg.ChannelID, msg.AuthorUserID, msg.Content, msg.CreatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *TicketsRepo) ListMessages(ctx context.Context, guildID string, ticketID int64, limit int) ([]models.TicketMessageRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, ticket_id, guild_id, channel_id, author_user_id, content, created_at
		FROM ticket_messages
		WHERE guild_id = ? AND ticket_id = ?
		ORDER BY created_at ASC
		LIMIT ?`, guildID, ticketID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.TicketMessageRow, 0)
	for rows.Next() {
		var item models.TicketMessageRow
		var created string
		if err := rows.Scan(&item.ID, &item.TicketID, &item.GuildID, &item.ChannelID, &item.AuthorUserID, &item.Content, &created); err != nil {
			return nil, err
		}
		item.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *TicketsRepo) OpenWithLastActivityBefore(ctx context.Context, guildID string, before time.Time, limit int) ([]models.TicketRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT t.id, t.guild_id, t.channel_id, t.creator_user_id, t.subject, t.status, t.created_at, t.closed_at
FROM tickets t
LEFT JOIN (
	SELECT ticket_id, MAX(created_at) AS last_message_at
	FROM ticket_messages
	GROUP BY ticket_id
) m ON m.ticket_id = t.id
WHERE t.guild_id = ? AND t.status = 'open' AND COALESCE(m.last_message_at, t.created_at) <= ?
ORDER BY COALESCE(m.last_message_at, t.created_at) ASC
LIMIT ?`, guildID, before.UTC().Format(time.RFC3339), limit)
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

func (r *TicketsRepo) CountByStatus(ctx context.Context, guildID, status string) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tickets WHERE guild_id = ? AND status = ?`, guildID, status)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *TicketsRepo) CountCreatedSince(ctx context.Context, guildID string, since time.Time) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tickets WHERE guild_id = ? AND created_at >= ?`,
		guildID, since.UTC().Format(time.RFC3339),
	)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
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
