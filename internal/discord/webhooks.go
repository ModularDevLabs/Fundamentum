package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (s *Service) emitWebhookEvent(guildID, eventType, message string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rows, err := s.repos.Webhooks.ListEnabledByGuild(ctx, guildID)
		if err != nil {
			s.logger.Error("webhook list failed guild=%s err=%v", guildID, err)
			return
		}
		for _, row := range rows {
			if !webhookWantsEvent(row.Events, eventType) {
				continue
			}
			payload := map[string]any{
				"guild_id":   guildID,
				"event_type": eventType,
				"message":    message,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}
			body, _ := json.Marshal(payload)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, row.URL, bytes.NewReader(body))
			if err != nil {
				_ = s.repos.Webhooks.SetLastError(ctx, row.ID, err.Error())
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				_ = s.repos.Webhooks.SetLastError(ctx, row.ID, err.Error())
				continue
			}
			_ = resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				_ = s.repos.Webhooks.SetLastError(ctx, row.ID, "non-2xx response")
				continue
			}
			_ = s.repos.Webhooks.SetLastError(ctx, row.ID, "")
		}
	}()
}

func webhookWantsEvent(events []string, eventType string) bool {
	if len(events) == 0 {
		return true
	}
	for _, evt := range events {
		if strings.EqualFold(strings.TrimSpace(evt), eventType) {
			return true
		}
	}
	return false
}
