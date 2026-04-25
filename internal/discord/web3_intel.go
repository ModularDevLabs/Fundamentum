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
	web3TrendingRe    = regexp.MustCompile(`(?i)^(?:!|/)trending(?:\s+(.+))?$`)
	web3HTTPClient    = &http.Client{Timeout: 4 * time.Second}
)

type web3Signal struct {
	Contract     string
	CashTag      string
	Chain        string
	InvalidChain string
	Dex          string
	Pair         string
	Exact        bool
	Trending     bool
	Source       string
	Limit        int
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
	Txns struct {
		H24 struct {
			Buys  int `json:"buys"`
			Sells int `json:"sells"`
		} `json:"h24"`
	} `json:"txns"`
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

type cgTrendingResponse struct {
	Coins []struct {
		Item struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			Symbol        string `json:"symbol"`
			MarketCapRank int    `json:"market_cap_rank"`
			Data          struct {
				Price                 any `json:"price"`
				PriceChangePercent24h struct {
					USD any `json:"usd"`
				} `json:"price_change_percentage_24h"`
			} `json:"data"`
		} `json:"item"`
	} `json:"coins"`
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
	TotalVolume              float64  `json:"total_volume"`
	MarketCap                float64  `json:"market_cap"`
	FDV                      *float64 `json:"fully_diluted_valuation"`
	PriceChangePercentage24H *float64 `json:"price_change_percentage_24h"`
}

type cgTickersResponse struct {
	Tickers []cgTicker `json:"tickers"`
}

type cgTicker struct {
	TrustScore        string             `json:"trust_score"`
	ConvertedVolume   map[string]float64 `json:"converted_volume"`
	CostToMoveUpUSD   *float64           `json:"cost_to_move_up_usd"`
	CostToMoveDownUSD *float64           `json:"cost_to_move_down_usd"`
	IsStale           bool               `json:"is_stale"`
	IsAnomaly         bool               `json:"is_anomaly"`
}

type cgTickerStats struct {
	TrustedMarkets int
	Volume24hUSD   float64
	DepthUpUSD     float64
	DepthDownUSD   float64
}

type gtTradesResponse struct {
	Data []struct {
		Attributes struct {
			Kind           string `json:"kind"`
			VolumeInUSD    string `json:"volume_in_usd"`
			BlockTimestamp string `json:"block_timestamp"`
			TxHash         string `json:"tx_hash"`
		} `json:"attributes"`
	} `json:"data"`
}

type gtTrendingPoolsResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Name                  string         `json:"name"`
			Address               string         `json:"address"`
			BaseTokenPriceUSD     any            `json:"base_token_price_usd"`
			PriceChangePercentage map[string]any `json:"price_change_percentage"`
			ReserveInUSD          any            `json:"reserve_in_usd"`
			MarketCapUSD          any            `json:"market_cap_usd"`
			FDVUSD                any            `json:"fdv_usd"`
			VolumeUSD             map[string]any `json:"volume_usd"`
		} `json:"attributes"`
	} `json:"data"`
}

type quickLink struct {
	Label string
	URL   string
}

type web3ModuleConfig struct {
	WhaleAlertsEnabled  bool
	WhaleMinTradeUSD    float64
	PriceAlertsEnabled  bool
	PriceAlertPumpPct   float64
	PriceAlertDumpPct   float64
	WalletWatchEnabled  bool
	WalletWatchlist     map[string]struct{}
	CommandsEnabled     bool
	AntiSpamEnabled     bool
	PerTokenCooldownSec time.Duration
	TrendingCount       int
}

func (s *Service) handleWeb3IntelMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if !settings.FeatureEnabled(models.FeatureWeb3Intel) {
		return
	}
	cfg := buildWeb3ModuleConfig(settings)
	signal := detectWeb3Signal(m.Content, cfg.CommandsEnabled)
	if signal.Contract == "" && signal.CashTag == "" && !signal.Trending {
		return
	}
	assetKey := ""
	if signal.Trending {
		scope := signal.Source
		if signal.Chain != "" {
			scope += ":" + signal.Chain
		}
		assetKey = "trending:" + scope
	} else if signal.Contract != "" {
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
	case signal.Trending:
		embed, err = s.resolveTrendingEmbed(lookupCtx, signal, cfg)
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
	trendingCount := settings.Web3TrendingCount
	if trendingCount <= 0 {
		trendingCount = 20
	}
	if trendingCount > 50 {
		trendingCount = 50
	}
	return web3ModuleConfig{
		WhaleAlertsEnabled:  settings.Web3WhaleAlertsEnabled,
		WhaleMinTradeUSD:    float64(settings.Web3WhaleMinTradeUSD),
		PriceAlertsEnabled:  settings.Web3PriceAlertsEnabled,
		PriceAlertPumpPct:   float64(settings.Web3PriceAlertPumpPct),
		PriceAlertDumpPct:   float64(settings.Web3PriceAlertDumpPct),
		WalletWatchEnabled:  settings.Web3WalletWatchEnabled,
		WalletWatchlist:     watch,
		CommandsEnabled:     settings.Web3CommandsEnabled,
		AntiSpamEnabled:     settings.Web3AntiSpamEnabled,
		PerTokenCooldownSec: time.Duration(cooldownSec) * time.Second,
		TrendingCount:       trendingCount,
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
	if m := web3TrendingRe.FindStringSubmatch(strings.TrimSpace(text)); len(m) > 0 {
		rest := ""
		if len(m) > 1 {
			rest = m[1]
		}
		return parseTrendingSignal(rest), true
	}
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

func parseTrendingSignal(rest string) web3Signal {
	sig := web3Signal{Trending: true, Source: "onchain"}
	args := strings.Fields(strings.TrimSpace(rest))
	if len(args) == 0 {
		return sig
	}
	for i := 0; i < len(args); i++ {
		a := strings.Trim(args[i], "\"'")
		lower := strings.ToLower(strings.TrimSpace(a))
		if lower == "" {
			continue
		}
		switch lower {
		case "coingecko", "cg":
			sig.Source = "coingecko"
			continue
		case "--chain":
			if i+1 < len(args) {
				chRaw := strings.Trim(args[i+1], "\"'")
				if ch := normalizeWeb3Chain(chRaw); ch != "" {
					sig.Chain = ch
				} else if chRaw != "" {
					sig.InvalidChain = chRaw
				}
				i++
			}
			continue
		case "--limit":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(strings.Trim(args[i+1], "\"'")); err == nil {
					sig.Limit = n
				}
				i++
			}
			continue
		}
		if strings.HasPrefix(lower, "--limit=") {
			if n, err := strconv.Atoi(strings.TrimPrefix(lower, "--limit=")); err == nil {
				sig.Limit = n
			}
			continue
		}
		if strings.HasPrefix(lower, "--chain=") {
			chRaw := strings.TrimPrefix(lower, "--chain=")
			if ch := normalizeWeb3Chain(chRaw); ch != "" {
				sig.Chain = ch
			} else if chRaw != "" {
				sig.InvalidChain = chRaw
			}
			continue
		}
		if n, err := strconv.Atoi(lower); err == nil {
			sig.Limit = n
			continue
		}
		if ch := normalizeWeb3Chain(lower); ch != "" {
			sig.Chain = ch
			continue
		}
		if !strings.HasPrefix(lower, "--") && sig.Source != "coingecko" && sig.InvalidChain == "" {
			sig.InvalidChain = lower
		}
	}
	if sig.Source == "coingecko" {
		sig.Chain = ""
		sig.InvalidChain = ""
	}
	return sig
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
	fields = append(fields, buildWeb3SignalFields(ctx, cfg, best, tokenAddr, currentPrice, best.PriceChange.H24, mcap, firstScan, created)...)

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
	tickerStats, _ := fetchCoinGeckoTickerStats(ctx, coin.ID)
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
	snapshotVol := m.TotalVolume
	if tickerStats.Volume24hUSD > snapshotVol {
		snapshotVol = tickerStats.Volume24hUSD
	}
	stats := marketSnapshotRow(
		formatUSD(m.CurrentPrice),
		formatPercent(change),
		formatUSDCompact(m.MarketCap),
		formatUSDCompact(fdv),
		"n/a",
		formatUSDCompact(snapshotVol),
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
	if tickerStats.DepthUpUSD > 0 || tickerStats.DepthDownUSD > 0 || tickerStats.Volume24hUSD > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "CEX Market Depth",
			Inline: false,
			Value:  fmt.Sprintf("Vol %s • +2%% depth %s • -2%% depth %s • trusted markets %d", formatUSDCompact(tickerStats.Volume24hUSD), formatUSDCompact(tickerStats.DepthUpUSD), formatUSDCompact(tickerStats.DepthDownUSD), tickerStats.TrustedMarkets),
		})
	}
	fields = append(fields, buildCoinGeckoSignalFields(cfg, m.CurrentPrice, change, m.MarketCap, fdv, snapshotVol, tickerStats, firstScan, created)...)

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
	fields = append(fields, buildWeb3SignalFields(ctx, cfg, best, tokenAddr, currentPrice, best.PriceChange.H24, maxFloat(best.MarketCap, best.FDV), firstScan, created)...)
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

func (s *Service) resolveTrendingEmbed(ctx context.Context, sig web3Signal, cfg web3ModuleConfig) (*discordgo.MessageEmbed, error) {
	limit := cfg.TrendingCount
	if sig.Limit > 0 {
		limit = sig.Limit
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	source := strings.ToLower(strings.TrimSpace(sig.Source))
	if source == "" {
		source = "onchain"
	}
	switch source {
	case "coingecko":
		return s.resolveCoinGeckoTrendingEmbed(ctx, limit)
	default:
		if sig.InvalidChain != "" {
			return unsupportedTrendingChainEmbed(sig.InvalidChain), nil
		}
		return s.resolveOnchainTrendingEmbed(ctx, sig.Chain, limit)
	}
}

func (s *Service) resolveCoinGeckoTrendingEmbed(ctx context.Context, limit int) (*discordgo.MessageEmbed, error) {
	var resp cgTrendingResponse
	endpoint := "https://api.coingecko.com/api/v3/search/trending"
	if limit > 15 {
		endpoint += "?show_max=coins"
	}
	if err := web3FetchJSON(ctx, endpoint, &resp); err != nil {
		return nil, err
	}
	if len(resp.Coins) == 0 {
		return nil, nil
	}
	maxN := minInt(limit, len(resp.Coins))
	lines := make([]string, 0, maxN)
	for i := 0; i < maxN; i++ {
		it := resp.Coins[i].Item
		symbol := strings.ToUpper(fallback(it.Symbol, "?"))
		name := fallback(it.Name, symbol)
		price := formatUSD(parseAnyFloat(it.Data.Price))
		chg := formatPercent(parseAnyFloat(it.Data.PriceChangePercent24h.USD))
		rank := "n/a"
		if it.MarketCapRank > 0 {
			rank = fmt.Sprintf("#%d", it.MarketCapRank)
		}
		url := "https://www.coingecko.com/en/coins/" + strings.TrimSpace(it.ID)
		lines = append(lines, fmt.Sprintf("%d. [%s](%s) (%s) • %s • 24h %s • mcap %s", i+1, trimEmbedText(name, 26), url, symbol, price, chg, rank))
	}
	return &discordgo.MessageEmbed{
		Title:       "Trending • CoinGecko",
		Description: trimEmbedText(strings.Join(lines, "\n"), 3900),
		Color:       0x3498db,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Requested %d • Returned %d", limit, len(lines)),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) resolveOnchainTrendingEmbed(ctx context.Context, chain string, limit int) (*discordgo.MessageEmbed, error) {
	ch := normalizeWeb3Chain(chain)
	if ch == "" {
		ch = "solana"
	}
	network := geckoNetworkForChain(ch)
	if network == "" {
		return unsupportedTrendingChainEmbed(ch), nil
	}
	rows := make([]struct {
		Name      string
		Address   string
		PriceUSD  float64
		Change24h float64
		LiqUSD    float64
		Vol24hUSD float64
	}, 0, limit)
	for page := 1; len(rows) < limit && page <= 5; page++ {
		var resp gtTrendingPoolsResponse
		endpoint := fmt.Sprintf("https://api.geckoterminal.com/api/v2/networks/%s/trending_pools?page=%d", url.PathEscape(network), page)
		if err := web3FetchJSON(ctx, endpoint, &resp); err != nil {
			return nil, err
		}
		if len(resp.Data) == 0 {
			break
		}
		for _, item := range resp.Data {
			addr := strings.TrimSpace(item.Attributes.Address)
			if addr == "" {
				addr = strings.TrimPrefix(strings.TrimSpace(item.ID), network+"_")
			}
			if addr == "" {
				continue
			}
			rows = append(rows, struct {
				Name      string
				Address   string
				PriceUSD  float64
				Change24h float64
				LiqUSD    float64
				Vol24hUSD float64
			}{
				Name:      strings.TrimSpace(item.Attributes.Name),
				Address:   addr,
				PriceUSD:  parseAnyFloat(item.Attributes.BaseTokenPriceUSD),
				Change24h: parseAnyFloat(item.Attributes.PriceChangePercentage["h24"]),
				LiqUSD:    parseAnyFloat(item.Attributes.ReserveInUSD),
				Vol24hUSD: parseAnyFloat(item.Attributes.VolumeUSD["h24"]),
			})
			if len(rows) >= limit {
				break
			}
		}
	}
	if len(rows) == 0 {
		return nil, nil
	}
	lines := make([]string, 0, len(rows))
	for i, r := range rows {
		name := trimEmbedText(fallback(r.Name, r.Address), 28)
		link := "https://www.geckoterminal.com/" + url.PathEscape(network) + "/pools/" + url.PathEscape(r.Address)
		lines = append(lines, fmt.Sprintf("%d. [%s](%s) • %s • 24h %s • Liq %s • Vol %s", i+1, name, link, formatUSD(r.PriceUSD), formatPercent(r.Change24h), formatUSDCompact(r.LiqUSD), formatUSDCompact(r.Vol24hUSD)))
	}
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Trending • %s", formatChain(ch)),
		Description: trimEmbedText(strings.Join(lines, "\n"), 3900),
		Color:       0x2ecc71,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Source: GeckoTerminal • Requested %d • Returned %d", limit, len(lines)),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func unsupportedTrendingChainEmbed(input string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "Trending • Unsupported Chain",
		Description: fmt.Sprintf(
			"`%s` is not supported.\nSupported chains: `%s`\nExamples: `!trending solana`, `!trending base`, `!trending coingecko`",
			strings.TrimSpace(input),
			strings.Join(supportedTrendingChains(), "`, `"),
		),
		Color:     0xe67e22,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func buildWeb3SignalFields(ctx context.Context, cfg web3ModuleConfig, pair *dexPair, tokenAddr string, _ float64, change24h, _ float64, _ models.Web3FirstScanRow, _ bool) []*discordgo.MessageEmbedField {
	if pair == nil {
		return nil
	}
	lines := make([]string, 0, 8)

	if cfg.WhaleAlertsEnabled {
		if whaleLine, ok := detectWhaleBuyTrade(ctx, pair, cfg.WhaleMinTradeUSD); ok {
			lines = append(lines, whaleLine)
		} else {
			buys := pair.Txns.H24.Buys
			sells := pair.Txns.H24.Sells
			total := buys + sells
			if total > 0 && pair.Volume.H24 > 0 {
				estBuyUSD := pair.Volume.H24 * (float64(buys) / float64(total))
				if estBuyUSD >= cfg.WhaleMinTradeUSD && buys > sells && change24h > 0 {
					lines = append(lines, fmt.Sprintf("Whale Buy Pressure (estimate): %s (thr %s)", formatUSDCompact(estBuyUSD), formatUSDCompact(cfg.WhaleMinTradeUSD)))
				}
			}
		}
	}
	if cfg.PriceAlertsEnabled {
		if change24h >= cfg.PriceAlertPumpPct {
			lines = append(lines, fmt.Sprintf("Price Alert: pump %s (thr +%.0f%%)", formatPercent(change24h), cfg.PriceAlertPumpPct))
		} else if change24h <= -cfg.PriceAlertDumpPct {
			lines = append(lines, fmt.Sprintf("Price Alert: dump %s (thr -%.0f%%)", formatPercent(change24h), cfg.PriceAlertDumpPct))
		}
	}
	if cfg.WalletWatchEnabled && tokenAddr != "" {
		if _, ok := cfg.WalletWatchlist[normalizeContractKey(tokenAddr)]; ok {
			lines = append(lines, "Watchlist: token matched your guild watchlist")
		}
	}
	if len(lines) == 0 {
		return nil
	}
	return []*discordgo.MessageEmbedField{
		{
			Name:   "Signal Summary",
			Inline: false,
			Value:  trimEmbedText(strings.Join(lines, "\n"), 1024),
		},
	}
}

func buildCoinGeckoSignalFields(cfg web3ModuleConfig, _ float64, change24h, _, _, _ float64, _ cgTickerStats, _ models.Web3FirstScanRow, _ bool) []*discordgo.MessageEmbedField {
	lines := make([]string, 0, 1)
	if cfg.PriceAlertsEnabled {
		if change24h >= cfg.PriceAlertPumpPct {
			lines = append(lines, fmt.Sprintf("Price Alert: pump %s (thr +%.0f%%)", formatPercent(change24h), cfg.PriceAlertPumpPct))
		} else if change24h <= -cfg.PriceAlertDumpPct {
			lines = append(lines, fmt.Sprintf("Price Alert: dump %s (thr -%.0f%%)", formatPercent(change24h), cfg.PriceAlertDumpPct))
		}
	}
	if len(lines) == 0 {
		return nil
	}
	return []*discordgo.MessageEmbedField{
		{
			Name:   "Signal Summary",
			Inline: false,
			Value:  trimEmbedText(strings.Join(lines, "\n"), 1024),
		},
	}
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

func fetchCoinGeckoTickerStats(ctx context.Context, coinID string) (cgTickerStats, error) {
	var resp cgTickersResponse
	u := "https://api.coingecko.com/api/v3/coins/" + url.PathEscape(strings.TrimSpace(coinID)) + "/tickers?depth=true"
	if err := web3FetchJSON(ctx, u, &resp); err != nil {
		return cgTickerStats{}, err
	}
	stats := cgTickerStats{}
	for _, t := range resp.Tickers {
		if t.IsStale || t.IsAnomaly {
			continue
		}
		trusted := strings.EqualFold(strings.TrimSpace(t.TrustScore), "green")
		if !trusted {
			continue
		}
		stats.TrustedMarkets++
		stats.Volume24hUSD += t.ConvertedVolume["usd"]
		if t.CostToMoveUpUSD != nil && *t.CostToMoveUpUSD > 0 {
			stats.DepthUpUSD += *t.CostToMoveUpUSD
		}
		if t.CostToMoveDownUSD != nil && *t.CostToMoveDownUSD > 0 {
			stats.DepthDownUSD += *t.CostToMoveDownUSD
		}
	}
	if stats.TrustedMarkets > 0 {
		return stats, nil
	}
	// Fallback to all non-stale markets when trust metadata is absent.
	for _, t := range resp.Tickers {
		if t.IsStale || t.IsAnomaly {
			continue
		}
		stats.TrustedMarkets++
		stats.Volume24hUSD += t.ConvertedVolume["usd"]
		if t.CostToMoveUpUSD != nil && *t.CostToMoveUpUSD > 0 {
			stats.DepthUpUSD += *t.CostToMoveUpUSD
		}
		if t.CostToMoveDownUSD != nil && *t.CostToMoveDownUSD > 0 {
			stats.DepthDownUSD += *t.CostToMoveDownUSD
		}
	}
	return stats, nil
}

func detectWhaleBuyTrade(ctx context.Context, pair *dexPair, minUSD float64) (string, bool) {
	if pair == nil || strings.TrimSpace(pair.PairAddr) == "" || strings.TrimSpace(pair.ChainID) == "" || minUSD <= 0 {
		return "", false
	}
	network := geckoNetworkForChain(pair.ChainID)
	if network == "" {
		return "", false
	}
	endpoint := fmt.Sprintf(
		"https://api.geckoterminal.com/api/v2/networks/%s/pools/%s/trades?trade_volume_in_usd_greater_than=%d",
		url.PathEscape(network),
		url.PathEscape(strings.TrimSpace(pair.PairAddr)),
		int(minUSD),
	)
	var resp gtTradesResponse
	if err := web3FetchJSON(ctx, endpoint, &resp); err != nil {
		return "", false
	}
	maxBuyUSD := 0.0
	var maxBuyAt time.Time
	for _, t := range resp.Data {
		kind := strings.ToLower(strings.TrimSpace(t.Attributes.Kind))
		if kind != "buy" {
			continue
		}
		usd := parseDecimal(t.Attributes.VolumeInUSD)
		if usd <= 0 {
			continue
		}
		ts := parseTimeBestEffort(t.Attributes.BlockTimestamp)
		if usd > maxBuyUSD {
			maxBuyUSD = usd
			maxBuyAt = ts
		}
	}
	if maxBuyUSD < minUSD {
		return "", false
	}
	when := "recently"
	if !maxBuyAt.IsZero() {
		mins := int(time.Since(maxBuyAt).Minutes())
		if mins < 1 {
			when = "<1m ago"
		} else if mins < 120 {
			when = fmt.Sprintf("%dm ago", mins)
		} else {
			when = maxBuyAt.UTC().Format("2006-01-02 15:04 UTC")
		}
	}
	return fmt.Sprintf("Whale Buy Trade: largest recent buy %s (%s)", formatUSDCompact(maxBuyUSD), when), true
}

func geckoNetworkForChain(chainID string) string {
	switch normalizeWeb3Chain(chainID) {
	case "ethereum":
		return "eth"
	case "arbitrum":
		return "arbitrum"
	case "optimism":
		return "optimism"
	case "base":
		return "base"
	case "polygon":
		return "polygon_pos"
	case "bsc":
		return "bsc"
	case "solana":
		return "solana"
	default:
		return ""
	}
}

func supportedTrendingChains() []string {
	return []string{
		"ethereum",
		"arbitrum",
		"optimism",
		"base",
		"polygon",
		"bsc",
		"solana",
	}
}

func parseTimeBestEffort(raw string) time.Time {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05Z07:00"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
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

func parseAnyFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	case string:
		return parseDecimal(t)
	default:
		return 0
	}
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

func minInt(a, b int) int {
	if a <= b {
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
