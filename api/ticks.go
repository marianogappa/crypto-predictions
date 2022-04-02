package api

import (
	"time"

	"github.com/marianogappa/predictions/types"
)

type predictionSummary struct {
	TickMap map[string][]types.Tick `json:"tickMap"`
}

func (a *API) BuildPredictionMarketSummary(p types.Prediction) (predictionSummary, error) {
	postedAt, err := p.PostedAt.Time()
	if err != nil {
		return predictionSummary{}, err
	}
	firstMinutelyTick := postedAt.Add(-101 * time.Minute)
	// firstDailyTick := postedAt.Add(-15*time.Hour*24)
	initialISO8601 := types.ISO8601(firstMinutelyTick.Format(time.RFC3339))
	p.Evaluate().IsFinal()

	ticks := map[string][]types.Tick{}
	for _, cond := range p.Given {
		for _, operand := range cond.NonNumberOperands() {
			opStr := operand.Str
			it, err := a.mkt.GetTickIterator(operand, initialISO8601, false)
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
		}
	}

	return predictionSummary{ticks}, nil
}
