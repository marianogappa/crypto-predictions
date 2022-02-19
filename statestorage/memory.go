package statestorage

import "github.com/marianogappa/predictions/types"

type MemoryStateStorage struct {
	predictions map[string]types.Prediction
}

func NewMemoryStateStorage() MemoryStateStorage {
	return MemoryStateStorage{predictions: map[string]types.Prediction{}}
}

func (s MemoryStateStorage) GetPredictions(css []types.PredictionStateValue) (map[string]types.Prediction, error) {
	csMap := map[types.PredictionStateValue]struct{}{}
	for _, cs := range css {
		csMap[cs] = struct{}{}
	}

	res := map[string]types.Prediction{}
	for k, v := range s.predictions {
		if _, ok := csMap[v.State.Value]; ok {
			res[k] = v
		}
	}
	return res, nil
}

func (s MemoryStateStorage) UpsertPredictions(ps map[string]types.Prediction) error {
	for _, prediction := range ps {
		s.predictions[prediction.Post] = prediction
	}
	return nil
}
