package web

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type triviaQuestion struct {
	Prompt string
	Answer string
}

var triviaBank = []triviaQuestion{
	{Prompt: "What command shows all channels in a Discord server settings export from Fundamentum? (single word)", Answer: "channels"},
	{Prompt: "In Discord, what permission is required to remove messages posted by others?", Answer: "manage messages"},
	{Prompt: "What does XP stand for in leveling systems?", Answer: "experience points"},
	{Prompt: "Which Discord object groups text channels together?", Answer: "category"},
	{Prompt: "What keyword starts a user mention in Discord markdown?", Answer: "<@"},
	{Prompt: "What HTTP method is typically used to create a new resource?", Answer: "post"},
	{Prompt: "What does RSVP mean for calendar events?", Answer: "please reply"},
	{Prompt: "What SQL keyword is used to sort rows?", Answer: "order by"},
}

func (s *Server) handleTriviaQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !s.ensureFeatureEnabled(w, r, guildID, models.FeatureTrivia, "trivia") {
		return
	}
	if len(triviaBank) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := rng.Intn(len(triviaBank))
	writeJSON(w, map[string]any{
		"id":       idx,
		"question": triviaBank[idx].Prompt,
	})
}

func (s *Server) handleTriviaAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !s.ensureFeatureEnabled(w, r, guildID, models.FeatureTrivia, "trivia") {
		return
	}
	var payload struct {
		UserID     string `json:"user_id"`
		QuestionID int    `json:"question_id"`
		Answer     string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload.UserID = strings.TrimSpace(payload.UserID)
	payload.Answer = strings.TrimSpace(payload.Answer)
	if payload.UserID == "" || payload.QuestionID < 0 || payload.QuestionID >= len(triviaBank) || payload.Answer == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expected := normalizeTriviaAnswer(triviaBank[payload.QuestionID].Answer)
	actual := normalizeTriviaAnswer(payload.Answer)
	correct := actual == expected
	if correct {
		if err := s.repos.Trivia.AddScore(r.Context(), guildID, payload.UserID, 1); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, map[string]any{
		"correct":         correct,
		"expected_answer": triviaBank[payload.QuestionID].Answer,
	})
}

func (s *Server) handleTriviaLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !s.ensureFeatureEnabled(w, r, guildID, models.FeatureTrivia, "trivia") {
		return
	}
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	rows, err := s.repos.Trivia.Leaderboard(r.Context(), guildID, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func normalizeTriviaAnswer(in string) string {
	normalized := strings.ToLower(strings.TrimSpace(in))
	replacer := strings.NewReplacer(".", "", ",", "", "!", "", "?", "", "'", "", "\"", "")
	return strings.Join(strings.Fields(replacer.Replace(normalized)), " ")
}
