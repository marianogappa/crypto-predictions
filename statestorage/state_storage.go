package statestorage

import "github.com/marianogappa/predictions/types"

type StateStorage interface {
	GetPredictions([]types.PredictionStateValue) (map[string]types.Prediction, error)
	UpsertPredictions(map[string]types.Prediction) error
}
