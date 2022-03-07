package statestorage

import "github.com/marianogappa/predictions/types"

type StateStorage interface {
	GetPredictions(filters types.APIFilters, orderBys []string) ([]types.Prediction, error)
	UpsertPredictions(map[string]types.Prediction) error
}
