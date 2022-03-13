package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/daemon"
	"github.com/marianogappa/predictions/types"
)

type newBody struct {
	Store      bool            `json:"store"`
	Prediction json.RawMessage `json:"prediction"`
}

func (a *API) newHandler(w http.ResponseWriter, r *http.Request) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		respond(w, nil, nil, pBool(false), fmt.Errorf("%w: %v", ErrInvalidRequestBody, err))
		return
	}
	defer r.Body.Close()

	var newBody newBody
	err = json.Unmarshal(bs, &newBody)
	if err != nil {
		respond(w, nil, nil, pBool(false), fmt.Errorf("%w: %v", ErrInvalidRequestJSON, err))
		return
	}

	log.Printf("API.newHandler: received request: %+v\n", newBody)

	pc := compiler.NewPredictionCompiler(&a.mFetcher, a.NowFunc)
	pred, err := pc.Compile(newBody.Prediction)
	if err != nil {
		respond(w, nil, nil, pBool(false), fmt.Errorf("%w: %v", ErrFailedToCompilePrediction, err))
		return
	}

	// If the state is empty, run one tick to see if the prediction is decided at start time. If so, it's invalid.
	if pred.State == (types.PredictionState{}) {
		predRunner, errs := daemon.NewPredRunner(&pred, a.mkt, int(a.NowFunc().Unix()))
		if len(errs) == 0 {
			predRunnerErrs := predRunner.Run()
			for _, err := range predRunnerErrs {
				if errors.Is(err, types.ErrInvalidMarketPair) {
					respond(w, nil, nil, pBool(false), types.ErrInvalidMarketPair)
					return
				}
			}
			if pred.Evaluate().IsFinal() {
				respond(w, nil, nil, pBool(false), types.ErrPredictionFinishedAtStartTime)
				return
			}
		}
	}

	if newBody.Store {
		// N.B. as per interface, UpsertPredictions may add UUIDs in-place on predictions
		_, err = a.store.UpsertPredictions([]*types.Prediction{&pred})
		if err != nil {
			respond(w, nil, nil, pBool(false), fmt.Errorf("%w: %v", ErrStorageErrorStoringPrediction, err))
			return
		}
	}

	bs, err = compiler.NewPredictionSerializer().Serialize(&pred)
	if err != nil {
		respond(w, nil, nil, pBool(false), fmt.Errorf("%w: %v", ErrFailedToSerializePredictions, err))
		return
	}
	raw := json.RawMessage(bs)

	respond(w, &raw, nil, pBool(newBody.Store), nil)
}

func pBool(b bool) *bool {
	return &b
}
