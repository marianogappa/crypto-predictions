package daemon

import (
	"errors"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marianogappa/crypto-candles/candles/common"
	"github.com/marianogappa/crypto-candles/candles/iterator"
	"github.com/marianogappa/predictions/core"
)

// PredEvolver is the struct that evolves a prediction's state upon reading market data.
type PredEvolver struct {
	prediction *core.Prediction
	tickers    map[string]map[string]iterator.Iterator
}

var (
	errPredictionAtFinalStateAtCreation = errors.New("prediction is in final state at creation time")
)

// NewPredEvolver is the constructor for PredEvolver.
func NewPredEvolver(prediction *core.Prediction, m core.IMarket, nowTs int) (*PredEvolver, []error) {
	errs := []error{}
	result := PredEvolver{prediction: prediction, tickers: make(map[string]map[string]iterator.Iterator)}

	predStateValue := prediction.Evaluate()
	if predStateValue != core.ONGOINGPREPREDICTION && predStateValue != core.ONGOINGPREDICTION {
		errs = append(errs, errPredictionAtFinalStateAtCreation)
		return nil, errs
	}

	for _, condition := range prediction.UndecidedConditions() {
		startTime, startFromNext := calculateStartTs(condition)

		result.tickers[condition.Name] = map[string]iterator.Iterator{}
		for _, operand := range condition.NonNumberOperands() {
			ticker, err := m.Iterator(operand.ToMarketSource(), startTime, 1*time.Minute)
			if err != nil {
				errs = append(errs, err)
				return &result, errs
			}
			ticker.SetStartFromNext(startFromNext)

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

func (r *PredEvolver) runCondition(cond *core.Condition) error {
	var (
		lowestTicks  = map[string]core.Tick{}
		highestTicks = map[string]core.Tick{}
	)
	for key, ticker := range r.tickers[cond.Name] {
		candlestick, err := ticker.Next()
		if err != nil {
			return err
		}
		lowestTicks[key] = core.Tick{Timestamp: candlestick.Timestamp, Value: candlestick.LowestPrice}
		highestTicks[key] = core.Tick{Timestamp: candlestick.Timestamp, Value: candlestick.HighestPrice}
	}

	err := cond.Run(lowestTicks)
	if err != nil {
		return err
	}
	return cond.Run(highestTicks)
}

func (r *PredEvolver) actionableNonStuckUndecidedConditions(stuckConditions map[string]struct{}) []*core.Condition {
	conds := []*core.Condition{}
	for _, cond := range r.prediction.ActionableUndecidedConditions() {
		if _, ok := stuckConditions[cond.Name]; !ok {
			conds = append(conds, cond)
		}
	}
	return conds
}

func calculateStartTs(c *core.Condition) (time.Time, bool) {
	startTs := c.FromTs
	startFromNext := false

	if c.State.LastTs > 0 {
		startTs = c.State.LastTs
		startFromNext = true
	}

	return time.Unix(int64(startTs), 0).UTC(), startFromNext
}
