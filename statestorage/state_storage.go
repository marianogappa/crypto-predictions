package statestorage

import "github.com/marianogappa/predictions/types"

type StateStorage interface {
	GetPredictions(filters types.APIFilters, orderBys []string, limit, offset int) ([]types.Prediction, error)
	GetAccounts(filters types.APIAccountFilters, orderBys []string, limit, offset int) ([]types.Account, error)
	// TODO: add interface contract
	UpsertPredictions([]*types.Prediction) ([]*types.Prediction, error)
	UpsertAccounts([]*types.Account) ([]*types.Account, error)
	LogPredictionStateValueChange(types.PredictionStateValueChange) error
	PausePrediction(uuid string) error
	UnpausePrediction(uuid string) error
	HidePrediction(uuid string) error
	UnhidePrediction(uuid string) error
	DeletePrediction(uuid string) error
	UndeletePrediction(uuid string) error
}
