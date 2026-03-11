# Settings Reference

This document lists every configurable setting used by GoBot.

## Dashboard Settings (per guild)

These are in the **Settings** view and are stored per server (`guild_settings`).

1. `inactive_days`
- UI label: `Inactive days`
- Type: integer (`>= 1`)
- Default: `180`
- Used for active/inactive status. A member is inactive when `last_message_at` is older than now minus this many days.

2. `backfill_days`
- UI label: `Backfill days`
- Type: integer (`>= 1`)
- Default: `60`
- Requested lookback window for backfill jobs.
- Effective window is `max(backfill_days, inactive_days)`.

3. `backfill_concurrency`
- UI label: `Backfill concurrency`
- Type: integer (`>= 1`; non-positive values fall back to `2`)
- Default: `2`
- Number of channels scanned in parallel during backfill.
- Higher values finish faster but increase API pressure and risk of rate-limit contention.
- Discord HTTP API limits are bucketed per-route plus a global app limit. Tune concurrency conservatively for small bots (for most servers, `1` to `3` is a safe start).
- Backfill deduplicates by user activity for the run: once a user is found with a qualifying message inside the inactivity window, later messages from that same user are skipped.

4. `admin_user_policy`
- UI label: `Admin user policy`
- Type: enum
- Default: `refuse`
- Options:
- `refuse`: block moderation action when target has Administrator.
- `quarantine`: allow action to proceed without pre-removing admin roles.
- `remove_admin_roles`: remove Administrator-permission roles first, then continue.

5. `quarantine_role_id`
- UI label: `Quarantine role ID`
- Type: Discord role ID (string)
- Default: empty (auto-provisioned)
- If empty, bot attempts to find/create role named `Quarantined`.

6. `readme_channel_id`
- UI label: `Readme channel ID`
- Type: Discord channel ID (string)
- Default: empty (auto-provisioned)
- If empty, bot attempts to find/create text channel named `quarantine-readme`.

7. `allowlist_role_ids`
- UI label: `Allowlist roles (comma)`
- Type: list of role IDs
- Default: empty list
- When quarantine removes roles, allowlisted roles are preserved.

8. `safe_quarantine_mode`
- UI label: `Safe quarantine mode`
- Type: boolean
- Default: `false`
- `false`: attempts guild-wide channel/category deny-overwrites for the quarantine role (except readme channel).
- `true`: skips guild-wide overwrite pass (safer for permission-limited bots).

9. `feature_flags`
- Type: object/map of booleans
- Default:
- `welcome_messages=false`
- `goodbye_messages=false`
- `audit_log_stream=false`
- `invite_tracker=false`
- `automod=false`
- `reaction_roles=false`
- `warnings=false`
- `scheduled_messages=false`
- Controls per-guild module enablement (features are off unless enabled for that server).

10. `welcome_channel_id`
- UI label: `Welcome channel ID`
- Type: Discord channel ID (string)
- Used when `feature_flags.welcome_messages=true`.

11. `welcome_message`
- UI label: `Welcome message template`
- Type: string
- Default: `Welcome {user} to {server}.`
- Tokens:
- `{user}`: member mention
- `{server}`: server name

12. `goodbye_channel_id`
- UI label: `Goodbye channel ID`
- Type: Discord channel ID (string)
- Used when `feature_flags.goodbye_messages=true`.

13. `goodbye_message`
- UI label: `Goodbye message template`
- Type: string
- Default: `Goodbye {user}.`
- Tokens:
- `{user}`: username
- `{server}`: server name

14. `audit_log_channel_id`
- UI label: `Channel ID` (Audit Log module)
- Type: Discord channel ID (string)
- Used when `feature_flags.audit_log_stream=true`.

15. `audit_log_event_types`
- UI label: `Event types (comma)` (Audit Log module)
- Type: list of strings
- Default:
- `ban_add`
- `ban_remove`
- `role_create`
- `role_update`
- `role_delete`
- `channel_create`
- `channel_update`
- `channel_delete`
- `action_success`
- `action_failed`
- `automod_action`

16. `invite_log_channel_id`
- UI label: `Log channel ID` (Invite Tracker module)
- Type: Discord channel ID (string)
- Used when `feature_flags.invite_tracker=true`.

17. `automod_block_links`
- UI label: `Block links` (AutoMod module)
- Type: boolean
- Default: `true`

18. `automod_blocked_words`
- UI label: `Blocked words (comma)` (AutoMod module)
- Type: list of strings
- Default: empty list

19. `automod_dup_window_sec`
- UI label: `Duplicate window (seconds)` (AutoMod module)
- Type: integer
- Default: `20`

20. `automod_dup_threshold`
- UI label: `Duplicate threshold` (AutoMod module)
- Type: integer
- Default: `3`
- When the same member sends the same content this many times inside the window, AutoMod triggers.

21. `automod_action`
- UI label: `Action` (AutoMod module)
- Type: enum
- Default: `delete_warn`
- Options:
- `delete_warn`
- `delete_only`
- `delete_quarantine`

22. `automod_ignore_channel_ids`
- UI label: `Ignored channel IDs (comma)` (AutoMod module)
- Type: list of Discord channel IDs
- Messages in these channels are ignored by AutoMod.

23. `automod_ignore_role_ids`
- UI label: `Ignored role IDs (comma)` (AutoMod module)
- Type: list of Discord role IDs
- Members with any ignored role are skipped by AutoMod.

24. `warning_log_channel_id`
- UI label: `Log channel ID` (Warnings module)
- Type: Discord channel ID (string)

25. `warn_quarantine_threshold`
- UI label: `Quarantine threshold` (Warnings module)
- Type: integer
- Default: `3`
- At or above this warning count, warning issuance auto-queues quarantine.

26. `warn_kick_threshold`
- UI label: `Kick threshold` (Warnings module)
- Type: integer
- Default: `5`
- At or above this warning count, warning issuance auto-queues kick.

## Reaction Roles Rules

Configured in the `Reaction Roles` module UI and stored in `reaction_role_rules`.

Each rule contains:

1. `channel_id`
2. `message_id`
3. `emoji` (unicode emoji or custom emoji ID)
4. `role_id`
5. `remove_on_unreact` (boolean)

## Scheduled Messages

Configured in the `Scheduled` module UI and stored in `scheduled_messages`.

Each schedule contains:

1. `channel_id`
2. `content`
3. `interval_minutes`
4. `enabled`
5. `next_run_at` (managed automatically by the worker)

## Advanced Per-Guild Setting (API/DB)

This setting exists in the model/API and is not currently exposed in the dashboard form.

1. `backfill_include_types`
- Type: list of channel-type names
- Default: empty list
- If empty, backfill scans `GUILD_TEXT` and `GUILD_NEWS`.
- If set, only listed types are scanned.
- Supported values:
- `GUILD_TEXT`
- `GUILD_NEWS`
- `GUILD_FORUM`
- `GUILD_PUBLIC_THREAD`
- `GUILD_PRIVATE_THREAD`
- `GUILD_NEWS_THREAD`

## Process Configuration (bot instance)

These configure the local bot process, not per-guild behavior.

1. `MODBOT_TOKEN` or `--token`
- Discord bot token.

2. `MODBOT_ADMIN_PASS` or `--admin-pass`
- Dashboard login password.

3. `MODBOT_DB` or `--db`
- SQLite path. Default: `modbot.sqlite`.

4. `MODBOT_BIND` or `--bind`
- Web bind address. Default: `127.0.0.1:8080`.

5. `MODBOT_LOG_LEVEL` or `--log-level`
- `info` or `debug`.

If token/password are missing at startup, the app prompts and saves to local `.modbot.config.json` (mode `0600`).

## Discord Rate-Limit References

Use these official docs when tuning backfill/action throughput:

1. HTTP API rate limits (headers, retry behavior, global and invalid-request limits):  
`https://docs.discord.com/developers/topics/rate-limits`
2. Gateway rate limits (websocket event send limits and identify constraints):  
`https://docs.discord.com/developers/docs/topics/gateway#rate-limiting`
3. Discord support explainer for diagnosing 429s:  
`https://support-dev.discord.com/hc/en-us/articles/6223003921559-My-Bot-Is-Being-Rate-Limited`
