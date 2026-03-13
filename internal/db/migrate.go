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
		`CREATE TABLE IF NOT EXISTS custom_commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			trigger TEXT NOT NULL,
			response TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_custom_commands_guild_trigger
		ON custom_commands(guild_id, trigger);`,
		`CREATE TABLE IF NOT EXISTS starboard_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			source_channel_id TEXT NOT NULL,
			source_message_id TEXT NOT NULL,
			starboard_channel_id TEXT NOT NULL,
			starboard_message_id TEXT NOT NULL,
			star_count INTEGER NOT NULL,
			last_updated_at TEXT NOT NULL,
			posted_at TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_starboard_unique_source
		ON starboard_entries(guild_id, source_channel_id, source_message_id);`,
		`CREATE TABLE IF NOT EXISTS member_levels (
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			username TEXT,
			xp INTEGER NOT NULL,
			level INTEGER NOT NULL,
			last_xp_at TEXT NOT NULL,
			PRIMARY KEY (guild_id, user_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_member_levels_guild_xp
		ON member_levels(guild_id, xp DESC);`,
		`CREATE TABLE IF NOT EXISTS giveaways (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			message_id TEXT NOT NULL,
			prize TEXT NOT NULL,
			winner_count INTEGER NOT NULL,
			ends_at TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_giveaways_guild_status
		ON giveaways(guild_id, status, ends_at);`,
		`CREATE TABLE IF NOT EXISTS giveaway_entries (
			giveaway_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			created_at TEXT NOT NULL,
			PRIMARY KEY (giveaway_id, user_id)
		);`,
		`CREATE TABLE IF NOT EXISTS polls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			message_id TEXT NOT NULL,
			question TEXT NOT NULL,
			options_json TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			closed_at TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_polls_guild_status
		ON polls(guild_id, status, created_at);`,
		`CREATE TABLE IF NOT EXISTS suggestions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			content TEXT NOT NULL,
			message_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			status TEXT NOT NULL,
			decision_by TEXT,
			decision_note TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_suggestions_guild_status
		ON suggestions(guild_id, status, created_at);`,
		`CREATE TABLE IF NOT EXISTS afk_status (
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			reason TEXT NOT NULL,
			created_at TEXT NOT NULL,
			PRIMARY KEY (guild_id, user_id)
		);`,
		`CREATE TABLE IF NOT EXISTS reminders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			content TEXT NOT NULL,
			run_at TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_reminders_due
		ON reminders(status, run_at);`,
		`CREATE TABLE IF NOT EXISTS member_notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			author_id TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at TEXT NOT NULL,
			resolved_at TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_member_notes_guild_user
		ON member_notes(guild_id, user_id, created_at);`,
		`CREATE TABLE IF NOT EXISTS webhook_integrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			url TEXT NOT NULL,
			events_json TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			last_error TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_webhooks_guild_enabled
		ON webhook_integrations(guild_id, enabled);`,
		`CREATE TABLE IF NOT EXISTS audit_trail_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			message TEXT NOT NULL,
			payload_json TEXT NOT NULL,
			prev_hash TEXT NOT NULL,
			event_hash TEXT NOT NULL,
			recorded_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_trail_guild_time
		ON audit_trail_events(guild_id, recorded_at DESC);`,
		`CREATE TABLE IF NOT EXISTS reputation_points (
			guild_id TEXT NOT NULL,
			from_user_id TEXT NOT NULL,
			to_user_id TEXT NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			last_given_at TEXT NOT NULL,
			PRIMARY KEY (guild_id, from_user_id, to_user_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_reputation_target
		ON reputation_points(guild_id, to_user_id);`,
		`CREATE TABLE IF NOT EXISTS economy_balances (
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			balance INTEGER NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (guild_id, user_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_economy_guild_balance
		ON economy_balances(guild_id, balance DESC);`,
		`CREATE TABLE IF NOT EXISTS shop_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			name TEXT NOT NULL,
			cost INTEGER NOT NULL,
			role_id TEXT,
			duration_minutes INTEGER NOT NULL DEFAULT 0,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_shop_items_guild
		ON shop_items(guild_id, enabled, cost);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
