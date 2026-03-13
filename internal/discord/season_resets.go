package discord

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

var seasonResetAllowedModules = []string{"leveling", "economy", "trivia"}

func normalizeSeasonResetModules(modules []string) []string {
	out := make([]string, 0, len(modules))
	seen := map[string]struct{}{}
	for _, m := range modules {
		name := strings.TrimSpace(strings.ToLower(m))
		if name == "" || !slices.Contains(seasonResetAllowedModules, name) {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func nextSeasonResetAt(now time.Time, cadence string) time.Time {
	now = now.UTC()
	switch strings.ToLower(strings.TrimSpace(cadence)) {
	case "quarterly":
		month := int(now.Month())
		nextQuarterStartMonth := ((month-1)/3)*3 + 4
		year := now.Year()
		if nextQuarterStartMonth > 12 {
			nextQuarterStartMonth = 1
			year++
		}
		return time.Date(year, time.Month(nextQuarterStartMonth), 1, 0, 0, 0, 0, time.UTC)
	default:
		year := now.Year()
		month := now.Month() + 1
		if month > 12 {
			month = 1
			year++
		}
		return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	}
}

func (s *Service) runSeasonResetWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		guildIDs, err := s.repos.Settings.ListGuildIDs(ctx)
		if err != nil {
			s.logger.Error("season reset list guilds failed: %v", err)
			continue
		}
		now := time.Now().UTC()
		for _, guildID := range guildIDs {
			cfg, err := s.repos.Settings.Get(ctx, guildID)
			if err != nil {
				continue
			}
			if !cfg.FeatureEnabled(models.FeatureSeasonResets) || !cfg.SeasonResetsEnabled {
				continue
			}

			modules := normalizeSeasonResetModules(cfg.SeasonResetModules)
			if len(modules) == 0 {
				modules = append([]string{}, seasonResetAllowedModules...)
			}
			if strings.TrimSpace(cfg.SeasonResetNextRunAt) == "" {
				cfg.SeasonResetNextRunAt = nextSeasonResetAt(now, cfg.SeasonResetCadence).Format(time.RFC3339)
				_ = s.repos.Settings.Upsert(ctx, cfg)
				continue
			}
			nextAt, err := time.Parse(time.RFC3339, cfg.SeasonResetNextRunAt)
			if err != nil || now.Before(nextAt) {
				continue
			}
			if _, err := s.runSeasonReset(ctx, guildID, "scheduler", modules); err != nil {
				s.logger.Error("season reset run failed guild=%s err=%v", guildID, err)
				continue
			}
			cfg.SeasonResetNextRunAt = nextSeasonResetAt(now, cfg.SeasonResetCadence).Format(time.RFC3339)
			if err := s.repos.Settings.Upsert(ctx, cfg); err != nil {
				s.logger.Error("season reset next run update failed guild=%s err=%v", guildID, err)
			}
		}
	}
}

func (s *Service) RunSeasonResetNow(ctx context.Context, guildID, actor string) (models.SeasonResetRunRow, error) {
	cfg, err := s.repos.Settings.Get(ctx, guildID)
	if err != nil {
		return models.SeasonResetRunRow{}, err
	}
	modules := normalizeSeasonResetModules(cfg.SeasonResetModules)
	if len(modules) == 0 {
		modules = append([]string{}, seasonResetAllowedModules...)
	}
	trigger := strings.TrimSpace(actor)
	if trigger == "" {
		trigger = "manual"
	}
	row, err := s.runSeasonReset(ctx, guildID, trigger, modules)
	if err != nil {
		return models.SeasonResetRunRow{}, err
	}
	cfg.SeasonResetNextRunAt = nextSeasonResetAt(time.Now().UTC(), cfg.SeasonResetCadence).Format(time.RFC3339)
	_ = s.repos.Settings.Upsert(ctx, cfg)
	return row, nil
}

func (s *Service) runSeasonReset(ctx context.Context, guildID, triggeredBy string, modules []string) (models.SeasonResetRunRow, error) {
	s.seasonResetMu.Lock()
	if s.seasonResetRuns[guildID] {
		s.seasonResetMu.Unlock()
		return models.SeasonResetRunRow{}, fmt.Errorf("season reset already running")
	}
	s.seasonResetRuns[guildID] = true
	s.seasonResetMu.Unlock()
	defer func() {
		s.seasonResetMu.Lock()
		delete(s.seasonResetRuns, guildID)
		s.seasonResetMu.Unlock()
	}()

	startedAt := time.Now().UTC()
	affected := map[string]int64{}
	status := "success"
	errMsg := ""

	for _, module := range modules {
		switch module {
		case "leveling":
			n, err := s.repos.Leveling.ResetGuild(ctx, guildID)
			if err != nil {
				status = "failed"
				errMsg = fmt.Sprintf("leveling reset failed: %v", err)
				goto done
			}
			affected[module] = n
		case "economy":
			n, err := s.repos.Economy.ResetBalances(ctx, guildID)
			if err != nil {
				status = "failed"
				errMsg = fmt.Sprintf("economy reset failed: %v", err)
				goto done
			}
			affected[module] = n
		case "trivia":
			n, err := s.repos.Trivia.ResetScores(ctx, guildID)
			if err != nil {
				status = "failed"
				errMsg = fmt.Sprintf("trivia reset failed: %v", err)
				goto done
			}
			affected[module] = n
		}
	}

done:
	completedAt := time.Now().UTC()
	if err := s.repos.SeasonResets.RecordRun(ctx, guildID, triggeredBy, modules, affected, status, errMsg, startedAt, completedAt); err != nil {
		s.logger.Error("season reset record failed guild=%s err=%v", guildID, err)
	}
	row := models.SeasonResetRunRow{
		GuildID:      guildID,
		TriggeredBy:  triggeredBy,
		Modules:      append([]string{}, modules...),
		AffectedRows: affected,
		Status:       status,
		Error:        errMsg,
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
	}
	if status == "failed" {
		return row, fmt.Errorf(errMsg)
	}
	s.logger.Info("season reset completed guild=%s modules=%v", guildID, modules)
	return row, nil
}

func (s *Service) SeasonResetHistory(ctx context.Context, guildID string, limit int) ([]models.SeasonResetRunRow, error) {
	return s.repos.SeasonResets.ListRuns(ctx, guildID, limit)
}
