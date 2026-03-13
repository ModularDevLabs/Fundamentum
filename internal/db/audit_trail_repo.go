package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type AuditTrailRepo struct {
	db *sql.DB
}

func (r *AuditTrailRepo) Append(ctx context.Context, guildID, eventType, message string, payload map[string]any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	prevHash := ""
	_ = r.db.QueryRowContext(ctx, `SELECT event_hash FROM audit_trail_events WHERE guild_id = ? ORDER BY id DESC LIMIT 1`, guildID).Scan(&prevHash)
	recorded := time.Now().UTC().Format(time.RFC3339)
	seed := fmt.Sprintf("%s|%s|%s|%s|%s|%s", guildID, eventType, message, string(payloadBytes), prevHash, recorded)
	sum := sha256.Sum256([]byte(seed))
	eventHash := hex.EncodeToString(sum[:])
	_, err = r.db.ExecContext(ctx, `INSERT INTO audit_trail_events(guild_id, event_type, message, payload_json, prev_hash, event_hash, recorded_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)`,
		guildID, eventType, message, string(payloadBytes), prevHash, eventHash, recorded,
	)
	return err
}

func (r *AuditTrailRepo) ListByGuild(ctx context.Context, guildID string, limit int) ([]models.AuditTrailRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, event_type, message, payload_json, prev_hash, event_hash, recorded_at
		FROM audit_trail_events WHERE guild_id = ? ORDER BY id DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.AuditTrailRow, 0, limit)
	for rows.Next() {
		var row models.AuditTrailRow
		var recordedRaw string
		if err := rows.Scan(&row.ID, &row.GuildID, &row.EventType, &row.Message, &row.Payload, &row.PrevHash, &row.EventHash, &recordedRaw); err != nil {
			return nil, err
		}
		row.RecordedAt, _ = time.Parse(time.RFC3339, recordedRaw)
		out = append(out, row)
	}
	return out, rows.Err()
}
