package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

func (s *Server) handleActions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		guildID := r.URL.Query().Get("guild_id")
		if guildID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		status := r.URL.Query().Get("status")
		limit := parseInt(r.URL.Query().Get("limit"), 50)
		offset := parseInt(r.URL.Query().Get("offset"), 0)

		rows, err := s.repos.Actions.List(r.Context(), guildID, status, limit, offset)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.enrichActionTargets(r.Context(), guildID, rows)
		writeJSON(w, rows)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleActionDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/actions/")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if strings.HasPrefix(path, "quarantine") || strings.HasPrefix(path, "kick") || strings.HasPrefix(path, "remove-roles") {
		s.handleActionCreate(w, r)
		return
	}

	segments := strings.Split(path, "/")
	id, err := strconv.ParseInt(segments[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(segments) == 1 {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		row, ok, err := s.repos.Actions.Get(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rows := []models.ActionRow{row}
		s.enrichActionTargets(r.Context(), row.GuildID, rows)
		row = rows[0]
		writeJSON(w, row)
		return
	}

	if len(segments) == 2 && segments[1] == "retry" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := s.repos.Actions.UpdateStatus(r.Context(), id, "queued", ""); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.discord.NotifyActionQueued()
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (s *Server) enrichActionTargets(ctx context.Context, guildID string, rows []models.ActionRow) {
	for i := range rows {
		targetID := rows[i].TargetUserID
		if targetID == "" {
			continue
		}
		if member, ok, err := s.repos.Activity.GetMember(ctx, guildID, targetID); err == nil && ok {
			if member.DisplayName != "" {
				rows[i].TargetName = member.DisplayName
				continue
			}
			if member.Username != "" {
				rows[i].TargetName = member.Username
				continue
			}
		}
		rows[i].TargetName = s.discord.ResolveMemberDisplayName(guildID, targetID)
		if rows[i].TargetName == "" {
			var payload struct {
				TargetName string `json:"target_name"`
			}
			if err := json.Unmarshal([]byte(rows[i].PayloadJSON), &payload); err == nil && payload.TargetName != "" {
				rows[i].TargetName = payload.TargetName
			}
		}
	}
}

func (s *Server) handleActionCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/actions/")
	path = strings.TrimSuffix(path, "/")

	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload struct {
		UserIDs              []string          `json:"user_ids"`
		Reason               string            `json:"reason"`
		RoleIDs              []string          `json:"role_ids"`
		RemoveAllExceptAllow bool              `json:"remove_all_except_allowlist"`
		TargetName           string            `json:"target_name"`
		TargetNames          map[string]string `json:"target_names"`
		ConfirmToken         string            `json:"confirm_token"`
		ApproverUser         string            `json:"approver_user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	actor := r.Header.Get("X-Actor-User")
	if actor == "" {
		actor = "dashboard"
	}
	settings, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	actionType := strings.ReplaceAll(path, "-", "_")
	isDestructive := actionType == "kick" || actionType == "quarantine" || actionType == "remove_roles"
	if settings.InMaintenanceWindow(time.Now().UTC()) && isDestructive {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("maintenance window active; destructive actions are paused"))
		return
	}
	incidentActive := settings.IncidentModeEnabled
	if incidentActive && settings.IncidentModeEndsAt != "" {
		if t, err := time.Parse(time.RFC3339, settings.IncidentModeEndsAt); err == nil && time.Now().UTC().After(t) {
			incidentActive = false
		}
	}
	if isDestructive && settings.ActionRequireConfirm {
		if strings.TrimSpace(strings.ToUpper(payload.ConfirmToken)) != "CONFIRM" {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("missing confirm token"))
			return
		}
	}
	if incidentActive && isDestructive {
		if strings.TrimSpace(strings.ToUpper(payload.ConfirmToken)) != "CONFIRM" {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("incident mode requires confirm token"))
			return
		}
	}
	if isDestructive && settings.ActionTwoPersonApproval {
		approver := strings.TrimSpace(payload.ApproverUser)
		if approver == "" || strings.EqualFold(approver, actor) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("two-person approval requires distinct approver user"))
			return
		}
	}
	if incidentActive && isDestructive && actionType != "quarantine" {
		approver := strings.TrimSpace(payload.ApproverUser)
		if approver == "" || strings.EqualFold(approver, actor) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("incident mode requires distinct approver user for this action"))
			return
		}
	}
	ids := make([]int64, 0, len(payload.UserIDs))
	if settings.ActionDryRun {
		writeJSON(w, map[string]any{
			"dry_run":      true,
			"action_type":  actionType,
			"target_count": len(payload.UserIDs),
			"reason":       strings.TrimSpace(payload.Reason),
		})
		return
	}
	for _, id := range payload.UserIDs {
		targetName := payload.TargetName
		if payload.TargetNames != nil {
			if name, ok := payload.TargetNames[id]; ok && name != "" {
				targetName = name
			}
		}
		row := models.ActionRow{
			GuildID:      guildID,
			ActorUserID:  actor,
			TargetUserID: id,
			Type:         actionType,
			PayloadJSON:  toJSON(map[string]any{"reason": payload.Reason, "role_ids": payload.RoleIDs, "remove_all_except_allowlist": payload.RemoveAllExceptAllow, "target_name": targetName}),
		}
		newID, err := s.repos.Actions.Enqueue(r.Context(), row)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ids = append(ids, newID)
	}
	s.discord.NotifyActionQueued()
	writeJSON(w, map[string]any{"action_ids": ids})
}

func (s *Server) handleActionPreflight(w http.ResponseWriter, r *http.Request) {
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
		ActionType string   `json:"action_type"`
		UserIDs    []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload.ActionType = strings.TrimSpace(strings.ReplaceAll(payload.ActionType, "-", "_"))
	if payload.ActionType == "" || len(payload.UserIDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	results := make([]any, 0, len(payload.UserIDs))
	for _, userID := range payload.UserIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			continue
		}
		res, err := s.discord.PreflightAction(r.Context(), guildID, userID, payload.ActionType)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		results = append(results, res)
	}
	writeJSON(w, map[string]any{"results": results})
}

func toJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
