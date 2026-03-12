package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	status := r.URL.Query().Get("status")
	rows, err := s.repos.Suggestions.ListByGuild(r.Context(), guildID, status, 100)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (s *Server) handleSuggestionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/modules/suggestions/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || (parts[1] != "approve" && parts[1] != "reject") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var payload struct {
		Note string `json:"note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)
	status := "approved"
	if parts[1] == "reject" {
		status = "rejected"
	}
	actor := r.Header.Get("X-Actor-User")
	if actor == "" {
		actor = "dashboard"
	}
	if err := s.repos.Suggestions.Decide(r.Context(), guildID, id, status, actor, strings.TrimSpace(payload.Note)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if cfg, err := s.repos.Settings.Get(r.Context(), guildID); err == nil && strings.TrimSpace(cfg.SuggestionsLogChannelID) != "" {
		msg := fmt.Sprintf("Suggestion #%d %s by %s", id, status, actor)
		if strings.TrimSpace(payload.Note) != "" {
			msg += ": " + strings.TrimSpace(payload.Note)
		}
		_, _ = s.discord.SendChannelMessage(cfg.SuggestionsLogChannelID, msg)
	}
	w.WriteHeader(http.StatusNoContent)
}
