package types

import "time"

type PredictionTypeCoinWillReachBeforeItReaches struct {
	P Prediction
}

func (p PredictionTypeCoinWillReachBeforeItReaches) Coin() Operand {
	return p.P.Predict.Predict.Literal.Operands[0]
}

func (p PredictionTypeCoinWillReachBeforeItReaches) WillReach() JsonFloat64 {
	return p.P.Predict.Predict.Operands[0].Literal.Operands[1].Number
}

func (p PredictionTypeCoinWillReachBeforeItReaches) BeforeItReaches() JsonFloat64 {
	return p.P.Predict.Predict.Operands[1].Operands[0].Literal.Operands[1].Number
}

func (p PredictionTypeCoinWillReachBeforeItReaches) WillReachWithError() JsonFloat64 {
	willReach := p.WillReach()
	beforeItReaches := p.BeforeItReaches()
	errorMarginRatio := p.ErrorMarginRatio()

	if willReach > beforeItReaches {
		return willReach * JsonFloat64(1-errorMarginRatio)
	}
	return willReach * JsonFloat64(1+errorMarginRatio)
}

func (p PredictionTypeCoinWillReachBeforeItReaches) BeforeItReachesWithError() JsonFloat64 {
	willReach := p.WillReach()
	beforeItReaches := p.BeforeItReaches()
	errorMarginRatio := p.ErrorMarginRatio()

	if beforeItReaches > willReach {
		return beforeItReaches * JsonFloat64(1+errorMarginRatio)
	}
	return beforeItReaches * JsonFloat64(1-errorMarginRatio)
}

func (p PredictionTypeCoinWillReachBeforeItReaches) ErrorMarginRatio() float64 {
	return p.P.Predict.Predict.Operands[0].Literal.ErrorMarginRatio
}

func (p PredictionTypeCoinWillReachBeforeItReaches) Deadline() time.Time {
	return time.Unix(int64(p.P.Predict.Predict.Operands[0].Literal.ToTs), 0)
}

func (p PredictionTypeCoinWillReachBeforeItReaches) EndTime() time.Time {
	deadline := p.Deadline()
	if p.P.State.Status != FINISHED {
		return deadline
	}
	return time.Unix(int64(p.P.State.LastTs), 0)
}

type PredictionTypeCoinOperatorFloatDeadline struct {
	P Prediction
}

func (p PredictionTypeCoinOperatorFloatDeadline) Coin() Operand {
	return p.P.Predict.Predict.Literal.Operands[0]
}

func (p PredictionTypeCoinOperatorFloatDeadline) Goal() JsonFloat64 {
	return p.P.Predict.Predict.Literal.Operands[1].Number
}

func (p PredictionTypeCoinOperatorFloatDeadline) Operator() string {
	return p.P.Predict.Predict.Literal.Operator
}

func (p PredictionTypeCoinOperatorFloatDeadline) GoalWithError() JsonFloat64 {
	goal := p.Goal()
	operator := p.Operator()
	errorMarginRatio := p.ErrorMarginRatio()

	errorDirection := 1.0
	if operator == ">" || operator == ">=" {
		errorDirection = -1.0
	}
	return goal * JsonFloat64(1.0+errorDirection*errorMarginRatio)
}

func (p PredictionTypeCoinOperatorFloatDeadline) ErrorMarginRatio() float64 {
	return p.P.Predict.Predict.Literal.ErrorMarginRatio
}

func (p PredictionTypeCoinOperatorFloatDeadline) Deadline() time.Time {
	return time.Unix(int64(p.P.Predict.Predict.Literal.ToTs), 0)
}

func (p PredictionTypeCoinOperatorFloatDeadline) EndTime() time.Time {
	deadline := p.Deadline()
	if p.P.State.Status != FINISHED {
		return deadline
	}
	return time.Unix(int64(p.P.State.LastTs), 0)
}

type PredictionTypeCoinWillRange struct {
	P Prediction
}

func (p PredictionTypeCoinWillRange) Coin() Operand {
	return p.P.Predict.Predict.Literal.Operands[0]
}

func (p PredictionTypeCoinWillRange) RangeLow() JsonFloat64 {
	return p.P.Predict.Predict.Literal.Operands[1].Number
}

func (p PredictionTypeCoinWillRange) RangeHigh() JsonFloat64 {
	return p.P.Predict.Predict.Literal.Operands[2].Number
}

func (p PredictionTypeCoinWillRange) Operator() string {
	return p.P.Predict.Predict.Literal.Operator
}

func (p PredictionTypeCoinWillRange) RangeLowWithError() JsonFloat64 {
	rangeLow := p.RangeLow()
	errorMarginRatio := p.ErrorMarginRatio()
	errorDirection := -1.0
	return rangeLow * JsonFloat64(1.0+errorDirection*errorMarginRatio)
}

func (p PredictionTypeCoinWillRange) RangeHighWithError() JsonFloat64 {
	rangeLow := p.RangeLow()
	errorMarginRatio := p.ErrorMarginRatio()
	errorDirection := 1.0
	return rangeLow * JsonFloat64(1.0+errorDirection*errorMarginRatio)
}

func (p PredictionTypeCoinWillRange) ErrorMarginRatio() float64 {
	return p.P.Predict.Predict.Literal.ErrorMarginRatio
}

func (p PredictionTypeCoinWillRange) Deadline() time.Time {
	return time.Unix(int64(p.P.Predict.Predict.Literal.ToTs), 0)
}

func (p PredictionTypeCoinWillRange) EndTime() time.Time {
	deadline := p.Deadline()
	if p.P.State.Status != FINISHED {
		return deadline
	}
	return time.Unix(int64(p.P.State.LastTs), 0)
}
