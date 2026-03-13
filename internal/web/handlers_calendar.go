package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/db"
	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleCalendarEvents(w http.ResponseWriter, r *http.Request) {
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !s.ensureFeatureEnabled(w, r, guildID, models.FeatureCalendar, "calendar") {
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := s.repos.Calendar.ListEvents(r.Context(), guildID, 100)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	case http.MethodPost:
		var payload struct {
			Title     string `json:"title"`
			Details   string `json:"details"`
			StartAt   string `json:"start_at"`
			CreatedBy string `json:"created_by"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		payload.Title = strings.TrimSpace(payload.Title)
		payload.StartAt = strings.TrimSpace(payload.StartAt)
		payload.CreatedBy = strings.TrimSpace(payload.CreatedBy)
		if payload.Title == "" || payload.StartAt == "" || payload.CreatedBy == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := s.repos.Calendar.CreateEvent(r.Context(), db.CalendarEventRow{
			GuildID:   guildID,
			Title:     payload.Title,
			Details:   payload.Details,
			StartAt:   payload.StartAt,
			CreatedBy: payload.CreatedBy,
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

func (s *Server) handleCalendarRSVP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !s.ensureFeatureEnabled(w, r, guildID, models.FeatureCalendar, "calendar") {
		return
	}
	var payload struct {
		EventID int64  `json:"event_id"`
		UserID  string `json:"user_id"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload.Status = strings.ToLower(strings.TrimSpace(payload.Status))
	if payload.EventID <= 0 || strings.TrimSpace(payload.UserID) == "" || (payload.Status != "yes" && payload.Status != "no" && payload.Status != "maybe") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.repos.Calendar.SetRSVP(r.Context(), payload.EventID, strings.TrimSpace(payload.UserID), payload.Status); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCalendarRSVPs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !s.ensureFeatureEnabled(w, r, guildID, models.FeatureCalendar, "calendar") {
		return
	}
	eventID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("event_id")), 10, 64)
	if err != nil || eventID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rows, err := s.repos.Calendar.ListRSVPs(r.Context(), eventID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}
