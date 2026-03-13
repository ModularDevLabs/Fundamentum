package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

func (s *Server) handleSeasonResetStatus(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, map[string]any{
		"enabled":           cfg.SeasonResetsEnabled,
		"feature_enabled":   cfg.FeatureEnabled(models.FeatureSeasonResets),
		"cadence":           cfg.SeasonResetCadence,
		"next_run_at":       cfg.SeasonResetNextRunAt,
		"modules":           cfg.SeasonResetModules,
		"supported_modules": []string{"leveling", "economy", "trivia"},
	})
}

func (s *Server) handleSeasonResetRun(w http.ResponseWriter, r *http.Request) {
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
		Actor string `json:"actor"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)
	row, err := s.discord.RunSeasonResetNow(r.Context(), guildID, strings.TrimSpace(payload.Actor))
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	writeJSON(w, row)
}

func (s *Server) handleSeasonResetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	rows, err := s.discord.SeasonResetHistory(r.Context(), guildID, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}
