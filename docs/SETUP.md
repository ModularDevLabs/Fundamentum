# Setup

This guide walks you through creating the Discord application, inviting the bot, and launching the dashboard.

## 1) Create the Discord app and bot

1. Go to the Discord Developer Portal and create a new application.
2. Add a Bot to the application.
3. Copy the bot token (you will need it to run the bot).

## 2) Enable privileged intents

Enable these privileged intents for the bot:

- `MESSAGE CONTENT` (for message activity tracking)
- `SERVER MEMBERS INTENT` (for member join/leave events)

Also enable gateway intent:

- `GUILD MESSAGE REACTIONS` (for Reaction Roles, Starboard, Giveaways, Polls, Suggestions)

## 3) Required permissions

The bot needs the following minimum permissions for the current feature set:

- View Channel
- Read Message History
- Send Messages
- Send Messages in Threads
- Add Reactions
- Manage Roles
- Manage Channels
- Kick Members
- Manage Server (required for Invite Tracker to resolve invite usage)
- Manage Messages (required for AutoMod message deletion)
- Embed Links (recommended for rich module posts such as giveaways/polls/summaries)

These map to the following permission bit values in the Discord docs:
`KICK_MEMBERS (1<<1)`, `MANAGE_CHANNELS (1<<4)`, `MANAGE_GUILD (1<<5)`, `ADD_REACTIONS (1<<6)`, `VIEW_CHANNEL (1<<10)`, `SEND_MESSAGES (1<<11)`, `MANAGE_MESSAGES (1<<13)`, `EMBED_LINKS (1<<14)`, `READ_MESSAGE_HISTORY (1<<16)`, `SEND_MESSAGES_IN_THREADS (1<<38)`, `MANAGE_ROLES (1<<28)`.

### Combined permissions integer

With `SEND_MESSAGES`, `EMBED_LINKS`, and `SEND_MESSAGES_IN_THREADS` included:

```
275146435698
```

Without `SEND_MESSAGES`:

```
275146433650
```

## 4) One-click invite

Replace `CLIENT_ID` with your application client ID and open the URL:

```
https://discord.com/oauth2/authorize?client_id=CLIENT_ID&scope=bot&permissions=275146435698
```

## 5) Role hierarchy

The bot’s highest role must be above any roles it needs to remove or assign, otherwise Discord will reject those actions.

For full quarantine lockdown (`SafeQuarantineMode=false`), the bot must also be able to edit overwrites on all channels/categories it should hide. If some channels are inaccessible, quarantine still applies but those channels are logged as skipped in the Events tab.

## 6) Run the bot

You can run it with env vars or just run and follow the prompt.

```bash
./modbot
```

Or with explicit env vars:

```bash
MODBOT_TOKEN=your_token MODBOT_ADMIN_PASS=your_pass ./modbot --db modbot.sqlite
```

Open the dashboard at `http://127.0.0.1:8080` and log in with:
- Username: `admin`
- Password: value from `MODBOT_ADMIN_PASS`

Create additional dashboard users/roles from `Settings -> Dashboard Users (Admin)`.

## 7) Exposing the dashboard safely

Default bind is local (`127.0.0.1:8080`). If multiple operators need access:

1. Keep Fundamentum bound to localhost.
2. Place a reverse proxy in front with HTTPS termination.
3. Restrict access (VPN, IP allowlist, or SSO at the proxy).
4. Optionally enable trusted auth-proxy mode for OIDC/SSO:
- `MODBOT_DASHBOARD_AUTH_PROXY_ENABLED=true`
- `MODBOT_DASHBOARD_AUTH_PROXY_SECRET=<shared-secret>`
- Proxy must inject `X-Modbot-Proxy-Secret`, user header, and role header.

On startup, and when the bot joins a guild while running, the bot attempts to auto-create/ensure the quarantine role and readme channel.

## References

```
https://docs.discord.com/developers/topics/permissions
https://docs.discord.com/developers/events/gateway
https://docs.discord.com/developers/topics/oauth2
https://support.discord.com/hc/en-us/articles/214836687-Role-Management-101
```
