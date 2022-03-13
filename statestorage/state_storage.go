package statestorage

import "github.com/marianogappa/predictions/types"

type StateStorage interface {
	GetPredictions(filters types.APIFilters, orderBys []string) ([]types.Prediction, error)
	// TODO: add interface contract
	UpsertPredictions([]*types.Prediction) ([]*types.Prediction, error)
}
