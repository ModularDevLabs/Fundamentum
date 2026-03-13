package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type JoinScreeningRepo struct {
	db *sql.DB
}

func (r *JoinScreeningRepo) Create(ctx context.Context, row models.JoinScreeningRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	accountCreated := ""
	if row.AccountCreatedAt != nil {
		accountCreated = row.AccountCreatedAt.UTC().Format(time.RFC3339)
	}
	res, err := r.db.ExecContext(ctx, `INSERT INTO join_screening_queue(
		guild_id, user_id, username, account_created_at, reason, status, reviewed_by, created_at, reviewed_at
	) VALUES(?, ?, ?, ?, ?, ?, '', ?, NULL)`,
		row.GuildID, row.UserID, row.Username, accountCreated, row.Reason, row.Status, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *JoinScreeningRepo) ListByStatus(ctx context.Context, guildID, status string, limit int) ([]models.JoinScreeningRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, user_id, username, account_created_at, reason, status, reviewed_by, created_at, reviewed_at
		FROM join_screening_queue WHERE guild_id = ? AND status = ? ORDER BY created_at DESC LIMIT ?`, guildID, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.JoinScreeningRow, 0)
	for rows.Next() {
		var row models.JoinScreeningRow
		var accountCreated, created, reviewed sql.NullString
		if err := rows.Scan(&row.ID, &row.GuildID, &row.UserID, &row.Username, &accountCreated, &row.Reason, &row.Status, &row.ReviewedBy, &created, &reviewed); err != nil {
			return nil, err
		}
		if accountCreated.Valid && accountCreated.String != "" {
			t, _ := time.Parse(time.RFC3339, accountCreated.String)
			row.AccountCreatedAt = &t
		}
		if created.Valid {
			row.CreatedAt, _ = time.Parse(time.RFC3339, created.String)
		}
		if reviewed.Valid && reviewed.String != "" {
			t, _ := time.Parse(time.RFC3339, reviewed.String)
			row.ReviewedAt = &t
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *JoinScreeningRepo) GetByID(ctx context.Context, guildID string, id int64) (models.JoinScreeningRow, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, guild_id, user_id, username, account_created_at, reason, status, reviewed_by, created_at, reviewed_at
		FROM join_screening_queue WHERE guild_id = ? AND id = ?`, guildID, id)
	var out models.JoinScreeningRow
	var accountCreated, created, reviewed sql.NullString
	if err := row.Scan(&out.ID, &out.GuildID, &out.UserID, &out.Username, &accountCreated, &out.Reason, &out.Status, &out.ReviewedBy, &created, &reviewed); err != nil {
		if err == sql.ErrNoRows {
			return models.JoinScreeningRow{}, false, nil
		}
		return models.JoinScreeningRow{}, false, err
	}
	if accountCreated.Valid && accountCreated.String != "" {
		t, _ := time.Parse(time.RFC3339, accountCreated.String)
		out.AccountCreatedAt = &t
	}
	if created.Valid {
		out.CreatedAt, _ = time.Parse(time.RFC3339, created.String)
	}
	if reviewed.Valid && reviewed.String != "" {
		t, _ := time.Parse(time.RFC3339, reviewed.String)
		out.ReviewedAt = &t
	}
	return out, true, nil
}

func (r *JoinScreeningRepo) Review(ctx context.Context, guildID string, id int64, decision, reviewedBy string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `UPDATE join_screening_queue SET status = ?, reviewed_by = ?, reviewed_at = ? WHERE guild_id = ? AND id = ?`,
		decision, reviewedBy, now, guildID, id,
	)
	return err
}
