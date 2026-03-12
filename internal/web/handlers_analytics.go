package web

import (
	"net/http"
	"time"
)

type analyticsTrendRow struct {
	Day      string `json:"day"`
	Warnings int    `json:"warnings"`
	Actions  int    `json:"actions"`
	Tickets  int    `json:"tickets"`
}

func (s *Server) handleAnalyticsTrends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	days := parseInt(r.URL.Query().Get("days"), 14)
	if days <= 0 {
		days = 14
	}
	if days > 90 {
		days = 90
	}

	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	rows := make([]analyticsTrendRow, 0, days)

	for i := days - 1; i >= 0; i-- {
		dayStart := startOfToday.AddDate(0, 0, -i)
		nextDay := dayStart.Add(24 * time.Hour)

		wFrom, err := s.repos.Warnings.CountSince(r.Context(), guildID, dayStart)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		wTo, err := s.repos.Warnings.CountSince(r.Context(), guildID, nextDay)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		aFrom, err := s.repos.Actions.CountSince(r.Context(), guildID, dayStart, "")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		aTo, err := s.repos.Actions.CountSince(r.Context(), guildID, nextDay, "")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tFrom, err := s.repos.Tickets.CountCreatedSince(r.Context(), guildID, dayStart)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tTo, err := s.repos.Tickets.CountCreatedSince(r.Context(), guildID, nextDay)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rows = append(rows, analyticsTrendRow{
			Day:      dayStart.Format("2006-01-02"),
			Warnings: wFrom - wTo,
			Actions:  aFrom - aTo,
			Tickets:  tFrom - tTo,
		})
	}

	writeJSON(w, rows)
}
