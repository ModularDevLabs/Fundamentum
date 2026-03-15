package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type dashboardUserDTO struct {
	Username    string `json:"username"`
	Role        string `json:"role"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	LastLoginAt string `json:"last_login_at"`
}

func (s *Server) handleDashboardUsers(w http.ResponseWriter, r *http.Request) {
	if dashboardRoleFromContext(r.Context()) != "admin" {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("admin role required"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := s.repos.DashboardAuth.ListUsers(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		out := make([]dashboardUserDTO, 0, len(rows))
		for _, row := range rows {
			item := dashboardUserDTO{
				Username:  row.Username,
				Role:      row.Role,
				Enabled:   row.Enabled,
				CreatedAt: row.CreatedAt.Format("2006-01-02 15:04:05Z07:00"),
				UpdatedAt: row.UpdatedAt.Format("2006-01-02 15:04:05Z07:00"),
			}
			if row.LastLoginAt != nil {
				item.LastLoginAt = row.LastLoginAt.Format("2006-01-02 15:04:05Z07:00")
			}
			out = append(out, item)
		}
		writeJSON(w, out)
	case http.MethodPost:
		var payload struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
			Enabled  *bool  `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		username := strings.ToLower(strings.TrimSpace(payload.Username))
		password := strings.TrimSpace(payload.Password)
		role := strings.ToLower(strings.TrimSpace(payload.Role))
		if username == "" || password == "" || role == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("username, password, and role are required"))
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		enabled := true
		if payload.Enabled != nil {
			enabled = *payload.Enabled
		}
		if err := s.repos.DashboardAuth.UpsertUser(r.Context(), db.DashboardUserRow{
			Username:     username,
			PasswordHash: string(hash),
			Role:         role,
			Enabled:      enabled,
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDashboardUserDetail(w http.ResponseWriter, r *http.Request) {
	if dashboardRoleFromContext(r.Context()) != "admin" {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("admin role required"))
		return
	}
	username := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/dashboard/users/"))
	username = strings.ToLower(username)
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPut:
		current, err := s.repos.DashboardAuth.GetUser(r.Context(), username)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var payload struct {
			Password *string `json:"password"`
			Role     *string `json:"role"`
			Enabled  *bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if payload.Password != nil {
			pwd := strings.TrimSpace(*payload.Password)
			if pwd == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("password cannot be empty"))
				return
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			current.PasswordHash = string(hash)
		}
		if payload.Role != nil {
			role := strings.ToLower(strings.TrimSpace(*payload.Role))
			if role == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("role cannot be empty"))
				return
			}
			current.Role = role
		}
		if payload.Enabled != nil {
			current.Enabled = *payload.Enabled
		}
		if current.Role == "admin" && !current.Enabled {
			count, err := s.repos.DashboardAuth.CountEnabledAdmins(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if count <= 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("cannot disable last enabled admin"))
				return
			}
		}
		if err := s.repos.DashboardAuth.UpsertUser(r.Context(), current); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !current.Enabled {
			_ = s.repos.DashboardAuth.RevokeSessionsForUser(r.Context(), current.Username)
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		current, err := s.repos.DashboardAuth.GetUser(r.Context(), username)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if current.Role == "admin" {
			count, err := s.repos.DashboardAuth.CountEnabledAdmins(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if count <= 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("cannot delete last enabled admin"))
				return
			}
		}
		if err := s.repos.DashboardAuth.DeleteUser(r.Context(), username); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = s.repos.DashboardAuth.RevokeSessionsForUser(r.Context(), username)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
