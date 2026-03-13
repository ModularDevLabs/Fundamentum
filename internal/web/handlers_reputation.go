package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleReputationGive(w http.ResponseWriter, r *http.Request) {
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
		FromUserID string `json:"from_user_id"`
		ToUserID   string `json:"to_user_id"`
		Delta      int    `json:"delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload.FromUserID = strings.TrimSpace(payload.FromUserID)
	payload.ToUserID = strings.TrimSpace(payload.ToUserID)
	if payload.FromUserID == "" || payload.ToUserID == "" || payload.FromUserID == payload.ToUserID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.Delta == 0 {
		payload.Delta = 1
	}
	if payload.Delta > 0 {
		payload.Delta = 1
	} else {
		payload.Delta = -1
	}
	last, found, err := s.repos.Reputation.LastGivenAt(r.Context(), guildID, payload.FromUserID, payload.ToUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if found && time.Since(last) < 12*time.Hour {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("reputation cooldown active for this pair"))
		return
	}
	if err := s.repos.Reputation.AddDelta(r.Context(), guildID, payload.FromUserID, payload.ToUserID, payload.Delta); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleReputationLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	rows, err := s.repos.Reputation.Leaderboard(r.Context(), guildID, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}
