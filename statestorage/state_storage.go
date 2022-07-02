package statestorage

import (
	"database/sql"

	"github.com/marianogappa/predictions/types"
)

// StateStorage is the interface to the storage-layer. Currently the only implementation is Postgres.
// It might be wise to keep this interface, because Postgres might be convenient but it's a terrible choice for
// this engine's persistence needs.
type StateStorage interface {
	GetPredictions(filters types.APIFilters, orderBys []string, limit, offset int) ([]types.Prediction, error)
	GetAccounts(filters types.APIAccountFilters, orderBys []string, limit, offset int) ([]types.Account, error)
	// TODO: add interface contract
	UpsertPredictions([]*types.Prediction) ([]*types.Prediction, error)
	UpsertAccounts([]*types.Account) ([]*types.Account, error)
	LogPredictionStateValueChange(types.PredictionStateValueChange) error

	NonPendingPredictionInteractionExists(types.PredictionInteraction) (bool, error)
	InsertPredictionInteraction(types.PredictionInteraction) error
	GetPendingPredictionInteractions() ([]types.PredictionInteraction, error)
	UpdatePredictionInteractionStatus(types.PredictionInteraction) error

	PausePrediction(uuid string) error
	UnpausePrediction(uuid string) error
	HidePrediction(uuid string) error
	UnhidePrediction(uuid string) error
	DeletePrediction(uuid string) error
	UndeletePrediction(uuid string) error

	DB() *sql.DB
}
