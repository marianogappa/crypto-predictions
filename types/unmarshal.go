package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/marianogappa/predictions/twitter"
	"github.com/marianogappa/signal-checker/common"
)

var (
	strCondition        = fmt.Sprintf(` *%v *%v *%v *`, strFloatOrVariable, strOperator, strFloatOrVariable)
	strBetweenCondition = fmt.Sprintf(` *%v +BETWEEN +%v +AND +%v *`, strFloatOrVariable, strFloatOrVariable, strFloatOrVariable)
	strFloatOrVariable  = fmt.Sprintf(`(%v|%v)`, strFloat, strVariable)
	strOperator         = `(	>=|<=|>|<)`
	strFloat            = `[0-9]+(.[0-9]*)?`
	strVariable         = `(COIN|MARKETCAP):([A-Z]+):([A-Z]+)(-([A-Z]+))?`
	rxVariable          = regexp.MustCompile(strVariable)
	rxCondition         = regexp.MustCompile(strCondition)
	rxBetweenCondition  = regexp.MustCompile(strBetweenCondition)
)

func mapOperand(v string) (Operand, error) {
	v = strings.ToUpper(v)
	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return Operand{Type: NUMBER, Number: common.JsonFloat64(f), Str: v}, nil
	}
	matches := rxVariable.FindStringSubmatch(v)
	if len(matches) == 0 {
		return Operand{}, fmt.Errorf("operand %v doesn't parse to float nor match the regex %v", v, strVariable)
	}
	operandType, err := OperandTypeFromString(matches[1])
	if err != nil {
		return Operand{}, fmt.Errorf("invalid operand type %v", matches[1])
	}
	return Operand{
		Type:       operandType,
		Provider:   matches[2],
		BaseAsset:  matches[3],
		QuoteAsset: matches[5],
		Str:        v,
	}, nil
}

func mapOperands(ss []string) ([]Operand, error) {
	ops := []Operand{}
	for _, s := range ss {
		op, err := mapOperand(s)
		if err != nil {
			return ops, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func mapFromTs(c condition, postedAt common.ISO8601) (int, error) {
	s, err := c.FromISO8601.Seconds()
	if err == nil {
		return s, nil
	}
	if c.FromISO8601 != "" && err != nil {
		return 0, fmt.Errorf("invalid FromISO8601 for condition: %v", c.FromISO8601)
	}
	return postedAt.Seconds()
}

func mapToTs(c condition, fromTs int) (int, error) {
	s, err := c.ToISO8601.Seconds()
	if err == nil {
		return s, nil
	}
	if c.ToISO8601 != "" && err != nil {
		return 0, fmt.Errorf("invalid ToISO8601 for condition: %v", c.ToISO8601)
	}
	fromTime := time.Unix(int64(fromTs), 0)
	duration, err := time.ParseDuration(c.ToDuration)
	if err != nil {
		return 0, fmt.Errorf("invalid ToDuration for condition: %v, error: %v", c.ToDuration, err)
	}
	return int(fromTime.Add(duration).Unix()), nil
}

func mapCondition(c condition, name string, postedAt common.ISO8601) (Condition, error) {
	var (
		operator    string
		strOperands []string
	)
	matchCondition := rxCondition.FindStringSubmatch(c.Condition)
	if len(matchCondition) == 0 {
		matchCondition := rxBetweenCondition.FindStringSubmatch(c.Condition)
		if len(matchCondition) == 0 {
			return Condition{}, fmt.Errorf("invalid condition; expecting regex match for '%v' or '%v' but got '%v'", rxCondition, rxBetweenCondition, c.Condition)
		}
		operator = "BETWEEN"
		strOperands = []string{matchCondition[1], matchCondition[8], matchCondition[15]}
	} else {
		operator = matchCondition[8]
		strOperands = []string{matchCondition[1], matchCondition[9]}
	}

	operands, err := mapOperands(strOperands)
	if err != nil {
		return Condition{}, err
	}

	if operator != ">" && operator != "<" && operator != ">=" && operator != "<=" && operator != "BETWEEN" {
		return Condition{}, fmt.Errorf("unknown operator %v", operator)
	}
	if operator == "BETWEEN" && len(operands) != 3 {
		return Condition{}, fmt.Errorf("operator BETWEEN requires 3 operands but %v were supplied", len(operands))
	}
	if operator != "BETWEEN" && len(operands) != 2 {
		return Condition{}, fmt.Errorf("operator %v requires 2 operands but %v were supplied", operator, len(operands))
	}

	if c.ErrorMarginRatio > 0.3 {
		return Condition{}, fmt.Errorf("error margin ratio above 30%% is not allowed, but was %v", c.ErrorMarginRatio)
	}

	stateValue, err := ConditionStateValueFromString(c.State.Value)
	if err != nil {
		return Condition{}, err
	}

	stateStatus, err := ConditionStatusFromString(c.State.Status)
	if err != nil {
		return Condition{}, err
	}

	fromTs, err := mapFromTs(c, postedAt)
	if err != nil {
		return Condition{}, err
	}

	toTs, err := mapToTs(c, fromTs)
	if err != nil {
		return Condition{}, err
	}

	return Condition{
		Name:             name,
		Operator:         operator,
		Operands:         operands,
		ErrorMarginRatio: c.ErrorMarginRatio,
		FromTs:           fromTs,
		ToTs:             toTs,
		ToDuration:       c.ToDuration,
		Assumed:          c.Assumed,
		State: ConditionState{
			Status: stateStatus,
			LastTs: c.State.LastTs,
			Value:  stateValue,
		},
	}, nil
}

func mapBoolExpr(expr *string, def map[string]*Condition) (*BoolExpr, error) {
	if expr == nil {
		return nil, nil
	}
	e, err := parseBoolExpr(*expr, def)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (p *Prediction) UnmarshalJSON(rawPredictionBs []byte) error {
	rawPrediction := string(rawPredictionBs)
	*p = Prediction{}

	raw := prediction{}
	err := json.Unmarshal([]byte(rawPrediction), &raw)
	if err != nil {
		return err
	}

	if raw.PostUrl == "" {
		return errors.New("postUrl cannot be empty")
	}
	// TODO these 3 fields should be fetchable using the Twitter API, but only if they don't exist (to allow caching)
	if raw.PostAuthor == "" || raw.PostedAt == "" {
		splitUrl := strings.Split(raw.PostUrl, "/")
		tweet, err := twitter.NewTwitter().GetTweetByID(splitUrl[len(splitUrl)-1])
		if err == nil {
			raw.PostAuthor = tweet.UserHandle
			raw.PostedAt = common.ISO8601(tweet.TweetCreatedAt)
		}
		log.Println(err)
	}
	if raw.PostAuthor == "" {
		return errors.New("postAuthor cannot be empty")
	}
	if raw.PostedAt == "" {
		return errors.New("postedAt cannot be empty")
	}
	if _, err := raw.PostedAt.Seconds(); err != nil {
		return errors.New("postedAt must be a valid ISO8601 timestamp")
	}
	if raw.Version == "" {
		raw.Version = "1.0.0"
	}
	if raw.CreatedAt == "" {
		raw.CreatedAt = common.ISO8601(time.Now().Format(time.RFC3339))
	}

	(*p).PostAuthor = raw.PostAuthor
	(*p).CreatedAt = raw.CreatedAt
	(*p).PostUrl = raw.PostUrl
	(*p).PostedAt = raw.PostedAt
	(*p).Version = raw.Version

	(*p).Given = map[string]*Condition{}
	for name, condition := range raw.Given {
		c, err := mapCondition(condition, name, raw.PostedAt)
		if err != nil {
			return err
		}
		(*p).Given[name] = &c
	}

	var (
		b *BoolExpr
	)
	if raw.PrePredict != nil {
		b, err = mapBoolExpr(raw.PrePredict.WrongIf, (*p).Given)
		if err != nil {
			return err
		}
		(*p).PrePredict.WrongIf = b

		b, err = mapBoolExpr(raw.PrePredict.AnnulledIf, (*p).Given)
		if err != nil {
			return err
		}
		(*p).PrePredict.AnnulledIf = b

		b, err = mapBoolExpr(raw.PrePredict.PredictIf, (*p).Given)
		if err != nil {
			return err
		}
		(*p).PrePredict.PredictIf = b

		if (*p).PrePredict.PredictIf == nil && ((*p).PrePredict.WrongIf != nil || (*p).PrePredict.AnnulledIf != nil) {
			err := errors.New("pre-predict clause must have predictIf if it has either wrongIf or annuledIf. Otherwise, add them directly on predict clause")
			return err
		}
	}

	if raw.Predict.WrongIf != nil {
		b, err = mapBoolExpr(raw.Predict.WrongIf, (*p).Given)
		if err != nil {
			return err
		}
		(*p).Predict.WrongIf = b
	}

	if raw.Predict.AnnulledIf != nil {
		b, err = mapBoolExpr(raw.Predict.AnnulledIf, (*p).Given)
		if err != nil {
			return err
		}
		(*p).Predict.AnnulledIf = b
	}

	b, err = mapBoolExpr(&raw.Predict.Predict, (*p).Given)
	if err != nil {
		return err
	}
	(*p).Predict.Predict = *b

	status, err := ConditionStatusFromString(raw.PredictionState.Status)
	if err != nil {
		return err
	}
	value, err := PredictionStateValueFromString(raw.PredictionState.Value)
	if err != nil {
		return err
	}
	(*p).State = PredictionState{
		Status: status,
		LastTs: raw.PredictionState.LastTs,
		Value:  value,
	}

	return nil
}

func CompilePrediction(bs []byte) (Prediction, error) {
	var prediction Prediction
	if err := json.Unmarshal(bs, &prediction); err != nil {
		return Prediction{}, err
	}
	return prediction, nil
}
