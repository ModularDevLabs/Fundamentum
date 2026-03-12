package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
)

type guildBackupSnapshot struct {
	GuildID           string                       `json:"guild_id"`
	ExportedAt        string                       `json:"exported_at"`
	Settings          models.GuildSettings         `json:"settings"`
	ReactionRules     []models.ReactionRoleRule    `json:"reaction_rules"`
	ScheduledMessages []models.ScheduledMessageRow `json:"scheduled_messages"`
	CustomCommands    []models.CustomCommandRow    `json:"custom_commands"`
}

func (s *Server) handleBackupExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cfg, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rules, err := s.repos.ReactionRoles.ListByGuild(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	schedules, err := s.repos.Scheduled.ListByGuild(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	commands, err := s.repos.CustomCommands.ListByGuild(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	snap := guildBackupSnapshot{
		GuildID:           guildID,
		ExportedAt:        time.Now().UTC().Format(time.RFC3339),
		Settings:          cfg,
		ReactionRules:     rules,
		ScheduledMessages: schedules,
		CustomCommands:    commands,
	}
	name := fmt.Sprintf("gobot_backup_%s_%s.json", guildID, time.Now().UTC().Format("20060102T150405"))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	_ = json.NewEncoder(w).Encode(snap)
}

func (s *Server) handleBackupImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var snap guildBackupSnapshot
	if err := json.NewDecoder(r.Body).Decode(&snap); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if snap.GuildID != "" && snap.GuildID != guildID {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("backup guild mismatch"))
		return
	}
	cfg := snap.Settings
	cfg.GuildID = guildID
	if err := s.repos.Settings.Upsert(r.Context(), cfg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := s.repos.ReactionRoles.DeleteAllByGuild(r.Context(), guildID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, row := range snap.ReactionRules {
		row.GuildID = guildID
		if _, err := s.repos.ReactionRoles.Create(r.Context(), row); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	if err := s.repos.Scheduled.DeleteAllByGuild(r.Context(), guildID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, row := range snap.ScheduledMessages {
		row.GuildID = guildID
		if row.NextRunAt.IsZero() {
			row.NextRunAt = time.Now().UTC().Add(5 * time.Minute)
		}
		if _, err := s.repos.Scheduled.Create(r.Context(), row); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	if err := s.repos.CustomCommands.DeleteAllByGuild(r.Context(), guildID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, row := range snap.CustomCommands {
		row.GuildID = guildID
		if _, err := s.repos.CustomCommands.Create(r.Context(), row); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, map[string]any{
		"restored": map[string]any{
			"reaction_rules":     len(snap.ReactionRules),
			"scheduled_messages": len(snap.ScheduledMessages),
			"custom_commands":    len(snap.CustomCommands),
		},
	})
}
