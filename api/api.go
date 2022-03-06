package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/smrunner"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
	"github.com/marianogappa/signal-checker/common"
)

type API struct {
	mkt     market.Market
	store   statestorage.StateStorage
	NowFunc func() time.Time
}

func NewAPI(mkt market.Market, store statestorage.StateStorage) *API {
	return &API{mkt: mkt, store: store, NowFunc: time.Now}
}

func (a *API) MustBlockinglyServe(port int) {
	http.HandleFunc("/new", a.newHandler)
	http.HandleFunc("/get", a.getHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func (a *API) newHandler(w http.ResponseWriter, r *http.Request) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		respond(w, nil, nil, err)
		return
	}
	defer r.Body.Close()

	pc := compiler.NewPredictionCompiler()
	pred, err := pc.Compile(bs)
	if err != nil {
		respond(w, nil, nil, err)
		return
	}

	// If the state is empty, run one tick to see if the prediction is decided at start time. If so, it's invalid.
	if pred.State == (types.PredictionState{}) {
		predRunner, errs := smrunner.NewPredRunner(&pred, a.mkt, int(a.NowFunc().Unix()))
		if len(errs) == 0 {
			predRunnerErrs := predRunner.Run()
			for _, err := range predRunnerErrs {
				if errors.Is(err, common.ErrInvalidMarketPair) {
					respond(w, nil, nil, common.ErrInvalidMarketPair)
					return
				}
			}
			if pred.Evaluate().IsFinal() {
				respond(w, nil, nil, types.ErrPredictionFinishedAtStartTime)
				return
			}
		}
	}

	err = a.store.UpsertPredictions(map[string]types.Prediction{"unused": pred})
	if err != nil {
		respond(w, nil, nil, err)
		return
	}

	bs, _ = compiler.NewPredictionSerializer().Serialize(&pred)
	raw := json.RawMessage(bs)

	respond(w, &raw, nil, err)
}

type response struct {
	Status          int                         `json:"status"`
	Message         string                      `json:"message,omitempty"`
	InternalMessage string                      `json:"internalMessage,omitempty"`
	ErrorCode       string                      `json:"errorCode,omitempty"`
	Prediction      *json.RawMessage            `json:"prediction,omitempty"`
	Predictions     *map[string]json.RawMessage `json:"predictions,omitempty"`
}

func respond(w http.ResponseWriter, pred *json.RawMessage, preds *map[string]json.RawMessage, err error) {
	if err == nil {
		doRespond(w, response{Message: "", Prediction: pred, Predictions: preds, Status: 200})
		return
	}

	r := response{Message: "Unknown internal error.", Status: 500, InternalMessage: err.Error()}
	for maybeErr, maybeResp := range errToResponse {
		if errors.Is(err, maybeErr) {
			r = maybeResp
			r.InternalMessage = err.Error()
		}
	}
	doRespond(w, r)
}

func doRespond(w http.ResponseWriter, r response) {
	w.WriteHeader(r.Status)
	enc := json.NewEncoder(w)
	enc.Encode(r)
}

func (a *API) getHandler(w http.ResponseWriter, r *http.Request) {
	preds, err := a.store.GetPredictions([]types.PredictionStateValue{
		types.ONGOING_PRE_PREDICTION,
		types.ONGOING_PREDICTION,
		types.ANNULLED,
		types.INCORRECT,
		types.CORRECT,
	})
	if err != nil {
		respond(w, nil, nil, err)
		return
	}

	ps := compiler.NewPredictionSerializer()

	raws := map[string]json.RawMessage{}
	for key, pred := range preds {
		bs, _ := ps.Serialize(&pred)
		raw := json.RawMessage(bs)
		raws[key] = raw
	}

	respond(w, nil, &raws, nil)
}
