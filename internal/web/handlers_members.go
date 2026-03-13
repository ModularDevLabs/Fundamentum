package web

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	statusFilter := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	offset := parseInt(r.URL.Query().Get("offset"), 0)

	settings, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	activityRows, err := s.repos.Activity.ListMembersAll(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	merged := make([]models.MemberRow, 0)
	activityByUser := make(map[string]models.MemberRow, len(activityRows))
	for _, row := range activityRows {
		activityByUser[row.UserID] = row
	}

	members, err := s.discord.ListGuildMembers(r.Context(), guildID)
	if err != nil {
		// Fallback: if guild member list fails, return activity-based rows.
		members = nil
	}
	if len(members) > 0 {
		for _, m := range members {
			if m == nil || m.User == nil {
				continue
			}
			row := models.MemberRow{
				GuildID: guildID,
				UserID:  m.User.ID,
			}
			if stored, ok := activityByUser[m.User.ID]; ok {
				row = stored
			} else {
				row.Username = m.User.Username
				row.DisplayName = m.Nick
				if row.DisplayName == "" {
					row.DisplayName = m.User.Username
				}
			}
			if settings.QuarantineRoleID != "" {
				for _, roleID := range m.Roles {
					if roleID == settings.QuarantineRoleID {
						row.Quarantined = true
						break
					}
				}
			}
			merged = append(merged, row)
		}
	} else {
		merged = activityRows
	}

	cutoff := time.Now().AddDate(0, 0, -settings.InactiveDays)
	filtered := make([]models.MemberRow, 0, len(merged))
	searchLower := strings.ToLower(strings.TrimSpace(search))
	for _, row := range merged {
		row.Status = statusFromLast(row.LastMessageAt, cutoff)
		if searchLower != "" {
			target := strings.ToLower(row.UserID + " " + row.Username + " " + row.GlobalName + " " + row.DisplayName)
			if !strings.Contains(target, searchLower) {
				continue
			}
		}
		if statusFilter != "" && row.Status != statusFilter {
			continue
		}
		filtered = append(filtered, row)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		a := strings.ToLower(filtered[i].DisplayName)
		if a == "" {
			a = strings.ToLower(filtered[i].Username)
		}
		b := strings.ToLower(filtered[j].DisplayName)
		if b == "" {
			b = strings.ToLower(filtered[j].Username)
		}
		return a < b
	})

	if offset > len(filtered) {
		offset = len(filtered)
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	out := filtered[offset:end]

	writeJSON(w, out)
}

func (s *Server) handleMemberDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID := parseIDFromPath(r.URL.Path, "/api/members/")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	row, ok, err := s.repos.Activity.GetMember(r.Context(), guildID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	settings, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cutoff := time.Now().AddDate(0, 0, -settings.InactiveDays)
	row.Status = statusFromLast(row.LastMessageAt, cutoff)
	writeJSON(w, row)
}

func statusFromLast(last *time.Time, cutoff time.Time) string {
	if last == nil {
		return "inactive"
	}
	if last.Before(cutoff) {
		return "inactive"
	}
	return "active"
}
