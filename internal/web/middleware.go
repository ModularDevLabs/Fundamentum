package web

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type authContextKey string

const (
	dashboardRoleContextKey     authContextKey = "dashboard_role"
	dashboardUserContextKey     authContextKey = "dashboard_user"
	dashboardCSRFContextKey     authContextKey = "dashboard_csrf"
	dashboardAuthModeContextKey authContextKey = "dashboard_auth_mode"
)

type authIdentity struct {
	Username  string
	Role      string
	CSRFToken string
	Mode      string // session | proxy | legacy
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := s.authenticatedIdentity(r)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("unauthorized"))
			return
		}
		ctx := context.WithValue(r.Context(), dashboardRoleContextKey, identity.Role)
		ctx = context.WithValue(ctx, dashboardUserContextKey, identity.Username)
		ctx = context.WithValue(ctx, dashboardCSRFContextKey, identity.CSRFToken)
		ctx = context.WithValue(ctx, dashboardAuthModeContextKey, identity.Mode)
		r = r.WithContext(ctx)

		if !s.authorizeRoleForRequest(w, r) {
			return
		}
		if !s.validateCSRFForRequest(w, r, identity) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) validateCSRFForRequest(w http.ResponseWriter, r *http.Request, identity authIdentity) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return true
	}
	if identity.Mode == "proxy" {
		if !sameOrigin(r) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("forbidden: cross-origin request blocked"))
			return false
		}
		return true
	}
	csrf := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
	if csrf == "" || csrf != identity.CSRFToken {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden: missing or invalid csrf token"))
		return false
	}
	return true
}

func sameOrigin(r *http.Request) bool {
	reqHost := strings.ToLower(strings.TrimSpace(r.Host))
	if reqHost == "" {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin != "" {
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, reqHost)
	}
	referer := strings.TrimSpace(r.Header.Get("Referer"))
	if referer != "" {
		u, err := url.Parse(referer)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, reqHost)
	}
	return false
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
	role := strings.TrimSpace(strings.ToLower(dashboardRoleFromContext(r.Context())))
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

func dashboardRoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(dashboardRoleContextKey).(string)
	return strings.TrimSpace(strings.ToLower(role))
}

func dashboardUsernameFromContext(ctx context.Context) string {
	name, _ := ctx.Value(dashboardUserContextKey).(string)
	return strings.TrimSpace(strings.ToLower(name))
}

func dashboardCSRFTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(dashboardCSRFContextKey).(string)
	return strings.TrimSpace(token)
}

func dashboardAuthModeFromContext(ctx context.Context) string {
	mode, _ := ctx.Value(dashboardAuthModeContextKey).(string)
	return strings.TrimSpace(strings.ToLower(mode))
}

func rbacPolicyKeyForPath(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/actions"):
		return "actions"
	case strings.HasPrefix(path, "/api/raid/panic"):
		return models.FeatureRaidPanic
	case strings.HasPrefix(path, "/api/settings"):
		return "settings"
	case strings.HasPrefix(path, "/api/modules/warnings"):
		return models.FeatureWarnings
	case strings.HasPrefix(path, "/api/modules/join-screening"):
		return models.FeatureJoinScreening
	case strings.HasPrefix(path, "/api/modules/reaction-roles"):
		return models.FeatureReactionRoles
	case strings.HasPrefix(path, "/api/modules/role-progression"):
		return models.FeatureRoleProgression
	case strings.HasPrefix(path, "/api/modules/scheduled"):
		return models.FeatureScheduled
	case strings.HasPrefix(path, "/api/modules/tickets"):
		return models.FeatureTickets
	case strings.HasPrefix(path, "/api/modules/appeals"):
		return models.FeatureAppeals
	case strings.HasPrefix(path, "/api/modules/custom-commands"):
		return models.FeatureCustomCommands
	case strings.HasPrefix(path, "/api/modules/birthdays"):
		return models.FeatureBirthdays
	case strings.HasPrefix(path, "/api/modules/streaks"):
		return models.FeatureStreaks
	case strings.HasPrefix(path, "/api/modules/season-resets"):
		return models.FeatureSeasonResets
	case strings.HasPrefix(path, "/api/modules/reputation"):
		return models.FeatureReputation
	case strings.HasPrefix(path, "/api/modules/economy"):
		return models.FeatureEconomy
	case strings.HasPrefix(path, "/api/modules/achievements"):
		return models.FeatureAchievements
	case strings.HasPrefix(path, "/api/modules/trivia"):
		return models.FeatureTrivia
	case strings.HasPrefix(path, "/api/modules/calendar"):
		return models.FeatureCalendar
	case strings.HasPrefix(path, "/api/modules/confessions"):
		return models.FeatureConfessions
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

func (s *Server) authenticatedIdentity(r *http.Request) (authIdentity, bool) {
	now := time.Now().UTC()
	if cookie, err := r.Cookie("modbot_session"); err == nil && strings.TrimSpace(cookie.Value) != "" {
		sess, err := s.repos.DashboardAuth.GetSession(r.Context(), strings.TrimSpace(cookie.Value))
		if err == nil && !sess.Revoked && sess.ExpiresAt.After(now) {
			if now.Sub(sess.LastSeenAt) > 2*time.Minute {
				_ = s.repos.DashboardAuth.TouchSession(r.Context(), sess.SessionID, now)
			}
			return authIdentity{Username: sess.Username, Role: sess.Role, CSRFToken: sess.CSRFToken, Mode: "session"}, true
		}
		if err != nil && err != sql.ErrNoRows {
			s.logger.Error("session lookup failed: %v", err)
		}
	}

	if s.authProxyEnabled && strings.TrimSpace(s.authProxySecret) != "" {
		proxySecret := strings.TrimSpace(r.Header.Get("X-Modbot-Proxy-Secret"))
		if proxySecret == strings.TrimSpace(s.authProxySecret) {
			username := strings.TrimSpace(strings.ToLower(r.Header.Get(s.authProxyUserHeader)))
			if username != "" {
				role := strings.TrimSpace(strings.ToLower(r.Header.Get(s.authProxyRoleHeader)))
				if role == "" {
					role = "support"
				}
				return authIdentity{Username: username, Role: role, Mode: "proxy"}, true
			}
		}
	}

	if s.allowLegacyBearer {
		if secret := legacySecretFromRequest(r); secret != "" {
			if role, ok := s.roleFromLegacySecret(secret); ok {
				return authIdentity{Username: role, Role: role, Mode: "legacy"}, true
			}
		}
	}

	return authIdentity{}, false
}

func legacySecretFromRequest(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	if cookie, err := r.Cookie("modbot_auth"); err == nil {
		return strings.TrimSpace(cookie.Value)
	}
	return ""
}

func (s *Server) roleFromLegacySecret(secret string) (string, bool) {
	if secret == s.adminPass {
		return "admin", true
	}
	for role, candidate := range s.dashboardRoleSecrets {
		if strings.TrimSpace(secret) == strings.TrimSpace(candidate) {
			normalized := strings.TrimSpace(strings.ToLower(role))
			if normalized == "" {
				continue
			}
			return normalized, true
		}
	}
	return "", false
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Debug("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}
