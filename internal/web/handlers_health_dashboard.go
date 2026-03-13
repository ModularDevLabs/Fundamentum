package web

import (
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleHealthDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cfg, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	queued, err := s.repos.Actions.CountByStatus(r.Context(), guildID, "queued")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	running, err := s.repos.Actions.CountByStatus(r.Context(), guildID, "running")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	failed24h, err := s.repos.Actions.CountSince(r.Context(), guildID, time.Now().UTC().Add(-24*time.Hour), "failed")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	warnings24h, err := s.repos.Warnings.CountSince(r.Context(), guildID, time.Now().UTC().Add(-24*time.Hour))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tickets24h, err := s.repos.Tickets.CountCreatedSince(r.Context(), guildID, time.Now().UTC().Add(-24*time.Hour))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	backfills := s.discord.BackfillStatus()
	activeBackfills := 0
	for _, job := range backfills {
		if job.GuildID != guildID {
			continue
		}
		if job.Status == "queued" || job.Status == "running" {
			activeBackfills++
		}
	}
	incidentActive := cfg.IncidentModeEnabled
	if incidentActive && cfg.IncidentModeEndsAt != "" {
		if t, err := time.Parse(time.RFC3339, cfg.IncidentModeEndsAt); err == nil && time.Now().UTC().After(t) {
			incidentActive = false
		}
	}
	writeJSON(w, map[string]any{
		"timestamp":                  time.Now().UTC().Format(time.RFC3339),
		"guild_id":                   guildID,
		"actions_queued":             queued,
		"actions_running":            running,
		"actions_failed_24h":         failed24h,
		"warnings_24h":               warnings24h,
		"tickets_created_24h":        tickets24h,
		"backfills_active":           activeBackfills,
		"incident_mode_active":       incidentActive,
		"retention_enabled":          cfg.RetentionDays > 0,
		"retention_days":             cfg.RetentionDays,
		"action_dry_run":             cfg.ActionDryRun,
		"action_two_person_approval": cfg.ActionTwoPersonApproval,
	})
}
