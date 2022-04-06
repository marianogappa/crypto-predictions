package api

import (
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
	mux.HandleFunc("/new", buildHandler(a.handleNew))
	mux.HandleFunc("/get", buildHandler(a.handleGet))
	mux.HandleFunc("/prediction", buildHandler(a.handleBodyPagePrediction))
	mux.HandleFunc("/predictionPause", buildHandler(a.handlePause))
	mux.HandleFunc("/predictionUnpause", buildHandler(a.handleUnpause))
	mux.HandleFunc("/predictionHide", buildHandler(a.handleHide))
	mux.HandleFunc("/predictionUnhide", buildHandler(a.handleUnhide))
	mux.HandleFunc("/predictionDelete", buildHandler(a.handleDelete))
	mux.HandleFunc("/predictionUndelete", buildHandler(a.handleUndelete))
	mux.HandleFunc("/predictionRefetchAccount", buildHandler(a.handleRefetchAccount))
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
