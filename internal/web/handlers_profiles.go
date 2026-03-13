package web

import (
	"encoding/json"
	"net/http"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) handleSettingsProfileApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var payload struct {
		Profile string `json:"profile"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cfg, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch payload.Profile {
	case "small_community":
		cfg.FeatureFlags[models.FeatureWelcomeMessages] = true
		cfg.FeatureFlags[models.FeatureGoodbyeMessages] = true
		cfg.FeatureFlags[models.FeatureSuggestions] = true
		cfg.FeatureFlags[models.FeaturePolls] = true
		cfg.FeatureFlags[models.FeatureCustomCommands] = true
		cfg.FeatureFlags[models.FeatureAutoMod] = true
		cfg.AutoModAction = "delete_warn"
		cfg.WarnQuarantineThreshold = 4
		cfg.WarnKickThreshold = 7
	case "gaming_server":
		cfg.FeatureFlags[models.FeatureLeveling] = true
		cfg.FeatureFlags[models.FeatureGiveaways] = true
		cfg.FeatureFlags[models.FeaturePolls] = true
		cfg.FeatureFlags[models.FeatureSuggestions] = true
		cfg.FeatureFlags[models.FeatureStarboard] = true
		cfg.FeatureFlags[models.FeatureScheduled] = true
		cfg.LevelingXPPerMessage = 15
		cfg.LevelingCooldownSeconds = 45
		cfg.StarboardThreshold = 4
	case "strict_moderation":
		cfg.FeatureFlags[models.FeatureAutoMod] = true
		cfg.FeatureFlags[models.FeatureWarnings] = true
		cfg.FeatureFlags[models.FeatureAntiRaid] = true
		cfg.FeatureFlags[models.FeatureAccountAgeGuard] = true
		cfg.FeatureFlags[models.FeatureAuditLogStream] = true
		cfg.FeatureFlags[models.FeatureMemberNotes] = true
		cfg.FeatureFlags[models.FeatureAppeals] = true
		cfg.AutoModAction = "delete_quarantine"
		cfg.WarnQuarantineThreshold = 2
		cfg.WarnKickThreshold = 4
		cfg.AntiRaidJoinThreshold = 5
		cfg.AntiRaidWindowSeconds = 20
		cfg.AntiRaidAction = "quarantine"
		cfg.AccountAgeAction = "quarantine"
		cfg.AccountAgeMinDays = 14
	default:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("unknown profile"))
		return
	}

	if err := s.repos.Settings.Upsert(r.Context(), cfg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, cfg)
}
