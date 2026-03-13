package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleConfessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cfg, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !cfg.FeatureEnabled(models.FeatureConfessions) || !cfg.ConfessionsEnabled {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("confessions module is disabled"))
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "pending"
	}
	rows, err := s.repos.Confessions.ListByStatus(r.Context(), guildID, status, 200)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (s *Server) handleConfessionReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cfg, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !cfg.FeatureEnabled(models.FeatureConfessions) || !cfg.ConfessionsEnabled {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("confessions module is disabled"))
		return
	}
	var payload struct {
		ID       int64  `json:"id"`
		Decision string `json:"decision"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	row, found, err := s.repos.Confessions.Get(r.Context(), payload.ID)
	if err != nil || !found || row.GuildID != guildID {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch strings.ToLower(strings.TrimSpace(payload.Decision)) {
	case "approve":
		msg, err := s.discord.SendChannelMessage(strings.TrimSpace((func() string {
			cfg, err := s.repos.Settings.Get(r.Context(), guildID)
			if err != nil {
				return ""
			}
			return cfg.ConfessionsChannelID
		})()), fmt.Sprintf("Anonymous confession #%d:\n%s", row.ID, row.Content))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = s.repos.Confessions.UpdateStatus(r.Context(), row.ID, "posted", msg)
	case "reject":
		_ = s.repos.Confessions.UpdateStatus(r.Context(), row.ID, "rejected", "")
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
