package web

import (
	"net/http"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleAchievements(w http.ResponseWriter, r *http.Request) {
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
	if !s.ensureFeatureEnabled(w, r, guildID, models.FeatureAchievements, "achievements") {
		return
	}

	levelRow, _, _ := s.repos.Leveling.GetMember(r.Context(), guildID, userID)
	repTotal, _ := s.repos.Reputation.TotalForUser(r.Context(), guildID, userID)
	balance, _ := s.repos.Economy.GetBalance(r.Context(), guildID, userID)

	if levelRow.Level >= 5 {
		_ = s.repos.Achievements.AwardIfMissing(r.Context(), guildID, userID, "level_5", "Rising Star", map[string]any{"level": levelRow.Level})
	}
	if levelRow.Level >= 10 {
		_ = s.repos.Achievements.AwardIfMissing(r.Context(), guildID, userID, "level_10", "Veteran", map[string]any{"level": levelRow.Level})
	}
	if repTotal >= 25 {
		_ = s.repos.Achievements.AwardIfMissing(r.Context(), guildID, userID, "rep_25", "Community Pillar", map[string]any{"reputation": repTotal})
	}
	if balance >= 500 {
		_ = s.repos.Achievements.AwardIfMissing(r.Context(), guildID, userID, "coins_500", "Coin Collector", map[string]any{"balance": balance})
	}

	rows, err := s.repos.Achievements.ListByUser(r.Context(), guildID, userID, 100)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}
