package api

import (
	"context"

	"github.com/swaggest/usecase"
)

type apiResHealthcheck struct {
	Success bool     `json:"success"`
	_       struct{} `query:"_" additionalProperties:"false"`
}

func (a *API) getHealthcheck(req struct{}) apiResponse[apiResHealthcheck] {
	return apiResponse[apiResHealthcheck]{Status: 200, Data: apiResHealthcheck{Success: true}}
}

func (a *API) apiHealthcheck() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input struct{}, output *apiResponse[apiResHealthcheck]) error {
		out := a.getHealthcheck(input)
		*output = out
		return nil
	})
	u.SetTags("Healthcheck")
	return u
}
