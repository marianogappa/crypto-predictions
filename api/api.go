package api

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/marianogappa/predictions/statestorage"
)

type API struct {
	mux      *http.ServeMux
	mkt      market.IMarket
	store    statestorage.StateStorage
	mFetcher metadatafetcher.MetadataFetcher
	NowFunc  func() time.Time
}

func NewAPI(mkt market.IMarket, store statestorage.StateStorage, mFetcher metadatafetcher.MetadataFetcher) *API {
	a := &API{mkt: mkt, store: store, NowFunc: time.Now, mFetcher: mFetcher}

	mux := http.NewServeMux()
	mux.HandleFunc("/new", a.newHandler)
	mux.HandleFunc("/get", a.getHandler)
	a.mux = mux

	return a
}

func (a *API) MustBlockinglyListenAndServe(apiURL string) {
	// If url starts with https?://, remove that part for the listener address
	rawUrlParts := strings.Split(apiURL, "//")
	listenUrl := rawUrlParts[len(rawUrlParts)-1]

	l, err := a.Listen(listenUrl)
	if err != nil {
		log.Fatal(err)
	}
	err = a.BlockinglyServe(l)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *API) Listen(url string) (net.Listener, error) {
	return net.Listen("tcp", url)
}

func (a *API) BlockinglyServe(l net.Listener) error {
	return http.Serve(l, a.mux)
}

type APIResponse struct {
	Status          int                `json:"status"`
	Message         string             `json:"message,omitempty"`
	InternalMessage string             `json:"internalMessage,omitempty"`
	ErrorCode       string             `json:"errorCode,omitempty"`
	Prediction      *json.RawMessage   `json:"prediction,omitempty"`
	Predictions     *[]json.RawMessage `json:"predictions,omitempty"`
	Stored          *bool
}

func respond(w http.ResponseWriter, pred *json.RawMessage, preds *[]json.RawMessage, stored *bool, err error) {
	if err == nil {
		doRespond(w, APIResponse{Message: "", Prediction: pred, Predictions: preds, Stored: stored, Status: 200})
		return
	}

	r := APIResponse{Message: "Unknown internal error.", Status: 500, InternalMessage: err.Error()}
	for maybeErr, maybeResp := range errToResponse {
		if errors.Is(err, maybeErr) {
			r = maybeResp
			r.InternalMessage = err.Error()
		}
	}
	doRespond(w, r)
}

func doRespond(w http.ResponseWriter, r APIResponse) {
	log.Printf("API.doRespond: responding request: %+v\n", r)

	w.WriteHeader(r.Status)
	enc := json.NewEncoder(w)
	enc.Encode(r)
}
