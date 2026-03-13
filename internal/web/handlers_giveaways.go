package web

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleGiveaways(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rows, err := s.repos.Giveaways.ListByGuild(r.Context(), guildID, 100)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (s *Server) handleGiveawayStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	settings, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var payload struct {
		ChannelID       string `json:"channel_id"`
		Prize           string `json:"prize"`
		DurationMinutes int    `json:"duration_minutes"`
		WinnerCount     int    `json:"winner_count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	channelID := strings.TrimSpace(payload.ChannelID)
	if channelID == "" {
		channelID = strings.TrimSpace(settings.GiveawaysChannelID)
	}
	if channelID == "" || strings.TrimSpace(payload.Prize) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.DurationMinutes <= 0 {
		payload.DurationMinutes = 60
	}
	if payload.WinnerCount <= 0 {
		payload.WinnerCount = 1
	}
	endsAt := time.Now().UTC().Add(time.Duration(payload.DurationMinutes) * time.Minute)
	id, err := s.repos.Giveaways.Create(r.Context(), models.GiveawayRow{
		GuildID:     guildID,
		ChannelID:   channelID,
		MessageID:   "pending",
		Prize:       strings.TrimSpace(payload.Prize),
		WinnerCount: payload.WinnerCount,
		EndsAt:      endsAt,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	emoji := strings.TrimSpace(settings.GiveawaysReactionEmoji)
	if emoji == "" {
		emoji = "🎉"
	}
	body := fmt.Sprintf("🎁 **Giveaway #%d**\nPrize: **%s**\nReact with %s to enter.\nWinners: %d\nEnds: %s", id, strings.TrimSpace(payload.Prize), emoji, payload.WinnerCount, endsAt.Local().Format(time.RFC1123))
	msgID, err := s.discord.SendChannelMessage(channelID, body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = s.repos.Giveaways.AttachMessageID(r.Context(), id, msgID)
	writeJSON(w, map[string]any{"id": id, "message_id": msgID})
}

func (s *Server) handleGiveawayDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/modules/giveaways/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "draw" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	giveaway, found, err := s.repos.Giveaways.GetByID(r.Context(), guildID, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	entries, err := s.repos.Giveaways.ListEntries(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	winners := drawWinners(entries, giveaway.WinnerCount)
	_ = s.repos.Giveaways.MarkEnded(r.Context(), id)
	if len(winners) == 0 {
		_, _ = s.discord.SendChannelMessage(giveaway.ChannelID, fmt.Sprintf("Giveaway #%d ended. No valid entries.", id))
		writeJSON(w, map[string]any{"winners": []string{}})
		return
	}
	mentions := make([]string, 0, len(winners))
	for _, userID := range winners {
		mentions = append(mentions, "<@"+userID+">")
	}
	_, _ = s.discord.SendChannelMessage(giveaway.ChannelID, fmt.Sprintf("Giveaway #%d ended. Winner(s): %s", id, strings.Join(mentions, ", ")))
	writeJSON(w, map[string]any{"winners": winners})
}

func drawWinners(entries []string, count int) []string {
	if count <= 0 || len(entries) == 0 {
		return []string{}
	}
	if count > len(entries) {
		count = len(entries)
	}
	cp := append([]string{}, entries...)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(cp) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		cp[i], cp[j] = cp[j], cp[i]
	}
	return cp[:count]
}
