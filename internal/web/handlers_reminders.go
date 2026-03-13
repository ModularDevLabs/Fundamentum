package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleReminders(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := s.repos.Reminders.ListByGuild(r.Context(), guildID, 100)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	case http.MethodPost:
		var payload struct {
			ChannelID string `json:"channel_id"`
			Content   string `json:"content"`
			RunAt     string `json:"run_at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		runAt, err := time.Parse(time.RFC3339, strings.TrimSpace(payload.RunAt))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := s.repos.Reminders.Create(r.Context(), models.ReminderRow{
			GuildID:   guildID,
			ChannelID: strings.TrimSpace(payload.ChannelID),
			Content:   strings.TrimSpace(payload.Content),
			RunAt:     runAt,
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
