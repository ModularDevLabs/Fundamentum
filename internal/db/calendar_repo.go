package db

import (
	"context"
	"database/sql"
	"time"
)

type CalendarRepo struct {
	db *sql.DB
}

type CalendarEventRow struct {
	ID        int64  `json:"id"`
	GuildID   string `json:"guild_id"`
	Title     string `json:"title"`
	Details   string `json:"details"`
	StartAt   string `json:"start_at"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

type CalendarRSVPRow struct {
	EventID   int64  `json:"event_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
}

func (r *CalendarRepo) CreateEvent(ctx context.Context, row CalendarEventRow) (int64, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO calendar_events(guild_id, title, details, start_at, created_by, created_at)
		VALUES(?, ?, ?, ?, ?, ?)`,
		row.GuildID, row.Title, row.Details, row.StartAt, row.CreatedBy, created,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CalendarRepo) ListEvents(ctx context.Context, guildID string, limit int) ([]CalendarEventRow, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, title, details, start_at, created_by, created_at
		FROM calendar_events WHERE guild_id = ? ORDER BY start_at ASC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CalendarEventRow, 0, limit)
	for rows.Next() {
		var row CalendarEventRow
		if err := rows.Scan(&row.ID, &row.GuildID, &row.Title, &row.Details, &row.StartAt, &row.CreatedBy, &row.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *CalendarRepo) SetRSVP(ctx context.Context, eventID int64, userID, status string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO calendar_event_rsvps(event_id, user_id, status, updated_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(event_id, user_id) DO UPDATE SET status=excluded.status, updated_at=excluded.updated_at`,
		eventID, userID, status, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (r *CalendarRepo) ListRSVPs(ctx context.Context, eventID int64) ([]CalendarRSVPRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT event_id, user_id, status, updated_at FROM calendar_event_rsvps WHERE event_id = ? ORDER BY updated_at DESC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CalendarRSVPRow, 0, 32)
	for rows.Next() {
		var row CalendarRSVPRow
		if err := rows.Scan(&row.EventID, &row.UserID, &row.Status, &row.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
