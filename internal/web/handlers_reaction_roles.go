package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

func (s *Server) handleReactionRoleRules(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := s.repos.ReactionRoles.ListByGuild(r.Context(), guildID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	case http.MethodPost:
		var payload struct {
			ChannelID       string `json:"channel_id"`
			MessageID       string `json:"message_id"`
			Emoji           string `json:"emoji"`
			RoleID          string `json:"role_id"`
			RemoveOnUnreact bool   `json:"remove_on_unreact"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.ChannelID) == "" || strings.TrimSpace(payload.MessageID) == "" || strings.TrimSpace(payload.Emoji) == "" || strings.TrimSpace(payload.RoleID) == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := s.repos.ReactionRoles.Create(r.Context(), models.ReactionRoleRule{
			GuildID:         guildID,
			ChannelID:       strings.TrimSpace(payload.ChannelID),
			MessageID:       strings.TrimSpace(payload.MessageID),
			Emoji:           normalizeEmoji(payload.Emoji),
			RoleID:          strings.TrimSpace(payload.RoleID),
			RemoveOnUnreact: payload.RemoveOnUnreact,
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

func (s *Server) handleReactionRoleRuleDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/modules/reaction-roles/rules/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.repos.ReactionRoles.Delete(r.Context(), guildID, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func normalizeEmoji(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "<:")
	v = strings.TrimPrefix(v, "<a:")
	v = strings.TrimSuffix(v, ">")
	parts := strings.Split(v, ":")
	if len(parts) == 2 {
		return parts[1]
	}
	return strings.TrimSpace(v)
}
