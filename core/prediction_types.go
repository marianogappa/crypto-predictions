package core

import (
	"math"
	"time"

	"github.com/marianogappa/crypto-candles/candles/common"
)

// PredictionTypeCoinWillReachBeforeItReachesWrapper is a prediction type. This type decorator provides value facades.
type PredictionTypeCoinWillReachBeforeItReachesWrapper struct {
	P Prediction
}

// Coin is the only relevant operand.
func (p PredictionTypeCoinWillReachBeforeItReachesWrapper) Coin() Operand {
	return p.P.Predict.Predict.Literal.Operands[0]
}

// WillReach is the price the coin must reach (without error).
func (p PredictionTypeCoinWillReachBeforeItReachesWrapper) WillReach() JSONFloat64 {
	return p.P.Predict.Predict.Operands[0].Literal.Operands[1].Number
}

// BeforeItReaches is the price the coin must not reach first (without error).
func (p PredictionTypeCoinWillReachBeforeItReachesWrapper) BeforeItReaches() JSONFloat64 {
	return p.P.Predict.Predict.Operands[1].Operands[0].Literal.Operands[1].Number
}

// WillReachWithError is the price the coin must reach (with error).
func (p PredictionTypeCoinWillReachBeforeItReachesWrapper) WillReachWithError() JSONFloat64 {
	willReach := p.WillReach()
	beforeItReaches := p.BeforeItReaches()
	errorMarginRatio := p.ErrorMarginRatio()

	if willReach > beforeItReaches {
		return willReach * JSONFloat64(1-errorMarginRatio)
	}
	return willReach * JSONFloat64(1+errorMarginRatio)
}

// BeforeItReachesWithError is the price the coin must not reach first (with error).
func (p PredictionTypeCoinWillReachBeforeItReachesWrapper) BeforeItReachesWithError() JSONFloat64 {
	willReach := p.WillReach()
	beforeItReaches := p.BeforeItReaches()
	errorMarginRatio := p.ErrorMarginRatio()

	if beforeItReaches > willReach {
		return beforeItReaches * JSONFloat64(1+errorMarginRatio)
	}
	return beforeItReaches * JSONFloat64(1-errorMarginRatio)
}

// ErrorMarginRatio is the error allowed to the price matching.
func (p PredictionTypeCoinWillReachBeforeItReachesWrapper) ErrorMarginRatio() float64 {
	return p.P.Predict.Predict.Operands[0].Literal.ErrorMarginRatio
}

// Deadline for the prediction (don't use for UI! Deprecated!)
func (p PredictionTypeCoinWillReachBeforeItReachesWrapper) Deadline() time.Time {
	return time.Unix(int64(p.P.Predict.Predict.Operands[0].Literal.ToTs), 0)
}

// EndTime is the time the prediction finished or will finish. If available, use EndTimeTruncatedDueToResultInvalidation
// instead!
func (p PredictionTypeCoinWillReachBeforeItReachesWrapper) EndTime() time.Time {
	deadline := p.Deadline()
	if p.P.State.Status != FINISHED {
		return deadline
	}
	return time.Unix(int64(p.P.State.LastTs), 0)
}

// PredictionTypeCoinOperatorFloatDeadlineWrapper is a prediction type. This type decorator provides value facades.
type PredictionTypeCoinOperatorFloatDeadlineWrapper struct {
	P Prediction
}

// Coin is the only relevant operand.
func (p PredictionTypeCoinOperatorFloatDeadlineWrapper) Coin() Operand {
	return p.P.Predict.Predict.Literal.Operands[0]
}

// Goal is the price goal for the coin (without error).
// Goal is the price goal for the coin (without error).
func (p PredictionTypeCoinOperatorFloatDeadlineWrapper) Goal() JSONFloat64 {
	return p.P.Predict.Predict.Literal.Operands[1].Number
}

// Operator is the operator for the prediction's condition: ">=" or "<=".
func (p PredictionTypeCoinOperatorFloatDeadlineWrapper) Operator() string {
	return p.P.Predict.Predict.Literal.Operator
}

// GoalWithError is the price goal for the coin (with error).
func (p PredictionTypeCoinOperatorFloatDeadlineWrapper) GoalWithError() JSONFloat64 {
	goal := p.Goal()
	operator := p.Operator()
	errorMarginRatio := p.ErrorMarginRatio()

	errorDirection := 1.0
	if operator == ">" || operator == ">=" {
		errorDirection = -1.0
	}
	return goal * JSONFloat64(1.0+errorDirection*errorMarginRatio)
}

// ErrorMarginRatio is the error allowed to the price matching.
func (p PredictionTypeCoinOperatorFloatDeadlineWrapper) ErrorMarginRatio() float64 {
	return p.P.Predict.Predict.Literal.ErrorMarginRatio
}

// Deadline for the prediction (don't use for UI! Deprecated!)
func (p PredictionTypeCoinOperatorFloatDeadlineWrapper) Deadline() time.Time {
	return time.Unix(int64(p.P.Predict.Predict.Literal.ToTs), 0)
}

// EndTime is the time the prediction finished or will finish. If available, use EndTimeTruncatedDueToResultInvalidation
// instead!
func (p PredictionTypeCoinOperatorFloatDeadlineWrapper) EndTime() time.Time {
	deadline := p.Deadline()
	if p.P.State.Status != FINISHED {
		return deadline
	}
	return time.Unix(int64(p.P.State.LastTs), 0)
}

// EndTimeTruncatedDueToResultInvalidation is the time the prediction finished or will finish, truncating for possible
// UI issues.
func (p PredictionTypeCoinOperatorFloatDeadlineWrapper) EndTimeTruncatedDueToResultInvalidation(candlesticks []common.Candlestick) time.Time {
	return endTimeTruncatedDueToResultInvalidation(p.P, p.P.Predict.Predict.Literal.Clone(), p.EndTime(), candlesticks)
}

// PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper is a prediction type. This type decorator provides value facades.
type PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper struct {
	P Prediction
}

// Coin is the only relevant operand.
func (p PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper) Coin() Operand {
	return p.P.Predict.Predict.Literal.Operands[0]
}

// Goal is the price goal for the coin (without error).
func (p PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper) Goal() JSONFloat64 {
	return p.P.Predict.Predict.Literal.Operands[1].Number
}

// InvalidatedIfItReaches is the price for the coin at which the prediction becomes annulled.
func (p PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper) InvalidatedIfItReaches() JSONFloat64 {
	return p.P.Predict.AnnulledIf.Literal.Operands[1].Number
}

// Operator is the operator for the prediction's condition: ">=" or "<=".
func (p PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper) Operator() string {
	return p.P.Predict.Predict.Literal.Operator
}

// GoalWithError is the price goal for the coin (with error).
func (p PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper) GoalWithError() JSONFloat64 {
	goal := p.Goal()
	operator := p.Operator()
	errorMarginRatio := p.ErrorMarginRatio()

	errorDirection := 1.0
	if operator == ">" || operator == ">=" {
		errorDirection = -1.0
	}
	return goal * JSONFloat64(1.0+errorDirection*errorMarginRatio)
}

// ErrorMarginRatio is the error allowed to the price matching.
func (p PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper) ErrorMarginRatio() float64 {
	return p.P.Predict.Predict.Literal.ErrorMarginRatio
}

// Deadline for the prediction (don't use for UI! Deprecated!)
func (p PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper) Deadline() time.Time {
	return time.Unix(int64(p.P.Predict.Predict.Literal.ToTs), 0)
}

// EndTime is the time the prediction finished or will finish. If available, use EndTimeTruncatedDueToResultInvalidation
// instead!
func (p PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper) EndTime() time.Time {
	deadline := p.Deadline()
	if p.P.State.Status != FINISHED {
		return deadline
	}
	return time.Unix(int64(p.P.State.LastTs), 0)
}

// EndTimeTruncatedDueToResultInvalidation is the time the prediction finished or will finish, truncating for possible
// UI issues.
func (p PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper) EndTimeTruncatedDueToResultInvalidation(candlesticks []common.Candlestick) time.Time {
	return endTimeTruncatedDueToResultInvalidation(p.P, p.P.Predict.Predict.Literal.Clone(), p.EndTime(), candlesticks)
}

func endTimeTruncatedDueToResultInvalidation(p Prediction, condition Condition, endTime time.Time, candlesticks []common.Candlestick) time.Time {
	if p.State.Status != FINISHED || p.State.Value != INCORRECT || len(candlesticks) < 2 {
		return endTime
	}

	// If prediction is FINISHED with value INCORRECT, there could be a big problem in the chart!
	// If the last candlestick extends over after endTime, it could happen that the prediction became true after it.
	// This will result in showing a chart where the prediction looks correct!
	// So, let's check if that's the case by evolving the prediction with the last candlestick.
	condition.ClearState()
	condition.FromTs = 0
	condition.ToTs = math.MaxInt // We must ignore the deadline

	var (
		coin                   = condition.Operands[0].Str
		lastCandlestick        = candlesticks[len(candlesticks)-1]
		penultimateCandlestick = candlesticks[len(candlesticks)-2]
		lastTimestamp          = lastCandlestick.Timestamp
		penultimateTimestamp   = penultimateCandlestick.Timestamp
		diffSeconds            = lastTimestamp - penultimateTimestamp
		nextTimestamp          = lastTimestamp + diffSeconds // Tick timestamp must be in the future
		lowTick                = Tick{Timestamp: nextTimestamp, Value: lastCandlestick.LowestPrice}
		highTick               = Tick{Timestamp: nextTimestamp + 1, Value: lastCandlestick.HighestPrice}
	)

	_ = condition.Run(map[string]Tick{coin: lowTick})
	_ = condition.Run(map[string]Tick{coin: highTick})

	if condition.Evaluate() != TRUE {
		return endTime
	}

	// Condition became TRUE after evolving it with the last candlestick! This means that the chart is going to look
	// wrong! To mitigate it, move back the end time between the last two candlesticks.
	var (
		midTimestamp = (lastTimestamp + penultimateTimestamp) / 2
		midTime      = time.Unix(int64(midTimestamp), 0)
	)

	return midTime
}

// PredictionTypeCoinWillRangeWrapper is a prediction type. This type decorator provides value facades.
type PredictionTypeCoinWillRangeWrapper struct {
	P Prediction
}

// Coin is the only relevant operand.
func (p PredictionTypeCoinWillRangeWrapper) Coin() Operand {
	return p.P.Predict.Predict.Literal.Operands[0]
}

// RangeLow that the coin will range between (without error)
func (p PredictionTypeCoinWillRangeWrapper) RangeLow() JSONFloat64 {
	return p.P.Predict.Predict.Literal.Operands[1].Number
}

// RangeHigh that the coin will range between (without error)
func (p PredictionTypeCoinWillRangeWrapper) RangeHigh() JSONFloat64 {
	return p.P.Predict.Predict.Literal.Operands[2].Number
}

// RangeLowWithError that the coin will range between (with error)
func (p PredictionTypeCoinWillRangeWrapper) RangeLowWithError() JSONFloat64 {
	rangeLow := p.RangeLow()
	errorMarginRatio := p.ErrorMarginRatio()
	errorDirection := -1.0
	return rangeLow * JSONFloat64(1.0+errorDirection*errorMarginRatio)
}

// RangeHighWithError that the coin will range between (with error)
func (p PredictionTypeCoinWillRangeWrapper) RangeHighWithError() JSONFloat64 {
	rangeLow := p.RangeLow()
	errorMarginRatio := p.ErrorMarginRatio()
	errorDirection := 1.0
	return rangeLow * JSONFloat64(1.0+errorDirection*errorMarginRatio)
}

// ErrorMarginRatio is the error allowed to the price matching.
func (p PredictionTypeCoinWillRangeWrapper) ErrorMarginRatio() float64 {
	return p.P.Predict.Predict.Literal.ErrorMarginRatio
}

// Deadline for the prediction (don't use for UI! Deprecated!)
func (p PredictionTypeCoinWillRangeWrapper) Deadline() time.Time {
	return time.Unix(int64(p.P.Predict.Predict.Literal.ToTs), 0)
}

// EndTime is the time the prediction finished or will finish. If available, use EndTimeTruncatedDueToResultInvalidation
// instead!
func (p PredictionTypeCoinWillRangeWrapper) EndTime() time.Time {
	deadline := p.Deadline()
	if p.P.State.Status != FINISHED {
		return deadline
	}
	return time.Unix(int64(p.P.State.LastTs), 0)
}
