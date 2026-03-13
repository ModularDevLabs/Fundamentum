# Operations

## Start / stop

Start:

```bash
./modbot
```

Stop: `Ctrl+C`

## Dashboard workflows

- **Backfill**: Use the Overview page to run a backfill job. It scans at least the inactivity window.
- **Backfill**: Job statuses are `queued`, `running`, `success`, `partial`, `failed`.
- **Backfill**: Channels the bot cannot read (`Missing Access` / `Missing Permissions`) are counted as skipped and do not fail the whole job.
- **Members**: Filter by status (active/inactive) and run bulk actions.
- **Members**: Rows with no tracked messages show a `No messages recorded` badge and are treated as inactive.
- **Members**: Quarantined users are visually highlighted and labeled with a `Quarantined` badge.
- **Events**: Shows raw process logs (INFO/DEBUG/ERROR) for troubleshooting actions and backfill.
- **Settings**: Update inactivity days, backfill days, and quarantine behavior.
- **Settings profiles**: Apply profile presets (`Small Community`, `Gaming Server`, `Strict Moderation`) for quick baseline module configuration.
- **Action safety controls**: Enable dry-run mode, confirm-token requirement, and optional two-person approval for destructive actions.
- **Incident mode**: Toggle incident mode in Settings to raise operator visibility and add extra safeguards during active incidents.
- **Retention + maintenance**: Configure retention purges and maintenance windows in Settings.
- **Dashboard RBAC**: Use `dashboard_role_policies` with `X-Dashboard-Role` context to restrict which dashboard roles can call module/action/settings APIs.
- **Exports**: Use Settings -> Exports to download JSON/CSV for members, actions, warnings, tickets, and per-user case timelines.
- **Backup / Restore**: Use Settings -> Backup / Restore to download a guild snapshot and restore settings + reaction roles + scheduled messages + custom commands.
- **Module pages**: Every module page has its own enable/disable control, save action, and quick how-to panel.
- **Modules (moderation/safety)**: Welcome, Goodbye, Audit Log, Invite Tracker, AutoMod, Warnings, Verification, Tickets, Anti-Raid, Account Age Guard, Join Screening, Member Notes, Appeals, Custom Commands, Raid Panic.
- **Modules (engagement/community)**: Reaction Roles, Scheduled Messages, Analytics, Starboard, Leveling, Role Progression, Giveaways, Polls, Suggestions, Keyword Alerts, AFK, Reminders, Birthdays, Streaks.
- **Modules (progression/economy)**: Reputation, Economy, Achievements, Trivia.
- **Modules (utility)**: Calendar, Confessions, Season Resets.
- **Module permission checks**: Each module card shows missing bot permissions and disables save/run actions until required permissions are granted.
- **Invite Tracker**: Includes a built-in permission check panel for `Manage Server` capability in the selected guild.
- **AutoMod**: Supports ignored channels and ignored roles to exempt staff/private workflows.
- **Reaction Roles**: Add rules mapping `(channel_id, message_id, emoji) -> role_id`; role can be removed on unreact.
- **Warnings**: Issue warnings from the dashboard; thresholds can auto-queue quarantine/kick actions.
- **Scheduled**: Configure recurring messages with channel, content, and interval minutes.
- **Verification**: On join, assigns unverified role and prompts user to type verification phrase in the configured channel.
- **Tickets**: Users open tickets from inbox phrase; bot creates private channels and supports close command/dashboard close.
- **Tickets**: Transcript can be viewed from the dashboard and is included in ticket-log messages on close.
- **Tickets**: Optional inactivity auto-close runs continuously based on `ticket_auto_close_minutes`.
- **Anti-Raid**: Monitors join spikes and applies a temporary protection action (`verification_only` or `quarantine`) during cooldown.
- **Analytics**: Periodic report worker posts moderation/activity summaries to configured channel.
- **Analytics trends**: Dashboard analytics page includes daily warnings/actions/tickets trend view for the selected period.
- **Starboard**: Reposts starred messages to a configured channel when reaction count reaches threshold.
- **Leveling**: Awards XP for chat activity with cooldown control and exposes a leaderboard in the dashboard.
- **Giveaways**: Starts giveaway posts and records entries from reaction emoji; draw winners from the dashboard.
- **Polls**: Starts multi-option reaction polls and closes them with a final vote summary.
- **Suggestions**: Converts messages in a suggestions channel into vote-ready cards and supports approve/reject from dashboard.
- **Keyword Alerts**: Scans messages for configured keywords and posts alert links into a dedicated channel.
- **AFK**: Users can set AFK with a phrase (default `!afk`); the bot clears AFK on return and warns mentioners.
- **Reminders**: Queue one-time reminder messages for a specific future time from the dashboard.
- **Account Age Guard**: On join, enforces minimum account age and can log-only, quarantine, or kick.
- **Member Notes**: Add moderator notes for members, filter by user, and resolve old notes.
- **Appeals**: Users submit appeals in one channel using a phrase (default `!appeal`); dashboard lists and resolves appeals.
- **Custom Commands**: Responds to exact message triggers with configured text replies for the selected guild.
- **Birthdays**: Store user birthdays in MM-DD format and post birthday greetings daily in a configured channel.
- **Role Progression**: Define metric thresholds (`leveling`, `reputation`, `economy`) to auto-assign progression roles.
- **Join Screening**: Queue suspicious joins for moderator review/approval/rejection workflows.
- **Raid Panic**: Activate temporary lockdown controls (slowmode + posting restrictions) and auto-expire based on configured duration.
- **Streaks**: Track daily activity streaks and grant configurable coin/XP rewards.
- **Season Resets**: Reset progression modules (`leveling`, `economy`, `trivia`) on schedule or manually from dashboard.
- **Reputation**: Give or remove rep and view leaderboard.
- **Economy**: Manage shop items, balances, and purchases.
- **Achievements**: Evaluate earned badges from progression milestones.
- **Trivia**: Run question/answer rounds and maintain trivia leaderboard.
- **Calendar**: Create events and collect RSVP responses.
- **Confessions**: Intake anonymous-style confessions and optionally require moderator review before posting.
- Full field-by-field settings reference: `docs/SETTINGS.md`.
- Module behavior/configuration guide: `docs/MODULES.md`.

## Activity tracking lifecycle

- Run backfill initially to seed last-seen data from recent history.
- During backfill, once a user is seen with any message inside the inactivity window, that user is considered active for the run and additional messages from that user are skipped.
- After that, if the bot stays online, incoming message events keep user state current.
- A scheduled per-user recheck is not required while events are flowing.
- Run backfill again after downtime if you want to reconcile missed messages.
- Users with no historical data and no new messages are treated as inactive.

## Quarantine behavior

- On startup, and on guild-join events, ensures `Quarantined` role and `quarantine-readme` channel exist for each guild.
- Applies channel overwrites unless `SafeQuarantineMode` is enabled (guild-level provisioning step, not repeated per-user action).
- Adds the quarantine role and attempts to remove non-allowlisted roles (best-effort; hierarchy-protected roles may be skipped and logged).
- Quarantine only posts a readme-channel user mention when an action reason is provided.

## Action queue

Actions are queued and executed by a worker. New and retried actions wake the worker immediately (no fixed poll delay).
You can view and retry failed actions from the Actions page.
The Actions table shows target display names and keeps names for historical rows via action payload metadata.
Guild selector and user-facing tables prefer server/user names over raw IDs.

## Troubleshooting

- **Unauthorized dashboard**: Make sure `MODBOT_ADMIN_PASS` matches what you enter.
- **Missing permissions**: Verify the bot role is above target roles and permissions are granted.
- **Backfill no data**: Confirm `Read Message History` and `View Channel` permissions on channels.
- **Backfill skipped channels**: Private/admin channels inaccessible to the bot are expected to be skipped.
- **HTTP 429 / slowed actions**: Lower `Backfill concurrency` and review `docs/SETTINGS.md` rate-limit links.
- **Action failed**: Check the Events tab for the raw error and validate role hierarchy and bot permissions.

## Log level

Set `MODBOT_LOG_LEVEL=debug` for more verbose logging.
