# Fundamentum

Fundamentum is a multi-guild Discord operations bot with a built-in web dashboard for moderation, safety automation, engagement systems, and incident response.

## Quickstart

1. Build local binary

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

3. Open dashboard

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

## Capabilities

All feature modules are disabled by default on a new guild and enabled per-guild from the module page.

- Moderation + safety:
- Welcome, Goodbye, Audit Log, Invite Tracker, AutoMod, Warnings, Verification, Tickets, Anti-Raid, Account Age Guard, Member Notes, Appeals, Custom Commands.
- Engagement + community:
- Reaction Roles, Starboard, Leveling, Role Progression, Giveaways, Polls, Suggestions, AFK, Reminders, Birthdays, Streaks.
- Economy + progression:
- Reputation, Economy shop, Achievements, Trivia.
- Operations + incident tooling:
- Backfill jobs, module permission checks, dependency checker, policy simulator, review queue, immutable audit trail option, retention worker, maintenance windows, raid panic controls, season resets, server pulse, health dashboard, webhook integrations, exports, backup/restore.
- Additional utilities:
- Calendar + RSVP, Confessions workflow, auto-thread helper, mod summaries, voice activity rewards.

## Docs Index

- Setup and invite guide: `docs/SETUP.md`
- Operations and day-2 workflows: `docs/OPERATIONS.md`
- Full settings catalog (fields, defaults, enums): `docs/SETTINGS.md`
- Module behavior and configuration guide: `docs/MODULES.md`

## Build Outputs

- Local build: `./modbot`
- Cross-platform build script output: `dist/`
