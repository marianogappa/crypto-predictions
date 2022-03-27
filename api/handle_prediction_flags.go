package api

import (
	"fmt"

	"github.com/marianogappa/predictions/types"
)

type uuidBody struct {
	UUID string `json:"uuid"`
}

func (a *API) handlePause(uuid uuidBody) (response, error) {
	ps, err := a.store.GetPredictions(types.APIFilters{UUIDs: []string{uuid.UUID}}, nil, 0, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrPredictionNotFound, err)
	}
	if len(ps) == 0 {
		return response{}, fmt.Errorf("%w", ErrPredictionNotFound)
	}

	if len(ps) != 1 {
		return response{}, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps))
	}

	pred := ps[0]
	if err := a.store.PausePrediction(pred.UUID); err != nil {
		return response{}, fmt.Errorf("%w: error pausing prediction", ErrFailedToCompilePrediction)
	}
	return response{stored: pBool(true)}, nil
}

func (a *API) handleUnpause(uuid uuidBody) (response, error) {
	ps, err := a.store.GetPredictions(types.APIFilters{UUIDs: []string{uuid.UUID}}, nil, 0, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrPredictionNotFound, err)
	}
	if len(ps) == 0 {
		return response{}, fmt.Errorf("%w", ErrPredictionNotFound)
	}

	if len(ps) != 1 {
		return response{}, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps))
	}

	pred := ps[0]
	if err := a.store.UnpausePrediction(pred.UUID); err != nil {
		return response{}, fmt.Errorf("%w: error pausing prediction", ErrFailedToCompilePrediction)
	}
	return response{stored: pBool(true)}, nil
}

func (a *API) handleHide(uuid uuidBody) (response, error) {
	ps, err := a.store.GetPredictions(types.APIFilters{UUIDs: []string{uuid.UUID}}, nil, 0, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrPredictionNotFound, err)
	}
	if len(ps) == 0 {
		return response{}, fmt.Errorf("%w", ErrPredictionNotFound)
	}

	if len(ps) != 1 {
		return response{}, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps))
	}

	pred := ps[0]
	if err := a.store.HidePrediction(pred.UUID); err != nil {
		return response{}, fmt.Errorf("%w: error pausing prediction", ErrFailedToCompilePrediction)
	}
	return response{stored: pBool(true)}, nil
}

func (a *API) handleUnhide(uuid uuidBody) (response, error) {
	ps, err := a.store.GetPredictions(types.APIFilters{UUIDs: []string{uuid.UUID}}, nil, 0, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrPredictionNotFound, err)
	}
	if len(ps) == 0 {
		return response{}, fmt.Errorf("%w", ErrPredictionNotFound)
	}

	if len(ps) != 1 {
		return response{}, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps))
	}

	pred := ps[0]
	if err := a.store.HidePrediction(pred.UUID); err != nil {
		return response{}, fmt.Errorf("%w: error pausing prediction", ErrFailedToCompilePrediction)
	}
	return response{stored: pBool(true)}, nil
}

func (a *API) handleDelete(uuid uuidBody) (response, error) {
	ps, err := a.store.GetPredictions(types.APIFilters{UUIDs: []string{uuid.UUID}}, nil, 0, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrPredictionNotFound, err)
	}
	if len(ps) == 0 {
		return response{}, fmt.Errorf("%w", ErrPredictionNotFound)
	}

	if len(ps) != 1 {
		return response{}, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps))
	}

	pred := ps[0]
	if err := a.store.DeletePrediction(pred.UUID); err != nil {
		return response{}, fmt.Errorf("%w: error pausing prediction", ErrFailedToCompilePrediction)
	}
	return response{stored: pBool(true)}, nil
}

func (a *API) handleUndelete(uuid uuidBody) (response, error) {
	ps, err := a.store.GetPredictions(types.APIFilters{UUIDs: []string{uuid.UUID}}, nil, 0, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrPredictionNotFound, err)
	}
	if len(ps) == 0 {
		return response{}, fmt.Errorf("%w", ErrPredictionNotFound)
	}

	if len(ps) != 1 {
		return response{}, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps))
	}

	pred := ps[0]
	if err := a.store.UndeletePrediction(pred.UUID); err != nil {
		return response{}, fmt.Errorf("%w: error pausing prediction", ErrFailedToCompilePrediction)
	}
	return response{stored: pBool(true)}, nil
}

func (a *API) handleRefetchAccount(uuid uuidBody) (response, error) {
	ps, err := a.store.GetPredictions(types.APIFilters{UUIDs: []string{uuid.UUID}}, nil, 0, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrPredictionNotFound, err)
	}
	if len(ps) == 0 {
		return response{}, fmt.Errorf("%w", ErrPredictionNotFound)
	}
	if len(ps) != 1 {
		return response{}, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps))
	}

	pred := ps[0]

	metadata, err := a.mFetcher.Fetch(pred.PostUrl)
	if err != nil {
		return response{}, fmt.Errorf("%w: error fetching metadata for url: %v", ErrFailedToCompilePrediction, pred.PostUrl)
	}

	if _, err := a.store.UpsertAccounts([]*types.Account{&metadata.Author}); err != nil {
		return response{}, fmt.Errorf("%w: error storing account: %v", ErrStorageErrorStoringAccount, err)
	}

	return response{stored: pBool(true)}, nil
}
