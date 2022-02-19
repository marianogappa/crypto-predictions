package smrunner

import (
	"fmt"
	"log"
	"time"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/types"
	"github.com/marianogappa/signal-checker/common"
)

type predRunner struct {
	prediction types.Prediction
	tickers    map[string]map[string]*market.TickIterator
	isInactive bool
}

func newPredRunner(prediction types.Prediction, m market.Market, nowTs int) (*predRunner, []error) {
	errs := []error{}
	result := predRunner{prediction: prediction, tickers: make(map[string]map[string]*market.TickIterator)}
	for name, condition := range prediction.Define {
		if condition.Evaluate() != types.UNDECIDED {
			continue
		}
		startTs := condition.FromTs
		if condition.State.LastTs > condition.FromTs {
			startTs = condition.State.LastTs
		}
		if startTs > nowTs {
			continue
		}
		result.tickers[name] = map[string]*market.TickIterator{}
		for _, operand := range condition.Operands {
			if operand.Type == types.COIN || operand.Type == types.MARKETCAP {
				ts := common.ISO8601(time.Unix(int64(startTs), 0).Format(time.RFC3339))
				ticker, err := m.GetTickIterator(operand, ts)
				if err != nil {
					errs = append(errs, err)
					return &result, errs
				}
				log.Printf("newPredRunner: created ticker for %v:%v:%v-%v at %v\n", operand.Type, operand.Provider, operand.BaseAsset, operand.QuoteAsset, ts)
				result.tickers[name][operand.Str] = ticker
			}
		}

	}
	if len(errs) > 0 {
		log.Printf("newPredRunner: errors creating new predRunner: %v\n", errs)
	}
	return &result, errs
}

func (r *predRunner) Run() []error {
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

	log.Printf("predRunner.Run: prediction: %+v, status: %v, value: %v, undecidedConditions: %v\n", r.prediction, r.prediction.State.Status, r.prediction.State.Value, len(undecidedConditions))
	errs := []error{}
	for _, cond := range undecidedConditions {
		tickers := r.tickers[cond.Name]
		if isAnyTickerOutOfTicks(tickers) {
			continue
		}

		var timestamp *int
		ticks := map[string]*common.Tick{}
		for key, ticker := range tickers {
			tick, err := ticker.Next()
			if err != nil {
				if err != common.ErrOutOfCandlesticks {
					errs = append(errs, err)
				}
				// TODO: check if error is retryable before bailing (e.g. greylist, rate-limit)
				r.isInactive = true
				continue
			}
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
			ticks[key] = &tick
		}

		cond.Run(ticks)
		log.Println(cond.Evaluate())
	}
	return errs
}

func isAnyTickerOutOfTicks(ts map[string]*market.TickIterator) bool {
	for _, ticker := range ts {
		if ticker.IsOutOfTicks {
			return true
		}
	}
	return false
}
