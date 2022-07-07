package daemon

import (
	"errors"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

// PredEvolver is the struct that evolves a prediction's state upon reading market data.
type PredEvolver struct {
	prediction *types.Prediction
	tickers    map[string]map[string]common.Iterator
}

var (
	errPredictionAtFinalStateAtCreation = errors.New("prediction is in final state at creation time")
)

// NewPredEvolver is the constructor for PredEvolver.
func NewPredEvolver(prediction *types.Prediction, m market.IMarket, nowTs int) (*PredEvolver, []error) {
	errs := []error{}
	result := PredEvolver{prediction: prediction, tickers: make(map[string]map[string]common.Iterator)}

	predStateValue := prediction.Evaluate()
	if predStateValue != types.ONGOINGPREPREDICTION && predStateValue != types.ONGOINGPREDICTION {
		errs = append(errs, errPredictionAtFinalStateAtCreation)
		return nil, errs
	}

	for _, condition := range prediction.UndecidedConditions() {
		startTime, startFromNext := calculateStartTs(condition)

		result.tickers[condition.Name] = map[string]common.Iterator{}
		for _, operand := range condition.NonNumberOperands() {
			ticker, err := m.GetIterator(operand.ToMarketSource(), startTime, startFromNext, 1)
			if err != nil {
				errs = append(errs, err)
				return &result, errs
			}

			result.tickers[condition.Name][operand.Str] = ticker
		}
	}

	if len(errs) > 0 {
		log.Info().Msgf("newPredEvolver: errors creating new PredRunner: %v\n", errs)
	}

	return &result, errs
}

// Run evolves the prediction until it hits an error, or there's no more recent market data, or the prediction finishes.
func (r *PredEvolver) Run(once bool) []error {
	var (
		errs            = []error{}
		stuckConditions = map[string]struct{}{}
		conds           = r.actionableNonStuckUndecidedConditions(stuckConditions)
	)
	for len(conds) > 0 {
		for _, cond := range conds {
			if err := r.runCondition(cond); err != nil {
				stuckConditions[cond.Name] = struct{}{}
				if err != common.ErrOutOfTicks && err != common.ErrNoNewTicksYet {
					errs = append(errs, err)
				}
			}
		}
		if once {
			break
		}
		conds = r.actionableNonStuckUndecidedConditions(stuckConditions)
	}

	return errs
}

func (r *PredEvolver) runCondition(cond *types.Condition) error {
	ticks := map[string]common.Tick{}
	for key, ticker := range r.tickers[cond.Name] {
		tick, err := ticker.NextTick()
		if err != nil {
			return err
		}
		ticks[key] = tick
	}

	return cond.Run(ticks)
}

func (r *PredEvolver) actionableNonStuckUndecidedConditions(stuckConditions map[string]struct{}) []*types.Condition {
	conds := []*types.Condition{}
	for _, cond := range r.prediction.ActionableUndecidedConditions() {
		if _, ok := stuckConditions[cond.Name]; !ok {
			conds = append(conds, cond)
		}
	}
	return conds
}

func calculateStartTs(c *types.Condition) (time.Time, bool) {
	startTs := c.FromTs
	startFromNext := false

	if c.State.LastTs > 0 {
		startTs = c.State.LastTs
		startFromNext = true
	}

	return time.Unix(int64(startTs), 0).UTC(), startFromNext
}
