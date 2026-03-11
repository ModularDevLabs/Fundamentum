package db

import "database/sql"

func Migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS activity (
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			last_message_at TEXT NOT NULL,
			last_channel_id TEXT,
			username TEXT,
			global_name TEXT,
			display_name TEXT,
			PRIMARY KEY (guild_id, user_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_activity_guild_last
		ON activity(guild_id, last_message_at);`,
		`CREATE TABLE IF NOT EXISTS guild_settings (
			guild_id TEXT PRIMARY KEY,
			config_json TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS actions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			actor_user_id TEXT NOT NULL,
			target_user_id TEXT NOT NULL,
			type TEXT NOT NULL,
			payload_json TEXT NOT NULL,
			status TEXT NOT NULL,
			error TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_actions_guild_status
		ON actions(guild_id, status, created_at);`,
		`CREATE TABLE IF NOT EXISTS backfill_state (
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			last_scanned_message_id TEXT,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (guild_id, channel_id)
		);`,
		`CREATE TABLE IF NOT EXISTS reaction_role_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			message_id TEXT NOT NULL,
			emoji TEXT NOT NULL,
			role_id TEXT NOT NULL,
			remove_on_unreact INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_reaction_rules_guild
		ON reaction_role_rules(guild_id, message_id);`,
		`CREATE TABLE IF NOT EXISTS warnings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			actor_user_id TEXT NOT NULL,
			reason TEXT,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_warnings_guild_user
		ON warnings(guild_id, user_id, created_at);`,
		`CREATE TABLE IF NOT EXISTS scheduled_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			content TEXT NOT NULL,
			interval_minutes INTEGER NOT NULL,
			next_run_at TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_scheduled_due
		ON scheduled_messages(enabled, next_run_at);`,
		`CREATE TABLE IF NOT EXISTS tickets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			creator_user_id TEXT NOT NULL,
			subject TEXT,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			closed_at TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_guild_status
		ON tickets(guild_id, status, created_at);`,
		`CREATE TABLE IF NOT EXISTS ticket_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ticket_id INTEGER NOT NULL,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			author_user_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket
		ON ticket_messages(ticket_id, created_at);`,
		`CREATE TABLE IF NOT EXISTS appeals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			reason TEXT NOT NULL,
			status TEXT NOT NULL,
			resolution TEXT,
			reviewed_by TEXT,
			created_at TEXT NOT NULL,
			reviewed_at TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_appeals_guild_status
		ON appeals(guild_id, status, created_at);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
