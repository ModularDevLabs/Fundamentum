package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

var pollOptionEmojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣"}

func (s *Server) handlePolls(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rows, err := s.repos.Polls.ListByGuild(r.Context(), guildID, 100)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (s *Server) handlePollStart(w http.ResponseWriter, r *http.Request) {
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
		ChannelID string   `json:"channel_id"`
		Question  string   `json:"question"`
		Options   []string `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	channelID := strings.TrimSpace(payload.ChannelID)
	if channelID == "" {
		channelID = strings.TrimSpace(settings.PollsChannelID)
	}
	question := strings.TrimSpace(payload.Question)
	if channelID == "" || question == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	options := make([]string, 0, 5)
	for _, option := range payload.Options {
		option = strings.TrimSpace(option)
		if option == "" {
			continue
		}
		options = append(options, option)
		if len(options) == 5 {
			break
		}
	}
	if len(options) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := s.repos.Polls.Create(r.Context(), models.PollRow{
		GuildID:   guildID,
		ChannelID: channelID,
		MessageID: "pending",
		Question:  question,
		Options:   options,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	body := fmt.Sprintf("📊 **Poll #%d**\n%s", id, question)
	for i, option := range options {
		body += fmt.Sprintf("\n%s %s", pollOptionEmojis[i], option)
	}
	body += "\nReact to vote."
	msgID, err := s.discord.SendChannelMessage(channelID, body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = s.repos.Polls.AttachMessageID(r.Context(), id, msgID)
	for i := range options {
		_ = s.discord.AddMessageReaction(channelID, msgID, pollOptionEmojis[i])
	}
	writeJSON(w, map[string]any{"id": id, "message_id": msgID})
}

func (s *Server) handlePollDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/modules/polls/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "close" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	poll, found, err := s.repos.Polls.GetByID(r.Context(), guildID, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	msg, err := s.discord.GetChannelMessage(poll.ChannelID, poll.MessageID)
	if err != nil || msg == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	votes := make([]int, len(poll.Options))
	total := 0
	for i := range poll.Options {
		emoji := pollOptionEmojis[i]
		for _, react := range msg.Reactions {
			if react == nil || react.Emoji == nil {
				continue
			}
			if react.Emoji.APIName() != emoji {
				continue
			}
			count := react.Count
			if count > 0 {
				count--
			}
			votes[i] = count
			total += count
			break
		}
	}
	result := fmt.Sprintf("📊 Poll #%d closed: %s", poll.ID, poll.Question)
	for i, option := range poll.Options {
		result += fmt.Sprintf("\n%s %s — %d", pollOptionEmojis[i], option, votes[i])
	}
	result += fmt.Sprintf("\nTotal votes: %d", total)
	_, _ = s.discord.SendChannelMessage(poll.ChannelID, result)
	_ = s.repos.Polls.MarkClosed(r.Context(), guildID, id)
	writeJSON(w, map[string]any{"total_votes": total})
}
