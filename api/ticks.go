package api

import (
	"time"

	"github.com/marianogappa/predictions/types"
)

type PredictionSummary struct {
	CandlestickMap  map[string][]types.Candlestick `json:"candlestickMap"`
	Coin            string                         `json:"coin"`
	Goal            types.JsonFloat64              `json:"goal"`
	RangeLow        types.JsonFloat64              `json:"rangeLow"`
	RangeHigh       types.JsonFloat64              `json:"rangeHigh"`
	WillReach       types.JsonFloat64              `json:"willReach"`
	BeforeItReaches types.JsonFloat64              `json:"beforeItReaches"`
	Operator        string                         `json:"operator"`
	Deadline        types.ISO8601                  `json:"deadline"`
}

func (a *API) BuildPredictionMarketSummary(p types.Prediction) (PredictionSummary, error) {
	switch p.Type {
	case types.PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE:
		return a.predictionTypeCoinOperatorFloatDeadline(p)
	case types.PREDICTION_TYPE_COIN_WILL_RANGE:
		return a.predictionTypeCoinWillRange(p)
	case types.PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES:
		return a.predictionTypeCoinWillReachBeforeItReaches(p)
	case types.PREDICTION_TYPE_THE_FLIPPENING:
		return a.predictionTypeTheFlippening(p)
	}
	return PredictionSummary{}, nil
}

func (a *API) predictionTypeCoinOperatorFloatDeadline(p types.Prediction) (PredictionSummary, error) {
	finalTs := time.Now().Round(time.Minute).Add(-2 * time.Minute)
	if p.State.Status == types.FINISHED {
		finalTs = time.Unix(int64(p.State.LastTs), 0)
	}

	tmFirstTick := finalTs.Add(-120 * time.Minute)
	coin := p.Predict.Predict.Literal.Operands[0]
	operator := p.Predict.Predict.Literal.Operator
	goal := p.Predict.Predict.Literal.Operands[1].Number
	initialISO8601 := types.ISO8601(tmFirstTick.Format(time.RFC3339))
	deadline := types.ISO8601(time.Unix(int64(p.Predict.Predict.Literal.ToTs), 0).UTC().Format(time.RFC3339))

	candlesticks := map[string][]types.Candlestick{}
	opStr := coin.Str
	it, err := a.mkt.GetIterator(coin, initialISO8601, false)
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < 120; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return PredictionSummary{
		CandlestickMap: candlesticks,
		Operator:       operator,
		Goal:           goal,
		Deadline:       deadline,
	}, nil
}

func (a *API) predictionTypeCoinWillRange(p types.Prediction) (PredictionSummary, error) {
	finalTs := time.Now().Round(time.Minute).Add(-2 * time.Minute)
	if p.State.Status == types.FINISHED {
		finalTs = time.Unix(int64(p.State.LastTs), 0)
	}

	tmFirstTick := finalTs.Add(-120 * time.Minute)
	coin := p.Predict.Predict.Literal.Operands[0]

	rangeLow := p.Predict.Predict.Literal.Operands[1].Number
	rangeHigh := p.Predict.Predict.Literal.Operands[2].Number
	if rangeLow > rangeHigh {
		rangeLow, rangeHigh = rangeHigh, rangeLow
	}

	initialISO8601 := types.ISO8601(tmFirstTick.Format(time.RFC3339))
	deadline := types.ISO8601(time.Unix(int64(p.Predict.Predict.Literal.ToTs), 0).UTC().Format(time.RFC3339))

	candlesticks := map[string][]types.Candlestick{}
	opStr := coin.Str
	it, err := a.mkt.GetIterator(coin, initialISO8601, false)
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < 120; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return PredictionSummary{
		CandlestickMap: candlesticks,
		Deadline:       deadline,
		RangeLow:       rangeLow,
		RangeHigh:      rangeHigh,
	}, nil
}

func (a *API) predictionTypeCoinWillReachBeforeItReaches(p types.Prediction) (PredictionSummary, error) {
	finalTs := time.Now().Round(time.Minute).Add(-2 * time.Minute)
	if p.State.Status == types.FINISHED {
		finalTs = time.Unix(int64(p.State.LastTs), 0)
	}

	tmFirstTick := finalTs.Add(-120 * time.Minute)
	coin := p.Predict.Predict.Literal.Operands[0]

	willReach := p.Predict.Predict.Operands[0].Literal.Operands[1].Number
	beforeItReaches := p.Predict.Predict.Operands[1].Operands[0].Literal.Operands[1].Number

	initialISO8601 := types.ISO8601(tmFirstTick.Format(time.RFC3339))
	deadline := types.ISO8601(time.Unix(int64(p.Predict.Predict.Literal.ToTs), 0).UTC().Format(time.RFC3339))

	candlesticks := map[string][]types.Candlestick{}
	opStr := coin.Str
	it, err := a.mkt.GetIterator(coin, initialISO8601, false)
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < 120; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return PredictionSummary{
		CandlestickMap:  candlesticks,
		Deadline:        deadline,
		WillReach:       willReach,
		BeforeItReaches: beforeItReaches,
	}, nil
}

func (a *API) predictionTypeTheFlippening(p types.Prediction) (PredictionSummary, error) {
	finalTs := time.Now().Truncate(time.Hour * 24)
	if p.State.Status == types.FINISHED {
		finalTs = time.Unix(int64(p.State.LastTs), 0)
	}

	tmFirstTick := finalTs.Add(-30 * 24 * time.Hour)
	marketCap1 := p.Predict.Predict.Literal.Operands[0]
	marketCap2 := p.Predict.Predict.Literal.Operands[1]
	operator := p.Predict.Predict.Literal.Operator
	initialISO8601 := types.ISO8601(tmFirstTick.Format(time.RFC3339))
	deadline := types.ISO8601(time.Unix(int64(p.Predict.Predict.Literal.ToTs), 0).UTC().Format(time.RFC3339))

	candlesticks := map[string][]types.Candlestick{}
	opStr1 := marketCap1.Str
	it1, err := a.mkt.GetIterator(marketCap1, initialISO8601, false)
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < 120; i++ {
		candlestick, err := it1.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr1] = append(candlesticks[opStr1], candlestick)
	}

	opStr2 := marketCap2.Str
	it2, err := a.mkt.GetIterator(marketCap2, initialISO8601, false)
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < 120; i++ {
		candlestick, err := it2.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr2] = append(candlesticks[opStr2], candlestick)
	}

	return PredictionSummary{
		CandlestickMap: candlesticks,
		Operator:       operator,
		Deadline:       deadline,
	}, nil
}
