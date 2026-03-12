const state = {
  token: localStorage.getItem('modbot_token') || '',
  guildId: localStorage.getItem('modbot_guild') || '',
  guilds: [],
  selectedUsers: new Map(),
  overviewPollTimer: null,
  overviewPollBusy: false,
  eventsPollTimer: null,
  memberFilterWatchTimer: null,
  lastMemberStatusValue: '',
};

const FEATURE_WELCOME = 'welcome_messages';
const FEATURE_GOODBYE = 'goodbye_messages';
const FEATURE_AUDIT = 'audit_log_stream';
const FEATURE_INVITE = 'invite_tracker';
const FEATURE_AUTOMOD = 'automod';
const FEATURE_REACTION_ROLES = 'reaction_roles';
const FEATURE_WARNINGS = 'warnings';
const FEATURE_SCHEDULED = 'scheduled_messages';
const FEATURE_VERIFICATION = 'verification';
const FEATURE_TICKETS = 'tickets';
const FEATURE_ANTI_RAID = 'anti_raid';
const FEATURE_ANALYTICS = 'analytics';
const FEATURE_STARBOARD = 'starboard';
const FEATURE_LEVELING = 'leveling';
const FEATURE_GIVEAWAYS = 'giveaways';
const FEATURE_POLLS = 'polls';
const FEATURE_SUGGESTIONS = 'suggestions';
const FEATURE_KEYWORD_ALERTS = 'keyword_alerts';
const FEATURE_APPEALS = 'appeals';
const FEATURE_CUSTOM_COMMANDS = 'custom_commands';

function setModuleBadge(enabled, badgeEl, cardEl) {
  if (!badgeEl || !cardEl) return;
  badgeEl.textContent = enabled ? 'Enabled' : 'Disabled';
  badgeEl.classList.toggle('on', enabled);
  badgeEl.classList.toggle('off', !enabled);
  cardEl.classList.toggle('enabled', enabled);
}

function syncModuleBadges() {
  const welcomeEnabled = qs('#settingsWelcomeEnabled').value === 'true';
  const goodbyeEnabled = qs('#settingsGoodbyeEnabled').value === 'true';
  const auditEnabled = qs('#settingsAuditEnabled').value === 'true';
  const inviteEnabled = qs('#settingsInviteEnabled').value === 'true';
  const autoModEnabled = qs('#settingsAutoModEnabled').value === 'true';
  const reactionRolesEnabled = qs('#settingsReactionRolesEnabled').value === 'true';
  const warningsEnabled = qs('#settingsWarningsEnabled').value === 'true';
  const scheduledEnabled = qs('#settingsScheduledEnabled').value === 'true';
  const verificationEnabled = qs('#settingsVerificationEnabled').value === 'true';
  const ticketsEnabled = qs('#settingsTicketsEnabled').value === 'true';
  const antiRaidEnabled = qs('#settingsAntiRaidEnabled').value === 'true';
  const analyticsEnabled = qs('#settingsAnalyticsEnabled').value === 'true';
  const starboardEnabled = qs('#settingsStarboardEnabled').value === 'true';
  const levelingEnabled = qs('#settingsLevelingEnabled').value === 'true';
  const giveawaysEnabled = qs('#settingsGiveawaysEnabled').value === 'true';
  const pollsEnabled = qs('#settingsPollsEnabled').value === 'true';
  const suggestionsEnabled = qs('#settingsSuggestionsEnabled').value === 'true';
  const keywordAlertsEnabled = qs('#settingsKeywordAlertsEnabled').value === 'true';
  const appealsEnabled = qs('#settingsAppealsEnabled').value === 'true';
  const customCommandsEnabled = qs('#settingsCustomCommandsEnabled').value === 'true';
  setModuleBadge(welcomeEnabled, qs('#moduleWelcomeBadge'), qs('#moduleWelcomeCard'));
  setModuleBadge(goodbyeEnabled, qs('#moduleGoodbyeBadge'), qs('#moduleGoodbyeCard'));
  setModuleBadge(auditEnabled, qs('#moduleAuditBadge'), qs('#moduleAuditCard'));
  setModuleBadge(inviteEnabled, qs('#moduleInviteBadge'), qs('#moduleInviteCard'));
  setModuleBadge(autoModEnabled, qs('#moduleAutoModBadge'), qs('#moduleAutoModCard'));
  setModuleBadge(reactionRolesEnabled, qs('#moduleReactionRolesBadge'), qs('#moduleReactionRolesCard'));
  setModuleBadge(warningsEnabled, qs('#moduleWarningsBadge'), qs('#moduleWarningsCard'));
  setModuleBadge(scheduledEnabled, qs('#moduleScheduledBadge'), qs('#moduleScheduledCard'));
  setModuleBadge(verificationEnabled, qs('#moduleVerificationBadge'), qs('#moduleVerificationCard'));
  setModuleBadge(ticketsEnabled, qs('#moduleTicketsBadge'), qs('#moduleTicketsCard'));
  setModuleBadge(antiRaidEnabled, qs('#moduleAntiRaidBadge'), qs('#moduleAntiRaidCard'));
  setModuleBadge(analyticsEnabled, qs('#moduleAnalyticsBadge'), qs('#moduleAnalyticsCard'));
  setModuleBadge(starboardEnabled, qs('#moduleStarboardBadge'), qs('#moduleStarboardCard'));
  setModuleBadge(levelingEnabled, qs('#moduleLevelingBadge'), qs('#moduleLevelingCard'));
  setModuleBadge(giveawaysEnabled, qs('#moduleGiveawaysBadge'), qs('#moduleGiveawaysCard'));
  setModuleBadge(pollsEnabled, qs('#modulePollsBadge'), qs('#modulePollsCard'));
  setModuleBadge(suggestionsEnabled, qs('#moduleSuggestionsBadge'), qs('#moduleSuggestionsCard'));
  setModuleBadge(keywordAlertsEnabled, qs('#moduleKeywordAlertsBadge'), qs('#moduleKeywordAlertsCard'));
  setModuleBadge(appealsEnabled, qs('#moduleAppealsBadge'), qs('#moduleAppealsCard'));
  setModuleBadge(customCommandsEnabled, qs('#moduleCustomCommandsBadge'), qs('#moduleCustomCommandsCard'));
}

const qs = (sel) => document.querySelector(sel);
const qsa = (sel) => Array.from(document.querySelectorAll(sel));

const loginModal = qs('#loginModal');
const loginError = qs('#loginError');
const loginInput = qs('#loginPassword');
const toastHost = qs('#toastHost');

function showToast(message, kind = 'success') {
  const toast = document.createElement('div');
  toast.className = `toast ${kind}`;
  toast.textContent = message;
  toastHost.appendChild(toast);
  setTimeout(() => {
    toast.remove();
  }, 3200);
}

function setBusy(button, busyLabel) {
  if (!button) return () => {};
  const original = button.textContent;
  button.disabled = true;
  button.textContent = busyLabel || 'Working...';
  return () => {
    button.disabled = false;
    button.textContent = original;
  };
}

function showLogin() {
  stopOverviewPolling();
  stopEventsPolling();
  loginModal.classList.remove('hidden');
}

function hideLogin() {
  loginModal.classList.add('hidden');
}

function stopOverviewPolling() {
  if (state.overviewPollTimer) {
    clearInterval(state.overviewPollTimer);
    state.overviewPollTimer = null;
  }
  state.overviewPollBusy = false;
}

function stopEventsPolling() {
  if (state.eventsPollTimer) {
    clearInterval(state.eventsPollTimer);
    state.eventsPollTimer = null;
  }
}

function startMemberFilterWatch() {
  if (state.memberFilterWatchTimer) return;
  const statusEl = qs('#memberStatus');
  if (!statusEl) return;
  state.lastMemberStatusValue = statusEl.value || '';
  state.memberFilterWatchTimer = setInterval(() => {
    const current = statusEl.value || '';
    if (current === state.lastMemberStatusValue) return;
    state.lastMemberStatusValue = current;
    loadMembers().catch((err) => showToast(`Members load failed: ${err.message}`, 'error'));
  }, 250);
}

function startEventsPolling() {
  if (state.eventsPollTimer) return;
  state.eventsPollTimer = setInterval(() => {
    loadEvents().catch(() => {});
  }, 2500);
}

function syncOverviewPolling(backfills) {
  const hasActive = (backfills || []).some((job) => job.status === 'queued' || job.status === 'running');
  if (!hasActive) {
    stopOverviewPolling();
    const status = qs('#overviewStatus');
    if (status && status.textContent.startsWith('Auto-refreshing')) {
      status.textContent = '';
    }
    return;
  }
  if (state.overviewPollTimer) {
    return;
  }
  const status = qs('#overviewStatus');
  status.textContent = 'Auto-refreshing while backfill is running...';
  state.overviewPollTimer = setInterval(async () => {
    if (state.overviewPollBusy) {
      return;
    }
    state.overviewPollBusy = true;
    try {
      await loadOverview();
    } finally {
      state.overviewPollBusy = false;
    }
  }, 3000);
}

async function apiFetch(path, options = {}) {
  const headers = options.headers || {};
  if (state.token) {
    headers['Authorization'] = `Bearer ${state.token}`;
  }
  const res = await fetch(path, { ...options, headers });
  if (res.status === 401) {
    showLogin();
    throw new Error('unauthorized');
  }
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || 'request failed');
  }
  if (res.status === 204) return null;
  return res.json();
}

async function login() {
  loginError.textContent = '';
  const password = loginInput.value.trim();
  if (!password) return;
  const res = await fetch('/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password }),
  });
  if (res.status !== 204) {
    loginError.textContent = 'Incorrect password.';
    return;
  }
  state.token = password;
  localStorage.setItem('modbot_token', password);
  hideLogin();
  await bootstrap();
}

async function loadGuilds() {
  const data = await apiFetch('/api/guilds');
  state.guilds = data || [];
  const select = qs('#guildSelect');
  select.innerHTML = '';
  state.guilds.forEach((g) => {
    const opt = document.createElement('option');
    opt.value = g.id;
    opt.textContent = g.name || 'Unknown Server';
    select.appendChild(opt);
  });
  if (!state.guildId && state.guilds.length) {
    state.guildId = state.guilds[0].id;
  }
  select.value = state.guildId;
  select.onchange = () => {
    state.guildId = select.value;
    localStorage.setItem('modbot_guild', state.guildId);
    refreshAll();
  };
}

async function loadSettings() {
  if (!state.guildId) return;
  const cfg = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
  const flags = cfg.feature_flags || {};
  qs('#settingsInactive').value = cfg.inactive_days;
  qs('#settingsBackfill').value = cfg.backfill_days;
  qs('#settingsConcurrency').value = cfg.backfill_concurrency;
  qs('#settingsAdminPolicy').value = cfg.admin_user_policy;
  qs('#settingsQuarantineRole').value = cfg.quarantine_role_id || '';
  qs('#settingsReadmeChannel').value = cfg.readme_channel_id || '';
  qs('#settingsAllowlist').value = (cfg.allowlist_role_ids || []).join(',');
  qs('#settingsSafeMode').value = String(cfg.safe_quarantine_mode);
  qs('#settingsWelcomeEnabled').value = String(!!flags[FEATURE_WELCOME]);
  qs('#settingsWelcomeChannel').value = cfg.welcome_channel_id || '';
  qs('#settingsWelcomeMessage').value = cfg.welcome_message || '';
  qs('#settingsGoodbyeEnabled').value = String(!!flags[FEATURE_GOODBYE]);
  qs('#settingsGoodbyeChannel').value = cfg.goodbye_channel_id || '';
  qs('#settingsGoodbyeMessage').value = cfg.goodbye_message || '';
  qs('#settingsAuditEnabled').value = String(!!flags[FEATURE_AUDIT]);
  qs('#settingsAuditChannel').value = cfg.audit_log_channel_id || '';
  qs('#settingsAuditEvents').value = (cfg.audit_log_event_types || []).join(',');
  qs('#settingsInviteEnabled').value = String(!!flags[FEATURE_INVITE]);
  qs('#settingsInviteChannel').value = cfg.invite_log_channel_id || '';
  qs('#settingsAutoModEnabled').value = String(!!flags[FEATURE_AUTOMOD]);
  qs('#settingsReactionRolesEnabled').value = String(!!flags[FEATURE_REACTION_ROLES]);
  qs('#settingsAutoModAction').value = cfg.automod_action || 'delete_warn';
  qs('#settingsAutoModBlockLinks').value = String(!!cfg.automod_block_links);
  qs('#settingsAutoModWords').value = (cfg.automod_blocked_words || []).join(',');
  qs('#settingsAutoModDupWindow').value = cfg.automod_dup_window_sec || 20;
  qs('#settingsAutoModDupThreshold').value = cfg.automod_dup_threshold || 3;
  qs('#settingsAutoModIgnoreChannels').value = (cfg.automod_ignore_channel_ids || []).join(',');
  qs('#settingsAutoModIgnoreRoles').value = (cfg.automod_ignore_role_ids || []).join(',');
  qs('#settingsWarningsEnabled').value = String(!!flags[FEATURE_WARNINGS]);
  qs('#settingsWarningLogChannel').value = cfg.warning_log_channel_id || '';
  qs('#settingsWarnQuarantineThreshold').value = cfg.warn_quarantine_threshold || 3;
  qs('#settingsWarnKickThreshold').value = cfg.warn_kick_threshold || 5;
  qs('#settingsScheduledEnabled').value = String(!!flags[FEATURE_SCHEDULED]);
  qs('#settingsVerificationEnabled').value = String(!!flags[FEATURE_VERIFICATION]);
  qs('#settingsVerificationChannel').value = cfg.verification_channel_id || '';
  qs('#settingsVerificationPhrase').value = cfg.verification_phrase || '!verify';
  qs('#settingsUnverifiedRole').value = cfg.unverified_role_id || '';
  qs('#settingsVerifiedRole').value = cfg.verified_role_id || '';
  qs('#settingsTicketsEnabled').value = String(!!flags[FEATURE_TICKETS]);
  qs('#settingsTicketInbox').value = cfg.ticket_inbox_channel_id || '';
  qs('#settingsTicketCategory').value = cfg.ticket_category_id || '';
  qs('#settingsTicketSupportRole').value = cfg.ticket_support_role_id || '';
  qs('#settingsTicketLogChannel').value = cfg.ticket_log_channel_id || '';
  qs('#settingsTicketOpenPhrase').value = cfg.ticket_open_phrase || '!ticket';
  qs('#settingsTicketClosePhrase').value = cfg.ticket_close_phrase || '!close';
  qs('#settingsTicketAutoClose').value = cfg.ticket_auto_close_minutes || 0;
  qs('#settingsAntiRaidEnabled').value = String(!!flags[FEATURE_ANTI_RAID]);
  qs('#settingsAntiRaidThreshold').value = cfg.anti_raid_join_threshold || 6;
  qs('#settingsAntiRaidWindow').value = cfg.anti_raid_window_seconds || 30;
  qs('#settingsAntiRaidCooldown').value = cfg.anti_raid_cooldown_minutes || 10;
  qs('#settingsAntiRaidAction').value = cfg.anti_raid_action || 'verification_only';
  qs('#settingsAntiRaidAlertChannel').value = cfg.anti_raid_alert_channel_id || '';
  qs('#settingsAnalyticsEnabled').value = String(!!flags[FEATURE_ANALYTICS]);
  qs('#settingsAnalyticsChannel').value = cfg.analytics_channel_id || '';
  qs('#settingsAnalyticsIntervalDays').value = cfg.analytics_interval_days || 7;
  qs('#settingsStarboardEnabled').value = String(!!flags[FEATURE_STARBOARD]);
  qs('#settingsStarboardChannel').value = cfg.starboard_channel_id || '';
  qs('#settingsStarboardEmoji').value = cfg.starboard_emoji || '⭐';
  qs('#settingsStarboardThreshold').value = cfg.starboard_threshold || 3;
  qs('#settingsLevelingEnabled').value = String(!!flags[FEATURE_LEVELING]);
  qs('#settingsLevelingChannel').value = cfg.leveling_channel_id || '';
  qs('#settingsLevelingXP').value = cfg.leveling_xp_per_message || 10;
  qs('#settingsLevelingCooldown').value = cfg.leveling_cooldown_seconds || 60;
  qs('#settingsGiveawaysEnabled').value = String(!!flags[FEATURE_GIVEAWAYS]);
  qs('#settingsGiveawaysChannel').value = cfg.giveaways_channel_id || '';
  qs('#settingsGiveawaysEmoji').value = cfg.giveaways_reaction_emoji || '🎉';
  qs('#settingsPollsEnabled').value = String(!!flags[FEATURE_POLLS]);
  qs('#settingsPollsChannel').value = cfg.polls_channel_id || '';
  qs('#settingsSuggestionsEnabled').value = String(!!flags[FEATURE_SUGGESTIONS]);
  qs('#settingsSuggestionsChannel').value = cfg.suggestions_channel_id || '';
  qs('#settingsSuggestionsLogChannel').value = cfg.suggestions_log_channel_id || '';
  qs('#settingsKeywordAlertsEnabled').value = String(!!flags[FEATURE_KEYWORD_ALERTS]);
  qs('#settingsKeywordAlertsChannel').value = cfg.keyword_alerts_channel_id || '';
  qs('#settingsKeywordAlertWords').value = (cfg.keyword_alert_words || []).join(',');
  qs('#settingsAppealsEnabled').value = String(!!flags[FEATURE_APPEALS]);
  qs('#settingsAppealsChannel').value = cfg.appeals_channel_id || '';
  qs('#settingsAppealsLogChannel').value = cfg.appeals_log_channel_id || '';
  qs('#settingsAppealsPhrase').value = cfg.appeals_open_phrase || '!appeal';
  qs('#settingsCustomCommandsEnabled').value = String(!!flags[FEATURE_CUSTOM_COMMANDS]);
  syncModuleBadges();
  await loadInvitePermissionStatus();
  await loadReactionRoleRules();
  await loadWarnings();
  await loadScheduledMessages();
  await loadTickets();
  await loadAppeals();
  await loadCustomCommands();
  await loadLeaderboard();
  await loadGiveaways();
  await loadPolls();
  await loadSuggestions();
}

async function loadInvitePermissionStatus() {
  const note = qs('#invitePermStatus');
  if (!note || !state.guildId) return;
  note.classList.remove('ok', 'warn');
  note.textContent = 'Checking bot permission requirements...';
  try {
    const status = await apiFetch(`/api/modules/invite/status?guild_id=${state.guildId}`);
    if (status.has_manage_guild) {
      note.classList.add('ok');
      note.textContent = 'Invite permission check passed: bot has Manage Server.';
      return;
    }
    note.classList.add('warn');
    note.textContent = 'Invite Tracker warning: bot lacks Manage Server in this guild, invite attribution may fail.';
  } catch (err) {
    note.classList.add('warn');
    note.textContent = `Invite permission check failed: ${err.message}`;
  }
}

async function saveSettings() {
  const restore = setBusy(qs('#settingsSave'), 'Saving...');
  const status = qs('#settingsStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      inactive_days: parseInt(qs('#settingsInactive').value, 10),
      backfill_days: parseInt(qs('#settingsBackfill').value, 10),
      backfill_concurrency: parseInt(qs('#settingsConcurrency').value, 10),
      admin_user_policy: qs('#settingsAdminPolicy').value,
      quarantine_role_id: qs('#settingsQuarantineRole').value.trim(),
      readme_channel_id: qs('#settingsReadmeChannel').value.trim(),
      allowlist_role_ids: qs('#settingsAllowlist').value.split(',').map((v) => v.trim()).filter(Boolean),
      safe_quarantine_mode: qs('#settingsSafeMode').value === 'true',
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Settings saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Settings save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveWelcome() {
  const restore = setBusy(qs('#welcomeSave'), 'Saving...');
  const status = qs('#welcomeStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_WELCOME]: qs('#settingsWelcomeEnabled').value === 'true',
      },
      welcome_channel_id: qs('#settingsWelcomeChannel').value.trim(),
      welcome_message: qs('#settingsWelcomeMessage').value,
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Welcome module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Welcome module save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveGoodbye() {
  const restore = setBusy(qs('#goodbyeSave'), 'Saving...');
  const status = qs('#goodbyeStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_GOODBYE]: qs('#settingsGoodbyeEnabled').value === 'true',
      },
      goodbye_channel_id: qs('#settingsGoodbyeChannel').value.trim(),
      goodbye_message: qs('#settingsGoodbyeMessage').value,
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Goodbye module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Goodbye module save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveAudit() {
  const restore = setBusy(qs('#auditSave'), 'Saving...');
  const status = qs('#auditStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_AUDIT]: qs('#settingsAuditEnabled').value === 'true',
      },
      audit_log_channel_id: qs('#settingsAuditChannel').value.trim(),
      audit_log_event_types: qs('#settingsAuditEvents').value.split(',').map((v) => v.trim()).filter(Boolean),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Audit module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Audit module save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveInviteTracker() {
  const restore = setBusy(qs('#inviteSave'), 'Saving...');
  const status = qs('#inviteStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_INVITE]: qs('#settingsInviteEnabled').value === 'true',
      },
      invite_log_channel_id: qs('#settingsInviteChannel').value.trim(),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    await loadInvitePermissionStatus();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Invite tracker saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Invite tracker save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveAutoMod() {
  const restore = setBusy(qs('#automodSave'), 'Saving...');
  const status = qs('#automodStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_AUTOMOD]: qs('#settingsAutoModEnabled').value === 'true',
      },
      automod_action: qs('#settingsAutoModAction').value,
      automod_block_links: qs('#settingsAutoModBlockLinks').value === 'true',
      automod_blocked_words: qs('#settingsAutoModWords').value.split(',').map((v) => v.trim()).filter(Boolean),
      automod_dup_window_sec: parseInt(qs('#settingsAutoModDupWindow').value, 10),
      automod_dup_threshold: parseInt(qs('#settingsAutoModDupThreshold').value, 10),
      automod_ignore_channel_ids: qs('#settingsAutoModIgnoreChannels').value.split(',').map((v) => v.trim()).filter(Boolean),
      automod_ignore_role_ids: qs('#settingsAutoModIgnoreRoles').value.split(',').map((v) => v.trim()).filter(Boolean),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('AutoMod saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`AutoMod save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveReactionRoles() {
  const restore = setBusy(qs('#reactionRolesSave'), 'Saving...');
  const status = qs('#reactionRolesStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_REACTION_ROLES]: qs('#settingsReactionRolesEnabled').value === 'true',
      },
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Reaction roles module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Reaction roles save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function loadReactionRoleRules() {
  const table = qs('#rrRulesTable');
  if (!table || !state.guildId) return;
  const rows = (await apiFetch(`/api/modules/reaction-roles/rules?guild_id=${state.guildId}`)) || [];
  table.innerHTML = '';
  rows.forEach((rule) => {
    const div = document.createElement('div');
    div.className = 'table-row rr-row';
    div.innerHTML = `
      <div>${rule.channel_id}</div>
      <div>${rule.message_id}</div>
      <div>${rule.emoji}</div>
      <div>${rule.role_id}</div>
      <div>${rule.remove_on_unreact ? 'yes' : 'no'}</div>
      <div><button class="ghost" data-rr-delete="${rule.id}">Delete</button></div>
    `;
    table.appendChild(div);
  });
}

async function addReactionRoleRule() {
  const restore = setBusy(qs('#rrAddRule'), 'Adding...');
  const status = qs('#rrStatus');
  status.textContent = 'Adding...';
  try {
    const payload = {
      channel_id: qs('#rrChannelId').value.trim(),
      message_id: qs('#rrMessageId').value.trim(),
      emoji: qs('#rrEmoji').value.trim(),
      role_id: qs('#rrRoleId').value.trim(),
      remove_on_unreact: qs('#rrRemoveOnUnreact').value === 'true',
    };
    await apiFetch(`/api/modules/reaction-roles/rules?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    status.textContent = `Added at ${new Date().toLocaleTimeString()}`;
    showToast('Reaction role rule added.');
    await loadReactionRoleRules();
  } catch (err) {
    status.textContent = 'Add failed.';
    showToast(`Add rule failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function deleteReactionRoleRule(id) {
  if (!id) return;
  await apiFetch(`/api/modules/reaction-roles/rules/${id}?guild_id=${state.guildId}`, { method: 'DELETE' });
}

async function saveWarningsModule() {
  const restore = setBusy(qs('#warningsSave'), 'Saving...');
  const status = qs('#warningsStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_WARNINGS]: qs('#settingsWarningsEnabled').value === 'true',
      },
      warning_log_channel_id: qs('#settingsWarningLogChannel').value.trim(),
      warn_quarantine_threshold: parseInt(qs('#settingsWarnQuarantineThreshold').value, 10),
      warn_kick_threshold: parseInt(qs('#settingsWarnKickThreshold').value, 10),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Warnings module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Warnings save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function loadWarnings() {
  const table = qs('#warningsTable');
  if (!table || !state.guildId) return;
  const rows = (await apiFetch(`/api/modules/warnings?guild_id=${state.guildId}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row warn-row';
    div.innerHTML = `
      <div>${row.user_id}</div>
      <div>${row.actor_user_id}</div>
      <div>${row.reason || ''}</div>
      <div>${formatDate(row.created_at)}</div>
    `;
    table.appendChild(div);
  });
}

async function issueWarning() {
  const restore = setBusy(qs('#warnIssue'), 'Issuing...');
  const status = qs('#warnStatus');
  status.textContent = 'Issuing...';
  try {
    const payload = {
      user_id: qs('#warnUserId').value.trim(),
      reason: qs('#warnReason').value.trim(),
    };
    const res = await apiFetch(`/api/modules/warnings/issue?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    status.textContent = `Warning issued (count=${res.count}${res.auto_action ? `, auto=${res.auto_action}` : ''})`;
    showToast('Warning issued.');
    await loadWarnings();
    await loadActions();
  } catch (err) {
    status.textContent = 'Issue failed.';
    showToast(`Issue warning failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveScheduledModule() {
  const restore = setBusy(qs('#scheduledSave'), 'Saving...');
  const status = qs('#scheduledStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_SCHEDULED]: qs('#settingsScheduledEnabled').value === 'true',
      },
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Scheduled module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Scheduled module save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveVerificationModule() {
  const restore = setBusy(qs('#verificationSave'), 'Saving...');
  const status = qs('#verificationStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_VERIFICATION]: qs('#settingsVerificationEnabled').value === 'true',
      },
      verification_channel_id: qs('#settingsVerificationChannel').value.trim(),
      verification_phrase: qs('#settingsVerificationPhrase').value.trim(),
      unverified_role_id: qs('#settingsUnverifiedRole').value.trim(),
      verified_role_id: qs('#settingsVerifiedRole').value.trim(),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Verification module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Verification save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function loadScheduledMessages() {
  const table = qs('#scheduledTable');
  if (!table || !state.guildId) return;
  const rows = (await apiFetch(`/api/modules/scheduled/messages?guild_id=${state.guildId}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row sched-row';
    div.innerHTML = `
      <div>${row.channel_id}</div>
      <div>${row.interval_minutes}m</div>
      <div>${formatDate(row.next_run_at)}</div>
      <div>${row.enabled ? 'yes' : 'no'}</div>
      <div>${row.content}</div>
      <div><button class="ghost" data-sched-del="${row.id}">Delete</button></div>
    `;
    table.appendChild(div);
  });
}

async function addScheduledMessage() {
  const restore = setBusy(qs('#schedAdd'), 'Adding...');
  const status = qs('#schedMsgStatus');
  status.textContent = 'Adding...';
  try {
    const payload = {
      channel_id: qs('#schedChannelId').value.trim(),
      interval_minutes: parseInt(qs('#schedInterval').value, 10),
      content: qs('#schedContent').value.trim(),
      enabled: qs('#schedEnabled').value === 'true',
    };
    await apiFetch(`/api/modules/scheduled/messages?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    status.textContent = `Added at ${new Date().toLocaleTimeString()}`;
    showToast('Schedule added.');
    await loadScheduledMessages();
  } catch (err) {
    status.textContent = 'Add failed.';
    showToast(`Add schedule failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function deleteScheduledMessage(id) {
  if (!id) return;
  await apiFetch(`/api/modules/scheduled/messages/${id}?guild_id=${state.guildId}`, { method: 'DELETE' });
}

async function saveTicketsModule() {
  const restore = setBusy(qs('#ticketsSave'), 'Saving...');
  const status = qs('#ticketsStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_TICKETS]: qs('#settingsTicketsEnabled').value === 'true',
      },
      ticket_inbox_channel_id: qs('#settingsTicketInbox').value.trim(),
      ticket_category_id: qs('#settingsTicketCategory').value.trim(),
      ticket_support_role_id: qs('#settingsTicketSupportRole').value.trim(),
      ticket_log_channel_id: qs('#settingsTicketLogChannel').value.trim(),
      ticket_open_phrase: qs('#settingsTicketOpenPhrase').value.trim(),
      ticket_close_phrase: qs('#settingsTicketClosePhrase').value.trim(),
      ticket_auto_close_minutes: parseInt(qs('#settingsTicketAutoClose').value, 10) || 0,
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Tickets module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Tickets save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function loadTickets() {
  const table = qs('#ticketsTable');
  if (!table || !state.guildId) return;
  const status = qs('#ticketStatusFilter')?.value || '';
  const rows = (await apiFetch(`/api/modules/tickets?guild_id=${state.guildId}&status=${encodeURIComponent(status)}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row ticket-row';
    div.innerHTML = `
      <div>${row.id}</div>
      <div>${row.creator_user_id}</div>
      <div>${row.channel_id}</div>
      <div>${row.subject || ''}</div>
      <div>${row.status}</div>
      <div>${formatDate(row.created_at)}</div>
      <div>
        ${row.status === 'open' ? `<button class="ghost" data-ticket-close="${row.id}">Close</button>` : ''}
        <button class="ghost" data-ticket-transcript="${row.id}">Transcript</button>
      </div>
    `;
    table.appendChild(div);
  });
}

async function closeTicket(id) {
  if (!id) return;
  await apiFetch(`/api/modules/tickets/${id}/close?guild_id=${state.guildId}`, { method: 'POST' });
}

async function loadTicketTranscript(id) {
  if (!id) return;
  const res = await apiFetch(`/api/modules/tickets/${id}/transcript?guild_id=${state.guildId}`);
  qs('#ticketTranscript').textContent = res.transcript || '';
}

async function saveAntiRaidModule() {
  const restore = setBusy(qs('#antiRaidSave'), 'Saving...');
  const status = qs('#antiRaidStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_ANTI_RAID]: qs('#settingsAntiRaidEnabled').value === 'true',
      },
      anti_raid_join_threshold: parseInt(qs('#settingsAntiRaidThreshold').value, 10),
      anti_raid_window_seconds: parseInt(qs('#settingsAntiRaidWindow').value, 10),
      anti_raid_cooldown_minutes: parseInt(qs('#settingsAntiRaidCooldown').value, 10),
      anti_raid_action: qs('#settingsAntiRaidAction').value,
      anti_raid_alert_channel_id: qs('#settingsAntiRaidAlertChannel').value.trim(),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Anti-raid module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Anti-raid save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveAnalyticsModule() {
  const restore = setBusy(qs('#analyticsSave'), 'Saving...');
  const status = qs('#analyticsStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_ANALYTICS]: qs('#settingsAnalyticsEnabled').value === 'true',
      },
      analytics_channel_id: qs('#settingsAnalyticsChannel').value.trim(),
      analytics_interval_days: parseInt(qs('#settingsAnalyticsIntervalDays').value, 10),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Analytics module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Analytics save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveAppealsModule() {
  const restore = setBusy(qs('#appealsSave'), 'Saving...');
  const status = qs('#appealsStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_APPEALS]: qs('#settingsAppealsEnabled').value === 'true',
      },
      appeals_channel_id: qs('#settingsAppealsChannel').value.trim(),
      appeals_log_channel_id: qs('#settingsAppealsLogChannel').value.trim(),
      appeals_open_phrase: qs('#settingsAppealsPhrase').value.trim(),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Appeals module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Appeals save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveStarboardModule() {
  const restore = setBusy(qs('#starboardSave'), 'Saving...');
  const status = qs('#starboardStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_STARBOARD]: qs('#settingsStarboardEnabled').value === 'true',
      },
      starboard_channel_id: qs('#settingsStarboardChannel').value.trim(),
      starboard_emoji: qs('#settingsStarboardEmoji').value.trim() || '⭐',
      starboard_threshold: parseInt(qs('#settingsStarboardThreshold').value, 10) || 3,
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Starboard module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Starboard save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveLevelingModule() {
  const restore = setBusy(qs('#levelingSave'), 'Saving...');
  const status = qs('#levelingStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_LEVELING]: qs('#settingsLevelingEnabled').value === 'true',
      },
      leveling_channel_id: qs('#settingsLevelingChannel').value.trim(),
      leveling_xp_per_message: parseInt(qs('#settingsLevelingXP').value, 10) || 10,
      leveling_cooldown_seconds: parseInt(qs('#settingsLevelingCooldown').value, 10) || 60,
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Leveling module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Leveling save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveGiveawaysModule() {
  const restore = setBusy(qs('#giveawaysSave'), 'Saving...');
  const status = qs('#giveawaysStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_GIVEAWAYS]: qs('#settingsGiveawaysEnabled').value === 'true',
      },
      giveaways_channel_id: qs('#settingsGiveawaysChannel').value.trim(),
      giveaways_reaction_emoji: qs('#settingsGiveawaysEmoji').value.trim() || '🎉',
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Giveaways module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Giveaways save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function savePollsModule() {
  const restore = setBusy(qs('#pollsSave'), 'Saving...');
  const status = qs('#pollsStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_POLLS]: qs('#settingsPollsEnabled').value === 'true',
      },
      polls_channel_id: qs('#settingsPollsChannel').value.trim(),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Polls module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Polls save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveSuggestionsModule() {
  const restore = setBusy(qs('#suggestionsSave'), 'Saving...');
  const status = qs('#suggestionsStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_SUGGESTIONS]: qs('#settingsSuggestionsEnabled').value === 'true',
      },
      suggestions_channel_id: qs('#settingsSuggestionsChannel').value.trim(),
      suggestions_log_channel_id: qs('#settingsSuggestionsLogChannel').value.trim(),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Suggestions module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Suggestions save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveKeywordAlertsModule() {
  const restore = setBusy(qs('#keywordAlertsSave'), 'Saving...');
  const status = qs('#keywordAlertsStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_KEYWORD_ALERTS]: qs('#settingsKeywordAlertsEnabled').value === 'true',
      },
      keyword_alerts_channel_id: qs('#settingsKeywordAlertsChannel').value.trim(),
      keyword_alert_words: qs('#settingsKeywordAlertWords').value.split(',').map((v) => v.trim()).filter(Boolean),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Keyword alerts module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Keyword alerts save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function loadGiveaways() {
  const table = qs('#giveawaysTable');
  if (!table || !state.guildId) return;
  const rows = (await apiFetch(`/api/modules/giveaways?guild_id=${state.guildId}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row giveaway-row';
    div.innerHTML = `
      <div>${row.id}</div>
      <div>${row.prize}</div>
      <div>${row.entry_count}</div>
      <div>${row.status}</div>
      <div>${formatDate(row.ends_at)}</div>
      <div>${row.status === 'open' ? `<button class="ghost" data-giveaway-draw="${row.id}">Draw</button>` : ''}</div>
    `;
    table.appendChild(div);
  });
}

async function startGiveaway() {
  const restore = setBusy(qs('#giveawayStart'), 'Starting...');
  const status = qs('#giveawaysRunStatus');
  status.textContent = 'Starting...';
  try {
    const payload = {
      channel_id: qs('#giveawayChannel').value.trim(),
      prize: qs('#giveawayPrize').value.trim(),
      duration_minutes: parseInt(qs('#giveawayDuration').value, 10) || 60,
      winner_count: parseInt(qs('#giveawayWinners').value, 10) || 1,
    };
    await apiFetch(`/api/modules/giveaways/start?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    status.textContent = `Started at ${new Date().toLocaleTimeString()}`;
    showToast('Giveaway started.');
    await loadGiveaways();
  } catch (err) {
    status.textContent = 'Start failed.';
    showToast(`Giveaway start failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function drawGiveaway(id) {
  if (!id) return;
  await apiFetch(`/api/modules/giveaways/${id}/draw?guild_id=${state.guildId}`, { method: 'POST' });
}

async function loadPolls() {
  const table = qs('#pollsTable');
  if (!table || !state.guildId) return;
  const rows = (await apiFetch(`/api/modules/polls?guild_id=${state.guildId}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row poll-row';
    div.innerHTML = `
      <div>${row.id}</div>
      <div>${row.question}</div>
      <div>${row.status}</div>
      <div>${formatDate(row.created_at)}</div>
      <div>${row.status === 'open' ? `<button class="ghost" data-poll-close="${row.id}">Close</button>` : ''}</div>
    `;
    table.appendChild(div);
  });
}

async function startPoll() {
  const restore = setBusy(qs('#pollStart'), 'Starting...');
  const status = qs('#pollsRunStatus');
  status.textContent = 'Starting...';
  try {
    const options = [
      qs('#pollOption1').value.trim(),
      qs('#pollOption2').value.trim(),
      qs('#pollOption3').value.trim(),
      qs('#pollOption4').value.trim(),
      qs('#pollOption5').value.trim(),
    ].filter(Boolean);
    const payload = {
      channel_id: qs('#pollChannel').value.trim(),
      question: qs('#pollQuestion').value.trim(),
      options,
    };
    await apiFetch(`/api/modules/polls/start?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    status.textContent = `Started at ${new Date().toLocaleTimeString()}`;
    showToast('Poll started.');
    await loadPolls();
  } catch (err) {
    status.textContent = 'Start failed.';
    showToast(`Poll start failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function closePoll(id) {
  if (!id) return;
  await apiFetch(`/api/modules/polls/${id}/close?guild_id=${state.guildId}`, { method: 'POST' });
}

async function loadSuggestions() {
  const table = qs('#suggestionsTable');
  if (!table || !state.guildId) return;
  const status = qs('#suggestionStatusFilter')?.value || '';
  const rows = (await apiFetch(`/api/modules/suggestions?guild_id=${state.guildId}&status=${encodeURIComponent(status)}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row suggestion-row';
    div.innerHTML = `
      <div>${row.id}</div>
      <div>${row.content}</div>
      <div>${row.status}</div>
      <div>${formatDate(row.created_at)}</div>
      <div>
        ${row.status === 'open' ? `<button class="ghost" data-suggestion-approve="${row.id}">Approve</button> <button class="ghost" data-suggestion-reject="${row.id}">Reject</button>` : ''}
      </div>
    `;
    table.appendChild(div);
  });
}

async function decideSuggestion(id, action) {
  if (!id) return;
  const note = prompt(`${action} note (optional):`) || '';
  await apiFetch(`/api/modules/suggestions/${id}/${action}?guild_id=${state.guildId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ note }),
  });
}

async function loadLeaderboard() {
  const table = qs('#levelingTable');
  if (!table || !state.guildId) return;
  const rows = (await apiFetch(`/api/modules/leveling/leaderboard?guild_id=${state.guildId}&limit=50`)) || [];
  table.innerHTML = '';
  rows.forEach((row, idx) => {
    const div = document.createElement('div');
    div.className = 'table-row leveling-row';
    div.innerHTML = `
      <div>${idx + 1}</div>
      <div>${row.username || row.user_id}</div>
      <div>${row.level}</div>
      <div>${row.xp}</div>
      <div>${formatDate(row.last_xp_at)}</div>
    `;
    table.appendChild(div);
  });
}

async function loadAppeals() {
  const table = qs('#appealsTable');
  if (!table || !state.guildId) return;
  const status = qs('#appealStatusFilter')?.value || '';
  const rows = (await apiFetch(`/api/modules/appeals?guild_id=${state.guildId}&status=${encodeURIComponent(status)}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row appeal-row';
    div.innerHTML = `
      <div>${row.id}</div>
      <div>${row.user_id}</div>
      <div>${row.reason || ''}</div>
      <div>${row.status}</div>
      <div>${formatDate(row.created_at)}</div>
      <div>
        ${row.status === 'open' ? `<button class="ghost" data-appeal-resolve="${row.id}">Resolve</button>` : ''}
      </div>
    `;
    table.appendChild(div);
  });
}

async function resolveAppeal(id) {
  if (!id) return;
  const resolution = prompt('Resolution notes (optional):') || '';
  await apiFetch(`/api/modules/appeals/${id}/resolve?guild_id=${state.guildId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ resolution }),
  });
}

async function saveCustomCommandsModule() {
  const restore = setBusy(qs('#customCommandsSave'), 'Saving...');
  const status = qs('#customCommandsStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_CUSTOM_COMMANDS]: qs('#settingsCustomCommandsEnabled').value === 'true',
      },
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Custom commands module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Custom commands save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function loadCustomCommands() {
  const table = qs('#customCommandsTable');
  if (!table || !state.guildId) return;
  const rows = (await apiFetch(`/api/modules/custom-commands/commands?guild_id=${state.guildId}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row custom-command-row';
    div.innerHTML = `
      <div>${row.trigger}</div>
      <div>${row.response}</div>
      <div>${formatDate(row.created_at)}</div>
      <div><button class="ghost" data-cc-delete="${row.id}">Delete</button></div>
    `;
    table.appendChild(div);
  });
}

async function addCustomCommand() {
  const restore = setBusy(qs('#customCommandAdd'), 'Adding...');
  const status = qs('#customCommandEditStatus');
  status.textContent = 'Adding...';
  try {
    const payload = {
      trigger: qs('#customCommandTrigger').value.trim(),
      response: qs('#customCommandResponse').value.trim(),
    };
    await apiFetch(`/api/modules/custom-commands/commands?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    status.textContent = `Added at ${new Date().toLocaleTimeString()}`;
    showToast('Custom command added.');
    qs('#customCommandTrigger').value = '';
    qs('#customCommandResponse').value = '';
    await loadCustomCommands();
  } catch (err) {
    status.textContent = 'Add failed.';
    showToast(`Add command failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function deleteCustomCommand(id) {
  if (!id) return;
  await apiFetch(`/api/modules/custom-commands/commands/${id}?guild_id=${state.guildId}`, { method: 'DELETE' });
}

function formatDate(value) {
  if (!value) return '—';
  const date = new Date(value);
  return date.toLocaleString();
}

function renderLastMessageCell(row) {
  if (!row.last_message_at) {
    return '<span class="muted">—</span> <span class="meta-badge">No messages recorded</span>';
  }
  return formatDate(row.last_message_at);
}

function escapeAttr(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('"', '&quot;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;');
}

async function loadMembers() {
  if (!state.guildId) return;
  const status = qs('#memberStatus').value;
  const search = qs('#memberSearch').value.trim();
  const query = new URLSearchParams({ guild_id: state.guildId, status, search, limit: 200 }).toString();
  const rows = (await apiFetch(`/api/members?${query}`)) || [];
  const visibleRows = status ? rows.filter((row) => row.status === status) : rows;
  const table = qs('#membersTable');
  table.innerHTML = '';
  state.selectedUsers.clear();
  updateSelectedCount();
  qs('#selectAllMembers').checked = false;
  visibleRows.forEach((row) => {
    const div = document.createElement('div');
    div.className = `table-row ${row.quarantined ? 'table-row-quarantined' : ''}`;
    const name = row.display_name || row.username || 'Unknown User';
    const quarantineBadge = row.quarantined ? '<span class="meta-badge quarantine-badge">Quarantined</span>' : '';
    const safeName = escapeAttr(name);
    div.innerHTML = `
      <div>
        <input type="checkbox" class="member-select" data-user="${row.user_id}" data-name="${safeName}" />
      </div>
      <div>
        <div>${name} ${quarantineBadge}</div>
      </div>
      <div>${renderLastMessageCell(row)}</div>
      <div><span class="status-pill ${row.status}">${row.status}</span></div>
      <div>
        <button class="ghost" data-action="quarantine" data-user="${row.user_id}" data-name="${safeName}">Quarantine</button>
        <button class="ghost" data-action="remove-roles" data-user="${row.user_id}" data-name="${safeName}">Remove Roles (Allowlist)</button>
        <button class="ghost" data-action="kick" data-user="${row.user_id}" data-name="${safeName}">Kick</button>
      </div>
    `;
    table.appendChild(div);
  });
}

async function loadActions() {
  if (!state.guildId) return;
  const status = qs('#actionStatus').value;
  const query = new URLSearchParams({ guild_id: state.guildId, status, limit: 100 }).toString();
  const rows = (await apiFetch(`/api/actions?${query}`)) || [];
  const table = qs('#actionsTable');
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row';
    const target = row.target_name || 'Unknown User';
    div.innerHTML = `
      <div>#${row.id}</div>
      <div>${target}</div>
      <div>${row.type}</div>
      <div>${row.status}</div>
      <div>${formatDate(row.updated_at)}</div>
    `;
    table.appendChild(div);
  });
}

async function loadOverview() {
  if (!state.guildId) return;
  const members = (await apiFetch(`/api/members?guild_id=${state.guildId}&limit=200`)) || [];
  const inactive = members.filter((m) => m.status === 'inactive').length;
  qs('#statTracked').textContent = members.length;
  qs('#statInactive').textContent = inactive;

  const queued = (await apiFetch(`/api/actions?guild_id=${state.guildId}&status=queued&limit=50`)) || [];
  qs('#statQueued').textContent = queued.length;

  const backfills = (await apiFetch('/api/backfill/status')) || [];
  const list = qs('#backfillList');
  list.innerHTML = '';
  backfills.forEach((job) => {
    const div = document.createElement('div');
    div.className = 'list-item';
    const skipped = job.skipped_channels || 0;
    div.textContent = `${job.guild_id} · ${job.status} · ${job.scanned_channels}/${job.total_channels} channels · ${job.checked_messages} msgs · ${job.updated_users} users · ${skipped} skipped`;
    list.appendChild(div);
  });
  syncOverviewPolling(backfills);
}

async function loadEvents() {
  const limit = parseInt(qs('#eventsLimit').value, 10) || 200;
  const rows = (await apiFetch(`/api/events?limit=${Math.min(Math.max(limit, 20), 1000)}`)) || [];
  qs('#eventsLog').textContent = rows.join('\n');
}

async function runBackfill() {
  const restore = setBusy(qs('#backfillBtn'), 'Starting...');
  const status = qs('#overviewStatus');
  status.textContent = 'Starting backfill...';
  try {
    const res = await apiFetch(`/api/backfill/start?guild_id=${state.guildId}`, { method: 'POST' });
    status.textContent = `Backfill started (${res.job_id || 'job created'}).`;
    showToast('Backfill started.');
    await loadOverview();
  } catch (err) {
    status.textContent = 'Backfill start failed.';
    showToast(`Backfill failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function createAction(userId, type, targetName) {
  if (!userId) return;
  const reason = prompt(`Reason for ${type} (optional):`);
  try {
    await apiFetch(`/api/actions/${type}?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_ids: [userId], reason: reason || '', target_name: targetName || '' }),
    });
    showToast(`Action queued: ${type}`);
    await loadActions();
  } catch (err) {
    showToast(`Action failed: ${err.message}`, 'error');
  }
}

async function createBulkAction(selectedUserMap, type) {
  const userIds = Array.from(selectedUserMap.keys());
  if (!userIds.length) {
    showToast('Select at least one member first.', 'error');
    return;
  }
  const reason = prompt(`Reason for ${type} (optional):`);
  const payload = { user_ids: userIds, reason: reason || '', target_names: Object.fromEntries(selectedUserMap) };
  if (type === 'remove-roles') {
    payload.remove_all_except_allowlist = true;
  }
  try {
    await apiFetch(`/api/actions/${type}?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    state.selectedUsers.clear();
    updateSelectedCount();
    showToast(`Bulk action queued: ${type} (${userIds.length})`);
    await loadActions();
    await loadMembers();
  } catch (err) {
    showToast(`Bulk action failed: ${err.message}`, 'error');
  }
}

function updateSelectedCount() {
  qs('#selectedCount').textContent = `${state.selectedUsers.size} selected`;
}

function wireEvents() {
  const reloadMembersForFilters = () => {
    loadMembers().catch((err) => showToast(`Members load failed: ${err.message}`, 'error'));
  };

  qs('#loginBtn').onclick = login;
  qs('#logoutBtn').onclick = () => {
    stopOverviewPolling();
    stopEventsPolling();
    state.token = '';
    localStorage.removeItem('modbot_token');
    showLogin();
  };
  qs('#settingsSave').onclick = saveSettings;
  qs('#welcomeSave').onclick = saveWelcome;
  qs('#goodbyeSave').onclick = saveGoodbye;
  qs('#auditSave').onclick = saveAudit;
  qs('#inviteSave').onclick = saveInviteTracker;
  qs('#automodSave').onclick = saveAutoMod;
  qs('#reactionRolesSave').onclick = saveReactionRoles;
  qs('#warningsSave').onclick = saveWarningsModule;
  qs('#scheduledSave').onclick = saveScheduledModule;
  qs('#verificationSave').onclick = saveVerificationModule;
  qs('#ticketsSave').onclick = saveTicketsModule;
  qs('#antiRaidSave').onclick = saveAntiRaidModule;
  qs('#analyticsSave').onclick = saveAnalyticsModule;
  qs('#appealsSave').onclick = saveAppealsModule;
  qs('#starboardSave').onclick = saveStarboardModule;
  qs('#levelingSave').onclick = saveLevelingModule;
  qs('#giveawaysSave').onclick = saveGiveawaysModule;
  qs('#pollsSave').onclick = savePollsModule;
  qs('#suggestionsSave').onclick = saveSuggestionsModule;
  qs('#keywordAlertsSave').onclick = saveKeywordAlertsModule;
  qs('#customCommandsSave').onclick = saveCustomCommandsModule;
  qs('#rrRefresh').onclick = () => loadReactionRoleRules().catch((err) => showToast(`Rule load failed: ${err.message}`, 'error'));
  qs('#rrAddRule').onclick = addReactionRoleRule;
  qs('#warnRefresh').onclick = () => loadWarnings().catch((err) => showToast(`Warnings load failed: ${err.message}`, 'error'));
  qs('#warnIssue').onclick = issueWarning;
  qs('#schedRefresh').onclick = () => loadScheduledMessages().catch((err) => showToast(`Schedules load failed: ${err.message}`, 'error'));
  qs('#schedAdd').onclick = addScheduledMessage;
  qs('#ticketsRefresh').onclick = () => loadTickets().catch((err) => showToast(`Tickets load failed: ${err.message}`, 'error'));
  qs('#ticketStatusFilter').addEventListener('change', () => loadTickets().catch((err) => showToast(`Tickets load failed: ${err.message}`, 'error')));
  qs('#appealsRefresh').onclick = () => loadAppeals().catch((err) => showToast(`Appeals load failed: ${err.message}`, 'error'));
  qs('#appealStatusFilter').addEventListener('change', () => loadAppeals().catch((err) => showToast(`Appeals load failed: ${err.message}`, 'error')));
  qs('#customCommandsRefresh').onclick = () => loadCustomCommands().catch((err) => showToast(`Commands load failed: ${err.message}`, 'error'));
  qs('#customCommandAdd').onclick = addCustomCommand;
  qs('#levelingRefresh').onclick = () => loadLeaderboard().catch((err) => showToast(`Leaderboard load failed: ${err.message}`, 'error'));
  qs('#giveawaysRefresh').onclick = () => loadGiveaways().catch((err) => showToast(`Giveaways load failed: ${err.message}`, 'error'));
  qs('#giveawayStart').onclick = startGiveaway;
  qs('#pollsRefresh').onclick = () => loadPolls().catch((err) => showToast(`Polls load failed: ${err.message}`, 'error'));
  qs('#pollStart').onclick = startPoll;
  qs('#suggestionsRefresh').onclick = () => loadSuggestions().catch((err) => showToast(`Suggestions load failed: ${err.message}`, 'error'));
  qs('#suggestionStatusFilter').addEventListener('change', () => loadSuggestions().catch((err) => showToast(`Suggestions load failed: ${err.message}`, 'error')));
  qs('#settingsWelcomeEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsGoodbyeEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsAuditEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsInviteEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsAutoModEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsReactionRolesEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsWarningsEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsScheduledEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsVerificationEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsTicketsEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsAntiRaidEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsAnalyticsEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsStarboardEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsLevelingEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsGiveawaysEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsPollsEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsSuggestionsEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsKeywordAlertsEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsAppealsEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsCustomCommandsEnabled').addEventListener('change', syncModuleBadges);
  qs('#memberRefresh').onclick = loadMembers;
  qs('#memberStatus').addEventListener('change', reloadMembersForFilters);
  qs('#memberStatus').addEventListener('input', reloadMembersForFilters);
  qs('#memberStatus').addEventListener('click', () => {
    setTimeout(reloadMembersForFilters, 0);
  });
  qs('#actionRefresh').onclick = loadActions;
  qs('#eventsRefresh').onclick = () => loadEvents().catch((err) => showToast(`Events load failed: ${err.message}`, 'error'));
  qs('#backfillBtn').onclick = runBackfill;
  qs('#refreshOverview').onclick = loadOverview;

  qs('#membersTable').addEventListener('click', (e) => {
    const btn = e.target.closest('button[data-action]');
    if (!btn) return;
    const userId = btn.getAttribute('data-user');
    const targetName = btn.getAttribute('data-name') || '';
    const type = btn.getAttribute('data-action');
    createAction(userId, type, targetName);
  });

  qs('#rrRulesTable').addEventListener('click', async (e) => {
    const btn = e.target.closest('button[data-rr-delete]');
    if (!btn) return;
    try {
      await deleteReactionRoleRule(btn.getAttribute('data-rr-delete'));
      showToast('Reaction role rule deleted.');
      await loadReactionRoleRules();
    } catch (err) {
      showToast(`Delete rule failed: ${err.message}`, 'error');
    }
  });

  qs('#scheduledTable').addEventListener('click', async (e) => {
    const btn = e.target.closest('button[data-sched-del]');
    if (!btn) return;
    try {
      await deleteScheduledMessage(btn.getAttribute('data-sched-del'));
      showToast('Schedule deleted.');
      await loadScheduledMessages();
    } catch (err) {
      showToast(`Delete schedule failed: ${err.message}`, 'error');
    }
  });

  qs('#ticketsTable').addEventListener('click', async (e) => {
    const closeBtn = e.target.closest('button[data-ticket-close]');
    if (closeBtn) {
      try {
        await closeTicket(closeBtn.getAttribute('data-ticket-close'));
        showToast('Ticket closed.');
        await loadTickets();
      } catch (err) {
        showToast(`Close ticket failed: ${err.message}`, 'error');
      }
      return;
    }
    const transcriptBtn = e.target.closest('button[data-ticket-transcript]');
    if (transcriptBtn) {
      try {
        await loadTicketTranscript(transcriptBtn.getAttribute('data-ticket-transcript'));
      } catch (err) {
        showToast(`Load transcript failed: ${err.message}`, 'error');
      }
    }
  });

  qs('#appealsTable').addEventListener('click', async (e) => {
    const resolveBtn = e.target.closest('button[data-appeal-resolve]');
    if (!resolveBtn) return;
    try {
      await resolveAppeal(resolveBtn.getAttribute('data-appeal-resolve'));
      showToast('Appeal resolved.');
      await loadAppeals();
    } catch (err) {
      showToast(`Resolve appeal failed: ${err.message}`, 'error');
    }
  });

  qs('#customCommandsTable').addEventListener('click', async (e) => {
    const delBtn = e.target.closest('button[data-cc-delete]');
    if (!delBtn) return;
    try {
      await deleteCustomCommand(delBtn.getAttribute('data-cc-delete'));
      showToast('Custom command deleted.');
      await loadCustomCommands();
    } catch (err) {
      showToast(`Delete command failed: ${err.message}`, 'error');
    }
  });

  qs('#giveawaysTable').addEventListener('click', async (e) => {
    const drawBtn = e.target.closest('button[data-giveaway-draw]');
    if (!drawBtn) return;
    try {
      await drawGiveaway(drawBtn.getAttribute('data-giveaway-draw'));
      showToast('Giveaway drawn.');
      await loadGiveaways();
    } catch (err) {
      showToast(`Giveaway draw failed: ${err.message}`, 'error');
    }
  });

  qs('#pollsTable').addEventListener('click', async (e) => {
    const closeBtn = e.target.closest('button[data-poll-close]');
    if (!closeBtn) return;
    try {
      await closePoll(closeBtn.getAttribute('data-poll-close'));
      showToast('Poll closed.');
      await loadPolls();
    } catch (err) {
      showToast(`Poll close failed: ${err.message}`, 'error');
    }
  });

  qs('#suggestionsTable').addEventListener('click', async (e) => {
    const approveBtn = e.target.closest('button[data-suggestion-approve]');
    if (approveBtn) {
      try {
        await decideSuggestion(approveBtn.getAttribute('data-suggestion-approve'), 'approve');
        showToast('Suggestion approved.');
        await loadSuggestions();
      } catch (err) {
        showToast(`Suggestion action failed: ${err.message}`, 'error');
      }
      return;
    }
    const rejectBtn = e.target.closest('button[data-suggestion-reject]');
    if (rejectBtn) {
      try {
        await decideSuggestion(rejectBtn.getAttribute('data-suggestion-reject'), 'reject');
        showToast('Suggestion rejected.');
        await loadSuggestions();
      } catch (err) {
        showToast(`Suggestion action failed: ${err.message}`, 'error');
      }
    }
  });

  qs('#membersTable').addEventListener('change', (e) => {
    const checkbox = e.target.closest('.member-select');
    if (!checkbox) return;
    const userId = checkbox.getAttribute('data-user');
    const name = checkbox.getAttribute('data-name') || 'Unknown User';
    if (checkbox.checked) {
      state.selectedUsers.set(userId, name);
    } else {
      state.selectedUsers.delete(userId);
    }
    updateSelectedCount();
  });

  qs('#selectAllMembers').addEventListener('change', (e) => {
    const checked = e.target.checked;
    qsa('.member-select').forEach((cb) => {
      cb.checked = checked;
      const userId = cb.getAttribute('data-user');
      const name = cb.getAttribute('data-name') || 'Unknown User';
      if (checked) {
        state.selectedUsers.set(userId, name);
      } else {
        state.selectedUsers.delete(userId);
      }
    });
    updateSelectedCount();
  });

  qs('#bulkQuarantine').onclick = () => createBulkAction(state.selectedUsers, 'quarantine');
  qs('#bulkKick').onclick = () => createBulkAction(state.selectedUsers, 'kick');
  qs('#bulkRemoveRoles').onclick = () => createBulkAction(state.selectedUsers, 'remove-roles');

  qsa('.nav button').forEach((btn) => {
    btn.onclick = () => {
      qsa('.nav button').forEach((b) => b.classList.remove('active'));
      btn.classList.add('active');
      const view = btn.getAttribute('data-view');
      qsa('.view').forEach((v) => v.classList.remove('active'));
      qs(`#view-${view}`).classList.add('active');
      if (view === 'events') {
        loadEvents().catch((err) => showToast(`Events load failed: ${err.message}`, 'error'));
        startEventsPolling();
      } else {
        stopEventsPolling();
      }
    };
  });

  startMemberFilterWatch();
}

async function refreshAll() {
  await loadOverview();
  await loadMembers();
  await loadActions();
  await loadEvents();
  await loadSettings();
  await loadReactionRoleRules();
  await loadWarnings();
  await loadScheduledMessages();
  await loadTickets();
  await loadAppeals();
  await loadCustomCommands();
  await loadLeaderboard();
  await loadGiveaways();
  await loadPolls();
  await loadSuggestions();
}

async function bootstrap() {
  await loadGuilds();
  await refreshAll();
}

wireEvents();
if (!state.token) {
  showLogin();
} else {
  bootstrap().catch(showLogin);
}
