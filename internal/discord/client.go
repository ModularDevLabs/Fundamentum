package discord

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ModularDevLabs/GoBot/internal/db"
	"github.com/ModularDevLabs/GoBot/internal/models"
	"github.com/bwmarrin/discordgo"
)

type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
}

type Service struct {
	session *discordgo.Session
	repos   *db.Repositories
	logger  Logger

	guildsCache  []models.GuildInfo
	backfillMu   sync.Mutex
	backfillJobs map[string]BackfillJob
	actionWakeCh chan struct{}
	invitesMu    sync.Mutex
	invitesCache map[string]map[string]int
	automodMu    sync.Mutex
	automodSeen  map[string][]time.Time
	raidMu       sync.Mutex
	raidJoins    map[string][]time.Time
	raidUntil    map[string]time.Time
}

func NewService(token string, repos *db.Repositories, logger Logger) (*Service, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions | discordgo.IntentsMessageContent

	svc := &Service{
		session:      s,
		repos:        repos,
		logger:       logger,
		actionWakeCh: make(chan struct{}, 1),
		invitesCache: make(map[string]map[string]int),
		automodSeen:  make(map[string][]time.Time),
		raidJoins:    make(map[string][]time.Time),
		raidUntil:    make(map[string]time.Time),
	}

	s.AddHandler(svc.onReady)
	s.AddHandler(svc.onGuildCreate)
	s.AddHandler(svc.OnMessageCreate)
	s.AddHandler(svc.OnGuildMemberAdd)
	s.AddHandler(svc.OnGuildMemberRemove)
	s.AddHandler(svc.OnGuildBanAdd)
	s.AddHandler(svc.OnGuildBanRemove)
	s.AddHandler(svc.OnGuildRoleCreate)
	s.AddHandler(svc.OnGuildRoleUpdate)
	s.AddHandler(svc.OnGuildRoleDelete)
	s.AddHandler(svc.OnChannelCreate)
	s.AddHandler(svc.OnChannelUpdate)
	s.AddHandler(svc.OnChannelDelete)
	s.AddHandler(svc.OnInviteCreate)
	s.AddHandler(svc.OnInviteDelete)
	s.AddHandler(svc.OnMessageReactionAdd)
	s.AddHandler(svc.OnMessageReactionRemove)
	svc.ensureBackfillInit()

	return svc, nil
}

func (s *Service) Open() error {
	return s.session.Open()
}

func (s *Service) Close() error {
	return s.session.Close()
}

func (s *Service) SendChannelMessage(channelID, content string) (string, error) {
	msg, err := s.session.ChannelMessageSend(channelID, content)
	if err != nil || msg == nil {
		return "", err
	}
	return msg.ID, nil
}

func (s *Service) StartWorkers(ctx context.Context) {
	go s.runActionWorker(ctx)
	go s.runScheduledWorker(ctx)
	go s.runTicketWorker(ctx)
	go s.runAnalyticsWorker(ctx)
}

func (s *Service) ListGuilds(ctx context.Context) ([]models.GuildInfo, error) {
	if len(s.guildsCache) > 0 {
		return s.guildsCache, nil
	}
	state := s.session.State
	if state != nil && len(state.Guilds) > 0 {
		out := make([]models.GuildInfo, 0, len(state.Guilds))
		for _, g := range state.Guilds {
			name := g.Name
			if name == "" {
				if full, err := s.session.Guild(g.ID); err == nil && full != nil && full.Name != "" {
					name = full.Name
				}
			}
			out = append(out, models.GuildInfo{ID: g.ID, Name: name})
		}
		s.guildsCache = out
		return out, nil
	}

	guilds, err := s.session.UserGuilds(200, "", "")
	if err != nil {
		return nil, err
	}
	out := make([]models.GuildInfo, 0, len(guilds))
	for _, g := range guilds {
		name := g.Name
		if name == "" {
			if full, err := s.session.Guild(g.ID); err == nil && full != nil && full.Name != "" {
				name = full.Name
			}
		}
		out = append(out, models.GuildInfo{ID: g.ID, Name: name})
	}
	s.guildsCache = out
	return out, nil
}

func (s *Service) ListGuildMembers(ctx context.Context, guildID string) ([]*discordgo.Member, error) {
	const pageSize = 1000
	after := ""
	out := make([]*discordgo.Member, 0)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		page, err := s.session.GuildMembers(guildID, after, pageSize)
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}
		out = append(out, page...)
		if len(page) < pageSize {
			break
		}
		after = page[len(page)-1].User.ID
	}

	return out, nil
}

func (s *Service) ResolveMemberDisplayName(guildID, userID string) string {
	if m, err := s.session.GuildMember(guildID, userID); err == nil && m != nil && m.User != nil {
		if m.Nick != "" {
			return m.Nick
		}
		if m.User.Username != "" {
			return m.User.Username
		}
	}
	return ""
}

func (s *Service) onReady(_ *discordgo.Session, _ *discordgo.Ready) {
	s.ensureGuildProvisioning("ready")
}

func (s *Service) onGuildCreate(_ *discordgo.Session, evt *discordgo.GuildCreate) {
	if evt == nil || evt.Guild == nil {
		return
	}
	if evt.Unavailable {
		return
	}
	s.updateGuildCache(evt.Guild.ID, evt.Guild.Name)
	s.ensureSingleGuildProvisioning(evt.Guild.ID, "guild_create")
}

func (s *Service) ensureGuildProvisioning(source string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	guilds, err := s.ListGuilds(ctx)
	if err != nil {
		s.logger.Error("fetch guilds failed: %v", err)
		return
	}
	for _, g := range guilds {
		s.ensureSingleGuildProvisioningWithContext(ctx, g.ID, source)
	}
}

func (s *Service) ensureSingleGuildProvisioning(guildID, source string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.ensureSingleGuildProvisioningWithContext(ctx, guildID, source)
}

func (s *Service) ensureSingleGuildProvisioningWithContext(ctx context.Context, guildID, source string) {
	cfg, err := s.repos.Settings.EnsureDefaults(ctx, guildID)
	if err != nil {
		s.logger.Error("ensure settings for guild %s failed source=%s: %v", guildID, source, err)
		return
	}
	if err := s.EnsureQuarantineBaseAssets(ctx, guildID, cfg); err != nil {
		s.logger.Error("ensure quarantine base assets for guild %s failed source=%s: %v", guildID, source, err)
		return
	}
	s.refreshInviteCache(guildID)
	s.logger.Info("guild provisioning complete guild=%s source=%s", guildID, source)
}

func (s *Service) updateGuildCache(guildID, guildName string) {
	for i := range s.guildsCache {
		if s.guildsCache[i].ID != guildID {
			continue
		}
		if guildName != "" {
			s.guildsCache[i].Name = guildName
		}
		return
	}
	s.guildsCache = append(s.guildsCache, models.GuildInfo{ID: guildID, Name: guildName})
}

var ErrNotReady = errors.New("discord session not ready")
