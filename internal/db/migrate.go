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
		`CREATE TABLE IF NOT EXISTS reaction_role_rule_constraints (
			rule_id INTEGER PRIMARY KEY,
			group_key TEXT NOT NULL DEFAULT '',
			max_select INTEGER NOT NULL DEFAULT 0,
			min_select INTEGER NOT NULL DEFAULT 0
		);`,
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
		`CREATE TABLE IF NOT EXISTS achievements (
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			badge_key TEXT NOT NULL,
			badge_name TEXT NOT NULL,
			awarded_at TEXT NOT NULL,
			meta_json TEXT NOT NULL,
			PRIMARY KEY (guild_id, user_id, badge_key)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_achievements_user
		ON achievements(guild_id, user_id, awarded_at DESC);`,
		`CREATE TABLE IF NOT EXISTS calendar_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			title TEXT NOT NULL,
			details TEXT,
			start_at TEXT NOT NULL,
			created_by TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_calendar_events_guild_start
		ON calendar_events(guild_id, start_at);`,
		`CREATE TABLE IF NOT EXISTS calendar_event_rsvps (
			event_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			status TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (event_id, user_id)
		);`,
		`CREATE TABLE IF NOT EXISTS role_rentals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role_id TEXT NOT NULL,
			started_at TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			status TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_role_rentals_due
		ON role_rentals(status, expires_at);`,
		`CREATE TABLE IF NOT EXISTS confessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL,
			posted_message_id TEXT,
			created_at TEXT NOT NULL,
			reviewed_at TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_confessions_guild_status
		ON confessions(guild_id, status, created_at DESC);`,
		`CREATE TABLE IF NOT EXISTS birthdays (
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			birthday_mmdd TEXT NOT NULL,
			timezone TEXT NOT NULL DEFAULT 'UTC',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (guild_id, user_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_birthdays_mmdd
		ON birthdays(guild_id, birthday_mmdd);`,
		`CREATE TABLE IF NOT EXISTS role_progression_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			metric TEXT NOT NULL,
			threshold INTEGER NOT NULL,
			role_id TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_role_progression_rules_guild
		ON role_progression_rules(guild_id, metric, threshold);`,
		`CREATE TABLE IF NOT EXISTS join_screening_queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			username TEXT NOT NULL,
			account_created_at TEXT,
			reason TEXT NOT NULL,
			status TEXT NOT NULL,
			reviewed_by TEXT,
			created_at TEXT NOT NULL,
			reviewed_at TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_join_screening_guild_status
		ON join_screening_queue(guild_id, status, created_at DESC);`,
		`CREATE TABLE IF NOT EXISTS raid_panic_lockdowns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			status TEXT NOT NULL,
			slowmode_seconds INTEGER NOT NULL,
			started_by TEXT NOT NULL,
			started_at TEXT NOT NULL,
			ends_at TEXT NOT NULL,
			ended_at TEXT,
			end_reason TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_raid_panic_guild_status
		ON raid_panic_lockdowns(guild_id, status, ends_at);`,
		`CREATE TABLE IF NOT EXISTS raid_panic_channel_states (
			lockdown_id INTEGER NOT NULL,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			previous_slowmode_seconds INTEGER NOT NULL,
			PRIMARY KEY(lockdown_id, channel_id)
		);`,
		`CREATE TABLE IF NOT EXISTS member_streaks (
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			current_streak INTEGER NOT NULL,
			best_streak INTEGER NOT NULL,
			last_active_date TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY(guild_id, user_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_member_streaks_guild_best
		ON member_streaks(guild_id, best_streak DESC, current_streak DESC);`,
		`CREATE TABLE IF NOT EXISTS trivia_scores (
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (guild_id, user_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_trivia_scores_guild
		ON trivia_scores(guild_id, score DESC);`,
		`CREATE TABLE IF NOT EXISTS season_reset_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			triggered_by TEXT NOT NULL,
			modules_json TEXT NOT NULL,
			affected_rows_json TEXT NOT NULL,
			status TEXT NOT NULL,
			error TEXT,
			started_at TEXT NOT NULL,
			completed_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_season_reset_runs_guild
		ON season_reset_runs(guild_id, started_at DESC);`,
		`CREATE TABLE IF NOT EXISTS dashboard_users (
			username TEXT PRIMARY KEY,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			last_login_at TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_dashboard_users_role
		ON dashboard_users(role, enabled);`,
		`CREATE TABLE IF NOT EXISTS dashboard_sessions (
			session_id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			role TEXT NOT NULL,
			csrf_token TEXT NOT NULL,
			created_at TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			last_seen_at TEXT NOT NULL,
			source_ip TEXT,
			user_agent TEXT,
			revoked INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE INDEX IF NOT EXISTS idx_dashboard_sessions_expiry
		ON dashboard_sessions(expires_at, revoked);`,
		`CREATE TABLE IF NOT EXISTS web3_first_scans (
			guild_id TEXT NOT NULL,
			asset_key TEXT NOT NULL,
			asset_type TEXT NOT NULL,
			display_symbol TEXT,
			display_name TEXT,
			first_scanner_user_id TEXT NOT NULL,
			first_scanner_name TEXT,
			first_price_usd REAL NOT NULL DEFAULT 0,
			first_scanned_at TEXT NOT NULL,
			PRIMARY KEY (guild_id, asset_key)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_web3_first_scans_guild_time
		ON web3_first_scans(guild_id, first_scanned_at DESC);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
