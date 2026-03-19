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
	web3HTTPClient    = &http.Client{Timeout: 4 * time.Second}
)

type web3Signal struct {
	Contract string
	CashTag  string
}

type dexScreenerTokenResponse struct {
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

func (s *Service) handleWeb3IntelMessage(ctx context.Context, m *discordgo.MessageCreate, settings models.GuildSettings) {
	if !settings.FeatureEnabled(models.FeatureWeb3Intel) {
		return
	}
	signal := detectWeb3Signal(m.Content)
	if signal.Contract == "" && signal.CashTag == "" {
		return
	}
	if !s.allowWeb3Lookup(m.ChannelID) {
		return
	}

	lookupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var embed *discordgo.MessageEmbed
	var err error
	switch {
	case signal.Contract != "":
		embed, err = s.resolveContractIntelEmbed(lookupCtx, signal.Contract)
	case signal.CashTag != "":
		embed, err = s.resolveCashTagEmbed(lookupCtx, signal.CashTag)
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

func (s *Service) allowWeb3Lookup(channelID string) bool {
	if channelID == "" {
		return false
	}
	s.web3Mu.Lock()
	defer s.web3Mu.Unlock()
	last := s.web3Last[channelID]
	now := time.Now().UTC()
	if !last.IsZero() && now.Sub(last) < 8*time.Second {
		return false
	}
	s.web3Last[channelID] = now
	return true
}

func detectWeb3Signal(content string) web3Signal {
	text := strings.TrimSpace(content)
	if text == "" {
		return web3Signal{}
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

func (s *Service) resolveContractIntelEmbed(ctx context.Context, contract string) (*discordgo.MessageEmbed, error) {
	var dex dexScreenerTokenResponse
	if err := web3FetchJSON(ctx, "https://api.dexscreener.com/latest/dex/tokens/"+url.PathEscape(contract), &dex); err != nil {
		return nil, err
	}
	best := chooseBestDexPair(dex.Pairs)
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
	stats := fmt.Sprintf(
		"Price: %s\n24h: %s\nMCap: %s\nFDV: %s\nLiquidity: %s\nVolume (24h): %s",
		formatUSD(parseDecimal(best.PriceUSD)),
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

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Market Snapshot",
			Value:  stats,
			Inline: true,
		},
	}
	if len(tradeLinks) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Quick Trade",
			Value:  formatQuickLinks(tradeLinks),
			Inline: true,
		})
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "Contract",
		Value: "`" + tokenAddr + "`",
	})

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

func (s *Service) resolveCashTagEmbed(ctx context.Context, token string) (*discordgo.MessageEmbed, error) {
	var search cgSearchResponse
	if err := web3FetchJSON(ctx, "https://api.coingecko.com/api/v3/search?query="+url.QueryEscape(token), &search); err != nil {
		return nil, err
	}
	coin := chooseCoinGeckoCoin(search.Coins, token)
	if coin == nil {
		return nil, nil
	}
	var markets []cgMarket
	u := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=" + url.QueryEscape(coin.ID)
	if err := web3FetchJSON(ctx, u, &markets); err != nil {
		return nil, err
	}
	if len(markets) == 0 {
		return nil, nil
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
	description := fmt.Sprintf(
		"%s (%s)\n%s",
		fallback(m.Name, coin.Name),
		strings.ToUpper(fallback(m.Symbol, coin.Symbol)),
		formatQuickLinks([]quickLink{
			{Label: "CoinGecko", URL: "https://www.coingecko.com/en/coins/" + coin.ID},
			{Label: "Search DexScreener", URL: "https://dexscreener.com/?q=" + url.QueryEscape(strings.ToUpper(token))},
		}),
	)
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Web3 Intel • $%s", strings.ToUpper(token)),
		Description: trimEmbedText(description, 1024),
		Color:       0x3498db,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Market Snapshot",
				Inline: false,
				Value: fmt.Sprintf("Price: %s\n24h: %s\nMCap: %s\nFDV: %s",
					formatUSD(m.CurrentPrice),
					formatPercent(change),
					formatUSDCompact(m.MarketCap),
					formatUSDCompact(fdv),
				),
			},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func chooseBestDexPair(pairs []dexPair) *dexPair {
	if len(pairs) == 0 {
		return nil
	}
	candidates := make([]dexPair, 0, len(pairs))
	for _, p := range pairs {
		if strings.TrimSpace(p.PriceUSD) == "" {
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
