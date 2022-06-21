package serializer

import (
	"errors"
	"time"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
)

// BuildPredictionMarketSummary builds a PredictionSummary from a prediction. It uses a market to get candlesticks.
func (s PredictionSerializer) BuildPredictionMarketSummary(p types.Prediction) (compiler.PredictionSummary, error) {
	switch p.Type {
	case types.PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE:
		return s.predictionTypeCoinOperatorFloatDeadline(p)
	case types.PREDICTION_TYPE_COIN_WILL_REACH_INVALIDATED_IF_IT_REACHES:
		return s.predictionTypeCoinWillReachInvalidatedIfItReaches(p)
	case types.PREDICTION_TYPE_COIN_WILL_RANGE:
		return s.predictionTypeCoinWillRange(p)
	case types.PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES:
		return s.predictionTypeCoinWillReachBeforeItReaches(p)
	case types.PREDICTION_TYPE_THE_FLIPPENING:
		return s.predictionTypeTheFlippening(p)
	}
	return compiler.PredictionSummary{}, nil
}

func (s PredictionSerializer) predictionTypeCoinOperatorFloatDeadline(p types.Prediction) (compiler.PredictionSummary, error) {
	typedPred := types.PredictionTypeCoinOperatorFloatDeadline{P: p}

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}

	candlesticks := map[string][]types.Candlestick{}
	opStr := typedPred.Coin().Str
	it, err := (*s.mkt).GetIterator(typedPred.Coin(), chartParams.startTimeISO8601(), false, chartParams.candlestickIntervalMinutes())
	if err != nil {
		return compiler.PredictionSummary{}, err
	}
	for i := 0; i < chartParams.candlestickCount; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return compiler.PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return compiler.PredictionSummary{
		PredictionType:                          p.Type.String(),
		CandlestickMap:                          candlesticks,
		Operator:                                typedPred.Operator(),
		Goal:                                    typedPred.Goal(),
		GoalWithError:                           typedPred.GoalWithError(),
		ErrorMarginRatio:                        types.JsonFloat64(typedPred.ErrorMarginRatio()),
		Deadline:                                types.ISO8601(typedPred.Deadline().Format(time.RFC3339)),
		EndedAt:                                 types.ISO8601(typedPred.EndTime().Format(time.RFC3339)),
		EndedAtTruncatedDueToResultInvalidation: types.ISO8601(typedPred.EndTimeTruncatedDueToResultInvalidation(candlesticks[opStr]).Format(time.RFC3339)),
		Coin:                                    opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeCoinWillReachInvalidatedIfItReaches(p types.Prediction) (compiler.PredictionSummary, error) {
	typedPred := types.PredictionTypeCoinWillReachInvalidatedIfItReaches{P: p}

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}

	candlesticks := map[string][]types.Candlestick{}
	opStr := typedPred.Coin().Str
	it, err := (*s.mkt).GetIterator(typedPred.Coin(), chartParams.startTimeISO8601(), false, chartParams.candlestickIntervalMinutes())
	if err != nil {
		return compiler.PredictionSummary{}, err
	}
	for i := 0; i < chartParams.candlestickCount; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return compiler.PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return compiler.PredictionSummary{
		PredictionType:                          p.Type.String(),
		CandlestickMap:                          candlesticks,
		Operator:                                typedPred.Operator(),
		Goal:                                    typedPred.Goal(),
		GoalWithError:                           typedPred.GoalWithError(),
		InvalidatedIfItReaches:                  typedPred.InvalidatedIfItReaches(),
		ErrorMarginRatio:                        types.JsonFloat64(typedPred.ErrorMarginRatio()),
		Deadline:                                types.ISO8601(typedPred.Deadline().Format(time.RFC3339)),
		EndedAt:                                 types.ISO8601(typedPred.EndTime().Format(time.RFC3339)),
		EndedAtTruncatedDueToResultInvalidation: types.ISO8601(typedPred.EndTimeTruncatedDueToResultInvalidation(candlesticks[opStr]).Format(time.RFC3339)),
		Coin:                                    opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeCoinWillRange(p types.Prediction) (compiler.PredictionSummary, error) {
	typedPred := types.PredictionTypeCoinWillRange{P: p}
	coin := typedPred.Coin()

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}

	candlesticks := map[string][]types.Candlestick{}
	opStr := coin.Str
	it, err := (*s.mkt).GetIterator(coin, chartParams.startTimeISO8601(), false, chartParams.candlestickIntervalMinutes())
	if err != nil {
		return compiler.PredictionSummary{}, err
	}
	for i := 0; i < chartParams.candlestickCount; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return compiler.PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return compiler.PredictionSummary{
		PredictionType:     p.Type.String(),
		CandlestickMap:     candlesticks,
		Deadline:           types.ISO8601(typedPred.Deadline().Format(time.RFC3339)),
		EndedAt:            types.ISO8601(typedPred.EndTime().Format(time.RFC3339)),
		ErrorMarginRatio:   types.JsonFloat64(typedPred.ErrorMarginRatio()),
		RangeLow:           typedPred.RangeLow(),
		RangeLowWithError:  typedPred.RangeLowWithError(),
		RangeHigh:          typedPred.RangeHigh(),
		RangeHighWithError: typedPred.RangeHighWithError(),
		Coin:               opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeCoinWillReachBeforeItReaches(p types.Prediction) (compiler.PredictionSummary, error) {
	typedPred := types.PredictionTypeCoinWillReachBeforeItReaches{P: p}
	coin := typedPred.Coin()

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}

	candlesticks := map[string][]types.Candlestick{}
	opStr := coin.Str
	it, err := (*s.mkt).GetIterator(coin, chartParams.startTimeISO8601(), false, chartParams.candlestickIntervalMinutes())
	if err != nil {
		return compiler.PredictionSummary{}, err
	}
	for i := 0; i < chartParams.candlestickCount; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return compiler.PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return compiler.PredictionSummary{
		PredictionType:           p.Type.String(),
		CandlestickMap:           candlesticks,
		Deadline:                 types.ISO8601(typedPred.Deadline().Format(time.RFC3339)),
		EndedAt:                  types.ISO8601(typedPred.EndTime().Format(time.RFC3339)),
		ErrorMarginRatio:         types.JsonFloat64(typedPred.ErrorMarginRatio()),
		WillReach:                typedPred.WillReach(),
		WillReachWithError:       typedPred.WillReachWithError(),
		BeforeItReaches:          typedPred.BeforeItReaches(),
		BeforeItReachesWithError: typedPred.BeforeItReachesWithError(),
		Coin:                     opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeTheFlippening(p types.Prediction) (compiler.PredictionSummary, error) {
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

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}

	candlesticks := map[string][]types.Candlestick{}
	opStr1 := marketCap1.Str
	it1, err := (*s.mkt).GetIterator(marketCap1, chartParams.startTimeISO8601(), false, 60*24)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}
	for i := 0; i < 30; i++ {
		candlestick, err := it1.NextCandlestick()
		if err != nil {
			return compiler.PredictionSummary{}, err
		}
		candlesticks[opStr1] = append(candlesticks[opStr1], candlestick)
	}

	opStr2 := marketCap2.Str
	it2, err := (*s.mkt).GetIterator(marketCap2, initialISO8601, false, 60*24)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}
	for i := 0; i < 30; i++ {
		candlestick, err := it2.NextCandlestick()
		if err != nil {
			return compiler.PredictionSummary{}, err
		}
		candlesticks[opStr2] = append(candlesticks[opStr2], candlestick)
	}

	return compiler.PredictionSummary{
		PredictionType: p.Type.String(),
		CandlestickMap: candlesticks,
		Operator:       operator,
		Deadline:       deadline,
		Coin:           opStr1,
		OtherCoin:      opStr2,
	}, nil
}

type candlestickChartParams struct {
	startTime           time.Time
	candlestickCount    int
	candlestickInterval time.Duration
}

func (p candlestickChartParams) startTimeISO8601() types.ISO8601 {
	return types.ISO8601(p.startTime.Format(time.RFC3339))
}

func (p candlestickChartParams) candlestickIntervalMinutes() int {
	return int(p.candlestickInterval / time.Minute)
}

func getCandlestickChartParams(p types.Prediction) (candlestickChartParams, error) {
	startTime, err := p.PostedAt.Time()
	if err != nil {
		return candlestickChartParams{}, err
	}

	// Compute endTime as the earliest of time.Now(), the lastTs when prediction finished & the prediction's deadline
	endTime := time.Now().UTC()
	if p.State.LastTs != 0 && p.State.Status == types.FINISHED {
		lastTime := time.Unix(int64(p.State.LastTs), 0)

		if lastTime.Before(endTime) {
			endTime = lastTime
		}
	}
	// TODO careful with this line! It's not guaranteed that all prediction types will have deadline here
	deadline := time.Unix(int64(p.Predict.Predict.Literal.ToTs), 0).UTC()
	if deadline.Before(endTime) {
		endTime = deadline
	}

	if endTime.Before(startTime) || endTime == startTime {
		return candlestickChartParams{}, errors.New("startTime is equal to endTime")
	}

	interval := endTime.Sub(startTime)

	switch {
	case interval < 5*24*time.Hour:
		return candlestickChartParams{startTime: endTime.Add(-30 * time.Hour), candlestickCount: 30, candlestickInterval: time.Hour}, nil
	default:
		return candlestickChartParams{startTime: endTime.Add(-30 * 24 * time.Hour), candlestickCount: 30, candlestickInterval: 24 * time.Hour}, nil
	}
}
