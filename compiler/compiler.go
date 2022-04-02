package compiler

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/marianogappa/predictions/types"
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

func MapOperandForTests(v string) (types.Operand, error) {
	return mapOperand(v)
}

func mapOperand(v string) (types.Operand, error) {
	v = strings.ToUpper(v)
	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return types.Operand{Type: types.NUMBER, Number: types.JsonFloat64(f), Str: v}, nil
	}
	matches := rxVariable.FindStringSubmatch(v)
	if len(matches) == 0 {
		return types.Operand{}, fmt.Errorf("%w: operand %v doesn't parse to float nor match the regex %v", types.ErrInvalidOperand, v, strVariable)
	}
	operandType, _ := types.OperandTypeFromString(matches[1])
	if operandType == types.MARKETCAP && matches[5] != "" {
		return types.Operand{}, types.ErrNonEmptyQuoteAssetOnNonCoin
	}
	if operandType == types.COIN && matches[5] == "" {
		return types.Operand{}, types.ErrEmptyQuoteAsset
	}
	if matches[3] == matches[5] {
		return types.Operand{}, types.ErrEqualBaseQuoteAssets
	}
	return types.Operand{
		Type:       operandType,
		Provider:   matches[2],
		BaseAsset:  matches[3],
		QuoteAsset: matches[5],
		Str:        v,
	}, nil
}

func mapOperands(ss []string) ([]types.Operand, error) {
	ops := []types.Operand{}
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
	return 0, fmt.Errorf("%w: %v, only `[0-9]+[mwdh]` or `eoy` are accepted", types.ErrInvalidDuration, dur)
}

func mapFromTs(c condition, postedAt types.ISO8601) (int, error) {
	s, err := c.FromISO8601.Seconds()
	if err == nil {
		return s, nil
	}
	if c.FromISO8601 != "" && err != nil {
		return 0, fmt.Errorf("%w for condition: %v", types.ErrInvalidFromISO8601, c.FromISO8601)
	}
	return postedAt.Seconds()
}

func mapToTs(c condition, fromTs int) (int, error) {
	s, err := c.ToISO8601.Seconds()
	if err == nil {
		return s, nil
	}
	if c.ToISO8601 != "" && err != nil {
		return 0, fmt.Errorf("%w for condition: %v", types.ErrInvalidToISO8601, c.ToISO8601)
	}
	if c.ToISO8601 == "" && c.ToDuration == "" {
		return 0, types.ErrOneOfToISO8601ToDurationRequired
	}
	fromTime := time.Unix(int64(fromTs), 0)
	duration, err := parseDuration(c.ToDuration, fromTime)
	if err != nil {
		return 0, fmt.Errorf("invalid ToDuration for condition: %v, error: %w", c.ToDuration, err)
	}
	return int(fromTime.Add(duration).Unix()), nil
}

func mapCondition(c condition, name string, postedAt types.ISO8601) (types.Condition, error) {
	var (
		operator    string
		strOperands []string
	)
	matchCondition := rxCondition.FindStringSubmatch(c.Condition)
	if len(matchCondition) == 0 {
		matchCondition := rxBetweenCondition.FindStringSubmatch(c.Condition)
		if len(matchCondition) == 0 {
			return types.Condition{}, fmt.Errorf("%w; expecting regex match for '%v' or '%v' but got '%v'", types.ErrInvalidConditionSyntax, rxCondition, rxBetweenCondition, c.Condition)
		}
		operator = "BETWEEN"
		strOperands = []string{matchCondition[1], matchCondition[8], matchCondition[15]}
	} else {
		operator = matchCondition[8]
		strOperands = []string{matchCondition[1], matchCondition[9]}
	}

	operands, err := mapOperands(strOperands)
	if err != nil {
		return types.Condition{}, err
	}

	if operator != ">" && operator != "<" && operator != ">=" && operator != "<=" && operator != "BETWEEN" {
		return types.Condition{}, fmt.Errorf("%w %v", types.ErrUnknownConditionOperator, operator)
	}

	if c.ErrorMarginRatio > 0.3 {
		return types.Condition{}, fmt.Errorf("%w, but was %v", types.ErrErrorMarginRatioAbove30, c.ErrorMarginRatio)
	}

	stateValue, err := types.ConditionStateValueFromString(c.State.Value)
	if err != nil {
		return types.Condition{}, err
	}

	stateStatus, err := types.ConditionStatusFromString(c.State.Status)
	if err != nil {
		return types.Condition{}, err
	}

	fromTs, err := mapFromTs(c, postedAt)
	if err != nil {
		return types.Condition{}, err
	}

	toTs, err := mapToTs(c, fromTs)
	if err != nil {
		return types.Condition{}, err
	}

	return types.Condition{
		Name:             name,
		Operator:         operator,
		Operands:         operands,
		ErrorMarginRatio: c.ErrorMarginRatio,
		FromTs:           fromTs,
		ToTs:             toTs,
		ToDuration:       c.ToDuration,
		Assumed:          c.Assumed,
		State: types.ConditionState{
			Status:    stateStatus,
			LastTs:    c.State.LastTs,
			LastTicks: c.State.LastTicks,
			Value:     stateValue,
		},
	}, nil
}

func mapBoolExpr(expr *string, def map[string]*types.Condition) (*types.BoolExpr, error) {
	if expr == nil {
		return nil, nil
	}
	e, err := parseBoolExpr(*expr, def)
	if err != nil {
		return nil, err
	}
	return e, nil
}

type PredictionCompiler struct {
	metadataFetcher *metadatafetcher.MetadataFetcher
	timeNow         func() time.Time
}

func NewPredictionCompiler(fetcher *metadatafetcher.MetadataFetcher, timeNow func() time.Time) PredictionCompiler {
	return PredictionCompiler{metadataFetcher: fetcher, timeNow: timeNow}
}

func (c PredictionCompiler) Compile(rawPredictionBs []byte) (types.Prediction, *types.Account, error) {
	rawPrediction := string(rawPredictionBs)
	p := types.Prediction{}

	raw := prediction{}
	err := json.Unmarshal([]byte(rawPrediction), &raw)
	if err != nil {
		return p, nil, fmt.Errorf("%w: %v", types.ErrInvalidJSON, err)
	}

	if raw.Reporter == "" {
		return p, nil, types.ErrEmptyReporter
	}

	if raw.PostUrl == "" {
		return p, nil, types.ErrEmptyPostURL
	}

	var account *types.Account

	// these fields should be fetchable using the Twitter/Youtube API, but only if they don't exist (to allow caching)
	if c.metadataFetcher != nil && (raw.PostAuthor == "" || raw.PostedAt == "") {
		metadata, err := c.metadataFetcher.Fetch(raw.PostUrl)
		if err == nil {
			account = &metadata.Author
			if raw.PostAuthor == "" {
				raw.PostAuthor = metadata.Author.Handle
			}
			if raw.PostAuthor == "" {
				raw.PostAuthor = metadata.Author.Name
			}
			if raw.PostedAt == "" {
				raw.PostedAt = metadata.PostCreatedAt
			}
		} else {
			log.Printf("PredictionCompiler.Compile: ignoring matadataFecher error: %v\n", err)
		}
	}
	if raw.PostAuthor == "" {
		return p, nil, types.ErrEmptyPostAuthor
	}
	if raw.PostedAt == "" {
		return p, nil, types.ErrEmptyPostedAt
	}
	if _, err := raw.PostedAt.Seconds(); err != nil {
		return p, nil, types.ErrInvalidPostedAt
	}
	if raw.Version == "" {
		raw.Version = "1.0.0"
	}
	if c.timeNow != nil && raw.CreatedAt == "" {
		raw.CreatedAt = types.ISO8601(c.timeNow().Format(time.RFC3339))
	}

	p.UUID = raw.UUID
	p.Reporter = raw.Reporter
	p.PostAuthor = raw.PostAuthor
	p.CreatedAt = raw.CreatedAt
	p.PostUrl = raw.PostUrl
	p.PostedAt = raw.PostedAt
	p.Version = raw.Version

	p.Given = map[string]*types.Condition{}
	for name, condition := range raw.Given {
		c, err := mapCondition(condition, name, raw.PostedAt)
		if err != nil {
			return p, nil, err
		}
		p.Given[name] = &c
	}

	var (
		b *types.BoolExpr
	)
	if raw.PrePredict != nil {
		b, err = mapBoolExpr(raw.PrePredict.WrongIf, p.Given)
		if err != nil {
			return p, nil, err
		}
		p.PrePredict.WrongIf = b

		b, err = mapBoolExpr(raw.PrePredict.AnnulledIf, p.Given)
		if err != nil {
			return p, nil, err
		}
		p.PrePredict.AnnulledIf = b

		b, err = mapBoolExpr(raw.PrePredict.Predict, p.Given)
		if err != nil {
			return p, nil, err
		}
		p.PrePredict.Predict = b

		p.PrePredict.IgnoreUndecidedIfPredictIsDefined = raw.PrePredict.IgnoreUndecidedIfPredictIsDefined
		p.PrePredict.AnnulledIfPredictIsFalse = raw.PrePredict.AnnulledIfPredictIsFalse

		if p.PrePredict.Predict == nil && (p.PrePredict.WrongIf != nil || p.PrePredict.AnnulledIf != nil) {
			return p, nil, types.ErrMissingRequiredPrePredictPredictIf
		}
	}

	if raw.Predict.WrongIf != nil {
		b, err = mapBoolExpr(raw.Predict.WrongIf, p.Given)
		if err != nil {
			return p, nil, err
		}
		p.Predict.WrongIf = b
	}

	if raw.Predict.AnnulledIf != nil {
		b, err = mapBoolExpr(raw.Predict.AnnulledIf, p.Given)
		if err != nil {
			return p, nil, err
		}
		p.Predict.AnnulledIf = b
	}

	if raw.Predict.Predict == "" {
		return p, nil, types.ErrEmptyPredict
	}
	b, err = mapBoolExpr(&raw.Predict.Predict, p.Given)
	if err != nil {
		return p, nil, err
	}
	p.Predict.Predict = *b
	p.Predict.IgnoreUndecidedIfPredictIsDefined = raw.Predict.IgnoreUndecidedIfPredictIsDefined

	p.Type = types.PredictionTypeFromString(raw.Type)
	if p.Type == types.PREDICTION_TYPE_UNSUPPORTED {
		p.Type = CalculatePredictionType(p)
	}

	status, err := types.ConditionStatusFromString(raw.PredictionState.Status)
	if err != nil {
		return p, nil, err
	}
	value, err := types.PredictionStateValueFromString(raw.PredictionState.Value)
	if err != nil {
		return p, nil, err
	}
	p.State = types.PredictionState{
		Status: status,
		LastTs: raw.PredictionState.LastTs,
		Value:  value,
	}

	return p, account, nil
}
