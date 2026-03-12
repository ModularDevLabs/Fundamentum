package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.isAuthorized(r) {
			if !s.authorizeRoleForRequest(w, r) {
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	})
}

func (s *Server) authorizeRoleForRequest(w http.ResponseWriter, r *http.Request) bool {
	policyKey := rbacPolicyKeyForPath(r.URL.Path)
	if policyKey == "" {
		return true
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		return true
	}
	cfg, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}
	role := strings.TrimSpace(strings.ToLower(r.Header.Get("X-Dashboard-Role")))
	if role == "" {
		role = "admin"
	}
	if role == "admin" {
		return true
	}
	allowed := cfg.DashboardRolePolicies[policyKey]
	if len(allowed) == 0 {
		return true
	}
	for _, candidate := range allowed {
		if strings.TrimSpace(strings.ToLower(candidate)) == role {
			return true
		}
	}
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte("forbidden by role policy"))
	return false
}

func rbacPolicyKeyForPath(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/actions"):
		return "actions"
	case strings.HasPrefix(path, "/api/settings"):
		return "settings"
	case strings.HasPrefix(path, "/api/modules/warnings"):
		return models.FeatureWarnings
	case strings.HasPrefix(path, "/api/modules/reaction-roles"):
		return models.FeatureReactionRoles
	case strings.HasPrefix(path, "/api/modules/scheduled"):
		return models.FeatureScheduled
	case strings.HasPrefix(path, "/api/modules/tickets"):
		return models.FeatureTickets
	case strings.HasPrefix(path, "/api/modules/appeals"):
		return models.FeatureAppeals
	case strings.HasPrefix(path, "/api/modules/custom-commands"):
		return models.FeatureCustomCommands
	case strings.HasPrefix(path, "/api/modules/giveaways"):
		return models.FeatureGiveaways
	case strings.HasPrefix(path, "/api/modules/polls"):
		return models.FeaturePolls
	case strings.HasPrefix(path, "/api/modules/suggestions"):
		return models.FeatureSuggestions
	case strings.HasPrefix(path, "/api/modules/reminders"):
		return models.FeatureReminders
	case strings.HasPrefix(path, "/api/modules/member-notes"):
		return models.FeatureMemberNotes
	case strings.HasPrefix(path, "/api/modules/invite"):
		return models.FeatureInviteTracker
	default:
		return ""
	}
}

func (s *Server) isAuthorized(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		if strings.TrimPrefix(auth, "Bearer ") == s.adminPass {
			return true
		}
	}
	cookie, err := r.Cookie("modbot_auth")
	if err == nil && cookie.Value == s.adminPass {
		return true
	}
	return false
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Debug("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}
