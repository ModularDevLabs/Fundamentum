# Module Guide

This document explains how each GoBot feature module behaves at runtime, how to configure it, and how modules interact.

For field-by-field definitions and defaults, see `docs/SETTINGS.md`.

## How Modules Are Enabled

- All feature modules are controlled by `feature_flags` in guild settings.
- Every module defaults to `disabled` on a new guild.
- Most modules no-op immediately when disabled.
- Optional per-module channel scoping can be applied with `module_channel_scopes`.
- Core tracking/action systems (activity tracking, backfill, action worker, quarantine provisioning) are always available because other modules depend on them.

## Core Systems

### Activity Tracking (core)

- Trigger: every non-bot `MESSAGE_CREATE` in a guild.
- Behavior:
- Runs module handlers (appeals, suggestions, tickets, verification, leveling, keyword alerts, AFK, custom commands/automod).
- Upserts `activity` only when stale relative to inactive cutoff (`inactive_days`), reducing write pressure.
- Data written: `last_message_at`, `last_channel_id`, username/display snapshots.
- Config keys: `inactive_days`.

### Backfill (core)

- Trigger: dashboard `POST /api/backfill/start`.
- Behavior:
- Scans channels for historical messages and seeds/reconciles activity data.
- Effective scan window is `max(backfill_days, inactive_days)`.
- Uses per-channel saved state so reruns continue efficiently.
- Marks inaccessible channels as skipped, not hard failure.
- Status values: `queued`, `running`, `success`, `partial`, `failed`.
- Config keys: `backfill_days`, `backfill_concurrency`, `backfill_include_types`.

### Action Queue (core)

- Trigger: dashboard actions, warnings thresholds, automod quarantine mode, anti-raid quarantine mode, account-age guard actions.
- Behavior:
- Queues actions as `quarantine`, `kick`, or `remove_roles`.
- Worker executes oldest queued action, updates `running/success/failed`, and emits audit events for success/failure.
- Enforces `admin_user_policy` for administrator targets.
- Config keys: `admin_user_policy`, quarantine-related settings.

### Quarantine Assets + Permissions (core)

- Trigger: startup/guild provisioning and quarantine execution.
- Behavior:
- Ensures role `Quarantined` and channel `quarantine-readme` exist (or uses configured IDs).
- Ensures readme channel overwrite (view/read allowed, send denied for quarantine role).
- If `safe_quarantine_mode=false`, attempts guild-wide deny-view overwrites for quarantine role (best effort).
- During quarantine action, assigns quarantine role and removes other roles except allowlist.
- Config keys: `quarantine_role_id`, `readme_channel_id`, `allowlist_role_ids`, `safe_quarantine_mode`.

## Feature Modules

### Message Retention Tooling (settings-driven)

- Trigger: background worker tick every 6 hours.
- Behavior:
- Reads guild `retention_days`.
- If `retention_days > 0`, computes cutoff (`now - retention_days`).
- Optionally records `retention_archive` action summary with row counts.
- Purges old records from `warnings`, `ticket_messages`, `appeals`, `suggestions`, `member_notes`, `reminders`, and `actions`.
- Config keys: `retention_days`, `retention_archive_before_purge`.

### Incident Mode Switch (settings-driven)

- Trigger: enabled from Settings view.
- Behavior:
- Shows an incident banner in dashboard.
- Adds stricter safety checks for destructive actions:
- Confirm token required.
- Distinct approver required for `kick` and `remove_roles`.
- Optional auto-disable timestamp can end incident mode automatically.
- Config keys: `incident_mode_enabled`, `incident_mode_reason`, `incident_mode_ends_at`.

### Policy Simulator (dashboard tool)

- Trigger: Actions view -> `Policy Simulator`.
- Behavior:
- Runs action preflight for each listed user.
- Shows whether action is allowed and what additional controls apply.
- Includes confirm/approver requirements derived from current safety and incident settings.
- API: `POST /api/policy/simulate?guild_id=...`.

### Dependency Checker (dashboard tool)

- Trigger: Settings view -> `Dependency Checker`.
- Behavior:
- Validates enabled modules against required channels/roles/phrases and basic numeric constraints.
- Returns `error`, `warn`, or `ok` checks with clear remediation messages.
- API: `GET /api/dependencies/check?guild_id=...`.

### Health Dashboard (dashboard tool)

- Trigger: Overview view auto-refresh and manual `Refresh health`.
- Behavior:
- Surfaces queue depth, running actions, failed actions (24h), warnings/tickets (24h), and active backfills.
- Shows current safety posture (`incident_mode`, `action_dry_run`, two-person approval) and retention state.
- API: `GET /api/health/dashboard?guild_id=...`.

### Webhook Integrations (dashboard tool)

- Trigger: Settings view -> `Webhook Integrations`.
- Behavior:
- Stores per-guild outbound webhook URLs with event subscriptions.
- On matching audit/action events, POSTs JSON payload to each enabled integration.
- Tracks last delivery error for each webhook in the dashboard.
- APIs:
- `GET /api/integrations/webhooks?guild_id=...`
- `POST /api/integrations/webhooks?guild_id=...`
- `DELETE /api/integrations/webhooks/{id}?guild_id=...`

### Immutable Audit Trail Option

- Trigger: `immutable_audit_trail=true`.
- Behavior:
- Appends emitted events into `audit_trail_events` with hash chaining (`prev_hash` -> `event_hash`).
- Provides tamper-evident chronological records separate from normal action/event tables.
- API: `GET /api/audit-trail?guild_id=...&limit=...`.

### Maintenance Windows (settings-driven)

- Trigger: `maintenance_window_enabled=true` and current UTC time inside configured window.
- Behavior:
- Blocks new destructive dashboard actions.
- Pauses scheduled messages, reminders, ticket auto-close, analytics posting, and retention purge passes.
- Configuration: `maintenance_window_start` and `maintenance_window_end` (`HH:MM` UTC).

### Review Queue (dashboard tool)

- Trigger: `review_queue_enabled=true` and destructive action request.
- Behavior:
- New destructive actions are stored as `review_pending`.
- Reviewers approve (moves to `queued`) or reject (moves to `failed` with reason).
- API: `GET/POST /api/review-queue?guild_id=...`.

### Reputation System

- Supports `+rep @user` and `-rep @user` chat commands.
- Enforces per giver/target cooldown (12h) to reduce abuse.
- Dashboard endpoints:
- `POST /api/modules/reputation/give?guild_id=...`
- `GET /api/modules/reputation/leaderboard?guild_id=...`

### Economy + Shop

- Members earn coins from chat activity (1 coin, 60s per-user cooldown).
- Tracks balances and leaderboard per guild.
- Supports shop items with optional role grants on purchase.
- APIs:
- `GET /api/modules/economy/balance?guild_id=...&user_id=...`
- `GET /api/modules/economy/leaderboard?guild_id=...`
- `GET/POST /api/modules/economy/shop?guild_id=...`
- `POST /api/modules/economy/purchase?guild_id=...`

### Achievement Engine

- Persists earned badges per user.
- Awards badges from milestone checks (level, reputation, economy balance).
- API: `GET /api/modules/achievements?guild_id=...&user_id=...`.

### Mod Summaries

- Generates periodic moderation digest messages (warnings/actions/tickets).
- Configurable via settings: `mod_summary_channel_id`, `mod_summary_interval_hours`.
- Supports on-demand generation endpoint:
- `POST /api/mod-summary/generate?guild_id=...&hours=24`

### Auto Thread Helper

- Watches a configured channel for keyword-matching questions.
- Auto-creates a public thread from matching messages.
- Settings: `auto_thread_enabled`, `auto_thread_channel_id`, `auto_thread_keywords`.

### Voice Activity Rewards

- Tracks voice join/leave sessions.
- Awards coins and XP based on session minutes.
- Settings:
- `voice_rewards_enabled`
- `voice_reward_coins_per_minute`
- `voice_reward_xp_per_minute`

### Welcome Messages (`welcome_messages`)

- Trigger: member joins.
- Behavior: posts templated welcome message in configured channel.
- Template tokens: `{user}`, `{server}`.
- Config keys: `welcome_channel_id`, `welcome_message`.

### Goodbye Messages (`goodbye_messages`)

- Trigger: member leaves.
- Behavior: posts templated goodbye message in configured channel.
- Template tokens: `{user}`, `{server}`.
- Config keys: `goodbye_channel_id`, `goodbye_message`.

### Audit Log Stream (`audit_log_stream`)

- Trigger: selected bot-observed events (ban/role/channel events, action queue outcomes, automod action, anti-raid trigger).
- Behavior: sends formatted event lines to audit channel if event type is allowlisted.
- Config keys: `audit_log_channel_id`, `audit_log_event_types`.

### Invite Tracker (`invite_tracker`)

- Trigger: member joins and invite create/delete updates.
- Behavior:
- Maintains per-guild invite-use cache.
- On join, compares current invite uses vs cache and posts inferred invite source.
- First join after startup may report source unknown while cache warms.
- Config keys: `invite_log_channel_id`.
- Notes: requires `Manage Server` permission to read invite usage.

### AutoMod (`automod`)

- Trigger: message create when custom command did not match.
- Behavior:
- Checks link blocking, blocked words, and duplicate-spam window.
- Deletes violating message.
- Action mode:
- `delete_warn`: posts channel notice.
- `delete_only`: no further action.
- `delete_quarantine`: enqueues quarantine action.
- Supports channel and role ignore lists.
- Config keys: `automod_block_links`, `automod_blocked_words`, `automod_dup_window_sec`, `automod_dup_threshold`, `automod_action`, `automod_ignore_channel_ids`, `automod_ignore_role_ids`.

### Reaction Roles (`reaction_roles`)

- Trigger: reaction add/remove events.
- Behavior:
- Matches `(channel_id, message_id, emoji)` rules.
- Add reaction => add role.
- Remove reaction => remove role only when `remove_on_unreact=true`.
- Config keys: feature flag only; rules managed via reaction-role rule records.

### Warnings (`warnings`)

- Trigger: dashboard warning issue endpoint.
- Behavior:
- Creates warning record.
- Computes user warning count.
- Auto-enqueues `quarantine` or `kick` when threshold reached.
- Optionally logs issuance to warning log channel.
- Config keys: `warning_log_channel_id`, `warn_quarantine_threshold`, `warn_kick_threshold`.

### Scheduled Messages (`scheduled_messages`)

- Trigger: background worker every 30 seconds.
- Behavior:
- Sends due recurring messages.
- Advances `next_run_at` by interval (minimum 1 minute).
- Config keys: feature flag only; schedules store `channel_id`, `content`, `interval_minutes`, `enabled`.

### Verification (`verification`)

- Trigger:
- On join: assigns `unverified_role_id` and prompts in verification channel.
- On message in verification channel: phrase match.
- Behavior:
- Exact phrase (case-insensitive) removes unverified role.
- Optionally adds `verified_role_id`.
- Deletes phrase message and posts success confirmation.
- Config keys: `verification_channel_id`, `verification_phrase`, `unverified_role_id`, `verified_role_id`.

### Tickets (`tickets`)

- Trigger:
- Inbox channel message starting with `ticket_open_phrase` (default `!ticket`).
- Ticket channel message equal to `ticket_close_phrase` (default `!close`) from creator or support role.
- Auto-close worker every minute for inactive tickets.
- Behavior:
- Creates private ticket channel with permission overwrites.
- Stores transcript messages for dashboard/export.
- Close path logs closure, sends transcript to log channel, then deletes ticket channel.
- Config keys: `ticket_inbox_channel_id`, `ticket_category_id`, `ticket_support_role_id`, `ticket_log_channel_id`, `ticket_open_phrase`, `ticket_close_phrase`, `ticket_auto_close_minutes`.

### Anti-Raid (`anti_raid`)

- Trigger: member joins.
- Behavior:
- Tracks join timestamps in configured sliding window.
- If threshold reached, activates cooldown and emits alert/audit message.
- While active:
- `verification_only`: applies unverified role when configured.
- `quarantine`: enqueues quarantine action.
- Config keys: `anti_raid_join_threshold`, `anti_raid_window_seconds`, `anti_raid_cooldown_minutes`, `anti_raid_action`, `anti_raid_alert_channel_id`.

### Analytics (`analytics`)

- Trigger: background worker tick every 6 hours.
- Behavior:
- Sends summary once configured interval has elapsed for each guild.
- Includes tracked/inactive users, warning counts, actions, and ticket metrics.
- Config keys: `analytics_channel_id`, `analytics_interval_days`, `inactive_days` (used in inactive count).

### Starboard (`starboard`)

- Trigger: reaction add/remove.
- Behavior:
- Watches configured emoji on non-starboard channels.
- When threshold first reached, posts starboard entry.
- Subsequent reaction changes update existing starboard message and stored count.
- Ignores bot-authored source messages.
- Config keys: `starboard_channel_id`, `starboard_emoji`, `starboard_threshold`.

### Leveling (`leveling`)

- Trigger: message create.
- Behavior:
- Awards XP with per-user cooldown.
- On level-up, posts announcement in configured channel or source channel.
- Config keys: `leveling_channel_id`, `leveling_xp_per_message`, `leveling_cooldown_seconds`.

### Giveaways (`giveaways`)

- Trigger:
- Dashboard starts giveaway (creates message + DB row).
- Reaction add on giveaway message.
- Dashboard draw action.
- Behavior:
- Tracks unique entries from configured reaction emoji before end time.
- Draw shuffles entrants and chooses `winner_count` unique winners.
- Marks giveaway ended and announces winners.
- Config keys: `giveaways_channel_id`, `giveaways_reaction_emoji`.

### Polls (`polls`)

- Trigger:
- Dashboard starts poll.
- Dashboard close action.
- Behavior:
- Creates poll message with 2-5 options using `1️⃣`..`5️⃣` reactions.
- On close, reads reaction counts, subtracts bot seed reaction, posts results summary, marks poll closed.
- Config keys: `polls_channel_id`.

### Suggestions (`suggestions`)

- Trigger: message in suggestions channel.
- Behavior:
- Reposts message as formatted suggestion card.
- Adds 👍 and 👎 reactions.
- Stores suggestion record and deletes original message.
- Dashboard approve/reject sets status and optional note; decision can be sent to log channel.
- Config keys: `suggestions_channel_id`, `suggestions_log_channel_id`.

### Keyword Alerts (`keyword_alerts`)

- Trigger: message create.
- Behavior:
- Case-insensitive substring match on configured keyword list.
- Sends alert with matched keyword and jump link to alert channel.
- Config keys: `keyword_alerts_channel_id`, `keyword_alert_words`.

### AFK (`afk`)

- Trigger: message create.
- Behavior:
- `afk_set_phrase` command sets AFK with optional reason.
- Any later message by AFK user clears AFK automatically.
- Mentioning AFK users replies with AFK reason.
- Config keys: `afk_set_phrase`.

### Reminders (`reminders`)

- Trigger:
- Dashboard creates one-time reminder (`run_at`).
- Worker tick every 20 seconds sends due reminders.
- Behavior:
- Uses reminder row channel or falls back to `reminders_channel_id`.
- On successful send, marks reminder `sent`.
- Config keys: `reminders_channel_id`.

### Account Age Guard (`account_age_guard`)

- Trigger: member joins.
- Behavior:
- Calculates account age from Discord snowflake timestamp.
- If age is below threshold, optionally logs and either:
- `log_only`
- enqueue `quarantine`
- enqueue `kick`
- Config keys: `account_age_min_days`, `account_age_action`, `account_age_log_channel_id`.

### Member Notes (`member_notes`)

- Trigger: dashboard create/resolve note actions.
- Behavior:
- Stores moderator notes per user.
- Supports list/filter and resolve workflow.
- Optional log message to notes log channel when note is created.
- Config keys: `notes_log_channel_id`.

### Appeals (`appeals`)

- Trigger: message in appeals channel beginning with open phrase.
- Behavior:
- Creates appeal record, deletes original message, posts submission confirmation.
- Optional log channel receives appeal details.
- Dashboard resolves appeal with actor + resolution text.
- Config keys: `appeals_channel_id`, `appeals_log_channel_id`, `appeals_open_phrase`.

### Custom Commands (`custom_commands`)

- Trigger: exact message match (case-insensitive) against configured command triggers.
- Behavior:
- Sends configured response in-channel.
- If matched, AutoMod is skipped for that message.
- Config keys: feature flag only; commands managed via custom-command records.

## Cross-Module Interactions

- `custom_commands` runs before `automod`; successful command responses bypass automod checks.
- `anti_raid` and `account_age_guard` can enqueue actions into the same queue used by dashboard moderation.
- `warnings` threshold actions use the same queue and same `admin_user_policy` behavior.
- `verification` and `anti_raid (verification_only)` both rely on `unverified_role_id`.
- `audit_log_stream` can include outcomes from actions and automod for centralized observability.

## Operational Advice

- Enable modules gradually per guild, beginning with logging modules first (`audit_log_stream`, `invite_tracker`) to validate permissions.
- For action modules (`automod`, `warnings`, `anti_raid`, `account_age_guard`), confirm role hierarchy and dry-run in a test guild.
- Keep `backfill_concurrency` conservative (1-3) if you see API 429 behavior.
- Use dedicated log channels for warnings, tickets, suggestions, appeals, and account-age events to keep moderation trails auditable.
