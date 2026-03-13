package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleMemberNotes(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		userID := r.URL.Query().Get("user_id")
		rows, err := s.repos.MemberNotes.List(r.Context(), guildID, userID, 200)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	case http.MethodPost:
		var payload struct {
			UserID string `json:"user_id"`
			Body   string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		actor := r.Header.Get("X-Actor-User")
		if actor == "" {
			actor = "dashboard"
		}
		id, err := s.repos.MemberNotes.Create(r.Context(), models.MemberNoteRow{
			GuildID:  guildID,
			UserID:   strings.TrimSpace(payload.UserID),
			AuthorID: actor,
			Body:     strings.TrimSpace(payload.Body),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if cfg, err := s.repos.Settings.Get(r.Context(), guildID); err == nil && strings.TrimSpace(cfg.NotesLogChannelID) != "" {
			_, _ = s.discord.SendChannelMessage(cfg.NotesLogChannelID, fmt.Sprintf("Member note #%d added for user %s by %s", id, strings.TrimSpace(payload.UserID), actor))
		}
		writeJSON(w, map[string]any{"id": id})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleMemberNoteDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/modules/member-notes/")
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
	if err := s.repos.MemberNotes.Resolve(r.Context(), guildID, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
