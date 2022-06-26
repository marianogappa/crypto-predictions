package api

import (
	"context"
	"errors"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5/middleware"
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
)

func buildAPIBoilerplate() (*openapi.Collector, *web.Service) {
	apiSchema := &openapi.Collector{}
	apiSchema.Reflector().SpecEns().Info.Title = "Crypto Predictions"
	apiSchema.Reflector().SpecEns().Info.WithDescription("Description!!!")
	apiSchema.Reflector().SpecEns().Info.Version = "v1.0.0"

	validatorFactory := jsonschema.NewFactory(apiSchema, apiSchema)
	decoderFactory := request.NewDecoderFactory()
	decoderFactory.ApplyDefaults = true
	decoderFactory.SetDecoderFunc(rest.ParamInPath, chirouter.PathToURLValues)

	service := web.DefaultService()

	service.OpenAPICollector.Reflector().DefaultOptions = append(service.OpenAPICollector.Reflector().DefaultOptions, func(rc *jsonschemago.ReflectContext) {
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

	service.Use(
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
		corsMiddleware,
	)

	return apiSchema, service
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}

		next.ServeHTTP(w, r)
	})
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
