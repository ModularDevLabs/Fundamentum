package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
	"github.com/bwmarrin/discordgo"
)

var (
	web3EVMContractRe = regexp.MustCompile(`(?i)\b0x[a-f0-9]{40}\b`)
	web3SolAddressRe  = regexp.MustCompile(`\b[1-9A-HJ-NP-Za-km-z]{32,44}\b`)
	web3CashTagRe     = regexp.MustCompile(`(?i)(?:^|\s)\$([a-z][a-z0-9._-]{1,31})\b`)
	web3CommandRe     = regexp.MustCompile(`(?i)^(?:!|/)(?:scan|token|ca)\s+(\S+)`)
	web3HTTPClient    = &http.Client{Timeout: 4 * time.Second}
)

type web3Signal struct {
	Contract string
	CashTag  string
	Chain    string
	Dex      string
	Pair     string
	Exact    bool
}

type dexScreenerTokenResponse struct {
	Pairs []dexPair `json:"pairs"`
}

type dexScreenerSearchResponse struct {
	Pairs []dexPair `json:"pairs"`
}

type dexPair struct {
	ChainID   string       `json:"chainId"`
	DexID     string       `json:"dexId"`
	URL       string       `json:"url"`
	PairAddr  string       `json:"pairAddress"`
	PriceUSD  string       `json:"priceUsd"`
	FDV       float64      `json:"fdv"`
	MarketCap float64      `json:"marketCap"`
	BaseToken dexTokenMeta `json:"baseToken"`
	Liquidity struct {
		USD float64 `json:"usd"`
	} `json:"liquidity"`
	Volume struct {
		H24 float64 `json:"h24"`
	} `json:"volume"`
	PriceChange struct {
		H24 float64 `json:"h24"`
	} `json:"priceChange"`
	Info struct {
		Websites []struct {
			Label string `json:"label"`
			URL   string `json:"url"`
		} `json:"websites"`
		Socials []struct {
			Type string `json:"type"`
			URL  string `json:"url"`
		} `json:"socials"`
	} `json:"info"`
}

type dexTokenMeta struct {
	Address string `json:"address"`
	Symbol  string `json:"symbol"`
	Name    string `json:"name"`
}

type cgSearchResponse struct {
	Coins []cgSearchCoin `json:"coins"`
}

type cgSearchCoin struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	MarketCapRank int    `json:"market_cap_rank"`
}

type cgMarket struct {
	ID                       string   `json:"id"`
	Symbol                   string   `json:"symbol"`
	Name                     string   `json:"name"`
	CurrentPrice             float64  `json:"current_price"`
	MarketCap                float64  `json:"market_cap"`
	FDV                      *float64 `json:"fully_diluted_valuation"`
	PriceChangePercentage24H *float64 `json:"price_change_percentage_24h"`
}

type quickLink struct {
	Label string
	URL   string
}

type web3ModuleConfig struct {
	WhaleAlertsEnabled    bool
	WhaleMinTradeUSD      float64
	PriceAlertsEnabled    bool
	PriceAlertPumpPct     float64
	PriceAlertDumpPct     float64
	HealthChecksEnabled   bool
	HealthMinLiquidityUSD float64
	MiniTAEnabled         bool
	TrendSignalsEnabled   bool
	RugRiskEnabled        bool
	HolderViewEnabled     bool
	WalletWatchEnabled    bool
	WalletWatchlist       map[string]struct{}
	ConfidenceEnabled     bool
	CommandsEnabled       bool
	AntiSpamEnabled       bool
	PerTokenCooldownSec   time.Duration
}

func (s *Service) handleWeb3IntelMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if !settings.FeatureEnabled(models.FeatureWeb3Intel) {
		return
	}
	cfg := buildWeb3ModuleConfig(settings)
	signal := detectWeb3Signal(m.Content, cfg.CommandsEnabled)
	if signal.Contract == "" && signal.CashTag == "" {
		return
	}
	assetKey := ""
	if signal.Contract != "" {
		assetKey = "contract:" + normalizeContractKey(signal.Contract)
	} else {
		assetKey = "cashtag:" + strings.ToLower(strings.TrimSpace(signal.CashTag))
	}
	if !s.allowWeb3Lookup(m.ChannelID, assetKey, cfg) {
		return
	}

	lookupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var embed *discordgo.MessageEmbed
	var err error
	switch {
	case signal.Contract != "":
		embed, err = s.resolveContractIntelEmbed(lookupCtx, m.GuildID, m.Author.ID, m.Author.Username, signal, cfg)
	case signal.CashTag != "":
		embed, err = s.resolveCashTagEmbed(lookupCtx, m.GuildID, m.Author.ID, m.Author.Username, signal, cfg)
	}
	if err != nil {
		s.logger.Debug("web3 intel lookup failed guild=%s channel=%s err=%v", m.GuildID, m.ChannelID, err)
		return
	}
	if embed == nil {
		return
	}
	_, err = s.session.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Embed: embed,
	})
	if err != nil {
		s.logger.Error("web3 intel reply failed guild=%s channel=%s err=%v", m.GuildID, m.ChannelID, err)
	}
}

func buildWeb3ModuleConfig(settings models.GuildSettings) web3ModuleConfig {
	watch := map[string]struct{}{}
	for _, raw := range settings.Web3WalletWatchlist {
		val := normalizeContractKey(raw)
		if val == "" {
			continue
		}
		watch[val] = struct{}{}
	}
	cooldownSec := settings.Web3PerTokenCooldownSec
	if cooldownSec <= 0 {
		cooldownSec = 30
	}
	return web3ModuleConfig{
		WhaleAlertsEnabled:    settings.Web3WhaleAlertsEnabled,
		WhaleMinTradeUSD:      float64(settings.Web3WhaleMinTradeUSD),
		PriceAlertsEnabled:    settings.Web3PriceAlertsEnabled,
		PriceAlertPumpPct:     float64(settings.Web3PriceAlertPumpPct),
		PriceAlertDumpPct:     float64(settings.Web3PriceAlertDumpPct),
		HealthChecksEnabled:   settings.Web3HealthChecksEnabled,
		HealthMinLiquidityUSD: float64(settings.Web3HealthMinLiquidityUSD),
		MiniTAEnabled:         settings.Web3MiniTAEnabled,
		TrendSignalsEnabled:   settings.Web3TrendSignalsEnabled,
		RugRiskEnabled:        settings.Web3RugRiskEnabled,
		HolderViewEnabled:     settings.Web3HolderViewEnabled,
		WalletWatchEnabled:    settings.Web3WalletWatchEnabled,
		WalletWatchlist:       watch,
		ConfidenceEnabled:     settings.Web3ConfidenceScoreEnabled,
		CommandsEnabled:       settings.Web3CommandsEnabled,
		AntiSpamEnabled:       settings.Web3AntiSpamEnabled,
		PerTokenCooldownSec:   time.Duration(cooldownSec) * time.Second,
	}
}

func (s *Service) allowWeb3Lookup(channelID, assetKey string, cfg web3ModuleConfig) bool {
	if !cfg.AntiSpamEnabled {
		return true
	}
	if channelID == "" {
		return false
	}
	s.web3Mu.Lock()
	defer s.web3Mu.Unlock()
	now := time.Now().UTC()
	perChannelCooldown := cfg.PerTokenCooldownSec / 2
	if perChannelCooldown < 4*time.Second {
		perChannelCooldown = 4 * time.Second
	}
	last := s.web3Last[channelID]
	if !last.IsZero() && now.Sub(last) < perChannelCooldown {
		return false
	}
	if assetKey != "" {
		lastAsset := s.web3LastAsset[assetKey]
		if !lastAsset.IsZero() && now.Sub(lastAsset) < cfg.PerTokenCooldownSec {
			return false
		}
		s.web3LastAsset[assetKey] = now
	}
	s.web3Last[channelID] = now
	return true
}

func detectWeb3Signal(content string, commandsEnabled bool) web3Signal {
	text := strings.TrimSpace(content)
	if text == "" {
		return web3Signal{}
	}
	if commandsEnabled {
		if sig, ok := parseWeb3CommandSignal(text); ok {
			return sig
		}
	}
	evmLoc := web3EVMContractRe.FindStringIndex(text)
	solLoc := web3SolAddressRe.FindStringIndex(text)
	if evmLoc != nil && (solLoc == nil || evmLoc[0] <= solLoc[0]) {
		return web3Signal{Contract: strings.ToLower(text[evmLoc[0]:evmLoc[1]])}
	}
	if solLoc != nil {
		return web3Signal{Contract: text[solLoc[0]:solLoc[1]]}
	}
	match := web3CashTagRe.FindStringSubmatch(text)
	if len(match) >= 2 {
		return web3Signal{CashTag: strings.ToLower(match[1])}
	}
	return web3Signal{}
}

func parseWeb3CommandSignal(text string) (web3Signal, bool) {
	if !web3CommandRe.MatchString(text) {
		return web3Signal{}, false
	}
	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) < 2 {
		return web3Signal{}, true
	}
	cmd := strings.TrimLeft(strings.ToLower(parts[0]), "!/")
	args := parts[1:]
	sig := web3Signal{}

	if cmd == "scan" && len(args) >= 2 && !strings.HasPrefix(args[0], "-") {
		if ch := normalizeWeb3Chain(args[0]); ch != "" {
			sig.Chain = ch
			args = args[1:]
		}
	}

	target := ""
	for i := 0; i < len(args); i++ {
		a := strings.TrimSpace(args[i])
		if a == "" {
			continue
		}
		switch strings.ToLower(a) {
		case "--chain":
			if i+1 < len(args) {
				sig.Chain = normalizeWeb3Chain(args[i+1])
				i++
			}
			continue
		case "--dex":
			if i+1 < len(args) {
				sig.Dex = strings.ToLower(strings.TrimSpace(args[i+1]))
				i++
			}
			continue
		case "--pair":
			if i+1 < len(args) {
				sig.Pair = strings.TrimSpace(args[i+1])
				i++
			}
			continue
		case "--exact":
			sig.Exact = true
			continue
		}
		if strings.HasPrefix(a, "--chain=") {
			sig.Chain = normalizeWeb3Chain(strings.TrimPrefix(a, "--chain="))
			continue
		}
		if strings.HasPrefix(a, "--dex=") {
			sig.Dex = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(a, "--dex=")))
			continue
		}
		if strings.HasPrefix(a, "--pair=") {
			sig.Pair = strings.TrimSpace(strings.TrimPrefix(a, "--pair="))
			continue
		}
		if a == "--exact=true" {
			sig.Exact = true
			continue
		}
		if target == "" {
			target = strings.Trim(a, "\"'")
		}
	}
	if target == "" {
		return web3Signal{}, true
	}
	if strings.HasPrefix(target, "$") {
		sig.CashTag = strings.ToLower(strings.TrimPrefix(target, "$"))
		return sig, true
	}
	if cmd == "ca" && target != "" {
		sig.Contract = normalizeContractKey(target)
		return sig, true
	}
	if web3EVMContractRe.MatchString(target) || web3SolAddressRe.MatchString(target) {
		sig.Contract = normalizeContractKey(target)
		return sig, true
	}
	sig.CashTag = strings.ToLower(target)
	return sig, true
}

func normalizeWeb3Chain(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "eth", "ethereum":
		return "ethereum"
	case "arb", "arbitrum":
		return "arbitrum"
	case "op", "optimism":
		return "optimism"
	case "base":
		return "base"
	case "poly", "polygon":
		return "polygon"
	case "bnb", "bsc":
		return "bsc"
	case "sol", "solana":
		return "solana"
	case "hl", "hyperliquid":
		return "hyperliquid"
	case "monad":
		return "monad"
	default:
		return ""
	}
}

func (s *Service) resolveContractIntelEmbed(ctx context.Context, guildID, scannerUserID, scannerName string, sig web3Signal, cfg web3ModuleConfig) (*discordgo.MessageEmbed, error) {
	var dex dexScreenerTokenResponse
	contract := sig.Contract
	if err := web3FetchJSON(ctx, "https://api.dexscreener.com/latest/dex/tokens/"+url.PathEscape(contract), &dex); err != nil {
		return nil, err
	}
	best := chooseBestDexPair(dex.Pairs, sig)
	if best == nil {
		return nil, nil
	}
	tokenAddr := best.BaseToken.Address
	if tokenAddr == "" {
		tokenAddr = contract
	}

	name := fallback(best.BaseToken.Name, "Token")
	symbol := strings.ToUpper(fallback(best.BaseToken.Symbol, "?"))
	chain := formatChain(best.ChainID)

	primaryLinks := make([]quickLink, 0, 5)
	if best.URL != "" {
		primaryLinks = append(primaryLinks, quickLink{Label: "Chart", URL: best.URL})
	}
	primaryLinks = append(primaryLinks, quickLink{Label: "Defined", URL: "https://www.defined.fi/search?query=" + url.QueryEscape(tokenAddr)})
	if explorer := explorerTokenURL(best.ChainID, tokenAddr); explorer != "" {
		primaryLinks = append(primaryLinks, quickLink{Label: "Explorer", URL: explorer})
	}
	primaryLinks = append(primaryLinks, quickLink{Label: "X Search", URL: "https://x.com/search?q=" + url.QueryEscape(tokenAddr)})
	primaryLinks = append(primaryLinks, quickLink{Label: "CG Search", URL: "https://www.coingecko.com/en/search?query=" + url.QueryEscape(tokenAddr)})

	socialLinks := socialLinksFromDex(best)
	tradeLinks := tradeLinksForChain(best.ChainID, tokenAddr, best.PairAddr)

	mcap := best.MarketCap
	if mcap <= 0 {
		mcap = best.FDV
	}
	currentPrice := parseDecimal(best.PriceUSD)
	firstScan, created, err := s.repos.Web3Scans.GetOrCreateFirstScan(ctx, models.Web3FirstScanRow{
		GuildID:            guildID,
		AssetKey:           "contract:" + normalizeContractKey(tokenAddr),
		AssetType:          "contract",
		DisplaySymbol:      symbol,
		DisplayName:        name,
		FirstScannerUserID: scannerUserID,
		FirstScannerName:   scannerName,
		FirstPriceUSD:      currentPrice,
	})
	if err != nil {
		return nil, err
	}
	stats := marketSnapshotRow(
		formatUSD(currentPrice),
		formatPercent(best.PriceChange.H24),
		formatUSDCompact(mcap),
		formatUSDCompact(best.FDV),
		formatUSDCompact(best.Liquidity.USD),
		formatUSDCompact(best.Volume.H24),
	)
	desc := fmt.Sprintf("%s (%s) on **%s** via **%s**", name, symbol, chain, fallback(strings.ToUpper(best.DexID), "DEX"))
	if len(primaryLinks) > 0 {
		desc += "\n" + formatQuickLinks(primaryLinks)
	}
	if len(socialLinks) > 0 {
		desc += "\nSocial: " + formatQuickLinks(socialLinks)
	}
	desc += "\nCA: `" + tokenAddr + "`"

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Market Snapshot",
			Value:  stats,
			Inline: false,
		},
	}
	if len(tradeLinks) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Quick Trade",
			Value:  formatQuickLinks(tradeLinks),
			Inline: false,
		})
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "First Scan",
		Value:  buildFirstScanField(firstScan, currentPrice, created),
		Inline: false,
	})
	fields = append(fields, buildWeb3SignalFields(cfg, best, tokenAddr, currentPrice, best.PriceChange.H24, mcap, firstScan, created)...)

	return &discordgo.MessageEmbed{
		Title:       "Web3 Intel",
		Description: trimEmbedText(desc, 1024),
		Fields:      fields,
		Color:       0x2ecc71,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Sources: Dexscreener / Defined / CoinGecko",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) resolveCashTagEmbed(ctx context.Context, guildID, scannerUserID, scannerName string, sig web3Signal, cfg web3ModuleConfig) (*discordgo.MessageEmbed, error) {
	token := sig.CashTag
	if sig.Chain != "" || sig.Dex != "" || sig.Pair != "" || sig.Exact {
		return s.resolveCashTagDexFallback(ctx, guildID, scannerUserID, scannerName, sig, cfg)
	}
	var search cgSearchResponse
	if err := web3FetchJSON(ctx, "https://api.coingecko.com/api/v3/search?query="+url.QueryEscape(token), &search); err != nil {
		return nil, err
	}
	coin := chooseCoinGeckoCoin(search.Coins, token)
	if coin == nil {
		return s.resolveCashTagDexFallback(ctx, guildID, scannerUserID, scannerName, sig, cfg)
	}
	var markets []cgMarket
	u := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=" + url.QueryEscape(coin.ID)
	if err := web3FetchJSON(ctx, u, &markets); err != nil {
		return nil, err
	}
	if len(markets) == 0 {
		return s.resolveCashTagDexFallback(ctx, guildID, scannerUserID, scannerName, sig, cfg)
	}
	m := markets[0]
	fdv := 0.0
	if m.FDV != nil {
		fdv = *m.FDV
	}
	change := 0.0
	if m.PriceChangePercentage24H != nil {
		change = *m.PriceChangePercentage24H
	}
	firstScan, created, err := s.repos.Web3Scans.GetOrCreateFirstScan(ctx, models.Web3FirstScanRow{
		GuildID:            guildID,
		AssetKey:           "coingecko:" + strings.TrimSpace(coin.ID),
		AssetType:          "coingecko",
		DisplaySymbol:      strings.ToUpper(fallback(m.Symbol, coin.Symbol)),
		DisplayName:        fallback(m.Name, coin.Name),
		FirstScannerUserID: scannerUserID,
		FirstScannerName:   scannerName,
		FirstPriceUSD:      m.CurrentPrice,
	})
	if err != nil {
		return nil, err
	}
	description := fmt.Sprintf(
		"%s (%s)\n%s",
		fallback(m.Name, coin.Name),
		strings.ToUpper(fallback(m.Symbol, coin.Symbol)),
		formatQuickLinks([]quickLink{
			{Label: "CoinGecko", URL: "https://www.coingecko.com/en/coins/" + coin.ID},
			{Label: "Search DexScreener", URL: "https://dexscreener.com/?q=" + url.QueryEscape(strings.ToUpper(token))},
		}),
	)
	stats := marketSnapshotRow(
		formatUSD(m.CurrentPrice),
		formatPercent(change),
		formatUSDCompact(m.MarketCap),
		formatUSDCompact(fdv),
		"n/a",
		"n/a",
	)
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Market Snapshot",
			Inline: false,
			Value:  stats,
		},
		{
			Name:   "Quick Trade",
			Inline: false,
			Value: formatQuickLinks([]quickLink{
				{Label: "Search DexScreener", URL: "https://dexscreener.com/?q=" + url.QueryEscape(strings.ToUpper(token))},
				{Label: "CoinGecko", URL: "https://www.coingecko.com/en/coins/" + coin.ID},
			}),
		},
		{
			Name:   "First Scan",
			Inline: false,
			Value:  buildFirstScanField(firstScan, m.CurrentPrice, created),
		},
	}
	fields = append(fields, buildCoinGeckoSignalFields(cfg, m.CurrentPrice, change, m.MarketCap, fdv, firstScan, created)...)

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Web3 Intel • $%s", strings.ToUpper(token)),
		Description: trimEmbedText(description, 1024),
		Color:       0x3498db,
		Fields:      fields,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) resolveCashTagDexFallback(ctx context.Context, guildID, scannerUserID, scannerName string, sig web3Signal, cfg web3ModuleConfig) (*discordgo.MessageEmbed, error) {
	token := sig.CashTag
	var search dexScreenerSearchResponse
	u := "https://api.dexscreener.com/latest/dex/search/?q=" + url.QueryEscape(strings.ToUpper(strings.TrimSpace(token)))
	if err := web3FetchJSON(ctx, u, &search); err != nil {
		return nil, err
	}
	best := chooseBestDexPairForTicker(search.Pairs, token, sig)
	if best == nil {
		return nil, nil
	}
	tokenAddr := best.BaseToken.Address
	if tokenAddr == "" {
		tokenAddr = best.PairAddr
	}
	name := fallback(best.BaseToken.Name, strings.ToUpper(token))
	symbol := strings.ToUpper(fallback(best.BaseToken.Symbol, token))
	chain := formatChain(best.ChainID)
	currentPrice := parseDecimal(best.PriceUSD)

	firstScan, created, err := s.repos.Web3Scans.GetOrCreateFirstScan(ctx, models.Web3FirstScanRow{
		GuildID:            guildID,
		AssetKey:           "contract:" + normalizeContractKey(tokenAddr),
		AssetType:          "contract",
		DisplaySymbol:      symbol,
		DisplayName:        name,
		FirstScannerUserID: scannerUserID,
		FirstScannerName:   scannerName,
		FirstPriceUSD:      currentPrice,
	})
	if err != nil {
		return nil, err
	}

	primaryLinks := make([]quickLink, 0, 5)
	if best.URL != "" {
		primaryLinks = append(primaryLinks, quickLink{Label: "Chart", URL: best.URL})
	}
	primaryLinks = append(primaryLinks, quickLink{Label: "Defined", URL: "https://www.defined.fi/search?query=" + url.QueryEscape(tokenAddr)})
	if explorer := explorerTokenURL(best.ChainID, tokenAddr); explorer != "" {
		primaryLinks = append(primaryLinks, quickLink{Label: "Explorer", URL: explorer})
	}
	primaryLinks = append(primaryLinks, quickLink{Label: "X Search", URL: "https://x.com/search?q=" + url.QueryEscape(tokenAddr)})
	primaryLinks = append(primaryLinks, quickLink{Label: "CG Search", URL: "https://www.coingecko.com/en/search?query=" + url.QueryEscape(tokenAddr)})
	socialLinks := socialLinksFromDex(best)
	tradeLinks := tradeLinksForChain(best.ChainID, tokenAddr, best.PairAddr)

	description := fmt.Sprintf("%s (%s) on **%s** via **%s**", name, symbol, chain, fallback(strings.ToUpper(best.DexID), "DEX"))
	if len(primaryLinks) > 0 {
		description += "\n" + formatQuickLinks(primaryLinks)
	}
	if len(socialLinks) > 0 {
		description += "\nSocial: " + formatQuickLinks(socialLinks)
	}

	stats := marketSnapshotRow(
		formatUSD(currentPrice),
		formatPercent(best.PriceChange.H24),
		formatUSDCompact(maxFloat(best.MarketCap, best.FDV)),
		formatUSDCompact(best.FDV),
		formatUSDCompact(best.Liquidity.USD),
		formatUSDCompact(best.Volume.H24),
	)
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Market Snapshot",
			Inline: false,
			Value:  stats,
		},
	}
	if len(tradeLinks) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Quick Trade",
			Value:  formatQuickLinks(tradeLinks),
			Inline: false,
		})
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "First Scan",
		Inline: false,
		Value:  buildFirstScanField(firstScan, currentPrice, created),
	})
	fields = append(fields, buildWeb3SignalFields(cfg, best, tokenAddr, currentPrice, best.PriceChange.H24, maxFloat(best.MarketCap, best.FDV), firstScan, created)...)
	description += "\nCA: `" + tokenAddr + "`"

	return &discordgo.MessageEmbed{
		Title:       "Web3 Intel",
		Description: trimEmbedText(description, 1024),
		Color:       0x2ecc71,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Sources: Dexscreener / Defined / CoinGecko",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func buildWeb3SignalFields(cfg web3ModuleConfig, pair *dexPair, tokenAddr string, currentPrice, change24h, mcap float64, first models.Web3FirstScanRow, created bool) []*discordgo.MessageEmbedField {
	if pair == nil {
		return nil
	}
	fields := make([]*discordgo.MessageEmbedField, 0, 8)

	if cfg.WhaleAlertsEnabled && pair.Volume.H24 >= cfg.WhaleMinTradeUSD {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Whale Flow",
			Inline: false,
			Value:  fmt.Sprintf("24h volume %s crossed whale threshold %s.", formatUSDCompact(pair.Volume.H24), formatUSDCompact(cfg.WhaleMinTradeUSD)),
		})
	}
	if cfg.PriceAlertsEnabled {
		if change24h >= cfg.PriceAlertPumpPct {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Price Alert",
				Inline: false,
				Value:  fmt.Sprintf("Pump signal: %s in 24h (threshold +%.0f%%).", formatPercent(change24h), cfg.PriceAlertPumpPct),
			})
		} else if change24h <= -cfg.PriceAlertDumpPct {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Price Alert",
				Inline: false,
				Value:  fmt.Sprintf("Dump signal: %s in 24h (threshold -%.0f%%).", formatPercent(change24h), cfg.PriceAlertDumpPct),
			})
		}
	}
	if cfg.HealthChecksEnabled {
		liqStatus := "healthy"
		if pair.Liquidity.USD < cfg.HealthMinLiquidityUSD {
			liqStatus = "thin-liquidity"
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Health Check",
			Inline: false,
			Value:  fmt.Sprintf("%s • Liquidity %s (min %s)", liqStatus, formatUSDCompact(pair.Liquidity.USD), formatUSDCompact(cfg.HealthMinLiquidityUSD)),
		})
	}
	if cfg.MiniTAEnabled {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Mini TA",
			Inline: false,
			Value:  miniTASummary(change24h, pair.Volume.H24, pair.Liquidity.USD),
		})
	}
	if cfg.TrendSignalsEnabled {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Trend Signal",
			Inline: false,
			Value:  trendSummary(change24h, pair.Volume.H24, pair.Liquidity.USD),
		})
	}
	if cfg.RugRiskEnabled {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Rug Risk",
			Inline: false,
			Value:  rugRiskSummary(pair.Liquidity.USD, mcap, pair.Volume.H24),
		})
	}
	if cfg.HolderViewEnabled {
		if holders := explorerHoldersURL(pair.ChainID, tokenAddr); holders != "" {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Holder View",
				Inline: false,
				Value:  fmt.Sprintf("[Open holders](%s)", holders),
			})
		}
	}
	if cfg.WalletWatchEnabled && tokenAddr != "" {
		if _, ok := cfg.WalletWatchlist[normalizeContractKey(tokenAddr)]; ok {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Watchlist Match",
				Inline: false,
				Value:  "Token contract matched this guild wallet/token watchlist.",
			})
		}
	}
	if cfg.ConfidenceEnabled {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Confidence",
			Inline: false,
			Value:  confidenceSummary(pair.Liquidity.USD, pair.Volume.H24, mcap, currentPrice, first, created),
		})
	}
	return fields
}

func buildCoinGeckoSignalFields(cfg web3ModuleConfig, currentPrice, change24h, mcap, fdv float64, first models.Web3FirstScanRow, created bool) []*discordgo.MessageEmbedField {
	fields := make([]*discordgo.MessageEmbedField, 0, 3)
	if cfg.PriceAlertsEnabled {
		if change24h >= cfg.PriceAlertPumpPct {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Price Alert",
				Inline: false,
				Value:  fmt.Sprintf("Pump signal: %s in 24h (threshold +%.0f%%).", formatPercent(change24h), cfg.PriceAlertPumpPct),
			})
		} else if change24h <= -cfg.PriceAlertDumpPct {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Price Alert",
				Inline: false,
				Value:  fmt.Sprintf("Dump signal: %s in 24h (threshold -%.0f%%).", formatPercent(change24h), cfg.PriceAlertDumpPct),
			})
		}
	}
	if cfg.TrendSignalsEnabled {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Trend Signal",
			Inline: false,
			Value:  trendSummary(change24h, 0, 0),
		})
	}
	if cfg.ConfidenceEnabled {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Confidence",
			Inline: false,
			Value:  confidenceSummary(0, 0, maxFloat(mcap, fdv), currentPrice, first, created),
		})
	}
	return fields
}

func miniTASummary(change24h, vol24h, liq float64) string {
	momentum := "sideways"
	switch {
	case change24h >= 12:
		momentum = "strong bullish momentum"
	case change24h >= 3:
		momentum = "bullish momentum"
	case change24h <= -12:
		momentum = "strong bearish momentum"
	case change24h <= -3:
		momentum = "bearish momentum"
	}
	participation := "normal participation"
	if liq > 0 && vol24h/liq > 2.0 {
		participation = "high turnover"
	} else if vol24h > 0 && liq > 0 && vol24h/liq < 0.2 {
		participation = "low turnover"
	}
	return fmt.Sprintf("%s • %s", momentum, participation)
}

func trendSummary(change24h, vol24h, liq float64) string {
	trend := "neutral"
	switch {
	case change24h >= 8:
		trend = "uptrend"
	case change24h <= -8:
		trend = "downtrend"
	}
	conviction := "moderate conviction"
	if liq > 0 && vol24h/liq >= 1.2 {
		conviction = "high conviction"
	} else if liq > 0 && vol24h/liq < 0.25 {
		conviction = "low conviction"
	}
	return fmt.Sprintf("%s • %s", trend, conviction)
}

func rugRiskSummary(liquidity, mcap, volume float64) string {
	if liquidity <= 0 {
		return "high risk • no liquidity data"
	}
	score := 0
	if liquidity < 10000 {
		score += 2
	} else if liquidity < 25000 {
		score++
	}
	if mcap > 0 {
		ratio := liquidity / mcap
		if ratio < 0.01 {
			score += 2
		} else if ratio < 0.03 {
			score++
		}
	}
	if volume > 0 && liquidity > 0 && volume/liquidity > 6 {
		score++
	}
	switch {
	case score >= 4:
		return "high risk • thin liquidity relative to size/flow"
	case score >= 2:
		return "medium risk • monitor liquidity and flow consistency"
	default:
		return "lower risk • liquidity profile looks healthier"
	}
}

func confidenceSummary(liquidity, volume, mcap, price float64, first models.Web3FirstScanRow, created bool) string {
	score := 50
	if liquidity >= 100000 {
		score += 20
	} else if liquidity >= 25000 {
		score += 10
	}
	if mcap >= 5_000_000 {
		score += 10
	}
	if volume > 0 && liquidity > 0 {
		ratio := volume / liquidity
		if ratio >= 0.2 && ratio <= 4 {
			score += 10
		}
	}
	if first.FirstPriceUSD > 0 && price > 0 && !created {
		move := math.Abs((price - first.FirstPriceUSD) / first.FirstPriceUSD * 100)
		if move > 75 {
			score -= 10
		}
	}
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	label := "moderate"
	if score >= 75 {
		label = "high"
	} else if score <= 35 {
		label = "low"
	}
	return fmt.Sprintf("%d/100 (%s confidence)", score, label)
}

func chooseBestDexPair(pairs []dexPair, sig web3Signal) *dexPair {
	if len(pairs) == 0 {
		return nil
	}
	candidates := make([]dexPair, 0, len(pairs))
	for _, p := range pairs {
		if strings.TrimSpace(p.PriceUSD) == "" {
			continue
		}
		if !pairMatchesSignalOptions(p, sig) {
			continue
		}
		candidates = append(candidates, p)
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Liquidity.USD == candidates[j].Liquidity.USD {
			return candidates[i].Volume.H24 > candidates[j].Volume.H24
		}
		return candidates[i].Liquidity.USD > candidates[j].Liquidity.USD
	})
	return &candidates[0]
}

func chooseCoinGeckoCoin(coins []cgSearchCoin, query string) *cgSearchCoin {
	if len(coins) == 0 {
		return nil
	}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	bestIdx := -1
	bestScore := -1
	bestRank := math.MaxInt32
	for i := range coins {
		c := coins[i]
		score := 0
		if strings.EqualFold(c.Symbol, q) {
			score = 3
		} else if strings.EqualFold(c.Name, q) {
			score = 2
		} else if strings.Contains(strings.ToLower(c.Name), q) || strings.Contains(strings.ToLower(c.Symbol), q) {
			score = 1
		}
		if score == 0 {
			continue
		}
		rank := c.MarketCapRank
		if rank <= 0 {
			rank = math.MaxInt32
		}
		if score > bestScore || (score == bestScore && rank < bestRank) {
			bestIdx = i
			bestScore = score
			bestRank = rank
		}
	}
	if bestIdx >= 0 {
		return &coins[bestIdx]
	}
	return &coins[0]
}

func web3FetchJSON(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "FundamentumBot/1.0 (+https://github.com/ModularDevLabs/Fundamentum)")

	res, err := web3HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("http %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}

func socialLinksFromDex(pair *dexPair) []quickLink {
	if pair == nil {
		return nil
	}
	out := make([]quickLink, 0, 2)
	for _, w := range pair.Info.Websites {
		if strings.TrimSpace(w.URL) == "" {
			continue
		}
		out = append(out, quickLink{Label: "Website", URL: w.URL})
		break
	}
	for _, s := range pair.Info.Socials {
		if strings.TrimSpace(s.URL) == "" {
			continue
		}
		t := strings.ToLower(strings.TrimSpace(s.Type))
		if strings.Contains(t, "twitter") || strings.Contains(t, "x") {
			out = append(out, quickLink{Label: "X", URL: s.URL})
			break
		}
	}
	return out
}

func tradeLinksForChain(chainID, tokenAddr, pairAddr string) []quickLink {
	chain := strings.ToLower(strings.TrimSpace(chainID))
	addr := strings.TrimSpace(tokenAddr)
	pair := strings.TrimSpace(pairAddr)
	if addr == "" {
		return nil
	}
	out := make([]quickLink, 0, 4)
	switch chain {
	case "solana":
		out = append(out,
			quickLink{Label: "Jupiter", URL: "https://jup.ag/swap/SOL-" + url.QueryEscape(addr)},
			quickLink{Label: "Birdeye", URL: "https://birdeye.so/token/" + url.PathEscape(addr) + "?chain=solana"},
		)
	case "bsc":
		out = append(out,
			quickLink{Label: "Pancake", URL: "https://pancakeswap.finance/swap?outputCurrency=" + url.QueryEscape(addr)},
		)
	default:
		out = append(out,
			quickLink{Label: "Uniswap", URL: "https://app.uniswap.org/swap?chain=" + url.QueryEscape(chainForUniswap(chain)) + "&outputCurrency=" + url.QueryEscape(addr)},
		)
	}
	if pair != "" && chain != "" {
		out = append(out, quickLink{Label: "DEXTools", URL: "https://www.dextools.io/app/en/" + url.PathEscape(chain) + "/pair-explorer/" + url.PathEscape(pair)})
	}
	return out
}

func chainForUniswap(chain string) string {
	switch chain {
	case "ethereum", "base", "arbitrum", "optimism", "polygon":
		return chain
	default:
		return "ethereum"
	}
}

func explorerTokenURL(chainID, tokenAddr string) string {
	chain := strings.ToLower(strings.TrimSpace(chainID))
	addr := strings.TrimSpace(tokenAddr)
	if addr == "" {
		return ""
	}
	switch chain {
	case "ethereum":
		return "https://etherscan.io/token/" + url.PathEscape(addr)
	case "arbitrum":
		return "https://arbiscan.io/token/" + url.PathEscape(addr)
	case "optimism":
		return "https://optimistic.etherscan.io/token/" + url.PathEscape(addr)
	case "base":
		return "https://basescan.org/token/" + url.PathEscape(addr)
	case "polygon":
		return "https://polygonscan.com/token/" + url.PathEscape(addr)
	case "linea":
		return "https://lineascan.build/token/" + url.PathEscape(addr)
	case "zksync":
		return "https://era.zksync.network/address/" + url.PathEscape(addr)
	case "bsc":
		return "https://bscscan.com/token/" + url.PathEscape(addr)
	case "solana":
		return "https://solscan.io/token/" + url.PathEscape(addr)
	default:
		return "https://dexscreener.com/search?q=" + url.QueryEscape(addr)
	}
}

func explorerHoldersURL(chainID, tokenAddr string) string {
	chain := strings.ToLower(strings.TrimSpace(chainID))
	addr := strings.TrimSpace(tokenAddr)
	if addr == "" {
		return ""
	}
	switch chain {
	case "ethereum":
		return "https://etherscan.io/token/" + url.PathEscape(addr) + "#balances"
	case "arbitrum":
		return "https://arbiscan.io/token/" + url.PathEscape(addr) + "#balances"
	case "optimism":
		return "https://optimistic.etherscan.io/token/" + url.PathEscape(addr) + "#balances"
	case "base":
		return "https://basescan.org/token/" + url.PathEscape(addr) + "#balances"
	case "polygon":
		return "https://polygonscan.com/token/" + url.PathEscape(addr) + "#balances"
	case "bsc":
		return "https://bscscan.com/token/" + url.PathEscape(addr) + "#balances"
	case "solana":
		return "https://solscan.io/token/" + url.PathEscape(addr) + "#holders"
	default:
		return ""
	}
}

func parseDecimal(raw string) float64 {
	n, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0
	}
	return n
}

func formatUSD(v float64) string {
	if v <= 0 {
		return "n/a"
	}
	if v >= 1 {
		return fmt.Sprintf("$%.4f", v)
	}
	return fmt.Sprintf("$%.8f", v)
}

func formatUSDCompact(v float64) string {
	if v <= 0 {
		return "n/a"
	}
	switch {
	case v >= 1_000_000_000:
		return fmt.Sprintf("$%.2fB", v/1_000_000_000)
	case v >= 1_000_000:
		return fmt.Sprintf("$%.2fM", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("$%.2fK", v/1_000)
	default:
		return fmt.Sprintf("$%.0f", v)
	}
}

func formatPercent(v float64) string {
	if math.Abs(v) < 0.0001 {
		return "0.00%"
	}
	return fmt.Sprintf("%+.2f%%", v)
}

func formatChain(chain string) string {
	switch strings.ToLower(strings.TrimSpace(chain)) {
	case "ethereum":
		return "Ethereum"
	case "arbitrum":
		return "Arbitrum"
	case "optimism":
		return "Optimism"
	case "base":
		return "Base"
	case "polygon":
		return "Polygon"
	case "linea":
		return "Linea"
	case "zksync":
		return "zkSync"
	case "bsc":
		return "BNB Chain"
	case "solana":
		return "Solana"
	case "hyperliquid":
		return "Hyperliquid"
	case "monad":
		return "Monad"
	default:
		return fallback(chain, "unknown")
	}
}

func formatQuickLinks(links []quickLink) string {
	out := make([]string, 0, len(links))
	for _, l := range links {
		if strings.TrimSpace(l.URL) == "" || strings.TrimSpace(l.Label) == "" {
			continue
		}
		out = append(out, fmt.Sprintf("[%s](%s)", l.Label, l.URL))
	}
	if len(out) == 0 {
		return "n/a"
	}
	return strings.Join(out, " • ")
}

func chooseBestDexPairForTicker(pairs []dexPair, ticker string, sig web3Signal) *dexPair {
	if len(pairs) == 0 {
		return nil
	}
	needle := strings.ToUpper(strings.TrimSpace(ticker))
	candidates := make([]dexPair, 0, len(pairs))
	for _, p := range pairs {
		if strings.TrimSpace(p.PriceUSD) == "" {
			continue
		}
		if !pairMatchesSignalOptions(p, sig) {
			continue
		}
		if needle != "" {
			sym := strings.ToUpper(strings.TrimSpace(p.BaseToken.Symbol))
			name := strings.ToUpper(strings.TrimSpace(p.BaseToken.Name))
			if sig.Exact {
				if sym != needle && name != needle {
					continue
				}
			} else if sym != needle && !strings.Contains(name, needle) {
				continue
			}
		}
		candidates = append(candidates, p)
	}
	if len(candidates) == 0 && !sig.Exact && sig.Chain == "" && sig.Dex == "" && sig.Pair == "" {
		candidates = append(candidates, pairs...)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Liquidity.USD == candidates[j].Liquidity.USD {
			return candidates[i].Volume.H24 > candidates[j].Volume.H24
		}
		return candidates[i].Liquidity.USD > candidates[j].Liquidity.USD
	})
	return &candidates[0]
}

func pairMatchesSignalOptions(p dexPair, sig web3Signal) bool {
	if sig.Pair != "" && !strings.EqualFold(strings.TrimSpace(p.PairAddr), strings.TrimSpace(sig.Pair)) {
		return false
	}
	if sig.Chain != "" && normalizeWeb3Chain(p.ChainID) != sig.Chain {
		return false
	}
	if sig.Dex != "" && strings.ToLower(strings.TrimSpace(p.DexID)) != sig.Dex {
		return false
	}
	return true
}

func maxFloat(a, b float64) float64 {
	if a >= b {
		return a
	}
	return b
}

func fallback(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}

func trimEmbedText(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func marketSnapshotRow(price, chg24, mcap, fdv, liq, vol string) string {
	return fmt.Sprintf("Price %s • 24h %s • MCap %s • FDV %s • Liq %s • Vol %s", price, chg24, mcap, fdv, liq, vol)
}

func normalizeContractKey(addr string) string {
	a := strings.TrimSpace(addr)
	if web3EVMContractRe.MatchString(a) {
		return strings.ToLower(a)
	}
	return a
}

func buildFirstScanField(first models.Web3FirstScanRow, currentPrice float64, created bool) string {
	whenUnix := first.FirstScannedAt.UTC().Unix()
	header := fmt.Sprintf("By <@%s> • <t:%d:R>", first.FirstScannerUserID, whenUnix)
	if created {
		return header + "\nInitial scan recorded now."
	}
	if first.FirstPriceUSD <= 0 || currentPrice <= 0 {
		return header + fmt.Sprintf("\nInitial price: %s", formatUSD(first.FirstPriceUSD))
	}
	since := ((currentPrice - first.FirstPriceUSD) / first.FirstPriceUSD) * 100
	return header + fmt.Sprintf("\nSince first: %s (from %s)", formatPercent(since), formatUSD(first.FirstPriceUSD))
}
