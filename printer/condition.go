package printer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/marianogappa/predictions/core"
)

func printCondition(c core.Condition, ignoreFromTs bool) string {
	fromTs := formatTs(c.FromTs)
	toTs := formatTs(c.ToTs)

	temporalPart := fmt.Sprintf("from %v to %v ", fromTs, toTs)
	if ignoreFromTs {
		temporalPart = fmt.Sprintf("by %v ", toTs)
		if c.ToDuration != "" {
			temporalPart = parseDuration(c.ToDuration, time.Unix(int64(c.FromTs), 0))
		}
	}

	suffix := ""
	if len(c.Assumed) > 0 {
		suffix = fmt.Sprintf("(%v assumed from prediction text)", strings.Join(c.Assumed, ", "))
	}
	if c.Operator == "BETWEEN" {
		return fmt.Sprintf("%v BETWEEN %v AND %v %v%v", legacyParseOperand(c.Operands[0]), legacyParseOperand(c.Operands[1]), legacyParseOperand(c.Operands[2]), temporalPart, suffix)
	}
	return fmt.Sprintf("%v %v %v %v%v", legacyParseOperand(c.Operands[0]), c.Operator, legacyParseOperand(c.Operands[1]), temporalPart, suffix)
}

func formatTs(ts int) string {
	return time.Unix(int64(ts), 0).Format(time.RFC3339)
}

var (
	rxDurationWeeks  = regexp.MustCompile(`([0-9]+)w`)
	rxDurationDays   = regexp.MustCompile(`([0-9]+)d`)
	rxDurationMonths = regexp.MustCompile(`([0-9]+)m`)
	rxDurationHours  = regexp.MustCompile(`([0-9]+)h`)

	knownCoinNames = map[string]string{
		"BTC":   "Bitcoin",
		"XBT":   "Bitcoin",
		"ETH":   "Ethereum",
		"USDT":  "Tether",
		"USDC":  "USD Coin",
		"BNB":   "BNB",
		"XRP":   "XRP",
		"BUSD":  "Binance USD",
		"ADA":   "Cardano",
		"SOL":   "Solana",
		"DOGE":  "Dogecoin",
		"DOT":   "Polkadot",
		"DAI":   "Dai",
		"SHIB":  "Shiba Inu",
		"TRX":   "TRON",
		"AVAX":  "Avalanche",
		"LEO":   "UNUS SED LEO",
		"WBTC":  "Wrapped Bitcoin",
		"MATIC": "Polygon",
		"UNI":   "Uniswap",
		"LTC":   "Litecoin",
		"FTT":   "FTX Token",
		"LINK":  "Chainlink",
		"CRO":   "Cronos",
		"XLM":   "Stellar",
		"NEAR":  "NEAR Protocol",
		"ATOM":  "Cosmos",
		"XMR":   "Monero",
		"ALGO":  "Algorand",
		"BCH":   "Bitcoin Cash",
		"ETC":   "Ethereum Classic",
		"VET":   "VeChain",
		"MANA":  "Decentraland",
		"FLOW":  "Flow",
		"SAND":  "The Sandbox",
		"HBAR":  "Hedera",
		"APE":   "ApeCoin",
		"ICP":   "Internet Computer",
		"THETA": "Theta Network",
		"AXS":   "Axie Infinity",
		"XTZ":   "Tezos",
		"FIL":   "Filecoin",
		"HNT":   "Helium",
		"EGLD":  "Elrond",
		"TUSD":  "TrueUSD",
		"BSV":   "Bitcoin SV",
		"KCS":   "KuCoin Token",
		"MKR":   "Maker",
		"EOS":   "EOS",
		"ZEC":   "Zcash",
		"USDP":  "Pax Dollar",
		"AAVE":  "Aave",
		"HT":    "Huobi Token",
		"MIOTA": "IOTA",
		"BTT":   "BitTorrent-New",
		"XEC":   "eCash",
		"USDN":  "Neutrino USD",
		"GRT":   "The Graph",
		"QNT":   "Quant",
		"OKB":   "OKB",
		"RUNE":  "THORChain",
		"FTM":   "Fantom",
		"KLAY":  "Klaytn",
		"USDD":  "USDD",
		"NEO":   "Neo",
		"WAVES": "Waves",
		"PAXG":  "PAX Gold",
		"ZIL":   "Zilliqa",
		"BAT":   "Basic Attention Token",
		"CHZ":   "Chiliz",
		"GMT":   "STEPN",
		"LRC":   "Loopring",
		"STX":   "Stacks",
		"DASH":  "Dash",
		"CAKE":  "PancakeSwap",
		"ENJ":   "Enjin Coin",
		"KSM":   "Kusama",
		"GALA":  "Gala",
		"CELO":  "Celo",
		"FEI":   "Fei USD",
		"CRV":   "Curve DAO Token",
		"KAVA":  "Kava",
		"HOT":   "Holo",
		"AMP":   "Amp",
		"MINA":  "Mina",
		"1INCH": "1inch Network",
		"XEM":   "NEM",
		"NEXO":  "Nexo",
		"DCR":   "Decred",
		"COMP":  "Compound",
		"XDC":   "XDC Network",
		"GT":    "GateToken",
		"AR":    "Arweave",
		"STORJ": "Storj",
		"KDA":   "Kadena",
		"GNO":   "Gnosis",
		"SNX":   "Synthetix",
		"QTUM":  "Qtum",
		"XYM":   "Symbol",
		"BORA":  "BORA",
		"BTG":   "Bitcoin Gold",
		"LUNA":  "Luna",
	}

	stablecoinNames = map[string]string{
		"USDT": "Tether",
		"USD":  "USD",
		"USDC": "USD Coin",
		"BUSD": "Binance USD",
		"DAI":  "Dai",
		"TUSD": "TrueUSD",
		"USDP": "Pax Dollar",
		"USDN": "Neutrino USD",
	}
)

func parseOperand(op core.Operand, useDollarSign bool) (string, bool) {
	if op.Type == core.NUMBER {
		return parseNumber(op.Number, useDollarSign), useDollarSign
	}
	if op.Type == core.MARKETCAP {
		return fmt.Sprintf("%v's MarketCap", op.BaseAsset), false
	}
	suffix := ""
	if op.Provider != "BINANCE" {
		suffix = fmt.Sprintf(" (on %v)", op.Provider)
	}

	baseAssetIsKnownCoin := knownCoinNames[op.BaseAsset] != ""
	quoteAssetIsStableCoin := stablecoinNames[op.QuoteAsset] != ""

	parsedMarket := fmt.Sprintf("%v/%v%v", op.BaseAsset, op.QuoteAsset, suffix)
	if baseAssetIsKnownCoin && quoteAssetIsStableCoin {
		parsedMarket = fmt.Sprintf("%v%v", knownCoinNames[op.BaseAsset], suffix)
	}

	return parsedMarket, quoteAssetIsStableCoin
}

func legacyParseOperand(op core.Operand) string {
	s, _ := parseOperand(op, false)
	return s
}

func parseNumber(num core.JSONFloat64, useDollarSign bool) string {
	dollarSign := ""
	if useDollarSign {
		dollarSign = "$"
	}

	if num/1000.0 > 1 && int(num)%1000 == 0.0 {
		return fmt.Sprintf("%v%vk", dollarSign, num/1000.0)
	}
	if num/100.0 > 1 && int(num)%100 == 0.0 {
		return fmt.Sprintf("%v%vk", dollarSign, num/1000.0)
	}
	return fmt.Sprintf("%v%v", dollarSign, num)
}

func parseDuration(dur string, fromTime time.Time) string {
	dur = strings.ToLower(dur)
	if dur == "eoy" {
		return "by end of year"
	}
	if dur == "eom" {
		return "by end of month"
	}
	if dur == "eow" {
		return "by end of week"
	}
	if dur == "eony" {
		return "by end of next year"
	}
	if dur == "eonm" {
		return "by end of next month"
	}
	if dur == "eonw" {
		return "by end of next week"
	}
	if dur == "eod" {
		return "by end of day"
	}
	if dur == "eond" {
		return "by end of next day"
	}
	matches := rxDurationMonths.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		if num == 1 {
			return "within a month"
		}
		return fmt.Sprintf("in %v months", num)
	}
	matches = rxDurationWeeks.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		if num == 1 {
			return "within a week"
		}
		return fmt.Sprintf("in %v weeks", num)
	}
	matches = rxDurationDays.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		if num == 1 {
			return "within 1 day"
		}
		return fmt.Sprintf("in %v days", num)
	}
	matches = rxDurationHours.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		if num == 1 {
			return "within the hour"
		}
		return fmt.Sprintf("in %v hours", num)
	}
	return "by ???"
}
