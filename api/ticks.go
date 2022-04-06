package api

import (
	"time"

	"github.com/marianogappa/predictions/types"
)

type predictionSummary struct {
	TickMap  map[string][]types.Tick `json:"tickMap"`
	Coin     string                  `json:"coin"`
	Goal     types.JsonFloat64       `json:"goal"`
	Operator string                  `json:"operator"`
	Deadline types.ISO8601           `json:"deadline"`
}

func (a *API) BuildPredictionMarketSummary(p types.Prediction) (predictionSummary, error) {
	switch p.Type {
	case types.PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE:
		return a.predictionTypeCoinOperatorFloatDeadline(p)
	}
	return predictionSummary{}, nil
}

func (a *API) predictionTypeCoinOperatorFloatDeadline(p types.Prediction) (predictionSummary, error) {
	postedAt, err := p.PostedAt.Time()
	if err != nil {
		return predictionSummary{}, err
	}
	tmFirstTick := postedAt.Add(-101 * time.Minute)
	p.Evaluate().IsFinal()

	coin := p.Predict.Predict.Literal.Operands[0]
	operator := p.Predict.Predict.Literal.Operator
	goal := p.Predict.Predict.Literal.Operands[1].Number
	initialISO8601 := types.ISO8601(tmFirstTick.Format(time.RFC3339))
	deadline := types.ISO8601(time.Unix(int64(p.Predict.Predict.Literal.ToTs), 0).UTC().Format(time.RFC3339))

	ticks := map[string][]types.Tick{}
	opStr := coin.Str
	it, err := a.mkt.GetTickIterator(coin, initialISO8601, false)
	if err != nil {
		return predictionSummary{}, err
	}
	for i := 0; i < 100; i++ {
		tick, err := it.Next()
		if err != nil {
			return predictionSummary{}, err
		}
		ticks[opStr] = append(ticks[opStr], tick)
	}

	return predictionSummary{
		TickMap:  ticks,
		Operator: operator,
		Goal:     goal,
		Deadline: deadline,
	}, nil
}
