package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/swaggest/rest"
	"github.com/swaggest/rest/chirouter"
	"github.com/swaggest/rest/jsonschema"
	"github.com/swaggest/rest/nethttp"
	"github.com/swaggest/rest/openapi"
	"github.com/swaggest/rest/request"
	restResponse "github.com/swaggest/rest/response"
	"github.com/swaggest/rest/response/gzip"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v4"
)

type API struct {
	mux      *web.Service
	mkt      market.IMarket
	store    statestorage.StateStorage
	mFetcher metadatafetcher.MetadataFetcher
	NowFunc  func() time.Time
	debug    bool
}

func NewAPI(mkt market.IMarket, store statestorage.StateStorage, mFetcher metadatafetcher.MetadataFetcher) *API {
	a := &API{mkt: mkt, store: store, NowFunc: time.Now, mFetcher: mFetcher}

	apiSchema := &openapi.Collector{}
	apiSchema.Reflector().SpecEns().Info.Title = "Crypto Predictions"
	apiSchema.Reflector().SpecEns().Info.WithDescription("Description!!!")
	apiSchema.Reflector().SpecEns().Info.Version = "v1.0.0"

	validatorFactory := jsonschema.NewFactory(apiSchema, apiSchema)
	decoderFactory := request.NewDecoderFactory()
	decoderFactory.ApplyDefaults = true
	decoderFactory.SetDecoderFunc(rest.ParamInPath, chirouter.PathToURLValues)

	s := web.DefaultService()

	s.Use(
		middleware.Recoverer,
		nethttp.OpenAPIMiddleware(apiSchema),
		request.DecoderMiddleware(decoderFactory),
		request.ValidatorMiddleware(validatorFactory),
		restResponse.EncoderMiddleware,
		gzip.Middleware,
	)

	s.Get("/predictions", a.apiGetPredictions())
	s.Get("/pages/prediction/{url}", a.apiGetPagesPrediction())
	s.Post("/predictions", a.apiPostPrediction())
	s.Post("/predictions/{uuid}/pause", a.apiPredictionStorageActionWithUUID(a.store.PausePrediction, "Paused predictions are not updated by daemon until unpaused. They are still returned by GET calls."))
	s.Post("/predictions/{uuid}/unpause", a.apiPredictionStorageActionWithUUID(a.store.UnpausePrediction, "Upausing makes daemon resume updating predictions."))
	s.Post("/predictions/{uuid}/hide", a.apiPredictionStorageActionWithUUID(a.store.HidePrediction, "Hidden predictions are not visible to any GET calls (unless showHidden is set), but they are still updated by daemon."))
	s.Post("/predictions/{uuid}/unhide", a.apiPredictionStorageActionWithUUID(a.store.UnhidePrediction, "Unhiding makes predictions visible to GET calls."))
	s.Post("/predictions/{uuid}/delete", a.apiPredictionStorageActionWithUUID(a.store.DeletePrediction, "Deleted predictions are not visible to any GET calls (unless showDeleted is set), nor updated by daemon."))
	s.Post("/predictions/{uuid}/undelete", a.apiPredictionStorageActionWithUUID(a.store.UndeletePrediction, "Undeleting predictions makes them visible to GET calls and updateable by daemon."))
	s.Post("/predictions/{uuid}/refetchAccount", a.apiPredictionRefetchAccount())

	s.Docs("/docs", swgui.New)

	a.mux = s

	return a
}

func (c *API) SetDebug(b bool) {
	c.debug = b
}

type apiResponse[D any] struct {
	Status               int    `json:"status"`
	ErrorMessage         string `json:"errorMessage,omitempty"`
	InternalErrorMessage string `json:"internalErrorMessage,omitempty"`
	ErrorCode            string `json:"errorCode,omitempty"`
	Data                 D      `json:"data,omitempty"`
}

func failWith[D any](errType, err error, zero D) apiResponse[D] {
	return apiResponse[D]{
		Status:               errToResponse[errType].Status,
		ErrorMessage:         errToResponse[errType].Message,
		InternalErrorMessage: err.Error(),
		ErrorCode:            errToResponse[errType].ErrorCode,
		Data:                 zero,
	}
}

func (a *API) MustBlockinglyListenAndServe(apiURL string) {
	// If url starts with https?://, remove that part for the listener address
	rawUrlParts := strings.Split(apiURL, "//")
	listenUrl := rawUrlParts[len(rawUrlParts)-1]

	l, err := a.Listen(listenUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	err = a.BlockinglyServe(l)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func (a *API) Listen(url string) (net.Listener, error) {
	return net.Listen("tcp", url)
}

func (a *API) BlockinglyServe(l net.Listener) error {
	log.Info().Str("docs", fmt.Sprintf("%v/docs", l.Addr().String())).Msgf("API listening on %v", l.Addr().String())
	return http.Serve(l, a.mux)
}
