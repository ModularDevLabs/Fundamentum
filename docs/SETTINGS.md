# Settings Reference

This document lists every configurable setting used by GoBot.
For module-level behavior details (triggers, workflows, and interactions), see `docs/MODULES.md`.

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

8a. `action_dry_run`
- UI label: `Action dry-run mode`
- Type: boolean
- Default: `false`
- When enabled, moderation actions are validated but not enqueued/executed.

8b. `action_require_confirm`
- UI label: `Require confirm token`
- Type: boolean
- Default: `true`
- Requires `confirm_token=CONFIRM` for destructive actions (`quarantine`, `kick`, `remove_roles`).

8c. `action_two_person_approval`
- UI label: `Two-person approval`
- Type: boolean
- Default: `false`
- Requires a distinct `approver_user` value for destructive action requests.

8d. `dashboard_role_policies`
- UI label: `Role policies JSON`
- Type: map of policy key -> list of dashboard roles
- Default: empty map
- Example:
- `{"actions":["moderator"],"tickets":["support","moderator"],"settings":["admin"]}`
- Empty/missing policy for a key means unrestricted.

8e. `retention_days`
- UI label: `Message retention days (0=disabled)`
- Type: integer (`>= 0`)
- Default: `0` (disabled)
- When greater than zero, the retention worker purges records older than this many days.

8f. `retention_archive_before_purge`
- UI label: `Archive summary before purge`
- Type: boolean
- Default: `true`
- Writes a `retention_archive` action row with table counts and cutoff timestamp before deleting old rows.

8g. `incident_mode_enabled`
- UI label: `Incident mode`
- Type: boolean
- Default: `false`
- Enables high-safety restrictions for destructive dashboard actions.

8h. `incident_mode_reason`
- UI label: `Incident reason`
- Type: string
- Default: empty
- Optional operator note shown in dashboard banner while incident mode is active.

8i. `incident_mode_ends_at`
- UI label: derived from `Incident auto-disable (minutes)`
- Type: RFC3339 timestamp string
- Default: empty
- When set, incident mode auto-deactivates after this timestamp.

8j. `immutable_audit_trail`
- UI label: `Immutable audit trail`
- Type: boolean
- Default: `false`
- When enabled, every emitted audit/action event is appended to a hash-linked audit trail record.

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
- `verification=false`
- `tickets=false`
- `anti_raid=false`
- `analytics=false`
- `starboard=false`
- `leveling=false`
- `giveaways=false`
- `polls=false`
- `suggestions=false`
- `keyword_alerts=false`
- `afk=false`
- `reminders=false`
- `account_age_guard=false`
- `member_notes=false`
- `appeals=false`
- `custom_commands=false`
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

24. `automod_rules`
- UI label: `Advanced rules JSON` (AutoMod module)
- Type: list of rule objects
- Default: empty list
- Supported rule `type` values:
- `regex` (matches message content against `pattern`)
- `file_ext` (matches attachment file extensions listed in `pattern`, comma-separated)
- `mention_spam` (uses `threshold` as mention count)
- `caps_ratio` (uses `threshold` as uppercase percentage; default `70`)
- Optional per-rule `action` override:
- `delete_warn`, `delete_only`, `delete_quarantine`

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

27. `verification_channel_id`
- UI label: `Verification channel ID` (Verification module)
- Type: Discord channel ID (string)

28. `verification_phrase`
- UI label: `Verify phrase` (Verification module)
- Type: string
- Default: `!verify`

29. `unverified_role_id`
- UI label: `Unverified role ID` (Verification module)
- Type: Discord role ID (string)

30. `verified_role_id`
- UI label: `Verified role ID (optional)` (Verification module)
- Type: Discord role ID (string)

31. `ticket_inbox_channel_id`
- UI label: `Inbox channel ID` (Tickets module)
- Type: Discord channel ID (string)

32. `ticket_category_id`
- UI label: `Category ID` (Tickets module)
- Type: Discord category channel ID (string)

33. `ticket_support_role_id`
- UI label: `Support role ID` (Tickets module)
- Type: Discord role ID (string)

34. `ticket_log_channel_id`
- UI label: `Log channel ID` (Tickets module)
- Type: Discord channel ID (string)

35. `ticket_open_phrase`
- UI label: `Open phrase` (Tickets module)
- Type: string
- Default: `!ticket`

36. `ticket_close_phrase`
- UI label: `Close phrase` (Tickets module)
- Type: string
- Default: `!close`

37. `ticket_auto_close_minutes`
- UI label: `Auto-close inactive (minutes, 0=off)` (Tickets module)
- Type: integer
- Default: `0`
- When greater than zero, open tickets with no recent ticket-channel activity for this many minutes are auto-closed.

38. `anti_raid_join_threshold`
- UI label: `Join threshold` (Anti-Raid module)
- Type: integer
- Default: `6`

39. `anti_raid_window_seconds`
- UI label: `Window (seconds)` (Anti-Raid module)
- Type: integer
- Default: `30`

40. `anti_raid_cooldown_minutes`
- UI label: `Cooldown (minutes)` (Anti-Raid module)
- Type: integer
- Default: `10`

41. `anti_raid_action`
- UI label: `Action` (Anti-Raid module)
- Type: enum
- Default: `verification_only`
- Options:
- `verification_only`
- `quarantine`

42. `anti_raid_alert_channel_id`
- UI label: `Alert channel ID` (Anti-Raid module)
- Type: Discord channel ID (string)

43. `analytics_channel_id`
- UI label: `Report channel ID` (Analytics module)
- Type: Discord channel ID (string)

44. `analytics_interval_days`
- UI label: `Interval (days)` (Analytics module)
- Type: integer
- Default: `7`

45. `starboard_channel_id`
- UI label: `Starboard channel ID` (Starboard module)
- Type: Discord channel ID (string)

46. `starboard_emoji`
- UI label: `Emoji` (Starboard module)
- Type: string
- Default: `⭐`

47. `starboard_threshold`
- UI label: `Threshold` (Starboard module)
- Type: integer
- Default: `3`

48. `leveling_channel_id`
- UI label: `Level-up channel ID (optional)` (Leveling module)
- Type: Discord channel ID (string)

49. `leveling_xp_per_message`
- UI label: `XP per message` (Leveling module)
- Type: integer
- Default: `10`

50. `leveling_cooldown_seconds`
- UI label: `XP cooldown (seconds)` (Leveling module)
- Type: integer
- Default: `60`

50a. `leveling_curve`
- UI label: `Level curve` (Leveling module)
- Type: enum
- Default: `quadratic`
- Options:
- `quadratic`: level requirement grows by `level^2 * leveling_xp_base`
- `linear`: level requirement grows by `level * leveling_xp_base`

50b. `leveling_xp_base`
- UI label: `XP base (level formula)` (Leveling module)
- Type: integer
- Default: `100`
- Used by `leveling_curve` to compute cumulative XP required for each level.

51. `giveaways_channel_id`
- UI label: `Default channel ID` (Giveaways module)
- Type: Discord channel ID (string)

52. `giveaways_reaction_emoji`
- UI label: `Entry emoji` (Giveaways module)
- Type: string
- Default: `🎉`

53. `polls_channel_id`
- UI label: `Default channel ID` (Polls module)
- Type: Discord channel ID (string)

54. `suggestions_channel_id`
- UI label: `Suggestions channel ID` (Suggestions module)
- Type: Discord channel ID (string)

55. `suggestions_log_channel_id`
- UI label: `Decision log channel ID (optional)` (Suggestions module)
- Type: Discord channel ID (string)

56. `keyword_alerts_channel_id`
- UI label: `Alert channel ID` (Keyword Alerts module)
- Type: Discord channel ID (string)

57. `keyword_alert_words`
- UI label: `Keywords (comma)` (Keyword Alerts module)
- Type: list of strings
- Default: empty list

58. `afk_set_phrase`
- UI label: `AFK set phrase` (AFK module)
- Type: string
- Default: `!afk`

59. `reminders_channel_id`
- UI label: `Default channel ID` (Reminders module)
- Type: Discord channel ID (string)

60. `account_age_min_days`
- UI label: `Minimum account age (days)` (Account Age Guard module)
- Type: integer
- Default: `7`

61. `account_age_action`
- UI label: `Action` (Account Age Guard module)
- Type: enum
- Default: `log_only`
- Options:
- `log_only`
- `quarantine`
- `kick`

62. `account_age_log_channel_id`
- UI label: `Log channel ID (optional)` (Account Age Guard module)
- Type: Discord channel ID (string)

63. `notes_log_channel_id`
- UI label: `Notes log channel ID (optional)` (Member Notes module)
- Type: Discord channel ID (string)

64. `appeals_channel_id`
- UI label: `Appeals channel ID` (Appeals module)
- Type: Discord channel ID (string)

65. `appeals_log_channel_id`
- UI label: `Log channel ID (optional)` (Appeals module)
- Type: Discord channel ID (string)

66. `appeals_open_phrase`
- UI label: `Open phrase` (Appeals module)
- Type: string
- Default: `!appeal`

## Giveaways Records

Configured in the `Giveaways` module UI and stored in `giveaways` / `giveaway_entries`.

Each giveaway contains:

1. `channel_id`
2. `message_id`
3. `prize`
4. `winner_count`
5. `ends_at`
6. `status`
7. `entry_count` (derived count)

## Poll Records

Configured in the `Polls` module UI and stored in `polls`.

Each poll contains:

1. `channel_id`
2. `message_id`
3. `question`
4. `options` (2-5 choices)
5. `status` (`open` or `closed`)

## Suggestions Records

Managed by the `Suggestions` module and stored in `suggestions`.

Each suggestion contains:

1. `user_id`
2. `content`
3. `message_id`
4. `status` (`open`, `approved`, `rejected`)
5. `decision_by`
6. `decision_note`

## Reminders Records

Configured in the `Reminders` module UI and stored in `reminders`.

Each reminder contains:

1. `channel_id`
2. `content`
3. `run_at`
4. `status` (`queued` or `sent`)

## Member Notes Records

Managed by the `Member Notes` module and stored in `member_notes`.

Each note contains:

1. `user_id`
2. `author_id`
3. `body`
4. `created_at`
5. `resolved_at`

## Custom Commands Rules

Configured in the `Custom Commands` module UI and stored in `custom_commands`.

Each command contains:

1. `trigger` (exact message text, case-insensitive match)
2. `response` (message sent when trigger matches)

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
