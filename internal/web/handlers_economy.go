package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ModularDevLabs/GoBot/internal/db"
)

func (s *Server) handleEconomyBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if guildID == "" || userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bal, err := s.repos.Economy.GetBalance(r.Context(), guildID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"user_id": userID, "balance": bal})
}

func (s *Server) handleEconomyLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	rows, err := s.repos.Economy.Leaderboard(r.Context(), guildID, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (s *Server) handleEconomyShop(w http.ResponseWriter, r *http.Request) {
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := s.repos.Economy.ListShopItems(r.Context(), guildID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	case http.MethodPost:
		var payload struct {
			Name            string `json:"name"`
			Cost            int    `json:"cost"`
			RoleID          string `json:"role_id"`
			DurationMinutes int    `json:"duration_minutes"`
			Enabled         bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		payload.Name = strings.TrimSpace(payload.Name)
		if payload.Name == "" || payload.Cost <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := s.repos.Economy.AddShopItem(r.Context(), db.ShopItemRow{
			GuildID:         guildID,
			Name:            payload.Name,
			Cost:            payload.Cost,
			RoleID:          strings.TrimSpace(payload.RoleID),
			DurationMinutes: payload.DurationMinutes,
			Enabled:         payload.Enabled,
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

func (s *Server) handleEconomyPurchase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := strings.TrimSpace(r.URL.Query().Get("guild_id"))
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var payload struct {
		UserID string `json:"user_id"`
		ItemID int64  `json:"item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.ItemID <= 0 {
		if raw := strings.TrimSpace(r.URL.Query().Get("item_id")); raw != "" {
			if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
				payload.ItemID = n
			}
		}
	}
	payload.UserID = strings.TrimSpace(payload.UserID)
	if payload.UserID == "" || payload.ItemID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result, err := s.discord.PurchaseShopItem(r.Context(), guildID, payload.UserID, payload.ItemID)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	writeJSON(w, map[string]any{"result": result})
}
