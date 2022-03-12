package printer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/marianogappa/predictions/types"
)

func printCondition(c types.Condition, ignoreFromTs bool) string {
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
		return fmt.Sprintf("%v BETWEEN %v AND %v %v%v", parseOperand(c.Operands[0]), parseOperand(c.Operands[1]), parseOperand(c.Operands[2]), temporalPart, suffix)
	}
	return fmt.Sprintf("%v %v %v %v%v", parseOperand(c.Operands[0]), c.Operator, parseOperand(c.Operands[1]), temporalPart, suffix)
}

func formatTs(ts int) string {
	return time.Unix(int64(ts), 0).Format(time.RFC3339)
}

var (
	rxDurationWeeks  = regexp.MustCompile(`([0-9]+)w`)
	rxDurationDays   = regexp.MustCompile(`([0-9]+)d`)
	rxDurationMonths = regexp.MustCompile(`([0-9]+)m`)
	rxDurationHours  = regexp.MustCompile(`([0-9]+)h`)
)

func parseOperand(op types.Operand) string {
	if op.Type == types.NUMBER {
		return parseNumber(op.Number)
	}
	if op.Type == types.MARKETCAP {
		return fmt.Sprintf("%v's MarketCap", op.BaseAsset)
	}
	suffix := ""
	if op.Provider != "BINANCE" {
		suffix = fmt.Sprintf(" (on %v)", op.Provider)
	}
	return fmt.Sprintf("%v/%v%v", op.BaseAsset, op.QuoteAsset, suffix)
}

func parseNumber(num types.JsonFloat64) string {
	if num/1000.0 > 1 && int(num)%1000 == 0.0 {
		return fmt.Sprintf("%vk", num/1000.0)
	}
	if num/100.0 > 1 && int(num)%100 == 0.0 {
		return fmt.Sprintf("%vk", num/1000.0)
	}
	return fmt.Sprintf("%v", num)
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
	matches := rxDurationMonths.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		return fmt.Sprintf("in %v months", num)
	}
	matches = rxDurationWeeks.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		return fmt.Sprintf("in %v weeks", num)
	}
	matches = rxDurationDays.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		return fmt.Sprintf("in %v days", num)
	}
	matches = rxDurationHours.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		return fmt.Sprintf("in %v hours", num)
	}
	return "by ???"
}
