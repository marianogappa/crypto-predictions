package statestorage

import (
	"github.com/google/uuid"
	"github.com/marianogappa/predictions/types"
)

// TODO implement paused/deleted/hidden filters

// MemoryStateStorage is the memory-level StateStorage implementation.
type MemoryStateStorage struct {
	Predictions                 map[string]types.Prediction
	PredictionStateValueChanges map[string]types.PredictionStateValueChange
	Accounts                    map[string]types.Account
}

// NewMemoryStateStorage constructs a MemoryStateStorage.
func NewMemoryStateStorage() MemoryStateStorage {
	return MemoryStateStorage{
		Predictions:                 map[string]types.Prediction{},
		PredictionStateValueChanges: map[string]types.PredictionStateValueChange{},
		Accounts:                    map[string]types.Account{},
	}
}

// GetPredictions SELECTs predictions from the memory store.
func (s MemoryStateStorage) GetPredictions(filters types.APIFilters, orderBys []string, limit, offset int) ([]types.Prediction, error) {
	authors := sliceToMap(filters.AuthorHandles)
	stateStatuses := sliceToMap(filters.PredictionStateStatus)
	stateValues := sliceToMap(filters.PredictionStateValues)
	uuids := sliceToMap(filters.UUIDs)

	res := []types.Prediction{}
	for _, v := range s.Predictions {
		if _, ok := authors[v.PostAuthor]; !ok && len(authors) > 0 {
			continue
		}
		if _, ok := stateStatuses[v.State.Status.String()]; !ok && len(stateStatuses) > 0 {
			continue
		}
		if _, ok := stateValues[v.State.Value.String()]; !ok && len(stateValues) > 0 {
			continue
		}
		if _, ok := uuids[v.UUID]; !ok && len(uuids) > 0 {
			continue
		}
		res = append(res, v)
	}
	if offset >= len(res) {
		return []types.Prediction{}, nil
	}
	res = res[offset:]
	if limit < len(res) {
		res = res[:limit]
	}
	return res, nil
}

// GetAccounts SELECTs accounts from the memory store.
func (s MemoryStateStorage) GetAccounts(filters types.APIAccountFilters, orderBys []string, limit, offset int) ([]types.Account, error) {
	handles := sliceToMap(filters.Handles)
	urls := sliceToMap(filters.URLs)

	res := []types.Account{}
	for _, v := range s.Accounts {
		if _, ok := handles[v.Handle]; !ok && len(handles) > 0 {
			continue
		}
		if _, ok := urls[v.URL.String()]; !ok && len(urls) > 0 {
			continue
		}
		res = append(res, v)
	}
	if offset >= len(res) {
		return []types.Account{}, nil
	}
	res = res[offset:]
	if limit < len(res) {
		res = res[:limit]
	}
	return res, nil
}

// UpsertPredictions upserts predictions to the memory store.
func (s MemoryStateStorage) UpsertPredictions(ps []*types.Prediction) ([]*types.Prediction, error) {
	for i, prediction := range ps {
		if prediction.UUID == "" {
			ps[i].UUID = uuid.NewString()
		}
		s.Predictions[prediction.PostUrl] = *prediction
	}
	return ps, nil
}

// LogPredictionStateValueChange logs the fact that a prediction changed PredictionStateValue to the memory store.
func (s MemoryStateStorage) LogPredictionStateValueChange(c types.PredictionStateValueChange) error {
	s.PredictionStateValueChanges[c.PK()] = c
	return nil
}

// UpsertAccounts upserts accounts to the memory store.
func (s MemoryStateStorage) UpsertAccounts(as []*types.Account) ([]*types.Account, error) {
	for _, a := range as {
		if a == nil {
			continue
		}
		s.Accounts[a.URL.String()] = *a
	}

	return as, nil
}

// TODO to implement these it's necessary to either wrap predictions in row objects or add fields onto Predictions

// PausePrediction pauses a prediction
func (s MemoryStateStorage) PausePrediction(uuid string) error {
	return nil
}

// UnpausePrediction unpauses a prediction.
func (s MemoryStateStorage) UnpausePrediction(uuid string) error {
	return nil
}

// HidePrediction hides a prediction.
func (s MemoryStateStorage) HidePrediction(uuid string) error {
	return nil
}

// UnhidePrediction unhides a prediction.
func (s MemoryStateStorage) UnhidePrediction(uuid string) error {
	return nil
}

// DeletePrediction deletes a prediction.
func (s MemoryStateStorage) DeletePrediction(uuid string) error {
	return nil
}

// UndeletePrediction undeletes a prediction.
func (s MemoryStateStorage) UndeletePrediction(uuid string) error {
	return nil
}

// PredictionInteractionExists answers if there was a Tweet for the given (prediction, actionType).
func (s MemoryStateStorage) PredictionInteractionExists(predictionUUID, postURL, actionType string) (bool, error) {
	return true, nil
}

// InsertPredictionInteraction notes down the fact that a Tweet for the given (prediction, actionType) happened.
func (s MemoryStateStorage) InsertPredictionInteraction(predictionUUID, postURL, actionType, interactionPostURL string) error {
	return nil
}

func sliceToMap(ss []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}
