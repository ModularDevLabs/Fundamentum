package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleWarnings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rows, err := s.repos.Warnings.ListByGuild(r.Context(), guildID, 200)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (s *Server) handleWarningIssue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var payload struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload.UserID = strings.TrimSpace(payload.UserID)
	payload.Reason = strings.TrimSpace(payload.Reason)
	if payload.UserID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	settings, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !settings.FeatureEnabled(models.FeatureWarnings) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("warnings module is disabled"))
		return
	}

	actor := r.Header.Get("X-Actor-User")
	if actor == "" {
		actor = "dashboard"
	}
	id, err := s.repos.Warnings.Create(r.Context(), models.WarningRow{
		GuildID:     guildID,
		UserID:      payload.UserID,
		ActorUserID: actor,
		Reason:      payload.Reason,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	count, err := s.repos.Warnings.CountByUser(r.Context(), guildID, payload.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	action := ""
	if settings.WarnKickThreshold > 0 && count >= settings.WarnKickThreshold {
		action = "kick"
	} else if settings.WarnQuarantineThreshold > 0 && count >= settings.WarnQuarantineThreshold {
		action = "quarantine"
	}
	if action != "" {
		row := models.ActionRow{
			GuildID:      guildID,
			ActorUserID:  "warnings",
			TargetUserID: payload.UserID,
			Type:         action,
			PayloadJSON:  toJSON(map[string]any{"reason": fmt.Sprintf("Warning threshold reached (%d)", count)}),
		}
		if _, err := s.repos.Actions.Enqueue(r.Context(), row); err == nil {
			s.discord.NotifyActionQueued()
		}
	}

	if settings.WarningLogChannelID != "" {
		msg := fmt.Sprintf("Warning issued to <@%s> (count=%d).", payload.UserID, count)
		if payload.Reason != "" {
			msg += " Reason: " + payload.Reason
		}
		_, _ = s.discordSessionMessage(settings.WarningLogChannelID, msg)
	}

	writeJSON(w, map[string]any{
		"id":          id,
		"count":       count,
		"auto_action": action,
	})
}

func (s *Server) discordSessionMessage(channelID, content string) (string, error) {
	// Keep web handlers independent from discordgo import details.
	if s.discord == nil {
		return "", nil
	}
	return s.discord.SendChannelMessage(channelID, content)
}
