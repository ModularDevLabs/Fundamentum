package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleRoleProgressionRules(w http.ResponseWriter, r *http.Request) {
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := s.repos.RoleProgression.ListByGuild(r.Context(), guildID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	case http.MethodPost:
		var payload struct {
			Metric    string `json:"metric"`
			Threshold int    `json:"threshold"`
			RoleID    string `json:"role_id"`
			Enabled   bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		payload.Metric = strings.TrimSpace(payload.Metric)
		payload.RoleID = strings.TrimSpace(payload.RoleID)
		if payload.Metric != "level" && payload.Metric != "reputation" && payload.Metric != "economy" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if payload.Threshold < 0 || payload.RoleID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := s.repos.RoleProgression.Create(r.Context(), models.RoleProgressionRuleRow{
			GuildID:   guildID,
			Metric:    payload.Metric,
			Threshold: payload.Threshold,
			RoleID:    payload.RoleID,
			Enabled:   payload.Enabled,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"id": id})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRoleProgressionRuleDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/modules/role-progression/rules/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.repos.RoleProgression.Delete(r.Context(), guildID, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRoleProgressionSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload.UserID = strings.TrimSpace(payload.UserID)
	if payload.UserID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := s.discord.SyncRoleProgressionForUser(r.Context(), guildID, payload.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, res)
}
