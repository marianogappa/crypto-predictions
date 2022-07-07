package serializer

import (
	"errors"
	"time"

	"github.com/marianogappa/crypto-candles/candles/common"
	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/core"
)

// BuildPredictionMarketSummary builds a PredictionSummary from a prediction. It uses a market to get candlesticks.
func (s PredictionSerializer) BuildPredictionMarketSummary(p core.Prediction) (compiler.PredictionSummary, error) {
	switch p.Type {
	case core.PredictionTypeCoinOperatorFloatDeadline:
		return s.predictionTypeCoinOperatorFloatDeadline(p)
	case core.PredictionTypeCoinWillReachInvalidatedIfItReaches:
		return s.predictionTypeCoinWillReachInvalidatedIfItReaches(p)
	case core.PredictionTypeCoinWillRange:
		return s.predictionTypeCoinWillRange(p)
	case core.PredictionTypeCoinWillReachBeforeItReaches:
		return s.predictionTypeCoinWillReachBeforeItReaches(p)
	}
	return compiler.PredictionSummary{}, nil
}

func (s PredictionSerializer) predictionTypeCoinOperatorFloatDeadline(p core.Prediction) (compiler.PredictionSummary, error) {
	typedPred := core.PredictionTypeCoinOperatorFloatDeadlineWrapper{P: p}

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}

	candlesticks := map[string][]common.Candlestick{}
	opStr := typedPred.Coin().Str
	it, err := (*s.mkt).GetIterator(typedPred.Coin().ToMarketSource(), chartParams.startTime, false, chartParams.candlestickIntervalMinutes())
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
		ErrorMarginRatio:                        core.JSONFloat64(typedPred.ErrorMarginRatio()),
		Deadline:                                core.ISO8601(typedPred.Deadline().Format(time.RFC3339)),
		EndedAt:                                 core.ISO8601(typedPred.EndTime().Format(time.RFC3339)),
		EndedAtTruncatedDueToResultInvalidation: core.ISO8601(typedPred.EndTimeTruncatedDueToResultInvalidation(candlesticks[opStr]).Format(time.RFC3339)),
		Coin:                                    opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeCoinWillReachInvalidatedIfItReaches(p core.Prediction) (compiler.PredictionSummary, error) {
	typedPred := core.PredictionTypeCoinWillReachInvalidatedIfItReachesWrapper{P: p}

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}

	candlesticks := map[string][]common.Candlestick{}
	opStr := typedPred.Coin().Str
	it, err := (*s.mkt).GetIterator(typedPred.Coin().ToMarketSource(), chartParams.startTime, false, chartParams.candlestickIntervalMinutes())
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
		ErrorMarginRatio:                        core.JSONFloat64(typedPred.ErrorMarginRatio()),
		Deadline:                                core.ISO8601(typedPred.Deadline().Format(time.RFC3339)),
		EndedAt:                                 core.ISO8601(typedPred.EndTime().Format(time.RFC3339)),
		EndedAtTruncatedDueToResultInvalidation: core.ISO8601(typedPred.EndTimeTruncatedDueToResultInvalidation(candlesticks[opStr]).Format(time.RFC3339)),
		Coin:                                    opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeCoinWillRange(p core.Prediction) (compiler.PredictionSummary, error) {
	typedPred := core.PredictionTypeCoinWillRangeWrapper{P: p}
	coin := typedPred.Coin()

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}

	candlesticks := map[string][]common.Candlestick{}
	opStr := coin.Str
	it, err := (*s.mkt).GetIterator(coin.ToMarketSource(), chartParams.startTime, false, chartParams.candlestickIntervalMinutes())
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
		Deadline:           core.ISO8601(typedPred.Deadline().Format(time.RFC3339)),
		EndedAt:            core.ISO8601(typedPred.EndTime().Format(time.RFC3339)),
		ErrorMarginRatio:   core.JSONFloat64(typedPred.ErrorMarginRatio()),
		RangeLow:           typedPred.RangeLow(),
		RangeLowWithError:  typedPred.RangeLowWithError(),
		RangeHigh:          typedPred.RangeHigh(),
		RangeHighWithError: typedPred.RangeHighWithError(),
		Coin:               opStr,
	}, nil
}

func (s PredictionSerializer) predictionTypeCoinWillReachBeforeItReaches(p core.Prediction) (compiler.PredictionSummary, error) {
	typedPred := core.PredictionTypeCoinWillReachBeforeItReachesWrapper{P: p}
	coin := typedPred.Coin()

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return compiler.PredictionSummary{}, err
	}

	candlesticks := map[string][]common.Candlestick{}
	opStr := coin.Str
	it, err := (*s.mkt).GetIterator(coin.ToMarketSource(), chartParams.startTime, false, chartParams.candlestickIntervalMinutes())
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
		Deadline:                 core.ISO8601(typedPred.Deadline().Format(time.RFC3339)),
		EndedAt:                  core.ISO8601(typedPred.EndTime().Format(time.RFC3339)),
		ErrorMarginRatio:         core.JSONFloat64(typedPred.ErrorMarginRatio()),
		WillReach:                typedPred.WillReach(),
		WillReachWithError:       typedPred.WillReachWithError(),
		BeforeItReaches:          typedPred.BeforeItReaches(),
		BeforeItReachesWithError: typedPred.BeforeItReachesWithError(),
		Coin:                     opStr,
	}, nil
}

type candlestickChartParams struct {
	startTime           time.Time
	candlestickCount    int
	candlestickInterval time.Duration
}

func (p candlestickChartParams) candlestickIntervalMinutes() int {
	return int(p.candlestickInterval / time.Minute)
}

func getCandlestickChartParams(p core.Prediction) (candlestickChartParams, error) {
	startTime, err := p.PostedAt.Time()
	if err != nil {
		return candlestickChartParams{}, err
	}

	// Compute endTime as the earliest of time.Now(), the lastTs when prediction finished & the prediction's deadline
	endTime := time.Now().UTC()
	if p.State.LastTs != 0 && p.State.Status == core.FINISHED {
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
