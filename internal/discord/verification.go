package discord

import (
	"strings"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleVerificationMessage(m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.FeatureEnabled(models.FeatureVerification) {
		return
	}
	if settings.VerificationChannelID == "" || settings.UnverifiedRoleID == "" {
		return
	}
	if m.ChannelID != settings.VerificationChannelID {
		return
	}
	phrase := strings.TrimSpace(settings.VerificationPhrase)
	if phrase == "" {
		phrase = "!verify"
	}
	if strings.TrimSpace(strings.ToLower(m.Content)) != strings.ToLower(phrase) {
		return
	}
	member, err := s.session.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		return
	}
	if !memberHasRole(member, settings.UnverifiedRoleID) {
		return
	}
	if err := s.session.GuildMemberRoleRemove(m.GuildID, m.Author.ID, settings.UnverifiedRoleID); err != nil {
		s.logger.Error("verification remove unverified failed guild=%s user=%s role=%s err=%v", m.GuildID, m.Author.ID, settings.UnverifiedRoleID, err)
		return
	}
	if settings.VerifiedRoleID != "" {
		if err := s.session.GuildMemberRoleAdd(m.GuildID, m.Author.ID, settings.VerifiedRoleID); err != nil {
			s.logger.Error("verification add verified failed guild=%s user=%s role=%s err=%v", m.GuildID, m.Author.ID, settings.VerifiedRoleID, err)
		}
	}
	_ = s.session.ChannelMessageDelete(m.ChannelID, m.ID)
	_, _ = s.session.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" verified successfully.")
}
