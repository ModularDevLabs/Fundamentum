package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleTicketMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !settings.FeatureEnabled(models.FeatureTickets) {
		return
	}

	openPhrase := strings.TrimSpace(settings.TicketOpenPhrase)
	if openPhrase == "" {
		openPhrase = "!ticket"
	}
	closePhrase := strings.TrimSpace(settings.TicketClosePhrase)
	if closePhrase == "" {
		closePhrase = "!close"
	}

	content := strings.TrimSpace(m.Content)
	if settings.TicketInboxChannelID != "" && m.ChannelID == settings.TicketInboxChannelID && strings.HasPrefix(strings.ToLower(content), strings.ToLower(openPhrase)) {
		subject := strings.TrimSpace(content[len(openPhrase):])
		if subject == "" {
			subject = "support"
		}
		_, err := s.openTicket(ctx, m.GuildID, m.Author.ID, subject, settings)
		if err != nil {
			s.logger.Error("open ticket failed guild=%s user=%s err=%v", m.GuildID, m.Author.ID, err)
			return
		}
		_ = s.session.ChannelMessageDelete(m.ChannelID, m.ID)
		return
	}

	ticket, ok, err := s.repos.Tickets.GetByChannel(ctx, m.GuildID, m.ChannelID)
	if err != nil || !ok || ticket.Status != "open" {
		return
	}
	if strings.EqualFold(content, closePhrase) && s.canCloseTicket(m, ticket, settings) {
		if err := s.closeTicket(ctx, ticket, settings, "closed by command"); err != nil {
			s.logger.Error("close ticket failed id=%d guild=%s err=%v", ticket.ID, ticket.GuildID, err)
		}
	}
}

func (s *Service) canCloseTicket(m *discordgo.MessageCreate, ticket models.TicketRow, settings models.GuildSettings) bool {
	if m.Author != nil && m.Author.ID == ticket.CreatorUserID {
		return true
	}
	if settings.TicketSupportRoleID != "" && memberHasRole(m.Member, settings.TicketSupportRoleID) {
		return true
	}
	return false
}

func (s *Service) openTicket(ctx context.Context, guildID, creatorUserID, subject string, settings models.GuildSettings) (models.TicketRow, error) {
	everyoneDeny := int64(discordgo.PermissionViewChannel)
	userAllow := int64(discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory)
	overwrites := []*discordgo.PermissionOverwrite{
		{ID: guildID, Type: discordgo.PermissionOverwriteTypeRole, Allow: 0, Deny: everyoneDeny},
		{ID: creatorUserID, Type: discordgo.PermissionOverwriteTypeMember, Allow: userAllow, Deny: 0},
	}
	if settings.TicketSupportRoleID != "" {
		overwrites = append(overwrites, &discordgo.PermissionOverwrite{
			ID: settings.TicketSupportRoleID, Type: discordgo.PermissionOverwriteTypeRole, Allow: userAllow, Deny: 0,
		})
	}
	name := fmt.Sprintf("ticket-%s", strings.ToLower(creatorUserID))
	if len(name) > 90 {
		name = name[:90]
	}
	ch, err := s.session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:                 name,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             settings.TicketCategoryID,
		PermissionOverwrites: overwrites,
	})
	if err != nil {
		return models.TicketRow{}, err
	}
	row := models.TicketRow{
		GuildID:       guildID,
		ChannelID:     ch.ID,
		CreatorUserID: creatorUserID,
		Subject:       subject,
		Status:        "open",
	}
	id, err := s.repos.Tickets.Create(ctx, row)
	if err != nil {
		return models.TicketRow{}, err
	}
	row.ID = id
	row.CreatedAt = time.Now().UTC()
	_, _ = s.session.ChannelMessageSend(ch.ID, fmt.Sprintf("Ticket opened by <@%s>. Subject: %s\nType `%s` to close.", creatorUserID, subject, settings.TicketClosePhrase))
	if settings.TicketLogChannelID != "" {
		_, _ = s.session.ChannelMessageSend(settings.TicketLogChannelID, fmt.Sprintf("Ticket #%d opened by <@%s> in <#%s>.", id, creatorUserID, ch.ID))
	}
	return row, nil
}

func (s *Service) closeTicket(ctx context.Context, ticket models.TicketRow, settings models.GuildSettings, reason string) error {
	if err := s.repos.Tickets.Close(ctx, ticket.GuildID, ticket.ID); err != nil {
		return err
	}
	_, _ = s.session.ChannelMessageSend(ticket.ChannelID, "Ticket closed. Archiving channel.")
	if settings.TicketLogChannelID != "" {
		_, _ = s.session.ChannelMessageSend(settings.TicketLogChannelID, fmt.Sprintf("Ticket #%d closed (%s).", ticket.ID, reason))
	}
	_, err := s.session.ChannelDelete(ticket.ChannelID)
	return err
}

func (s *Service) CloseTicketByID(ctx context.Context, guildID string, id int64) error {
	t, ok, err := s.repos.Tickets.GetByID(ctx, guildID, id)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	settings, err := s.repos.Settings.Get(ctx, guildID)
	if err != nil {
		return err
	}
	return s.closeTicket(ctx, t, settings, "closed from dashboard")
}
