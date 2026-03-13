package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleScheduledMessages(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := s.repos.Scheduled.ListByGuild(r.Context(), guildID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	case http.MethodPost:
		var payload struct {
			ChannelID       string `json:"channel_id"`
			Content         string `json:"content"`
			IntervalMinutes int    `json:"interval_minutes"`
			Enabled         bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if payload.IntervalMinutes < 1 || strings.TrimSpace(payload.ChannelID) == "" || strings.TrimSpace(payload.Content) == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := s.repos.Scheduled.Create(r.Context(), models.ScheduledMessageRow{
			GuildID:         guildID,
			ChannelID:       strings.TrimSpace(payload.ChannelID),
			Content:         strings.TrimSpace(payload.Content),
			IntervalMinutes: payload.IntervalMinutes,
			NextRunAt:       time.Now().UTC().Add(time.Duration(payload.IntervalMinutes) * time.Minute),
			Enabled:         payload.Enabled,
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

func (s *Server) handleScheduledMessageDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/modules/scheduled/messages/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.repos.Scheduled.Delete(r.Context(), guildID, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
