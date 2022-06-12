package compiler

import (
	"errors"
	"time"

	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/jsonschema-go"
)

// PredictionSummary contains all necessary information about the prediction to make a candlestick chart of it.
type PredictionSummary struct {
	// Only in "PredictionTheFlippening" type
	OtherCoin string `json:"otherCoin,omitempty"`

	// Only in "PredictionTypeCoinOperatorFloatDeadline" type
	Goal                                    types.JsonFloat64 `json:"goal,omitempty"`
	GoalWithError                           types.JsonFloat64 `json:"goalWithError,omitempty"`
	EndedAtTruncatedDueToResultInvalidation types.ISO8601     `json:"endedAtTruncatedDueToResultInvalidation,omitempty"`

	// Only in "PredictionWillRange type"
	RangeLow           types.JsonFloat64 `json:"rangeLow,omitempty"`
	RangeLowWithError  types.JsonFloat64 `json:"rangeLowWithError,omitempty"`
	RangeHigh          types.JsonFloat64 `json:"rangeHigh,omitempty"`
	RangeHighWithError types.JsonFloat64 `json:"rangeHighWithError,omitempty"`

	// Only in "PredictionWillReachBeforeItReaches type"
	WillReach                types.JsonFloat64 `json:"willReach,omitempty"`
	WillReachWithError       types.JsonFloat64 `json:"willReachWithError,omitempty"`
	BeforeItReaches          types.JsonFloat64 `json:"beforeItReaches,omitempty"`
	BeforeItReachesWithError types.JsonFloat64 `json:"beforeItReachesWithError,omitempty"`

	// In all prediction types
	CandlestickMap   map[string][]types.Candlestick `json:"candlestickMap,omitempty"`
	Coin             string                         `json:"coin,omitempty"`
	ErrorMarginRatio types.JsonFloat64              `json:"errorMarginRatio,omitempty"`
	Operator         string                         `json:"operator,omitempty"`
	Deadline         types.ISO8601                  `json:"deadline,omitempty"`
	EndedAt          types.ISO8601                  `json:"endedAt,omitempty"`
	PredictionType   string                         `json:"predictionType,omitempty"`
}

// PrepareJSONSchema provides an example of the structure for Swagger docs
func (PredictionSummary) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithExamples(PredictionSummary{
		CandlestickMap: map[string][]types.Candlestick{
			"BINANCE:COIN:BTC-USDT": {
				{Timestamp: 1651161957, OpenPrice: 39000, HighestPrice: 39500, LowestPrice: 39000, ClosePrice: 39050},
				{Timestamp: 1651162017, OpenPrice: 39500, HighestPrice: 39550, LowestPrice: 39200, ClosePrice: 39020},
			},
		},
		Coin:                                    "BINANCE:COIN:BTC-USDT",
		Goal:                                    45000,
		GoalWithError:                           43650,
		ErrorMarginRatio:                        0.03,
		Operator:                                ">=",
		Deadline:                                "2022-06-24T07:51:06Z",
		EndedAt:                                 "2022-06-24T07:51:06Z",
		EndedAtTruncatedDueToResultInvalidation: "2022-06-23T00:00:00Z",
		PredictionType:                          "PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE",
	})

	return nil
}

// BuildPredictionMarketSummary builds a PredictionSummary from a prediction. It uses a market to get candlesticks.
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
	typedPred := types.PredictionTypeCoinOperatorFloatDeadline{P: p}

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return PredictionSummary{}, err
	}

	candlesticks := map[string][]types.Candlestick{}
	opStr := typedPred.Coin().Str
	it, err := (*s.mkt).GetIterator(typedPred.Coin(), chartParams.startTimeISO8601(), false, chartParams.candlestickIntervalMinutes())
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < chartParams.candlestickCount; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return PredictionSummary{
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

func (s PredictionSerializer) predictionTypeCoinWillRange(p types.Prediction) (PredictionSummary, error) {
	typedPred := types.PredictionTypeCoinWillRange{P: p}
	coin := typedPred.Coin()

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return PredictionSummary{}, err
	}

	candlesticks := map[string][]types.Candlestick{}
	opStr := coin.Str
	it, err := (*s.mkt).GetIterator(coin, chartParams.startTimeISO8601(), false, chartParams.candlestickIntervalMinutes())
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < chartParams.candlestickCount; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return PredictionSummary{
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

func (s PredictionSerializer) predictionTypeCoinWillReachBeforeItReaches(p types.Prediction) (PredictionSummary, error) {
	typedPred := types.PredictionTypeCoinWillReachBeforeItReaches{P: p}
	coin := typedPred.Coin()

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return PredictionSummary{}, err
	}

	candlesticks := map[string][]types.Candlestick{}
	opStr := coin.Str
	it, err := (*s.mkt).GetIterator(coin, chartParams.startTimeISO8601(), false, chartParams.candlestickIntervalMinutes())
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < chartParams.candlestickCount; i++ {
		candlestick, err := it.NextCandlestick()
		if err != nil {
			return PredictionSummary{}, err
		}
		candlesticks[opStr] = append(candlesticks[opStr], candlestick)
	}

	return PredictionSummary{
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

	chartParams, err := getCandlestickChartParams(p)
	if err != nil {
		return PredictionSummary{}, err
	}

	candlesticks := map[string][]types.Candlestick{}
	opStr1 := marketCap1.Str
	it1, err := (*s.mkt).GetIterator(marketCap1, chartParams.startTimeISO8601(), false, 60*24)
	if err != nil {
		return PredictionSummary{}, err
	}
	for i := 0; i < 30; i++ {
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
	for i := 0; i < 30; i++ {
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
