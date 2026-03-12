# GoBot

Multi-guild Discord moderation bot with a built-in dashboard to track inactivity, quarantine users, and manage actions.

## Quickstart

1. Build

```bash
go build -o modbot ./cmd/modbot
```

2. Run

```bash
MODBOT_TOKEN=your_token MODBOT_ADMIN_PASS=your_pass ./modbot --db modbot.sqlite
```

Or use the helper script:

```bash
MODBOT_TOKEN=your_token MODBOT_ADMIN_PASS=your_pass ./run.sh
```

3. Open the dashboard

Visit `http://127.0.0.1:8080` and enter the admin password.

## Configuration

Environment variables:

- `MODBOT_TOKEN` Discord bot token
- `MODBOT_ADMIN_PASS` Dashboard admin password
- `MODBOT_DB` SQLite path (default: `modbot.sqlite`)
- `MODBOT_BIND` HTTP bind address (default: `127.0.0.1:8080`)
- `MODBOT_LOG_LEVEL` Log level: `info` or `debug`

Flags override env vars:

- `--token`
- `--admin-pass`
- `--db`
- `--bind`
- `--log-level`

If token/password are not provided, startup prompts for them and saves them to local file `.modbot.config.json` (permissions `0600`) for future runs.

For a complete settings catalog (dashboard fields, enums, defaults, and advanced/API-only settings), see `docs/SETTINGS.md`.

## Notes

- Inactivity tracking is day-based via `InactiveDays`.
- Members without any recorded messages are shown as `inactive`.
- Backfill automatically scans at least the inactivity window.
- Backfill is primarily for initial seeding and reconciliation after downtime.
- After seeding, real-time message events keep last-seen current while the bot is online.
- If the bot misses time offline, run backfill again to catch missed activity.
- Members view highlights quarantined users and shows a `Quarantined` badge.
- Actions view shows target usernames/display names and preserves names even after the user leaves (for queued/history rows).
- Events view exposes recent raw logs to help troubleshoot failed actions/backfills.
- Quarantine uses a `Quarantined` role and `quarantine-readme` channel (auto-created on startup and when the bot joins a new guild, if missing).
- Feature modules are configured per guild via dedicated dashboard menus (Welcome, Goodbye, Audit Log, Invite Tracker, AutoMod, Reaction Roles, Warnings, Scheduled, Verification, Tickets, Anti-Raid, Analytics, Starboard, Leveling, Giveaways, Polls, Suggestions, Keyword Alerts, AFK, Reminders, Account Age Guard, Appeals, Custom Commands) backed by per-guild `feature_flags`.
- Tickets include transcript viewing/export in the dashboard and optional inactivity auto-close.
- AutoMod supports ignored channels/roles to exempt moderation or staff workflows.
