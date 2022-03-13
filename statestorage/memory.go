package statestorage

import (
	"github.com/google/uuid"
	"github.com/marianogappa/predictions/types"
)

type MemoryStateStorage struct {
	Predictions map[string]types.Prediction
}

func NewMemoryStateStorage() MemoryStateStorage {
	return MemoryStateStorage{Predictions: map[string]types.Prediction{}}
}

func sliceToMap(ss []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func (s MemoryStateStorage) GetPredictions(filters types.APIFilters, orderBys []string) ([]types.Prediction, error) {
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
	return res, nil
}

func (s MemoryStateStorage) UpsertPredictions(ps []*types.Prediction) ([]*types.Prediction, error) {
	for i, prediction := range ps {
		if prediction.UUID == "" {
			ps[i].UUID = uuid.NewString()
		}
		s.Predictions[prediction.PostUrl] = *prediction
	}
	return ps, nil
}
