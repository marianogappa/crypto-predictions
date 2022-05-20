package compiler

import (
	"time"

	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/jsonschema-go"
)

type PredictionSummary struct {
	CandlestickMap  map[string][]types.Candlestick `json:"candlestickMap,omitempty"`
	Coin            string                         `json:"coin,omitempty"`
	OtherCoin       string                         `json:"otherCoin,omitempty"`
	Goal            types.JsonFloat64              `json:"goal,omitempty"`
	RangeLow        types.JsonFloat64              `json:"rangeLow,omitempty"`
	RangeHigh       types.JsonFloat64              `json:"rangeHigh,omitempty"`
	WillReach       types.JsonFloat64              `json:"willReach,omitempty"`
	BeforeItReaches types.JsonFloat64              `json:"beforeItReaches,omitempty"`
	Operator        string                         `json:"operator,omitempty"`
	Deadline        types.ISO8601                  `json:"deadline,omitempty"`
	PredictionType  string                         `json:"predictionType,omitempty"`
}

func (PredictionSummary) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithExamples(PredictionSummary{
		CandlestickMap: map[string][]types.Candlestick{
			"BINANCE:COIN:BTC-USDT": {
				{Timestamp: 1651161957, OpenPrice: 39000, HighestPrice: 39500, LowestPrice: 39000, ClosePrice: 39050},
				{Timestamp: 1651162017, OpenPrice: 39500, HighestPrice: 39550, LowestPrice: 39200, ClosePrice: 39020},
			},
		},
		Coin:           "BINANCE:COIN:BTC-USDT",
		Goal:           45000,
		Operator:       ">=",
		Deadline:       "2022-06-24T07:51:06Z",
		PredictionType: "PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE",
	})

	return nil
}

func (s PredictionSerializer) BuildPredictionMarketSummary(p types.Prediction) (PredictionSummary, error) {
	switch p.Type {
	case types.PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE:
		return s.predictionTypeCoinOperatorFloatDeadline(p)
	case types.PREDICTION_TYPE_COIN_WILL_RANGE:
		return s.predictionTypeCoinWillRange(p)
	case types.PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES:
		return s.predictionTypeCoinWillReachBeforeItReaches(p)
	case types.PREDICTION_TYPE_THE_FLIPPENING:
		return s.predictionTypeTheFlippening(p)
	}
	return PredictionSummary{}, nil
}

func (s PredictionSerializer) predictionTypeCoinOperatorFloatDeadline(p types.Prediction) (PredictionSummary, error) {
	finalTs := time.Now().Round(time.Minute).Add(-2 * time.Minute)
	if p.State.Status == types.FINISHED {
		finalTs = time.Unix(int64(p.State.LastTs), 0)
	}

	tmFirstTick := finalTs.Add(-60 * time.Minute)
	coin := p.Predict.Predict.Literal.Operands[0]
	operator := p.Predict.Predict.Literal.Operator
	goal := p.Predict.Predict.Literal.Operands[1].Number
	initialISO8601 := types.ISO8601(tmFirstTick.Format(time.RFC3339))
	deadline := types.ISO8601(time.Unix(int64(p.Predict.Predict.Literal.ToTs), 0).UTC().Format(time.RFC3339))

	candlesticks := map[string][]types.Candlestick{}
	opStr := coin.Str
	it, err := (*s.mkt).GetIterator(coin, initialISO8601, false, 1)
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < 60; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return PredictionSummary{
		PredictionType: p.Type.String(),
		CandlestickMap: candlesticks,
		Operator:       operator,
		Goal:           goal,
		Deadline:       deadline,
		Coin:           opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeCoinWillRange(p types.Prediction) (PredictionSummary, error) {
	finalTs := time.Now().Round(time.Minute).Add(-2 * time.Minute)
	if p.State.Status == types.FINISHED {
		finalTs = time.Unix(int64(p.State.LastTs), 0)
	}

	tmFirstTick := finalTs.Add(-60 * time.Minute)
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
	it, err := (*s.mkt).GetIterator(coin, initialISO8601, false, 1)
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < 60; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return PredictionSummary{
		PredictionType: p.Type.String(),
		CandlestickMap: candlesticks,
		Deadline:       deadline,
		RangeLow:       rangeLow,
		RangeHigh:      rangeHigh,
		Coin:           opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeCoinWillReachBeforeItReaches(p types.Prediction) (PredictionSummary, error) {
	finalTs := time.Now().Round(time.Minute).Add(-2 * time.Minute)
	if p.State.Status == types.FINISHED {
		finalTs = time.Unix(int64(p.State.LastTs), 0)
	}

	tmFirstTick := finalTs.Add(-60 * time.Minute)
	coin := p.Predict.Predict.Literal.Operands[0]

	willReach := p.Predict.Predict.Operands[0].Literal.Operands[1].Number
	beforeItReaches := p.Predict.Predict.Operands[1].Operands[0].Literal.Operands[1].Number

	initialISO8601 := types.ISO8601(tmFirstTick.Format(time.RFC3339))
	deadline := types.ISO8601(time.Unix(int64(p.Predict.Predict.Literal.ToTs), 0).UTC().Format(time.RFC3339))

	candlesticks := map[string][]types.Candlestick{}
	opStr := coin.Str
	it, err := (*s.mkt).GetIterator(coin, initialISO8601, false, 1)
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < 60; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return PredictionSummary{
		PredictionType:  p.Type.String(),
		CandlestickMap:  candlesticks,
		Deadline:        deadline,
		WillReach:       willReach,
		BeforeItReaches: beforeItReaches,
		Coin:            opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeTheFlippening(p types.Prediction) (PredictionSummary, error) {
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
	it1, err := (*s.mkt).GetIterator(marketCap1, initialISO8601, false, 60*24)
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
	it2, err := (*s.mkt).GetIterator(marketCap2, initialISO8601, false, 60*24)
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
		PredictionType: p.Type.String(),
		CandlestickMap: candlesticks,
		Operator:       operator,
		Deadline:       deadline,
		Coin:           opStr1,
		OtherCoin:      opStr2,
	}, nil
}
