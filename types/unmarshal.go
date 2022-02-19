package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/marianogappa/signal-checker/common"
)

var (
	rxVariableStr = `(COIN|MARKETCAP):([A-Z]+):([A-Z]+)-([A-Z]+)`
	rxVariable    = regexp.MustCompile(rxVariableStr)
)

func mapOperand(v string) (Operand, error) {
	v = strings.ToUpper(v)
	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return Operand{Type: NUMBER, Number: common.JsonFloat64(f), Str: v}, nil
	}
	matches := rxVariable.FindStringSubmatch(v)
	if len(matches) == 0 {
		return Operand{}, fmt.Errorf("operand %v doesn't parse to float nor match the regex %v", v, rxVariableStr)
	}
	operandType, err := OperandTypeFromString(matches[1])
	if err != nil {
		return Operand{}, fmt.Errorf("invalid operand type %v", matches[1])
	}
	return Operand{
		Type:       operandType,
		Provider:   matches[2],
		BaseAsset:  matches[3],
		QuoteAsset: matches[4],
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

func mapCondition(c condition, name string) (Condition, error) {
	log.Println("operands", c.Operands)
	operands, err := mapOperands(c.Operands)
	if err != nil {
		return Condition{}, err
	}

	if c.Operator != "==" && c.Operator != ">" && c.Operator != "<" && c.Operator != ">=" && c.Operator != "<=" && c.Operator != "BETWEEN" {
		return Condition{}, fmt.Errorf("unknown operator %v", c.Operator)
	}
	if c.Operator == "BETWEEN" && len(c.Operands) != 3 {
		return Condition{}, fmt.Errorf("operator BETWEEN requires 3 operands but %v were supplied", len(c.Operands))
	}
	if c.Operator != "BETWEEN" && len(c.Operands) != 2 {
		return Condition{}, fmt.Errorf("operator %v requires 2 operands but %v were supplied", c.Operator, len(c.Operands))
	}

	stateValue, err := ConditionStateValueFromString(c.State.Value)
	if err != nil {
		return Condition{}, err
	}

	stateStatus, err := ConditionStatusFromString(c.State.Status)
	if err != nil {
		return Condition{}, err
	}

	fromTs, err := c.FromTs.Seconds()
	if err != nil {
		return Condition{}, err
	}

	toTs, err := c.ToTs.Seconds()
	if err != nil {
		return Condition{}, err
	}

	return Condition{
		Name:     name,
		Operator: c.Operator,
		Operands: operands,
		FromTs:   fromTs,
		ToTs:     toTs,
		Assumed:  c.Assumed,
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

	if raw.AuthorHandle == "" {
		return errors.New("authorHandle cannot be empty")
	}
	(*p).AuthorHandle = raw.AuthorHandle
	(*p).CreatedAt = raw.CreatedAt
	(*p).Post = raw.Post
	(*p).Version = raw.Version

	(*p).Define = map[string]*Condition{}
	for name, condition := range raw.Define {
		c, err := mapCondition(condition, name)
		if err != nil {
			return err
		}
		(*p).Define[name] = &c
	}

	var (
		b *BoolExpr
	)
	if raw.PrePredict != nil {
		b, err = mapBoolExpr(raw.PrePredict.WrongIf, (*p).Define)
		if err != nil {
			return err
		}
		(*p).PrePredict.WrongIf = b

		b, err = mapBoolExpr(raw.PrePredict.AnnulledIf, (*p).Define)
		if err != nil {
			return err
		}
		(*p).PrePredict.AnnulledIf = b

		b, err = mapBoolExpr(raw.PrePredict.PredictIf, (*p).Define)
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
		b, err = mapBoolExpr(raw.Predict.WrongIf, (*p).Define)
		if err != nil {
			return err
		}
		(*p).Predict.WrongIf = b
	}

	if raw.Predict.AnnulledIf != nil {
		b, err = mapBoolExpr(raw.Predict.AnnulledIf, (*p).Define)
		if err != nil {
			return err
		}
		(*p).Predict.AnnulledIf = b
	}

	b, err = mapBoolExpr(&raw.Predict.Predict, (*p).Define)
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
