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

Visit `http://127.0.0.1:8080` and sign in as:
- Username: `admin`
- Password: value of `MODBOT_ADMIN_PASS`

## Configuration

Environment variables:

- `MODBOT_TOKEN` Discord bot token
- `MODBOT_ADMIN_PASS` Dashboard admin password
- `MODBOT_DB` SQLite path (default: `modbot.sqlite`)
- `MODBOT_BIND` HTTP bind address (default: `127.0.0.1:8080`)
- `MODBOT_LOG_LEVEL` Log level: `info` or `debug`
- `MODBOT_DASHBOARD_ROLE_SECRETS` Optional JSON map for non-admin dashboard login credentials (example: `{"moderator":"mod-pass","support":"support-pass"}`)
- `MODBOT_DASHBOARD_SESSION_TTL_MINUTES` Session lifetime in minutes (default: `480`)
- `MODBOT_DASHBOARD_ALLOW_LEGACY_BEARER` Allow legacy secret bearer auth (`true`/`false`, default `false`)
- `MODBOT_DASHBOARD_AUTH_PROXY_ENABLED` Enable trusted auth-proxy mode for reverse-proxy SSO (`true`/`false`)
- `MODBOT_DASHBOARD_AUTH_PROXY_SECRET` Shared secret required in `X-Modbot-Proxy-Secret` header when auth-proxy mode is enabled
- `MODBOT_DASHBOARD_AUTH_PROXY_USER_HEADER` Username header name in proxy mode (default: `X-Auth-Request-User`)
- `MODBOT_DASHBOARD_AUTH_PROXY_ROLE_HEADER` Role header name in proxy mode (default: `X-Auth-Request-Role`)

Flags override env vars:

- `--token`
- `--admin-pass`
- `--db`
- `--bind`
- `--log-level`
- `--dashboard-role-secrets`
- `--dashboard-session-ttl-minutes`
- `--dashboard-allow-legacy-bearer`
- `--dashboard-auth-proxy-enabled`
- `--dashboard-auth-proxy-secret`
- `--dashboard-auth-proxy-user-header`
- `--dashboard-auth-proxy-role-header`

If token/password are not provided, startup prompts for them and saves them to local file `.modbot.config.json` (permissions `0600`) for future runs.

## Dashboard Access Model

- Dashboard users are username/password accounts.
- A bootstrap admin account is created/updated at startup:
- Username: `admin`
- Password: `MODBOT_ADMIN_PASS`
- Additional users can be managed in **Settings -> Dashboard Users (Admin)**.
- API write requests use session-bound CSRF tokens.
- Non-admin authorization is enforced server-side via `dashboard_role_policies`.
- Optional OIDC/SSO integration is supported through trusted reverse-proxy headers (auth-proxy mode).

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

- `./scripts/build.sh` (or `build.ps1`) now builds both:
- Local run binary: `./modbot` (recommended for normal operation)
- Cross-platform release artifacts: `dist/`
- Runtime config file `.modbot.config.json` is loaded from the current working directory, so launching from repo root keeps using your existing local config.

## GitHub Pages Website

Marketing site files live in `website/`.

- Page entrypoint: `website/index.html`
- Styling: `website/styles.css`
- Theme toggle and small UI behavior: `website/app.js`
- Deploy workflow: `.github/workflows/pages.yml`

Deployment steps:

1. In GitHub repo settings, open **Pages**.
2. Set **Source** to **GitHub Actions**.
3. Push to `main` with changes under `website/`.
4. Workflow publishes to `https://modulardevlabs.github.io/Fundamentum/`.
