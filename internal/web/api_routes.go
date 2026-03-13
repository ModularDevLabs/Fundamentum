package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/login", s.handleLogin)

	api := http.NewServeMux()
	api.HandleFunc("/api/health", s.handleHealth)
	api.HandleFunc("/api/health/dashboard", s.handleHealthDashboard)
	api.HandleFunc("/api/guilds", s.handleGuilds)
	api.HandleFunc("/api/settings", s.handleSettings)
	api.HandleFunc("/api/settings/profile/apply", s.handleSettingsProfileApply)
	api.HandleFunc("/api/backfill/start", s.handleBackfillStart)
	api.HandleFunc("/api/backfill/status", s.handleBackfillStatus)
	api.HandleFunc("/api/analytics/trends", s.handleAnalyticsTrends)
	api.HandleFunc("/api/export", s.handleExport)
	api.HandleFunc("/api/backup/export", s.handleBackupExport)
	api.HandleFunc("/api/backup/import", s.handleBackupImport)
	api.HandleFunc("/api/events", s.handleEvents)
	api.HandleFunc("/api/audit-trail", s.handleAuditTrail)
	api.HandleFunc("/api/members", s.handleMembers)
	api.HandleFunc("/api/members/", s.handleMemberDetail)
	api.HandleFunc("/api/cases", s.handleCases)
	api.HandleFunc("/api/actions", s.handleActions)
	api.HandleFunc("/api/actions/preflight", s.handleActionPreflight)
	api.HandleFunc("/api/actions/", s.handleActionDetail)
	api.HandleFunc("/api/review-queue", s.handleReviewQueue)
	api.HandleFunc("/api/policy/simulate", s.handlePolicySimulate)
	api.HandleFunc("/api/dependencies/check", s.handleDependencyCheck)
	api.HandleFunc("/api/modules/invite/status", s.handleInviteModuleStatus)
	api.HandleFunc("/api/modules/permissions", s.handleModulePermissions)
	api.HandleFunc("/api/modules/reaction-roles/rules", s.handleReactionRoleRules)
	api.HandleFunc("/api/modules/reaction-roles/rules/", s.handleReactionRoleRuleDetail)
	api.HandleFunc("/api/modules/warnings", s.handleWarnings)
	api.HandleFunc("/api/modules/warnings/issue", s.handleWarningIssue)
	api.HandleFunc("/api/modules/scheduled/messages", s.handleScheduledMessages)
	api.HandleFunc("/api/modules/scheduled/messages/", s.handleScheduledMessageDetail)
	api.HandleFunc("/api/modules/tickets", s.handleTickets)
	api.HandleFunc("/api/modules/tickets/", s.handleTicketDetail)
	api.HandleFunc("/api/modules/appeals", s.handleAppeals)
	api.HandleFunc("/api/modules/appeals/", s.handleAppealDetail)
	api.HandleFunc("/api/modules/custom-commands/commands", s.handleCustomCommands)
	api.HandleFunc("/api/modules/custom-commands/commands/", s.handleCustomCommandDetail)
	api.HandleFunc("/api/modules/leveling/leaderboard", s.handleLevelingLeaderboard)
	api.HandleFunc("/api/modules/giveaways", s.handleGiveaways)
	api.HandleFunc("/api/modules/giveaways/start", s.handleGiveawayStart)
	api.HandleFunc("/api/modules/giveaways/", s.handleGiveawayDetail)
	api.HandleFunc("/api/modules/polls", s.handlePolls)
	api.HandleFunc("/api/modules/polls/start", s.handlePollStart)
	api.HandleFunc("/api/modules/polls/", s.handlePollDetail)
	api.HandleFunc("/api/modules/suggestions", s.handleSuggestions)
	api.HandleFunc("/api/modules/suggestions/", s.handleSuggestionDetail)
	api.HandleFunc("/api/modules/reminders", s.handleReminders)
	api.HandleFunc("/api/modules/member-notes", s.handleMemberNotes)
	api.HandleFunc("/api/modules/member-notes/", s.handleMemberNoteDetail)
	api.HandleFunc("/api/modules/reputation/give", s.handleReputationGive)
	api.HandleFunc("/api/modules/reputation/leaderboard", s.handleReputationLeaderboard)
	api.HandleFunc("/api/integrations/webhooks", s.handleWebhooks)
	api.HandleFunc("/api/integrations/webhooks/", s.handleWebhookDetail)

	mux.Handle("/api/", s.authMiddleware(api))

	mux.Handle("/", s.handleStatic())
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var payload struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.Password != s.adminPass {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "modbot_auth",
		Value:    s.adminPass,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	if n, err := strconv.Atoi(value); err == nil {
		return n
	}
	return fallback
}

func parseIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}
