package smrunner

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/types"
	"github.com/marianogappa/signal-checker/common"
)

type PredRunner struct {
	prediction *types.Prediction
	tickers    map[string]map[string]types.TickIterator
	isInactive bool
}

var (
	errPredictionAtFinalStateAtCreation = errors.New("prediction is in final state at creation time")
)

func isDateEqual(ts1, ts2 int) bool {
	t1, t2 := time.Unix(int64(ts1), 0), time.Unix(int64(ts2), 0)
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func NewPredRunner(prediction *types.Prediction, m market.IMarket, nowTs int) (*PredRunner, []error) {
	errs := []error{}
	result := PredRunner{prediction: prediction, tickers: make(map[string]map[string]types.TickIterator)}

	predStateValue := prediction.Evaluate()
	if predStateValue != types.ONGOING_PRE_PREDICTION && predStateValue != types.ONGOING_PREDICTION {
		errs = append(errs, errPredictionAtFinalStateAtCreation)
		return nil, errs
	}

	var undecidedConditions []*types.Condition
	undecidedConditions = append(undecidedConditions, prediction.PrePredict.UndecidedConditions()...)
	undecidedConditions = append(undecidedConditions, prediction.Predict.UndecidedConditions()...)

	for _, condition := range undecidedConditions {
		if condition.Evaluate() != types.UNDECIDED {
			continue
		}
		startTs := calculateStartTs(condition)
		// N.B. For Messari API, current day is not available until day's end, so a special condition is needed.
		if startTs > nowTs || (condition.Operands[0].Type == types.MARKETCAP && isDateEqual(startTs, nowTs)) {
			continue
		}
		result.tickers[condition.Name] = map[string]types.TickIterator{}
		for _, operand := range condition.Operands {
			if operand.Type == types.COIN || operand.Type == types.MARKETCAP {
				ts := common.ISO8601(time.Unix(int64(startTs), 0).Format(time.RFC3339))
				ticker, err := m.GetTickIterator(operand, ts)
				if err != nil {
					errs = append(errs, err)
					return &result, errs
				}
				log.Printf("newPredRunner: created ticker for %v:%v:%v-%v at %v\n", operand.Type, operand.Provider, operand.BaseAsset, operand.QuoteAsset, ts)
				result.tickers[condition.Name][operand.Str] = ticker
			}
		}

	}
	if len(errs) > 0 {
		log.Printf("newPredRunner: errors creating new PredRunner: %v\n", errs)
	}
	return &result, errs
}

func calculateStartTs(c *types.Condition) int {
	if c.State.LastTs > c.FromTs {
		tickDurationSecs := 60
		if c.Operands[0].Type == types.MARKETCAP {
			tickDurationSecs = 60 * 60 * 24
		}
		return c.State.LastTs + tickDurationSecs
	}
	return c.FromTs
}

func (r *PredRunner) Run() []error {
	if r.isInactive {
		return nil
	}

	var undecidedConditions []*types.Condition
	switch r.prediction.Evaluate() {
	case types.ONGOING_PRE_PREDICTION:
		undecidedConditions = r.prediction.PrePredict.UndecidedConditions()
	case types.ONGOING_PREDICTION:
		undecidedConditions = r.prediction.Predict.UndecidedConditions()
	default:
		r.isInactive = true
		return nil
	}

	if !r.isAnyConditionHaveTickers(undecidedConditions) {
		r.isInactive = true
		return nil
	}

	errs := []error{}
	for _, cond := range undecidedConditions {
		tickers := r.tickers[cond.Name]
		if isAnyTickerOutOfTicks(tickers) {
			continue
		}

		var timestamp *int
		ticks := map[string]types.Tick{}
		for key, ticker := range tickers {
			tick, err := ticker.Next()
			if err != nil {
				if err != types.ErrOutOfTicks {
					errs = append(errs, err)
				}
				// log.Printf("For %v: for error after reading tick %v: %v\n", printer.NewPredictionPrettyPrinter(r.prediction).Default(), tick, err)
				// TODO: check if error is retryable before bailing (e.g. greylist, rate-limit)
				r.isInactive = true
				continue
			}
			log.Printf("For %v: read tick %v = %v\n", printer.NewPredictionPrettyPrinter(*r.prediction).Default(), time.Unix(int64(tick.Timestamp), 0).Format(time.RFC3339), tick.Value)
			// Timestamps must match on these ticks! Otherwise we're comparing apples & oranges!
			if timestamp == nil {
				timestamp = &tick.Timestamp
			} else if tick.Timestamp != *timestamp {
				// TODO: look into solving this problem: what if tickers don't synch on timestamps??? They should!
				err := fmt.Errorf("tickers have out-of-sync timestamps! %v vs %v", *timestamp, tick.Timestamp)
				errs = append(errs, err)
				r.isInactive = true
				continue
			}
			ticks[key] = tick
		}

		if !r.isInactive {
			err := cond.Run(ticks)
			if err != nil {
				errs = append(errs, err)
				r.isInactive = true
				return errs
			}
		}
		// log.Printf("Evaluating %v: %v", cond.Name, cond.Evaluate())
	}
	return errs
}

func (r *PredRunner) isAnyConditionHaveTickers(conds []*types.Condition) bool {
	for _, cond := range conds {
		if len(r.tickers[cond.Name]) > 0 {
			return true
		}
	}
	return false
}

func isAnyTickerOutOfTicks(ts map[string]types.TickIterator) bool {
	for _, ticker := range ts {
		if ticker.IsOutOfTicks() {
			return true
		}
	}
	return false
}
