const state = {
  token: localStorage.getItem('modbot_token') || '',
  guildId: localStorage.getItem('modbot_guild') || '',
  guilds: [],
  dashboardRole: localStorage.getItem('modbot_dashboard_role') || 'admin',
  currentSettings: null,
  modulePermissions: {},
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
const FEATURE_AFK = 'afk';
const FEATURE_REMINDERS = 'reminders';
const FEATURE_ACCOUNT_AGE_GUARD = 'account_age_guard';
const FEATURE_MEMBER_NOTES = 'member_notes';
const FEATURE_APPEALS = 'appeals';
const FEATURE_CUSTOM_COMMANDS = 'custom_commands';
const FEATURE_BY_VIEW = {
  welcome: FEATURE_WELCOME,
  goodbye: FEATURE_GOODBYE,
  auditlog: FEATURE_AUDIT,
  invites: FEATURE_INVITE,
  automod: FEATURE_AUTOMOD,
  reactionroles: FEATURE_REACTION_ROLES,
  warnings: FEATURE_WARNINGS,
  scheduled: FEATURE_SCHEDULED,
  verification: FEATURE_VERIFICATION,
  tickets: FEATURE_TICKETS,
  antiraid: FEATURE_ANTI_RAID,
  analytics: FEATURE_ANALYTICS,
  starboard: FEATURE_STARBOARD,
  leveling: FEATURE_LEVELING,
  giveaways: FEATURE_GIVEAWAYS,
  polls: FEATURE_POLLS,
  suggestions: FEATURE_SUGGESTIONS,
  keywordalerts: FEATURE_KEYWORD_ALERTS,
  afk: FEATURE_AFK,
  reminders: FEATURE_REMINDERS,
  accountageguard: FEATURE_ACCOUNT_AGE_GUARD,
  membernotes: FEATURE_MEMBER_NOTES,
  appeals: FEATURE_APPEALS,
  customcommands: FEATURE_CUSTOM_COMMANDS,
};
const NAV_GROUPS_STORAGE_KEY = 'modbot_nav_groups';
const ACTIVE_VIEW_STORAGE_KEY = 'modbot_active_view';
const THEME_STORAGE_KEY = 'modbot_theme';
const MODULE_GUIDES = {
  welcome: { title: 'How To Use', points: ['Enable the module and set a channel ID.', 'Use {user} and {server} tokens in the message template.', 'Save, then test with a new account join.'] },
  goodbye: { title: 'How To Use', points: ['Enable and set a goodbye channel ID.', 'Tune the message template to match your community tone.', 'Save and verify with a member leave event.'] },
  auditlog: { title: 'How To Use', points: ['Set the audit log channel ID first.', 'Keep only event types you care about in the list.', 'Use this as the central trail for moderation actions.'] },
  invites: { title: 'How To Use', points: ['Set an invite log channel and enable the module.', 'Ensure bot has Manage Server permission in the guild.', 'Expect one warm-up join after restart before exact attribution.'] },
  automod: { title: 'How To Use', points: ['Start with delete_warn for safe rollout.', 'Add blocked words and duplicate thresholds gradually.', 'Use ignored channels/roles to avoid staff workflow conflicts.'] },
  reactionroles: { title: 'How To Use', points: ['Enable module, then add one rule per message/emoji mapping.', 'Use the exact message ID and channel ID from Discord.', 'Set remove-on-unreact if roles should be reversible.'] },
  warnings: { title: 'How To Use', points: ['Enable and set optional warning log channel.', 'Issue warnings from the panel below to track history.', 'Configure quarantine/kick thresholds for auto-escalation.'] },
  scheduled: { title: 'How To Use', points: ['Enable module and create recurring messages below.', 'Use conservative intervals at first to validate behavior.', 'Delete schedules that are no longer relevant.'] },
  verification: { title: 'How To Use', points: ['Set verification channel + unverified role ID.', 'Keep phrase short and easy to type.', 'Optionally set verified role for post-verification assignment.'] },
  tickets: { title: 'How To Use', points: ['Configure inbox channel, category, and support role.', 'Users open with the open phrase; staff/creator close via close phrase.', 'Set auto-close minutes to clean stale tickets automatically.'] },
  antiraid: { title: 'How To Use', points: ['Set join threshold/window/cooldown to your server baseline.', 'Use verification_only first, then quarantine if needed.', 'Set alert channel so staff can react quickly during spikes.'] },
  analytics: { title: 'How To Use', points: ['Enable module and set report channel ID.', 'Choose a weekly interval first for signal over noise.', 'Use reports to tune inactivity, warnings, and action policies.'] },
  starboard: { title: 'How To Use', points: ['Set starboard channel + emoji + threshold.', 'Avoid setting threshold too low to prevent noise.', 'Verify the configured emoji matches your community usage.'] },
  leveling: { title: 'How To Use', points: ['Set XP per message and cooldown to control XP velocity.', 'Choose curve + base to define XP needed per level.', 'Use leaderboard refresh to verify progression behavior.'] },
  giveaways: { title: 'How To Use', points: ['Set default channel and entry emoji.', 'Start giveaways from the run panel below.', 'Draw winners after end time to announce results.'] },
  polls: { title: 'How To Use', points: ['Set default poll channel and enable module.', 'Create polls with 2-5 options.', 'Close polls from the table to publish final vote summary.'] },
  suggestions: { title: 'How To Use', points: ['Set suggestions channel (and optional log channel).', 'Users post suggestions; bot converts them into vote cards.', 'Approve/reject from the table and include moderation notes.'] },
  keywordalerts: { title: 'How To Use', points: ['Set alert channel and comma-separated keywords.', 'Use specific terms to reduce false positives.', 'Review jump links from alerts for context before acting.'] },
  afk: { title: 'How To Use', points: ['Set the AFK phrase (default !afk).', 'Users set AFK with optional reason.', 'Bot auto-clears AFK when users send a new message.'] },
  reminders: { title: 'How To Use', points: ['Set default reminder channel (optional).', 'Create one-time reminders with exact run time below.', 'Worker sends due reminders and marks them sent.'] },
  accountageguard: { title: 'How To Use', points: ['Set minimum account age in days.', 'Start with log_only to observe impact.', 'Escalate to quarantine/kick once thresholds are validated.'] },
  membernotes: { title: 'How To Use', points: ['Enable and optionally set notes log channel.', 'Add moderation notes per user from the panel below.', 'Resolve notes when issues are closed out.'] },
  appeals: { title: 'How To Use', points: ['Set appeals intake channel + phrase.', 'Users submit appeals in that channel.', 'Resolve with clear outcome notes for future audits.'] },
  customcommands: { title: 'How To Use', points: ['Enable module and add trigger/response rules below.', 'Triggers are exact matches (case-insensitive).', 'Keep responses concise to avoid channel spam.'] },
};

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
  const afkEnabled = qs('#settingsAFKEnabled').value === 'true';
  const remindersEnabled = qs('#settingsRemindersEnabled').value === 'true';
  const accountAgeGuardEnabled = qs('#settingsAccountAgeGuardEnabled').value === 'true';
  const memberNotesEnabled = qs('#settingsMemberNotesEnabled').value === 'true';
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
  setModuleBadge(afkEnabled, qs('#moduleAFKBadge'), qs('#moduleAFKCard'));
  setModuleBadge(remindersEnabled, qs('#moduleRemindersBadge'), qs('#moduleRemindersCard'));
  setModuleBadge(accountAgeGuardEnabled, qs('#moduleAccountAgeGuardBadge'), qs('#moduleAccountAgeGuardCard'));
  setModuleBadge(memberNotesEnabled, qs('#moduleMemberNotesBadge'), qs('#moduleMemberNotesCard'));
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

function preferredTheme() {
  const saved = localStorage.getItem(THEME_STORAGE_KEY);
  if (saved === 'dark' || saved === 'light') {
    return saved;
  }
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
    return 'light';
  }
  return 'dark';
}

function applyTheme(theme) {
  const normalized = theme === 'light' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', normalized);
  localStorage.setItem(THEME_STORAGE_KEY, normalized);
  const select = qs('#themeSelect');
  if (select && select.value !== normalized) {
    select.value = normalized;
  }
}

function injectModuleGuides() {
  qsa('section.view[id^="view-"]').forEach((section) => {
    const grid = section.querySelector('.modules-grid');
    if (!grid) return;
    if (grid.querySelector('.module-guide-card')) return;
    const view = section.id.replace('view-', '');
    const guide = MODULE_GUIDES[view];
    if (!guide) return;

    const card = document.createElement('article');
    card.className = 'module-card module-guide-card';
    const points = (guide.points || []).map((point) => `<li>${point}</li>`).join('');
    const dynamicLeveling = view === 'leveling'
      ? '<div class="module-guide-hint" id="levelingGuideExamples"></div>'
      : '';
    card.innerHTML = `
      <div class="module-head">
        <div class="module-title">${guide.title}</div>
      </div>
      <ul class="module-guide-list">${points}</ul>
      ${dynamicLeveling}
    `;
    grid.appendChild(card);
  });
}

function xpForLevelPreview(level, curve, base) {
  if (level <= 0) return 0;
  if (curve === 'linear') return level * base;
  return level * level * base;
}

function updateLevelingGuideExamples() {
  const host = qs('#levelingGuideExamples');
  if (!host) return;
  const curve = qs('#settingsLevelingCurve')?.value || 'quadratic';
  const base = parseInt(qs('#settingsLevelingBase')?.value || '100', 10) || 100;
  const xpPerMessage = parseInt(qs('#settingsLevelingXP')?.value || '10', 10) || 10;
  const levels = [1, 2, 3, 5, 10];
  const rows = levels.map((level) => {
    const xp = xpForLevelPreview(level, curve, base);
    const msgs = Math.ceil(xp / Math.max(1, xpPerMessage));
    return `L${level}: ${xp} XP (~${msgs} msgs)`;
  });
  const curveLabel = curve === 'linear' ? 'linear' : 'quadratic';
  host.textContent = `Current curve: ${curveLabel}, base: ${base}. Milestones -> ${rows.join(' | ')}`;
}

function refreshIncidentBanner(cfg) {
  const wrap = qs('#incidentBanner');
  const text = qs('#incidentBannerText');
  if (!wrap || !text) return;
  const enabled = !!(cfg && cfg.incident_mode_enabled);
  if (!enabled) {
    wrap.classList.add('hidden');
    text.textContent = '';
    return;
  }
  const reason = (cfg.incident_mode_reason || '').trim();
  const endsAt = (cfg.incident_mode_ends_at || '').trim();
  const parts = [];
  if (reason) parts.push(`Reason: ${reason}`);
  if (endsAt) parts.push(`Ends: ${formatDate(endsAt)}`);
  text.textContent = parts.length ? parts.join(' | ') : 'High-safety restrictions are enabled for destructive actions.';
  wrap.classList.remove('hidden');
}

function modulePermissionState(featureKey) {
  if (!featureKey) return null;
  return (state.modulePermissions && state.modulePermissions[featureKey]) || null;
}

function moduleMissingPermissions(featureKey) {
  const status = modulePermissionState(featureKey);
  if (!status || status.has_all) return [];
  return status.missing_permissions || [];
}

function moduleHasPermissions(featureKey) {
  const missing = moduleMissingPermissions(featureKey);
  return missing.length === 0;
}

function requireModulePermissions(featureKey, actionLabel) {
  if (!featureKey) return true;
  const missing = moduleMissingPermissions(featureKey);
  if (!missing.length) return true;
  const text = `${actionLabel} blocked. Missing bot permissions: ${missing.join(', ')}`;
  showToast(text, 'error');
  return false;
}

function renderModulePermissionNotes() {
  qsa('section.view[id^="view-"]').forEach((section) => {
    const view = section.id.replace('view-', '');
    const featureKey = FEATURE_BY_VIEW[view];
    if (!featureKey) return;
    const card = section.querySelector('.module-card[id^="module"]');
    if (!card) return;
    let note = card.querySelector('.module-perm-note');
    if (!note) {
      note = document.createElement('p');
      note.className = 'module-note module-perm-note';
      const desc = card.querySelector('.module-desc');
      if (desc && desc.nextSibling) {
        card.insertBefore(note, desc.nextSibling);
      } else {
        card.appendChild(note);
      }
    }
    const missing = moduleMissingPermissions(featureKey);
    note.classList.remove('ok', 'warn');
    if (!missing.length) {
      note.classList.add('ok');
      note.textContent = 'Permission check: all required bot permissions are present.';
    } else {
      note.classList.add('warn');
      note.textContent = `Missing bot permissions: ${missing.join(', ')}`;
    }
  });
}

function applyModulePermissionDisabling() {
  const buttonFeatureMap = {
    welcomeSave: FEATURE_WELCOME,
    goodbyeSave: FEATURE_GOODBYE,
    auditSave: FEATURE_AUDIT,
    inviteSave: FEATURE_INVITE,
    automodSave: FEATURE_AUTOMOD,
    reactionRolesSave: FEATURE_REACTION_ROLES,
    warningsSave: FEATURE_WARNINGS,
    warnIssue: FEATURE_WARNINGS,
    scheduledSave: FEATURE_SCHEDULED,
    schedAdd: FEATURE_SCHEDULED,
    verificationSave: FEATURE_VERIFICATION,
    ticketsSave: FEATURE_TICKETS,
    antiRaidSave: FEATURE_ANTI_RAID,
    analyticsSave: FEATURE_ANALYTICS,
    starboardSave: FEATURE_STARBOARD,
    levelingSave: FEATURE_LEVELING,
    giveawaysSave: FEATURE_GIVEAWAYS,
    giveawayStart: FEATURE_GIVEAWAYS,
    pollsSave: FEATURE_POLLS,
    pollStart: FEATURE_POLLS,
    suggestionsSave: FEATURE_SUGGESTIONS,
    keywordAlertsSave: FEATURE_KEYWORD_ALERTS,
    afkSave: FEATURE_AFK,
    remindersSave: FEATURE_REMINDERS,
    reminderAdd: FEATURE_REMINDERS,
    accountAgeGuardSave: FEATURE_ACCOUNT_AGE_GUARD,
    memberNotesSave: FEATURE_MEMBER_NOTES,
    memberNoteAdd: FEATURE_MEMBER_NOTES,
    appealsSave: FEATURE_APPEALS,
    customCommandsSave: FEATURE_CUSTOM_COMMANDS,
    customCommandAdd: FEATURE_CUSTOM_COMMANDS,
  };
  Object.entries(buttonFeatureMap).forEach(([id, feature]) => {
    const button = qs(`#${id}`);
    if (!button) return;
    const missing = moduleMissingPermissions(feature);
    const blocked = missing.length > 0;
    button.disabled = blocked;
    if (blocked) {
      button.title = `Missing bot permissions: ${missing.join(', ')}`;
    } else {
      button.removeAttribute('title');
    }
  });
}

async function loadModulePermissions() {
  if (!state.guildId) return;
  try {
    const res = await apiFetch(`/api/modules/permissions?guild_id=${state.guildId}`);
    state.modulePermissions = (res && res.modules) || {};
  } catch (err) {
    state.modulePermissions = {};
    showToast(`Module permission check failed: ${err.message}`, 'error');
  }
  renderModulePermissionNotes();
  applyModulePermissionDisabling();
}

function loadNavGroupState() {
  const raw = localStorage.getItem(NAV_GROUPS_STORAGE_KEY);
  if (!raw) return {};
  try {
    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === 'object' ? parsed : {};
  } catch (_) {
    return {};
  }
}

function saveNavGroupState(groups) {
  localStorage.setItem(NAV_GROUPS_STORAGE_KEY, JSON.stringify(groups));
}

function setNavGroupExpanded(groupEl, expanded) {
  if (!groupEl) return;
  groupEl.classList.toggle('expanded', expanded);
  const toggle = groupEl.querySelector('.nav-group-toggle');
  if (toggle) {
    toggle.setAttribute('aria-expanded', expanded ? 'true' : 'false');
  }
}

function ensureViewGroupExpanded(view) {
  const btn = qs(`.nav [data-view="${view}"]`);
  if (!btn) return;
  const groupEl = btn.closest('.nav-group');
  if (!groupEl) return;
  const groupName = groupEl.getAttribute('data-group');
  setNavGroupExpanded(groupEl, true);
  const groups = loadNavGroupState();
  groups[groupName] = true;
  saveNavGroupState(groups);
}

function setActiveView(view, persist = true) {
  const targetView = qs(`#view-${view}`);
  if (!targetView) return;
  qsa('.nav [data-view]').forEach((b) => b.classList.remove('active'));
  const navBtn = qs(`.nav [data-view="${view}"]`);
  if (navBtn) {
    navBtn.classList.add('active');
  }
  qsa('.view').forEach((v) => v.classList.remove('active'));
  targetView.classList.add('active');
  ensureViewGroupExpanded(view);
  if (persist) {
    localStorage.setItem(ACTIVE_VIEW_STORAGE_KEY, view);
  }
  if (view === 'events') {
    loadEvents().catch((err) => showToast(`Events load failed: ${err.message}`, 'error'));
    startEventsPolling();
  } else {
    stopEventsPolling();
  }
}

function initNavUI() {
  const groups = loadNavGroupState();
  qsa('.nav-group').forEach((groupEl) => {
    const groupName = groupEl.getAttribute('data-group');
    const stored = groups[groupName];
    const expanded = typeof stored === 'boolean' ? stored : true;
    setNavGroupExpanded(groupEl, expanded);
  });

  qsa('.nav .nav-group-toggle').forEach((btn) => {
    btn.onclick = () => {
      const groupName = btn.getAttribute('data-group-toggle');
      const groupEl = qs(`.nav-group[data-group="${groupName}"]`);
      if (!groupEl) return;
      const expanded = !groupEl.classList.contains('expanded');
      setNavGroupExpanded(groupEl, expanded);
      const next = loadNavGroupState();
      next[groupName] = expanded;
      saveNavGroupState(next);
    };
  });

  qsa('.nav [data-view]').forEach((btn) => {
    btn.onclick = () => {
      const view = btn.getAttribute('data-view');
      setActiveView(view);
    };
  });
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
  if (state.dashboardRole) {
    headers['X-Dashboard-Role'] = state.dashboardRole;
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
  state.currentSettings = cfg;
  const flags = cfg.feature_flags || {};
  qs('#settingsInactive').value = cfg.inactive_days;
  qs('#settingsBackfill').value = cfg.backfill_days;
  qs('#settingsConcurrency').value = cfg.backfill_concurrency;
  qs('#settingsAdminPolicy').value = cfg.admin_user_policy;
  qs('#settingsQuarantineRole').value = cfg.quarantine_role_id || '';
  qs('#settingsReadmeChannel').value = cfg.readme_channel_id || '';
  qs('#settingsAllowlist').value = (cfg.allowlist_role_ids || []).join(',');
  qs('#settingsSafeMode').value = String(cfg.safe_quarantine_mode);
  qs('#settingsActionDryRun').value = String(!!cfg.action_dry_run);
  qs('#settingsActionRequireConfirm').value = String(cfg.action_require_confirm !== false);
  qs('#settingsActionTwoPerson').value = String(!!cfg.action_two_person_approval);
  qs('#settingsRolePolicies').value = JSON.stringify(cfg.dashboard_role_policies || {}, null, 2);
  qs('#settingsModuleScopes').value = JSON.stringify(cfg.module_channel_scopes || {}, null, 2);
  qs('#settingsRetentionDays').value = Number.isFinite(cfg.retention_days) ? cfg.retention_days : 0;
  qs('#settingsRetentionArchive').value = String(cfg.retention_archive_before_purge !== false);
  qs('#settingsIncidentModeEnabled').value = String(!!cfg.incident_mode_enabled);
  qs('#settingsIncidentModeReason').value = cfg.incident_mode_reason || '';
  qs('#settingsImmutableAuditTrail').value = String(!!cfg.immutable_audit_trail);
  qs('#settingsMaintenanceEnabled').value = String(!!cfg.maintenance_window_enabled);
  qs('#settingsMaintenanceStart').value = cfg.maintenance_window_start || '02:00';
  qs('#settingsMaintenanceEnd').value = cfg.maintenance_window_end || '03:00';
  qs('#settingsReviewQueueEnabled').value = String(!!cfg.review_queue_enabled);
  const incidentEndsRaw = (cfg.incident_mode_ends_at || '').trim();
  let incidentDuration = 0;
  if (incidentEndsRaw) {
    const end = new Date(incidentEndsRaw).getTime();
    const now = Date.now();
    if (Number.isFinite(end) && end > now) {
      incidentDuration = Math.ceil((end - now) / 60000);
    }
  }
  qs('#settingsIncidentModeDuration').value = incidentDuration;
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
  qs('#settingsAutoModRules').value = JSON.stringify(cfg.automod_rules || [], null, 2);
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
  qs('#settingsLevelingCurve').value = cfg.leveling_curve || 'quadratic';
  qs('#settingsLevelingBase').value = cfg.leveling_xp_base || 100;
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
  qs('#settingsAFKEnabled').value = String(!!flags[FEATURE_AFK]);
  qs('#settingsAFKPhrase').value = cfg.afk_set_phrase || '!afk';
  qs('#settingsRemindersEnabled').value = String(!!flags[FEATURE_REMINDERS]);
  qs('#settingsRemindersChannel').value = cfg.reminders_channel_id || '';
  qs('#settingsAccountAgeGuardEnabled').value = String(!!flags[FEATURE_ACCOUNT_AGE_GUARD]);
  qs('#settingsAccountAgeMinDays').value = cfg.account_age_min_days || 7;
  qs('#settingsAccountAgeAction').value = cfg.account_age_action || 'log_only';
  qs('#settingsAccountAgeLogChannel').value = cfg.account_age_log_channel_id || '';
  qs('#settingsMemberNotesEnabled').value = String(!!flags[FEATURE_MEMBER_NOTES]);
  qs('#settingsNotesLogChannel').value = cfg.notes_log_channel_id || '';
  qs('#settingsAppealsEnabled').value = String(!!flags[FEATURE_APPEALS]);
  qs('#settingsAppealsChannel').value = cfg.appeals_channel_id || '';
  qs('#settingsAppealsLogChannel').value = cfg.appeals_log_channel_id || '';
  qs('#settingsAppealsPhrase').value = cfg.appeals_open_phrase || '!appeal';
  qs('#settingsCustomCommandsEnabled').value = String(!!flags[FEATURE_CUSTOM_COMMANDS]);
  refreshIncidentBanner(cfg);
  syncModuleBadges();
  updateLevelingGuideExamples();
  await loadInvitePermissionStatus();
  await loadAuditTrail();
  await loadReactionRoleRules();
  await loadWarnings();
  await loadScheduledMessages();
  await loadTickets();
  await loadAppeals();
  await loadCustomCommands();
  await loadLeaderboard();
  await loadGiveaways();
  await loadReputationLeaderboard();
  await loadEconomy();
  await loadPolls();
  await loadSuggestions();
  await loadReminders();
  await loadMemberNotes();
  await loadDependencyChecks();
  await loadWebhooks();
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

async function loadAuditTrail() {
  if (!state.guildId) return;
  const table = qs('#auditTrailTable');
  const status = qs('#auditTrailStatus');
  if (!table || !status) return;
  status.textContent = 'Loading...';
  const rows = (await apiFetch(`/api/audit-trail?guild_id=${state.guildId}&limit=100`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const hash = (row.event_hash || '').slice(0, 12);
    const div = document.createElement('div');
    div.className = 'table-row';
    div.innerHTML = `
      <div>${formatDate(row.recorded_at)}</div>
      <div>${row.event_type || ''}</div>
      <div>${row.message || ''}</div>
      <div title="${row.event_hash || ''}">${hash}</div>
    `;
    table.appendChild(div);
  });
  status.textContent = `Loaded ${rows.length} entries`;
}

async function saveSettings() {
  const restore = setBusy(qs('#settingsSave'), 'Saving...');
  const status = qs('#settingsStatus');
  status.textContent = 'Saving...';
  try {
    let rolePolicies = {};
    const rolePoliciesRaw = qs('#settingsRolePolicies').value.trim();
    if (rolePoliciesRaw) {
      const parsed = JSON.parse(rolePoliciesRaw);
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
        throw new Error('Role policies JSON must be an object.');
      }
      rolePolicies = parsed;
    }
    let moduleScopes = {};
    const moduleScopesRaw = qs('#settingsModuleScopes').value.trim();
    if (moduleScopesRaw) {
      const parsed = JSON.parse(moduleScopesRaw);
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
        throw new Error('Module channel scopes JSON must be an object.');
      }
      moduleScopes = parsed;
    }
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const incidentModeEnabled = qs('#settingsIncidentModeEnabled').value === 'true';
    const incidentDurationMin = parseInt(qs('#settingsIncidentModeDuration').value, 10) || 0;
    let incidentEndsAt = (current.incident_mode_ends_at || '').trim();
    if (incidentModeEnabled && incidentDurationMin > 0) {
      incidentEndsAt = new Date(Date.now() + incidentDurationMin * 60000).toISOString();
    }
    if (!incidentModeEnabled) {
      incidentEndsAt = '';
    }
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
      action_dry_run: qs('#settingsActionDryRun').value === 'true',
      action_require_confirm: qs('#settingsActionRequireConfirm').value === 'true',
      action_two_person_approval: qs('#settingsActionTwoPerson').value === 'true',
      dashboard_role_policies: rolePolicies,
      module_channel_scopes: moduleScopes,
      retention_days: parseInt(qs('#settingsRetentionDays').value, 10) || 0,
      retention_archive_before_purge: qs('#settingsRetentionArchive').value === 'true',
      incident_mode_enabled: incidentModeEnabled,
      incident_mode_reason: qs('#settingsIncidentModeReason').value.trim(),
      incident_mode_ends_at: incidentEndsAt,
      immutable_audit_trail: qs('#settingsImmutableAuditTrail').value === 'true',
      maintenance_window_enabled: qs('#settingsMaintenanceEnabled').value === 'true',
      maintenance_window_start: (qs('#settingsMaintenanceStart').value || '').trim(),
      maintenance_window_end: (qs('#settingsMaintenanceEnd').value || '').trim(),
      review_queue_enabled: qs('#settingsReviewQueueEnabled').value === 'true',
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

async function applySettingsProfile() {
  const profile = (qs('#settingsProfilePreset').value || '').trim();
  if (!profile) {
    showToast('Select a profile preset first.', 'error');
    return;
  }
  const restore = setBusy(qs('#settingsApplyProfile'), 'Applying...');
  const status = qs('#settingsStatus');
  status.textContent = 'Applying profile...';
  try {
    await apiFetch(`/api/settings/profile/apply?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ profile }),
    });
    await loadSettings();
    status.textContent = `Profile applied at ${new Date().toLocaleTimeString()}`;
    showToast('Settings profile applied.');
  } catch (err) {
    status.textContent = 'Profile apply failed.';
    showToast(`Apply profile failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

function downloadExport() {
  if (!state.guildId) return;
  const type = (qs('#exportType').value || '').trim();
  const format = (qs('#exportFormat').value || 'json').trim();
  const userID = (qs('#exportCaseUserId').value || '').trim();
  if (type === 'cases' && !userID) {
    showToast('Case export requires user ID.', 'error');
    return;
  }
  const params = new URLSearchParams({
    guild_id: state.guildId,
    type,
    format,
  });
  if (userID) {
    params.set('user_id', userID);
  }
  const url = `/api/export?${params.toString()}`;
  const status = qs('#exportStatus');
  status.textContent = `Downloading ${type}.${format}...`;
  window.open(url, '_blank');
}

function downloadBackupSnapshot() {
  if (!state.guildId) return;
  const status = qs('#backupStatus');
  status.textContent = 'Downloading backup snapshot...';
  window.open(`/api/backup/export?guild_id=${encodeURIComponent(state.guildId)}`, '_blank');
}

async function restoreBackupSnapshot() {
  if (!state.guildId) return;
  const raw = (qs('#backupRestoreJson').value || '').trim();
  if (!raw) {
    showToast('Paste backup JSON first.', 'error');
    return;
  }
  const restore = setBusy(qs('#backupRestore'), 'Restoring...');
  const status = qs('#backupStatus');
  status.textContent = 'Restoring snapshot...';
  try {
    const payload = JSON.parse(raw);
    await apiFetch(`/api/backup/import?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await refreshAll();
    status.textContent = `Restore completed at ${new Date().toLocaleTimeString()}`;
    showToast('Backup restore completed.');
  } catch (err) {
    status.textContent = 'Restore failed.';
    showToast(`Backup restore failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function loadDependencyChecks() {
  if (!state.guildId) return;
  const status = qs('#dependencyCheckStatus');
  const table = qs('#dependencyCheckTable');
  if (!status || !table) return;
  status.textContent = 'Checking...';
  const res = await apiFetch(`/api/dependencies/check?guild_id=${state.guildId}`);
  const checks = (res && res.checks) || [];
  table.innerHTML = '';
  checks.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row';
    div.innerHTML = `
      <div>${row.module || ''}</div>
      <div>${row.severity || ''}</div>
      <div>${row.message || ''}</div>
    `;
    table.appendChild(div);
  });
  status.textContent = `Checked at ${new Date().toLocaleTimeString()}`;
}

async function loadWebhooks() {
  if (!state.guildId) return;
  const table = qs('#webhookTable');
  if (!table) return;
  const rows = (await apiFetch(`/api/integrations/webhooks?guild_id=${state.guildId}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row';
    const shortURL = row.url && row.url.length > 56 ? `${row.url.slice(0, 56)}...` : row.url;
    div.innerHTML = `
      <div>${row.id}</div>
      <div title="${row.url || ''}">${shortURL || ''}</div>
      <div>${(row.events || []).join(', ')}</div>
      <div>${row.enabled ? 'yes' : 'no'}</div>
      <div>${row.last_error || ''}</div>
      <div><button class="ghost" data-webhook-del="${row.id}">Delete</button></div>
    `;
    table.appendChild(div);
  });
}

async function addWebhook() {
  if (!state.guildId) return;
  const restore = setBusy(qs('#webhookAdd'), 'Adding...');
  const status = qs('#webhookStatus');
  status.textContent = 'Adding...';
  try {
    const payload = {
      url: (qs('#webhookUrl').value || '').trim(),
      events: (qs('#webhookEvents').value || '').split(',').map((v) => v.trim()).filter(Boolean),
      enabled: qs('#webhookEnabled').value === 'true',
    };
    await apiFetch(`/api/integrations/webhooks?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    qs('#webhookUrl').value = '';
    status.textContent = `Added at ${new Date().toLocaleTimeString()}`;
    await loadWebhooks();
  } catch (err) {
    status.textContent = 'Add failed.';
    showToast(`Webhook add failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function deleteWebhook(id) {
  if (!id || !state.guildId) return;
  await apiFetch(`/api/integrations/webhooks/${id}?guild_id=${state.guildId}`, { method: 'DELETE' });
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
    let advancedRules = [];
    const rawRules = qs('#settingsAutoModRules').value.trim();
    if (rawRules) {
      const parsed = JSON.parse(rawRules);
      if (!Array.isArray(parsed)) {
        throw new Error('Advanced rules JSON must be an array.');
      }
      advancedRules = parsed;
    }
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
      automod_rules: advancedRules,
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

async function loadAnalyticsTrends() {
  const table = qs('#analyticsTrendsTable');
  if (!table || !state.guildId) return;
  const days = parseInt(qs('#analyticsTrendDays')?.value || '14', 10) || 14;
  const rows = (await apiFetch(`/api/analytics/trends?guild_id=${state.guildId}&days=${days}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row analytics-trend-row';
    div.innerHTML = `
      <div>${row.day}</div>
      <div>${row.warnings}</div>
      <div>${row.actions}</div>
      <div>${row.tickets}</div>
    `;
    table.appendChild(div);
  });
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
      leveling_curve: qs('#settingsLevelingCurve').value || 'quadratic',
      leveling_xp_base: parseInt(qs('#settingsLevelingBase').value, 10) || 100,
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

async function saveAFKModule() {
  const restore = setBusy(qs('#afkSave'), 'Saving...');
  const status = qs('#afkStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_AFK]: qs('#settingsAFKEnabled').value === 'true',
      },
      afk_set_phrase: qs('#settingsAFKPhrase').value.trim() || '!afk',
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('AFK module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`AFK save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveRemindersModule() {
  const restore = setBusy(qs('#remindersSave'), 'Saving...');
  const status = qs('#remindersStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_REMINDERS]: qs('#settingsRemindersEnabled').value === 'true',
      },
      reminders_channel_id: qs('#settingsRemindersChannel').value.trim(),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Reminders module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Reminders save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveAccountAgeGuardModule() {
  const restore = setBusy(qs('#accountAgeGuardSave'), 'Saving...');
  const status = qs('#accountAgeGuardStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_ACCOUNT_AGE_GUARD]: qs('#settingsAccountAgeGuardEnabled').value === 'true',
      },
      account_age_min_days: parseInt(qs('#settingsAccountAgeMinDays').value, 10) || 7,
      account_age_action: qs('#settingsAccountAgeAction').value,
      account_age_log_channel_id: qs('#settingsAccountAgeLogChannel').value.trim(),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Account-age guard module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Account-age guard save failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function saveMemberNotesModule() {
  const restore = setBusy(qs('#memberNotesSave'), 'Saving...');
  const status = qs('#memberNotesStatus');
  status.textContent = 'Saving...';
  try {
    const current = await apiFetch(`/api/settings?guild_id=${state.guildId}`);
    const payload = {
      ...current,
      feature_flags: {
        ...(current.feature_flags || {}),
        [FEATURE_MEMBER_NOTES]: qs('#settingsMemberNotesEnabled').value === 'true',
      },
      notes_log_channel_id: qs('#settingsNotesLogChannel').value.trim(),
    };
    await apiFetch(`/api/settings?guild_id=${state.guildId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    await loadSettings();
    status.textContent = `Saved at ${new Date().toLocaleTimeString()}`;
    showToast('Member notes module saved.');
  } catch (err) {
    status.textContent = 'Save failed.';
    showToast(`Member notes save failed: ${err.message}`, 'error');
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

async function loadReminders() {
  const table = qs('#remindersTable');
  if (!table || !state.guildId) return;
  const rows = (await apiFetch(`/api/modules/reminders?guild_id=${state.guildId}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row reminder-row';
    div.innerHTML = `
      <div>${row.id}</div>
      <div>${row.content}</div>
      <div>${formatDate(row.run_at)}</div>
      <div>${row.status}</div>
    `;
    table.appendChild(div);
  });
}

async function addReminder() {
  const restore = setBusy(qs('#reminderAdd'), 'Adding...');
  const status = qs('#remindersRunStatus');
  status.textContent = 'Adding...';
  try {
    const payload = {
      channel_id: qs('#reminderChannel').value.trim(),
      content: qs('#reminderContent').value.trim(),
      run_at: new Date(qs('#reminderRunAt').value).toISOString(),
    };
    await apiFetch(`/api/modules/reminders?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    status.textContent = `Added at ${new Date().toLocaleTimeString()}`;
    showToast('Reminder queued.');
    await loadReminders();
  } catch (err) {
    status.textContent = 'Add failed.';
    showToast(`Reminder add failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function loadMemberNotes() {
  const table = qs('#memberNotesTable');
  if (!table || !state.guildId) return;
  const userId = qs('#memberNoteUserFilter')?.value.trim() || '';
  const query = userId ? `?guild_id=${state.guildId}&user_id=${encodeURIComponent(userId)}` : `?guild_id=${state.guildId}`;
  const rows = (await apiFetch(`/api/modules/member-notes${query}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row member-note-row';
    div.innerHTML = `
      <div>${row.id}</div>
      <div>${row.user_id}</div>
      <div>${row.body}</div>
      <div>${row.resolved_at ? 'resolved' : 'open'}</div>
      <div>${row.resolved_at ? '' : `<button class="ghost" data-note-resolve="${row.id}">Resolve</button>`}</div>
    `;
    table.appendChild(div);
  });
}

async function addMemberNote() {
  const restore = setBusy(qs('#memberNoteAdd'), 'Adding...');
  const status = qs('#memberNoteStatus');
  status.textContent = 'Adding...';
  try {
    const payload = {
      user_id: qs('#memberNoteUser').value.trim(),
      body: qs('#memberNoteBody').value.trim(),
    };
    await apiFetch(`/api/modules/member-notes?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    status.textContent = `Added at ${new Date().toLocaleTimeString()}`;
    showToast('Member note added.');
    await loadMemberNotes();
  } catch (err) {
    status.textContent = 'Add failed.';
    showToast(`Member note add failed: ${err.message}`, 'error');
  } finally {
    restore();
  }
}

async function resolveMemberNote(id) {
  if (!id) return;
  await apiFetch(`/api/modules/member-notes/${id}/resolve?guild_id=${state.guildId}`, { method: 'POST' });
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
  await loadReviewQueue();
}

async function loadReviewQueue() {
  if (!state.guildId) return;
  const status = qs('#reviewQueueStatus');
  const table = qs('#reviewQueueTable');
  if (!status || !table) return;
  const rows = (await apiFetch(`/api/review-queue?guild_id=${state.guildId}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const target = row.target_name || row.target_user_id || 'Unknown';
    const div = document.createElement('div');
    div.className = 'table-row';
    div.innerHTML = `
      <div>#${row.id}</div>
      <div>${target}</div>
      <div>${row.type}</div>
      <div>${formatDate(row.created_at)}</div>
      <div>
        <button class="ghost" data-review-approve="${row.id}">Approve</button>
        <button class="ghost" data-review-reject="${row.id}">Reject</button>
      </div>
    `;
    table.appendChild(div);
  });
  status.textContent = `${rows.length} pending`;
}

async function loadReputationLeaderboard() {
  if (!state.guildId) return;
  const table = qs('#repTable');
  const status = qs('#repStatus');
  if (!table || !status) return;
  const rows = (await apiFetch(`/api/modules/reputation/leaderboard?guild_id=${state.guildId}&limit=20`)) || [];
  table.innerHTML = '';
  rows.forEach((row, idx) => {
    const div = document.createElement('div');
    div.className = 'table-row';
    div.innerHTML = `
      <div>${idx + 1}</div>
      <div>${row.user_id}</div>
      <div>${row.score}</div>
    `;
    table.appendChild(div);
  });
  status.textContent = `Loaded ${rows.length} rows`;
}

async function giveReputation(delta) {
  if (!state.guildId) return;
  const from = (qs('#repFromUser')?.value || '').trim();
  const to = (qs('#repToUser')?.value || '').trim();
  if (!from || !to) {
    showToast('Enter from/to user IDs.', 'error');
    return;
  }
  await apiFetch(`/api/modules/reputation/give?guild_id=${state.guildId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ from_user_id: from, to_user_id: to, delta }),
  });
  await loadReputationLeaderboard();
}

async function loadEconomy() {
  if (!state.guildId) return;
  const status = qs('#ecoStatus');
  const board = qs('#ecoLeaderboardTable');
  const shop = qs('#ecoShopTable');
  if (!status || !board || !shop) return;
  const leaderboard = (await apiFetch(`/api/modules/economy/leaderboard?guild_id=${state.guildId}&limit=20`)) || [];
  const items = (await apiFetch(`/api/modules/economy/shop?guild_id=${state.guildId}`)) || [];
  board.innerHTML = '';
  leaderboard.forEach((row, idx) => {
    const div = document.createElement('div');
    div.className = 'table-row';
    div.innerHTML = `<div>${idx + 1}</div><div>${row.user_id}</div><div>${row.balance}</div>`;
    board.appendChild(div);
  });
  shop.innerHTML = '';
  items.forEach((item) => {
    const div = document.createElement('div');
    div.className = 'table-row';
    div.innerHTML = `<div>${item.id}</div><div>${item.name}</div><div>${item.cost}</div><div>${item.role_id || ''}</div><div>${item.enabled ? 'yes' : 'no'}</div>`;
    shop.appendChild(div);
  });
  status.textContent = `Loaded leaderboard (${leaderboard.length}) and shop (${items.length})`;
}

async function addEconomyItem() {
  if (!state.guildId) return;
  const name = (qs('#ecoNewItemName').value || '').trim();
  const cost = parseInt(qs('#ecoNewItemCost').value || '0', 10);
  const roleID = (qs('#ecoNewItemRole').value || '').trim();
  if (!name || cost <= 0) {
    showToast('Item name and cost are required.', 'error');
    return;
  }
  await apiFetch(`/api/modules/economy/shop?guild_id=${state.guildId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, cost, role_id: roleID, enabled: true }),
  });
  await loadEconomy();
}

async function purchaseEconomyItem() {
  if (!state.guildId) return;
  const userID = (qs('#ecoUserId').value || '').trim();
  const itemID = parseInt(qs('#ecoItemId').value || '0', 10);
  if (!userID || itemID <= 0) {
    showToast('User ID and item ID are required.', 'error');
    return;
  }
  await apiFetch(`/api/modules/economy/purchase?guild_id=${state.guildId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id: userID, item_id: itemID }),
  });
  await loadEconomy();
}

async function loadAchievements() {
  if (!state.guildId) return;
  const userID = (qs('#achUserId')?.value || '').trim();
  if (!userID) {
    showToast('Enter a user ID first.', 'error');
    return;
  }
  const table = qs('#achTable');
  if (!table) return;
  const rows = (await apiFetch(`/api/modules/achievements?guild_id=${state.guildId}&user_id=${encodeURIComponent(userID)}`)) || [];
  table.innerHTML = '';
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row';
    div.innerHTML = `<div>${row.badge_key || ''}</div><div>${row.badge_name || ''}</div><div>${formatDate(row.awarded_at)}</div>`;
    table.appendChild(div);
  });
}

async function reviewQueueDecision(actionID, decision) {
  if (!actionID || !decision) return;
  let reason = '';
  if (decision === 'reject') {
    reason = prompt('Optional rejection reason:') || '';
  }
  await apiFetch(`/api/review-queue?guild_id=${state.guildId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action_id: Number(actionID), decision, reason }),
  });
}

async function simulatePolicy() {
  if (!state.guildId) return;
  const actionType = (qs('#policyActionType')?.value || 'quarantine').trim();
  const userIDs = (qs('#policyUserIds')?.value || '')
    .split(',')
    .map((v) => v.trim())
    .filter(Boolean);
  if (!userIDs.length) {
    showToast('Enter at least one user ID for simulation.', 'error');
    return;
  }
  const table = qs('#policySimTable');
  const details = qs('#policySimDetails');
  if (table) table.innerHTML = '';
  if (details) details.textContent = 'Running simulation...';
  const res = await apiFetch(`/api/policy/simulate?guild_id=${state.guildId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action_type: actionType, user_ids: userIDs }),
  });
  const rows = (res && res.results) || [];
  if (table) {
    table.innerHTML = '';
    rows.forEach((row) => {
      const preflight = row.preflight || {};
      const div = document.createElement('div');
      div.className = 'table-row';
      div.innerHTML = `
        <div>${row.user_id || ''}</div>
        <div>${preflight.allowed ? 'yes' : 'no'}</div>
        <div>${row.confirm_required ? 'yes' : 'no'}</div>
        <div>${row.distinct_approver_needed ? 'yes' : 'no'}</div>
        <div>${row.incident_mode_active ? 'on' : 'off'}</div>
      `;
      table.appendChild(div);
    });
  }
  if (details) {
    details.textContent = JSON.stringify(rows, null, 2);
  }
}

async function loadCases() {
  if (!state.guildId) return;
  const userID = (qs('#caseUserId')?.value || '').trim();
  const limit = parseInt(qs('#caseLimit')?.value || '100', 10) || 100;
  const table = qs('#casesTable');
  if (!table) return;
  table.innerHTML = '';
  if (!userID) {
    return;
  }
  const query = new URLSearchParams({ guild_id: state.guildId, user_id: userID, limit: String(limit) }).toString();
  const rows = (await apiFetch(`/api/cases?${query}`)) || [];
  rows.forEach((row) => {
    const div = document.createElement('div');
    div.className = 'table-row case-row';
    div.innerHTML = `
      <div>${formatDate(row.time)}</div>
      <div>${row.type || ''}</div>
      <div>${row.actor || ''}</div>
      <div>${row.summary || ''}</div>
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
  await loadHealthDashboard();
}

async function loadHealthDashboard() {
  if (!state.guildId) return;
  const status = qs('#healthStatus');
  const table = qs('#healthTable');
  if (!status || !table) return;
  status.textContent = 'Loading...';
  const data = await apiFetch(`/api/health/dashboard?guild_id=${state.guildId}`);
  const rows = [
    ['Actions queued', data.actions_queued],
    ['Actions running', data.actions_running],
    ['Actions failed (24h)', data.actions_failed_24h],
    ['Warnings (24h)', data.warnings_24h],
    ['Tickets created (24h)', data.tickets_created_24h],
    ['Backfills active', data.backfills_active],
    ['Incident mode', data.incident_mode_active ? 'on' : 'off'],
    ['Retention', data.retention_enabled ? `${data.retention_days} days` : 'disabled'],
    ['Action dry-run', data.action_dry_run ? 'on' : 'off'],
    ['Two-person approval', data.action_two_person_approval ? 'on' : 'off'],
  ];
  table.innerHTML = '';
  rows.forEach(([metric, value]) => {
    const div = document.createElement('div');
    div.className = 'table-row';
    div.innerHTML = `<div>${metric}</div><div>${value}</div>`;
    table.appendChild(div);
  });
  status.textContent = `Updated at ${new Date().toLocaleTimeString()}`;
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
  const preflight = await runActionPreflight(type, [userId]);
  if (!preflight.allowed) {
    showToast(`Action blocked: ${preflight.summary}`, 'error');
    return;
  }
  if (preflight.summary) {
    const proceed = confirm(`Preflight warning:\n${preflight.summary}\n\nContinue?`);
    if (!proceed) return;
  }
  const safeguards = collectActionSafeguards(type);
  if (safeguards.cancelled) return;
  const reason = prompt(`Reason for ${type} (optional):`);
  try {
    await apiFetch(`/api/actions/${type}?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        user_ids: [userId],
        reason: reason || '',
        target_name: targetName || '',
        confirm_token: safeguards.confirmToken,
        approver_user: safeguards.approverUser,
      }),
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
  const preflight = await runActionPreflight(type, userIds);
  if (!preflight.allowed) {
    showToast(`Bulk action blocked: ${preflight.summary}`, 'error');
    return;
  }
  if (preflight.summary) {
    const proceed = confirm(`Bulk preflight warning:\n${preflight.summary}\n\nContinue?`);
    if (!proceed) return;
  }
  const safeguards = collectActionSafeguards(type);
  if (safeguards.cancelled) return;
  const reason = prompt(`Reason for ${type} (optional):`);
  const payload = {
    user_ids: userIds,
    reason: reason || '',
    target_names: Object.fromEntries(selectedUserMap),
    confirm_token: safeguards.confirmToken,
    approver_user: safeguards.approverUser,
  };
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

function isDestructiveAction(type) {
  return type === 'kick' || type === 'quarantine' || type === 'remove-roles';
}

function collectActionSafeguards(type) {
  const cfg = state.currentSettings || {};
  const destructive = isDestructiveAction(type);
  let confirmToken = '';
  let approverUser = '';
  if (!destructive) {
    return { cancelled: false, confirmToken, approverUser };
  }
  if (cfg.action_require_confirm !== false) {
    const token = prompt('Type CONFIRM to queue this destructive action:');
    if (!token || token.trim().toUpperCase() !== 'CONFIRM') {
      showToast('Action cancelled: confirm token not provided.', 'error');
      return { cancelled: true, confirmToken: '', approverUser: '' };
    }
    confirmToken = 'CONFIRM';
  }
  if (cfg.action_two_person_approval) {
    const approver = prompt('Enter second approver user ID (must be different from actor):');
    if (!approver || !approver.trim()) {
      showToast('Action cancelled: approver required by policy.', 'error');
      return { cancelled: true, confirmToken: '', approverUser: '' };
    }
    approverUser = approver.trim();
  }
  return { cancelled: false, confirmToken, approverUser };
}

async function runActionPreflight(type, userIds) {
  if (!state.guildId || !userIds || !userIds.length) {
    return { allowed: true, summary: '' };
  }
  const actionType = String(type || '').replaceAll('-', '_');
  try {
    const res = await apiFetch(`/api/actions/preflight?guild_id=${state.guildId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action_type: actionType, user_ids: userIds }),
    });
    const results = (res && res.results) || [];
    const blocked = results.some((row) => row && row.allowed === false);
    const messages = [];
    results.forEach((row) => {
      const issues = (row && row.issues) || [];
      issues.forEach((issue) => {
        if (!issue || !issue.message) return;
        messages.push(`User ${row.target_user_id}: ${issue.message}`);
      });
    });
    return {
      allowed: !blocked,
      summary: messages.slice(0, 5).join('\n'),
    };
  } catch (err) {
    return { allowed: false, summary: `preflight failed: ${err.message}` };
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
  const themeSelect = qs('#themeSelect');
  if (themeSelect) {
    themeSelect.onchange = () => applyTheme(themeSelect.value);
  }
  const roleSelect = qs('#dashboardRoleSelect');
  if (roleSelect) {
    roleSelect.value = state.dashboardRole || 'admin';
    roleSelect.onchange = () => {
      state.dashboardRole = roleSelect.value || 'admin';
      localStorage.setItem('modbot_dashboard_role', state.dashboardRole);
      refreshAll().catch((err) => showToast(`Reload failed: ${err.message}`, 'error'));
    };
  }
  qs('#settingsSave').onclick = saveSettings;
  qs('#settingsApplyProfile').onclick = applySettingsProfile;
  qs('#exportDownload').onclick = downloadExport;
  qs('#backupDownload').onclick = downloadBackupSnapshot;
  qs('#backupRestore').onclick = restoreBackupSnapshot;
  qs('#dependencyCheckRun').onclick = () => loadDependencyChecks().catch((err) => showToast(`Dependency check failed: ${err.message}`, 'error'));
  qs('#webhookRefresh').onclick = () => loadWebhooks().catch((err) => showToast(`Webhook load failed: ${err.message}`, 'error'));
  qs('#webhookAdd').onclick = () => addWebhook().catch((err) => showToast(`Webhook add failed: ${err.message}`, 'error'));
  qs('#welcomeSave').onclick = () => { if (requireModulePermissions(FEATURE_WELCOME, 'Save welcome module')) saveWelcome(); };
  qs('#goodbyeSave').onclick = () => { if (requireModulePermissions(FEATURE_GOODBYE, 'Save goodbye module')) saveGoodbye(); };
  qs('#auditSave').onclick = () => { if (requireModulePermissions(FEATURE_AUDIT, 'Save audit module')) saveAudit(); };
  qs('#auditTrailRefresh').onclick = () => loadAuditTrail().catch((err) => showToast(`Audit trail load failed: ${err.message}`, 'error'));
  qs('#inviteSave').onclick = () => { if (requireModulePermissions(FEATURE_INVITE, 'Save invite tracker module')) saveInviteTracker(); };
  qs('#automodSave').onclick = () => { if (requireModulePermissions(FEATURE_AUTOMOD, 'Save automod module')) saveAutoMod(); };
  qs('#reactionRolesSave').onclick = () => { if (requireModulePermissions(FEATURE_REACTION_ROLES, 'Save reaction roles module')) saveReactionRoles(); };
  qs('#warningsSave').onclick = () => { if (requireModulePermissions(FEATURE_WARNINGS, 'Save warnings module')) saveWarningsModule(); };
  qs('#scheduledSave').onclick = () => { if (requireModulePermissions(FEATURE_SCHEDULED, 'Save scheduled module')) saveScheduledModule(); };
  qs('#verificationSave').onclick = () => { if (requireModulePermissions(FEATURE_VERIFICATION, 'Save verification module')) saveVerificationModule(); };
  qs('#ticketsSave').onclick = () => { if (requireModulePermissions(FEATURE_TICKETS, 'Save tickets module')) saveTicketsModule(); };
  qs('#antiRaidSave').onclick = () => { if (requireModulePermissions(FEATURE_ANTI_RAID, 'Save anti-raid module')) saveAntiRaidModule(); };
  qs('#analyticsSave').onclick = () => { if (requireModulePermissions(FEATURE_ANALYTICS, 'Save analytics module')) saveAnalyticsModule(); };
  qs('#analyticsTrendsRefresh').onclick = () => loadAnalyticsTrends().catch((err) => showToast(`Analytics trends failed: ${err.message}`, 'error'));
  qs('#analyticsTrendDays').addEventListener('change', () => loadAnalyticsTrends().catch((err) => showToast(`Analytics trends failed: ${err.message}`, 'error')));
  qs('#appealsSave').onclick = () => { if (requireModulePermissions(FEATURE_APPEALS, 'Save appeals module')) saveAppealsModule(); };
  qs('#starboardSave').onclick = () => { if (requireModulePermissions(FEATURE_STARBOARD, 'Save starboard module')) saveStarboardModule(); };
  qs('#levelingSave').onclick = () => { if (requireModulePermissions(FEATURE_LEVELING, 'Save leveling module')) saveLevelingModule(); };
  qs('#giveawaysSave').onclick = () => { if (requireModulePermissions(FEATURE_GIVEAWAYS, 'Save giveaways module')) saveGiveawaysModule(); };
  qs('#pollsSave').onclick = () => { if (requireModulePermissions(FEATURE_POLLS, 'Save polls module')) savePollsModule(); };
  qs('#suggestionsSave').onclick = () => { if (requireModulePermissions(FEATURE_SUGGESTIONS, 'Save suggestions module')) saveSuggestionsModule(); };
  qs('#keywordAlertsSave').onclick = () => { if (requireModulePermissions(FEATURE_KEYWORD_ALERTS, 'Save keyword alerts module')) saveKeywordAlertsModule(); };
  qs('#afkSave').onclick = () => { if (requireModulePermissions(FEATURE_AFK, 'Save AFK module')) saveAFKModule(); };
  qs('#remindersSave').onclick = () => { if (requireModulePermissions(FEATURE_REMINDERS, 'Save reminders module')) saveRemindersModule(); };
  qs('#accountAgeGuardSave').onclick = () => { if (requireModulePermissions(FEATURE_ACCOUNT_AGE_GUARD, 'Save account age guard module')) saveAccountAgeGuardModule(); };
  qs('#memberNotesSave').onclick = () => { if (requireModulePermissions(FEATURE_MEMBER_NOTES, 'Save member notes module')) saveMemberNotesModule(); };
  qs('#customCommandsSave').onclick = () => { if (requireModulePermissions(FEATURE_CUSTOM_COMMANDS, 'Save custom commands module')) saveCustomCommandsModule(); };
  qs('#rrRefresh').onclick = () => loadReactionRoleRules().catch((err) => showToast(`Rule load failed: ${err.message}`, 'error'));
  qs('#rrAddRule').onclick = () => { if (requireModulePermissions(FEATURE_REACTION_ROLES, 'Add reaction role rule')) addReactionRoleRule(); };
  qs('#warnRefresh').onclick = () => loadWarnings().catch((err) => showToast(`Warnings load failed: ${err.message}`, 'error'));
  qs('#warnIssue').onclick = () => { if (requireModulePermissions(FEATURE_WARNINGS, 'Issue warning')) issueWarning(); };
  qs('#schedRefresh').onclick = () => loadScheduledMessages().catch((err) => showToast(`Schedules load failed: ${err.message}`, 'error'));
  qs('#schedAdd').onclick = () => { if (requireModulePermissions(FEATURE_SCHEDULED, 'Add scheduled message')) addScheduledMessage(); };
  qs('#ticketsRefresh').onclick = () => loadTickets().catch((err) => showToast(`Tickets load failed: ${err.message}`, 'error'));
  qs('#ticketStatusFilter').addEventListener('change', () => loadTickets().catch((err) => showToast(`Tickets load failed: ${err.message}`, 'error')));
  qs('#appealsRefresh').onclick = () => loadAppeals().catch((err) => showToast(`Appeals load failed: ${err.message}`, 'error'));
  qs('#appealStatusFilter').addEventListener('change', () => loadAppeals().catch((err) => showToast(`Appeals load failed: ${err.message}`, 'error')));
  qs('#customCommandsRefresh').onclick = () => loadCustomCommands().catch((err) => showToast(`Commands load failed: ${err.message}`, 'error'));
  qs('#customCommandAdd').onclick = () => { if (requireModulePermissions(FEATURE_CUSTOM_COMMANDS, 'Add custom command')) addCustomCommand(); };
  qs('#levelingRefresh').onclick = () => loadLeaderboard().catch((err) => showToast(`Leaderboard load failed: ${err.message}`, 'error'));
  qs('#giveawaysRefresh').onclick = () => loadGiveaways().catch((err) => showToast(`Giveaways load failed: ${err.message}`, 'error'));
  qs('#giveawayStart').onclick = () => { if (requireModulePermissions(FEATURE_GIVEAWAYS, 'Start giveaway')) startGiveaway(); };
  qs('#pollsRefresh').onclick = () => loadPolls().catch((err) => showToast(`Polls load failed: ${err.message}`, 'error'));
  qs('#pollStart').onclick = () => { if (requireModulePermissions(FEATURE_POLLS, 'Start poll')) startPoll(); };
  qs('#suggestionsRefresh').onclick = () => loadSuggestions().catch((err) => showToast(`Suggestions load failed: ${err.message}`, 'error'));
  qs('#suggestionStatusFilter').addEventListener('change', () => loadSuggestions().catch((err) => showToast(`Suggestions load failed: ${err.message}`, 'error')));
  qs('#remindersRefresh').onclick = () => loadReminders().catch((err) => showToast(`Reminders load failed: ${err.message}`, 'error'));
  qs('#reminderAdd').onclick = () => { if (requireModulePermissions(FEATURE_REMINDERS, 'Add reminder')) addReminder(); };
  qs('#memberNotesRefresh').onclick = () => loadMemberNotes().catch((err) => showToast(`Member notes load failed: ${err.message}`, 'error'));
  qs('#memberNoteAdd').onclick = () => { if (requireModulePermissions(FEATURE_MEMBER_NOTES, 'Add member note')) addMemberNote(); };
  qs('#memberNoteUserFilter').addEventListener('input', () => loadMemberNotes().catch((err) => showToast(`Member notes load failed: ${err.message}`, 'error')));
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
  qs('#settingsAFKEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsRemindersEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsAccountAgeGuardEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsMemberNotesEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsAppealsEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsCustomCommandsEnabled').addEventListener('change', syncModuleBadges);
  qs('#settingsLevelingCurve').addEventListener('change', updateLevelingGuideExamples);
  qs('#settingsLevelingBase').addEventListener('input', updateLevelingGuideExamples);
  qs('#settingsLevelingXP').addEventListener('input', updateLevelingGuideExamples);
  qs('#memberRefresh').onclick = loadMembers;
  qs('#memberStatus').addEventListener('change', reloadMembersForFilters);
  qs('#memberStatus').addEventListener('input', reloadMembersForFilters);
  qs('#memberStatus').addEventListener('click', () => {
    setTimeout(reloadMembersForFilters, 0);
  });
  qs('#actionRefresh').onclick = loadActions;
  qs('#reviewQueueRefresh').onclick = () => loadReviewQueue().catch((err) => showToast(`Review queue load failed: ${err.message}`, 'error'));
  qs('#policySimulate').onclick = () => simulatePolicy().catch((err) => showToast(`Policy simulation failed: ${err.message}`, 'error'));
  qs('#caseRefresh').onclick = () => loadCases().catch((err) => showToast(`Cases load failed: ${err.message}`, 'error'));
  qs('#caseUserId').addEventListener('change', () => loadCases().catch((err) => showToast(`Cases load failed: ${err.message}`, 'error')));
  qs('#eventsRefresh').onclick = () => loadEvents().catch((err) => showToast(`Events load failed: ${err.message}`, 'error'));
  qs('#backfillBtn').onclick = runBackfill;
  qs('#refreshOverview').onclick = loadOverview;
  qs('#healthRefresh').onclick = () => loadHealthDashboard().catch((err) => showToast(`Health load failed: ${err.message}`, 'error'));
  qs('#repRefresh').onclick = () => loadReputationLeaderboard().catch((err) => showToast(`Reputation load failed: ${err.message}`, 'error'));
  qs('#repGivePlus').onclick = () => giveReputation(1).catch((err) => showToast(`Give rep failed: ${err.message}`, 'error'));
  qs('#repGiveMinus').onclick = () => giveReputation(-1).catch((err) => showToast(`Give rep failed: ${err.message}`, 'error'));
  qs('#ecoRefresh').onclick = () => loadEconomy().catch((err) => showToast(`Economy load failed: ${err.message}`, 'error'));
  qs('#ecoAddItem').onclick = () => addEconomyItem().catch((err) => showToast(`Add item failed: ${err.message}`, 'error'));
  qs('#ecoPurchase').onclick = () => purchaseEconomyItem().catch((err) => showToast(`Purchase failed: ${err.message}`, 'error'));
  qs('#achLoad').onclick = () => loadAchievements().catch((err) => showToast(`Achievements load failed: ${err.message}`, 'error'));

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
    if (!requireModulePermissions(FEATURE_REACTION_ROLES, 'Delete reaction role rule')) return;
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

  qs('#reviewQueueTable').addEventListener('click', async (e) => {
    const approve = e.target.closest('button[data-review-approve]');
    const reject = e.target.closest('button[data-review-reject]');
    if (!approve && !reject) return;
    const id = approve ? approve.getAttribute('data-review-approve') : reject.getAttribute('data-review-reject');
    const decision = approve ? 'approve' : 'reject';
    try {
      await reviewQueueDecision(id, decision);
      showToast(`Review ${decision}d.`);
      await loadReviewQueue();
      await loadActions();
    } catch (err) {
      showToast(`Review decision failed: ${err.message}`, 'error');
    }
  });

  qs('#webhookTable').addEventListener('click', async (e) => {
    const btn = e.target.closest('button[data-webhook-del]');
    if (!btn) return;
    try {
      await deleteWebhook(btn.getAttribute('data-webhook-del'));
      showToast('Webhook deleted.');
      await loadWebhooks();
    } catch (err) {
      showToast(`Delete webhook failed: ${err.message}`, 'error');
    }
  });

  qs('#ticketsTable').addEventListener('click', async (e) => {
    const closeBtn = e.target.closest('button[data-ticket-close]');
    if (closeBtn) {
      if (!requireModulePermissions(FEATURE_TICKETS, 'Close ticket')) return;
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
    if (!requireModulePermissions(FEATURE_APPEALS, 'Resolve appeal')) return;
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
    if (!requireModulePermissions(FEATURE_GIVEAWAYS, 'Draw giveaway')) return;
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
    if (!requireModulePermissions(FEATURE_POLLS, 'Close poll')) return;
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
      if (!requireModulePermissions(FEATURE_SUGGESTIONS, 'Approve suggestion')) return;
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
      if (!requireModulePermissions(FEATURE_SUGGESTIONS, 'Reject suggestion')) return;
      try {
        await decideSuggestion(rejectBtn.getAttribute('data-suggestion-reject'), 'reject');
        showToast('Suggestion rejected.');
        await loadSuggestions();
      } catch (err) {
        showToast(`Suggestion action failed: ${err.message}`, 'error');
      }
    }
  });

  qs('#memberNotesTable').addEventListener('click', async (e) => {
    const resolveBtn = e.target.closest('button[data-note-resolve]');
    if (!resolveBtn) return;
    try {
      await resolveMemberNote(resolveBtn.getAttribute('data-note-resolve'));
      showToast('Member note resolved.');
      await loadMemberNotes();
    } catch (err) {
      showToast(`Resolve note failed: ${err.message}`, 'error');
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

  initNavUI();
  injectModuleGuides();
  updateLevelingGuideExamples();
  startMemberFilterWatch();
}

async function refreshAll() {
  await loadOverview();
  await loadMembers();
  await loadActions();
  await loadCases();
  await loadEvents();
  await loadSettings();
  await loadModulePermissions();
  await loadAnalyticsTrends();
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
  await loadReminders();
  await loadMemberNotes();
}

async function bootstrap() {
  await loadGuilds();
  await refreshAll();
  const preferredView = localStorage.getItem(ACTIVE_VIEW_STORAGE_KEY) || 'overview';
  if (qs(`#view-${preferredView}`)) {
    setActiveView(preferredView, false);
  } else {
    setActiveView('overview', false);
  }
}

applyTheme(preferredTheme());
wireEvents();
if (!state.token) {
  showLogin();
} else {
  bootstrap().catch(showLogin);
}
