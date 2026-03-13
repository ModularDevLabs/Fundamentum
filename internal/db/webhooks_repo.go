package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type WebhooksRepo struct {
	db *sql.DB
}

func (r *WebhooksRepo) ListByGuild(ctx context.Context, guildID string) ([]models.WebhookIntegrationRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, url, events_json, enabled, last_error, created_at, updated_at
		FROM webhook_integrations WHERE guild_id = ? ORDER BY id DESC`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.WebhookIntegrationRow, 0)
	for rows.Next() {
		var row models.WebhookIntegrationRow
		var eventsRaw string
		var createdRaw string
		var updatedRaw string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.URL, &eventsRaw, &row.Enabled, &row.LastError, &createdRaw, &updatedRaw); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(eventsRaw), &row.Events)
		row.CreatedAt, _ = time.Parse(time.RFC3339, createdRaw)
		row.UpdatedAt, _ = time.Parse(time.RFC3339, updatedRaw)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *WebhooksRepo) ListEnabledByGuild(ctx context.Context, guildID string) ([]models.WebhookIntegrationRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, url, events_json, enabled, last_error, created_at, updated_at
		FROM webhook_integrations WHERE guild_id = ? AND enabled = 1 ORDER BY id DESC`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.WebhookIntegrationRow, 0)
	for rows.Next() {
		var row models.WebhookIntegrationRow
		var eventsRaw string
		var createdRaw string
		var updatedRaw string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.URL, &eventsRaw, &row.Enabled, &row.LastError, &createdRaw, &updatedRaw); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(eventsRaw), &row.Events)
		row.CreatedAt, _ = time.Parse(time.RFC3339, createdRaw)
		row.UpdatedAt, _ = time.Parse(time.RFC3339, updatedRaw)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *WebhooksRepo) Create(ctx context.Context, row models.WebhookIntegrationRow) (int64, error) {
	eventsJSON, _ := json.Marshal(row.Events)
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO webhook_integrations(guild_id, url, events_json, enabled, last_error, created_at, updated_at)
		VALUES(?, ?, ?, ?, '', ?, ?)`, row.GuildID, row.URL, string(eventsJSON), row.Enabled, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *WebhooksRepo) Delete(ctx context.Context, guildID string, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM webhook_integrations WHERE guild_id = ? AND id = ?`, guildID, id)
	return err
}

func (r *WebhooksRepo) SetLastError(ctx context.Context, id int64, errText string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE webhook_integrations SET last_error = ?, updated_at = ? WHERE id = ?`,
		errText, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}
