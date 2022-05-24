package api

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/marianogappa/predictions/imagebuilder"
	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/marianogappa/predictions/statestorage"
	jsonschemago "github.com/swaggest/jsonschema-go"
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
	mux          *web.Service
	mkt          market.IMarket
	store        statestorage.StateStorage
	mFetcher     metadatafetcher.MetadataFetcher
	NowFunc      func() time.Time
	debug        bool
	imageBuilder imagebuilder.PredictionImageBuilder
}

func NewAPI(mkt market.IMarket, store statestorage.StateStorage, mFetcher metadatafetcher.MetadataFetcher, files embed.FS) *API {
	a := &API{mkt: mkt, store: store, NowFunc: time.Now, mFetcher: mFetcher, imageBuilder: imagebuilder.NewPredictionImageBuilder(mkt, files)}

	apiSchema := &openapi.Collector{}
	apiSchema.Reflector().SpecEns().Info.Title = "Crypto Predictions"
	apiSchema.Reflector().SpecEns().Info.WithDescription("Description!!!")
	apiSchema.Reflector().SpecEns().Info.Version = "v1.0.0"

	validatorFactory := jsonschema.NewFactory(apiSchema, apiSchema)
	decoderFactory := request.NewDecoderFactory()
	decoderFactory.ApplyDefaults = true
	decoderFactory.SetDecoderFunc(rest.ParamInPath, chirouter.PathToURLValues)

	s := web.DefaultService()

	s.OpenAPICollector.Reflector().DefaultOptions = append(s.OpenAPICollector.Reflector().DefaultOptions, func(rc *jsonschemago.ReflectContext) {
		it := rc.InterceptType
		rc.InterceptType = func(value reflect.Value, schema *jsonschemago.Schema) (bool, error) {
			stop, err := it(value, schema)
			if err != nil {
				return stop, err
			}

			if schema.HasType(jsonschemago.Object) && len(schema.Properties) > 0 && schema.AdditionalProperties == nil {
				schema.AdditionalProperties = (&jsonschemago.SchemaOrBool{}).WithTypeBoolean(false)
			}

			return stop, nil
		}
	})

	s.Use(
		middleware.Recoverer,
		nethttp.OpenAPIMiddleware(apiSchema),
		request.DecoderMiddleware(decoderFactory),
		request.ValidatorMiddleware(validatorFactory),

		// Example middleware to setup custom error responses.
		func(handler http.Handler) http.Handler {
			var h *nethttp.Handler
			if nethttp.HandlerAs(handler, &h) {
				h.MakeErrResp = func(ctx context.Context, err error) (int, interface{}) {
					code, er := rest.Err(err)

					return code, apiResponse[map[string]interface{}]{
						Status:               code,
						ErrorMessage:         ErrInvalidRequestJSON.Error(),
						InternalErrorMessage: er.ErrorText,
						ErrorCode:            "ErrInvalidRequestJSON",
						Data:                 er.Context,
					}
				}
			}

			return handler
		},
		restResponse.EncoderMiddleware,
		gzip.Middleware,
	)

	s.Get("/predictions", a.apiGetPredictions())
	s.Get("/pages/prediction/{url}", a.apiGetPagesPrediction())
	s.Post("/predictions", a.apiPostPrediction())
	s.Get("/predictions/{uuid}/image", a.apiGetPredictionImage())
	s.Post("/predictions/{uuid}/pause", a.apiPredictionStorageActionWithUUID(a.store.PausePrediction, "Paused predictions are not updated by daemon until unpaused. They are still returned by GET calls."))
	s.Post("/predictions/{uuid}/unpause", a.apiPredictionStorageActionWithUUID(a.store.UnpausePrediction, "Upausing makes daemon resume updating predictions."))
	s.Post("/predictions/{uuid}/hide", a.apiPredictionStorageActionWithUUID(a.store.HidePrediction, "Hidden predictions are not visible to any GET calls (unless showHidden is set), but they are still updated by daemon."))
	s.Post("/predictions/{uuid}/unhide", a.apiPredictionStorageActionWithUUID(a.store.UnhidePrediction, "Unhiding makes predictions visible to GET calls."))
	s.Post("/predictions/{uuid}/delete", a.apiPredictionStorageActionWithUUID(a.store.DeletePrediction, "Deleted predictions are not visible to any GET calls (unless showDeleted is set), nor updated by daemon."))
	s.Post("/predictions/{uuid}/undelete", a.apiPredictionStorageActionWithUUID(a.store.UndeletePrediction, "Undeleting predictions makes them visible to GET calls and updateable by daemon."))
	s.Post("/predictions/{uuid}/refetchAccount", a.apiPredictionRefetchAccount())
	s.Post("/predictions/{uuid}/clearState", a.apiPredictionClearState())
	s.Post("/maintenance/{action}", a.apiMaintenance())

	s.Docs("/docs", swgui.New)

	a.mux = s

	return a
}

func (c *API) SetDebug(b bool) {
	c.debug = b
}

type apiResponse[D any] struct {
	Status               int    `json:"status" example:"200"`
	ErrorMessage         string `json:"errorMessage,omitempty" example:""`
	InternalErrorMessage string `json:"internalErrorMessage,omitempty" example:""`
	ErrorCode            string `json:"errorCode,omitempty" example:""`
	Data                 D      `json:"data,omitempty"`
}

func failWith[D any](errType, err error, zero D) apiResponse[D] {
	if err == nil {
		err = errors.New(errToResponse[errType].Message)
	}
	return apiResponse[D]{
		Status:               errToResponse[errType].StatusCode,
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
