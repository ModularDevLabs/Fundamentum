package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) OnMessageCreate(_ *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}
	if m.GuildID == "" {
		return
	}

	username := m.Author.Username
	globalName := ""
	displayName := ""
	if m.Member != nil {
		displayName = m.Member.Nick
	}
	if displayName == "" {
		displayName = username
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := s.repos.Settings.Get(ctx, m.GuildID)
	if err != nil {
		settings = models.DefaultGuildSettings(m.GuildID)
	}
	if settings.FeatureAllowedInChannel(models.FeatureAppeals, m.ChannelID) {
		s.handleAppealMessage(ctx, m, settings)
	}
	if settings.FeatureAllowedInChannel(models.FeatureSuggestions, m.ChannelID) {
		s.handleSuggestionMessage(ctx, m, settings)
	}
	if settings.FeatureAllowedInChannel(models.FeatureTickets, m.ChannelID) {
		s.handleTicketMessage(ctx, m, settings)
	}
	if settings.FeatureAllowedInChannel(models.FeatureVerification, m.ChannelID) {
		s.handleVerificationMessage(m, settings)
	}
	if settings.FeatureAllowedInChannel(models.FeatureLeveling, m.ChannelID) {
		s.handleLevelingMessage(ctx, m, settings)
	}
	if settings.FeatureAllowedInChannel(models.FeatureKeywordAlerts, m.ChannelID) {
		s.handleKeywordAlerts(ctx, m, settings)
	}
	if settings.FeatureAllowedInChannel(models.FeatureAFK, m.ChannelID) {
		s.handleAFKMessage(ctx, m, settings)
	}
	handledCustomCommand := false
	if settings.FeatureAllowedInChannel(models.FeatureCustomCommands, m.ChannelID) {
		handledCustomCommand = s.handleCustomCommandMessage(ctx, m, settings)
	}
	if !handledCustomCommand && settings.FeatureAllowedInChannel(models.FeatureAutoMod, m.ChannelID) {
		s.handleAutoMod(ctx, m, settings)
	}
	cutoff := time.Now().AddDate(0, 0, -settings.InactiveDays)
	_, _ = s.repos.Activity.UpsertActivityIfStale(ctx, m.GuildID, m.Author.ID, m.ChannelID, m.Timestamp, username, globalName, displayName, cutoff)
}

func (s *Service) OnGuildMemberAdd(_ *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if m == nil || m.GuildID == "" || m.User == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := s.repos.Settings.Get(ctx, m.GuildID)
	if err != nil {
		return
	}
	s.handleAccountAgeGuardOnJoin(ctx, m, settings)
	s.handleAntiRaidOnJoin(ctx, m.GuildID, m.User.ID, settings)

	if settings.FeatureEnabled(models.FeatureWelcomeMessages) && settings.WelcomeChannelID != "" {
		content := renderMessageTemplate(settings.WelcomeMessage, m.User.Mention(), s.guildName(m.GuildID))
		if strings.TrimSpace(content) != "" {
			if _, err := s.session.ChannelMessageSend(settings.WelcomeChannelID, content); err != nil {
				s.logger.Error("welcome message failed guild=%s channel=%s user=%s err=%v", m.GuildID, settings.WelcomeChannelID, m.User.ID, err)
			}
		}
	}

	if settings.FeatureEnabled(models.FeatureVerification) && settings.UnverifiedRoleID != "" {
		if err := s.session.GuildMemberRoleAdd(m.GuildID, m.User.ID, settings.UnverifiedRoleID); err != nil {
			s.logger.Error("verification add unverified role failed guild=%s user=%s role=%s err=%v", m.GuildID, m.User.ID, settings.UnverifiedRoleID, err)
		}
		if settings.VerificationChannelID != "" {
			phrase := settings.VerificationPhrase
			if strings.TrimSpace(phrase) == "" {
				phrase = "!verify"
			}
			_, _ = s.session.ChannelMessageSend(settings.VerificationChannelID, fmt.Sprintf("%s welcome. Type `%s` here to verify.", m.User.Mention(), phrase))
		}
	}

	s.handleInviteTracking(m.GuildID, m.User)
}

func (s *Service) OnGuildMemberRemove(_ *discordgo.Session, m *discordgo.GuildMemberRemove) {
	if m == nil || m.GuildID == "" || m.User == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := s.repos.Settings.Get(ctx, m.GuildID)
	if err != nil || !settings.FeatureEnabled(models.FeatureGoodbyeMessages) || settings.GoodbyeChannelID == "" {
		return
	}
	content := renderMessageTemplate(settings.GoodbyeMessage, m.User.Username, s.guildName(m.GuildID))
	if strings.TrimSpace(content) == "" {
		return
	}
	if _, err := s.session.ChannelMessageSend(settings.GoodbyeChannelID, content); err != nil {
		s.logger.Error("goodbye message failed guild=%s channel=%s user=%s err=%v", m.GuildID, settings.GoodbyeChannelID, m.User.ID, err)
	}
}

func (s *Service) OnGuildBanAdd(_ *discordgo.Session, evt *discordgo.GuildBanAdd) {
	if evt == nil || evt.GuildID == "" || evt.User == nil {
		return
	}
	s.emitAuditEvent(evt.GuildID, "ban_add", fmt.Sprintf("User banned: %s (%s)", evt.User.Username, evt.User.ID))
}

func (s *Service) OnGuildBanRemove(_ *discordgo.Session, evt *discordgo.GuildBanRemove) {
	if evt == nil || evt.GuildID == "" || evt.User == nil {
		return
	}
	s.emitAuditEvent(evt.GuildID, "ban_remove", fmt.Sprintf("User unbanned: %s (%s)", evt.User.Username, evt.User.ID))
}

func (s *Service) OnGuildRoleCreate(_ *discordgo.Session, evt *discordgo.GuildRoleCreate) {
	if evt == nil || evt.GuildID == "" || evt.Role == nil {
		return
	}
	s.emitAuditEvent(evt.GuildID, "role_create", fmt.Sprintf("Role created: %s (%s)", evt.Role.Name, evt.Role.ID))
}

func (s *Service) OnGuildRoleUpdate(_ *discordgo.Session, evt *discordgo.GuildRoleUpdate) {
	if evt == nil || evt.GuildID == "" || evt.Role == nil {
		return
	}
	s.emitAuditEvent(evt.GuildID, "role_update", fmt.Sprintf("Role updated: %s (%s)", evt.Role.Name, evt.Role.ID))
}

func (s *Service) OnGuildRoleDelete(_ *discordgo.Session, evt *discordgo.GuildRoleDelete) {
	if evt == nil || evt.GuildID == "" || evt.RoleID == "" {
		return
	}
	s.emitAuditEvent(evt.GuildID, "role_delete", fmt.Sprintf("Role deleted: %s", evt.RoleID))
}

func (s *Service) OnChannelCreate(_ *discordgo.Session, evt *discordgo.ChannelCreate) {
	if evt == nil || evt.GuildID == "" || evt.ID == "" {
		return
	}
	name := evt.Name
	if name == "" {
		name = "<unnamed>"
	}
	s.emitAuditEvent(evt.GuildID, "channel_create", fmt.Sprintf("Channel created: %s (%s)", name, evt.ID))
}

func (s *Service) OnChannelUpdate(_ *discordgo.Session, evt *discordgo.ChannelUpdate) {
	if evt == nil || evt.GuildID == "" || evt.ID == "" {
		return
	}
	name := evt.Name
	if name == "" {
		name = "<unnamed>"
	}
	s.emitAuditEvent(evt.GuildID, "channel_update", fmt.Sprintf("Channel updated: %s (%s)", name, evt.ID))
}

func (s *Service) OnChannelDelete(_ *discordgo.Session, evt *discordgo.ChannelDelete) {
	if evt == nil || evt.GuildID == "" || evt.ID == "" {
		return
	}
	name := evt.Name
	if name == "" {
		name = "<unnamed>"
	}
	s.emitAuditEvent(evt.GuildID, "channel_delete", fmt.Sprintf("Channel deleted: %s (%s)", name, evt.ID))
}

func (s *Service) guildName(guildID string) string {
	if g, err := s.session.Guild(guildID); err == nil && g != nil && g.Name != "" {
		return g.Name
	}
	return "this server"
}

func renderMessageTemplate(template, user, server string) string {
	out := strings.TrimSpace(template)
	if out == "" {
		return ""
	}
	out = strings.ReplaceAll(out, "{user}", user)
	out = strings.ReplaceAll(out, "{server}", server)
	return out
}

func (s *Service) emitAuditEvent(guildID, eventType, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	settings, err := s.repos.Settings.Get(ctx, guildID)
	if err != nil {
		return
	}
	s.emitWebhookEvent(guildID, eventType, message)
	if settings.ImmutableAuditTrail {
		payload := map[string]any{
			"guild_id":   guildID,
			"event_type": eventType,
			"message":    message,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		}
		if err := s.repos.AuditTrail.Append(ctx, guildID, eventType, message, payload); err != nil {
			s.logger.Error("immutable audit append failed guild=%s type=%s err=%v", guildID, eventType, err)
		}
	}
	if !settings.FeatureEnabled(models.FeatureAuditLogStream) || settings.AuditLogChannelID == "" {
		return
	}
	allowed := false
	for _, t := range settings.AuditLogEventTypes {
		if t == eventType {
			allowed = true
			break
		}
	}
	if !allowed {
		return
	}
	if _, err := s.session.ChannelMessageSend(settings.AuditLogChannelID, fmt.Sprintf("[%s] %s", eventType, message)); err != nil {
		s.logger.Error("audit stream failed guild=%s channel=%s type=%s err=%v", guildID, settings.AuditLogChannelID, eventType, err)
	}
}
