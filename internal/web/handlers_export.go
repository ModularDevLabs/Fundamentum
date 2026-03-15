package web

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	exportType := strings.TrimSpace(r.URL.Query().Get("type"))
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "json"
	}
	filename := fmt.Sprintf("%s_%s_%s.%s", exportType, guildID, time.Now().UTC().Format("20060102T150405"), format)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	switch exportType {
	case "members":
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
		rows := make([]models.MemberRow, 0, len(activityRows))
		activityByUser := make(map[string]models.MemberRow, len(activityRows))
		for _, row := range activityRows {
			activityByUser[row.UserID] = row
		}

		members, err := s.discord.ListGuildMembers(r.Context(), guildID)
		if err != nil {
			// Fallback: export activity rows when live guild member fetch is unavailable.
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
				rows = append(rows, row)
			}
		} else {
			rows = activityRows
		}

		cutoff := time.Now().AddDate(0, 0, -settings.InactiveDays)
		for i := range rows {
			rows[i].Status = statusFromLast(rows[i].LastMessageAt, cutoff)
		}
		sort.SliceStable(rows, func(i, j int) bool {
			a := strings.ToLower(rows[i].DisplayName)
			if a == "" {
				a = strings.ToLower(rows[i].Username)
			}
			b := strings.ToLower(rows[j].DisplayName)
			if b == "" {
				b = strings.ToLower(rows[j].Username)
			}
			return a < b
		})
		if format == "csv" {
			w.Header().Set("Content-Type", "text/csv")
			cw := csv.NewWriter(w)
			_ = cw.Write([]string{"user_id", "username", "global_name", "display_name", "status", "quarantined", "last_message_at", "last_channel_id"})
			for _, row := range rows {
				last := ""
				if row.LastMessageAt != nil {
					last = row.LastMessageAt.UTC().Format(time.RFC3339)
				}
				_ = cw.Write([]string{row.UserID, row.Username, row.GlobalName, row.DisplayName, row.Status, strconv.FormatBool(row.Quarantined), last, row.LastChannelID})
			}
			cw.Flush()
			return
		}
		writeJSON(w, rows)
	case "actions":
		rows, err := s.repos.Actions.List(r.Context(), guildID, "", 5000, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if format == "csv" {
			w.Header().Set("Content-Type", "text/csv")
			cw := csv.NewWriter(w)
			_ = cw.Write([]string{"id", "target_user_id", "type", "status", "actor_user_id", "created_at", "error"})
			for _, row := range rows {
				_ = cw.Write([]string{
					strconv.FormatInt(row.ID, 10),
					row.TargetUserID,
					row.Type,
					row.Status,
					row.ActorUserID,
					row.CreatedAt.UTC().Format(time.RFC3339),
					row.Error,
				})
			}
			cw.Flush()
			return
		}
		writeJSON(w, rows)
	case "warnings":
		rows, err := s.repos.Warnings.ListByGuild(r.Context(), guildID, 5000)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if format == "csv" {
			w.Header().Set("Content-Type", "text/csv")
			cw := csv.NewWriter(w)
			_ = cw.Write([]string{"id", "user_id", "actor_user_id", "reason", "created_at"})
			for _, row := range rows {
				_ = cw.Write([]string{
					strconv.FormatInt(row.ID, 10),
					row.UserID,
					row.ActorUserID,
					row.Reason,
					row.CreatedAt.UTC().Format(time.RFC3339),
				})
			}
			cw.Flush()
			return
		}
		writeJSON(w, rows)
	case "tickets":
		rows, err := s.repos.Tickets.ListByGuild(r.Context(), guildID, "", 5000)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if format == "csv" {
			w.Header().Set("Content-Type", "text/csv")
			cw := csv.NewWriter(w)
			_ = cw.Write([]string{"id", "creator_user_id", "subject", "status", "created_at", "closed_at"})
			for _, row := range rows {
				closed := ""
				if row.ClosedAt != nil {
					closed = row.ClosedAt.UTC().Format(time.RFC3339)
				}
				_ = cw.Write([]string{
					strconv.FormatInt(row.ID, 10),
					row.CreatorUserID,
					row.Subject,
					row.Status,
					row.CreatedAt.UTC().Format(time.RFC3339),
					closed,
				})
			}
			cw.Flush()
			return
		}
		writeJSON(w, rows)
	case "cases":
		userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
		if userID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("user_id required for cases export"))
			return
		}
		req, _ := http.NewRequest(http.MethodGet, "/api/cases", nil)
		q := req.URL.Query()
		q.Set("guild_id", guildID)
		q.Set("user_id", userID)
		q.Set("limit", "1000")
		req.URL.RawQuery = q.Encode()
		rr := newResponseCapture()
		s.handleCases(rr, req.WithContext(r.Context()))
		if rr.statusCode >= 400 {
			w.WriteHeader(rr.statusCode)
			_, _ = w.Write(rr.body.Bytes())
			return
		}
		var rows []caseTimelineItem
		if err := json.Unmarshal(rr.body.Bytes(), &rows); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if format == "csv" {
			w.Header().Set("Content-Type", "text/csv")
			cw := csv.NewWriter(w)
			_ = cw.Write([]string{"time", "type", "actor", "summary"})
			for _, row := range rows {
				_ = cw.Write([]string{row.Time.UTC().Format(time.RFC3339), row.Type, row.Actor, row.Summary})
			}
			cw.Flush()
			return
		}
		writeJSON(w, rows)
	default:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("unknown export type"))
	}
}

type responseCapture struct {
	header     http.Header
	body       bytes.Buffer
	statusCode int
}

func newResponseCapture() *responseCapture {
	return &responseCapture{
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}
}

func (r *responseCapture) Header() http.Header {
	return r.header
}

func (r *responseCapture) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseCapture) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}
