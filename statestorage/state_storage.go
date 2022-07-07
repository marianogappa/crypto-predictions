package statestorage

import (
	"database/sql"

	"github.com/marianogappa/predictions/core"
)

// StateStorage is the interface to the storage-layer. Currently the only implementation is Postgres.
// It might be wise to keep this interface, because Postgres might be convenient but it's a terrible choice for
// this engine's persistence needs.
type StateStorage interface {
	GetPredictions(filters core.APIFilters, orderBys []string, limit, offset int) ([]core.Prediction, error)
	GetAccounts(filters core.APIAccountFilters, orderBys []string, limit, offset int) ([]core.Account, error)
	// TODO: add interface contract
	UpsertPredictions([]*core.Prediction) ([]*core.Prediction, error)
	UpsertAccounts([]*core.Account) ([]*core.Account, error)
	LogPredictionStateValueChange(core.PredictionStateValueChange) error

	NonPendingPredictionInteractionExists(core.PredictionInteraction) (bool, error)
	InsertPredictionInteraction(core.PredictionInteraction) error
	GetPendingPredictionInteractions() ([]core.PredictionInteraction, error)
	UpdatePredictionInteractionStatus(core.PredictionInteraction) error

	PausePrediction(uuid string) error
	UnpausePrediction(uuid string) error
	HidePrediction(uuid string) error
	UnhidePrediction(uuid string) error
	DeletePrediction(uuid string) error
	UndeletePrediction(uuid string) error

	DB() *sql.DB
}
