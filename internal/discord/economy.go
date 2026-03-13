package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleEconomyEarn(ctx context.Context, m *discordgo.MessageCreate) {
	if m == nil || m.Author == nil || m.Author.Bot {
		return
	}
	key := fmt.Sprintf("%s:%s", m.GuildID, m.Author.ID)
	now := time.Now().UTC()
	s.economyMu.Lock()
	last := s.economyLast[key]
	if !last.IsZero() && now.Sub(last) < 60*time.Second {
		s.economyMu.Unlock()
		return
	}
	s.economyLast[key] = now
	s.economyMu.Unlock()
	_ = s.repos.Economy.AddBalance(ctx, m.GuildID, m.Author.ID, 1)
}

func (s *Service) grantShopItem(ctx context.Context, guildID, userID string, itemID int64) (string, error) {
	item, found, err := s.repos.Economy.GetShopItem(ctx, guildID, itemID)
	if err != nil {
		return "", err
	}
	if !found || !item.Enabled {
		return "", fmt.Errorf("shop item not found")
	}
	balance, err := s.repos.Economy.GetBalance(ctx, guildID, userID)
	if err != nil {
		return "", err
	}
	if balance < item.Cost {
		return "", fmt.Errorf("insufficient balance")
	}
	if err := s.repos.Economy.AddBalance(ctx, guildID, userID, -item.Cost); err != nil {
		return "", err
	}
	if strings.TrimSpace(item.RoleID) != "" {
		if err := s.session.GuildMemberRoleAdd(guildID, userID, strings.TrimSpace(item.RoleID)); err != nil {
			return "", err
		}
		return "role granted", nil
	}
	return "purchase recorded", nil
}

func (s *Service) PurchaseShopItem(ctx context.Context, guildID, userID string, itemID int64) (string, error) {
	return s.grantShopItem(ctx, guildID, userID, itemID)
}
