package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/imagebuilder"
	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/swaggest/rest/nethttp"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v4"
)

// API is the main API struct.
type API struct {
	mux          *web.Service
	mkt          core.IMarket
	store        statestorage.StateStorage
	mFetcher     metadatafetcher.MetadataFetcher
	NowFunc      func() time.Time
	debug        bool
	imageBuilder imagebuilder.PredictionImageBuilder
}

// NewAPI is the constructor for the API.
func NewAPI(
	mkt core.IMarket,
	store statestorage.StateStorage,
	mFetcher metadatafetcher.MetadataFetcher,
	imgBuilder imagebuilder.PredictionImageBuilder,
	basicAuthUser string,
	basicAuthPass string,
) *API {
	a := &API{mkt: mkt, store: store, NowFunc: time.Now, mFetcher: mFetcher, imageBuilder: imgBuilder}

	apiSchema, service := buildAPIBoilerplate()

	// Prepare middleware with suitable security schema.
	// It will perform actual security check for every relevant request.
	adminAuth := middleware.BasicAuth("Admin Access", map[string]string{basicAuthUser: basicAuthPass})

	var (
		handlerGetPagesPrediction   = a.apiGetPagesPrediction()
		handlerHealthcheck          = nethttp.NewHandler(a.apiHealthcheck())
		handlerGetPredictions       = nethttp.NewHandler(a.apiGetPredictions())
		handlerPostPrediction       = nethttp.NewHandler(a.apiPostPrediction())
		handlerGetPredictionImage   = nethttp.NewHandler(a.apiGetPredictionImage())
		handlerPausePrediction      = nethttp.NewHandler(a.apiPredictionStorageActionWithUUID(a.store.PausePrediction, "Paused predictions are not updated by daemon until unpaused. They are still returned by GET calls."))
		handlerUnpausePrediction    = nethttp.NewHandler(a.apiPredictionStorageActionWithUUID(a.store.UnpausePrediction, "Upausing makes daemon resume updating predictions."))
		handlerHidePrediction       = nethttp.NewHandler(a.apiPredictionStorageActionWithUUID(a.store.HidePrediction, "Hidden predictions are not visible to any GET calls (unless showHidden is set), but they are still updated by daemon."))
		handlerUnhidePrediction     = nethttp.NewHandler(a.apiPredictionStorageActionWithUUID(a.store.UnhidePrediction, "Unhiding makes predictions visible to GET calls."))
		handlerDeletePrediction     = nethttp.NewHandler(a.apiPredictionStorageActionWithUUID(a.store.DeletePrediction, "Deleted predictions are not visible to any GET calls (unless showDeleted is set), nor updated by daemon."))
		handlerUndeletePrediction   = nethttp.NewHandler(a.apiPredictionStorageActionWithUUID(a.store.UndeletePrediction, "Undeleting predictions makes them visible to GET calls and updateable by daemon."))
		handlerRefetchAccount       = nethttp.NewHandler(a.apiPredictionRefetchAccount())
		handlerPredictionClearState = nethttp.NewHandler(a.apiPredictionClearState())
		handlerMaintenance          = nethttp.NewHandler(a.apiMaintenance())
	)

	service.Get("/pages/prediction/{id}", handlerGetPagesPrediction)
	service.Group(func(r chi.Router) {
		r.Use(adminAuth, nethttp.HTTPBasicSecurityMiddleware(apiSchema, "Admin", "Admin access"))
		r.Method(http.MethodGet, "/", handlerHealthcheck)
		r.Method(http.MethodGet, "/predictions", handlerGetPredictions)
		r.Method(http.MethodPost, "/predictions", handlerPostPrediction)
		r.Method(http.MethodGet, "/predictions/{uuid}/image", handlerGetPredictionImage)
		r.Method(http.MethodPost, "/predictions/{uuid}/pause", handlerPausePrediction)
		r.Method(http.MethodPost, "/predictions/{uuid}/unpause", handlerUnpausePrediction)
		r.Method(http.MethodPost, "/predictions/{uuid}/hide", handlerHidePrediction)
		r.Method(http.MethodPost, "/predictions/{uuid}/unhide", handlerUnhidePrediction)
		r.Method(http.MethodPost, "/predictions/{uuid}/delete", handlerDeletePrediction)
		r.Method(http.MethodPost, "/predictions/{uuid}/undelete", handlerUndeletePrediction)
		r.Method(http.MethodPost, "/predictions/{uuid}/refetchAccount", handlerRefetchAccount)
		r.Method(http.MethodPost, "/predictions/{uuid}/clearState", handlerPredictionClearState)
		r.Method(http.MethodPost, "/maintenance/{action}", handlerMaintenance)
		service.Docs("/docs", swgui.New)
	})

	a.mux = service

	return a
}

// SetDebug enables debug logging across the entire API.
func (a *API) SetDebug(b bool) {
	a.debug = b
}

// MustBlockinglyListenAndServe serves the API.
func (a *API) MustBlockinglyListenAndServe(apiURL string) {
	// If url starts with https?://, remove that part for the listener address
	var (
		rawURLParts = strings.Split(apiURL, "//")
		listenURL   = rawURLParts[len(rawURLParts)-1]
		l, err      = net.Listen("tcp", listenURL)
	)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Info().Str("docs", fmt.Sprintf("%v/docs", l.Addr().String())).Msgf("API listening on %v", l.Addr().String())

	if err := http.Serve(l, a.mux); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}
