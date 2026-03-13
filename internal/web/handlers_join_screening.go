package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleJoinScreening(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "pending"
	}
	rows, err := s.repos.JoinScreening.ListByStatus(r.Context(), guildID, status, 200)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (s *Server) handleJoinScreeningReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var payload struct {
		ID         int64  `json:"id"`
		Decision   string `json:"decision"`
		ReviewedBy string `json:"reviewed_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload.Decision = strings.TrimSpace(strings.ToLower(payload.Decision))
	payload.ReviewedBy = strings.TrimSpace(payload.ReviewedBy)
	if payload.ID <= 0 || (payload.Decision != "approved" && payload.Decision != "rejected") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	row, found, err := s.repos.JoinScreening.GetByID(r.Context(), guildID, payload.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if row.Status != "pending" {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("already reviewed"))
		return
	}
	if err := s.repos.JoinScreening.Review(r.Context(), guildID, payload.ID, payload.Decision, payload.ReviewedBy); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	queuedKick := false
	if payload.Decision == "rejected" {
		actionPayload, _ := json.Marshal(map[string]any{
			"reason": "Join screening rejected by moderation review",
		})
		if _, err := s.repos.Actions.Enqueue(r.Context(), models.ActionRow{
			GuildID:      guildID,
			ActorUserID:  payload.ReviewedBy,
			TargetUserID: row.UserID,
			Type:         "kick",
			PayloadJSON:  string(actionPayload),
			Status:       "queued",
		}); err == nil {
			queuedKick = true
			s.discord.NotifyActionQueued()
		}
	}
	writeJSON(w, map[string]any{
		"ok":          true,
		"id":          strconv.FormatInt(payload.ID, 10),
		"queued_kick": queuedKick,
	})
}
