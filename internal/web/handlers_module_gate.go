package web

import (
	"net/http"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

func (s *Server) ensureFeatureEnabled(w http.ResponseWriter, r *http.Request, guildID, featureKey, moduleName string) bool {
	cfg, err := s.repos.Settings.Get(r.Context(), guildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}
	if cfg.FeatureEnabled(featureKey) {
		return true
	}
	w.WriteHeader(http.StatusForbidden)
	if moduleName == "" {
		moduleName = featureKey
	}
	_, _ = w.Write([]byte(moduleName + " module is disabled"))
	return false
}

func isConfessionsModuleEnabled(cfg models.GuildSettings) bool {
	return cfg.FeatureEnabled(models.FeatureConfessions) && cfg.ConfessionsEnabled
}
