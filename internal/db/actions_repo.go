package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type ActionsRepo struct {
	db *sql.DB
}

func (r *ActionsRepo) Enqueue(ctx context.Context, row models.ActionRow) (int64, error) {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, `INSERT INTO actions(
		guild_id, actor_user_id, target_user_id, type, payload_json, status, error, created_at, updated_at
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.GuildID, row.ActorUserID, row.TargetUserID, row.Type, row.PayloadJSON, "queued", "", now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ActionsRepo) List(ctx context.Context, guildID string, status string, limit, offset int) ([]models.ActionRow, error) {
	query := `SELECT id, guild_id, actor_user_id, target_user_id, type, payload_json, status, error, created_at, updated_at
		FROM actions WHERE guild_id = ?`
	args := []any{guildID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.ActionRow, 0)
	for rows.Next() {
		var row models.ActionRow
		var created, updated string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.ActorUserID, &row.TargetUserID, &row.Type, &row.PayloadJSON, &row.Status, &row.Error, &created, &updated); err != nil {
			return nil, err
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		row.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *ActionsRepo) Get(ctx context.Context, id int64) (models.ActionRow, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, guild_id, actor_user_id, target_user_id, type, payload_json, status, error, created_at, updated_at
		FROM actions WHERE id = ?`, id)

	var out models.ActionRow
	var created, updated string
	if err := row.Scan(&out.ID, &out.GuildID, &out.ActorUserID, &out.TargetUserID, &out.Type, &out.PayloadJSON, &out.Status, &out.Error, &created, &updated); err != nil {
		if err == sql.ErrNoRows {
			return models.ActionRow{}, false, nil
		}
		return models.ActionRow{}, false, err
	}
	out.CreatedAt, _ = time.Parse(time.RFC3339, created)
	out.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return out, true, nil
}

func (r *ActionsRepo) UpdateStatus(ctx context.Context, id int64, status, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE actions SET status = ?, error = ?, updated_at = ? WHERE id = ?`,
		status, errMsg, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (r *ActionsRepo) NextQueued(ctx context.Context) (models.ActionRow, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, guild_id, actor_user_id, target_user_id, type, payload_json, status, error, created_at, updated_at
		FROM actions WHERE status = 'queued' ORDER BY created_at ASC LIMIT 1`)

	var out models.ActionRow
	var created, updated string
	if err := row.Scan(&out.ID, &out.GuildID, &out.ActorUserID, &out.TargetUserID, &out.Type, &out.PayloadJSON, &out.Status, &out.Error, &created, &updated); err != nil {
		if err == sql.ErrNoRows {
			return models.ActionRow{}, false, nil
		}
		return models.ActionRow{}, false, err
	}
	out.CreatedAt, _ = time.Parse(time.RFC3339, created)
	out.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return out, true, nil
}

func (r *ActionsRepo) CountSince(ctx context.Context, guildID string, since time.Time, status string) (int, error) {
	query := `SELECT COUNT(*) FROM actions WHERE guild_id = ? AND created_at >= ?`
	args := []any{guildID, since.UTC().Format(time.RFC3339)}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	row := r.db.QueryRowContext(ctx, query, args...)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *ActionsRepo) CountByStatus(ctx context.Context, guildID, status string) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM actions WHERE guild_id = ? AND status = ?`, guildID, status)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
