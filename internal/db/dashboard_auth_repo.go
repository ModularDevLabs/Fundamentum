package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type DashboardUserRow struct {
	Username     string
	PasswordHash string
	Role         string
	Enabled      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}

type DashboardSessionRow struct {
	SessionID  string
	Username   string
	Role       string
	CSRFToken  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastSeenAt time.Time
	SourceIP   string
	UserAgent  string
	Revoked    bool
}

type DashboardAuthRepo struct {
	db *sql.DB
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func (r *DashboardAuthRepo) UpsertUser(ctx context.Context, row DashboardUserRow) error {
	now := time.Now().UTC()
	if row.CreatedAt.IsZero() {
		row.CreatedAt = now
	}
	if row.UpdatedAt.IsZero() {
		row.UpdatedAt = now
	}
	row.Username = normalizeUsername(row.Username)
	row.Role = normalizeRole(row.Role)
	if row.Username == "" || row.PasswordHash == "" || row.Role == "" {
		return fmt.Errorf("missing required user fields")
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO dashboard_users(username, password_hash, role, enabled, created_at, updated_at, last_login_at)
	VALUES(?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(username) DO UPDATE SET
		password_hash=excluded.password_hash,
		role=excluded.role,
		enabled=excluded.enabled,
		updated_at=excluded.updated_at,
		last_login_at=COALESCE(excluded.last_login_at, dashboard_users.last_login_at)`,
		row.Username,
		row.PasswordHash,
		row.Role,
		boolToInt(row.Enabled),
		row.CreatedAt.Format(time.RFC3339),
		row.UpdatedAt.Format(time.RFC3339),
		timePtrString(row.LastLoginAt),
	)
	return err
}

func (r *DashboardAuthRepo) GetUser(ctx context.Context, username string) (DashboardUserRow, error) {
	username = normalizeUsername(username)
	row := r.db.QueryRowContext(ctx, `SELECT username, password_hash, role, enabled, created_at, updated_at, last_login_at FROM dashboard_users WHERE username = ?`, username)
	var out DashboardUserRow
	var enabled int
	var createdRaw, updatedRaw string
	var lastLogin sql.NullString
	if err := row.Scan(&out.Username, &out.PasswordHash, &out.Role, &enabled, &createdRaw, &updatedRaw, &lastLogin); err != nil {
		return DashboardUserRow{}, err
	}
	created, err := time.Parse(time.RFC3339, createdRaw)
	if err != nil {
		return DashboardUserRow{}, err
	}
	updated, err := time.Parse(time.RFC3339, updatedRaw)
	if err != nil {
		return DashboardUserRow{}, err
	}
	out.CreatedAt = created
	out.UpdatedAt = updated
	out.Enabled = enabled == 1
	if lastLogin.Valid && lastLogin.String != "" {
		parsed, err := time.Parse(time.RFC3339, lastLogin.String)
		if err == nil {
			out.LastLoginAt = &parsed
		}
	}
	return out, nil
}

func (r *DashboardAuthRepo) ListUsers(ctx context.Context) ([]DashboardUserRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT username, password_hash, role, enabled, created_at, updated_at, last_login_at FROM dashboard_users ORDER BY username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]DashboardUserRow, 0, 16)
	for rows.Next() {
		var item DashboardUserRow
		var enabled int
		var createdRaw, updatedRaw string
		var lastLogin sql.NullString
		if err := rows.Scan(&item.Username, &item.PasswordHash, &item.Role, &enabled, &createdRaw, &updatedRaw, &lastLogin); err != nil {
			return nil, err
		}
		created, err := time.Parse(time.RFC3339, createdRaw)
		if err != nil {
			return nil, err
		}
		updated, err := time.Parse(time.RFC3339, updatedRaw)
		if err != nil {
			return nil, err
		}
		item.CreatedAt = created
		item.UpdatedAt = updated
		item.Enabled = enabled == 1
		if lastLogin.Valid && lastLogin.String != "" {
			parsed, err := time.Parse(time.RFC3339, lastLogin.String)
			if err == nil {
				item.LastLoginAt = &parsed
			}
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *DashboardAuthRepo) DeleteUser(ctx context.Context, username string) error {
	username = normalizeUsername(username)
	_, err := r.db.ExecContext(ctx, `DELETE FROM dashboard_users WHERE username = ?`, username)
	return err
}

func (r *DashboardAuthRepo) CountEnabledAdmins(ctx context.Context) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM dashboard_users WHERE role = 'admin' AND enabled = 1`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *DashboardAuthRepo) SetLastLogin(ctx context.Context, username string, at time.Time) error {
	username = normalizeUsername(username)
	_, err := r.db.ExecContext(ctx, `UPDATE dashboard_users SET last_login_at = ?, updated_at = ? WHERE username = ?`, at.UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339), username)
	return err
}

func (r *DashboardAuthRepo) CreateSession(ctx context.Context, row DashboardSessionRow) error {
	if row.SessionID == "" || row.Username == "" || row.Role == "" || row.CSRFToken == "" {
		return fmt.Errorf("missing required session fields")
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO dashboard_sessions(session_id, username, role, csrf_token, created_at, expires_at, last_seen_at, source_ip, user_agent, revoked)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		row.SessionID,
		normalizeUsername(row.Username),
		normalizeRole(row.Role),
		row.CSRFToken,
		row.CreatedAt.UTC().Format(time.RFC3339),
		row.ExpiresAt.UTC().Format(time.RFC3339),
		row.LastSeenAt.UTC().Format(time.RFC3339),
		strings.TrimSpace(row.SourceIP),
		strings.TrimSpace(row.UserAgent),
	)
	return err
}

func (r *DashboardAuthRepo) GetSession(ctx context.Context, sessionID string) (DashboardSessionRow, error) {
	row := r.db.QueryRowContext(ctx, `SELECT session_id, username, role, csrf_token, created_at, expires_at, last_seen_at, source_ip, user_agent, revoked
	FROM dashboard_sessions WHERE session_id = ?`, strings.TrimSpace(sessionID))
	var out DashboardSessionRow
	var createdRaw, expiresRaw, lastSeenRaw string
	var revoked int
	if err := row.Scan(&out.SessionID, &out.Username, &out.Role, &out.CSRFToken, &createdRaw, &expiresRaw, &lastSeenRaw, &out.SourceIP, &out.UserAgent, &revoked); err != nil {
		return DashboardSessionRow{}, err
	}
	createdAt, err := time.Parse(time.RFC3339, createdRaw)
	if err != nil {
		return DashboardSessionRow{}, err
	}
	expiresAt, err := time.Parse(time.RFC3339, expiresRaw)
	if err != nil {
		return DashboardSessionRow{}, err
	}
	lastSeenAt, err := time.Parse(time.RFC3339, lastSeenRaw)
	if err != nil {
		return DashboardSessionRow{}, err
	}
	out.CreatedAt = createdAt
	out.ExpiresAt = expiresAt
	out.LastSeenAt = lastSeenAt
	out.Revoked = revoked == 1
	return out, nil
}

func (r *DashboardAuthRepo) TouchSession(ctx context.Context, sessionID string, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE dashboard_sessions SET last_seen_at = ? WHERE session_id = ? AND revoked = 0`, now.UTC().Format(time.RFC3339), strings.TrimSpace(sessionID))
	return err
}

func (r *DashboardAuthRepo) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE dashboard_sessions SET revoked = 1 WHERE session_id = ?`, strings.TrimSpace(sessionID))
	return err
}

func (r *DashboardAuthRepo) RevokeSessionsForUser(ctx context.Context, username string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE dashboard_sessions SET revoked = 1 WHERE username = ?`, normalizeUsername(username))
	return err
}

func (r *DashboardAuthRepo) PurgeExpiredSessions(ctx context.Context, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM dashboard_sessions WHERE revoked = 1 OR expires_at < ?`, now.UTC().Format(time.RFC3339))
	return err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func timePtrString(v *time.Time) any {
	if v == nil {
		return nil
	}
	return v.UTC().Format(time.RFC3339)
}
