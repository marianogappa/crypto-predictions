package statestorage

import (
	"encoding/json"
	"log"

	"github.com/boltdb/bolt"
	"github.com/marianogappa/predictions/types"
)

type BoltDBStateStorage struct {
	db *bolt.DB
}

func NewBoltDBStateStorage() (BoltDBStateStorage, error) {
	db, err := bolt.Open("predictions.db", 0600, nil)
	if err != nil {
		return BoltDBStateStorage{}, err
	}
	tx, err := db.Begin(true)
	if err != nil {
		return BoltDBStateStorage{}, err
	}
	defer tx.Rollback()

	// Use the transaction...
	if _, err := tx.CreateBucketIfNotExists([]byte("predictions")); err != nil {
		return BoltDBStateStorage{}, err
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		return BoltDBStateStorage{}, err
	}
	return BoltDBStateStorage{db: db}, nil
}

func (s BoltDBStateStorage) GetPredictions(css []types.PredictionStateValue) (map[string]types.Prediction, error) {
	csMap := map[types.PredictionStateValue]struct{}{}
	for _, cs := range css {
		csMap[cs] = struct{}{}
	}

	res := map[string]types.Prediction{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("predictions"))
		b.ForEach(func(k, v []byte) error {
			var p types.Prediction
			if err := json.Unmarshal(v, &p); err != nil {
				log.Printf("BoltDBStateStorage.GetPredictions: error unmarshalling prediction: %v\n", err)
				return nil
			}
			if _, ok := csMap[p.State.Value]; ok {
				res[string(k)] = p
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s BoltDBStateStorage) UpsertPredictions(ps map[string]types.Prediction) error {
	s.db.Update(func(tx *bolt.Tx) error {
		for _, prediction := range ps {
			b := tx.Bucket([]byte("predictions"))
			bs, err := json.Marshal(&prediction)
			if err != nil {
				log.Printf("BoltDBStateStorage.UpsertPredictions: error marshalling prediction: %v\n", err)
				continue
			}
			err = b.Put([]byte(prediction.PostUrl), bs)
			if err != nil {
				log.Printf("BoltDBStateStorage.UpsertPredictions: error putting marshalled prediction: %v\n", err)
				continue
			}
		}
		return nil
	})
	return nil
}
