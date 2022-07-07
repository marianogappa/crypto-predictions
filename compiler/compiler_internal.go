package compiler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/marianogappa/predictions/core"
)

var (
	strCondition        = fmt.Sprintf(` *%v *%v *%v *`, strFloatOrVariable, strOperator, strFloatOrVariable)
	strBetweenCondition = fmt.Sprintf(` *%v +BETWEEN +%v +AND +%v *`, strFloatOrVariable, strFloatOrVariable, strFloatOrVariable)
	strFloatOrVariable  = fmt.Sprintf(`(%v|%v)`, strFloat, strVariable)
	strOperator         = `([>=!<]+)`
	strFloat            = `[0-9]+(.[0-9]*)?`
	strVariable         = `(COIN|MARKETCAP):([A-Z]+):([A-Z]+)(-([A-Z]+))?`
	rxVariable          = regexp.MustCompile(fmt.Sprintf("^%v$", strVariable))
	rxCondition         = regexp.MustCompile(strCondition)
	rxBetweenCondition  = regexp.MustCompile(strBetweenCondition)
	rxDurationWeeks     = regexp.MustCompile(`([0-9]+)w`)
	rxDurationDays      = regexp.MustCompile(`([0-9]+)d`)
	rxDurationMonths    = regexp.MustCompile(`([0-9]+)m`)
	rxDurationHours     = regexp.MustCompile(`([0-9]+)h`)
)

func mapOperandForTests(v string) (core.Operand, error) {
	return mapOperand(v)
}

func mapOperand(v string) (core.Operand, error) {
	v = strings.ToUpper(v)
	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return core.Operand{Type: core.NUMBER, Number: core.JSONFloat64(f), Str: v}, nil
	}
	matches := rxVariable.FindStringSubmatch(v)
	if len(matches) == 0 {
		return core.Operand{}, fmt.Errorf("%w: operand %v doesn't parse to float nor match the regex %v", core.ErrInvalidOperand, v, strVariable)
	}
	operandType, _ := core.OperandTypeFromString(matches[1])
	if operandType == core.MARKETCAP && matches[5] != "" {
		return core.Operand{}, core.ErrNonEmptyQuoteAssetOnNonCoin
	}
	if operandType == core.COIN && matches[5] == "" {
		return core.Operand{}, core.ErrEmptyQuoteAsset
	}
	if matches[3] == matches[5] {
		return core.Operand{}, core.ErrEqualBaseQuoteAssets
	}
	return core.Operand{
		Type:       operandType,
		Provider:   matches[2],
		BaseAsset:  matches[3],
		QuoteAsset: matches[5],
		Str:        v,
	}, nil
}

func mapOperands(ss []string) ([]core.Operand, error) {
	ops := []core.Operand{}
	for _, s := range ss {
		op, err := mapOperand(s)
		if err != nil {
			return ops, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func parseDuration(dur string, fromTime time.Time) (time.Duration, error) {
	dur = strings.ToLower(dur)
	if dur == "eoy" {
		year, _, _ := fromTime.Date()
		nextYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
		return nextYear.Sub(fromTime), nil
	}
	if dur == "eod" {
		year, month, day := fromTime.Date()
		nextYear := time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
		return nextYear.Sub(fromTime), nil
	}
	if dur == "eom" {
		year, month, _ := fromTime.Date()
		nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
		return nextMonth.Sub(fromTime), nil
	}
	if dur == "eony" {
		year, _, _ := fromTime.Date()
		nextYear := time.Date(year+2, 1, 1, 0, 0, 0, 0, time.UTC)
		return nextYear.Sub(fromTime), nil
	}
	if dur == "eonm" {
		year, month, _ := fromTime.Date()
		nextMonth := time.Date(year, month+2, 1, 0, 0, 0, 0, time.UTC)
		return nextMonth.Sub(fromTime), nil
	}
	if dur == "eond" {
		year, month, day := fromTime.Date()
		nextYear := time.Date(year, month, day+2, 0, 0, 0, 0, time.UTC)
		return nextYear.Sub(fromTime), nil
	}
	matches := rxDurationMonths.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		return time.Duration(24*30*num) * time.Hour, nil
	}
	matches = rxDurationWeeks.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		return time.Duration(24*7*num) * time.Hour, nil
	}
	matches = rxDurationDays.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		return time.Duration(24*num) * time.Hour, nil
	}
	matches = rxDurationHours.FindStringSubmatch(dur)
	if len(matches) == 2 {
		num, _ := strconv.Atoi(matches[1])
		return time.Duration(num) * time.Hour, nil
	}
	return 0, fmt.Errorf("%w: %v, only `[0-9]+[mwdh]` or `eoy` are accepted", core.ErrInvalidDuration, dur)
}

func mapFromTs(c Condition, postedAt core.ISO8601) (int, error) {
	s, err := c.FromISO8601.Seconds()
	if err == nil {
		return s, nil
	}
	if c.FromISO8601 != "" && err != nil {
		return 0, fmt.Errorf("%w for condition: %v", core.ErrInvalidFromISO8601, c.FromISO8601)
	}
	return postedAt.Seconds()
}

func mapToTs(c Condition, fromTs int) (int, error) {
	s, err := c.ToISO8601.Seconds()
	if err == nil {
		return s, nil
	}
	if c.ToISO8601 != "" && err != nil {
		return 0, fmt.Errorf("%w for condition: %v", core.ErrInvalidToISO8601, c.ToISO8601)
	}
	if c.ToISO8601 == "" && c.ToDuration == "" {
		return 0, core.ErrOneOfToISO8601ToDurationRequired
	}
	fromTime := time.Unix(int64(fromTs), 0)
	duration, err := parseDuration(c.ToDuration, fromTime)
	if err != nil {
		return 0, fmt.Errorf("invalid ToDuration for condition: %v, error: %w", c.ToDuration, err)
	}
	return int(fromTime.Add(duration).Unix()), nil
}

func mapCondition(c Condition, name string, postedAt core.ISO8601) (core.Condition, error) {
	var (
		operator    string
		strOperands []string
	)
	matchCondition := rxCondition.FindStringSubmatch(c.Condition)
	if len(matchCondition) == 0 {
		matchCondition := rxBetweenCondition.FindStringSubmatch(c.Condition)
		if len(matchCondition) == 0 {
			return core.Condition{}, fmt.Errorf("%w; expecting regex match for '%v' or '%v' but got '%v'", core.ErrInvalidConditionSyntax, rxCondition, rxBetweenCondition, c.Condition)
		}
		operator = "BETWEEN"
		strOperands = []string{matchCondition[1], matchCondition[8], matchCondition[15]}
	} else {
		operator = matchCondition[8]
		strOperands = []string{matchCondition[1], matchCondition[9]}
	}

	operands, err := mapOperands(strOperands)
	if err != nil {
		return core.Condition{}, fmt.Errorf("while parsing condition's operands: %w", err)
	}

	if operator != ">" && operator != "<" && operator != ">=" && operator != "<=" && operator != "BETWEEN" {
		return core.Condition{}, fmt.Errorf("%w %v", core.ErrUnknownConditionOperator, operator)
	}

	if c.ErrorMarginRatio > 0.3 {
		return core.Condition{}, fmt.Errorf("%w, but was %v", core.ErrErrorMarginRatioAbove30, c.ErrorMarginRatio)
	}

	stateValue, err := core.ConditionStateValueFromString(c.State.Value)
	if err != nil {
		return core.Condition{}, fmt.Errorf("while parsing condition's stateValue: %w", err)
	}

	stateStatus, err := core.ConditionStatusFromString(c.State.Status)
	if err != nil {
		return core.Condition{}, fmt.Errorf("while parsing condition's stateStatus: %w", err)
	}

	fromTs, err := mapFromTs(c, postedAt)
	if err != nil {
		return core.Condition{}, fmt.Errorf("while parsing condition's fromTs: %w", err)
	}

	toTs, err := mapToTs(c, fromTs)
	if err != nil {
		return core.Condition{}, fmt.Errorf("while parsing condition's toTs: %w", err)
	}

	return core.Condition{
		Name:             name,
		Operator:         operator,
		Operands:         operands,
		ErrorMarginRatio: c.ErrorMarginRatio,
		FromTs:           fromTs,
		ToTs:             toTs,
		ToDuration:       c.ToDuration,
		Assumed:          c.Assumed,
		State: core.ConditionState{
			Status:    stateStatus,
			LastTs:    c.State.LastTs,
			LastTicks: c.State.LastTicks,
			Value:     stateValue,
		},
	}, nil
}

func mapBoolExpr(expr *string, def map[string]*core.Condition) (*core.BoolExpr, error) {
	if expr == nil {
		return nil, nil
	}
	e, err := parseBoolExpr(*expr, def)
	if err != nil {
		return nil, err
	}
	return e, nil
}
