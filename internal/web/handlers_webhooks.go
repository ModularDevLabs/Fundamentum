package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

func (s *Server) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := s.repos.Webhooks.ListByGuild(r.Context(), guildID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	case http.MethodPost:
		var payload struct {
			URL     string   `json:"url"`
			Events  []string `json:"events"`
			Enabled bool     `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		payload.URL = strings.TrimSpace(payload.URL)
		if payload.URL == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(payload.Events) == 0 {
			payload.Events = []string{"action_success", "action_failed"}
		}
		id, err := s.repos.Webhooks.Create(r.Context(), models.WebhookIntegrationRow{
			GuildID: guildID,
			URL:     payload.URL,
			Events:  payload.Events,
			Enabled: payload.Enabled,
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

func (s *Server) handleWebhookDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	idText := strings.TrimPrefix(r.URL.Path, "/api/integrations/webhooks/")
	idText = strings.TrimSpace(idText)
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.repos.Webhooks.Delete(r.Context(), guildID, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
