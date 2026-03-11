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
	api.HandleFunc("/api/guilds", s.handleGuilds)
	api.HandleFunc("/api/settings", s.handleSettings)
	api.HandleFunc("/api/backfill/start", s.handleBackfillStart)
	api.HandleFunc("/api/backfill/status", s.handleBackfillStatus)
	api.HandleFunc("/api/events", s.handleEvents)
	api.HandleFunc("/api/members", s.handleMembers)
	api.HandleFunc("/api/members/", s.handleMemberDetail)
	api.HandleFunc("/api/actions", s.handleActions)
	api.HandleFunc("/api/actions/", s.handleActionDetail)
	api.HandleFunc("/api/modules/invite/status", s.handleInviteModuleStatus)
	api.HandleFunc("/api/modules/reaction-roles/rules", s.handleReactionRoleRules)
	api.HandleFunc("/api/modules/reaction-roles/rules/", s.handleReactionRoleRuleDetail)

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
