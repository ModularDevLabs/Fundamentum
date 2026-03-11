package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleAppeals(w http.ResponseWriter, r *http.Request) {
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
	rows, err := s.repos.Appeals.ListByGuild(r.Context(), guildID, status, 200)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (s *Server) handleAppealDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/modules/appeals/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "resolve" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var payload struct {
		Resolution string `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	actor := r.Header.Get("X-Actor-User")
	if actor == "" {
		actor = "dashboard"
	}
	if err := s.repos.Appeals.Resolve(r.Context(), guildID, id, actor, strings.TrimSpace(payload.Resolution)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
