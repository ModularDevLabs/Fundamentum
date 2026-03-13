package web

import (
	"net/http"
	"strings"
)

func (s *Server) handleStreaksLeaderboard(w http.ResponseWriter, r *http.Request) {
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
	rows, err := s.repos.Streaks.Leaderboard(r.Context(), guildID, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (s *Server) handleStreaksUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if guildID == "" || userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	row, found, err := s.repos.Streaks.GetUser(r.Context(), guildID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !found {
		writeJSON(w, map[string]any{"user_id": userID, "current_streak": 0, "best_streak": 0})
		return
	}
	writeJSON(w, row)
}
