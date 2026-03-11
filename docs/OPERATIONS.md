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
- **Welcome / Goodbye / Audit Log / Invite Tracker / AutoMod / Reaction Roles**: Each module has its own menu and save action, and settings are per-guild.
- **Invite Tracker**: Includes a built-in permission check panel for `Manage Server` capability in the selected guild.
- **AutoMod**: Supports ignored channels and ignored roles to exempt staff/private workflows.
- **Reaction Roles**: Add rules mapping `(channel_id, message_id, emoji) -> role_id`; role can be removed on unreact.
- Full field-by-field settings reference: `docs/SETTINGS.md`.

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
