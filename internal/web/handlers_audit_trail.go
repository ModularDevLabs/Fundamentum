package web

import (
	"net/http"
	"strings"
)

func (s *Server) handleAuditTrail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	limit := parseInt(r.URL.Query().Get("limit"), 100)
	if limit > 500 {
		limit = 500
	}
	rows, err := s.repos.AuditTrail.ListByGuild(r.Context(), guildID, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}
