package daemon

import (
	"errors"

	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
)

var (
	ErrNoMorePredictions = errors.New("no more predictions")
)

type EvolvablePredictionsScanner struct {
	store       statestorage.StateStorage
	predictions []types.Prediction
	lastUUID    string
	Error       error
}

func NewEvolvablePredictionsScanner(store statestorage.StateStorage) *EvolvablePredictionsScanner {
	return &EvolvablePredictionsScanner{store: store}
}

func (it *EvolvablePredictionsScanner) query() ([]types.Prediction, error) {
	filters := types.APIFilters{
		PredictionStateValues: []string{
			types.ONGOING_PRE_PREDICTION.String(),
			types.ONGOING_PREDICTION.String(),
		},
		Paused:  pBool(false),
		Deleted: pBool(false),
	}

	if it.lastUUID != "" {
		filters.GreaterThanUUID = it.lastUUID
	}

	preds, err := it.store.GetPredictions(
		filters,
		[]string{types.UUID_ASC.String()},
		10, 0,
	)
	if err != nil {
		return nil, err
	}
	if len(preds) > 0 {
		it.lastUUID = preds[len(preds)-1].UUID
	}

	return preds, nil
}

func (it *EvolvablePredictionsScanner) Scan(prediction *types.Prediction) bool {
	it.Error = nil
	if len(it.predictions) == 0 {
		var err error
		it.predictions, err = it.query()
		if err != nil {
			it.Error = err
			*prediction = types.Prediction{}
			return false
		}
	}

	if len(it.predictions) > 0 {
		*prediction = it.predictions[0]
		it.predictions = it.predictions[1:]
		return true
	}

	return false
}

func pBool(b bool) *bool { return &b }