package web

import (
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleServerPulse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	settings, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	now := time.Now().UTC()
	since24 := now.Add(-24 * time.Hour)
	inactiveCutoff := now.AddDate(0, 0, -settings.InactiveDays)

	tracked, err := s.repos.Activity.CountTracked(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	activeMap, err := s.repos.Activity.ActiveUsersSince(r.Context(), guildID, since24)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	inactive, err := s.repos.Activity.CountInactiveBefore(r.Context(), guildID, inactiveCutoff)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	warnings24, err := s.repos.Warnings.CountSince(r.Context(), guildID, since24)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	actions24, err := s.repos.Actions.CountSince(r.Context(), guildID, since24, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	queued, err := s.repos.Actions.CountByStatus(r.Context(), guildID, "queued")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	openTickets, err := s.repos.Tickets.CountByStatus(r.Context(), guildID, "open")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	topRepUser, topRepScore := "", 0
	if rep, err := s.repos.Reputation.Leaderboard(r.Context(), guildID, 1); err == nil && len(rep) > 0 {
		topRepUser = rep[0].UserID
		topRepScore = rep[0].Score
	}
	topTriviaUser, topTriviaScore := "", 0
	if trivia, err := s.repos.Trivia.Leaderboard(r.Context(), guildID, 1); err == nil && len(trivia) > 0 {
		topTriviaUser = trivia[0].UserID
		topTriviaScore = trivia[0].Score
	}

	writeJSON(w, map[string]any{
		"guild_id":           guildID,
		"generated_at":       now.Format(time.RFC3339),
		"tracked_members":    tracked,
		"active_members_24h": len(activeMap),
		"inactive_members":   inactive,
		"warnings_24h":       warnings24,
		"actions_24h":        actions24,
		"actions_queued":     queued,
		"open_tickets":       openTickets,
		"top_reputation": map[string]any{
			"user_id": topRepUser,
			"score":   topRepScore,
		},
		"top_trivia": map[string]any{
			"user_id": topTriviaUser,
			"score":   topTriviaScore,
		},
	})
}
