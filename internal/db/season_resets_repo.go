package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type SeasonResetsRepo struct {
	db *sql.DB
}

func (r *SeasonResetsRepo) RecordRun(ctx context.Context, guildID, triggeredBy string, modules []string, affectedRows map[string]int64, status, runError string, startedAt, completedAt time.Time) error {
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	if completedAt.IsZero() {
		completedAt = startedAt
	}
	modulesData, err := json.Marshal(modules)
	if err != nil {
		return err
	}
	affectedData, err := json.Marshal(affectedRows)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO season_reset_runs(guild_id, triggered_by, modules_json, affected_rows_json, status, error, started_at, completed_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		guildID,
		triggeredBy,
		string(modulesData),
		string(affectedData),
		status,
		runError,
		startedAt.UTC().Format(time.RFC3339),
		completedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *SeasonResetsRepo) ListRuns(ctx context.Context, guildID string, limit int) ([]models.SeasonResetRunRow, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, triggered_by, modules_json, affected_rows_json, status, error, started_at, completed_at
		FROM season_reset_runs
		WHERE guild_id = ?
		ORDER BY started_at DESC
		LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.SeasonResetRunRow, 0, limit)
	for rows.Next() {
		var row models.SeasonResetRunRow
		var modulesRaw, affectedRaw string
		var started, completed string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.TriggeredBy, &modulesRaw, &affectedRaw, &row.Status, &row.Error, &started, &completed); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(modulesRaw), &row.Modules)
		_ = json.Unmarshal([]byte(affectedRaw), &row.AffectedRows)
		row.StartedAt, _ = time.Parse(time.RFC3339, started)
		row.CompletedAt, _ = time.Parse(time.RFC3339, completed)
		out = append(out, row)
	}
	return out, rows.Err()
}
