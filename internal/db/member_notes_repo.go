package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type MemberNotesRepo struct {
	db *sql.DB
}

func (r *MemberNotesRepo) Create(ctx context.Context, row models.MemberNoteRow) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO member_notes(
		guild_id, user_id, author_id, body, created_at, resolved_at
	) VALUES(?, ?, ?, ?, ?, NULL)`,
		row.GuildID, row.UserID, row.AuthorID, row.Body, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *MemberNotesRepo) List(ctx context.Context, guildID, userID string, limit int) ([]models.MemberNoteRow, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, guild_id, user_id, author_id, body, created_at, resolved_at
		FROM member_notes
		WHERE guild_id = ?`
	args := []any{guildID}
	if userID != "" {
		query += " AND user_id = ?"
		args = append(args, userID)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.MemberNoteRow, 0)
	for rows.Next() {
		var row models.MemberNoteRow
		var created string
		var resolved sql.NullString
		if err := rows.Scan(&row.ID, &row.GuildID, &row.UserID, &row.AuthorID, &row.Body, &created, &resolved); err != nil {
			return nil, err
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		if resolved.Valid {
			t, _ := time.Parse(time.RFC3339, resolved.String)
			row.ResolvedAt = &t
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *MemberNotesRepo) Resolve(ctx context.Context, guildID string, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE member_notes SET resolved_at = ? WHERE guild_id = ? AND id = ?`,
		time.Now().UTC().Format(time.RFC3339), guildID, id,
	)
	return err
}
