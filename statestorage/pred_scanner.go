package statestorage

import (
	"github.com/marianogappa/predictions/types"
)

// PredictionScanner is a storage-layer Prediction iterator that follows the Scanner interface.
type PredictionScanner struct {
	Error error

	store       StateStorage
	predictions []types.Prediction
	lastUUID    string
	filters     types.APIFilters
	limit       int
}

// NewEvolvablePredictionsScanner constructs a PredictionScanner that only retrieves predictions that are not in a
// final state, nor paused nor deleted.
func NewEvolvablePredictionsScanner(store StateStorage) *PredictionScanner {
	return newPredictionScanner(store, filterEvolvable, 100)
}

// NewAllPredictionsScanner constructs a PredictionScanner that retrieves all available predictions in the
// storage-layer, even if they are final, paused or deleted.
func NewAllPredictionsScanner(store StateStorage) *PredictionScanner {
	return newPredictionScanner(store, filterAll, 100)
}

var (
	filterEvolvable = types.APIFilters{
		PredictionStateValues: []string{
			types.ONGOING_PRE_PREDICTION.String(),
			types.ONGOING_PREDICTION.String(),
		},
		Paused:  pBool(false),
		Deleted: pBool(false),
	}

	filterAll = types.APIFilters{
		Paused:  nil,
		Deleted: nil,
		Hidden:  nil,
	}
)

func newPredictionScanner(store StateStorage, filters types.APIFilters, batchSize int) *PredictionScanner {
	if batchSize == 0 {
		batchSize = 100
	}
	return &PredictionScanner{store: store, filters: filters, limit: batchSize}
}

func (it *PredictionScanner) query() ([]types.Prediction, error) {
	filters := it.filters

	// Note that the iterator strategy is WHERE uuid > lastUUID.
	if it.lastUUID != "" {
		filters.GreaterThanUUID = it.lastUUID
	}

	preds, err := it.store.GetPredictions(
		filters,
		[]string{types.UUID_ASC.String()},
		it.limit, 0,
	)
	if err != nil {
		return nil, err
	}
	if len(preds) > 0 {
		it.lastUUID = preds[len(preds)-1].UUID
	}

	return preds, nil
}

// Scan retrieves the next prediction and stores it within the supplied *types.Prediction, and returns false when
// there are no predictions left or there is an error. To differentiate these cases, inspect the Error property.
func (it *PredictionScanner) Scan(prediction *types.Prediction) bool {
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
