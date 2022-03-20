package daemon

import (
	"errors"
	"log"
	"time"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/types"
)

type PredRunner struct {
	prediction *types.Prediction
	tickers    map[string]map[string]types.TickIterator
}

var (
	errPredictionAtFinalStateAtCreation = errors.New("prediction is in final state at creation time")
)

func NewPredRunner(prediction *types.Prediction, m market.IMarket, nowTs int) (*PredRunner, []error) {
	errs := []error{}
	result := PredRunner{prediction: prediction, tickers: make(map[string]map[string]types.TickIterator)}

	predStateValue := prediction.Evaluate()
	if predStateValue != types.ONGOING_PRE_PREDICTION && predStateValue != types.ONGOING_PREDICTION {
		errs = append(errs, errPredictionAtFinalStateAtCreation)
		return nil, errs
	}

	for _, condition := range prediction.UndecidedConditions() {
		startISO8601, startFromNext := calculateStartTs(condition)

		result.tickers[condition.Name] = map[string]types.TickIterator{}
		for _, operand := range condition.NonNumberOperands() {
			ticker, err := m.GetTickIterator(operand, startISO8601, startFromNext)
			if err != nil {
				errs = append(errs, err)
				return &result, errs
			}

			result.tickers[condition.Name][operand.Str] = ticker
		}
	}

	if len(errs) > 0 {
		log.Printf("newPredRunner: errors creating new PredRunner: %v\n", errs)
	}

	return &result, errs
}

func (r *PredRunner) Run() []error {
	var (
		errs            = []error{}
		stuckConditions = map[string]struct{}{}
		conds           = r.actionableNonStuckUndecidedConditions(stuckConditions)
	)
	for len(conds) > 0 {
		for _, cond := range conds {
			if err := r.runCondition(cond); err != nil {
				stuckConditions[cond.Name] = struct{}{}
				if err != types.ErrOutOfTicks && err != types.ErrNoNewTicksYet {
					errs = append(errs, err)
				}
			}
		}
		conds = r.actionableNonStuckUndecidedConditions(stuckConditions)
	}

	return errs
}

func (r *PredRunner) runCondition(cond *types.Condition) error {
	ticks := map[string]types.Tick{}
	for key, ticker := range r.tickers[cond.Name] {
		tick, err := ticker.Next()
		if err != nil {
			return err
		}
		ticks[key] = tick
	}

	return cond.Run(ticks)
}

func (r *PredRunner) actionableNonStuckUndecidedConditions(stuckConditions map[string]struct{}) []*types.Condition {
	conds := []*types.Condition{}
	for _, cond := range r.prediction.ActionableUndecidedConditions() {
		if _, ok := stuckConditions[cond.Name]; !ok {
			conds = append(conds, cond)
		}
	}
	return conds
}

func calculateStartTs(c *types.Condition) (types.ISO8601, bool) {
	startTs := c.FromTs
	startFromNext := false

	if c.State.LastTs > 0 {
		startTs = c.State.LastTs
		startFromNext = true
	}

	return types.ISO8601(time.Unix(int64(startTs), 0).Format(time.RFC3339)), startFromNext
}
